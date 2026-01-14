package validation

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestLiteralEvaluator_Numeric tests literal numeric value evaluation
func TestLiteralEvaluator_Numeric(t *testing.T) {
	tests := []struct {
		name     string
		literal  *ast.Literal
		expected float64
	}{
		{
			name:     "positive_integer",
			literal:  &ast.Literal{Value: 42},
			expected: 42.0,
		},
		{
			name:     "negative_integer",
			literal:  &ast.Literal{Value: -15},
			expected: -15.0,
		},
		{
			name:     "zero",
			literal:  &ast.Literal{Value: 0},
			expected: 0.0,
		},
		{
			name:     "positive_float",
			literal:  &ast.Literal{Value: 3.14159},
			expected: 3.14159,
		},
		{
			name:     "negative_float",
			literal:  &ast.Literal{Value: -2.5},
			expected: -2.5,
		},
		{
			name:     "large_integer",
			literal:  &ast.Literal{Value: 1260},
			expected: 1260.0,
		},
		{
			name:     "float64_directly",
			literal:  &ast.Literal{Value: float64(252.5)},
			expected: 252.5,
		},
	}

	evaluator := NewLiteralEvaluator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.Evaluate(tt.literal)

			if math.IsNaN(result) {
				t.Errorf("expected %.2f, got NaN", tt.expected)
				return
			}

			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("expected %.4f, got %.4f", tt.expected, result)
			}
		})
	}
}

// TestLiteralEvaluator_NonNumeric tests non-numeric literals return NaN
func TestLiteralEvaluator_NonNumeric(t *testing.T) {
	tests := []struct {
		name    string
		literal *ast.Literal
	}{
		{
			name:    "string_literal",
			literal: &ast.Literal{Value: "hello"},
		},
		{
			name:    "boolean_true",
			literal: &ast.Literal{Value: true},
		},
		{
			name:    "boolean_false",
			literal: &ast.Literal{Value: false},
		},
		{
			name:    "nil_value",
			literal: &ast.Literal{Value: nil},
		},
	}

	evaluator := NewLiteralEvaluator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.Evaluate(tt.literal)

			if !math.IsNaN(result) {
				t.Errorf("expected NaN for non-numeric literal, got %.2f", result)
			}
		})
	}
}

// TestConstantRegistry_Operations tests storage and retrieval
func TestConstantRegistry_Operations(t *testing.T) {
	registry := NewConstantRegistry()

	t.Run("set_and_get", func(t *testing.T) {
		registry.Set("pi", 3.14159)
		val, exists := registry.Get("pi")

		if !exists {
			t.Fatal("expected constant to exist")
		}

		if math.Abs(val-3.14159) > 0.0001 {
			t.Errorf("expected 3.14159, got %.5f", val)
		}
	})

	t.Run("get_nonexistent", func(t *testing.T) {
		_, exists := registry.Get("nonexistent")

		if exists {
			t.Error("expected constant to not exist")
		}
	})

	t.Run("overwrite_existing", func(t *testing.T) {
		registry.Set("value", 10.0)
		registry.Set("value", 20.0)

		val, _ := registry.Get("value")
		if math.Abs(val-20.0) > 0.0001 {
			t.Errorf("expected 20.0 after overwrite, got %.1f", val)
		}
	})

	t.Run("clear_all", func(t *testing.T) {
		registry.Set("a", 1.0)
		registry.Set("b", 2.0)
		registry.Clear()

		_, existsA := registry.Get("a")
		_, existsB := registry.Get("b")

		if existsA || existsB {
			t.Error("expected all constants to be cleared")
		}
	})
}

// TestIdentifierLookup_Variables tests variable resolution
func TestIdentifierLookup_Variables(t *testing.T) {
	registry := NewConstantRegistry()
	registry.Set("period", 252.0)
	registry.Set("multiplier", 5.0)
	registry.Set("zero", 0.0)

	lookup := NewIdentifierLookup(registry)

	tests := []struct {
		name       string
		identifier *ast.Identifier
		expected   float64
		shouldFail bool
	}{
		{
			name:       "existing_variable",
			identifier: &ast.Identifier{Name: "period"},
			expected:   252.0,
		},
		{
			name:       "another_variable",
			identifier: &ast.Identifier{Name: "multiplier"},
			expected:   5.0,
		},
		{
			name:       "zero_value",
			identifier: &ast.Identifier{Name: "zero"},
			expected:   0.0,
		},
		{
			name:       "nonexistent_variable",
			identifier: &ast.Identifier{Name: "unknown"},
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lookup.Resolve(tt.identifier)

			if tt.shouldFail {
				if !math.IsNaN(result) {
					t.Errorf("expected NaN for nonexistent variable, got %.2f", result)
				}
			} else {
				if math.IsNaN(result) {
					t.Errorf("expected %.2f, got NaN", tt.expected)
					return
				}
				if math.Abs(result-tt.expected) > 0.0001 {
					t.Errorf("expected %.2f, got %.2f", tt.expected, result)
				}
			}
		})
	}
}

