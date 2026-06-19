package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

// ── UDF fixtures ───────────────────────────────────────────────────────────────

// cumulativeSumUDF accumulates close prices using Pine nz() pre-history semantics.
// Monotonically growing output makes AdvanceAll ordering errors visible as a wrong
// cumulative total rather than a silent NaN.
func cumulativeSumUDF(ac *context.ArrowContext) float64 {
	acc := ac.GetOrCreateSeries("acc")
	prev := acc.Get(1)
	if math.IsNaN(prev) {
		prev = 0
	}
	v := ac.Context.Data[ac.Context.BarIndex].Close + prev
	acc.Set(v)
	return v
}

// directionByMomentumUDF classifies bars as +1 (rising), -1 (falling), or carries
// the prior-bar value when flat.  The flat branch is the only path that reads series
// history, so fixtures must include at least one flat bar to exercise the carry path.
func directionByMomentumUDF(ac *context.ArrowContext) float64 {
	dir := ac.GetOrCreateSeries("_dir")
	cur := ac.Context.GetClose(0)
	prev := ac.Context.GetClose(1)
	var v float64
	switch {
	case cur > prev:
		v = 1
	case cur < prev:
		v = -1
	default:
		pv := dir.Get(1)
		if math.IsNaN(pv) {
			pv = 0
		}
		v = pv
	}
	dir.Set(v)
	return v
}

// cumulativeBarCountUDF produces [1, 2, 3, …].  Output disjoint from
// directionByMomentumUDF so state bleed between evaluators surfaces as a wrong
// count rather than a coincidentally-correct value.
func cumulativeBarCountUDF(ac *context.ArrowContext) float64 {
	acc := ac.GetOrCreateSeries("_cnt")
	prev := acc.Get(1)
	if math.IsNaN(prev) {
		prev = 0
	}
	v := prev + 1
	acc.Set(v)
	return v
}

// dualSeriesUDF accumulates a running close-sum in "_sum" and a bar count in
// "_cnt", returning their product.  State bleed between the two named series
// produces a wrong product immediately, verifying GetOrCreateSeries key isolation
// holds across shared AdvanceAll calls.
func dualSeriesUDF(ac *context.ArrowContext) float64 {
	sumS := ac.GetOrCreateSeries("_sum")
	cntS := ac.GetOrCreateSeries("_cnt")

	prevSum := sumS.Get(1)
	if math.IsNaN(prevSum) {
		prevSum = 0
	}
	prevCnt := cntS.Get(1)
	if math.IsNaN(prevCnt) {
		prevCnt = 0
	}

	curSum := ac.Context.Data[ac.Context.BarIndex].Close + prevSum
	curCnt := prevCnt + 1

	sumS.Set(curSum)
	cntS.Set(curCnt)
	return curSum * curCnt
}

// ── Assertion helper ───────────────────────────────────────────────────────────

// requireBarSequence asserts ev's output against want for bars 0…len(want)-1.
// floatEq handles NaN expectations so callers pass math.NaN() for data-gap bars
// without per-caller NaN branches.
func requireBarSequence(t *testing.T, ev *UDFBarEvaluator, want []float64) {
	t.Helper()
	for bar, w := range want {
		got, err := ev.EvaluateAtBar(nil, nil, bar)
		if err != nil {
			t.Fatalf("bar %d: unexpected error: %v", bar, err)
		}
		if !floatEq(got, w) {
			t.Errorf("bar %d: got %v, want %v", bar, got, w)
		}
	}
}

// ── Per-bar OHLCV field access ─────────────────────────────────────────────────

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
			requireBarSequence(t, ev, tt.want)
		})
	}
}

func TestUDFBarEvaluator_HistoricalOffset(t *testing.T) {
	sec := closesOnlyCtx(10, 20, 30, 40)
	ev := NewUDFBarEvaluator(4, sec, func(ac *context.ArrowContext) float64 {
		ctx := ac.Context
		if ctx.BarIndex < 2 {
			return math.NaN()
		}
		return ctx.Data[ctx.BarIndex-2].Close
	})
	requireBarSequence(t, ev, []float64{math.NaN(), math.NaN(), 10, 20})
}

// ── Self-referential series state ──────────────────────────────────────────────

