package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func TestStreamingBarEvaluator_ConditionalExpression(t *testing.T) {
	ctx := createTestContext()
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name     string
		expr     *ast.ConditionalExpression
		barIdx   int
		expected float64
		desc     string
	}{
		{
			name: "true_branch",
			expr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Operator: ">",
					Left:     &ast.Identifier{Name: "close"},
					Right:    &ast.Literal{Value: 100.0},
				},
				Consequent: &ast.Literal{Value: 1.0},
				Alternate:  &ast.Literal{Value: 0.0},
			},
			barIdx:   1,
			expected: 1.0,
			desc:     "condition true, returns consequent",
		},
		{
			name: "false_branch",
			expr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Operator: ">",
					Left:     &ast.Identifier{Name: "close"},
					Right:    &ast.Literal{Value: 200.0},
				},
				Consequent: &ast.Literal{Value: 1.0},
				Alternate:  &ast.Literal{Value: 0.0},
			},
			barIdx:   1,
			expected: 0.0,
			desc:     "condition false, returns alternate",
		},
		{
			name: "nested_conditional",
			expr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Operator: ">",
					Left:     &ast.Identifier{Name: "close"},
					Right:    &ast.Literal{Value: 103.0},
				},
				Consequent: &ast.ConditionalExpression{
					Test: &ast.BinaryExpression{
						Operator: ">",
						Left:     &ast.Identifier{Name: "close"},
						Right:    &ast.Literal{Value: 105.0},
					},
					Consequent: &ast.Literal{Value: 2.0},
					Alternate:  &ast.Literal{Value: 1.0},
				},
				Alternate: &ast.Literal{Value: 0.0},
			},
			barIdx:   2,
			expected: 2.0,
			desc:     "nested conditional with multiple levels",
		},
		{
			name: "expression_in_branches",
			expr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Operator: ">=",
					Left:     &ast.Identifier{Name: "close"},
					Right:    &ast.Literal{Value: 104.0},
				},
				Consequent: &ast.BinaryExpression{
					Operator: "+",
					Left:     &ast.Identifier{Name: "close"},
					Right:    &ast.Literal{Value: 10.0},
				},
				Alternate: &ast.BinaryExpression{
					Operator: "-",
					Left:     &ast.Identifier{Name: "close"},
					Right:    &ast.Literal{Value: 10.0},
				},
			},
			barIdx:   1,
			expected: 114.0,
			desc:     "expressions in both branches",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := evaluator.EvaluateAtBar(tt.expr, ctx, tt.barIdx)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tt.desc, err)
			}

			if math.Abs(value-tt.expected) > 1e-10 {
				t.Errorf("%s: expected %.2f, got %.2f", tt.desc, tt.expected, value)
			}
		})
	}
}

func TestStreamingBarEvaluator_NaNPropagation(t *testing.T) {
	ctx := createTestContext()
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name string
		expr ast.Expression
		desc string
	}{
		{
			name: "nan_addition",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Literal{Value: math.NaN()},
				Right:    &ast.Literal{Value: 5.0},
			},
			desc: "NaN + number = NaN",
		},
		{
			name: "nan_multiplication",
			expr: &ast.BinaryExpression{
				Operator: "*",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: math.NaN()},
			},
			desc: "number * NaN = NaN",
		},
		{
			name: "nan_division",
			expr: &ast.BinaryExpression{
				Operator: "/",
				Left:     &ast.Literal{Value: math.NaN()},
				Right:    &ast.Literal{Value: 2.0},
			},
			desc: "NaN / number = NaN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := evaluator.EvaluateAtBar(tt.expr, ctx, 1)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tt.desc, err)
			}

			if !math.IsNaN(value) {
				t.Errorf("%s: expected NaN, got %.2f", tt.desc, value)
			}
		})
	}
}