// TestIdentifierLookup_ParserWrappedVariables tests parser quirk handling
func TestIdentifierLookup_ParserWrappedVariables(t *testing.T) {
	registry := NewConstantRegistry()
	registry.Set("wrapped", 100.0)

	lookup := NewIdentifierLookup(registry)

	wrappedExpr := &ast.MemberExpression{
		Object:   &ast.Identifier{Name: "wrapped"},
		Property: &ast.Literal{Value: 0},
		Computed: true,
	}

	result := lookup.ResolveWrappedVariable(wrappedExpr)

	if math.IsNaN(result) {
		t.Error("expected 100.0 for wrapped variable, got NaN")
		return
	}

	if math.Abs(result-100.0) > 0.0001 {
		t.Errorf("expected 100.0, got %.1f", result)
	}
}

// TestUnaryEvaluator_Operators tests unary operator evaluation
func TestUnaryEvaluator_Operators(t *testing.T) {
	registry := NewConstantRegistry()
	evaluator := NewExpressionEvaluator(registry)
	unaryEval := NewUnaryEvaluator(evaluator)

	tests := []struct {
		name     string
		operator string
		operand  ast.Expression
		expected float64
	}{
		{
			name:     "negation_positive",
			operator: "-",
			operand:  &ast.Literal{Value: 5.0},
			expected: -5.0,
		},
		{
			name:     "negation_negative",
			operator: "-",
			operand:  &ast.Literal{Value: -3.0},
			expected: 3.0,
		},
		{
			name:     "negation_zero",
			operator: "-",
			operand:  &ast.Literal{Value: 0.0},
			expected: 0.0,
		},
		{
			name:     "plus_operator",
			operator: "+",
			operand:  &ast.Literal{Value: 42.0},
			expected: 42.0,
		},
		{
			name:     "logical_not_nonzero",
			operator: "!",
			operand:  &ast.Literal{Value: 5.0},
			expected: 0.0,
		},
		{
			name:     "logical_not_zero",
			operator: "!",
			operand:  &ast.Literal{Value: 0.0},
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.UnaryExpression{
				Operator: tt.operator,
				Argument: tt.operand,
			}

			result := unaryEval.Evaluate(expr)

			if math.IsNaN(result) {
				t.Errorf("expected %.2f, got NaN", tt.expected)
				return
			}

			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("expected %.2f, got %.2f", tt.expected, result)
			}
		})
	}
}

// TestUnaryEvaluator_NestedExpressions tests nested unary operations
func TestUnaryEvaluator_NestedExpressions(t *testing.T) {
	registry := NewConstantRegistry()
	registry.Set("value", 10.0)

	evaluator := NewExpressionEvaluator(registry)
	unaryEval := NewUnaryEvaluator(evaluator)

	innerNeg := &ast.UnaryExpression{
		Operator: "-",
		Argument: &ast.Literal{Value: 10.0},
	}
	outerNeg := &ast.UnaryExpression{
		Operator: "-",
		Argument: innerNeg,
	}

	result := unaryEval.Evaluate(outerNeg)

	if math.IsNaN(result) {
		t.Error("expected 10.0 for double negation, got NaN")
		return
	}

	if math.Abs(result-10.0) > 0.0001 {
		t.Errorf("expected 10.0, got %.2f", result)
	}
}

