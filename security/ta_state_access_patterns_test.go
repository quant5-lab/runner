package security

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func TestTAStateManager_ArbitraryAccessOrderContract(t *testing.T) {
	tests := []struct {
		name          string
		createManager func() TAStateManager
		dataSize      int
	}{
		{
			name: "EMA",
			createManager: func() TAStateManager {
				return &EMAStateManager{
					cacheKey:   "ema_close_3",
					period:     3,
					storage:    NewSeriesStorage(10),
					multiplier: 2.0 / 4.0,
					computed:   0,
				}
			},
			dataSize: 10,
		},
		{
			name: "RMA",
			createManager: func() TAStateManager {
				return &RMAStateManager{
					cacheKey: "rma_close_3",
					period:   3,
					storage:  NewSeriesStorage(10),
					computed: 0,
				}
			},
			dataSize: 10,
		},
		{
			name: "ATR",
			createManager: func() TAStateManager {
				return NewATRStateManager("atr_test", 3, 10)
			},
			dataSize: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := buildIncrementingOHLCV(tt.dataSize, 100.0, 2.0)
			manager := tt.createManager()
			sourceID := &ast.Identifier{Name: "close"}

			forward, _ := manager.ComputeAtBar(ctx, sourceID, 7)
			backward, _ := manager.ComputeAtBar(ctx, sourceID, 5)
			forward2, _ := manager.ComputeAtBar(ctx, sourceID, 6)

			if forward == backward {
				t.Errorf("Forward access (bar 7) should differ from backward (bar 5), both = %f", forward)
			}

			if forward2 == forward {
				t.Errorf("Bar 6 should differ from bar 7, both = %f", forward)
			}

			if forward2 == backward {
				t.Errorf("Bar 6 should differ from bar 5, both = %f", forward2)
			}
		})
	}
}

func TestTAStateManager_IdempotentRepeatedAccess(t *testing.T) {
	tests := []struct {
		name          string
		createManager func() TAStateManager
	}{
		{
			name: "EMA_repeated_access_idempotent",
			createManager: func() TAStateManager {
				return &EMAStateManager{
					cacheKey:   "ema_close_2",
					period:     2,
					storage:    NewSeriesStorage(5),
					multiplier: 2.0 / 3.0,
					computed:   0,
				}
			},
		},
		{
			name: "RMA_repeated_access_idempotent",
			createManager: func() TAStateManager {
				return &RMAStateManager{
					cacheKey: "rma_close_2",
					period:   2,
					storage:  NewSeriesStorage(5),
					computed: 0,
				}
			},
		},
		{
			name: "ATR_repeated_access_idempotent",
			createManager: func() TAStateManager {
				return NewATRStateManager("atr_test", 2, 5)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := buildIncrementingOHLCV(5, 100.0, 2.0)
			manager := tt.createManager()
			sourceID := &ast.Identifier{Name: "close"}

			value1, _ := manager.ComputeAtBar(ctx, sourceID, 3)
			value2, _ := manager.ComputeAtBar(ctx, sourceID, 3)
			value3, _ := manager.ComputeAtBar(ctx, sourceID, 3)

			if value1 != value2 || value2 != value3 {
				t.Errorf("Repeated access not idempotent: %f, %f, %f", value1, value2, value3)
			}
		})
	}
}

func TestEMAStateManager_MonotonicIncreasingData(t *testing.T) {
	ctx := buildIncrementingOHLCV(10, 100.0, 2.0)
	manager := &EMAStateManager{
		cacheKey:   "ema_close_3",
		period:     3,
		storage:    NewSeriesStorage(10),
		multiplier: 2.0 / 4.0,
		computed:   0,
	}
	sourceID := &ast.Identifier{Name: "close"}

	var prevValue float64
	for bar := 2; bar < 10; bar++ {
		value, _ := manager.ComputeAtBar(ctx, sourceID, bar)
		if bar > 2 && value <= prevValue {
			t.Errorf("EMA should increase monotonically for increasing data: bar %d = %f, bar %d = %f",
				bar-1, prevValue, bar, value)
		}
		prevValue = value
	}
}

func TestRMAStateManager_MonotonicIncreasingData(t *testing.T) {
	ctx := buildIncrementingOHLCV(10, 100.0, 5.0)
	manager := &RMAStateManager{
		cacheKey: "rma_close_3",
		period:   3,
		storage:  NewSeriesStorage(10),
		computed: 0,
	}
	sourceID := &ast.Identifier{Name: "close"}

	var prevValue float64
	for bar := 2; bar < 10; bar++ {
		value, _ := manager.ComputeAtBar(ctx, sourceID, bar)
		if bar > 2 && value <= prevValue {
			t.Errorf("RMA should increase for strongly increasing data: bar %d = %f, bar %d = %f",
				bar-1, prevValue, bar, value)
		}
		prevValue = value
	}
}

