package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

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

// TestStreamingBarEvaluator_ValuewhenCacheIsolation verifies that two valuewhen
// calls with distinct condition expressions maintain independent cache entries and
// independent state — a different condition threshold must never bleed into another
// valuewhen's match history.
func TestStreamingBarEvaluator_ValuewhenCacheIsolation(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100.0, High: 105.0},
			{Close: 103.0, High: 108.0},
			{Close: 101.0, High: 106.0},
			{Close: 104.0, High: 109.0},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	makecall := func(threshold float64) *ast.CallExpression {
		return &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "valuewhen"},
			},
			Arguments: []ast.Expression{
				&ast.BinaryExpression{
					Operator: ">",
					Left:     &ast.Identifier{Name: "close"},
					Right:    &ast.Literal{Value: threshold},
				},
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: 0.0},
			},
		}
	}

	call1 := makecall(102.0)
	call2 := makecall(103.0)

	r1, err := evaluator.EvaluateAtBar(call1, ctx, 2)
	if err != nil {
		t.Fatalf("call1 bar 2: %v", err)
	}
	r2, err := evaluator.EvaluateAtBar(call2, ctx, 2)
	if err != nil {
		t.Fatalf("call2 bar 2: %v", err)
	}

	if got := len(evaluator.valuewhenCache); got != 2 {
		t.Errorf("expected 2 independent cache entries, got %d", got)
	}
	if math.IsNaN(r1) || r1 != 108.0 {
		t.Errorf("call1 bar 2: expected 108.0 (carry-forward from bar 1), got %.4f", r1)
	}
	if !math.IsNaN(r2) {
		t.Errorf("call2 bar 2: expected NaN (no match before bar 3), got %.4f", r2)
	}
}

// TestStreamingBarEvaluator_ValuewhenCachesStateAcrossBars verifies that a single
// ValuewhenStateManager is created and reused across all calls sharing the same
// condition, source, and occurrence — confirming cache key correctness and
// ForwardSeriesBuffer historical access via the evaluator dispatch path.
func TestStreamingBarEvaluator_ValuewhenCachesStateAcrossBars(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100.0, High: 105.0},
			{Close: 103.0, High: 108.0},
			{Close: 101.0, High: 106.0},
			{Close: 104.0, High: 109.0},
			{Close: 105.0, High: 110.0},
		},
	}

	evaluator := NewStreamingBarEvaluator()
	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "valuewhen"},
		},
		Arguments: []ast.Expression{
			&ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 102.0},
			},
			&ast.Identifier{Name: "high"},
			&ast.Literal{Value: 0.0},
		},
	}

	saved := make([]float64, len(ctx.Data))
	for i := range ctx.Data {
		v, err := evaluator.EvaluateAtBar(call, ctx, i)
		if err != nil {
			t.Fatalf("bar %d: %v", i, err)
		}
		saved[i] = v
	}

	if got := len(evaluator.valuewhenCache); got != 1 {
		t.Errorf("expected exactly 1 cached ValuewhenStateManager, got %d", got)
	}

	requeried, err := evaluator.EvaluateAtBar(call, ctx, 2)
	if err != nil {
		t.Fatalf("historical re-request bar 2: %v", err)
	}
	if requeried != saved[2] {
		t.Errorf("historical re-request bar 2: expected %.4f, got %.4f", saved[2], requeried)
	}
}

// TestStreamingBarEvaluator_ValuewhenUnqualifiedAlias verifies that the bare
// "valuewhen" identifier (without "ta." prefix) resolves to the same handler
// as "ta.valuewhen" and produces correct output.
func TestStreamingBarEvaluator_ValuewhenUnqualifiedAlias(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100.0, High: 105.0},
			{Close: 103.0, High: 108.0},
			{Close: 101.0, High: 106.0},
		},
	}
	evaluator := NewStreamingBarEvaluator()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "valuewhen"}, // unqualified alias
		Arguments: []ast.Expression{
			&ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 102.0},
			},
			&ast.Identifier{Name: "high"},
			&ast.Literal{Value: 0.0},
		},
	}

	result, err := evaluator.EvaluateAtBar(call, ctx, 2)
	if err != nil {
		t.Fatalf("unqualified valuewhen alias: unexpected error: %v", err)
	}
	assertFloat64(t, "valuewhen_alias_carry_forward", result, 108.0, 0)
}