// TestBinaryEvaluator_BasicOperations tests basic arithmetic
func TestBinaryEvaluator_BasicOperations(t *testing.T) {
	registry := NewConstantRegistry()
	evaluator := NewExpressionEvaluator(registry)
	binaryEval := NewBinaryEvaluator(evaluator)

	tests := []struct {
		name     string
		operator string
		left     float64
		right    float64
		expected float64
	}{
		{
			name:     "addition_positive",
			operator: "+",
			left:     10.0,
			right:    5.0,
			expected: 15.0,
		},
		{
			name:     "addition_negative",
			operator: "+",
			left:     -10.0,
			right:    5.0,
			expected: -5.0,
		},
		{
			name:     "subtraction",
			operator: "-",
			left:     100.0,
			right:    42.0,
			expected: 58.0,
		},
		{
			name:     "subtraction_negative_result",
			operator: "-",
			left:     10.0,
			right:    20.0,
			expected: -10.0,
		},
		{
			name:     "multiplication",
			operator: "*",
			left:     5.0,
			right:    252.0,
			expected: 1260.0,
		},
		{
			name:     "multiplication_by_zero",
			operator: "*",
			left:     42.0,
			right:    0.0,
			expected: 0.0,
		},
		{
			name:     "division",
			operator: "/",
			left:     100.0,
			right:    4.0,
			expected: 25.0,
		},
		{
			name:     "division_fractional",
			operator: "/",
			left:     5.0,
			right:    2.0,
			expected: 2.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.BinaryExpression{
				Operator: tt.operator,
				Left:     &ast.Literal{Value: tt.left},
				Right:    &ast.Literal{Value: tt.right},
			}

			result := binaryEval.Evaluate(expr)

			if math.IsNaN(result) {
				t.Errorf("expected %.2f, got NaN", tt.expected)
				return
			}

			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("expected %.4f, got %.4f", tt.expected, result)
			}
		})
	}
}

// TestBinaryEvaluator_EdgeCases tests edge cases and error conditions
func TestBinaryEvaluator_EdgeCases(t *testing.T) {
	registry := NewConstantRegistry()
	evaluator := NewExpressionEvaluator(registry)
	binaryEval := NewBinaryEvaluator(evaluator)

	tests := []struct {
		name        string
		operator    string
		left        float64
		right       float64
		expectNaN   bool
		expectValue float64
	}{
		{
			name:      "division_by_zero",
			operator:  "/",
			left:      10.0,
			right:     0.0,
			expectNaN: true,
		},
		{
			name:      "division_zero_by_zero",
			operator:  "/",
			left:      0.0,
			right:     0.0,
			expectNaN: true,
		},
		{
			name:        "division_zero_numerator",
			operator:    "/",
			left:        0.0,
			right:       5.0,
			expectValue: 0.0,
		},
		{
			name:        "large_numbers_multiplication",
			operator:    "*",
			left:        1000000.0,
			right:       1000.0,
			expectValue: 1000000000.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.BinaryExpression{
				Operator: tt.operator,
				Left:     &ast.Literal{Value: tt.left},
				Right:    &ast.Literal{Value: tt.right},
			}

			result := binaryEval.Evaluate(expr)

			if tt.expectNaN {
				if !math.IsNaN(result) {
					t.Errorf("expected NaN, got %.2f", result)
				}
			} else {
				if math.IsNaN(result) {
					t.Errorf("expected %.2f, got NaN", tt.expectValue)
					return
				}
				if math.Abs(result-tt.expectValue) > 0.0001 {
					t.Errorf("expected %.2f, got %.2f", tt.expectValue, result)
				}
			}
		})
	}
}

// TestBinaryEvaluator_OperatorPrecedence tests precedence through nested expressions
func TestBinaryEvaluator_OperatorPrecedence(t *testing.T) {
	registry := NewConstantRegistry()
	evaluator := NewExpressionEvaluator(registry)

	expr := &ast.BinaryExpression{
		Operator: "*",
		Left: &ast.BinaryExpression{
			Operator: "+",
			Left:     &ast.Literal{Value: 10.0},
			Right:    &ast.Literal{Value: 5.0},
		},
		Right: &ast.Literal{Value: 2.0},
	}

	result := evaluator.Evaluate(expr)

	if math.IsNaN(result) {
		t.Error("expected 30.0, got NaN")
		return
	}

	if math.Abs(result-30.0) > 0.0001 {
		t.Errorf("expected 30.0, got %.2f", result)
	}
}

