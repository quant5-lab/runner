package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func newTestValuewhenState(cond, src ast.Expression, occurrence, capacity int) *ValuewhenStateManager {
	return newValuewhenStateManager(cond, src, occurrence, capacity, NewStreamingBarEvaluator())
}

func closeGtExpr(threshold float64) *ast.BinaryExpression {
	return &ast.BinaryExpression{
		Operator: ">",
		Left:     &ast.Identifier{Name: "close"},
		Right:    &ast.Literal{Value: threshold},
	}
}

// ── matchRing invariants ─────────────────────────────────────────────────────

// TestMatchRing_OccurrenceOrdering verifies that nthMostRecent returns pushed
// values in reverse-insertion order.
func TestMatchRing_OccurrenceOrdering(t *testing.T) {
	ring := newMatchRing(3)
	ring.push(10.0)
	ring.push(20.0)
	ring.push(30.0)

	cases := []struct {
		n    int
		want float64
	}{
		{0, 30.0},
		{1, 20.0},
		{2, 10.0},
	}
	for _, c := range cases {
		v, ok := ring.nthMostRecent(c.n)
		if !ok || v != c.want {
			t.Errorf("nthMostRecent(%d): expected (%.1f, true), got (%.1f, %v)", c.n, c.want, v, ok)
		}
	}
}

// TestMatchRing_InsufficientCount verifies that nthMostRecent returns false
// when fewer than n+1 values have been pushed, including on a completely empty ring.
func TestMatchRing_InsufficientCount(t *testing.T) {
	empty := newMatchRing(3)
	for _, n := range []int{0, 1, 5} {
		if _, ok := empty.nthMostRecent(n); ok {
			t.Errorf("nthMostRecent(%d) on empty ring: expected false, got true", n)
		}
	}

	ring := newMatchRing(3)
	ring.push(5.0)
	if _, ok := ring.nthMostRecent(0); !ok {
		t.Error("nthMostRecent(0) after 1 push: expected true, got false")
	}
	for _, n := range []int{1, 2, 3, 100} {
		if _, ok := ring.nthMostRecent(n); ok {
			t.Errorf("nthMostRecent(%d): expected false for n >= count (count=1)", n)
		}
	}
}

// TestMatchRing_OverwritesOldest verifies that once the ring is full, new
// pushes overwrite the oldest entry and the accessible window stays correct.
// Also covers the degenerate capacity=1 case where each push evicts the previous.
func TestMatchRing_OverwritesOldest(t *testing.T) {
	t.Run("capacity2_overwrite", func(t *testing.T) {
		ring := newMatchRing(2)
		ring.push(10.0)
		ring.push(20.0)
		ring.push(30.0)

		v0, ok0 := ring.nthMostRecent(0)
		v1, ok1 := ring.nthMostRecent(1)
		_, ok2 := ring.nthMostRecent(2)

		if !ok0 || v0 != 30.0 {
			t.Errorf("nthMostRecent(0): expected (30, true), got (%.1f, %v)", v0, ok0)
		}
		if !ok1 || v1 != 20.0 {
			t.Errorf("nthMostRecent(1): expected (20, true), got (%.1f, %v)", v1, ok1)
		}
		if ok2 {
			t.Error("nthMostRecent(2): expected false after overwrite, got true")
		}
	})

	t.Run("capacity1_always_newest", func(t *testing.T) {
		ring := newMatchRing(1)
		ring.push(10.0)
		if v, ok := ring.nthMostRecent(0); !ok || v != 10.0 {
			t.Errorf("after 1st push: expected (10, true), got (%.1f, %v)", v, ok)
		}
		ring.push(20.0)
		if v, ok := ring.nthMostRecent(0); !ok || v != 20.0 {
			t.Errorf("after 2nd push: expected (20, true), got (%.1f, %v)", v, ok)
		}
		if _, ok := ring.nthMostRecent(1); ok {
			t.Error("nthMostRecent(1) on capacity=1 ring: expected false, got true")
		}
	})
}

// ── Condition semantics ──────────────────────────────────────────────────────

// TestValuewhenStateManager_ConditionSemantics verifies the three critical
// condition types: false (0.0), NaN (na), and true (1.0).
func TestValuewhenStateManager_ConditionSemantics(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 110, High: 99},
			{Close: 105, High: 97},
			{Close: 112, High: 98},
		},
	}
	src := &ast.Identifier{Name: "high"}

	tests := []struct {
		name    string
		cond    ast.Expression
		barIdx  int
		wantNaN bool
	}{
		{"false_condition_never_matches", closeGtExpr(200.0), 2, true},
		{"nan_condition_never_matches", &ast.Literal{Value: math.NaN()}, 2, true},
		{"always_true_condition_matches", closeGtExpr(100.0), 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := newTestValuewhenState(tt.cond, src, 0, len(ctx.Data))
			result, err := state.ComputeAtBar(ctx, tt.barIdx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNaN != math.IsNaN(result) {
				t.Errorf("wantNaN=%v, got %.2f (isNaN=%v)", tt.wantNaN, result, math.IsNaN(result))
			}
		})
	}
}

// ── Occurrence semantics ─────────────────────────────────────────────────────