func TestStreamingBarEvaluator_ComparisonEdgeCases(t *testing.T) {
	ctx := createTestContext()
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name     string
		expr     *ast.BinaryExpression
		expected float64
		desc     string
	}{
		{
			name: "equal_floats",
			expr: &ast.BinaryExpression{
				Operator: "==",
				Left:     &ast.Literal{Value: 100.0},
				Right:    &ast.Literal{Value: 100.0},
			},
			expected: 1.0,
			desc:     "exact equality",
		},
		{
			name: "nearly_equal_floats",
			expr: &ast.BinaryExpression{
				Operator: "==",
				Left:     &ast.Literal{Value: 100.0},
				Right:    &ast.Literal{Value: 100.0 + 1e-11},
			},
			expected: 1.0,
			desc:     "within epsilon tolerance",
		},
		{
			name: "not_equal_outside_tolerance",
			expr: &ast.BinaryExpression{
				Operator: "==",
				Left:     &ast.Literal{Value: 100.0},
				Right:    &ast.Literal{Value: 100.0 + 1e-9},
			},
			expected: 0.0,
			desc:     "outside epsilon tolerance",
		},
		{
			name: "inequality_inverted",
			expr: &ast.BinaryExpression{
				Operator: "!=",
				Left:     &ast.Literal{Value: 100.0},
				Right:    &ast.Literal{Value: 100.0},
			},
			expected: 0.0,
			desc:     "exact equality negated",
		},
		{
			name: "zero_comparison",
			expr: &ast.BinaryExpression{
				Operator: "==",
				Left:     &ast.Literal{Value: 0.0},
				Right:    &ast.Literal{Value: -0.0},
			},
			expected: 1.0,
			desc:     "positive and negative zero equal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := evaluator.EvaluateAtBar(tt.expr, ctx, 1)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tt.desc, err)
			}

			if math.Abs(value-tt.expected) > 1e-10 {
				t.Errorf("%s: expected %.2f, got %.2f", tt.desc, tt.expected, value)
			}
		})
	}
}

func TestStreamingBarEvaluator_ComplexNestedExpressions(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100, High: 105, Low: 95},
			{Close: 110, High: 115, Low: 105},
			{Close: 120, High: 125, Low: 115},
			{Close: 130, High: 135, Low: 125},
		},
	}
	evaluator := NewStreamingBarEvaluator()

	expr := &ast.BinaryExpression{
		Operator: "/",
		Left: &ast.BinaryExpression{
			Operator: "-",
			Left:     &ast.Identifier{Name: "high"},
			Right:    &ast.Identifier{Name: "low"},
		},
		Right: &ast.Identifier{Name: "close"},
	}

	value, err := evaluator.EvaluateAtBar(expr, ctx, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := (125.0 - 115.0) / 120.0
	if math.Abs(value-expected) > 1e-10 {
		t.Errorf("expected %.6f, got %.6f", expected, value)
	}
}

func TestStreamingBarEvaluator_OperatorPrecedence(t *testing.T) {
	ctx := createTestContext()
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name     string
		expr     *ast.BinaryExpression
		expected float64
		desc     string
	}{
		{
			name: "multiplication_before_addition",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left: &ast.BinaryExpression{
					Operator: "*",
					Left:     &ast.Literal{Value: 2.0},
					Right:    &ast.Literal{Value: 3.0},
				},
				Right: &ast.Literal{Value: 4.0},
			},
			expected: 10.0,
			desc:     "(2 * 3) + 4 = 10",
		},
		{
			name: "division_before_subtraction",
			expr: &ast.BinaryExpression{
				Operator: "-",
				Left:     &ast.Literal{Value: 20.0},
				Right: &ast.BinaryExpression{
					Operator: "/",
					Left:     &ast.Literal{Value: 10.0},
					Right:    &ast.Literal{Value: 2.0},
				},
			},
			expected: 15.0,
			desc:     "20 - (10 / 2) = 15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := evaluator.EvaluateAtBar(tt.expr, ctx, 1)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tt.desc, err)
			}

			if math.Abs(value-tt.expected) > 1e-10 {
				t.Errorf("%s: expected %.2f, got %.2f", tt.desc, tt.expected, value)
			}
		})
	}
}

func TestStreamingBarEvaluator_StateIsolation(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 102},
			{Close: 104},
			{Close: 106},
			{Close: 108},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	sma2 := createTACallExpression("sma", "close", 2.0)
	sma3 := createTACallExpression("sma", "close", 3.0)

	val2_at_3, err := evaluator.EvaluateAtBar(sma2, ctx, 3)
	if err != nil {
		t.Fatalf("sma(2) at bar 3 failed: %v", err)
	}

	val3_at_3, err := evaluator.EvaluateAtBar(sma3, ctx, 3)
	if err != nil {
		t.Fatalf("sma(3) at bar 3 failed: %v", err)
	}

	if math.Abs(val2_at_3-105.0) > 1e-10 {
		t.Errorf("sma(2) at bar 3: expected 105.0, got %.2f", val2_at_3)
	}

	if math.Abs(val3_at_3-104.0) > 1e-10 {
		t.Errorf("sma(3) at bar 3: expected 104.0, got %.2f", val3_at_3)
	}

	val2_at_4, err := evaluator.EvaluateAtBar(sma2, ctx, 4)
	if err != nil {
		t.Fatalf("sma(2) at bar 4 failed: %v", err)
	}

	if math.Abs(val2_at_4-107.0) > 1e-10 {
		t.Errorf("sma(2) at bar 4: expected 107.0, got %.2f (state isolation failed)", val2_at_4)
	}
}

