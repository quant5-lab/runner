package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

var allTATypes = []struct {
	name     string
	cacheKey string
	period   int
}{
	{"SMA", "sma_close_5", 5},
	{"EMA", "ema_close_5", 5},
	{"RMA", "rma_close_5", 5},
	{"RSI", "rsi_close_5", 5},
	{"ATR", "atr_hlc_5", 5},
	{"STDEV", "stdev_close_5", 5},
}

// TestTAStateManager_NonSequentialAccess verifies that requesting a historical
// bar after advancing the cursor returns the same value as the original query.
// This is the core ForwardSeriesBuffer historical-look-back invariant.
func TestTAStateManager_NonSequentialAccess(t *testing.T) {
	for _, tt := range allTATypes {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createContextWithBars(40)
			m := NewTAStateManager(tt.cacheKey, tt.period, 40, NewStreamingBarEvaluator())
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

// TestTAStateManager_ColdJumpCatchUp verifies that jumping directly to a far
// bar triggers the full catch-up loop so all intermediate bars become accessible.
func TestTAStateManager_ColdJumpCatchUp(t *testing.T) {
	for _, tt := range allTATypes {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createContextWithBars(50)
			m := NewTAStateManager(tt.cacheKey, tt.period, 50, NewStreamingBarEvaluator())
			src := &ast.Identifier{Name: "close"}

			if _, err := m.ComputeAtBar(ctx, src, 40); err != nil {
				t.Fatalf("cold jump to bar 40: %v", err)
			}

			warmup := tt.period - 1
			if tt.name == "RSI" {
				warmup = tt.period
			}
			for _, bar := range []int{warmup, warmup + 5, 30, 40} {
				if bar > 40 {
					continue
				}
				v, err := m.ComputeAtBar(ctx, src, bar)
				if err != nil {
					t.Fatalf("bar %d: %v", bar, err)
				}
				if math.IsNaN(v) && tt.name != "ATR" {
					t.Errorf("bar %d: unexpected NaN after cold-jump catch-up", bar)
				}
			}
		})
	}
}

// TestTAStateManager_PartialCatchUp verifies that a partial advance followed by
// a later query correctly computes all intermediate state without gaps.
func TestTAStateManager_PartialCatchUp(t *testing.T) {
	for _, tt := range allTATypes {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createContextWithBars(30)
			m := NewTAStateManager(tt.cacheKey, tt.period, 30, NewStreamingBarEvaluator())
			src := &ast.Identifier{Name: "close"}

			earlyBar := tt.period + 1
			lateBar := tt.period + 15

			_, _ = m.ComputeAtBar(ctx, src, earlyBar)

			v, err := m.ComputeAtBar(ctx, src, lateBar)
			if err != nil {
				t.Fatalf("bar %d: %v", lateBar, err)
			}
			if math.IsNaN(v) && tt.name != "ATR" {
				t.Errorf("bar %d: unexpected NaN after partial catch-up from bar %d", lateBar, earlyBar)
			}
		})
	}
}

// TestTAStateManager_MultipleHistoricalLookbacks verifies that after a single
// full catch-up, all prior bars can be queried in arbitrary order and return
// consistent values.
func TestTAStateManager_MultipleHistoricalLookbacks(t *testing.T) {
	for _, tt := range allTATypes {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createContextWithBars(50)
			m := NewTAStateManager(tt.cacheKey, tt.period, 50, NewStreamingBarEvaluator())
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
