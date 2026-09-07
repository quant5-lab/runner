package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/quant5-lab/runner/codegen"
	"github.com/quant5-lab/runner/parser"
	"github.com/quant5-lab/runner/preprocessor"
	"github.com/quant5-lab/runner/runtime/validation"
)

var (
	inputFlag    = flag.String("input", "", "Input Pine strategy file (.pine)")
	outputFlag   = flag.String("output", "", "Output Go binary path")
	templateFlag = flag.String("template", "template/main.go.tmpl", "Template file path")
)

func main() {
	flag.Parse()

	if *inputFlag == "" || *outputFlag == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -input STRATEGY.pine -output BINARY [-template TEMPLATE.tmpl]\n", os.Args[0])
		os.Exit(1)
	}

	sourceContent, err := os.ReadFile(*inputFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read input: %v\n", err)
		os.Exit(1)
	}

	sourceStr := string(sourceContent)
	sourceStr = preprocessor.ExpandTabs(sourceStr)
	sourceStr = preprocessor.RewriteGenericTypeSyntax(sourceStr)

	pineVersion := detectPineVersion(sourceStr)
	if pineVersion < 5 {
		sourceStr = transformInputTypeParameters(sourceStr)
	}

	pineParser, err := parser.NewParser()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create parser: %v\n", err)
		os.Exit(1)
	}

	sourceFilename := filepath.Base(*inputFlag)
	parsedAST, err := pineParser.ParseString(sourceFilename, sourceStr)
	if err != nil {
		reportBlockedParse(sourceFilename, err)
		os.Exit(1)
	}

	if pineVersion < 5 {
		fmt.Printf("Detected Pine v%d - applying v4→v5 preprocessing\n", pineVersion)
		preprocessingPipeline := preprocessor.NewV4ToV5Pipeline()
		parsedAST, err = preprocessingPipeline.Run(parsedAST)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Preprocessing error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Preprocessing complete\n")
	} else {
		fmt.Printf("Detected Pine v%d - no preprocessing needed\n", pineVersion)
	}

	astConverter := parser.NewConverter()
	estreeAST, err := astConverter.ToESTree(parsedAST)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Conversion error: %v\n", err)
		os.Exit(1)
	}
	estreeAST.PineVersion = pineVersion

	identifierSanitizer := preprocessor.NewIdentifierSanitizer()
	estreeAST, err = identifierSanitizer.Transform(estreeAST)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Identifier sanitization error: %v\n", err)
		os.Exit(1)
	}

	astJSON, err := astConverter.ToJSON(estreeAST)
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON error: %v\n", err)
		os.Exit(1)
	}

	warmupAnalyzer := validation.NewWarmupAnalyzer()
	warmupRequirements := warmupAnalyzer.AnalyzeScript(estreeAST)
	if len(warmupRequirements) > 0 {
		fmt.Printf("Warmup requirements detected:\n")
		maxLookbackBars := 0
		for _, requirement := range warmupRequirements {
			fmt.Printf("  - %s (lookback: %d bars)\n", requirement.Source, requirement.MaxLookback)
			if requirement.MaxLookback > maxLookbackBars {
				maxLookbackBars = requirement.MaxLookback
			}
		}
		fmt.Printf("  ⚠️  Strategy requires at least %d bars of historical data\n", maxLookbackBars+1)
		fmt.Printf("  💡 First %d bars will produce null/NaN values (warmup period)\n", maxLookbackBars)
	}

	strategyCode, err := codegen.GenerateStrategyCodeFromAST(estreeAST)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Codegen error: %v\n", err)
		os.Exit(1)
	}

	strategyCode.StrategyName = deriveStrategyNameFromSourceFile(*inputFlag)

	strategyCode, err = codegen.InjectSecurityCode(strategyCode, estreeAST)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Security injection error: %v\n", err)
		os.Exit(1)
	}

	reportFeatureGaps(strategyCode.FeatureGaps)

	temporaryDirectory := os.TempDir()

	/* Create unique temp file to avoid conflicts when running tests in parallel */
	tempFile, err := os.CreateTemp(temporaryDirectory, "pine_strategy_*.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp file: %v\n", err)
		os.Exit(1)
	}
	temporaryGoFile := tempFile.Name()
	tempFile.Close()

	err = codegen.InjectStrategy(*templateFlag, temporaryGoFile, strategyCode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Injection error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Parsed: %s\n", *inputFlag)
	fmt.Printf("Generated: %s\n", temporaryGoFile)
	fmt.Printf("AST size: %d bytes\n", len(astJSON))
	fmt.Printf("Next: Compile with: go build -o %s %s\n", *outputFlag, temporaryGoFile)
}