// TestStreamingBarEvaluator_ValuewhenCacheIsolationBySourceExpression verifies that
// two ta.valuewhen calls with identical condition and occurrence but different source
// expressions maintain independent cache entries — the source expression is part of
// the cache key.
func TestStreamingBarEvaluator_ValuewhenCacheIsolationBySourceExpression(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100.0, High: 105.0, Low: 95.0},
			{Close: 103.0, High: 108.0, Low: 98.0},
			{Close: 101.0, High: 106.0, Low: 96.0},
		},
	}
	evaluator := NewStreamingBarEvaluator()

	cond := &ast.BinaryExpression{
		Operator: ">",
		Left:     &ast.Identifier{Name: "close"},
		Right:    &ast.Literal{Value: 102.0},
	}

	makeCall := func(srcField string) *ast.CallExpression {
		return &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "valuewhen"},
			},
			Arguments: []ast.Expression{cond, &ast.Identifier{Name: srcField}, &ast.Literal{Value: 0.0}},
		}
	}

	rHigh, err := evaluator.EvaluateAtBar(makeCall("high"), ctx, 2)
	if err != nil {
		t.Fatalf("source=high: %v", err)
	}
	rLow, err := evaluator.EvaluateAtBar(makeCall("low"), ctx, 2)
	if err != nil {
		t.Fatalf("source=low: %v", err)
	}

	if got := len(evaluator.valuewhenCache); got != 2 {
		t.Errorf("expected 2 independent cache entries (one per source), got %d", got)
	}
	assertFloat64(t, "source_high_carry_forward", rHigh, 108.0, 0)
	assertFloat64(t, "source_low_carry_forward", rLow, 98.0, 0)
}

// TestStreamingBarEvaluator_ValuewhenCarryForwardBetweenMatches verifies that on
// non-matching bars valuewhen returns the source value captured at the most recent
// matching bar, not NaN — the canonical carry-forward contract. This covers the
// evaluator dispatch path (not just the state manager directly).
func TestStreamingBarEvaluator_ValuewhenCarryForwardBetweenMatches(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 111},
			{Close: 108},
			{Close: 107},
			{Close: 121},
			{Close: 109},
		},
	}
	evaluator := NewStreamingBarEvaluator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "valuewhen"},
		},
		Arguments: []ast.Expression{
			closeGtExpr(110),
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 0.0},
		},
	}

	tests := []struct {
		barIdx  int
		wantNaN bool
		want    float64
		label   string
	}{
		{0, true, 0, "no_match_yet_NaN"},
		{1, false, 111, "match_captures_value"},
		{2, false, 111, "carry_forward_after_first_match"},
		{3, false, 111, "carry_forward_two_bars_after"},
		{4, false, 121, "new_match_updates_value"},
		{5, false, 121, "carry_forward_after_second_match"},
	}

	for _, bc := range tests {
		t.Run(bc.label, func(t *testing.T) {
			result, err := evaluator.EvaluateAtBar(call, ctx, bc.barIdx)
			if err != nil {
				t.Fatalf("bar %d: EvaluateAtBar failed: %v", bc.barIdx, err)
			}
			if bc.wantNaN {
				if !math.IsNaN(result) {
					t.Errorf("bar %d: expected NaN, got %.4f", bc.barIdx, result)
				}
			} else {
				assertFloat64(t, bc.label, result, bc.want, 0)
			}
		})
	}
}

// TestStreamingBarEvaluator_ValuewhenCacheIsolationByOccurrence verifies that two
// ta.valuewhen calls with identical condition and source but different occurrence
// values maintain independent cache entries and independent state machines.
func TestStreamingBarEvaluator_ValuewhenCacheIsolationByOccurrence(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100.0, High: 105.0},
			{Close: 103.0, High: 108.0},
			{Close: 101.0, High: 106.0},
			{Close: 104.0, High: 109.0},
		},
	}
	evaluator := NewStreamingBarEvaluator()

	cond := &ast.BinaryExpression{
		Operator: ">",
		Left:     &ast.Identifier{Name: "close"},
		Right:    &ast.Literal{Value: 102.0},
	}
	src := &ast.Identifier{Name: "high"}

	makeCall := func(occurrence int) *ast.CallExpression {
		return &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "valuewhen"},
			},
			Arguments: []ast.Expression{cond, src, &ast.Literal{Value: float64(occurrence)}},
		}
	}

	r0, err := evaluator.EvaluateAtBar(makeCall(0), ctx, 3)
	if err != nil {
		t.Fatalf("occ=0: %v", err)
	}
	r1, err := evaluator.EvaluateAtBar(makeCall(1), ctx, 3)
	if err != nil {
		t.Fatalf("occ=1: %v", err)
	}

	if got := len(evaluator.valuewhenCache); got != 2 {
		t.Errorf("expected 2 independent cache entries (one per occurrence), got %d", got)
	}
	assertFloat64(t, "occ0_most_recent", r0, 109.0, 0)
	assertFloat64(t, "occ1_previous", r1, 108.0, 0)
}
