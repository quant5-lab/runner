package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestParsePlotOptions_SimpleVariable(t *testing.T) {
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "sma20"},
		},
	}

	opts := ParsePlotOptions(call)

	if opts.Variable != "sma20" {
		t.Errorf("Expected variable 'sma20', got '%s'", opts.Variable)
	}
	if opts.Title != "sma20" {
		t.Errorf("Expected title 'sma20', got '%s'", opts.Title)
	}
}

func TestParsePlotOptions_MemberExpression(t *testing.T) {
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "sma50"},
				Property: &ast.Literal{Value: 0},
				Computed: true,
			},
		},
	}

	opts := ParsePlotOptions(call)

	if opts.Variable != "sma50" {
		t.Errorf("Expected variable 'sma50', got '%s'", opts.Variable)
	}
}

func TestParsePlotOptions_WithTitle(t *testing.T) {
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "ema20"},
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key:   &ast.Identifier{Name: "title"},
						Value: &ast.Literal{Value: "EMA 20"},
					},
				},
			},
		},
	}

	opts := ParsePlotOptions(call)

	if opts.Variable != "ema20" {
		t.Errorf("Expected variable 'ema20', got '%s'", opts.Variable)
	}
	if opts.Title != "EMA 20" {
		t.Errorf("Expected title 'EMA 20', got '%s'", opts.Title)
	}
}

func TestParsePlotOptions_EmptyCall(t *testing.T) {
	call := &ast.CallExpression{
		Arguments: []ast.Expression{},
	}

	opts := ParsePlotOptions(call)

	if opts.Variable != "" {
		t.Errorf("Expected empty variable, got '%s'", opts.Variable)
	}
	if opts.Title != "" {
		t.Errorf("Expected empty title, got '%s'", opts.Title)
	}
}

func TestParsePlotOptions_MultipleProperties(t *testing.T) {
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "rsi"},
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key:   &ast.Identifier{Name: "color"},
						Value: &ast.Identifier{Name: "blue"},
					},
					{
						Key:   &ast.Identifier{Name: "title"},
						Value: &ast.Literal{Value: "RSI Indicator"},
					},
					{
						Key:   &ast.Identifier{Name: "linewidth"},
						Value: &ast.Literal{Value: 2},
					},
				},
			},
		},
	}

	opts := ParsePlotOptions(call)

	if opts.Variable != "rsi" {
		t.Errorf("Expected variable 'rsi', got '%s'", opts.Variable)
	}
	if opts.Title != "RSI Indicator" {
		t.Errorf("Expected title 'RSI Indicator', got '%s'", opts.Title)
	}
}
