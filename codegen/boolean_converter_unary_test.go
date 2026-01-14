package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestBooleanConverter_UnaryExpression validates UnaryExpression boolean recognition */
func TestBooleanConverter_UnaryExpression(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	converter := NewBooleanConverter(typeSystem)

	tests := []struct {
		name       string
		expr       ast.Expression
		wantIsBool bool
	}{
		{
			name: "not operator produces boolean",
			expr: &ast.UnaryExpression{
				Operator: "not",
				Argument: &ast.Identifier{Name: "condition"},
			},
			wantIsBool: true,
		},
		{
			name: "exclamation operator produces boolean",
			expr: &ast.UnaryExpression{
				Operator: "!",
				Argument: &ast.Identifier{Name: "enabled"},
			},
			wantIsBool: true,
		},
		{
			name: "negation operator does not produce boolean",
			expr: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.Identifier{Name: "value"},
			},
			wantIsBool: false,
		},
		{
			name: "positive operator does not produce boolean",
			expr: &ast.UnaryExpression{
				Operator: "+",
				Argument: &ast.Identifier{Name: "delta"},
			},
			wantIsBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.IsAlreadyBoolean(tt.expr)
			if result != tt.wantIsBool {
				t.Errorf("IsAlreadyBoolean() = %v, want %v", result, tt.wantIsBool)
			}
		})
	}
}

/* TestBooleanConverter_NaFunction validates na() function boolean recognition */
func TestBooleanConverter_NaFunction(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	converter := NewBooleanConverter(typeSystem)

	tests := []struct {
		name       string
		expr       ast.Expression
		wantIsBool bool
	}{
		{
			name: "na() function produces boolean",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "na"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			wantIsBool: true,
		},
		{
			name: "sma() function does not produce boolean",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "sma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 20.0},
				},
			},
			wantIsBool: false,
		},
		{
			name: "ta.crossover produces boolean",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossover"},
				},
			},
			wantIsBool: true,
		},
		{
			name: "ta.crossunder produces boolean",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossunder"},
				},
			},
			wantIsBool: true,
		},
		{
			name: "ta.sma does not produce boolean",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
			},
			wantIsBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.IsAlreadyBoolean(tt.expr)
			if result != tt.wantIsBool {
				t.Errorf("IsAlreadyBoolean() = %v, want %v", result, tt.wantIsBool)
			}
		})
	}
}

/* TestBooleanConverter_UnaryWithSeries validates Series handling in unary expressions */
func TestBooleanConverter_UnaryWithSeries(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	typeSystem.RegisterVariable("enabled", "bool")
	typeSystem.RegisterVariable("value", "float64")
	converter := NewBooleanConverter(typeSystem)

	tests := []struct {
		name     string
		code     string
		expr     ast.Expression
		expected string
	}{
		{
			name: "not with Series requires boolean conversion",
			code: "enabledSeries.GetCurrent()",
			expr: &ast.UnaryExpression{
				Operator: "not",
				Argument: &ast.Identifier{Name: "enabled"},
			},
			expected: "(value.IsTrue(enabledSeries.GetCurrent()))",
		},
		{
			name: "not with comparison unchanged",
			code: "(close > open)",
			expr: &ast.UnaryExpression{
				Operator: "not",
				Argument: &ast.BinaryExpression{
					Operator: ">",
					Left:     &ast.Identifier{Name: "close"},
					Right:    &ast.Identifier{Name: "open"},
				},
			},
			expected: "(close > open)",
		},
		{
			name: "not with na() unchanged",
			code: "math.IsNaN(bar.Close)",
			expr: &ast.UnaryExpression{
				Operator: "not",
				Argument: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "na"},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
					},
				},
			},
			expected: "math.IsNaN(bar.Close)",
		},
		{
			name: "negation with Series still needs boolean check for unary context",
			code: "valueSeries.GetCurrent()",
			expr: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.Identifier{Name: "value"},
			},
			expected: "(value.IsTrue(valueSeries.GetCurrent()))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.EnsureBooleanOperand(tt.expr.(*ast.UnaryExpression).Argument, tt.code)
			if result != tt.expected {
				t.Errorf("EnsureBooleanOperand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

/* TestBooleanConverter_NestedUnary validates nested unary expression handling */
func TestBooleanConverter_NestedUnary(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	converter := NewBooleanConverter(typeSystem)

	tests := []struct {
		name       string
		expr       ast.Expression
		wantIsBool bool
	}{
		{
			name: "double not",
			expr: &ast.UnaryExpression{
				Operator: "not",
				Argument: &ast.UnaryExpression{
					Operator: "not",
					Argument: &ast.Identifier{Name: "condition"},
				},
			},
			wantIsBool: true,
		},
		{
			name: "not with negation argument",
			expr: &ast.UnaryExpression{
				Operator: "not",
				Argument: &ast.UnaryExpression{
					Operator: "-",
					Argument: &ast.Identifier{Name: "value"},
				},
			},
			wantIsBool: true,
		},
		{
			name: "negation of boolean expression",
			expr: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.BinaryExpression{
					Operator: ">",
					Left:     &ast.Identifier{Name: "a"},
					Right:    &ast.Identifier{Name: "b"},
				},
			},
			wantIsBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.IsAlreadyBoolean(tt.expr)
			if result != tt.wantIsBool {
				t.Errorf("IsAlreadyBoolean() = %v, want %v", result, tt.wantIsBool)
			}
		})
	}
}

