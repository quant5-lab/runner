package security

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func TestTAStateManager_RepeatedCallsEquivalentToSequential(t *testing.T) {
	tests := []struct {
		name          string
		createManager func() (TAStateManager, TAStateManager)
		dataSize      int
		period        int
	}{
		{
			name: "EMA_catch_up_loop",
			createManager: func() (TAStateManager, TAStateManager) {
				m1 := &EMAStateManager{
					cacheKey:   "ema_close_3",
					period:     3,
					storage:    NewSeriesStorage(8),
					multiplier: 2.0 / 4.0,
					computed:   0,
				}
				m2 := &EMAStateManager{
					cacheKey:   "ema_close_3",
					period:     3,
					storage:    NewSeriesStorage(8),
					multiplier: 2.0 / 4.0,
					computed:   0,
				}
				return m1, m2
			},
			dataSize: 8,
			period:   3,
		},
		{
			name: "RMA_catch_up_loop",
			createManager: func() (TAStateManager, TAStateManager) {
				m1 := &RMAStateManager{
					cacheKey: "rma_close_3",
					period:   3,
					storage:  NewSeriesStorage(8),
					computed: 0,
				}
				m2 := &RMAStateManager{
					cacheKey: "rma_close_3",
					period:   3,
					storage:  NewSeriesStorage(8),
					computed: 0,
				}
				return m1, m2
			},
			dataSize: 8,
			period:   3,
		},
		{
			name: "RSI_catch_up_loop",
			createManager: func() (TAStateManager, TAStateManager) {
				m1 := &RSIStateManager{
					cacheKey: "rsi_close_3",
					period:   3,
					rmaGain: &RMAStateManager{
						cacheKey: "rsi_close_3_gain",
						period:   3,
						storage:  NewSeriesStorage(8),
						computed: 0,
					},
					rmaLoss: &RMAStateManager{
						cacheKey: "rsi_close_3_loss",
						period:   3,
						storage:  NewSeriesStorage(8),
						computed: 0,
					},
					computed: 0,
				}
				m2 := &RSIStateManager{
					cacheKey: "rsi_close_3",
					period:   3,
					rmaGain: &RMAStateManager{
						cacheKey: "rsi_close_3_gain",
						period:   3,
						storage:  NewSeriesStorage(8),
						computed: 0,
					},
					rmaLoss: &RMAStateManager{
						cacheKey: "rsi_close_3_loss",
						period:   3,
						storage:  NewSeriesStorage(8),
						computed: 0,
					},
					computed: 0,
				}
				return m1, m2
			},
			dataSize: 8,
			period:   3,
		},
		{
			name: "ATR_catch_up_loop",
			createManager: func() (TAStateManager, TAStateManager) {
				return NewATRStateManager("atr_test_1", 3, 8),
					NewATRStateManager("atr_test_2", 3, 8)
			},
			dataSize: 8,
			period:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &context.Context{
				Data: []context.OHLCV{
					{Close: 100.0, High: 102.0, Low: 98.0},
					{Close: 102.0, High: 104.0, Low: 100.0},
					{Close: 99.0, High: 103.0, Low: 97.0},
					{Close: 103.0, High: 105.0, Low: 101.0},
					{Close: 101.0, High: 104.0, Low: 99.0},
					{Close: 105.0, High: 107.0, Low: 103.0},
					{Close: 104.0, High: 106.0, Low: 102.0},
					{Close: 108.0, High: 110.0, Low: 106.0},
				},
			}

			managerRepeated, managerSequential := tt.createManager()
			sourceID := &ast.Identifier{Name: "close"}

			firstValidBar := tt.period
			testBars := []int{firstValidBar + 1, firstValidBar + 2, firstValidBar + 4}

			_, _ = managerRepeated.ComputeAtBar(ctx, sourceID, firstValidBar)
			_, _ = managerRepeated.ComputeAtBar(ctx, sourceID, firstValidBar)
			_, _ = managerRepeated.ComputeAtBar(ctx, sourceID, firstValidBar)

			for _, targetBar := range testBars {
				repeatedVal, err1 := managerRepeated.ComputeAtBar(ctx, sourceID, targetBar)
				sequentialVal, err2 := managerSequential.ComputeAtBar(ctx, sourceID, targetBar)

				if err1 != nil || err2 != nil {
					t.Fatalf("ComputeAtBar(%d) failed: repeated=%v, sequential=%v", targetBar, err1, err2)
				}

				if repeatedVal != sequentialVal {
					t.Errorf("Bar %d: repeated-then-forward (%.6f) != sequential (%.6f)",
						targetBar, repeatedVal, sequentialVal)
				}
			}
		})
	}
}
