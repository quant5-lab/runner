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

	/* Parse Pine strategy */
	content, err := os.ReadFile(*inputFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read input: %v\n", err)
		os.Exit(1)
	}

	p, err := parser.NewParser()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create parser: %v\n", err)
		os.Exit(1)
	}

	ast, err := p.ParseString(filepath.Base(*inputFlag), string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	/* Detect Pine version and apply preprocessing if needed */
	version := detectPineVersion(string(content))
	if version < 5 {
		fmt.Printf("Detected Pine v%d - applying v4→v5 preprocessing\n", version)
		pipeline := preprocessor.NewV4ToV5Pipeline()
		ast, err = pipeline.Run(ast)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Preprocessing error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Preprocessing complete\n")
	} else {
		fmt.Printf("Detected Pine v%d - no preprocessing needed\n", version)
	}

	/* Convert to ESTree */
	converter := parser.NewConverter()
	estree, err := converter.ToESTree(ast)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Conversion error: %v\n", err)
		os.Exit(1)
	}

	astJSON, err := converter.ToJSON(estree)
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON error: %v\n", err)
		os.Exit(1)
	}

	/* Analyze warmup requirements */
	analyzer := validation.NewWarmupAnalyzer()
	warmupReqs := analyzer.AnalyzeScript(estree)
	if len(warmupReqs) > 0 {
		fmt.Printf("Warmup requirements detected:\n")
		maxLookback := 0
		for _, req := range warmupReqs {
			fmt.Printf("  - %s (lookback: %d bars)\n", req.Source, req.MaxLookback)
			if req.MaxLookback > maxLookback {
				maxLookback = req.MaxLookback
			}
		}
		fmt.Printf("  ⚠️  Strategy requires at least %d bars of historical data\n", maxLookback+1)
		fmt.Printf("  💡 First %d bars will produce null/NaN values (warmup period)\n", maxLookback)
	}

	/* Generate Go code from AST */
	strategyCode, err := codegen.GenerateStrategyCodeFromAST(estree)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Codegen error: %v\n", err)
		os.Exit(1)
	}

	/* Inject security() prefetch code if needed */
	strategyCode, err = codegen.InjectSecurityCode(strategyCode, estree)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Security injection error: %v\n", err)
		os.Exit(1)
	}

	/* Create temp Go source file */
	tempDir := os.TempDir()
	tempGoFile := filepath.Join(tempDir, "pine_strategy_temp.go")

	err = codegen.InjectStrategy(*templateFlag, tempGoFile, strategyCode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Injection error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Parsed: %s\n", *inputFlag)
	fmt.Printf("Generated: %s\n", tempGoFile)
	fmt.Printf("AST size: %d bytes\n", len(astJSON))
	fmt.Printf("Next: Compile with: go build -o %s %s\n", *outputFlag, tempGoFile)
}

// detectPineVersion extracts version from //@version=N comment
func detectPineVersion(content string) int {
	re := regexp.MustCompile(`//@version\s*=\s*(\d+)`)
	matches := re.FindStringSubmatch(content)
	if len(matches) >= 2 {
		var version int
		fmt.Sscanf(matches[1], "%d", &version)
		return version
	}
	// No version comment = assume v4 (pre-v5 default)
	return 4
}
