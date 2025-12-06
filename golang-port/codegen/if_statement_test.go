package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

func TestIfStatementCodegen(t *testing.T) {
	// Create a minimal strategy with if statement
	pineScript := `//@version=5
strategy("Test If", overlay=true)

signal = close > open

if (signal)
    strategy.entry("Long", strategy.long)
`

	// Parse Pine Script
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test-if.pine", []byte(pineScript))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Convert to AST
	converter := parser.NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion error: %v", err)
	}

	// Generate Go code
	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Codegen error: %v", err)
	}

	generated := code.FunctionBody
	t.Logf("Generated code:\n%s", generated)

	// Verify ForwardSeriesBuffer paradigm (ALL variables use Series)
	if !strings.Contains(generated, "signalSeries") {
		t.Errorf("Expected signalSeries (ForwardSeriesBuffer paradigm), got:\n%s", generated)
	}
	if !strings.Contains(generated, "signalSeries.Set(") {
		t.Errorf("Expected Series.Set() assignment, got:\n%s", generated)
	}
	// Bool variable stored as float64 in Series, needs != 0 for if condition
	if !strings.Contains(generated, "if signalSeries.GetCurrent() != 0") {
		t.Errorf("Expected 'if signalSeries.GetCurrent() != 0', got:\n%s", generated)
	}
	if !strings.Contains(generated, "strat.Entry(") {
		t.Errorf("Expected 'strat.Entry(', got:\n%s", generated)
	}

	// Make sure no TODO placeholders
	if strings.Contains(generated, "TODO: implement") {
		t.Errorf("Found TODO placeholder, if statement not properly generated:\n%s", generated)
	}
}