// TestUDFBarEvaluator_SelfReferentialHistory verifies Pine _var[1] semantics:
// AdvanceAll must expose the prior-bar UDF output to the next invocation.
// Monotonically growing output makes any ordering error immediately visible.
func TestUDFBarEvaluator_SelfReferentialHistory(t *testing.T) {
	requireBarSequence(t,
		NewUDFBarEvaluator(4, closesOnlyCtx(10, 20, 30, 40), cumulativeSumUDF),
		[]float64{10, 30, 60, 100},
	)
}

// ── Conditional self-ref: flat-bar carry ───────────────────────────────────────

// TestUDFBarEvaluator_FlatBarCarryChain verifies that a self-referential UDF
// correctly propagates the prior-bar series value through consecutive flat bars.
// Sub-tests cover three structurally distinct carry scenarios:
//   - flat at bar 0 (pre-history NaN must coerce to neutral, not crash)
//   - a single flat bar mid-sequence
//   - N>1 consecutive flat bars (every link of the carry chain must hold)
func TestUDFBarEvaluator_FlatBarCarryChain(t *testing.T) {
	tests := []struct {
		name  string
		input []float64
		want  []float64
	}{
		{
			// GetClose(1) returns 0 for OOB at bar 0, making close==prev (flat).
			// dir.Get(1)=NaN (pre-history) must coerce to neutral 0, not propagate.
			name:  "flat_at_bar0_prehistory_nan_coerces_to_neutral",
			input: []float64{0, 5, 3},
			want:  []float64{0, 1, -1},
		},
		{
			name:  "single_flat_bar_carries_prior_direction",
			input: []float64{10, 11, 9, 9, 12},
			want:  []float64{1, 1, -1, -1, 1},
		},
		{
			// Breaks if AdvanceAll is skipped or called out of order at any link.
			name:  "three_consecutive_flats_chain_each_link",
			input: []float64{10, 9, 9, 9, 11},
			want:  []float64{1, -1, -1, -1, 1},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ev := NewUDFBarEvaluator(len(tt.input), closesOnlyCtx(tt.input...), directionByMomentumUDF)
			requireBarSequence(t, ev, tt.want)
		})
	}
}

// ── Boundary conditions ────────────────────────────────────────────────────────

// TestUDFBarEvaluator_SingleBarContext verifies a one-bar secondary series.
// AdvanceAll is skipped at the series capacity boundary, so the evaluator must
// not panic or produce a wrong result.
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

