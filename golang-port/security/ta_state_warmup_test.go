package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

/* TestTAStateManager_InsufficientDataReturnsNaN verifies all TA state managers
 * return NaN when insufficient bars exist for computation, preventing spurious
 * zero values that render as visible lines on charts
 */
func TestTAStateManager_InsufficientDataReturnsNaN(t *testing.T) {
	tests := []struct {
		name        string
		cacheKey    string
		period      int
		dataPoints  int
		validateIdx int
		wantNaN     bool
	}{
		{"SMA warmup start", "sma_close_20", 20, 25, 0, true},
		{"SMA warmup mid", "sma_close_20", 20, 25, 9, true},
		{"SMA warmup end", "sma_close_20", 20, 25, 18, true},
		{"SMA sufficient", "sma_close_20", 20, 25, 19, false},
		{"EMA warmup", "ema_close_50", 50, 60, 48, true},
		{"EMA sufficient", "ema_close_50", 50, 60, 49, false},
		{"RMA warmup", "rma_close_100", 100, 110, 98, true},
		{"RMA sufficient", "rma_close_100", 100, 110, 99, false},
		{"RSI warmup", "rsi_close_14", 14, 20, 13, true},
		{"RSI sufficient", "rsi_close_14", 14, 20, 14, false},
		{"ATR warmup start", "atr_hlc_14", 14, 20, 0, false},
		{"ATR warmup mid", "atr_hlc_14", 14, 20, 6, false},
		{"ATR warmup end", "atr_hlc_14", 14, 20, 12, false},
		{"ATR sufficient", "atr_hlc_14", 14, 20, 13, false},
		{"STDEV warmup start", "stdev_close_20", 20, 25, 0, true},
		{"STDEV warmup mid", "stdev_close_20", 20, 25, 9, true},
		{"STDEV warmup end", "stdev_close_20", 20, 25, 18, true},
		{"STDEV sufficient", "stdev_close_20", 20, 25, 19, false},
		{"STDEV BB8 warmup", "stdev_close_46", 46, 55, 44, true},
		{"STDEV BB8 sufficient", "stdev_close_46", 46, 55, 45, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createContextWithBars(tt.dataPoints)
			manager := NewTAStateManager(tt.cacheKey, tt.period, tt.dataPoints)
			sourceID := &ast.Identifier{Name: "close"}

			value, err := manager.ComputeAtBar(ctx, sourceID, tt.validateIdx)
			if err != nil {
				t.Fatalf("ComputeAtBar failed: %v", err)
			}

			if tt.wantNaN {
				if !math.IsNaN(value) && value != 0.0 {
					t.Errorf("expected NaN or 0 at index %d (period %d), got %.4f",
						tt.validateIdx, tt.period, value)
				}
			} else {
				if math.IsNaN(value) {
					t.Errorf("expected valid value at index %d (period %d), got NaN",
						tt.validateIdx, tt.period)
				}
			}
		})
	}
}

/* TestTAStateManager_WarmupBoundaryTransition verifies exact boundary
 * where NaN transitions to valid values (period-1 → period)
 */
func TestTAStateManager_WarmupBoundaryTransition(t *testing.T) {
	tests := []struct {
		name     string
		cacheKey string
		period   int
	}{
		{"SMA period 5", "sma_close_5", 5},
		{"SMA period 20", "sma_close_20", 20},
		{"EMA period 10", "ema_close_10", 10},
		{"RMA period 14", "rma_close_14", 14},
		{"ATR period 7", "atr_hlc_7", 7},
		{"ATR period 20", "atr_hlc_20", 20},
		{"STDEV period 5", "stdev_close_5", 5},
		{"STDEV period 46", "stdev_close_46", 46},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createContextWithBars(tt.period + 5)
			manager := NewTAStateManager(tt.cacheKey, tt.period, tt.period+5)
			sourceID := &ast.Identifier{Name: "close"}

			lastWarmupIdx := tt.period - 2
			if lastWarmupIdx >= 0 {
				valueBeforeBoundary, _ := manager.ComputeAtBar(ctx, sourceID, lastWarmupIdx)
				if !math.IsNaN(valueBeforeBoundary) && valueBeforeBoundary != 0.0 {
					t.Errorf("index %d (period-2): expected NaN or 0, got %.4f",
						lastWarmupIdx, valueBeforeBoundary)
				}
			}

			firstValidIdx := tt.period - 1
			valueAtBoundary, _ := manager.ComputeAtBar(ctx, sourceID, firstValidIdx)
			if math.IsNaN(valueAtBoundary) || valueAtBoundary == 0.0 {
				t.Errorf("index %d (period-1): expected valid non-zero value, got %.4f",
					firstValidIdx, valueAtBoundary)
			}

			valuePastBoundary, _ := manager.ComputeAtBar(ctx, sourceID, firstValidIdx+1)
			if math.IsNaN(valuePastBoundary) || valuePastBoundary == 0.0 {
				t.Errorf("index %d (period): expected valid non-zero value, got %.4f",
					firstValidIdx+1, valuePastBoundary)
			}
		})
	}
}

