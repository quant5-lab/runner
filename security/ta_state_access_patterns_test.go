package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

var allTATypes = []struct {
	name          string
	cacheKey      string
	period        int
	warmup        int
	nanPostWarmup bool
}{
	{"SMA", "sma_close_5", 5, 4, false},
	{"EMA", "ema_close_5", 5, 4, false},
	{"RMA", "rma_close_5", 5, 4, false},
	{"RSI", "rsi_close_5", 5, 5, false},
	{"ATR", "atr_hlc_5", 5, 4, true},
	{"STDEV", "stdev_close_5", 5, 4, false},
	{"TSI", "tsi_close_5_13", 5, 17, false},
	{"SAR", "sar_0.0200_0.0200_0.2000", 5, 1, true},
}

func createTAManager(name, cacheKey string, period, capacity int, evaluator BarEvaluator) TAStateManager {
	switch name {
	case "TSI":
		return NewTSIStateManager(cacheKey, 5, 13, capacity, evaluator)
	case "SAR":
		return NewSARStateManager(cacheKey, 0.02, 0.02, 0.20, capacity)
	default:
		return NewTAStateManager(cacheKey, period, capacity, evaluator)
	}
}

func TestTAStateManager_NonSequentialAccess(t *testing.T) {
	for _, tt := range allTATypes {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createContextWithBars(40)
			m := createTAManager(tt.name, tt.cacheKey, tt.period, 40, NewStreamingBarEvaluator())
			src := &ast.Identifier{Name: "close"}

			anchor := 20
			vAnchorFirst, err := m.ComputeAtBar(ctx, src, anchor)
			if err != nil {
				t.Fatalf("anchor bar %d: %v", anchor, err)
			}

			_, err = m.ComputeAtBar(ctx, src, 35)
			if err != nil {
				t.Fatalf("advance bar 35: %v", err)
			}

			vAnchorSecond, err := m.ComputeAtBar(ctx, src, anchor)
			if err != nil {
				t.Fatalf("re-query anchor %d: %v", anchor, err)
			}

			if !floatEq(vAnchorFirst, vAnchorSecond) {
				t.Errorf("historical bar %d changed after advance: %.6f → %.6f", anchor, vAnchorFirst, vAnchorSecond)
			}
		})
	}
}

func TestTAStateManager_ColdJumpCatchUp(t *testing.T) {
	for _, tt := range allTATypes {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createContextWithBars(50)
			m := createTAManager(tt.name, tt.cacheKey, tt.period, 50, NewStreamingBarEvaluator())
			src := &ast.Identifier{Name: "close"}

			if _, err := m.ComputeAtBar(ctx, src, 40); err != nil {
				t.Fatalf("cold jump to bar 40: %v", err)
			}

			for _, bar := range []int{tt.warmup, tt.warmup + 5, 30, 40} {
				if bar > 40 {
					continue
				}
				v, err := m.ComputeAtBar(ctx, src, bar)
				if err != nil {
					t.Fatalf("bar %d: %v", bar, err)
				}
				if math.IsNaN(v) && !tt.nanPostWarmup {
					t.Errorf("bar %d: unexpected NaN after cold-jump catch-up", bar)
				}
			}
		})
	}
}

func TestTAStateManager_PartialCatchUp(t *testing.T) {
	for _, tt := range allTATypes {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createContextWithBars(30)
			m := createTAManager(tt.name, tt.cacheKey, tt.period, 30, NewStreamingBarEvaluator())
			src := &ast.Identifier{Name: "close"}

			earlyBar := tt.period + 1
			lateBar := tt.period + 15

			_, _ = m.ComputeAtBar(ctx, src, earlyBar)

			v, err := m.ComputeAtBar(ctx, src, lateBar)
			if err != nil {
				t.Fatalf("bar %d: %v", lateBar, err)
			}
			if math.IsNaN(v) && !tt.nanPostWarmup {
				t.Errorf("bar %d: unexpected NaN after partial catch-up from bar %d", lateBar, earlyBar)
			}
		})
	}
}

func TestTAStateManager_MultipleHistoricalLookbacks(t *testing.T) {
	for _, tt := range allTATypes {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createContextWithBars(50)
			m := createTAManager(tt.name, tt.cacheKey, tt.period, 50, NewStreamingBarEvaluator())
			src := &ast.Identifier{Name: "close"}

			if _, err := m.ComputeAtBar(ctx, src, 45); err != nil {
				t.Fatalf("seed bar 45: %v", err)
			}

			probePoints := []int{20, 25, 30, 35, 40, 45}
			saved := make(map[int]float64, len(probePoints))
			for _, bar := range probePoints {
				v, err := m.ComputeAtBar(ctx, src, bar)
				if err != nil {
					t.Fatalf("probe bar %d: %v", bar, err)
				}
				saved[bar] = v
			}

			for i := len(probePoints) - 1; i >= 0; i-- {
				bar := probePoints[i]
				v, err := m.ComputeAtBar(ctx, src, bar)
				if err != nil {
					t.Fatalf("re-probe bar %d: %v", bar, err)
				}
				if !floatEq(v, saved[bar]) {
					t.Errorf("bar %d: value changed on re-query %.6f → %.6f", bar, saved[bar], v)
				}
			}
		})
	}
}

func floatEq(a, b float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	return a == b
}
