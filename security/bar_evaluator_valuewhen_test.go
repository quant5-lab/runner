package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

/* TestStreamingBarEvaluator_ValuewhenOccurrenceSelection verifies occurrence-based lookback */
func TestStreamingBarEvaluator_ValuewhenOccurrenceSelection(t *testing.T) {
	data := []context.OHLCV{
		{Close: 100.0, High: 105.0},
		{Close: 103.0, High: 108.0},
		{Close: 101.0, High: 106.0},
		{Close: 104.0, High: 109.0},
		{Close: 105.0, High: 110.0},
	}

	ctx := &context.Context{Data: data}
	evaluator := NewStreamingBarEvaluator()

	conditionExpr := &ast.BinaryExpression{
		Operator: ">",
		Left:     &ast.Identifier{Name: "close"},
		Right:    &ast.Literal{Value: 102.0},
	}

	tests := []struct {
		name       string
		occurrence int
		barIdx     int
		expected   float64
		desc       string
	}{
		{"most_recent", 0, 4, 110.0, "occurrence=0 returns current bar (most recent match)"},
		{"second_recent", 1, 4, 109.0, "occurrence=1 returns 2nd most recent match"},
		{"third_recent", 2, 4, 108.0, "occurrence=2 returns 3rd most recent match"},
		{"earlier_bar_context", 0, 3, 109.0, "at bar 3, occurrence=0 returns bar 3"},
		{"earlier_bar_second", 1, 3, 108.0, "at bar 3, occurrence=1 returns bar 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valuewhenCall := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "valuewhen"},
				},
				Arguments: []ast.Expression{
					conditionExpr,
					&ast.Identifier{Name: "high"},
					&ast.Literal{Value: float64(tt.occurrence)},
				},
			}

			result, err := evaluator.EvaluateAtBar(valuewhenCall, ctx, tt.barIdx)
			if err != nil {
				t.Fatalf("%s: EvaluateAtBar failed: %v", tt.desc, err)
			}

			if result != tt.expected {
				t.Errorf("%s: expected %.2f, got %.2f", tt.desc, tt.expected, result)
			}
		})
	}
}

/* TestStreamingBarEvaluator_ValuewhenBoundaryConditions verifies edge cases */
func TestStreamingBarEvaluator_ValuewhenBoundaryConditions(t *testing.T) {
	data := []context.OHLCV{
		{Close: 100.0, High: 105.0},
		{Close: 101.0, High: 106.0},
		{Close: 102.0, High: 107.0},
		{Close: 104.0, High: 109.0},
	}

	ctx := &context.Context{Data: data}
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name       string
		condition  ast.Expression
		occurrence int
		barIdx     int
		expectNaN  bool
		desc       string
	}{
		{
			name: "no_matches",
			condition: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 200.0},
			},
			occurrence: 0,
			barIdx:     3,
			expectNaN:  true,
			desc:       "condition never true in entire history",
		},
		{
			name: "occurrence_beyond_available",
			condition: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 102.0},
			},
			occurrence: 10,
			barIdx:     3,
			expectNaN:  true,
			desc:       "occurrence exceeds match count",
		},
		{
			name: "exact_match_count",
			condition: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 102.0},
			},
			occurrence: 1,
			barIdx:     3,
			expectNaN:  true,
			desc:       "only 1 match exists (bar 3), requesting 2nd",
		},
		{
			name: "valid_at_boundary",
			condition: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 102.0},
			},
			occurrence: 0,
			barIdx:     3,
			expectNaN:  false,
			desc:       "1 match exists, requesting 1st is valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valuewhenCall := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "valuewhen"},
				},
				Arguments: []ast.Expression{
					tt.condition,
					&ast.Identifier{Name: "high"},
					&ast.Literal{Value: float64(tt.occurrence)},
				},
			}

			result, err := evaluator.EvaluateAtBar(valuewhenCall, ctx, tt.barIdx)
			if err != nil {
				t.Fatalf("%s: EvaluateAtBar failed: %v", tt.desc, err)
			}

			isNaN := math.IsNaN(result)
			if isNaN != tt.expectNaN {
				t.Errorf("%s: expectNaN=%v, got isNaN=%v (result=%.2f)",
					tt.desc, tt.expectNaN, isNaN, result)
			}
		})
	}
}

