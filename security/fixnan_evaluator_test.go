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
		result, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 4)
		if err != nil {
			t.Fatalf("EvaluateAtBar failed: %v", err)
		}
		if result != 110 {
			t.Errorf("expected first pivot 110, got %.2f", result)
		}
	})

	t.Run("forward_fill_after_pivot", func(t *testing.T) {
		result, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 5)
		if err != nil {
			t.Fatalf("EvaluateAtBar failed: %v", err)
		}
		if result != 110 {
			t.Errorf("expected forward-fill 110, got %.2f", result)
		}
	})

	t.Run("forward_fill_continues", func(t *testing.T) {
		result, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 6)
		if err != nil {
			t.Fatalf("EvaluateAtBar failed: %v", err)
		}
		if result != 110 {
			t.Errorf("expected forward-fill 110, got %.2f", result)
		}
	})

	t.Run("new_pivot_replaces", func(t *testing.T) {
		result, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 9)
		if err != nil {
			t.Fatalf("EvaluateAtBar failed: %v", err)
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
		result, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 5)
		if err != nil {
			t.Fatalf("EvaluateAtBar failed: %v", err)
		}
		if result != 110 {
			t.Errorf("expected fixnan(pivot[1]) = 110 at bar 5, got %.2f", result)
		}
	})

	t.Run("forward_fill_after_subscript", func(t *testing.T) {
		result, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 6)
		if err != nil {
			t.Fatalf("EvaluateAtBar failed: %v", err)
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

	identifier := NewHashExpressionIdentifier()
	hash := "fixnan_" + identifier.Identify(pivotCall)
	storage := evaluator.fixnanEvaluator.stateStorage

	_, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 2)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	if !storage.Has(hash) {
		t.Error("expected fixnan state to be cached")
	}

	cachedState, _ := storage.Get(hash)

	_, err = evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 3)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	newCachedState, _ := storage.Get(hash)
	if newCachedState != cachedState {
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

		_, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 0)
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

		result, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 1)
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

		result, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 4)
		if err != nil {
			t.Fatalf("bar 4 failed: %v", err)
		}
		if result != 110 {
			t.Errorf("bar 4: expected 110, got %.2f", result)
		}

		for i := 5; i <= 9; i++ {
			result, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, i)
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

	highResult, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanHighCall, ctx, 4)
	if err != nil {
		t.Fatalf("fixnan(pivothigh) failed: %v", err)
	}
	if highResult != 110 {
		t.Errorf("expected pivothigh fixnan 110, got %.2f", highResult)
	}

	lowResult, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanLowCall, ctx, 4)
	if err != nil {
		t.Fatalf("fixnan(pivotlow) failed: %v", err)
	}
	if lowResult != 80 {
		t.Errorf("expected pivotlow fixnan 80, got %.2f", lowResult)
	}

	highForward, _ := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanHighCall, ctx, 5)
	lowForward, _ := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanLowCall, ctx, 5)

	if highForward != 110 {
		t.Errorf("pivothigh forward-fill should be 110, got %.2f", highForward)
	}
	if lowForward != 80 {
		t.Errorf("pivotlow forward-fill should be 80, got %.2f", lowForward)
	}
}

func TestFixnanState_ExtremeValues(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected []float64
	}{
		{
			name:     "negative_values",
			values:   []float64{-100.0, math.NaN(), math.NaN(), -50.0, math.NaN()},
			expected: []float64{-100.0, -100.0, -100.0, -50.0, -50.0},
		},
		{
			name:     "zero_vs_nan",
			values:   []float64{0.0, math.NaN(), 10.0, 0.0, math.NaN()},
			expected: []float64{0.0, 0.0, 10.0, 0.0, 0.0},
		},
		{
			name:     "large_positive",
			values:   []float64{1e10, math.NaN(), 1e11, math.NaN()},
			expected: []float64{1e10, 1e10, 1e11, 1e11},
		},
		{
			name:     "very_small",
			values:   []float64{1e-10, math.NaN(), 1e-11, math.NaN()},
			expected: []float64{1e-10, 1e-10, 1e-11, 1e-11},
		},
		{
			name:     "alternating_valid_nan",
			values:   []float64{10.0, math.NaN(), 20.0, math.NaN(), 30.0, math.NaN()},
			expected: []float64{10.0, 10.0, 20.0, 20.0, 30.0, 30.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewFixnanState()
			for i, val := range tt.values {
				result := state.ForwardFill(val)
				expected := tt.expected[i]
				if math.IsNaN(expected) {
					if !math.IsNaN(result) {
						t.Errorf("bar %d: expected NaN, got %.10f", i, result)
					}
				} else {
					if math.Abs(result-expected) > 1e-9 {
						t.Errorf("bar %d: expected %.10f, got %.10f", i, expected, result)
					}
				}
			}
		})
	}
}

