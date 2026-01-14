package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func TestSMAStateManager_CircularBufferBehavior(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 10},
			{Close: 20},
			{Close: 30},
			{Close: 40},
			{Close: 50},
		},
	}

	manager := &SMAStateManager{
		cacheKey: "sma_close_3",
		period:   3,
		buffer:   make([]float64, 3),
		computed: 0,
	}

	sourceID := &ast.Identifier{Name: "close"}

	tests := []struct {
		barIdx   int
		expected float64
	}{
		{0, 0.0},
		{1, 0.0},
		{2, 20.0},
		{3, 30.0},
		{4, 40.0},
	}

	for _, tt := range tests {
		value, err := manager.ComputeAtBar(ctx, sourceID, tt.barIdx)
		if err != nil {
			t.Fatalf("bar %d: ComputeAtBar failed: %v", tt.barIdx, err)
		}

		if math.Abs(value-tt.expected) > 0.0001 {
			t.Errorf("bar %d: expected %.4f, got %.4f", tt.barIdx, tt.expected, value)
		}
	}
}

func TestSMAStateManager_IncrementalComputation(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 110},
			{Close: 120},
			{Close: 130},
		},
	}

	manager := &SMAStateManager{
		cacheKey: "sma_close_2",
		period:   2,
		buffer:   make([]float64, 2),
		computed: 0,
	}

	sourceID := &ast.Identifier{Name: "close"}

	value1, _ := manager.ComputeAtBar(ctx, sourceID, 1)
	if math.Abs(value1-105.0) > 0.0001 {
		t.Errorf("bar 1: expected 105.0, got %.4f", value1)
	}

	value2, _ := manager.ComputeAtBar(ctx, sourceID, 2)
	if math.Abs(value2-115.0) > 0.0001 {
		t.Errorf("bar 2: expected 115.0, got %.4f", value2)
	}

	if manager.computed != 3 {
		t.Errorf("expected computed=3, got %d", manager.computed)
	}
}

func TestEMAStateManager_ExponentialSmoothing(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 110},
			{Close: 120},
			{Close: 130},
			{Close: 140},
		},
	}

	multiplier := 2.0 / float64(3+1)
	manager := &EMAStateManager{
		cacheKey:   "ema_close_3",
		period:     3,
		multiplier: multiplier,
		computed:   0,
	}

	sourceID := &ast.Identifier{Name: "close"}

	value2, _ := manager.ComputeAtBar(ctx, sourceID, 2)
	if value2 == 0.0 {
		t.Error("EMA at warmup boundary should not be zero")
	}

	value4, _ := manager.ComputeAtBar(ctx, sourceID, 4)
	if value4 < 120.0 || value4 > 135.0 {
		t.Errorf("EMA bar 4: expected [120, 135], got %.4f", value4)
	}

	if value4 <= value2 {
		t.Errorf("EMA should increase: bar2=%.4f, bar4=%.4f", value2, value4)
	}
}

func TestEMAStateManager_StatePreservation(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 102},
			{Close: 104},
			{Close: 106},
		},
	}

	manager := &EMAStateManager{
		cacheKey:   "ema_close_3",
		period:     3,
		multiplier: 2.0 / 4.0,
		computed:   0,
	}

	sourceID := &ast.Identifier{Name: "close"}

	value2First, _ := manager.ComputeAtBar(ctx, sourceID, 2)
	value2Second, _ := manager.ComputeAtBar(ctx, sourceID, 2)

	if value2First != value2Second {
		t.Errorf("state not preserved: first=%.4f, second=%.4f", value2First, value2Second)
	}
}

func TestRMAStateManager_AlphaSmoothing(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 120},
			{Close: 110},
			{Close: 130},
			{Close: 115},
		},
	}

	manager := &RMAStateManager{
		cacheKey: "rma_close_3",
		period:   3,
		computed: 0,
	}

	sourceID := &ast.Identifier{Name: "close"}

	value4, err := manager.ComputeAtBar(ctx, sourceID, 4)
	if err != nil {
		t.Fatalf("ComputeAtBar failed: %v", err)
	}

	if value4 < 110.0 || value4 > 125.0 {
		t.Errorf("RMA bar 4: expected smoothed [110, 125], got %.4f", value4)
	}
}