// TestValuewhenStateManager_OccurrenceSemantics verifies carry-forward
// behaviour, occurrence indexing (0 = most recent), and NaN when fewer
// matches than requested have occurred.
func TestValuewhenStateManager_OccurrenceSemantics(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100, High: 101},
			{Close: 110, High: 102},
			{Close: 105, High: 103},
			{Close: 115, High: 104},
			{Close: 108, High: 105},
			{Close: 120, High: 106},
			{Close: 107, High: 107},
		},
	}
	cond := closeGtExpr(109.0)
	src := &ast.Identifier{Name: "high"}

	tests := []struct {
		name       string
		occurrence int
		barIdx     int
		want       float64
	}{
		{"no_match_at_bar0", 0, 0, math.NaN()},
		{"match1_occ0_at_bar1", 0, 1, 102.0},
		{"insufficient_occ1_at_bar1", 1, 1, math.NaN()},
		{"carry_forward_occ0_at_bar2", 0, 2, 102.0},
		{"match2_occ0_at_bar3", 0, 3, 104.0},
		{"match2_occ1_at_bar3", 1, 3, 102.0},
		{"insufficient_occ2_at_bar3", 2, 3, math.NaN()},
		{"match3_occ0_at_bar5", 0, 5, 106.0},
		{"match3_occ1_at_bar5", 1, 5, 104.0},
		{"match3_occ2_at_bar5", 2, 5, 102.0},
		{"carry_occ0_at_bar6", 0, 6, 106.0},
		{"carry_occ1_at_bar6", 1, 6, 104.0},
		{"carry_occ2_at_bar6", 2, 6, 102.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := newTestValuewhenState(cond, src, tt.occurrence, len(ctx.Data))
			result, err := state.ComputeAtBar(ctx, tt.barIdx)
			if err != nil {
				t.Fatalf("bar %d occ %d: unexpected error: %v", tt.barIdx, tt.occurrence, err)
			}
			assertFloat64(t, tt.name, result, tt.want, 0)
		})
	}
}

// ── Access patterns ──────────────────────────────────────────────────────────

// TestValuewhenStateManager_Idempotency verifies that re-querying the same bar
// multiple times never mutates the buffered result.
func TestValuewhenStateManager_Idempotency(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 110, High: 101},
			{Close: 105, High: 102},
			{Close: 115, High: 103},
			{Close: 108, High: 104},
		},
	}
	state := newTestValuewhenState(closeGtExpr(109.0), &ast.Identifier{Name: "high"}, 0, len(ctx.Data))

	first, err := state.ComputeAtBar(ctx, 2)
	if err != nil {
		t.Fatalf("initial query: %v", err)
	}

	for rep := 1; rep <= 5; rep++ {
		v, err := state.ComputeAtBar(ctx, 2)
		if err != nil {
			t.Fatalf("rep %d: %v", rep, err)
		}
		if !floatEq(v, first) {
			t.Errorf("rep %d: value mutated %.6f → %.6f", rep, first, v)
		}
	}
}

// TestValuewhenStateManager_HistoricalAccessAfterAdvance verifies that
// after advancing the cursor to a later bar, earlier bars can still be
// retrieved via the ForwardSeriesBuffer.
func TestValuewhenStateManager_HistoricalAccessAfterAdvance(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 110, High: 100},
			{Close: 105, High: 95},
			{Close: 112, High: 98},
			{Close: 108, High: 91},
			{Close: 115, High: 99},
		},
	}
	state := newTestValuewhenState(closeGtExpr(109.0), &ast.Identifier{Name: "high"}, 0, len(ctx.Data))

	result4, err := state.ComputeAtBar(ctx, 4)
	if err != nil {
		t.Fatalf("bar 4: %v", err)
	}
	assertFloat64(t, "bar4", result4, 99.0, 0)

	result2, err := state.ComputeAtBar(ctx, 2)
	if err != nil {
		t.Fatalf("bar 2 historical: %v", err)
	}
	assertFloat64(t, "bar2_historical", result2, 98.0, 0)
}

// TestValuewhenStateManager_FullHistoricalConsistency computes all bars
// sequentially, saves results, then re-queries every bar and verifies no
// value has changed.
func TestValuewhenStateManager_FullHistoricalConsistency(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100, High: 91},
			{Close: 112, High: 92},
			{Close: 105, High: 93},
			{Close: 115, High: 94},
			{Close: 108, High: 95},
			{Close: 120, High: 96},
			{Close: 103, High: 97},
		},
	}
	n := len(ctx.Data)
	state := newTestValuewhenState(closeGtExpr(109.0), &ast.Identifier{Name: "high"}, 0, n)

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

// TestValuewhenStateManager_GrowsWhenContextExpands verifies that growIfNeeded
// reinitialises the buffer and ring when a larger context is presented.
func TestValuewhenStateManager_GrowsWhenContextExpands(t *testing.T) {
	smallCtx := &context.Context{
		Data: []context.OHLCV{
			{Close: 110, High: 99},
			{Close: 105, High: 95},
		},
	}
	largeCtx := &context.Context{
		Data: []context.OHLCV{
			{Close: 110, High: 99},
			{Close: 105, High: 95},
			{Close: 112, High: 97},
			{Close: 108, High: 93},
			{Close: 115, High: 101},
		},
	}

	state := newTestValuewhenState(closeGtExpr(109.0), &ast.Identifier{Name: "high"}, 0, len(smallCtx.Data))
	state.ComputeAtBar(smallCtx, 1) //nolint:errcheck

	result, err := state.ComputeAtBar(largeCtx, 4)
	if err != nil {
		t.Fatalf("after growth: %v", err)
	}
	assertFloat64(t, "after_growth", result, 101.0, 0)
}

// ── Bounds ───────────────────────────────────────────────────────────────────

// TestValuewhenStateManager_OutOfBoundsReturnsNaN verifies that barIdx values
// outside [0, len(data)) return NaN without error.
func TestValuewhenStateManager_OutOfBoundsReturnsNaN(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{{Close: 100}, {Close: 110}},
	}
	state := newTestValuewhenState(closeGtExpr(105.0), &ast.Identifier{Name: "close"}, 0, len(ctx.Data))

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
