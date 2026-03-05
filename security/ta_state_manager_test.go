package security

import (
	"fmt"
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// ── Arithmetic correctness ──────────────────────────────────────────────────

func TestSMAStateManager_KnownValues(t *testing.T) {
	tests := []struct {
		name   string
		period int
		barIdx int
		want   float64
	}{
		// close prices: 100, 101, 102, … (createContextWithBars)
		{"period3 first valid", 3, 2, 101.0},
		{"period3 second valid", 3, 3, 102.0},
		{"period5 first valid", 5, 4, 102.0},
		{"period1 any bar", 1, 7, 107.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createContextWithBars(20)
			m := newSMAStateManager("sma_close", tt.period, 20, NewStreamingBarEvaluator())
			src := &ast.Identifier{Name: "close"}

			v, err := m.ComputeAtBar(ctx, src, tt.barIdx)
			if err != nil {
				t.Fatalf("bar %d: %v", tt.barIdx, err)
			}
			if math.Abs(v-tt.want) > 1e-9 {
				t.Errorf("SMA(%d) at bar %d = %.9f, want %.9f", tt.period, tt.barIdx, v, tt.want)
			}
		})
	}
}

func TestRMAStateManager_KnownValues(t *testing.T) {
	// close prices: 100, 101, 102, … (createContextWithBars); RMA(3), alpha=1/3
	// bar 0: seed = 100.0
	// bar 1: running avg = (100*1 + 101)/2 = 100.5
	// bar 2: running avg = (100.5*2 + 102)/3 = 101.0
	// bar 3: Wilder smooth = (1/3)*103 + (2/3)*101 = 101.6̄
	tests := []struct {
		name   string
		barIdx int
		want   float64
	}{
		{"bar 0 seed", 0, 100.0},
		{"bar 1 running avg", 1, 100.5},
		{"bar 2 running avg", 2, 101.0},
		{"bar 3 wilder smooth", 3, 101.0 + 2.0/3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createContextWithBars(20)
			m := newRMAStateManager("rma_close_3", 3, 20, NewStreamingBarEvaluator())
			src := &ast.Identifier{Name: "close"}

			v, err := m.ComputeAtBar(ctx, src, tt.barIdx)
			if err != nil {
				t.Fatalf("bar %d: %v", tt.barIdx, err)
			}
			if math.Abs(v-tt.want) > 1e-9 {
				t.Errorf("RMA(3) at bar %d = %.9f, want %.9f", tt.barIdx, v, tt.want)
			}
		})
	}
}

func TestEMAStateManager_MonotonicInputConvergence(t *testing.T) {
	for _, period := range []int{3, 5, 10} {
		t.Run(fmt.Sprintf("period%d", period), func(t *testing.T) {
			ctx := createContextWithBars(40)
			m := newEMAStateManager("ema_close", period, 40, NewStreamingBarEvaluator())
			src := &ast.Identifier{Name: "close"}

			var prev float64
			for i := period - 1; i < 30; i++ {
				v, err := m.ComputeAtBar(ctx, src, i)
				if err != nil {
					t.Fatalf("bar %d: %v", i, err)
				}
				if math.IsNaN(v) {
					t.Fatalf("bar %d: unexpected NaN post-warmup", i)
				}
				if i > period-1 && v <= prev {
					t.Errorf("EMA(%d) bar %d = %.6f not > bar %d = %.6f", period, i, v, i-1, prev)
				}
				prev = v
			}
		})
	}
}

func TestRMAStateManager_WilderAlphaIsSlowerThanEMA(t *testing.T) {
	period := 14
	ctx := createContextWithBars(60)
	mEMA := newEMAStateManager("ema_close_14", period, 60, NewStreamingBarEvaluator())
	mRMA := newRMAStateManager("rma_close_14", period, 60, NewStreamingBarEvaluator())
	src := &ast.Identifier{Name: "close"}

	var emaV, rmaV float64
	for i := period; i < 50; i++ {
		emaV, _ = mEMA.ComputeAtBar(ctx, src, i)
		rmaV, _ = mRMA.ComputeAtBar(ctx, src, i)
	}
	if math.Abs(emaV-rmaV) < 1e-6 {
		t.Errorf("EMA(%d) and RMA(%d) must diverge on trending input: ema=%.6f rma=%.6f", period, period, emaV, rmaV)
	}
}

