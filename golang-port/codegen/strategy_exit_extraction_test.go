package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Test strategy.exit() argument extraction pattern */
func TestStrategyExitArgumentExtraction(t *testing.T) {
	g := &generator{}
	extractor := &ArgumentExtractor{generator: g}

	/* Simulate: strategy.exit("Exit", "Long", stop=95.0, limit=110.0) */
	fullArgs := []ast.Expression{
		&ast.Literal{Value: "Exit"}, // exitID
		&ast.Literal{Value: "Long"}, // fromEntry
		&ast.ObjectExpression{ // Named args
			NodeType: ast.TypeObjectExpression,
			Properties: []ast.Property{
				{
					NodeType: ast.TypeProperty,
					Key:      &ast.Identifier{Name: "stop"},
					Value:    &ast.Literal{Value: 95.0},
				},
				{
					NodeType: ast.TypeProperty,
					Key:      &ast.Identifier{Name: "limit"},
					Value:    &ast.Literal{Value: 110.0},
				},
			},
		},
	}

	/* After skipping first 2 args (exitID, fromEntry) */
	remainingArgs := fullArgs[2:]

	/* Named extraction should work */
	stopCode, stopFound := extractor.ExtractNamedArgument(remainingArgs, "stop")
	if !stopFound {
		t.Fatal("Expected stop argument to be found in remainingArgs")
	}
	if stopCode != "95" {
		t.Errorf("Expected '95.00', got %q", stopCode)
	}

	limitCode, limitFound := extractor.ExtractNamedArgument(remainingArgs, "limit")
	if !limitFound {
		t.Fatal("Expected limit argument to be found in remainingArgs")
	}
	if limitCode != "110" {
		t.Errorf("Expected '110.00', got %q", limitCode)
	}
}

/* Test with variables instead of literals */
func TestStrategyExitWithVariables(t *testing.T) {
	g := &generator{
		variables: map[string]string{
			"stop_val":  "float64",
			"limit_val": "float64",
		},
	}
	extractor := &ArgumentExtractor{generator: g}

	/* Simulate: strategy.exit("Exit", "Long", stop=stop_val, limit=limit_val) */
	fullArgs := []ast.Expression{
		&ast.Literal{Value: "Exit"},
		&ast.Literal{Value: "Long"},
		&ast.ObjectExpression{
			NodeType: ast.TypeObjectExpression,
			Properties: []ast.Property{
				{
					NodeType: ast.TypeProperty,
					Key:      &ast.Identifier{Name: "stop"},
					Value:    &ast.Identifier{Name: "stop_val"},
				},
				{
					NodeType: ast.TypeProperty,
					Key:      &ast.Identifier{Name: "limit"},
					Value:    &ast.Identifier{Name: "limit_val"},
				},
			},
		},
	}

	remainingArgs := fullArgs[2:]

	stopCode, stopFound := extractor.ExtractNamedArgument(remainingArgs, "stop")
	if !stopFound {
		t.Fatal("Expected stop argument to be found")
	}
	if stopCode != "stop_valSeries.GetCurrent()" {
		t.Errorf("Expected series access, got %q", stopCode)
	}

	limitCode, limitFound := extractor.ExtractNamedArgument(remainingArgs, "limit")
	if !limitFound {
		t.Fatal("Expected limit argument to be found")
	}
	if limitCode != "limit_valSeries.GetCurrent()" {
		t.Errorf("Expected series access, got %q", limitCode)
	}
}
