package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func TestFixnanState_ForwardFillProgression(t *testing.T) {
	state := NewFixnanState()

	tests := []struct {
		barIdx   int
		input    float64
		expected float64
		desc     string
	}{
		{0, math.NaN(), math.NaN(), "first bar NaN - no prior valid value"},
		{1, 100.0, 100.0, "first valid value propagates"},
		{2, math.NaN(), 100.0, "forward-fill from bar 1"},
		{3, math.NaN(), 100.0, "forward-fill continues"},
		{4, 105.0, 105.0, "new valid value replaces"},
		{5, math.NaN(), 105.0, "forward-fill from bar 4"},
		{6, math.NaN(), 105.0, "forward-fill continues"},
		{7, 110.0, 110.0, "another valid value"},
		{8, math.NaN(), 110.0, "forward-fill from bar 7"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := state.ForwardFill(tt.input)
			if math.IsNaN(tt.expected) {
				if !math.IsNaN(result) {
					t.Errorf("bar %d: expected NaN, got %.2f", tt.barIdx, result)
				}
			} else {
				if result != tt.expected {
					t.Errorf("bar %d: expected %.2f, got %.2f", tt.barIdx, tt.expected, result)
				}
			}
		})
	}
}

func TestFixnanState_ConsecutiveNaNs(t *testing.T) {
	state := NewFixnanState()

	state.ForwardFill(100.0)

	for i := 0; i < 100; i++ {
		result := state.ForwardFill(math.NaN())
		if result != 100.0 {
			t.Errorf("bar %d: forward-fill should persist 100.0, got %.2f", i+1, result)
		}
	}
}

func TestFixnanState_IsolationBetweenInstances(t *testing.T) {
	state1 := NewFixnanState()
	state2 := NewFixnanState()

	state1.ForwardFill(100.0)
	state2.ForwardFill(200.0)

	result1 := state1.ForwardFill(math.NaN())
	result2 := state2.ForwardFill(math.NaN())

	if result1 != 100.0 {
		t.Errorf("state1: expected 100.0, got %.2f", result1)
	}
	if result2 != 200.0 {
		t.Errorf("state2: expected 200.0, got %.2f", result2)
	}
}

func TestFixnanEvaluator_BasicForwardFill(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100}, {High: 105}, {High: 110},
			{High: 108}, {High: 103}, {High: 102},
			{High: 104}, {High: 107}, {High: 106}, {High: 101},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	pivotCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "pivothigh"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "high"},
			&ast.Literal{Value: float64(2)},
			&ast.Literal{Value: float64(2)},
		},
	}

	fixnanCall := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "fixnan"},
		Arguments: []ast.Expression{pivotCall},
	}

	t.Run("first_valid_pivot", func(t *testing.T) {
		result, err := evaluator.evaluateFixnanAtBar(fixnanCall, ctx, 2)
		if err != nil {
			t.Fatalf("evaluateFixnanAtBar failed: %v", err)
		}
		if result != 110 {
			t.Errorf("expected first pivot 110, got %.2f", result)
		}
	})

	t.Run("forward_fill_after_pivot", func(t *testing.T) {
		result, err := evaluator.evaluateFixnanAtBar(fixnanCall, ctx, 3)
		if err != nil {
			t.Fatalf("evaluateFixnanAtBar failed: %v", err)
		}
		if result != 110 {
			t.Errorf("expected forward-fill 110, got %.2f", result)
		}
	})

	t.Run("forward_fill_continues", func(t *testing.T) {
		result, err := evaluator.evaluateFixnanAtBar(fixnanCall, ctx, 5)
		if err != nil {
			t.Fatalf("evaluateFixnanAtBar failed: %v", err)
		}
		if result != 110 {
			t.Errorf("expected forward-fill 110, got %.2f", result)
		}
	})

	t.Run("new_pivot_replaces", func(t *testing.T) {
		result, err := evaluator.evaluateFixnanAtBar(fixnanCall, ctx, 7)
		if err != nil {
			t.Fatalf("evaluateFixnanAtBar failed: %v", err)
		}
		if result != 107 {
			t.Errorf("expected new pivot 107, got %.2f", result)
		}
	})
}

func TestFixnanEvaluator_WithMemberExpression(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100}, {High: 105}, {High: 110},
			{High: 108}, {High: 103}, {High: 102},
			{High: 104}, {High: 107}, {High: 106}, {High: 101},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	pivotMember := &ast.MemberExpression{
		Object: &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "pivothigh"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: float64(2)},
				&ast.Literal{Value: float64(2)},
			},
		},
		Property: &ast.Literal{Value: float64(1)},
	}

	fixnanCall := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "fixnan"},
		Arguments: []ast.Expression{pivotMember},
	}

	t.Run("fixnan_with_subscript", func(t *testing.T) {
		result, err := evaluator.evaluateFixnanAtBar(fixnanCall, ctx, 3)
		if err != nil {
			t.Fatalf("evaluateFixnanAtBar failed: %v", err)
		}
		if result != 110 {
			t.Errorf("expected fixnan(pivot[1]) = 110 at bar 3, got %.2f", result)
		}
	})

	t.Run("forward_fill_after_subscript", func(t *testing.T) {
		result, err := evaluator.evaluateFixnanAtBar(fixnanCall, ctx, 4)
		if err != nil {
			t.Fatalf("evaluateFixnanAtBar failed: %v", err)
		}
		if result != 110 {
			t.Errorf("expected forward-fill 110, got %.2f", result)
		}
	})
}

