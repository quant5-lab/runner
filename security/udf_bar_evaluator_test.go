package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

// ── Per-bar OHLCV input ────────────────────────────────────────────────────────

// TestUDFBarEvaluator_PerBarOHLCVInput verifies that every OHLCV field from each
// secondary bar is accessible to the UDF via ac.Context.Data[ac.Context.BarIndex].
// Each sub-test binds to a different field so the coverage is exhaustive and the
// table can grow without adding test functions.
func TestUDFBarEvaluator_PerBarOHLCVInput(t *testing.T) {
	bars := []context.OHLCV{
		{Open: 1, High: 5, Low: 0.5, Close: 2, Volume: 100},
		{Open: 2, High: 6, Low: 1.0, Close: 3, Volume: 200},
		{Open: 3, High: 7, Low: 1.5, Close: 4, Volume: 300},
		{Open: 4, High: 8, Low: 2.0, Close: 5, Volume: 400},
	}
	sec := makeSecCtx(bars)

	tests := []struct {
		name    string
		extract func(context.OHLCV) float64
		want    []float64
	}{
		{"Open", func(b context.OHLCV) float64 { return b.Open }, []float64{1, 2, 3, 4}},
		{"High", func(b context.OHLCV) float64 { return b.High }, []float64{5, 6, 7, 8}},
		{"Low", func(b context.OHLCV) float64 { return b.Low }, []float64{0.5, 1.0, 1.5, 2.0}},
		{"Close", func(b context.OHLCV) float64 { return b.Close }, []float64{2, 3, 4, 5}},
		{"Volume", func(b context.OHLCV) float64 { return b.Volume }, []float64{100, 200, 300, 400}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			extract := tt.extract
			ev := NewUDFBarEvaluator(len(bars), sec, func(ac *context.ArrowContext) float64 {
				ctx := ac.Context
				return extract(ctx.Data[ctx.BarIndex])
			})
			for bar, want := range tt.want {
				got, err := ev.EvaluateAtBar(nil, nil, bar)
				if err != nil {
					t.Fatalf("bar %d: %v", bar, err)
				}
				if got != want {
					t.Errorf("bar %d: got %.1f, want %.1f", bar, got, want)
				}
			}
		})
	}
}

// TestUDFBarEvaluator_HistoricalOffset verifies that a UDF reading
// Data[BarIndex-N] accesses the correct prior bar and returns NaN-equivalent
// when the requested bar is before history start.
func TestUDFBarEvaluator_HistoricalOffset(t *testing.T) {
	sec := closesOnlyCtx(10, 20, 30, 40)
	ev := NewUDFBarEvaluator(4, sec, func(ac *context.ArrowContext) float64 {
		ctx := ac.Context
		if ctx.BarIndex < 2 {
			return math.NaN()
		}
		return ctx.Data[ctx.BarIndex-2].Close
	})

	cases := []struct {
		bar  int
		want float64
	}{
		{0, math.NaN()},
		{1, math.NaN()},
		{2, 10},
		{3, 20},
	}
	for _, tc := range cases {
		got, err := ev.EvaluateAtBar(nil, nil, tc.bar)
		if err != nil {
			t.Fatalf("bar %d: %v", tc.bar, err)
		}
		if math.IsNaN(tc.want) {
			if !math.IsNaN(got) {
				t.Errorf("bar %d: got %.1f, want NaN", tc.bar, got)
			}
		} else if got != tc.want {
			t.Errorf("bar %d: got %.1f, want %.1f", tc.bar, got, tc.want)
		}
	}
}

