package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestBooleanConverter_Integration_EndToEnd(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	typeSystem.RegisterVariable("enabled", "bool")
	typeSystem.RegisterVariable("condition", "bool")
	converter := NewBooleanConverter(typeSystem)

	tests := []struct {
		name        string
		method      string
		expr        ast.Expression
		code        string
		expected    string
		description string
	}{
		{
			name:        "EnsureBooleanOperand: Series access gets parentheses",
			method:      "EnsureBooleanOperand",
			expr:        &ast.Identifier{Name: "signal"},
			code:        "signalSeries.GetCurrent()",
			expected:    "(signalSeries.GetCurrent() != 0)",
			description: "Series variables in logical expressions need parentheses",
		},
		{
			name:   "EnsureBooleanOperand: comparison unchanged",
			method: "EnsureBooleanOperand",
			expr: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "price"},
				Right:    &ast.Literal{Value: 100.0},
			},
			code:        "price > 100",
			expected:    "price > 100",
			description: "Comparisons are already boolean",
		},
		{
			name:        "ConvertBoolSeriesForIfStatement: Series gets != 0",
			method:      "ConvertBoolSeriesForIfStatement",
			expr:        &ast.Identifier{Name: "value"},
			code:        "valueSeries.GetCurrent()",
			expected:    "valueSeries.GetCurrent() != 0",
			description: "If conditions need explicit != 0 for Series",
		},
		{
			name:   "ConvertBoolSeriesForIfStatement: comparison unchanged",
			method: "ConvertBoolSeriesForIfStatement",
			expr: &ast.BinaryExpression{
				Operator: "<",
				Left:     &ast.Identifier{Name: "low"},
				Right:    &ast.Identifier{Name: "stopLevel"},
			},
			code:        "bar.Low < stopLevelSeries.GetCurrent()",
			expected:    "bar.Low < stopLevelSeries.GetCurrent()",
			description: "Skip conversion when comparison already present",
		},
		{
			name:        "ConvertBoolSeriesForIfStatement: bool type",
			method:      "ConvertBoolSeriesForIfStatement",
			expr:        &ast.Identifier{Name: "enabled"},
			code:        "enabledSeries.GetCurrent()",
			expected:    "enabledSeries.GetCurrent() != 0",
			description: "Bool-typed variables converted even without pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			switch tt.method {
			case "EnsureBooleanOperand":
				result = converter.EnsureBooleanOperand(tt.expr, tt.code)
			case "ConvertBoolSeriesForIfStatement":
				result = converter.ConvertBoolSeriesForIfStatement(tt.expr, tt.code)
			default:
				t.Fatalf("unknown method: %s", tt.method)
			}

			if result != tt.expected {
				t.Errorf("%s\ncode=%q\nexpected: %q\ngot:      %q",
					tt.description, tt.code, tt.expected, result)
			}
		})
	}
}