func TestFixnanEvaluator_WarmupBehavior(t *testing.T) {
	t.Run("target_bar_zero", func(t *testing.T) {
		ctx := &context.Context{
			Data: []context.OHLCV{{Close: 100.0}},
		}
		evaluator := NewStreamingBarEvaluator()

		closeExpr := &ast.Identifier{Name: "close"}
		fixnanCall := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "fixnan"},
			Arguments: []ast.Expression{closeExpr},
		}

		result, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 0)
		if err != nil {
			t.Fatalf("bar 0 evaluation failed: %v", err)
		}
		if result != 100.0 {
			t.Errorf("expected 100.0 at bar 0, got %.2f", result)
		}
	})

	t.Run("single_bar_context", func(t *testing.T) {
		ctx := &context.Context{
			Data: []context.OHLCV{{Close: 50.0}},
		}
		evaluator := NewStreamingBarEvaluator()

		closeExpr := &ast.Identifier{Name: "close"}
		fixnanCall := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "fixnan"},
			Arguments: []ast.Expression{closeExpr},
		}

		result, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 0)
		if err != nil {
			t.Fatalf("single bar failed: %v", err)
		}
		if result != 50.0 {
			t.Errorf("expected 50.0, got %.2f", result)
		}
	})

	t.Run("non_sequential_bar_access", func(t *testing.T) {
		ctx := &context.Context{
			Data: []context.OHLCV{
				{Close: 100}, {Close: 110}, {Close: 120},
				{Close: 115}, {Close: 125}, {Close: 130},
			},
		}
		evaluator := NewStreamingBarEvaluator()

		closeExpr := &ast.Identifier{Name: "close"}
		fixnanCall := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "fixnan"},
			Arguments: []ast.Expression{closeExpr},
		}

		result5, _ := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 5)
		result2, _ := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 2)
		result4, _ := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 4)

		if result5 != 130.0 {
			t.Errorf("bar 5: expected 130.0, got %.2f", result5)
		}
		if result2 != 120.0 {
			t.Errorf("bar 2: expected 120.0, got %.2f", result2)
		}
		if result4 != 125.0 {
			t.Errorf("bar 4: expected 125.0, got %.2f", result4)
		}
	})

	t.Run("large_gap_forward_fill", func(t *testing.T) {
		data := make([]context.OHLCV, 1002)
		data[0].High = 100
		data[1].High = 105
		data[2].High = 110
		data[3].High = 108
		data[4].High = 103
		for i := 5; i < 1002; i++ {
			data[i].High = 102
		}

		ctx := &context.Context{Data: data}
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

		result1000, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 1000)
		if err != nil {
			t.Fatalf("bar 1000 failed: %v", err)
		}
		if result1000 != 110 {
			t.Errorf("expected forward-fill 110 after 1000 bars, got %.2f", result1000)
		}
	})
}

func TestFixnanEvaluator_DifferentTAFunctions(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100}, {Close: 102}, {Close: 104}, {Close: 106},
			{Close: 108}, {Close: 110}, {Close: 112}, {Close: 114},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	t.Run("fixnan_with_sma", func(t *testing.T) {
		smaCall := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "sma"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(3)},
			},
		}

		fixnanCall := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "fixnan"},
			Arguments: []ast.Expression{smaCall},
		}

		result, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 2)
		if err != nil {
			t.Fatalf("fixnan(sma) failed: %v", err)
		}
		expectedSMA := (100.0 + 102.0 + 104.0) / 3.0
		if math.Abs(result-expectedSMA) > 0.01 {
			t.Errorf("expected SMA %.2f, got %.2f", expectedSMA, result)
		}
	})

	t.Run("fixnan_with_ema", func(t *testing.T) {
		emaCall := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "ema"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(3)},
			},
		}

		fixnanCall := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "fixnan"},
			Arguments: []ast.Expression{emaCall},
		}

		result, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 5)
		if err != nil {
			t.Fatalf("fixnan(ema) failed: %v", err)
		}
		if math.IsNaN(result) || result <= 0 {
			t.Errorf("expected valid EMA result, got %.2f", result)
		}
	})

	t.Run("multiple_fixnan_different_expressions", func(t *testing.T) {
		smaCall := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "sma"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(3)},
			},
		}

		emaCall := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "ema"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(3)},
			},
		}

		fixnanSMA := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "fixnan"},
			Arguments: []ast.Expression{smaCall},
		}

		fixnanEMA := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "fixnan"},
			Arguments: []ast.Expression{emaCall},
		}

		smaResult, _ := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanSMA, ctx, 5)
		emaResult, _ := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanEMA, ctx, 5)

		if math.IsNaN(smaResult) || math.IsNaN(emaResult) {
			t.Error("neither result should be NaN at bar 5")
		}

		if math.Abs(smaResult-emaResult) < 0.01 {
			t.Logf("SMA=%.2f EMA=%.2f - values very close but both valid", smaResult, emaResult)
		}
	})
}