func TestUDFBarEvaluator_OutOfRangeBarIdx(t *testing.T) {
	cases := []struct {
		name   string
		barIdx int
	}{
		{"negative", -1},
		{"one_past_end", 3},
		{"far_past_end", 1000},
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

// TestUDFBarEvaluator_ZeroBarContext verifies that an evaluator with zero bars
// never panics and returns (NaN, nil) for any barIdx.  series.NewSeries panics
// on capacity=0; the catch-up loop guard ensures the UDF is never invoked.
func TestUDFBarEvaluator_ZeroBarContext(t *testing.T) {
	sec := closesOnlyCtx()
	ev := NewUDFBarEvaluator(0, sec, func(ac *context.ArrowContext) float64 { return 99 })
	for _, barIdx := range []int{-1, 0, 1, 100} {
		got, err := ev.EvaluateAtBar(nil, nil, barIdx)
		if err != nil {
			t.Errorf("barIdx %d: unexpected error: %v", barIdx, err)
		}
		if !math.IsNaN(got) {
			t.Errorf("barIdx %d: got %v, want NaN", barIdx, got)
		}
	}
}

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

// TestUDFBarEvaluator_IdempotentRequery verifies that re-querying a bar returns
// the cached value without re-invoking the UDF.  The call counter must equal
// barIdx+1 regardless of how many times that bar is re-queried.
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
	if calls != targetBar+1 {
		t.Errorf("UDF called %d times, want %d", calls, targetBar+1)
	}
}

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

// TestUDFBarEvaluator_FullHistoricalConsistency snapshots every bar with a
// stateful UDF, then re-queries every bar and confirms no value drifted.
// A stateful UDF is required: if EvaluateAtBar re-invokes on a cache hit,
// the series cursor advances twice and the cumulative total diverges.
func TestUDFBarEvaluator_FullHistoricalConsistency(t *testing.T) {
	closes := []float64{5, 3, 8, 1, 9, 2, 7, 4, 6}
	sec := closesOnlyCtx(closes...)
	ev := NewUDFBarEvaluator(len(closes), sec, cumulativeSumUDF)

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

func TestUDFBarEvaluator_StateIntactAfterOutOfRangeQuery(t *testing.T) {
	ev := NewUDFBarEvaluator(3, closesOnlyCtx(10, 20, 30), cumulativeSumUDF)

	oob, err := ev.EvaluateAtBar(nil, nil, 999)
	if err != nil {
		t.Fatalf("OOB query: unexpected error: %v", err)
	}
	if !math.IsNaN(oob) {
		t.Errorf("OOB query: got %v, want NaN", oob)
	}

	requireBarSequence(t, ev, []float64{10, 30, 60})
}

func TestUDFBarEvaluator_SkipToMidBarFirstQueryWithStatefulUDF(t *testing.T) {
	sec := closesOnlyCtx(10, 20, 30, 40, 50)
	ev := NewUDFBarEvaluator(len(sec.Data), sec, cumulativeSumUDF)

	got2, err := ev.EvaluateAtBar(nil, nil, 2)
	if err != nil || got2 != 60 {
		t.Fatalf("first query bar 2: got %v err %v, want 60 nil", got2, err)
	}

	got0, err := ev.EvaluateAtBar(nil, nil, 0)
	if err != nil || got0 != 10 {
		t.Errorf("cache hit bar 0: got %v err %v, want 10 nil", got0, err)
	}

	// State must continue from bar 2; the catch-up must not restart from bar 0.
	got4, err := ev.EvaluateAtBar(nil, nil, 4)
	if err != nil || got4 != 150 {
		t.Errorf("advance to bar 4: got %v err %v, want 150 nil", got4, err)
	}

	requireBarSequence(t, ev, []float64{10, 30, 60, 100, 150})
}

// ── Evaluator isolation ────────────────────────────────────────────────────────

// TestUDFBarEvaluator_EvaluatorIsolation verifies that two evaluators built from
// the same secCtx maintain independent ArrowContext histories.  Bars are queried
// in interleaved order so that advancing one evaluator cannot advance the other.
func TestUDFBarEvaluator_EvaluatorIsolation(t *testing.T) {
	sec := closesOnlyCtx(10, 11, 9, 9, 12)
	dirEval := NewUDFBarEvaluator(len(sec.Data), sec, directionByMomentumUDF)
	cntEval := NewUDFBarEvaluator(len(sec.Data), sec, cumulativeBarCountUDF)

	wantDir := []float64{1, 1, -1, -1, 1}
	wantCnt := []float64{1, 2, 3, 4, 5}

	for bar := range sec.Data {
		gotDir, errD := dirEval.EvaluateAtBar(nil, nil, bar)
		gotCnt, errC := cntEval.EvaluateAtBar(nil, nil, bar)
		if errD != nil || errC != nil {
			t.Fatalf("bar %d: dirEval err=%v cntEval err=%v", bar, errD, errC)
		}
		if gotDir != wantDir[bar] {
			t.Errorf("dirEval bar %d: got %v, want %v", bar, gotDir, wantDir[bar])
		}
		if gotCnt != wantCnt[bar] {
			t.Errorf("cntEval bar %d: got %v, want %v", bar, gotCnt, wantCnt[bar])
		}
	}
}

// TestUDFBarEvaluator_MultipleNamedSeriesInUDF verifies that named series within
// a single UDF maintain independent histories across shared AdvanceAll calls.
func TestUDFBarEvaluator_MultipleNamedSeriesInUDF(t *testing.T) {
	requireBarSequence(t,
		NewUDFBarEvaluator(4, closesOnlyCtx(10, 20, 30, 40), dualSeriesUDF),
		[]float64{10, 60, 180, 400},
	)
}

// ── NaN propagation through series advance ─────────────────────────────────────

// TestUDFBarEvaluator_NaNPropagatedThroughSeriesAdvance verifies that NaN stored
// in a series slot survives AdvanceAll and is readable as NaN on the next bar —
// it must not be coerced to zero or any other sentinel by the advance mechanism.
func TestUDFBarEvaluator_NaNPropagatedThroughSeriesAdvance(t *testing.T) {
	ev := NewUDFBarEvaluator(4, closesOnlyCtx(5, 0, 7, 8), func(ac *context.ArrowContext) float64 {
		s := ac.GetOrCreateSeries("_v")
		cl := ac.Context.Data[ac.Context.BarIndex].Close
		if cl == 0 {
			s.Set(math.NaN())
			return math.NaN()
		}
		prev := s.Get(1)
		s.Set(cl)
		return prev
	})
	requireBarSequence(t, ev, []float64{math.NaN(), math.NaN(), math.NaN(), 7})
}
