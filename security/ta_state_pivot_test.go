package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func newTestPivotState(kind pivotKind, left, right, capacity int, fieldName string) *PivotStateManager {
	return newPivotStateManager(kind, left, right, capacity,
		&ast.Identifier{Name: fieldName}, NewStreamingBarEvaluator())
}

func pivotHighData(length, pivotBar int, peakVal, baseVal float64) *context.Context {
	data := make([]context.OHLCV, length)
	for i := range data {
		data[i] = context.OHLCV{High: baseVal, Low: baseVal - 10}
	}
	data[pivotBar] = context.OHLCV{High: peakVal, Low: baseVal - 10}
	return &context.Context{Data: data}
}

// ── Warmup boundary and output sequence ─────────────────────────────────────

// TestPivotStateManager_OutputSequence verifies the full per-bar output
// sequence: NaN during warmup, the pivot value at the detection bar, and NaN
// at non-extremum bars — exercising the catch-up loop from scratch.
func TestPivotStateManager_OutputSequence(t *testing.T) {
	ctx := pivotHighData(7, 2, 110.0, 95.0)

	state := newTestPivotState(pivotKindHigh, 2, 2, len(ctx.Data), "high")

	tests := []struct {
		name   string
		barIdx int
		want   float64
	}{
		{"bar0_pre_warmup", 0, math.NaN()},
		{"bar1_pre_warmup", 1, math.NaN()},
		{"bar2_pre_warmup", 2, math.NaN()},
		{"bar3_pre_warmup", 3, math.NaN()},
		{"bar4_first_detection", 4, 110.0},
		{"bar5_no_pivot", 5, math.NaN()},
		{"bar6_no_pivot", 6, math.NaN()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := state.ComputeAtBar(ctx, tt.barIdx)
			if err != nil {
				t.Fatalf("bar %d: unexpected error: %v", tt.barIdx, err)
			}
			assertFloat64(t, tt.name, result, tt.want, 0)
		})
	}
}

// ── Asymmetric window ────────────────────────────────────────────────────────

// TestPivotStateManager_AsymmetricWindow verifies detection with unequal left
// and right bar counts, confirming the delayed-detection formula holds.
func TestPivotStateManager_AsymmetricWindow(t *testing.T) {
	tests := []struct {
		name      string
		left      int
		right     int
		pivotBar  int
		detectBar int
	}{
		{"left3_right1", 3, 1, 3, 4},
		{"left1_right3", 1, 3, 1, 4},
		{"left2_right3", 2, 3, 2, 5},
		// Zero-window edge cases: empty left or right neighbor set.
		{"left0_right2", 0, 2, 0, 2}, // no left neighbors; detection still delayed by rightBars
		{"left2_right0", 2, 0, 2, 2}, // no right neighbors; detection on same bar as center
		{"left0_right0", 0, 0, 2, 2}, // no neighbors; every bar is trivially a pivot
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := tt.detectBar + 1
			ctx := pivotHighData(total, tt.pivotBar, 110.0, 95.0)
			state := newTestPivotState(pivotKindHigh, tt.left, tt.right, total, "high")

			result, err := state.ComputeAtBar(ctx, tt.detectBar)
			if err != nil {
				t.Fatalf("detect bar %d: %v", tt.detectBar, err)
			}
			assertFloat64(t, tt.name, result, 110.0, 0)
		})
	}
}

// ── Access patterns ──────────────────────────────────────────────────────────

// TestPivotStateManager_Idempotency verifies that re-querying the same bar
// multiple times never mutates the buffered result.
func TestPivotStateManager_Idempotency(t *testing.T) {
	ctx := pivotHighData(7, 2, 110.0, 95.0)
	state := newTestPivotState(pivotKindHigh, 2, 2, len(ctx.Data), "high")

	first, err := state.ComputeAtBar(ctx, 4)
	if err != nil {
		t.Fatalf("initial query: %v", err)
	}

	for rep := 1; rep <= 5; rep++ {
		v, err := state.ComputeAtBar(ctx, 4)
		if err != nil {
			t.Fatalf("rep %d: %v", rep, err)
		}
		if !floatEq(v, first) {
			t.Errorf("rep %d: value mutated %.6f → %.6f", rep, first, v)
		}
	}
}