func TestEMAStateManager_BackwardThenForwardAccess(t *testing.T) {
	ctx := buildIncrementingOHLCV(8, 100.0, 3.0)
	manager := &EMAStateManager{
		cacheKey:   "ema_close_3",
		period:     3,
		storage:    NewSeriesStorage(8),
		multiplier: 2.0 / 4.0,
		computed:   0,
	}
	sourceID := &ast.Identifier{Name: "close"}

	value7, _ := manager.ComputeAtBar(ctx, sourceID, 7)
	value5, _ := manager.ComputeAtBar(ctx, sourceID, 5)
	value6, _ := manager.ComputeAtBar(ctx, sourceID, 6)
	value4, _ := manager.ComputeAtBar(ctx, sourceID, 4)
	value7Again, _ := manager.ComputeAtBar(ctx, sourceID, 7)

	if value7 != value7Again {
		t.Errorf("Accessing bar 7 again after backward access changed value: first=%f, second=%f",
			value7, value7Again)
	}

	if value4 >= value5 || value5 >= value6 || value6 >= value7 {
		t.Errorf("EMA ordering broken: bar4=%f, bar5=%f, bar6=%f, bar7=%f",
			value4, value5, value6, value7)
	}
}

func TestRSIStateManager_HistoricalAccessIntegrity(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 102},
			{Close: 101},
			{Close: 103},
			{Close: 102},
			{Close: 105},
			{Close: 104},
			{Close: 106},
		},
	}

	manager := &RSIStateManager{
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
	sourceID := &ast.Identifier{Name: "close"}

	value7, _ := manager.ComputeAtBar(ctx, sourceID, 7)
	value5, _ := manager.ComputeAtBar(ctx, sourceID, 5)
	value6, _ := manager.ComputeAtBar(ctx, sourceID, 6)

	if value5 < 0 || value5 > 100 {
		t.Errorf("RSI out of range [0, 100]: bar5 = %f", value5)
	}
	if value6 < 0 || value6 > 100 {
		t.Errorf("RSI out of range [0, 100]: bar6 = %f", value6)
	}
	if value7 < 0 || value7 > 100 {
		t.Errorf("RSI out of range [0, 100]: bar7 = %f", value7)
	}

	if value5 == value6 && value6 == value7 {
		t.Errorf("RSI values should vary for changing data: all = %f", value5)
	}
}

func TestTAStateManager_WarmupPeriodBehavior(t *testing.T) {
	tests := []struct {
		name          string
		createManager func() TAStateManager
		period        int
	}{
		{
			name: "EMA_warmup",
			createManager: func() TAStateManager {
				return &EMAStateManager{
					cacheKey:   "ema_close_5",
					period:     5,
					storage:    NewSeriesStorage(10),
					multiplier: 2.0 / 6.0,
					computed:   0,
				}
			},
			period: 5,
		},
		{
			name: "RMA_warmup",
			createManager: func() TAStateManager {
				return &RMAStateManager{
					cacheKey: "rma_close_5",
					period:   5,
					storage:  NewSeriesStorage(10),
					computed: 0,
				}
			},
			period: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := buildIncrementingOHLCV(10, 100.0, 2.0)
			manager := tt.createManager()
			sourceID := &ast.Identifier{Name: "close"}

			beforeWarmup, _ := manager.ComputeAtBar(ctx, sourceID, tt.period-2)
			atWarmup, _ := manager.ComputeAtBar(ctx, sourceID, tt.period-1)
			afterWarmup, _ := manager.ComputeAtBar(ctx, sourceID, tt.period)

			t.Logf("Warmup transition: bar%d=%f, bar%d=%f, bar%d=%f",
				tt.period-2, beforeWarmup,
				tt.period-1, atWarmup,
				tt.period, afterWarmup)

			if beforeWarmup == atWarmup && atWarmup == afterWarmup {
				t.Errorf("Values should change across warmup boundary")
			}
		})
	}
}

func buildIncrementingOHLCV(size int, start, increment float64) *context.Context {
	data := make([]context.OHLCV, size)
	for i := range data {
		data[i].Close = start + float64(i)*increment
	}
	return &context.Context{Data: data}
}