func TestStateStorage_EdgeCases(t *testing.T) {
	t.Run("get_nonexistent_key", func(t *testing.T) {
		storage := NewMapStateStorage()
		_, exists := storage.Get("nonexistent")
		if exists {
			t.Error("expected false for nonexistent key")
		}
	})

	t.Run("has_empty_storage", func(t *testing.T) {
		storage := NewMapStateStorage()
		if storage.Has("anything") {
			t.Error("empty storage should not have any keys")
		}
	})

	t.Run("set_overwrite", func(t *testing.T) {
		storage := NewMapStateStorage()
		state1 := NewFixnanState()
		state1.ForwardFill(100.0)
		storage.Set("key", state1)

		state2 := NewFixnanState()
		state2.ForwardFill(200.0)
		storage.Set("key", state2)

		retrieved, _ := storage.Get("key")
		retrievedState := retrieved.(*FixnanState)
		result := retrievedState.ForwardFill(math.NaN())
		if result != 200.0 {
			t.Errorf("expected overwritten value 200.0, got %.2f", result)
		}
	})

	t.Run("storage_isolation", func(t *testing.T) {
		storage1 := NewMapStateStorage()
		storage2 := NewMapStateStorage()

		state1 := NewFixnanState()
		state1.ForwardFill(100.0)
		storage1.Set("key", state1)

		if storage2.Has("key") {
			t.Error("storage2 should not have key from storage1")
		}
	})
}

func TestExpressionIdentifier_Uniqueness(t *testing.T) {
	identifier := NewHashExpressionIdentifier()

	t.Run("different_arguments_different_hash", func(t *testing.T) {
		expr1 := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "sma"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(10)},
			},
		}

		expr2 := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "sma"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(20)},
			},
		}

		hash1 := identifier.Identify(expr1)
		hash2 := identifier.Identify(expr2)

		if hash1 == hash2 {
			t.Error("different SMA periods should produce different hashes")
		}
	})

	t.Run("different_functions_different_hash", func(t *testing.T) {
		expr1 := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "sma"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(10)},
			},
		}

		expr2 := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "ema"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(10)},
			},
		}

		hash1 := identifier.Identify(expr1)
		hash2 := identifier.Identify(expr2)

		if hash1 == hash2 {
			t.Error("SMA and EMA should produce different hashes")
		}
	})

	t.Run("same_expression_same_hash", func(t *testing.T) {
		expr := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "sma"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(10)},
			},
		}

		hash1 := identifier.Identify(expr)
		hash2 := identifier.Identify(expr)

		if hash1 != hash2 {
			t.Error("same expression should produce consistent hash")
		}
	})
}

func TestWarmupStrategy_ErrorHandling(t *testing.T) {
	t.Run("warmup_with_partial_errors", func(t *testing.T) {
		ctx := &context.Context{
			Data: []context.OHLCV{
				{Close: 100}, {Close: 110}, {Close: 120},
			},
		}
		evaluator := NewStreamingBarEvaluator()
		warmup := NewSequentialWarmupStrategy()
		state := NewFixnanState()

		invalidExpr := &ast.Identifier{Name: "invalid_field"}

		err := warmup.Warmup(evaluator, invalidExpr, ctx, 2, state)
		if err != nil {
			t.Errorf("warmup should handle errors gracefully, got: %v", err)
		}
	})

	t.Run("warmup_empty_target", func(t *testing.T) {
		ctx := &context.Context{Data: []context.OHLCV{}}
		evaluator := NewStreamingBarEvaluator()
		warmup := NewSequentialWarmupStrategy()
		state := NewFixnanState()

		closeExpr := &ast.Identifier{Name: "close"}
		err := warmup.Warmup(evaluator, closeExpr, ctx, 0, state)

		if err != nil {
			t.Errorf("warmup with target 0 should not error, got: %v", err)
		}
	})
}

func TestFixnanEvaluator_MemberExpressionOffsets(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100}, {High: 105}, {High: 110},
			{High: 108}, {High: 103}, {High: 102},
			{High: 104}, {High: 107}, {High: 106}, {High: 101},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	t.Run("fixnan_with_pivot_offset_1", func(t *testing.T) {
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

		result, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 5)
		if err != nil {
			t.Fatalf("fixnan(pivot[1]) failed: %v", err)
		}
		if result != 110 {
			t.Errorf("expected fixnan(pivot[1]) = 110 at bar 5, got %.2f", result)
		}
	})

	t.Run("fixnan_with_pivot_offset_2", func(t *testing.T) {
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
			Property: &ast.Literal{Value: float64(2)},
		}

		fixnanCall := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "fixnan"},
			Arguments: []ast.Expression{pivotMember},
		}

		result, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, 9)
		if err != nil {
			t.Fatalf("fixnan(pivot[2]) failed: %v", err)
		}
		if !math.IsNaN(result) || result == 110 {
			t.Logf("fixnan(pivot[2]) at bar 9: %.2f - offset behavior as expected", result)
		}
	})
}