// TestPivotStateManager_HistoricalAccessAfterAdvance verifies that after
// advancing the cursor to a later bar, earlier bars can still be retrieved
// via the ForwardSeriesBuffer.
func TestPivotStateManager_HistoricalAccessAfterAdvance(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100}, {High: 105}, {High: 110},
			{High: 108}, {High: 103},
			{High: 120}, {High: 125}, {High: 130},
			{High: 128}, {High: 122},
		},
	}
	state := newTestPivotState(pivotKindHigh, 2, 2, len(ctx.Data), "high")

	result9, err := state.ComputeAtBar(ctx, 9)
	if err != nil {
		t.Fatalf("bar 9: %v", err)
	}
	assertFloat64(t, "bar9", result9, 130.0, 0)

	result4, err := state.ComputeAtBar(ctx, 4)
	if err != nil {
		t.Fatalf("bar 4: %v", err)
	}
	assertFloat64(t, "bar4_historical", result4, 110.0, 0)
}

// TestPivotStateManager_FullHistoricalConsistency computes all bars
// sequentially, saves results, then re-queries every bar and verifies no
// value has changed.
func TestPivotStateManager_FullHistoricalConsistency(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100}, {High: 105}, {High: 110},
			{High: 108}, {High: 103}, {High: 101},
			{High: 115}, {High: 120}, {High: 118},
			{High: 116}, {High: 112},
		},
	}
	n := len(ctx.Data)
	state := newTestPivotState(pivotKindHigh, 2, 2, n, "high")

	saved := make([]float64, n)
	for i := 0; i < n; i++ {
		v, err := state.ComputeAtBar(ctx, i)
		if err != nil {
			t.Fatalf("bar %d: %v", i, err)
		}
		saved[i] = v
	}

	for i := 0; i < n; i++ {
		v, err := state.ComputeAtBar(ctx, i)
		if err != nil {
			t.Fatalf("re-query bar %d: %v", i, err)
		}
		if !floatEq(v, saved[i]) {
			t.Errorf("bar %d: historical value changed %.6f → %.6f", i, saved[i], v)
		}
	}
}

// ── Context growth ───────────────────────────────────────────────────────────

// TestPivotStateManager_GrowsWhenContextExpands verifies that growIfNeeded
// reinitialises the buffer when a larger context is presented, producing
// correct results without panicking.
func TestPivotStateManager_GrowsWhenContextExpands(t *testing.T) {
	smallCtx := &context.Context{
		Data: []context.OHLCV{
			{High: 100}, {High: 105}, {High: 103},
		},
	}
	largeCtx := pivotHighData(5, 2, 110.0, 90.0)

	state := newTestPivotState(pivotKindHigh, 2, 2, len(smallCtx.Data), "high")
	state.ComputeAtBar(smallCtx, 2) //nolint:errcheck

	result, err := state.ComputeAtBar(largeCtx, 4)
	if err != nil {
		t.Fatalf("after growth: %v", err)
	}
	assertFloat64(t, "after_growth", result, 110.0, 0)
}

// ── Bounds ───────────────────────────────────────────────────────────────────

// TestPivotStateManager_OutOfBoundsReturnsNaN verifies that barIdx values
// outside [0, len(data)) return NaN without error.
func TestPivotStateManager_OutOfBoundsReturnsNaN(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{{High: 100}, {High: 105}, {High: 103}},
	}
	state := newTestPivotState(pivotKindHigh, 1, 1, len(ctx.Data), "high")

	for _, barIdx := range []int{-1, -100, len(ctx.Data), len(ctx.Data) + 5} {
		result, err := state.ComputeAtBar(ctx, barIdx)
		if err != nil {
			t.Fatalf("barIdx %d: unexpected error: %v", barIdx, err)
		}
		if !math.IsNaN(result) {
			t.Errorf("barIdx %d: expected NaN, got %.2f", barIdx, result)
		}
	}
}