// TestUDFBarEvaluator_SelfReferentialHistory verifies that local series advanced
// by AdvanceAll expose prior-bar UDF output to the next bar's invocation, matching
// Pine _var[1] semantics.  The running-sum UDF accumulates the secondary context's
// close prices; each bar's expected value is checkable by hand.
func TestUDFBarEvaluator_SelfReferentialHistory(t *testing.T) {
	// closes: 10, 20, 30, 40
	// running_sum[i] = close[i] + running_sum[i-1], running_sum[-1] = 0
	sec := closesOnlyCtx(10, 20, 30, 40)
	ev := NewUDFBarEvaluator(4, sec, func(ac *context.ArrowContext) float64 {
		ctx := ac.Context
		s := ac.GetOrCreateSeries("running_sum")
		prev := s.Get(1)
		if math.IsNaN(prev) {
			prev = 0
		}
		sum := ctx.Data[ctx.BarIndex].Close + prev
		s.Set(sum)
		return sum
	})

	want := []float64{10, 30, 60, 100}
	for bar, w := range want {
		got, err := ev.EvaluateAtBar(nil, nil, bar)
		if err != nil {
			t.Fatalf("bar %d: %v", bar, err)
		}
		if got != w {
			t.Errorf("bar %d: got %.0f, want %.0f", bar, got, w)
		}
	}
}

// ── Boundary conditions ────────────────────────────────────────────────────────

// TestUDFBarEvaluator_SingleBarContext verifies that a one-bar secondary series
// evaluates correctly.  On the only bar, AdvanceAll is skipped (series capacity
// boundary), so the evaluator must neither panic nor produce a wrong result.
func TestUDFBarEvaluator_SingleBarContext(t *testing.T) {
	sec := closesOnlyCtx(42)
	ev := NewUDFBarEvaluator(1, sec, func(ac *context.ArrowContext) float64 {
		ctx := ac.Context
		return ctx.Data[ctx.BarIndex].Close
	})
	got, err := ev.EvaluateAtBar(nil, nil, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 42 {
		t.Errorf("got %.1f, want 42", got)
	}
}

// TestUDFBarEvaluator_OutOfRangeBarIdx verifies that barIdx values outside
// [0, barCount) return (NaN, nil) without error, panic, or mutating evaluator
// state.  Each case uses an independent evaluator so side-effects cannot
// carry between sub-tests.
func TestUDFBarEvaluator_OutOfRangeBarIdx(t *testing.T) {
	cases := []struct {
		name   string
		barIdx int
	}{
		{"negative", -1},
		{"one past end", 3},
		{"far past end", 1000},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			sec := closesOnlyCtx(1, 2, 3)
			ev := NewUDFBarEvaluator(3, sec, func(ac *context.ArrowContext) float64 { return 99 })
			got, err := ev.EvaluateAtBar(nil, nil, tc.barIdx)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !math.IsNaN(got) {
				t.Errorf("barIdx %d: got %v, want NaN", tc.barIdx, got)
			}
		})
	}
}

// TestUDFBarEvaluator_ReplayDoesNotMutateSourceContext verifies that secondary
// context replay is fully isolated: the original secCtx.BarIndex is unchanged
// after EvaluateAtBar triggers a full replay sequence.
func TestUDFBarEvaluator_ReplayDoesNotMutateSourceContext(t *testing.T) {
	sec := closesOnlyCtx(1, 2, 3, 4, 5)
	originalBarIndex := sec.BarIndex

	ev := NewUDFBarEvaluator(len(sec.Data), sec, func(ac *context.ArrowContext) float64 {
		return ac.Context.Data[ac.Context.BarIndex].Close
	})
	if _, err := ev.EvaluateAtBar(nil, nil, 4); err != nil {
		t.Fatal(err)
	}

	if sec.BarIndex != originalBarIndex {
		t.Errorf("source BarIndex mutated: was %d, now %d", originalBarIndex, sec.BarIndex)
	}
}

// ── Cache invariants ───────────────────────────────────────────────────────────

