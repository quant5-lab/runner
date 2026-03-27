package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

// TestValuewhenStateManager_RingCapacityInvariant verifies that the internal
// match index ring never grows beyond O(occurrence+1) capacity regardless of
// how many bars are processed or how frequently the condition matches.
func TestValuewhenStateManager_RingCapacityInvariant(t *testing.T) {
	tests := []struct {
		name        string
		occurrence  int
		barCount    int
		matchEveryN int
	}{
		{"occ0_dense_matches", 0, 1000, 1},
		{"occ1_dense_matches", 1, 5000, 1},
		{"occ5_sparse_matches", 5, 10000, 10},
		{"occ10_dense_matches", 10, 50000, 1},
		{"occ2_sparse_matches", 2, 1000, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]context.OHLCV, tt.barCount)
			for i := 0; i < tt.barCount; i++ {
				data[i] = context.OHLCV{Close: float64(100 + i)}
			}

			ctx := &context.Context{Data: data}

			conditionExpr := &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "bar_index"},
				Operator: "%",
				Right:    &ast.Literal{Value: float64(tt.matchEveryN)},
			}
			conditionExpr = &ast.BinaryExpression{
				Left:     conditionExpr,
				Operator: "==",
				Right:    &ast.Literal{Value: 0.0},
			}

			sourceExpr := &ast.Identifier{Name: "close"}
			evaluator := NewStreamingBarEvaluator()

			mgr := NewValuewhenStateManager(
				"test_capacity_invariant",
				tt.occurrence,
				conditionExpr,
				sourceExpr,
				len(data),
				evaluator,
			)

			finalBar := tt.barCount - 1
			_, err := mgr.ComputeAtBar(ctx, nil, finalBar)
			if err != nil {
				t.Fatalf("ComputeAtBar(%d) failed: %v", finalBar, err)
			}

			expectedCapacity := tt.occurrence + 1
			actualCapacity := mgr.matchRing.capacity

			if actualCapacity != expectedCapacity {
				t.Errorf("ring capacity: expected %d, got %d (occurrence+1 invariant violated)",
					expectedCapacity, actualCapacity)
			}

			if len(mgr.matchRing.buffer) != expectedCapacity {
				t.Errorf("ring buffer allocation: expected %d, got %d",
					expectedCapacity, len(mgr.matchRing.buffer))
			}

			if mgr.matchRing.Size() > expectedCapacity {
				t.Errorf("ring size %d exceeds capacity %d",
					mgr.matchRing.Size(), expectedCapacity)
			}
		})
	}
}

// TestValuewhenStateManager_RepeatedBarIdempotencyWithRing extends the
// TAStateManager idempotency contract to valuewhen, verifying that repeated
// queries for the same bar return identical values without mutating ring state.
func TestValuewhenStateManager_RepeatedBarIdempotencyWithRing(t *testing.T) {
	data := make([]context.OHLCV, 30)
	for i := 0; i < 30; i++ {
		data[i] = context.OHLCV{Close: float64(100 + i)}
	}

	ctx := &context.Context{Data: data}

	conditionExpr := &ast.BinaryExpression{
		Operator: ">",
		Left:     &ast.Identifier{Name: "close"},
		Right:    &ast.Literal{Value: 110.0},
	}
	sourceExpr := &ast.Identifier{Name: "close"}
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		occurrence int
		targetBar  int
	}{
		{0, 20},
		{1, 25},
		{2, 29},
	}

	for _, tt := range tests {
		t.Run("occurrence_"+string(rune('0'+tt.occurrence)), func(t *testing.T) {
			mgr := NewValuewhenStateManager(
				"test_idempotency",
				tt.occurrence,
				conditionExpr,
				sourceExpr,
				len(data),
				evaluator,
			)

			first, err := mgr.ComputeAtBar(ctx, nil, tt.targetBar)
			if err != nil {
				t.Fatalf("first call bar %d: %v", tt.targetBar, err)
			}

			initialSize := mgr.matchRing.Size()

			for rep := 1; rep <= 5; rep++ {
				v, err := mgr.ComputeAtBar(ctx, nil, tt.targetBar)
				if err != nil {
					t.Fatalf("rep %d bar %d: %v", rep, tt.targetBar, err)
				}

				if (math.IsNaN(first) && !math.IsNaN(v)) || (!math.IsNaN(first) && math.Abs(v-first) > 1e-9) {
					t.Errorf("rep %d: value mutated %.6f → %.6f", rep, first, v)
				}

				if mgr.matchRing.Size() != initialSize {
					t.Errorf("rep %d: ring size mutated %d → %d (idempotency violated)",
						rep, initialSize, mgr.matchRing.Size())
				}
			}
		})
	}
}