// ── Kind routing ─────────────────────────────────────────────────────────────

// TestPivotStateManager_LowDetection verifies the full per-bar output sequence
// for pivotKindLow: NaN during warmup, the trough value at detection, and NaN
// on non-trough bars — symmetric counterpart to TestPivotStateManager_OutputSequence.
func TestPivotStateManager_LowDetection(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Low: 90}, {Low: 85}, {Low: 80},
			{Low: 82}, {Low: 87},
			{Low: 89}, {Low: 91},
		},
	}
	state := newTestPivotState(pivotKindLow, 2, 2, len(ctx.Data), "low")

	tests := []struct {
		name   string
		barIdx int
		want   float64
	}{
		{"bar0_pre_warmup", 0, math.NaN()},
		{"bar1_pre_warmup", 1, math.NaN()},
		{"bar2_pre_warmup", 2, math.NaN()},
		{"bar3_pre_warmup", 3, math.NaN()},
		{"bar4_detection", 4, 80.0},
		{"bar5_no_trough", 5, math.NaN()},
		{"bar6_no_trough", 6, math.NaN()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := state.ComputeAtBar(ctx, tt.barIdx)
			if err != nil {
				t.Fatalf("bar %d: unexpected error: %v", tt.barIdx, err)
			}
			assertFloat64(t, tt.name, result, tt.want, 0)
		})
	}
}

// ── Strict inequality ─────────────────────────────────────────────────────────

// TestPivotStateManager_StrictNeighborInequality verifies that an equal-value
// neighbor prevents pivot detection.  PineScript requires strict inequality:
// the center must be strictly greater (high) or less (low) than every neighbor.
func TestPivotStateManager_StrictNeighborInequality(t *testing.T) {
	tests := []struct {
		name    string
		kind    pivotKind
		data    []context.OHLCV
		barIdx  int
		wantNaN bool
		want    float64
	}{
		{
			name: "high_equal_right_neighbor",
			kind: pivotKindHigh,
			data: []context.OHLCV{
				{High: 90}, {High: 90}, {High: 100},
				{High: 100}, {High: 90},
			},
			barIdx: 4, wantNaN: true,
		},
		{
			name: "high_equal_left_neighbor",
			kind: pivotKindHigh,
			data: []context.OHLCV{
				{High: 90}, {High: 100}, {High: 100},
				{High: 90}, {High: 80},
			},
			barIdx: 4, wantNaN: true,
		},
		{
			name: "high_strictly_greater_contrast",
			kind: pivotKindHigh,
			data: []context.OHLCV{
				{High: 90}, {High: 95}, {High: 100},
				{High: 95}, {High: 90},
			},
			barIdx: 4, wantNaN: false, want: 100.0,
		},
		{
			name: "low_equal_right_neighbor",
			kind: pivotKindLow,
			data: []context.OHLCV{
				{Low: 90}, {Low: 90}, {Low: 80},
				{Low: 80}, {Low: 90},
			},
			barIdx: 4, wantNaN: true,
		},
		{
			name: "low_equal_left_neighbor",
			kind: pivotKindLow,
			data: []context.OHLCV{
				{Low: 90}, {Low: 80}, {Low: 80},
				{Low: 90}, {Low: 95},
			},
			barIdx: 4, wantNaN: true,
		},
		{
			name: "low_strictly_less_contrast",
			kind: pivotKindLow,
			data: []context.OHLCV{
				{Low: 90}, {Low: 85}, {Low: 80},
				{Low: 85}, {Low: 90},
			},
			barIdx: 4, wantNaN: false, want: 80.0,
		},
		{
			name: "high_all_equal",
			kind: pivotKindHigh,
			data: []context.OHLCV{
				{High: 5}, {High: 5}, {High: 5},
				{High: 5}, {High: 5},
			},
			barIdx: 4, wantNaN: true,
		},
		{
			name: "low_all_equal",
			kind: pivotKindLow,
			data: []context.OHLCV{
				{Low: 1}, {Low: 1}, {Low: 1},
				{Low: 1}, {Low: 1},
			},
			barIdx: 4, wantNaN: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &context.Context{Data: tt.data}
			fieldName := "high"
			if tt.kind == pivotKindLow {
				fieldName = "low"
			}
			state := newTestPivotState(tt.kind, 2, 2, len(tt.data), fieldName)

			result, err := state.ComputeAtBar(ctx, tt.barIdx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNaN {
				if !math.IsNaN(result) {
					t.Errorf("expected NaN (equal neighbor disqualifies), got %.2f", result)
				}
			} else {
				assertFloat64(t, tt.name, result, tt.want, 0)
			}
		})
	}
}