type blockedCompatibilityResult struct {
	ScriptCompatibility blockedCompatibility `json:"scriptCompatibility"`
	Backtest            blockedBacktest      `json:"backtest"`
}

type blockedCompatibility struct {
	Status      string              `json:"status"`
	Diagnostics []blockedDiagnostic `json:"diagnostics"`
}

type blockedBacktest struct {
	Status string `json:"status"`
}

type blockedDiagnostic struct {
	FeatureID    string `json:"featureId"`
	Phase        string `json:"phase"`
	Impact       string `json:"impact"`
	Substitution string `json:"substitution"`
	Source       string `json:"source"`
	File         string `json:"file,omitempty"`
	Line         int    `json:"line,omitempty"`
	Column       int    `json:"column,omitempty"`
	Message      string `json:"message"`
}

func reportBlockedParse(sourceFilename string, err error) {
	pos := errorPosition(err)
	result := blockedCompatibilityResult{
		ScriptCompatibility: blockedCompatibility{
			Status: "blocked",
			Diagnostics: []blockedDiagnostic{{
				FeatureID:    "parser.unrecoverable",
				Phase:        "parse",
				Impact:       "structural",
				Substitution: "none",
				Source:       sourceFilename,
				File:         fallbackString(pos.Filename, sourceFilename),
				Line:         pos.Line,
				Column:       pos.Column,
				Message:      err.Error(),
			}},
		},
		Backtest: blockedBacktest{Status: "blocked"},
	}
	payload, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "%s\n", payload)
}

func errorPosition(err error) lexer.Position {
	positioned, ok := err.(interface{ Position() lexer.Position })
	if !ok {
		return lexer.Position{}
	}
	return positioned.Position()
}

func fallbackString(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func reportFeatureGaps(gaps []string) {
	if len(gaps) == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "WARNING: %d unimplemented Pine function(s) — binary runs with stubs:\n", len(gaps))
	for _, name := range gaps {
		fmt.Fprintf(os.Stderr, "  - %s()\n", name)
	}
}

func deriveStrategyNameFromSourceFile(inputPath string) string {
	baseFilename := filepath.Base(inputPath)
	extension := filepath.Ext(baseFilename)
	return baseFilename[:len(baseFilename)-len(extension)]
}

func detectPineVersion(content string) int {
	versionPattern := regexp.MustCompile(`//@version\s*=\s*(\d+)`)
	matches := versionPattern.FindStringSubmatch(content)

	if len(matches) >= 2 {
		var versionNumber int
		fmt.Sscanf(matches[1], "%d", &versionNumber)
		return versionNumber
	}

	const defaultPineVersion = 4
	return defaultPineVersion
}

/* transformInputTypeParameters converts V4 input(..., type=input.X) to V5 input.X()
 * Removes type=input.X argument and renames function to input.X()
 */
func transformInputTypeParameters(source string) string {
	/* Pattern: input(..., type=input.X, ...) - captures type name and removes that arg */
	inputPattern := regexp.MustCompile(`input\s*\(([^)]*)\btype\s*=\s*input\.(\w+)\b\s*,?\s*([^)]*)\)`)

	return inputPattern.ReplaceAllStringFunc(source, func(match string) string {
		submatches := inputPattern.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		beforeType := submatches[1]
		inputType := submatches[2]
		afterType := submatches[3]

		/* Normalize v4 type names to v5 */
		switch inputType {
		case "integer":
			inputType = "int"
		}

		/* Combine arguments, removing trailing/leading commas */
		beforeType = regexp.MustCompile(`,\s*$`).ReplaceAllString(beforeType, "")
		afterType = regexp.MustCompile(`^\s*,\s*`).ReplaceAllString(afterType, "")

		args := ""
		if beforeType != "" && afterType != "" {
			args = beforeType + ", " + afterType
		} else if beforeType != "" {
			args = beforeType
		} else {
			args = afterType
		}

		/* Remove trailing comma if present */
		args = regexp.MustCompile(`,\s*$`).ReplaceAllString(args, "")

		return "input." + inputType + "(" + args + ")"
	})
}
