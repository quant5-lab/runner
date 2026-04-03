package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

// closesOnlyCtx creates a context where only the Close field is set per bar.
// Used when the tested expression or condition accesses only the close price.
func closesOnlyCtx(closes ...float64) *context.Context {
	data := make([]context.OHLCV, len(closes))
	for i, c := range closes {
		data[i] = context.OHLCV{Close: c}
	}
	return &context.Context{Data: data}
}

// ── CUMStateManager ───────────────────────────────────────────────────────────

func TestCUMStateManager_RunningSum(t *testing.T) {
	closes := []float64{10, 20, 30, 40, 50}
	ctx := makeCtxOHLCV(closes...)
	m := NewCUMStateManager(len(closes), NewStreamingBarEvaluator())
	src := &ast.Identifier{Name: "close"}

	cumulative := 0.0
	for i, c := range closes {
		cumulative += c
		v, err := m.ComputeAtBar(ctx, src, i)
		if err != nil {
			t.Fatalf("bar %d: %v", i, err)
		}
		if math.Abs(v-cumulative) > 1e-9 {
			t.Errorf("bar %d: want %.1f, got %.9f", i, cumulative, v)
		}
	}
}

func TestCUMStateManager_AllZeroSource(t *testing.T) {
	ctx := closesOnlyCtx(0, 0, 0, 0, 0)
	m := NewCUMStateManager(5, NewStreamingBarEvaluator())
	src := &ast.Identifier{Name: "close"}

	for i := 0; i < 5; i++ {
		v, err := m.ComputeAtBar(ctx, src, i)
		if err != nil {
			t.Fatalf("bar %d: %v", i, err)
		}
		if v != 0 {
			t.Errorf("bar %d: all-zero source must produce 0, got %.6f", i, v)
		}
	}
}

func TestCUMStateManager_NegativeValues(t *testing.T) {
	closes := []float64{10, -5, 3, -8, 2}
	ctx := makeCtxOHLCV(closes...)
	m := NewCUMStateManager(len(closes), NewStreamingBarEvaluator())
	src := &ast.Identifier{Name: "close"}

	cumulative := 0.0
	for i, c := range closes {
		cumulative += c
		v, err := m.ComputeAtBar(ctx, src, i)
		if err != nil {
			t.Fatalf("bar %d: %v", i, err)
		}
		if math.Abs(v-cumulative) > 1e-9 {
			t.Errorf("bar %d: want %.1f, got %.9f", i, cumulative, v)
		}
	}
}

func TestCUMStateManager_HistoricalLookbackStableAfterAdvance(t *testing.T) {
	ctx := makeCtxOHLCV(5, 10, 15, 20, 25)
	m := NewCUMStateManager(5, NewStreamingBarEvaluator())
	src := &ast.Identifier{Name: "close"}

	anchor := 2
	first, err := m.ComputeAtBar(ctx, src, anchor)
	if err != nil {
		t.Fatalf("anchor bar %d: %v", anchor, err)
	}

	_, _ = m.ComputeAtBar(ctx, src, 4)

	second, err := m.ComputeAtBar(ctx, src, anchor)
	if err != nil {
		t.Fatalf("re-query anchor bar %d: %v", anchor, err)
	}
	if math.Abs(first-second) > 1e-9 {
		t.Errorf("anchor bar %d: value changed after advance: %.1f → %.1f", anchor, first, second)
	}
}

func TestCUMStateManager_FullHistoricalConsistency(t *testing.T) {
	ctx := makeCtxOHLCV(3, 7, 2, 9, 4, 1, 8)
	m := NewCUMStateManager(len(ctx.Data), NewStreamingBarEvaluator())
	src := &ast.Identifier{Name: "close"}

	saved := make([]float64, len(ctx.Data))
	for i := range ctx.Data {
		v, err := m.ComputeAtBar(ctx, src, i)
		if err != nil {
			t.Fatalf("bar %d: %v", i, err)
		}
		saved[i] = v
	}

	for i := range ctx.Data {
		v, _ := m.ComputeAtBar(ctx, src, i)
		if math.Abs(v-saved[i]) > 1e-9 {
			t.Errorf("bar %d: historical value changed on re-query: %.6f → %.6f", i, saved[i], v)
		}
	}
}