/* TestStreamingBarEvaluator_ValuewhenComplexExpressions verifies expression support */
func TestStreamingBarEvaluator_ValuewhenComplexExpressions(t *testing.T) {
	data := []context.OHLCV{
		{Close: 100.0, High: 105.0, Low: 95.0},
		{Close: 103.0, High: 108.0, Low: 98.0},
		{Close: 101.0, High: 106.0, Low: 96.0},
		{Close: 104.0, High: 109.0, Low: 99.0},
	}

	ctx := &context.Context{Data: data}
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name      string
		condition ast.Expression
		source    ast.Expression
		barIdx    int
		expected  float64
		desc      string
	}{
		{
			name: "binary_expression_source",
			condition: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 102.0},
			},
			source: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Identifier{Name: "high"},
				Right:    &ast.Identifier{Name: "low"},
			},
			barIdx:   3,
			expected: 109.0 + 99.0,
			desc:     "source expression with arithmetic",
		},
		{
			name: "complex_condition",
			condition: &ast.BinaryExpression{
				Operator: ">=",
				Left: &ast.BinaryExpression{
					Operator: "+",
					Left:     &ast.Identifier{Name: "close"},
					Right:    &ast.Literal{Value: 1.0},
				},
				Right: &ast.Literal{Value: 104.0},
			},
			source:   &ast.Identifier{Name: "high"},
			barIdx:   3,
			expected: 109.0,
			desc:     "condition with arithmetic expression",
		},
		{
			name: "source_arithmetic_division",
			condition: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 102.0},
			},
			source: &ast.BinaryExpression{
				Operator: "/",
				Left:     &ast.Identifier{Name: "high"},
				Right:    &ast.Literal{Value: 2.0},
			},
			barIdx:   3,
			expected: 109.0 / 2.0,
			desc:     "source with division operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valuewhenCall := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "valuewhen"},
				},
				Arguments: []ast.Expression{
					tt.condition,
					tt.source,
					&ast.Literal{Value: 0.0},
				},
			}

			result, err := evaluator.EvaluateAtBar(valuewhenCall, ctx, tt.barIdx)
			if err != nil {
				t.Fatalf("%s: EvaluateAtBar failed: %v", tt.desc, err)
			}

			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("%s: expected %.2f, got %.2f", tt.desc, tt.expected, result)
			}
		})
	}
}

/* TestStreamingBarEvaluator_ValuewhenConditionTypes verifies condition expression handling */
func TestStreamingBarEvaluator_ValuewhenConditionTypes(t *testing.T) {
	data := []context.OHLCV{
		{Close: 100.0, High: 105.0},
		{Close: 103.0, High: 108.0},
		{Close: 101.0, High: 106.0},
		{Close: 104.0, High: 109.0},
		{Close: 105.0, High: 110.0},
	}

	ctx := &context.Context{Data: data}
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name      string
		condition ast.Expression
		expected  float64
		desc      string
	}{
		{
			name: "greater_than",
			condition: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 103.5},
			},
			expected: 110.0,
			desc:     "condition with > operator",
		},
		{
			name: "less_than",
			condition: &ast.BinaryExpression{
				Operator: "<",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 102.0},
			},
			expected: 106.0,
			desc:     "condition with < operator",
		},
		{
			name: "equality",
			condition: &ast.BinaryExpression{
				Operator: "==",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 104.0},
			},
			expected: 109.0,
			desc:     "condition with == operator",
		},
		{
			name: "greater_equal",
			condition: &ast.BinaryExpression{
				Operator: ">=",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 104.0},
			},
			expected: 110.0,
			desc:     "condition with >= operator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valuewhenCall := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "valuewhen"},
				},
				Arguments: []ast.Expression{
					tt.condition,
					&ast.Identifier{Name: "high"},
					&ast.Literal{Value: 0.0},
				},
			}

			result, err := evaluator.EvaluateAtBar(valuewhenCall, ctx, 4)
			if err != nil {
				t.Fatalf("%s: EvaluateAtBar failed: %v", tt.desc, err)
			}

			if result != tt.expected {
				t.Errorf("%s: expected %.2f, got %.2f", tt.desc, tt.expected, result)
			}
		})
	}
}