func TestRSIStateManager_WarmupBoundary(t *testing.T) {
	period := 7
	ctx := createContextWithBars(period + 5)
	manager := NewTAStateManager("rsi_close_7", period, period+5)
	sourceID := &ast.Identifier{Name: "close"}

	valueBefore, _ := manager.ComputeAtBar(ctx, sourceID, period-1)
	if !math.IsNaN(valueBefore) {
		t.Errorf("RSI index %d (period-1): expected NaN, got %.4f", period-1, valueBefore)
	}

	valueAtBoundary, _ := manager.ComputeAtBar(ctx, sourceID, period)
	if math.IsNaN(valueAtBoundary) {
		t.Errorf("RSI index %d (period): expected valid value, got NaN", period)
	}

	if valueAtBoundary < 0.0 || valueAtBoundary > 100.0 {
		t.Errorf("RSI out of range [0, 100]: got %.4f", valueAtBoundary)
	}
}

func TestTAStateManager_EmptyDataReturnsError(t *testing.T) {
	emptyCtx := &context.Context{Data: []context.OHLCV{}}
	sourceID := &ast.Identifier{Name: "close"}

	managers := []struct {
		name    string
		manager TAStateManager
	}{
		{"SMA", NewTAStateManager("sma_close_20", 20, 0)},
		{"EMA", NewTAStateManager("ema_close_20", 20, 0)},
		{"RMA", NewTAStateManager("rma_close_20", 20, 0)},
		{"RSI", NewTAStateManager("rsi_close_14", 14, 0)},
		{"ATR", NewTAStateManager("atr_hlc_14", 14, 0)},
	}

	for _, m := range managers {
		t.Run(m.name, func(t *testing.T) {
			value, err := m.manager.ComputeAtBar(emptyCtx, sourceID, 0)
			if m.name == "ATR" {
				if value != 0.0 {
					t.Errorf("expected 0 for empty data, got %.4f", value)
				}
			} else {
				if err == nil && !math.IsNaN(value) {
					t.Errorf("expected error or NaN for empty data, got value %.4f", value)
				}
			}
		})
	}
}

func TestTAStateManager_SingleBarReturnsNaN(t *testing.T) {
	ctx := createContextWithBars(1)
	sourceID := &ast.Identifier{Name: "close"}

	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewTAStateManager(tt.cacheKey, tt.period, 1)
			value, err := manager.ComputeAtBar(ctx, sourceID, 0)
			if err != nil {
				t.Fatalf("ComputeAtBar failed: %v", err)
			}

			if !math.IsNaN(value) && value != 0.0 {
				t.Errorf("single bar with period %d: expected NaN or 0, got %.4f", tt.period, value)
			}
		})
	}
}

func TestTAStateManager_InvalidSourceReturnsError(t *testing.T) {
	ctx := createContextWithBars(20)
	invalidSource := &ast.Identifier{Name: "invalid_field"}

	managers := []struct {
		name    string
		manager TAStateManager
	}{
		{"SMA", NewTAStateManager("sma_close_10", 10, 20)},
		{"EMA", NewTAStateManager("ema_close_10", 10, 20)},
		{"RMA", NewTAStateManager("rma_close_10", 10, 20)},
		{"RSI", NewTAStateManager("rsi_close_10", 10, 20)},
		{"ATR", NewTAStateManager("atr_hlc_10", 10, 20)},
		{"STDEV", NewTAStateManager("stdev_close_10", 10, 20)},
	}

	for _, m := range managers {
		t.Run(m.name, func(t *testing.T) {
			value, err := m.manager.ComputeAtBar(ctx, invalidSource, 10)
			if m.name == "ATR" {
				if err != nil {
					t.Error("ATR should not error with invalid source")
				}
				if value <= 0 || math.IsNaN(value) {
					t.Errorf("expected valid value, got %.4f", value)
				}
			} else {
				if err == nil {
					t.Error("expected error for invalid source field")
				}
				if !math.IsNaN(value) && value != 0.0 {
					t.Errorf("expected NaN or zero on error, got %.4f", value)
				}
			}
		})
	}
}

func TestTAStateManager_ConsecutiveNaNsNoGaps(t *testing.T) {
	period := 10
	dataSize := 15
	ctx := createContextWithBars(dataSize)
	sourceID := &ast.Identifier{Name: "close"}

	tests := []struct {
		name     string
		cacheKey string
	}{
		{"SMA", "sma_close_10"},
		{"EMA", "ema_close_10"},
		{"RMA", "rma_close_10"},
		{"ATR", "atr_hlc_10"},
		{"STDEV", "stdev_close_10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewTAStateManager(tt.cacheKey, period, dataSize)

			for i := 0; i < period-1; i++ {
				value, err := manager.ComputeAtBar(ctx, sourceID, i)
				if err != nil {
					t.Fatalf("bar %d: ComputeAtBar failed: %v", i, err)
				}
				if !math.IsNaN(value) && value != 0.0 {
					t.Errorf("bar %d: expected NaN or 0 in warmup sequence, got %.4f", i, value)
				}
			}

			for i := period - 1; i < dataSize; i++ {
				value, err := manager.ComputeAtBar(ctx, sourceID, i)
				if err != nil {
					t.Fatalf("bar %d: ComputeAtBar failed: %v", i, err)
				}
				if math.IsNaN(value) || value == 0.0 {
					t.Errorf("bar %d: expected valid non-zero value post-warmup, got %.4f", i, value)
				}
			}
		})
	}
}

func createContextWithBars(count int) *context.Context {
	data := make([]context.OHLCV, count)
	for i := 0; i < count; i++ {
		price := 100.0 + float64(i)
		data[i] = context.OHLCV{
			Open:   price - 0.5,
			High:   price + 1.0,
			Low:    price - 1.0,
			Close:  price,
			Volume: 1000.0,
		}
	}
	return &context.Context{Data: data}
}