func TestCUMStateManager_IdempotentRequery(t *testing.T) {
	ctx := makeCtxOHLCV(1, 2, 3, 4, 5)
	m := NewCUMStateManager(5, NewStreamingBarEvaluator())
	src := &ast.Identifier{Name: "close"}

	_, _ = m.ComputeAtBar(ctx, src, 4)
	first, _ := m.ComputeAtBar(ctx, src, 2)

	for rep := 0; rep < 3; rep++ {
		v, _ := m.ComputeAtBar(ctx, src, 2)
		if math.Abs(v-first) > 1e-9 {
			t.Errorf("rep %d: bar 2 changed on repeated query: %.6f → %.6f", rep, first, v)
		}
	}
}

func TestCUMStateManager_InvalidSourceReturnsError(t *testing.T) {
	ctx := createContextWithBars(5)
	m := NewCUMStateManager(5, NewStreamingBarEvaluator())
	invalid := &ast.Identifier{Name: "nonexistent_field"}

	_, err := m.ComputeAtBar(ctx, invalid, 0)
	if err == nil {
		t.Error("expected error for invalid source field")
	}
}

// ── BarsSinceStateManager ─────────────────────────────────────────────────────

func newBarsSince(ctx *context.Context, condExpr ast.Expression) *BarsSinceStateManager {
	return NewBarsSinceStateManager(condExpr, NewStreamingBarEvaluator(), len(ctx.Data))
}

// closeCondition returns a condition expression that evaluates to close == 1.
// When used with closesOnlyCtx, bars with Close=1 are true and Close=0 are false.
func closeCondition() *ast.Identifier {
	return &ast.Identifier{Name: "close"}
}

func TestBarsSinceStateManager_NaNBeforeFirstOccurrence(t *testing.T) {
	ctx := closesOnlyCtx(0, 0, 0, 0, 0)
	m := newBarsSince(ctx, closeCondition())

	for i := 0; i < len(ctx.Data); i++ {
		v, err := m.ComputeAtBar(ctx, i)
		if err != nil {
			t.Fatalf("bar %d: %v", i, err)
		}
		if !math.IsNaN(v) {
			t.Errorf("bar %d: condition never true — want NaN, got %.1f", i, v)
		}
	}
}

func TestBarsSinceStateManager_ZeroOnConditionBar(t *testing.T) {
	ctx := closesOnlyCtx(0, 0, 1, 0, 0)
	m := newBarsSince(ctx, closeCondition())

	v, err := m.ComputeAtBar(ctx, 2)
	if err != nil {
		t.Fatalf("bar 2: %v", err)
	}
	if v != 0 {
		t.Errorf("bar where condition is true: want 0, got %.1f", v)
	}
}

func TestBarsSinceStateManager_CountingAfterCondition(t *testing.T) {
	ctx := closesOnlyCtx(0, 1, 0, 0, 0, 0)
	m := newBarsSince(ctx, closeCondition())

	tests := []struct {
		bar     int
		wantNaN bool
		want    float64
	}{
		{0, true, 0},
		{1, false, 0},
		{2, false, 1},
		{3, false, 2},
		{4, false, 3},
		{5, false, 4},
	}

	for _, tt := range tests {
		v, err := m.ComputeAtBar(ctx, tt.bar)
		if err != nil {
			t.Fatalf("bar %d: %v", tt.bar, err)
		}
		if tt.wantNaN {
			if !math.IsNaN(v) {
				t.Errorf("bar %d: want NaN (before first occurrence), got %.1f", tt.bar, v)
			}
		} else if math.Abs(v-tt.want) > 1e-9 {
			t.Errorf("bar %d: want %.1f, got %.1f", tt.bar, tt.want, v)
		}
	}
}

