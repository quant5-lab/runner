package integration

import (
	"os"
	"strings"
	"testing"

	"github.com/borisquantlab/pinescript-go/codegen"
	"github.com/borisquantlab/pinescript-go/parser"
)

func TestCrossoverCodegen(t *testing.T) {
	input := `
//@version=5
strategy("Crossover Test", overlay=true)

sma20 = ta.sma(close, 20)
longCrossover = ta.crossover(close, sma20)

if longCrossover
    strategy.entry("long", strategy.long)
`

	// Parse
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Convert to ESTree
	converter := parser.NewConverter()
	estree, err := converter.ToESTree(ast)
	if err != nil {
		t.Fatalf("ESTree conversion failed: %v", err)
	}

	// Generate code
	stratCode, err := codegen.GenerateStrategyCodeFromAST(estree)
	if err != nil {
		t.Fatalf("Codegen failed: %v", err)
	}

	goCode := stratCode.FunctionBody

	// Write to temp file
	tmpFile := "/tmp/test_crossover.go"
	err = os.WriteFile(tmpFile, []byte(goCode), 0644)
	if err != nil {
		t.Fatalf("Failed to write generated code: %v", err)
	}
	defer os.Remove(tmpFile)

	t.Logf("Generated code written to %s", tmpFile)
	t.Logf("Generated code:\n%s", goCode)

	// Verify key elements in generated code
	if !strings.Contains(goCode, "var sma20 float64") {
		t.Error("Missing sma20 float64 declaration")
	}
	if !strings.Contains(goCode, "var longCrossover bool") {
		t.Error("Missing longCrossover bool declaration")
	}
	if !strings.Contains(goCode, "Crossover") {
		t.Error("Missing crossover comment")
	}
	if !strings.Contains(goCode, "if i > 0") {
		t.Error("Missing warmup check for crossover")
	}
	if !strings.Contains(goCode, "bar.Close > sma20") {
		t.Error("Missing crossover condition")
	}
}
