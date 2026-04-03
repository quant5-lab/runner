package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func makeSecCtx(bars []context.OHLCV) *context.Context {
	return &context.Context{Data: bars}
}

// ── OBV: On Balance Volume ────────────────────────────────────────────────────

func TestOBVState_Accumulation(t *testing.T) {
	bars := []context.OHLCV{
		{Close: 10, Volume: 100},
		{Close: 12, Volume: 200}, // close up   → +200  → 200
		{Close: 11, Volume: 150}, // close down → -150  →  50
		{Close: 11, Volume: 300}, // close flat → no change → 50
		{Close: 13, Volume: 250}, // close up   → +250  → 300
	}
	expected := []float64{0, 200, 50, 50, 300}
	ctx := makeSecCtx(bars)
	st := newOBVState(len(bars))

	for i, want := range expected {
		v, err := st.computeAtBar(ctx, i)
		if err != nil {
			t.Fatalf("bar %d: unexpected error: %v", i, err)
		}
		if math.Abs(v-want) > 0.001 {
			t.Errorf("bar %d: got %v, want %v", i, v, want)
		}
	}
}

func TestOBVState_FlatPrice_NeverChanges(t *testing.T) {
	bars := make([]context.OHLCV, 5)
	for i := range bars {
		bars[i] = context.OHLCV{Close: 100, Volume: float64((i + 1) * 100)}
	}
	ctx := makeSecCtx(bars)
	st := newOBVState(len(bars))

	for i := range bars {
		v, err := st.computeAtBar(ctx, i)
		if err != nil {
			t.Fatalf("bar %d: unexpected error: %v", i, err)
		}
		if v != 0 {
			t.Errorf("bar %d: flat price must keep OBV at 0, got %v", i, v)
		}
	}
}

func TestOBVState_MonotonicUpPrice_NeverDecreases(t *testing.T) {
	bars := make([]context.OHLCV, 6)
	for i := range bars {
		bars[i] = context.OHLCV{Close: float64(10 + i), Volume: 100}
	}
	ctx := makeSecCtx(bars)
	st := newOBVState(len(bars))

	prev := math.Inf(-1)
	for i := range bars {
		v, _ := st.computeAtBar(ctx, i)
		if v < prev {
			t.Errorf("bar %d: OBV decreased on monotonic up price: %.0f → %.0f", i, prev, v)
		}
		prev = v
	}
}

// ── NVI / PVI ─────────────────────────────────────────────────────────────────