func TestBooleanConverter_Integration_ComplexExpressions(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	typeSystem.RegisterVariable("bullish", "bool")
	converter := NewBooleanConverter(typeSystem)

	tests := []struct {
		name     string
		expr     ast.Expression
		code     string
		expected string
	}{
		{
			name:     "nested ternary with Series",
			expr:     &ast.Identifier{Name: "signal"},
			code:     "func() float64 { if sma_bullishSeries.GetCurrent() { return 1.0 } else { return 0.0 } }()",
			expected: "(func() float64 { if sma_bullishSeries.GetCurrent() { return 1.0 } else { return 0.0 } }() != 0)",
		},
		{
			name:     "multiple Series in expression",
			expr:     &ast.Identifier{Name: "combined"},
			code:     "aSeries.GetCurrent() + bSeries.GetCurrent()",
			expected: "(aSeries.GetCurrent() + bSeries.GetCurrent() != 0)",
		},
		{
			name:     "Series within function call",
			expr:     &ast.Identifier{Name: "result"},
			code:     "ta.sma(closeSeries.GetCurrent(), 20)",
			expected: "(ta.sma(closeSeries.GetCurrent(), 20) != 0)",
		},
		{
			name:     "empty code handled",
			expr:     &ast.Identifier{Name: "test"},
			code:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.EnsureBooleanOperand(tt.expr, tt.code)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBooleanConverter_Integration_RuleOrdering(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	typeSystem.RegisterVariable("signal", "bool")
	converter := NewBooleanConverter(typeSystem)

	tests := []struct {
		name            string
		expr            ast.Expression
		code            string
		expectedIf      string
		expectedOperand string
		description     string
	}{
		{
			name:            "comparison skips if conversion, wraps in operand context",
			expr:            &ast.Identifier{Name: "signal"},
			code:            "signalSeries.GetCurrent() > 0",
			expectedIf:      "signalSeries.GetCurrent() > 0",
			expectedOperand: "(signalSeries.GetCurrent() > 0 != 0)",
			description:     "If statement: comparison blocks conversion; Operand: Series pattern still applies",
		},
		{
			name:            "Series pattern applies before type",
			expr:            &ast.Identifier{Name: "signal"},
			code:            "signalSeries.GetCurrent()",
			expectedIf:      "signalSeries.GetCurrent() != 0",
			expectedOperand: "(signalSeries.GetCurrent() != 0)",
			description:     "Series rule takes precedence over type rule",
		},
		{
			name:            "type rule as fallback",
			expr:            &ast.Identifier{Name: "signal"},
			code:            "signal",
			expectedIf:      "signal != 0",
			expectedOperand: "(signal != 0)",
			description:     "Type rule applies when no Series pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultIf := converter.ConvertBoolSeriesForIfStatement(tt.expr, tt.code)
			if resultIf != tt.expectedIf {
				t.Errorf("%s (if statement)\nexpected: %q\ngot:      %q",
					tt.description, tt.expectedIf, resultIf)
			}

			resultOperand := converter.EnsureBooleanOperand(tt.expr, tt.code)
			if resultOperand != tt.expectedOperand {
				t.Errorf("%s (operand)\nexpected: %q\ngot:      %q",
					tt.description, tt.expectedOperand, resultOperand)
			}
		})
	}
}

func TestBooleanConverter_Integration_EdgeCases(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	converter := NewBooleanConverter(typeSystem)

	t.Run("nil expression handled gracefully", func(t *testing.T) {
		result := converter.EnsureBooleanOperand(nil, "someCode")
		expected := "someCode"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("unregistered variable with Series pattern", func(t *testing.T) {
		expr := &ast.Identifier{Name: "unknown"}
		code := "unknownSeries.GetCurrent()"
		result := converter.ConvertBoolSeriesForIfStatement(expr, code)
		expected := "unknownSeries.GetCurrent() != 0"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("unregistered variable without Series pattern", func(t *testing.T) {
		expr := &ast.Identifier{Name: "unknown"}
		code := "unknown"
		result := converter.ConvertBoolSeriesForIfStatement(expr, code)
		expected := "unknown"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("LogicalExpression is already boolean", func(t *testing.T) {
		expr := &ast.LogicalExpression{
			Operator: "&&",
			Left:     &ast.Identifier{Name: "a"},
			Right:    &ast.Identifier{Name: "b"},
		}
		code := "a && b"
		result := converter.EnsureBooleanOperand(expr, code)
		expected := "a && b"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("crossover function is already boolean", func(t *testing.T) {
		expr := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "crossover"},
			},
		}
		code := "ta.Crossover(fast, slow)"
		result := converter.EnsureBooleanOperand(expr, code)
		expected := "ta.Crossover(fast, slow)"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("crossunder function is already boolean", func(t *testing.T) {
		expr := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "crossunder"},
			},
		}
		code := "ta.Crossunder(fast, slow)"
		result := converter.EnsureBooleanOperand(expr, code)
		expected := "ta.Crossunder(fast, slow)"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}