/* TestBooleanConverter_EdgeCases_Unary validates unary expression edge cases */
func TestBooleanConverter_EdgeCases_Unary(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	converter := NewBooleanConverter(typeSystem)

	t.Run("nil unary expression", func(t *testing.T) {
		result := converter.IsAlreadyBoolean(nil)
		if result {
			t.Error("Expected false for nil expression")
		}
	})

	t.Run("unary with nil argument", func(t *testing.T) {
		expr := &ast.UnaryExpression{
			Operator: "not",
			Argument: nil,
		}
		result := converter.IsAlreadyBoolean(expr)
		if !result {
			t.Error("Expected true for 'not' operator regardless of argument")
		}
	})

	t.Run("empty operator string", func(t *testing.T) {
		expr := &ast.UnaryExpression{
			Operator: "",
			Argument: &ast.Identifier{Name: "test"},
		}
		result := converter.IsAlreadyBoolean(expr)
		if result {
			t.Error("Expected false for empty operator")
		}
	})

	t.Run("unknown operator", func(t *testing.T) {
		expr := &ast.UnaryExpression{
			Operator: "~",
			Argument: &ast.Identifier{Name: "bits"},
		}
		result := converter.IsAlreadyBoolean(expr)
		if result {
			t.Error("Expected false for unknown operator '~'")
		}
	})
}

/* TestBooleanConverter_Integration_UnaryInLogical validates unary in logical expressions */
func TestBooleanConverter_Integration_UnaryInLogical(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	typeSystem.RegisterVariable("has_trade", "bool")
	typeSystem.RegisterVariable("buy_signal", "bool")
	converter := NewBooleanConverter(typeSystem)

	tests := []struct {
		name       string
		expr       ast.Expression
		exprCode   string
		wantIsBool bool
	}{
		{
			name: "not X produces boolean",
			expr: &ast.UnaryExpression{
				Operator: "not",
				Argument: &ast.Identifier{Name: "has_trade"},
			},
			exprCode:   "!(has_tradeSeries.GetCurrent() != 0)",
			wantIsBool: true,
		},
		{
			name:       "identifier in logical context needs conversion",
			expr:       &ast.Identifier{Name: "buy_signal"},
			exprCode:   "buy_signalSeries.GetCurrent()",
			wantIsBool: false,
		},
		{
			name: "not X and Y - both operands need boolean context",
			expr: &ast.LogicalExpression{
				Operator: "&&",
				Left: &ast.UnaryExpression{
					Operator: "not",
					Argument: &ast.Identifier{Name: "has_trade"},
				},
				Right: &ast.Identifier{Name: "buy_signal"},
			},
			exprCode:   "!(has_tradeSeries.GetCurrent() != 0) && (buy_signalSeries.GetCurrent() != 0)",
			wantIsBool: true,
		},
		{
			name: "X or not Y - logical expression produces boolean",
			expr: &ast.LogicalExpression{
				Operator: "||",
				Left:     &ast.Identifier{Name: "has_trade"},
				Right: &ast.UnaryExpression{
					Operator: "not",
					Argument: &ast.Identifier{Name: "buy_signal"},
				},
			},
			exprCode:   "(has_tradeSeries.GetCurrent() != 0) || !(buy_signalSeries.GetCurrent() != 0)",
			wantIsBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.IsAlreadyBoolean(tt.expr)
			if result != tt.wantIsBool {
				t.Errorf("IsAlreadyBoolean() = %v, want %v", result, tt.wantIsBool)
			}
		})
	}
}

/* TestBooleanConverter_UnaryOperatorCoverage ensures all operators handled */
func TestBooleanConverter_UnaryOperatorCoverage(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	converter := NewBooleanConverter(typeSystem)

	operators := []struct {
		op         string
		shouldBool bool
	}{
		{"not", true},
		{"!", true},
		{"-", false},
		{"+", false},
		{"~", false},
	}

	for _, op := range operators {
		t.Run("operator_"+op.op, func(t *testing.T) {
			expr := &ast.UnaryExpression{
				Operator: op.op,
				Argument: &ast.Identifier{Name: "x"},
			}

			result := converter.IsAlreadyBoolean(expr)
			if result != op.shouldBool {
				t.Errorf("Operator %q: IsAlreadyBoolean() = %v, want %v", op.op, result, op.shouldBool)
			}
		})
	}
}

/* TestBooleanConverter_CodegenIntegration validates generated code patterns */
func TestBooleanConverter_CodegenIntegration(t *testing.T) {
	t.Run("not generates negation without extra != 0", func(t *testing.T) {
		typeSystem := NewTypeInferenceEngine()
		typeSystem.RegisterVariable("enabled", "bool")
		converter := NewBooleanConverter(typeSystem)

		expr := &ast.UnaryExpression{
			Operator: "not",
			Argument: &ast.Identifier{Name: "enabled"},
		}

		code := "!(enabledSeries.GetCurrent() != 0)"

		result := converter.ConvertBoolSeriesForIfStatement(expr, code)

		if result != code {
			t.Errorf("Expected unchanged code %q, got %q", code, result)
		}

		if strings.Count(result, "!= 0") > 1 {
			t.Error("Double boolean conversion detected")
		}
	})

	t.Run("na() generates IsNaN without != 0", func(t *testing.T) {
		typeSystem := NewTypeInferenceEngine()
		converter := NewBooleanConverter(typeSystem)

		expr := &ast.CallExpression{
			Callee: &ast.Identifier{Name: "na"},
		}

		code := "math.IsNaN(bar.Close)"

		result := converter.ConvertBoolSeriesForIfStatement(expr, code)

		if result != code {
			t.Errorf("Expected unchanged code %q, got %q", code, result)
		}

		if strings.Contains(result, "!= 0") {
			t.Error("Unexpected boolean conversion for IsNaN")
		}
	})
}