// TestMathFunctionEvaluator_Functions tests math library functions
func TestMathFunctionEvaluator_Functions(t *testing.T) {
	registry := NewConstantRegistry()
	evaluator := NewExpressionEvaluator(registry)
	mathEval := NewMathFunctionEvaluator(evaluator)

	tests := []struct {
		name     string
		function string
		args     []ast.Expression
		expected float64
	}{
		{
			name:     "pow_positive_exponent",
			function: "math.pow",
			args: []ast.Expression{
				&ast.Literal{Value: 2.0},
				&ast.Literal{Value: 3.0},
			},
			expected: 8.0,
		},
		{
			name:     "pow_zero_exponent",
			function: "math.pow",
			args: []ast.Expression{
				&ast.Literal{Value: 5.0},
				&ast.Literal{Value: 0.0},
			},
			expected: 1.0,
		},
		{
			name:     "pow_fractional_exponent",
			function: "math.pow",
			args: []ast.Expression{
				&ast.Literal{Value: 4.0},
				&ast.Literal{Value: 0.5},
			},
			expected: 2.0,
		},
		{
			name:     "sqrt_perfect_square",
			function: "math.sqrt",
			args: []ast.Expression{
				&ast.Literal{Value: 16.0},
			},
			expected: 4.0,
		},
		{
			name:     "sqrt_non_perfect",
			function: "math.sqrt",
			args: []ast.Expression{
				&ast.Literal{Value: 2.0},
			},
			expected: 1.41421356,
		},
		{
			name:     "round_up",
			function: "math.round",
			args: []ast.Expression{
				&ast.Literal{Value: 3.6},
			},
			expected: 4.0,
		},
		{
			name:     "round_down",
			function: "math.round",
			args: []ast.Expression{
				&ast.Literal{Value: 3.4},
			},
			expected: 3.0,
		},
		{
			name:     "round_half",
			function: "math.round",
			args: []ast.Expression{
				&ast.Literal{Value: 3.5},
			},
			expected: 4.0,
		},
		{
			name:     "floor_positive",
			function: "math.floor",
			args: []ast.Expression{
				&ast.Literal{Value: 3.9},
			},
			expected: 3.0,
		},
		{
			name:     "floor_negative",
			function: "math.floor",
			args: []ast.Expression{
				&ast.Literal{Value: -2.1},
			},
			expected: -3.0,
		},
		{
			name:     "ceil_positive",
			function: "math.ceil",
			args: []ast.Expression{
				&ast.Literal{Value: 3.1},
			},
			expected: 4.0,
		},
		{
			name:     "ceil_negative",
			function: "math.ceil",
			args: []ast.Expression{
				&ast.Literal{Value: -2.9},
			},
			expected: -2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callExpr := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: tt.function[5:]}, // Remove "math." prefix
				},
				Arguments: tt.args,
			}

			result := mathEval.Evaluate(callExpr)

			if math.IsNaN(result) {
				t.Errorf("expected %.5f, got NaN", tt.expected)
				return
			}

			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("expected %.5f, got %.5f", tt.expected, result)
			}
		})
	}
}

// TestMathFunctionEvaluator_EdgeCases tests math function edge cases
func TestMathFunctionEvaluator_EdgeCases(t *testing.T) {
	registry := NewConstantRegistry()
	evaluator := NewExpressionEvaluator(registry)
	mathEval := NewMathFunctionEvaluator(evaluator)

	tests := []struct {
		name      string
		function  string
		args      []ast.Expression
		expectNaN bool
	}{
		{
			name:     "sqrt_negative",
			function: "sqrt",
			args: []ast.Expression{
				&ast.Literal{Value: -1.0},
			},
			expectNaN: true,
		},
		{
			name:     "pow_invalid_args_count",
			function: "pow",
			args: []ast.Expression{
				&ast.Literal{Value: 2.0},
			},
			expectNaN: true,
		},
		{
			name:     "sqrt_invalid_args_count",
			function: "sqrt",
			args: []ast.Expression{
				&ast.Literal{Value: 2.0},
				&ast.Literal{Value: 3.0},
			},
			expectNaN: true,
		},
		{
			name:      "unknown_function",
			function:  "unknown",
			args:      []ast.Expression{},
			expectNaN: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callExpr := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: tt.function},
				},
				Arguments: tt.args,
			}

			result := mathEval.Evaluate(callExpr)

			if !tt.expectNaN {
				t.Fatal("test configuration error: expectNaN should be true")
			}

			if !math.IsNaN(result) {
				t.Errorf("expected NaN for invalid operation, got %.2f", result)
			}
		})
	}

	t.Run("plain_identifier_callee", func(t *testing.T) {
		callExpr := &ast.CallExpression{
			Callee: &ast.Identifier{Name: "plainFunc"},
			Arguments: []ast.Expression{
				&ast.Literal{Value: 2.0},
			},
		}

		result := mathEval.Evaluate(callExpr)

		if !math.IsNaN(result) {
			t.Errorf("expected NaN for plain identifier function, got %.2f", result)
		}
	})
}