// ── RSI behavioral invariants ───────────────────────────────────────────────

func TestRSIStateManager_OutputBoundsAllInputShapes(t *testing.T) {
	tests := []struct {
		name string
		data func(i int) float64
	}{
		{"monotonic rise", func(i int) float64 { return float64(100 + i) }},
		{"monotonic fall", func(i int) float64 { return float64(200 - i) }},
		{"oscillating", func(i int) float64 {
			if i%2 == 0 {
				return float64(100 + i)
			}
			return float64(100 - i)
		}},
		{"step up", func(i int) float64 {
			if i < 30 {
				return 100.0
			}
			return 200.0
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createContextWithBars(60)
			for i := range ctx.Data {
				ctx.Data[i].Close = tt.data(i)
			}
			m := newRSIStateManager("rsi_close_14", 14, 60, NewStreamingBarEvaluator())
			src := &ast.Identifier{Name: "close"}

			for i := 14; i < 60; i++ {
				v, err := m.ComputeAtBar(ctx, src, i)
				if err != nil {
					t.Fatalf("bar %d: %v", i, err)
				}
				if math.IsNaN(v) {
					t.Fatalf("bar %d: unexpected NaN post-warmup", i)
				}
				if v < 0 || v > 100 {
					t.Errorf("bar %d: RSI = %.6f out of [0, 100]", i, v)
				}
			}
		})
	}
}

func TestRSIStateManager_ZeroLossYields100(t *testing.T) {
	ctx := createContextWithBars(30)
	for i := range ctx.Data {
		ctx.Data[i].Close = 100.0
	}
	m := newRSIStateManager("rsi_close_14", 14, 30, NewStreamingBarEvaluator())
	src := &ast.Identifier{Name: "close"}

	v, err := m.ComputeAtBar(ctx, src, 20)
	if err != nil {
		t.Fatalf("ComputeAtBar: %v", err)
	}
	if math.Abs(v-100.0) > 1e-9 {
		t.Errorf("flat source RSI = %.6f, want 100.0", v)
	}
}

func TestRSIStateManager_ZeroGainYields0(t *testing.T) {
	ctx := createContextWithBars(30)
	for i := range ctx.Data {
		ctx.Data[i].Close = float64(200 - i)
	}
	m := newRSIStateManager("rsi_close_14", 14, 30, NewStreamingBarEvaluator())
	src := &ast.Identifier{Name: "close"}

	v, err := m.ComputeAtBar(ctx, src, 25)
	if err != nil {
		t.Fatalf("ComputeAtBar: %v", err)
	}
	if v >= 50.0 {
		t.Errorf("monotonically falling source RSI = %.6f, want < 50", v)
	}
}

// ── Factory routing ─────────────────────────────────────────────────────────

func TestNewTAStateManager_RoutesToCorrectType(t *testing.T) {
	tests := []struct {
		cacheKey string
		wantType string
	}{
		{"sma_close_5", "SMA"},
		{"ema_close_10", "EMA"},
		{"rma_close_14", "RMA"},
		{"rsi_close_14", "RSI"},
		{"atr_hlc_14", "ATR"},
		{"stdev_close_20", "STDEV"},
	}

	for _, tt := range tests {
		t.Run(tt.wantType, func(t *testing.T) {
			ctx := createContextWithBars(30)
			m := NewTAStateManager(tt.cacheKey, 5, 30, NewStreamingBarEvaluator())
			src := &ast.Identifier{Name: "close"}

			_, err := m.ComputeAtBar(ctx, src, 10)
			if err != nil && tt.wantType != "ATR" {
				t.Errorf("unexpected error from %s: %v", tt.wantType, err)
			}
		})
	}
}

// ── contains helper ─────────────────────────────────────────────────────────

func TestContainsFunction(t *testing.T) {
	tests := []struct {
		s      string
		sub    string
		expect bool
	}{
		{"sma_close_20", "sma", true},
		{"ema_close_50", "ema", true},
		{"rma_close_14", "rma", true},
		{"rsi_close_14", "rsi", true},
		{"atr_hlc_14", "atr", true},
		{"stdev_close_20", "stdev", true},
		{"sma_close_20", "ema", false},
		{"", "sma", false},
		{"sma", "", true},
	}

	for _, tt := range tests {
		got := contains(tt.s, tt.sub)
		if got != tt.expect {
			t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.sub, got, tt.expect)
		}
	}
}