func TestBarsSinceStateManager_ResetOnReoccurrence(t *testing.T) {
	ctx := closesOnlyCtx(0, 0, 1, 0, 0, 1, 0, 0)
	m := newBarsSince(ctx, closeCondition())

	tests := []struct {
		bar  int
		want float64
	}{
		{2, 0},
		{3, 1},
		{4, 2},
		{5, 0},
		{6, 1},
		{7, 2},
	}

	for _, tt := range tests {
		v, err := m.ComputeAtBar(ctx, tt.bar)
		if err != nil {
			t.Fatalf("bar %d: %v", tt.bar, err)
		}
		if math.Abs(v-tt.want) > 1e-9 {
			t.Errorf("bar %d: want %.1f, got %.1f", tt.bar, tt.want, v)
		}
	}
}

func TestBarsSinceStateManager_ConsecutiveTrueBarsEachResetToZero(t *testing.T) {
	ctx := closesOnlyCtx(0, 1, 1, 1, 0, 0)
	m := newBarsSince(ctx, closeCondition())

	for _, bar := range []int{1, 2, 3} {
		v, err := m.ComputeAtBar(ctx, bar)
		if err != nil {
			t.Fatalf("bar %d: %v", bar, err)
		}
		if v != 0 {
			t.Errorf("bar %d: consecutive true condition — want 0, got %.1f", bar, v)
		}
	}
}

func TestBarsSinceStateManager_HistoricalLookbackStableAfterAdvance(t *testing.T) {
	ctx := closesOnlyCtx(0, 0, 1, 0, 0, 0, 0)
	m := newBarsSince(ctx, closeCondition())

	anchor := 4
	first, err := m.ComputeAtBar(ctx, anchor)
	if err != nil {
		t.Fatalf("anchor bar %d: %v", anchor, err)
	}

	_, _ = m.ComputeAtBar(ctx, 6)

	second, err := m.ComputeAtBar(ctx, anchor)
	if err != nil {
		t.Fatalf("re-query anchor bar %d: %v", anchor, err)
	}
	if !floatEq(first, second) {
		t.Errorf("anchor bar %d: value changed after advance: %.1f → %.1f", anchor, first, second)
	}
}

func TestBarsSinceStateManager_FullHistoricalConsistency(t *testing.T) {
	ctx := closesOnlyCtx(0, 1, 0, 0, 1, 0, 0)
	m := newBarsSince(ctx, closeCondition())

	saved := make([]float64, len(ctx.Data))
	for i := range ctx.Data {
		v, _ := m.ComputeAtBar(ctx, i)
		saved[i] = v
	}

	for i := range ctx.Data {
		v, _ := m.ComputeAtBar(ctx, i)
		if !floatEq(v, saved[i]) {
			t.Errorf("bar %d: historical value changed on re-query: %.1f → %.1f", i, saved[i], v)
		}
	}
}

func TestBarsSinceStateManager_IdempotentRequery(t *testing.T) {
	ctx := closesOnlyCtx(0, 1, 0, 0, 0, 0)
	m := newBarsSince(ctx, closeCondition())

	_, _ = m.ComputeAtBar(ctx, 5)
	first, _ := m.ComputeAtBar(ctx, 3)

	for rep := 0; rep < 3; rep++ {
		v, _ := m.ComputeAtBar(ctx, 3)
		if !floatEq(v, first) {
			t.Errorf("rep %d: bar 3 changed on repeated query: %.1f → %.1f", rep, first, v)
		}
	}
}

func TestBarsSinceStateManager_InvalidConditionReturnsError(t *testing.T) {
	ctx := createContextWithBars(5)
	invalid := &ast.Identifier{Name: "nonexistent_field"}
	m := NewBarsSinceStateManager(invalid, NewStreamingBarEvaluator(), 5)

	_, err := m.ComputeAtBar(ctx, 0)
	if err == nil {
		t.Error("expected error for invalid condition expression")
	}
}