// ── Source expression ────────────────────────────────────────────────────────

// TestPivotStateManager_NaNNeighborBlocks verifies that when a source expression
// produces NaN at any neighbor bar, the pivot is not detected — matching PineScript
// where "center > na" evaluates to false and the bar is not confirmed as extremum.
func TestPivotStateManager_NaNNeighborBlocks(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 1}, {Close: 2},
			{Close: 10},
			{Close: 7}, {Close: 6},
		},
	}

	nanWhenSmall := &ast.ConditionalExpression{
		Test: &ast.BinaryExpression{
			Operator: ">=",
			Left:     &ast.Identifier{Name: "close"},
			Right:    &ast.Literal{Value: 5.0},
		},
		Consequent: &ast.Identifier{Name: "close"},
		Alternate:  &ast.Literal{Value: math.NaN()},
	}

	state := newPivotStateManager(pivotKindHigh, 2, 2, len(ctx.Data), nanWhenSmall, NewStreamingBarEvaluator())

	result, err := state.ComputeAtBar(ctx, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !math.IsNaN(result) {
		t.Errorf("expected NaN (NaN left neighbors block), got %.4f", result)
	}
}

// TestPivotStateManager_NaNSourceDetectsAfterWarmup verifies that once a NaN-producing
// source expression has warmed up and all neighbors are valid, pivot detection proceeds
// normally — NaN blocking only applies to bars where the source actually yields NaN.
func TestPivotStateManager_NaNSourceDetectsAfterWarmup(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 1}, {Close: 2}, {Close: 3}, {Close: 4},
			{Close: 5}, {Close: 6}, {Close: 10}, {Close: 8},
			{Close: 7}, {Close: 6},
		},
	}

	nanWhenSmall := &ast.ConditionalExpression{
		Test: &ast.BinaryExpression{
			Operator: ">=",
			Left:     &ast.Identifier{Name: "close"},
			Right:    &ast.Literal{Value: 5.0},
		},
		Consequent: &ast.Identifier{Name: "close"},
		Alternate:  &ast.Literal{Value: math.NaN()},
	}

	state := newPivotStateManager(pivotKindHigh, 2, 2, len(ctx.Data), nanWhenSmall, NewStreamingBarEvaluator())

	result, err := state.ComputeAtBar(ctx, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFloat64(t, "post_warmup_pivot", result, 10.0, 0)
}

// TestPivotStateManager_ArbitrarySourceExpression verifies that any evaluatable
// expression (not only bare field names) can serve as the source.
func TestPivotStateManager_ArbitrarySourceExpression(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Low: 90},
			{High: 110, Low: 100},
			{High: 120, Low: 110},
			{High: 115, Low: 105},
			{High: 110, Low: 100},
		},
	}
	hl2Expr := &ast.Identifier{Name: "hl2"}
	state := newPivotStateManager(pivotKindHigh, 2, 2, len(ctx.Data), hl2Expr, NewStreamingBarEvaluator())

	result, err := state.ComputeAtBar(ctx, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFloat64(t, "hl2_pivot_high", result, 115.0, 0)
}