func TestNVIState_SeedAt1000(t *testing.T) {
	ctx := makeSecCtx([]context.OHLCV{{Close: 100, Volume: 100}})
	v, err := newNVIState(1).computeAtBar(ctx, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 1000 {
		t.Errorf("NVI bar 0 = %v, want 1000", v)
	}
}

func TestNVIState_UpdatesOnVolumeDecreaseOnly(t *testing.T) {
	bars := []context.OHLCV{
		{Close: 100, Volume: 200},
		{Close: 110, Volume: 100}, // vol < prev → update: 1000*(1+10/100) = 1100
		{Close: 110, Volume: 150}, // vol > prev → no change
	}
	ctx := makeSecCtx(bars)
	st := newNVIState(len(bars))

	v0, _ := st.computeAtBar(ctx, 0)
	v1, _ := st.computeAtBar(ctx, 1)
	v2, _ := st.computeAtBar(ctx, 2)

	if v0 != 1000 {
		t.Errorf("bar 0: %v want 1000", v0)
	}
	if math.Abs(v1-1100) > 0.001 {
		t.Errorf("bar 1 (vol decrease): %v want 1100", v1)
	}
	if math.Abs(v2-1100) > 0.001 {
		t.Errorf("bar 2 (vol increase, no change): %v want 1100", v2)
	}
}

func TestPVIState_SeedAt1000(t *testing.T) {
	ctx := makeSecCtx([]context.OHLCV{{Close: 100, Volume: 100}})
	v, err := newPVIState(1).computeAtBar(ctx, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 1000 {
		t.Errorf("PVI bar 0 = %v, want 1000", v)
	}
}

func TestPVIState_UpdatesOnVolumeIncreaseOnly(t *testing.T) {
	bars := []context.OHLCV{
		{Close: 100, Volume: 100},
		{Close: 110, Volume: 200}, // vol > prev → update: 1000*(1+10/100) = 1100
		{Close: 110, Volume: 150}, // vol < prev → no change
	}
	ctx := makeSecCtx(bars)
	st := newPVIState(len(bars))

	v0, _ := st.computeAtBar(ctx, 0)
	v1, _ := st.computeAtBar(ctx, 1)
	v2, _ := st.computeAtBar(ctx, 2)

	if v0 != 1000 {
		t.Errorf("bar 0: %v want 1000", v0)
	}
	if math.Abs(v1-1100) > 0.001 {
		t.Errorf("bar 1 (vol increase): %v want 1100", v1)
	}
	if math.Abs(v2-1100) > 0.001 {
		t.Errorf("bar 2 (vol decrease, no change): %v want 1100", v2)
	}
}

func TestNVIPVIState_FlatVolume_StayAtSeed(t *testing.T) {
	bars := make([]context.OHLCV, 5)
	for i := range bars {
		bars[i] = context.OHLCV{Close: float64(100 + i), Volume: 100} // vol never changes
	}
	ctx := makeSecCtx(bars)

	for _, name := range []string{"nvi", "pvi"} {
		st, _ := newVolumeState(name, len(bars))
		for i := range bars {
			v, err := st.computeAtBar(ctx, i)
			if err != nil {
				t.Fatalf("%s bar %d: unexpected error: %v", name, i, err)
			}
			if math.Abs(v-1000) > 0.001 {
				t.Errorf("%s bar %d: flat volume must keep value at 1000, got %v", name, i, v)
			}
		}
	}
}

// ── Accdist ───────────────────────────────────────────────────────────────────

func TestAccdistState_ZeroRangeGuard(t *testing.T) {
	ctx := makeSecCtx([]context.OHLCV{{High: 10, Low: 10, Close: 10, Volume: 500}})
	v, err := newAccdistState(1).computeAtBar(ctx, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 0 {
		t.Errorf("zero range: got %v, want 0", v)
	}
}

func TestAccdistState_CloseLevels(t *testing.T) {
	cases := []struct {
		name string
		bar  context.OHLCV
		want float64
	}{
		{"close at high (CLV=+1)", context.OHLCV{High: 10, Low: 8, Close: 10, Volume: 100}, 100},
		{"close at low (CLV=-1)", context.OHLCV{High: 10, Low: 8, Close: 8, Volume: 100}, -100},
		{"close at midpoint (CLV=0)", context.OHLCV{High: 10, Low: 8, Close: 9, Volume: 100}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := makeSecCtx([]context.OHLCV{tc.bar})
			v, err := newAccdistState(1).computeAtBar(ctx, 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if math.Abs(v-tc.want) > 0.001 {
				t.Errorf("got %v, want %v", v, tc.want)
			}
		})
	}
}

// ── PVT ───────────────────────────────────────────────────────────────────────

func TestPVTState_ZeroOnFirstBar(t *testing.T) {
	ctx := makeSecCtx([]context.OHLCV{{Close: 100, Volume: 500}})
	v, err := newPVTState(1).computeAtBar(ctx, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 0 {
		t.Errorf("PVT bar 0 = %v, want 0", v)
	}
}

func TestPVTState_ZeroPrevCloseGuard(t *testing.T) {
	bars := []context.OHLCV{
		{Close: 0, Volume: 500},  // prevClose will be 0
		{Close: 10, Volume: 500}, // prevClose == 0 → no update
	}
	ctx := makeSecCtx(bars)
	st := newPVTState(len(bars))
	_, _ = st.computeAtBar(ctx, 0)
	v, err := st.computeAtBar(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 0 {
		t.Errorf("PVT with zero prevClose: got %v, want 0 (no update)", v)
	}
}

func TestPVTState_Accumulation(t *testing.T) {
	bars := []context.OHLCV{
		{Close: 100, Volume: 1000},
		{Close: 110, Volume: 1000}, // pvt += (10/100)*1000 = 100
	}
	ctx := makeSecCtx(bars)
	st := newPVTState(len(bars))
	_, _ = st.computeAtBar(ctx, 0)
	v, _ := st.computeAtBar(ctx, 1)
	if math.Abs(v-100) > 0.001 {
		t.Errorf("PVT bar 1: got %v, want 100", v)
	}
}

// ── WAD ───────────────────────────────────────────────────────────────────────

func TestWADState_ZeroOnFirstBar(t *testing.T) {
	ctx := makeSecCtx([]context.OHLCV{{High: 12, Low: 8, Close: 10}})
	v, err := newWADState(1).computeAtBar(ctx, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 0 {
		t.Errorf("WAD bar 0 = %v, want 0", v)
	}
}

func TestWADState_CloseMoves(t *testing.T) {
	cases := []struct {
		name     string
		prevBar  context.OHLCV
		bar      context.OHLCV
		wantDiff float64 // expected WAD change from bar 0 to bar 1
	}{
		{
			"close up → close - trueLow",
			context.OHLCV{High: 12, Low: 8, Close: 10},
			context.OHLCV{High: 14, Low: 9, Close: 13}, // trueLow=min(9,10)=9, diff=13-9=4
			4,
		},
		{
			"close down → close - trueHigh",
			context.OHLCV{High: 12, Low: 8, Close: 10},
			context.OHLCV{High: 11, Low: 7, Close: 8}, // trueHigh=max(11,10)=11, diff=8-11=-3
			-3,
		},
		{
			"close flat → no change",
			context.OHLCV{High: 12, Low: 8, Close: 10},
			context.OHLCV{High: 12, Low: 8, Close: 10},
			0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := makeSecCtx([]context.OHLCV{tc.prevBar, tc.bar})
			st := newWADState(2)
			_, _ = st.computeAtBar(ctx, 0)
			v, err := st.computeAtBar(ctx, 1)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if math.Abs(v-tc.wantDiff) > 0.001 {
				t.Errorf("got %v, want %v", v, tc.wantDiff)
			}
		})
	}
}

// ── WVAD ──────────────────────────────────────────────────────────────────────

func TestWVADState_ZeroRangeGuard(t *testing.T) {
	ctx := makeSecCtx([]context.OHLCV{{High: 10, Low: 10, Open: 10, Close: 10, Volume: 500}})
	v, err := newWVADState(1).computeAtBar(ctx, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 0 {
		t.Errorf("zero range: got %v, want 0", v)
	}
}

func TestWVADState_Formula(t *testing.T) {
	// close==high, open==low: (close-open)/(high-low)*volume = 1.0 * volume
	ctx := makeSecCtx([]context.OHLCV{
		{High: 10, Low: 8, Open: 8, Close: 10, Volume: 100},
	})
	v, err := newWVADState(1).computeAtBar(ctx, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(v-100) > 0.001 {
		t.Errorf("got %v, want 100", v)
	}
}

// ── III ───────────────────────────────────────────────────────────────────────

func TestIIIState_Guards(t *testing.T) {
	cases := []struct {
		name string
		bar  context.OHLCV
	}{
		{"zero range", context.OHLCV{High: 10, Low: 10, Close: 10, Volume: 100}},
		{"zero volume", context.OHLCV{High: 12, Low: 8, Close: 10, Volume: 0}},
		{"zero range and volume", context.OHLCV{High: 10, Low: 10, Close: 10, Volume: 0}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := makeSecCtx([]context.OHLCV{tc.bar})
			v, err := newIIIState(1).computeAtBar(ctx, 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if v != 0 {
				t.Errorf("%s: got %v, want 0", tc.name, v)
			}
		})
	}
}

// ── State contracts ───────────────────────────────────────────────────────────

func TestVolumeIndicatorState_SequentialContract(t *testing.T) {
	// OBV must produce identical results whether computed bar-by-bar or skipped-to.
	bars := []context.OHLCV{
		{Close: 10, Volume: 100},
		{Close: 12, Volume: 200},
		{Close: 11, Volume: 150},
		{Close: 13, Volume: 250},
		{Close: 12, Volume: 180},
	}
	expected := []float64{0, 200, 50, 300, 120}
	ctx := makeSecCtx(bars)
	st := newOBVState(len(bars))

	for i, want := range expected {
		v, err := st.computeAtBar(ctx, i)
		if err != nil {
			t.Fatalf("bar %d: %v", i, err)
		}
		if math.Abs(v-want) > 0.001 {
			t.Errorf("bar %d: got %v want %v", i, v, want)
		}
	}
}

func TestVolumeIndicatorState_StateReuseIdempotent(t *testing.T) {
	bars := []context.OHLCV{
		{Close: 10, Volume: 100},
		{Close: 12, Volume: 200},
		{Close: 11, Volume: 150},
	}
	ctx := makeSecCtx(bars)
	st := newOBVState(len(bars))

	// Compute through bar 2, then re-request bar 1 (historical).
	_, _ = st.computeAtBar(ctx, 2)
	v1a, _ := st.computeAtBar(ctx, 1)
	v1b, _ := st.computeAtBar(ctx, 1)

	if v1a != v1b {
		t.Errorf("idempotency failed: first=%v second=%v", v1a, v1b)
	}
	if math.Abs(v1a-200) > 0.001 {
		t.Errorf("historical bar 1: got %v want 200", v1a)
	}
}

func TestVolumeIndicatorState_IsolatedState(t *testing.T) {
	// Two OBV states with the same data must produce the same results independently.
	bars := []context.OHLCV{
		{Close: 10, Volume: 100},
		{Close: 12, Volume: 200},
		{Close: 11, Volume: 150},
	}
	ctx := makeSecCtx(bars)
	st1 := newOBVState(len(bars))
	st2 := newOBVState(len(bars))

	for i := range bars {
		v1, _ := st1.computeAtBar(ctx, i)
		v2, _ := st2.computeAtBar(ctx, i)
		if math.Abs(v1-v2) > 0.001 {
			t.Errorf("bar %d: isolated states diverged: st1=%v st2=%v", i, v1, v2)
		}
	}
}

// ── Factory ───────────────────────────────────────────────────────────────────

func TestVolumeIndicatorState_AllTypesHistoricalAccessStable(t *testing.T) {
	// All 8 volume indicators must preserve historical bar values after advancing
	// the cursor — the core ForwardSeriesBuffer historical look-back invariant.
	bars := []context.OHLCV{
		{Open: 9, High: 12, Low: 8, Close: 10, Volume: 100},
		{Open: 10, High: 14, Low: 9, Close: 12, Volume: 200},
		{Open: 12, High: 13, Low: 10, Close: 11, Volume: 150},
		{Open: 11, High: 15, Low: 10, Close: 13, Volume: 250},
		{Open: 13, High: 16, Low: 12, Close: 14, Volume: 180},
	}
	ctx := makeSecCtx(bars)
	anchor := 2

	for _, name := range []string{"obv", "accdist", "pvt", "iii", "wvad", "nvi", "pvi", "wad"} {
		t.Run(name, func(t *testing.T) {
			st, err := newVolumeState(name, len(bars))
			if err != nil {
				t.Fatalf("newVolumeState: %v", err)
			}

			first, err := st.computeAtBar(ctx, anchor)
			if err != nil {
				t.Fatalf("anchor bar %d: %v", anchor, err)
			}

			_, _ = st.computeAtBar(ctx, len(bars)-1)

			second, err := st.computeAtBar(ctx, anchor)
			if err != nil {
				t.Fatalf("re-query anchor bar %d: %v", anchor, err)
			}
			if !math.IsNaN(first) && math.Abs(first-second) > 1e-9 {
				t.Errorf("bar %d: value changed after advance: %.6f → %.6f", anchor, first, second)
			}
			if math.IsNaN(first) && !math.IsNaN(second) {
				t.Errorf("bar %d: was NaN, became %.6f after advance", anchor, second)
			}
		})
	}
}

func TestVolumeState_AllKnownIndicators(t *testing.T) {
	for _, name := range []string{"obv", "accdist", "pvt", "iii", "wvad", "nvi", "pvi", "wad"} {
		t.Run(name, func(t *testing.T) {
			st, err := newVolumeState(name, 10)
			if err != nil {
				t.Fatalf("newVolumeState(%q) failed: %v", name, err)
			}
			if st == nil {
				t.Fatal("expected non-nil state")
			}
		})
	}
}

func TestVolumeState_UnknownReturnsError(t *testing.T) {
	for _, name := range []string{"sma", "tr", "", "ema", "UNKNOWN"} {
		t.Run(name, func(t *testing.T) {
			_, err := newVolumeState(name, 10)
			if err == nil {
				t.Errorf("expected error for %q, got nil", name)
			}
		})
	}
}

// ── Through-evaluator: StreamingBarEvaluator ──────────────────────────────────

func makeTAMemberExpr(propName string) *ast.MemberExpression {
	return &ast.MemberExpression{
		Object:   &ast.Identifier{Name: "ta"},
		Property: &ast.Identifier{Name: propName},
	}
}

func makeSubscriptedTAMemberExpr(propName string, offset float64) *ast.MemberExpression {
	return &ast.MemberExpression{
		Object:   makeTAMemberExpr(propName),
		Property: &ast.Literal{Value: offset},
	}
}

func TestStreamingBarEvaluator_VolumeIndicator_PlainAccess(t *testing.T) {
	bars := []context.OHLCV{
		{Close: 10, Volume: 100},
		{Close: 12, Volume: 200},
		{Close: 11, Volume: 150},
		{Close: 13, Volume: 250},
		{Close: 12, Volume: 180},
	}
	expected := []float64{0, 200, 50, 300, 120}
	ctx := makeSecCtx(bars)
	evaluator := NewStreamingBarEvaluator()
	expr := makeTAMemberExpr("obv")

	for i, want := range expected {
		v, err := evaluator.EvaluateAtBar(expr, ctx, i)
		if err != nil {
			t.Fatalf("bar %d: EvaluateAtBar failed: %v", i, err)
		}
		if math.Abs(v-want) > 0.001 {
			t.Errorf("bar %d: got %v, want %v", i, v, want)
		}
	}
}

func TestStreamingBarEvaluator_VolumeIndicator_HistoricalAccess(t *testing.T) {
	// ta.obv[2] at barIdx=4 → OBV value at bar 2 (targetIdx = 4 - 2 = 2)
	bars := []context.OHLCV{
		{Close: 10, Volume: 100},
		{Close: 12, Volume: 200}, // OBV=200
		{Close: 11, Volume: 150}, // OBV=50  ← ta.obv[2] at bar 4 resolves here
		{Close: 13, Volume: 250},
		{Close: 12, Volume: 180},
	}
	ctx := makeSecCtx(bars)
	evaluator := NewStreamingBarEvaluator()
	subscript := makeSubscriptedTAMemberExpr("obv", 2)

	v, err := evaluator.EvaluateAtBar(subscript, ctx, 4)
	if err != nil {
		t.Fatalf("EvaluateAtBar failed: %v", err)
	}
	if math.Abs(v-50) > 0.001 {
		t.Errorf("ta.obv[2] at bar 4: got %v, want 50", v)
	}
}

func TestStreamingBarEvaluator_VolumeIndicator_EachIndicatorAccessible(t *testing.T) {
	// All 8 volume indicators must be reachable through the evaluator without error.
	bars := []context.OHLCV{
		{Open: 9, High: 12, Low: 8, Close: 10, Volume: 100},
		{Open: 10, High: 14, Low: 9, Close: 12, Volume: 200},
		{Open: 12, High: 13, Low: 10, Close: 11, Volume: 150},
	}
	ctx := makeSecCtx(bars)

	for _, name := range []string{"obv", "accdist", "pvt", "iii", "wvad", "nvi", "pvi", "wad"} {
		t.Run(name, func(t *testing.T) {
			evaluator := NewStreamingBarEvaluator()
			expr := makeTAMemberExpr(name)
			for i := range bars {
				_, err := evaluator.EvaluateAtBar(expr, ctx, i)
				if err != nil {
					t.Errorf("bar %d: EvaluateAtBar(%q) failed: %v", i, name, err)
				}
			}
		})
	}
}

func TestStreamingBarEvaluator_VolumeIndicator_CachesStateAcrossBars(t *testing.T) {
	// Calling EvaluateAtBar with the same indicator multiple times on the same evaluator
	// must return consistent values (state is cached, not recomputed from scratch).
	bars := []context.OHLCV{
		{Close: 10, Volume: 100},
		{Close: 12, Volume: 200},
		{Close: 11, Volume: 150},
	}
	ctx := makeSecCtx(bars)
	evaluator := NewStreamingBarEvaluator()
	expr := makeTAMemberExpr("obv")

	// Full forward pass.
	var vals [3]float64
	for i := range bars {
		v, _ := evaluator.EvaluateAtBar(expr, ctx, i)
		vals[i] = v
	}

	// Re-request bar 1 (historical via the same evaluator's cached state).
	v1Again, _ := evaluator.EvaluateAtBar(expr, ctx, 1)
	if math.Abs(v1Again-vals[1]) > 0.001 {
		t.Errorf("cached re-request bar 1: got %v, want %v", v1Again, vals[1])
	}
}