func TestStreamingBarEvaluator_BoundaryConditions(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
		},
	}
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name string
		expr ast.Expression
		desc string
	}{
		{
			name: "single_bar_identifier",
			expr: &ast.Identifier{Name: "close"},
			desc: "single bar context with identifier",
		},
		{
			name: "single_bar_binary",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 10.0},
			},
			desc: "single bar context with binary operation",
		},
		{
			name: "single_bar_conditional",
			expr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Operator: ">",
					Left:     &ast.Identifier{Name: "close"},
					Right:    &ast.Literal{Value: 50.0},
				},
				Consequent: &ast.Literal{Value: 1.0},
				Alternate:  &ast.Literal{Value: 0.0},
			},
			desc: "single bar context with conditional",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := evaluator.EvaluateAtBar(tt.expr, ctx, 0)
			if err != nil {
				t.Errorf("%s: should handle single bar context, got error: %v", tt.desc, err)
			}
		})
	}
}

func TestStreamingBarEvaluator_UnsupportedOperator(t *testing.T) {
	ctx := createTestContext()
	evaluator := NewStreamingBarEvaluator()

	expr := &ast.BinaryExpression{
		Operator: "**",
		Left:     &ast.Literal{Value: 2.0},
		Right:    &ast.Literal{Value: 3.0},
	}

	_, err := evaluator.EvaluateAtBar(expr, ctx, 1)
	if err == nil {
		t.Error("expected error for unsupported operator, got nil")
	}
}

func TestStreamingBarEvaluator_ConditionalWithTAFunctions(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 102},
			{Close: 104},
			{Close: 106},
			{Close: 108},
		},
	}
	evaluator := NewStreamingBarEvaluator()

	smaCall := createTACallExpression("sma", "close", 3.0)

	expr := &ast.ConditionalExpression{
		Test: &ast.BinaryExpression{
			Operator: ">",
			Left:     smaCall,
			Right:    &ast.Literal{Value: 104.0},
		},
		Consequent: &ast.Literal{Value: 1.0},
		Alternate:  &ast.Literal{Value: 0.0},
	}

	value, err := evaluator.EvaluateAtBar(expr, ctx, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != 0.0 {
		t.Errorf("expected 0.0 (sma=104.0 not > 104.0), got %.2f", value)
	}

	value, err = evaluator.EvaluateAtBar(expr, ctx, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != 1.0 {
		t.Errorf("expected 1.0 (sma=106.0 > 104.0), got %.2f", value)
	}
}

func TestStreamingBarEvaluator_LogicalOperatorsEdgeCases(t *testing.T) {
	ctx := createTestContext()
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name     string
		expr     *ast.BinaryExpression
		expected float64
		desc     string
	}{
		{
			name: "non_zero_and_non_zero",
			expr: &ast.BinaryExpression{
				Operator: "and",
				Left:     &ast.Literal{Value: 5.0},
				Right:    &ast.Literal{Value: 10.0},
			},
			expected: 1.0,
			desc:     "any non-zero values with 'and' return 1",
		},
		{
			name: "negative_and_positive",
			expr: &ast.BinaryExpression{
				Operator: "and",
				Left:     &ast.Literal{Value: -1.0},
				Right:    &ast.Literal{Value: 1.0},
			},
			expected: 1.0,
			desc:     "negative and positive both non-zero",
		},
		{
			name: "zero_or_zero",
			expr: &ast.BinaryExpression{
				Operator: "or",
				Left:     &ast.Literal{Value: 0.0},
				Right:    &ast.Literal{Value: 0.0},
			},
			expected: 0.0,
			desc:     "both zeros with 'or' return 0",
		},
		{
			name: "negative_or_zero",
			expr: &ast.BinaryExpression{
				Operator: "or",
				Left:     &ast.Literal{Value: -5.0},
				Right:    &ast.Literal{Value: 0.0},
			},
			expected: 1.0,
			desc:     "negative value is non-zero for 'or'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := evaluator.EvaluateAtBar(tt.expr, ctx, 1)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tt.desc, err)
			}

			if math.Abs(value-tt.expected) > 1e-10 {
				t.Errorf("%s: expected %.2f, got %.2f", tt.desc, tt.expected, value)
			}
		})
	}
}