func TestRSIStateManager_DualRMAIntegration(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 102},
			{Close: 101},
			{Close: 103},
			{Close: 102},
			{Close: 104},
			{Close: 103},
		},
	}

	manager := &RSIStateManager{
		cacheKey: "rsi_close_3",
		period:   3,
		rmaGain: &RMAStateManager{
			cacheKey: "rsi_close_3_gain",
			period:   3,
			computed: 0,
		},
		rmaLoss: &RMAStateManager{
			cacheKey: "rsi_close_3_loss",
			period:   3,
			computed: 0,
		},
		computed: 0,
	}

	sourceID := &ast.Identifier{Name: "close"}

	value6, err := manager.ComputeAtBar(ctx, sourceID, 6)
	if err != nil {
		t.Fatalf("ComputeAtBar failed: %v", err)
	}

	if value6 < 0.0 || value6 > 100.0 {
		t.Errorf("RSI must be [0, 100], got %.4f", value6)
	}
}

func TestRSIStateManager_AllGainsScenario(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 102},
			{Close: 104},
			{Close: 106},
			{Close: 108},
		},
	}

	manager := &RSIStateManager{
		cacheKey: "rsi_close_3",
		period:   3,
		rmaGain: &RMAStateManager{
			cacheKey: "rsi_close_3_gain",
			period:   3,
			computed: 0,
		},
		rmaLoss: &RMAStateManager{
			cacheKey: "rsi_close_3_loss",
			period:   3,
			computed: 0,
		},
		computed: 0,
	}

	sourceID := &ast.Identifier{Name: "close"}

	value4, err := manager.ComputeAtBar(ctx, sourceID, 4)
	if err != nil {
		t.Fatalf("ComputeAtBar failed: %v", err)
	}

	if value4 < 80.0 || value4 > 100.0 {
		t.Errorf("RSI all gains: expected [80, 100], got %.4f", value4)
	}
}

func TestRSIStateManager_AllLossesScenario(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 108},
			{Close: 106},
			{Close: 104},
			{Close: 102},
			{Close: 100},
		},
	}

	manager := &RSIStateManager{
		cacheKey: "rsi_close_3",
		period:   3,
		rmaGain: &RMAStateManager{
			cacheKey: "rsi_close_3_gain",
			period:   3,
			computed: 0,
		},
		rmaLoss: &RMAStateManager{
			cacheKey: "rsi_close_3_loss",
			period:   3,
			computed: 0,
		},
		computed: 0,
	}

	sourceID := &ast.Identifier{Name: "close"}

	value4, err := manager.ComputeAtBar(ctx, sourceID, 4)
	if err != nil {
		t.Fatalf("ComputeAtBar failed: %v", err)
	}

	if value4 < 0.0 || value4 > 20.0 {
		t.Errorf("RSI all losses: expected [0, 20], got %.4f", value4)
	}
}

func TestNewTAStateManager_FactoryPattern(t *testing.T) {
	tests := []struct {
		cacheKey     string
		period       int
		capacity     int
		expectedType string
	}{
		{"sma_close_20", 20, 100, "SMA"},
		{"ema_high_14", 14, 100, "EMA"},
		{"rma_low_10", 10, 100, "RMA"},
		{"rsi_close_14", 14, 100, "RSI"},
	}

	for _, tt := range tests {
		t.Run(tt.cacheKey, func(t *testing.T) {
			manager := NewTAStateManager(tt.cacheKey, tt.period, tt.capacity)
			if manager == nil {
				t.Fatal("NewTAStateManager returned nil")
			}

			switch tt.expectedType {
			case "SMA":
				if _, ok := manager.(*SMAStateManager); !ok {
					t.Errorf("expected SMAStateManager, got %T", manager)
				}
			case "EMA":
				if _, ok := manager.(*EMAStateManager); !ok {
					t.Errorf("expected EMAStateManager, got %T", manager)
				}
			case "RMA":
				if _, ok := manager.(*RMAStateManager); !ok {
					t.Errorf("expected RMAStateManager, got %T", manager)
				}
			case "RSI":
				if _, ok := manager.(*RSIStateManager); !ok {
					t.Errorf("expected RSIStateManager, got %T", manager)
				}
			}
		})
	}
}

func TestNewTAStateManager_UnknownFunction(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for unknown TA function")
		}
	}()

	NewTAStateManager("unknown_close_14", 14, 100)
}

func TestContainsFunction(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"sma_close_20", "sma", true},
		{"ema_high_14", "ema", true},
		{"rma_low_10", "rma", true},
		{"rsi_close_14", "rsi", true},
		{"sma_close_20", "ema", false},
		{"ta_ema_14", "ema", true},
		{"close", "sma", false},
		{"", "sma", false},
		{"sma", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, expected %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}