// TestExpressionEvaluator_ComplexNestedExpressions tests integration
func TestExpressionEvaluator_ComplexNestedExpressions(t *testing.T) {
	registry := NewConstantRegistry()
	registry.Set("base", 10.0)
	registry.Set("multiplier", 5.0)

	evaluator := NewExpressionEvaluator(registry)

	tests := []struct {
		name     string
		expr     ast.Expression
		expected float64
	}{
		{
			name: "variable_multiplication_and_addition",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left: &ast.BinaryExpression{
					Operator: "*",
					Left:     &ast.Identifier{Name: "base"},
					Right:    &ast.Identifier{Name: "multiplier"},
				},
				Right: &ast.Literal{Value: 200.0},
			},
			expected: 250.0,
		},
		{
			name: "negation_of_multiplication",
			expr: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.BinaryExpression{
					Operator: "*",
					Left:     &ast.Identifier{Name: "base"},
					Right:    &ast.Identifier{Name: "multiplier"},
				},
			},
			expected: -50.0,
		},
		{
			name: "division_with_addition",
			expr: &ast.BinaryExpression{
				Operator: "/",
				Left: &ast.BinaryExpression{
					Operator: "+",
					Left:     &ast.Identifier{Name: "base"},
					Right:    &ast.Literal{Value: 5.0},
				},
				Right: &ast.Identifier{Name: "multiplier"},
			},
			expected: 3.0,
		},
		{
			name: "math_pow_with_variables",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: "pow"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "base"},
					&ast.Literal{Value: 2.0},
				},
			},
			expected: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.Evaluate(tt.expr)

			if math.IsNaN(result) {
				t.Errorf("expected %.2f, got NaN", tt.expected)
				return
			}

			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("expected %.2f, got %.2f", tt.expected, result)
			}
		})
	}
}

// TestExpressionEvaluator_NaNPropagation tests NaN propagates correctly
func TestExpressionEvaluator_NaNPropagation(t *testing.T) {
	registry := NewConstantRegistry()
	evaluator := NewExpressionEvaluator(registry)

	tests := []struct {
		name string
		expr ast.Expression
	}{
		{
			name: "addition_with_nonexistent_variable",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Identifier{Name: "nonexistent"},
				Right:    &ast.Literal{Value: 10.0},
			},
		},
		{
			name: "multiplication_with_NaN",
			expr: &ast.BinaryExpression{
				Operator: "*",
				Left: &ast.BinaryExpression{
					Operator: "/",
					Left:     &ast.Literal{Value: 1.0},
					Right:    &ast.Literal{Value: 0.0},
				},
				Right: &ast.Literal{Value: 5.0},
			},
		},
		{
			name: "unknown_expression_type",
			expr: &ast.ConditionalExpression{
				Test:       &ast.Literal{Value: true},
				Consequent: &ast.Literal{Value: 1.0},
				Alternate:  &ast.Literal{Value: 0.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.Evaluate(tt.expr)

			if !math.IsNaN(result) {
				t.Errorf("expected NaN propagation, got %.2f", result)
			}
		})
	}
}

// TestExpressionEvaluator_RealWorldScenarios tests realistic use cases
func TestExpressionEvaluator_RealWorldScenarios(t *testing.T) {
	registry := NewConstantRegistry()
	registry.Set("rightBars", 15.0)
	registry.Set("period", 252.0)
	registry.Set("years", 5.0)

	evaluator := NewExpressionEvaluator(registry)

	tests := []struct {
		name     string
		expr     ast.Expression
		expected float64
		desc     string
	}{
		{
			name: "plot_offset_calculation",
			expr: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.BinaryExpression{
					Operator: "+",
					Left:     &ast.Identifier{Name: "rightBars"},
					Right:    &ast.Literal{Value: 1.0},
				},
			},
			expected: -16.0,
			desc:     "Plot offset for drawing ahead of current bar",
		},
		{
			name: "lookback_period_calculation",
			expr: &ast.BinaryExpression{
				Operator: "*",
				Left:     &ast.Identifier{Name: "period"},
				Right:    &ast.Identifier{Name: "years"},
			},
			expected: 1260.0,
			desc:     "5-year lookback in trading days",
		},
		{
			name: "moving_average_period",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: "round"},
				},
				Arguments: []ast.Expression{
					&ast.BinaryExpression{
						Operator: "/",
						Left:     &ast.Identifier{Name: "period"},
						Right:    &ast.Literal{Value: 12.0},
					},
				},
			},
			expected: 21.0,
			desc:     "Monthly period from annual trading days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.Evaluate(tt.expr)

			if math.IsNaN(result) {
				t.Errorf("%s: expected %.2f, got NaN", tt.desc, tt.expected)
				return
			}

			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("%s: expected %.2f, got %.2f", tt.desc, tt.expected, result)
			}
		})
	}
}