func TestFixnanEvaluator_StateCaching(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100}, {High: 105}, {High: 110},
			{High: 108}, {High: 103}, {High: 102},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	pivotCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "pivothigh"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "high"},
			&ast.Literal{Value: float64(2)},
			&ast.Literal{Value: float64(2)},
		},
	}

	fixnanCall := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "fixnan"},
		Arguments: []ast.Expression{pivotCall},
	}

	hash := "fixnan_" + computeExpressionHash(pivotCall)

	_, err := evaluator.evaluateFixnanAtBar(fixnanCall, ctx, 2)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	if evaluator.fixnanStateCache[hash] == nil {
		t.Error("expected fixnan state to be cached")
	}

	cachedState := evaluator.fixnanStateCache[hash]

	_, err = evaluator.evaluateFixnanAtBar(fixnanCall, ctx, 3)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if evaluator.fixnanStateCache[hash] != cachedState {
		t.Error("expected same cached state instance to be reused")
	}
}

func TestFixnanEvaluator_EdgeCases(t *testing.T) {
	evaluator := NewStreamingBarEvaluator()

	t.Run("no_arguments", func(t *testing.T) {
		ctx := &context.Context{Data: []context.OHLCV{{High: 100}}}
		fixnanCall := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "fixnan"},
			Arguments: []ast.Expression{},
		}

		_, err := evaluator.evaluateFixnanAtBar(fixnanCall, ctx, 0)
		if err == nil {
			t.Error("expected error for no arguments")
		}
	})

	t.Run("all_nan_sequence", func(t *testing.T) {
		ctx := &context.Context{
			Data: []context.OHLCV{
				{High: 100}, {High: 102}, {High: 103}, {High: 101},
			},
		}

		pivotCall := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "pivothigh"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: float64(2)},
				&ast.Literal{Value: float64(2)},
			},
		}

		fixnanCall := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "fixnan"},
			Arguments: []ast.Expression{pivotCall},
		}

		result, err := evaluator.evaluateFixnanAtBar(fixnanCall, ctx, 1)
		if err != nil {
			t.Fatalf("evaluateFixnanAtBar failed: %v", err)
		}
		if !math.IsNaN(result) {
			t.Errorf("expected NaN when no pivots exist yet, got %.2f", result)
		}
	})

	t.Run("single_valid_then_all_nan", func(t *testing.T) {
		ctx := &context.Context{
			Data: []context.OHLCV{
				{High: 100}, {High: 105}, {High: 110},
				{High: 108}, {High: 103}, {High: 102},
				{High: 101}, {High: 100}, {High: 99}, {High: 98},
			},
		}

		pivotCall := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "pivothigh"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: float64(2)},
				&ast.Literal{Value: float64(2)},
			},
		}

		fixnanCall := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "fixnan"},
			Arguments: []ast.Expression{pivotCall},
		}

		result, err := evaluator.evaluateFixnanAtBar(fixnanCall, ctx, 2)
		if err != nil {
			t.Fatalf("bar 2 failed: %v", err)
		}
		if result != 110 {
			t.Errorf("bar 2: expected 110, got %.2f", result)
		}

		for i := 3; i <= 9; i++ {
			result, err := evaluator.evaluateFixnanAtBar(fixnanCall, ctx, i)
			if err != nil {
				t.Fatalf("bar %d failed: %v", i, err)
			}
			if result != 110 {
				t.Errorf("bar %d: expected forward-fill 110, got %.2f", i, result)
			}
		}
	})
}

func TestFixnanEvaluator_MultipleSeriesIsolation(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Low: 90},
			{High: 105, Low: 85},
			{High: 110, Low: 80},
			{High: 108, Low: 82},
			{High: 103, Low: 87},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	pivotHighCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "pivothigh"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "high"},
			&ast.Literal{Value: float64(2)},
			&ast.Literal{Value: float64(2)},
		},
	}

	pivotLowCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "pivotlow"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "low"},
			&ast.Literal{Value: float64(2)},
			&ast.Literal{Value: float64(2)},
		},
	}

	fixnanHighCall := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "fixnan"},
		Arguments: []ast.Expression{pivotHighCall},
	}

	fixnanLowCall := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "fixnan"},
		Arguments: []ast.Expression{pivotLowCall},
	}

	highResult, err := evaluator.evaluateFixnanAtBar(fixnanHighCall, ctx, 2)
	if err != nil {
		t.Fatalf("fixnan(pivothigh) failed: %v", err)
	}
	if highResult != 110 {
		t.Errorf("expected pivothigh fixnan 110, got %.2f", highResult)
	}

	lowResult, err := evaluator.evaluateFixnanAtBar(fixnanLowCall, ctx, 2)
	if err != nil {
		t.Fatalf("fixnan(pivotlow) failed: %v", err)
	}
	if lowResult != 80 {
		t.Errorf("expected pivotlow fixnan 80, got %.2f", lowResult)
	}

	highForward, _ := evaluator.evaluateFixnanAtBar(fixnanHighCall, ctx, 3)
	lowForward, _ := evaluator.evaluateFixnanAtBar(fixnanLowCall, ctx, 3)

	if highForward != 110 {
		t.Errorf("pivothigh forward-fill should be 110, got %.2f", highForward)
	}
	if lowForward != 80 {
		t.Errorf("pivotlow forward-fill should be 80, got %.2f", lowForward)
	}
}