/* TestStreamingBarEvaluator_ValuewhenArgumentValidation verifies error handling */
func TestStreamingBarEvaluator_ValuewhenArgumentValidation(t *testing.T) {
	ctx := &context.Context{Data: []context.OHLCV{{Close: 100.0, High: 105.0}}}
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name string
		call *ast.CallExpression
		desc string
	}{
		{
			name: "insufficient_arguments",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "valuewhen"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Identifier{Name: "high"},
				},
			},
			desc: "missing occurrence argument",
		},
		{
			name: "non_literal_occurrence",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "valuewhen"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Identifier{Name: "high"},
					&ast.Identifier{Name: "somevar"},
				},
			},
			desc: "occurrence must be literal",
		},
		{
			name: "zero_arguments",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "valuewhen"},
				},
				Arguments: []ast.Expression{},
			},
			desc: "no arguments provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := evaluator.EvaluateAtBar(tt.call, ctx, 0)
			if err == nil {
				t.Errorf("%s: expected error, got nil", tt.desc)
			}
		})
	}
}

/* TestStreamingBarEvaluator_ValuewhenBarProgression verifies behavior across bars */
func TestStreamingBarEvaluator_ValuewhenBarProgression(t *testing.T) {
	data := []context.OHLCV{
		{Close: 100.0, High: 105.0},
		{Close: 103.0, High: 108.0},
		{Close: 101.0, High: 106.0},
		{Close: 104.0, High: 109.0},
		{Close: 105.0, High: 110.0},
		{Close: 106.0, High: 111.0},
	}

	ctx := &context.Context{Data: data}
	evaluator := NewStreamingBarEvaluator()

	conditionExpr := &ast.BinaryExpression{
		Operator: ">",
		Left:     &ast.Identifier{Name: "close"},
		Right:    &ast.Literal{Value: 102.0},
	}

	tests := []struct {
		barIdx   int
		expected float64
		desc     string
	}{
		{1, 108.0, "bar 1: first match, returns self"},
		{3, 109.0, "bar 3: most recent match is bar 3"},
		{5, 111.0, "bar 5: most recent match is bar 5"},
		{2, 108.0, "bar 2: no match at bar 2, returns bar 1"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			valuewhenCall := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "valuewhen"},
				},
				Arguments: []ast.Expression{
					conditionExpr,
					&ast.Identifier{Name: "high"},
					&ast.Literal{Value: 0.0},
				},
			}

			result, err := evaluator.EvaluateAtBar(valuewhenCall, ctx, tt.barIdx)
			if err != nil {
				t.Fatalf("bar %d: EvaluateAtBar failed: %v", tt.barIdx, err)
			}

			if result != tt.expected {
				t.Errorf("bar %d: expected %.2f, got %.2f", tt.barIdx, tt.expected, result)
			}
		})
	}
}

/* TestStreamingBarEvaluator_ValuewhenStateIsolation verifies independent evaluation */
func TestStreamingBarEvaluator_ValuewhenStateIsolation(t *testing.T) {
	data := []context.OHLCV{
		{Close: 100.0, High: 105.0},
		{Close: 103.0, High: 108.0},
		{Close: 104.0, High: 109.0},
	}

	ctx := &context.Context{Data: data}
	evaluator := NewStreamingBarEvaluator()

	conditionExpr := &ast.BinaryExpression{
		Operator: ">",
		Left:     &ast.Identifier{Name: "close"},
		Right:    &ast.Literal{Value: 102.0},
	}

	valuewhenCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "valuewhen"},
		},
		Arguments: []ast.Expression{
			conditionExpr,
			&ast.Identifier{Name: "high"},
			&ast.Literal{Value: 0.0},
		},
	}

	result1, err1 := evaluator.EvaluateAtBar(valuewhenCall, ctx, 2)
	result2, err2 := evaluator.EvaluateAtBar(valuewhenCall, ctx, 2)

	if err1 != nil || err2 != nil {
		t.Fatalf("EvaluateAtBar failed: err1=%v, err2=%v", err1, err2)
	}

	if result1 != result2 {
		t.Errorf("state isolation failed: first=%.2f, second=%.2f", result1, result2)
	}

	result3, err3 := evaluator.EvaluateAtBar(valuewhenCall, ctx, 1)
	if err3 != nil {
		t.Fatalf("EvaluateAtBar at bar 1 failed: %v", err3)
	}

	if result1 == result3 {
		t.Errorf("expected different results for different bars, got %.2f for both", result1)
	}
}