// TestUDFBarEvaluator_IdempotentRequery verifies that querying the same barIdx
// multiple times after initial computation returns an identical value without
// re-invoking the UDF.
func TestUDFBarEvaluator_IdempotentRequery(t *testing.T) {
	calls := 0
	sec := closesOnlyCtx(10, 20, 30, 40, 50)
	ev := NewUDFBarEvaluator(5, sec, func(ac *context.ArrowContext) float64 {
		calls++
		ctx := ac.Context
		return ctx.Data[ctx.BarIndex].Close
	})

	targetBar := 3
	first, err := ev.EvaluateAtBar(nil, nil, targetBar)
	if err != nil {
		t.Fatalf("first call bar %d: %v", targetBar, err)
	}
	for rep := 1; rep <= 5; rep++ {
		got, err := ev.EvaluateAtBar(nil, nil, targetBar)
		if err != nil {
			t.Fatalf("rep %d: %v", rep, err)
		}
		if !floatEq(got, first) {
			t.Errorf("rep %d: value mutated %.6f → %.6f", rep, first, got)
		}
	}
	// bars 0..targetBar were each evaluated exactly once during catch-up
	wantCalls := targetBar + 1
	if calls != wantCalls {
		t.Errorf("UDF called %d times, want %d (re-evaluation on re-query)", calls, wantCalls)
	}
}

// TestUDFBarEvaluator_HistoricalAnchorStableAfterAdvance verifies that a value
// cached at an anchor bar remains unchanged after the catch-up cursor advances
// past that anchor to a later bar.
func TestUDFBarEvaluator_HistoricalAnchorStableAfterAdvance(t *testing.T) {
	closes := []float64{5, 3, 8, 1, 9, 2, 7, 4, 6, 10}
	sec := closesOnlyCtx(closes...)
	ev := NewUDFBarEvaluator(len(closes), sec, func(ac *context.ArrowContext) float64 {
		ctx := ac.Context
		return ctx.Data[ctx.BarIndex].Close
	})

	anchor := 3
	first, err := ev.EvaluateAtBar(nil, nil, anchor)
	if err != nil {
		t.Fatalf("anchor bar %d: %v", anchor, err)
	}

	if _, err := ev.EvaluateAtBar(nil, nil, len(closes)-1); err != nil {
		t.Fatalf("advance to last bar: %v", err)
	}

	for rep := 1; rep <= 3; rep++ {
		got, err := ev.EvaluateAtBar(nil, nil, anchor)
		if err != nil {
			t.Fatalf("rep %d re-query anchor: %v", rep, err)
		}
		if !floatEq(got, first) {
			t.Errorf("rep %d: anchor bar %d changed after advance %.6f → %.6f", rep, anchor, first, got)
		}
	}
}

// TestUDFBarEvaluator_FullHistoricalConsistency computes every bar in sequence,
// snapshots the results, re-queries every bar, and verifies no value drifted.
// Uses a stateful (self-referential) UDF so the test detects any cache bypass
// that would cause the series cursor to compound twice.
func TestUDFBarEvaluator_FullHistoricalConsistency(t *testing.T) {
	closes := []float64{5, 3, 8, 1, 9, 2, 7, 4, 6}
	sec := closesOnlyCtx(closes...)
	ev := NewUDFBarEvaluator(len(closes), sec, func(ac *context.ArrowContext) float64 {
		ctx := ac.Context
		s := ac.GetOrCreateSeries("acc")
		prev := s.Get(1)
		if math.IsNaN(prev) {
			prev = 0
		}
		sum := ctx.Data[ctx.BarIndex].Close + prev
		s.Set(sum)
		return sum
	})

	n := len(closes)
	saved := make([]float64, n)
	for i := 0; i < n; i++ {
		v, err := ev.EvaluateAtBar(nil, nil, i)
		if err != nil {
			t.Fatalf("bar %d: %v", i, err)
		}
		saved[i] = v
	}

	for i := 0; i < n; i++ {
		got, err := ev.EvaluateAtBar(nil, nil, i)
		if err != nil {
			t.Fatalf("re-query bar %d: %v", i, err)
		}
		if !floatEq(got, saved[i]) {
			t.Errorf("bar %d: value drifted on re-query %.6f → %.6f", i, saved[i], got)
		}
	}
}