// TestValuewhenStateManager_HistoricalAnchorStableAfterAdvance verifies that
// ring-based lookback maintains historical stability: a value computed at an
// anchor bar remains identical after advancing past it and re-querying.
func TestValuewhenStateManager_HistoricalAnchorStableAfterAdvance(t *testing.T) {
	data := make([]context.OHLCV, 50)
	for i := 0; i < 50; i++ {
		data[i] = context.OHLCV{Close: float64(100 + i)}
	}

	ctx := &context.Context{Data: data}

	conditionExpr := &ast.BinaryExpression{
		Operator: ">",
		Left:     &ast.Identifier{Name: "close"},
		Right:    &ast.Literal{Value: 110.0},
	}
	sourceExpr := &ast.Identifier{Name: "close"}
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		occurrence int
		anchorBar  int
		advanceBar int
	}{
		{0, 15, 45},
		{1, 20, 40},
		{2, 25, 49},
	}

	for _, tt := range tests {
		t.Run("occurrence_"+string(rune('0'+tt.occurrence)), func(t *testing.T) {
			mgr := NewValuewhenStateManager(
				"test_historical_anchor",
				tt.occurrence,
				conditionExpr,
				sourceExpr,
				len(data),
				evaluator,
			)

			first, err := mgr.ComputeAtBar(ctx, nil, tt.anchorBar)
			if err != nil {
				t.Fatalf("anchor bar %d: %v", tt.anchorBar, err)
			}

			_, _ = mgr.ComputeAtBar(ctx, nil, tt.advanceBar)

			for rep := 1; rep <= 3; rep++ {
				v, err := mgr.ComputeAtBar(ctx, nil, tt.anchorBar)
				if err != nil {
					t.Fatalf("rep %d anchor bar %d: %v", rep, tt.anchorBar, err)
				}

				if (math.IsNaN(first) && !math.IsNaN(v)) || (!math.IsNaN(first) && math.Abs(v-first) > 1e-9) {
					t.Errorf("rep %d: anchor value changed after advance %.6f → %.6f",
						rep, first, v)
				}
			}
		})
	}
}

// TestValuewhenStateManager_RingEvictionSemantics validates that when matches
// exceed ring capacity, the oldest indices are evicted and the most recent
// (occurrence+1) matches remain accessible for correct occurrence-based lookback.
func TestValuewhenStateManager_RingEvictionSemantics(t *testing.T) {
	tests := []struct {
		name                 string
		occurrence           int
		totalMatches         int
		expectedAccessible   int
		expectedInaccessible int
	}{
		{"occ0_10_matches", 0, 10, 1, 9},
		{"occ2_100_matches", 2, 100, 3, 97},
		{"occ5_50_matches", 5, 50, 6, 44},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]context.OHLCV, tt.totalMatches)
			for i := 0; i < tt.totalMatches; i++ {
				data[i] = context.OHLCV{Close: float64(100 + i)}
			}

			ctx := &context.Context{Data: data}

			conditionExpr := &ast.Literal{Value: 1.0}
			sourceExpr := &ast.Identifier{Name: "close"}
			evaluator := NewStreamingBarEvaluator()

			mgr := NewValuewhenStateManager(
				"test_eviction",
				tt.occurrence,
				conditionExpr,
				sourceExpr,
				len(data),
				evaluator,
			)

			finalBar := tt.totalMatches - 1
			_, err := mgr.ComputeAtBar(ctx, nil, finalBar)
			if err != nil {
				t.Fatalf("ComputeAtBar(%d) failed: %v", finalBar, err)
			}

			for n := 0; n < tt.expectedAccessible; n++ {
				idx, found := mgr.matchRing.GetNthMostRecent(n)
				if !found {
					t.Errorf("GetNthMostRecent(%d): expected accessible (within capacity), got not found", n)
				}
				expectedIdx := tt.totalMatches - 1 - n
				if idx != expectedIdx {
					t.Errorf("GetNthMostRecent(%d): expected bar %d, got %d", n, expectedIdx, idx)
				}
			}

			for n := tt.expectedAccessible; n < tt.totalMatches; n++ {
				_, found := mgr.matchRing.GetNthMostRecent(n)
				if found {
					t.Errorf("GetNthMostRecent(%d): expected not found (evicted), got accessible", n)
				}
			}

			if mgr.matchRing.Size() != tt.expectedAccessible {
				t.Errorf("ring size: expected %d, got %d", tt.expectedAccessible, mgr.matchRing.Size())
			}
		})
	}
}
