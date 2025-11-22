package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

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

	pineParser, err := parser.NewParser()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create parser: %v\n", err)
		os.Exit(1)
	}

	sourceFilename := filepath.Base(*inputFlag)
	parsedAST, err := pineParser.ParseString(sourceFilename, string(sourceContent))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	pineVersion := detectPineVersion(string(sourceContent))
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

	temporaryDirectory := os.TempDir()
	temporaryGoFile := filepath.Join(temporaryDirectory, "pine_strategy_temp.go")

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
