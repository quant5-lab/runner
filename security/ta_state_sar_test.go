package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

// ascendingOHLCV produces n bars with monotonically rising high/low/close,
// guaranteeing uptrend initialization (high[1] > high[0]) and no reversal.
func ascendingOHLCV(n int) []context.OHLCV {
	bars := make([]context.OHLCV, n)
	for i := range bars {
		low := float64(8 + 2*i)
		bars[i] = context.OHLCV{High: low + 4, Low: low, Close: low + 2}
	}
	return bars
}

// descendingOHLCV produces n bars with monotonically falling high/low/close,
// guaranteeing downtrend initialization (high[1] < high[0]) and no reversal.
func descendingOHLCV(n int) []context.OHLCV {
	bars := make([]context.OHLCV, n)
	for i := range bars {
		high := float64(26 - 2*i)
		bars[i] = context.OHLCV{High: high, Low: high - 4, Close: high - 2}
	}
	return bars
}

func newTestSARStateManager(bars []context.OHLCV) *SARStateManager {
	return NewSARStateManager("sar_test", 0.02, 0.02, 0.2, len(bars))
}

// ── Warmup / NaN boundary ─────────────────────────────────────────────────────

func TestSARStateManager_Bar0AlwaysNaN(t *testing.T) {
	for _, name := range []string{"uptrend", "downtrend"} {
		var bars []context.OHLCV
		if name == "uptrend" {
			bars = ascendingOHLCV(5)
		} else {
			bars = descendingOHLCV(5)
		}
		t.Run(name, func(t *testing.T) {
			ctx := makeSecCtx(bars)
			m := newTestSARStateManager(bars)
			v, err := m.ComputeAtBar(ctx, nil, 0)
			if err != nil {
				t.Fatalf("bar 0: %v", err)
			}
			if !math.IsNaN(v) {
				t.Errorf("bar 0: want NaN (insufficient history), got %.6f", v)
			}
		})
	}
}

func TestSARStateManager_Bar1AndBeyondAreValid(t *testing.T) {
	for _, name := range []string{"uptrend", "downtrend"} {
		var bars []context.OHLCV
		if name == "uptrend" {
			bars = ascendingOHLCV(6)
		} else {
			bars = descendingOHLCV(6)
		}
		t.Run(name, func(t *testing.T) {
			ctx := makeSecCtx(bars)
			m := newTestSARStateManager(bars)
			for i := 1; i < len(bars); i++ {
				v, err := m.ComputeAtBar(ctx, nil, i)
				if err != nil {
					t.Fatalf("bar %d: %v", i, err)
				}
				if math.IsNaN(v) {
					t.Errorf("bar %d: want valid value, got NaN", i)
				}
			}
		})
	}
}

// ── Directional invariants ────────────────────────────────────────────────────

// TestSARStateManager_UptrendSARBelowLows verifies that in a persistent uptrend
// the SAR value stays below each bar's low — the core SAR directional invariant.
func TestSARStateManager_UptrendSARBelowLows(t *testing.T) {
	bars := ascendingOHLCV(8)
	ctx := makeSecCtx(bars)
	m := newTestSARStateManager(bars)

	for i := 2; i < len(bars); i++ {
		v, err := m.ComputeAtBar(ctx, nil, i)
		if err != nil {
			t.Fatalf("bar %d: %v", i, err)
		}
		if v >= bars[i].Low {
			t.Errorf("bar %d: uptrend SAR %.4f must be < low %.4f", i, v, bars[i].Low)
		}
	}
}

// TestSARStateManager_DowntrendSARAboveHighs verifies that in a persistent downtrend
// the SAR value stays above each bar's high — the core SAR directional invariant.
func TestSARStateManager_DowntrendSARAboveHighs(t *testing.T) {
	bars := descendingOHLCV(8)
	ctx := makeSecCtx(bars)
	m := newTestSARStateManager(bars)

	for i := 2; i < len(bars); i++ {
		v, err := m.ComputeAtBar(ctx, nil, i)
		if err != nil {
			t.Fatalf("bar %d: %v", i, err)
		}
		if v <= bars[i].High {
			t.Errorf("bar %d: downtrend SAR %.4f must be > high %.4f", i, v, bars[i].High)
		}
	}
}

// TestSARStateManager_PositivePricesYieldPositiveSAR verifies that SAR is always
// strictly positive when all price data is positive — ensures no degenerate zero/NaN output.
func TestSARStateManager_PositivePricesYieldPositiveSAR(t *testing.T) {
	for _, name := range []string{"uptrend", "downtrend"} {
		var bars []context.OHLCV
		if name == "uptrend" {
			bars = ascendingOHLCV(8)
		} else {
			bars = descendingOHLCV(8)
		}
		t.Run(name, func(t *testing.T) {
			ctx := makeSecCtx(bars)
			m := newTestSARStateManager(bars)
			for i := 1; i < len(bars); i++ {
				v, err := m.ComputeAtBar(ctx, nil, i)
				if err != nil {
					t.Fatalf("bar %d: %v", i, err)
				}
				if v <= 0 {
					t.Errorf("bar %d: positive price data must yield positive SAR, got %.6f", i, v)
				}
			}
		})
	}
}

// ── Historical access contracts ───────────────────────────────────────────────

func TestSARStateManager_HistoricalLookbackStableAfterAdvance(t *testing.T) {
	bars := ascendingOHLCV(8)
	ctx := makeSecCtx(bars)
	m := newTestSARStateManager(bars)

	anchor := 3
	first, err := m.ComputeAtBar(ctx, nil, anchor)
	if err != nil {
		t.Fatalf("anchor bar %d: %v", anchor, err)
	}

	_, _ = m.ComputeAtBar(ctx, nil, len(bars)-1)

	second, err := m.ComputeAtBar(ctx, nil, anchor)
	if err != nil {
		t.Fatalf("re-query anchor bar %d: %v", anchor, err)
	}
	if math.Abs(first-second) > 1e-9 {
		t.Errorf("anchor bar %d: value changed after advance: %.6f → %.6f", anchor, first, second)
	}
}

func TestSARStateManager_IdempotentRequery(t *testing.T) {
	bars := ascendingOHLCV(8)
	ctx := makeSecCtx(bars)
	m := newTestSARStateManager(bars)

	_, _ = m.ComputeAtBar(ctx, nil, len(bars)-1)
	first, _ := m.ComputeAtBar(ctx, nil, 3)

	for rep := 0; rep < 3; rep++ {
		v, _ := m.ComputeAtBar(ctx, nil, 3)
		if math.Abs(v-first) > 1e-9 {
			t.Errorf("rep %d: bar 3 changed on repeated query: %.6f → %.6f", rep, first, v)
		}
	}
}

func TestSARStateManager_FullHistoricalConsistency(t *testing.T) {
	bars := ascendingOHLCV(8)
	ctx := makeSecCtx(bars)
	m := newTestSARStateManager(bars)

	saved := make([]float64, len(bars))
	for i := range bars {
		v, err := m.ComputeAtBar(ctx, nil, i)
		if err != nil {
			t.Fatalf("bar %d: %v", i, err)
		}
		saved[i] = v
	}

	for i := range bars {
		v, _ := m.ComputeAtBar(ctx, nil, i)
		if !floatEq(v, saved[i]) {
			t.Errorf("bar %d: historical value changed on re-query: %.6f → %.6f", i, saved[i], v)
		}
	}
}
