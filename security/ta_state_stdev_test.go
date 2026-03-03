package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func TestSTDEVStateManager_PopulationStandardDeviation(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 10}, // Bar 0
			{Close: 12}, // Bar 1
			{Close: 14}, // Bar 2: stdev([10,12,14]) = sqrt(((10-12)^2 + (12-12)^2 + (14-12)^2)/3) = sqrt(8/3) = 1.6329
			{Close: 16}, // Bar 3: stdev([12,14,16]) = sqrt(8/3) = 1.6329
			{Close: 10}, // Bar 4: stdev([14,16,10]) = sqrt(((14-13.33)^2 + (16-13.33)^2 + (10-13.33)^2)/3) = 2.494
		},
	}

	manager := NewSTDEVStateManager("stdev_close_3", 3, 10)
	sourceID := &ast.Identifier{Name: "close"}

	tests := []struct {
		name     string
		barIdx   int
		expected float64
		isNaN    bool
	}{
		{"warmup bar 0", 0, 0, true},
		{"warmup bar 1", 1, 0, true},
		{"first valid bar", 2, 1.6329, false},
		{"uniform growth", 3, 1.6329, false},
		{"with variance", 4, 2.494, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := manager.ComputeAtBar(ctx, sourceID, tt.barIdx)
			if err != nil {
				t.Fatalf("ComputeAtBar failed: %v", err)
			}

			if tt.isNaN {
				if !math.IsNaN(value) {
					t.Errorf("expected NaN, got %.4f", value)
				}
			} else {
				if math.Abs(value-tt.expected) > 0.001 {
					t.Errorf("expected %.4f, got %.4f", tt.expected, value)
				}
			}
		})
	}
}

func TestSTDEVStateManager_ZeroVariance(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 100},
			{Close: 100},
			{Close: 100},
		},
	}

	manager := NewSTDEVStateManager("stdev_close_3", 3, 10)
	sourceID := &ast.Identifier{Name: "close"}

	value, err := manager.ComputeAtBar(ctx, sourceID, 2)
	if err != nil {
		t.Fatalf("ComputeAtBar failed: %v", err)
	}

	if value != 0.0 {
		t.Errorf("constant values should have stdev=0, got %.6f", value)
	}
}

func TestSTDEVStateManager_RollingWindowCorrectness(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 2}, // Bar 0
			{Close: 4}, // Bar 1
			{Close: 4}, // Bar 2: [2,4,4] mean=3.33, stdev=0.9428
			{Close: 4}, // Bar 3: [4,4,4] mean=4, stdev=0
			{Close: 5}, // Bar 4: [4,4,5] mean=4.33, stdev=0.4714
			{Close: 5}, // Bar 5: [4,5,5] mean=4.67, stdev=0.4714
			{Close: 7}, // Bar 6: [5,5,7] mean=5.67, stdev=0.9428
			{Close: 9}, // Bar 7: [5,7,9] mean=7, stdev=1.6329
		},
	}

	manager := NewSTDEVStateManager("stdev_close_3", 3, 10)
	sourceID := &ast.Identifier{Name: "close"}

	tests := []struct {
		barIdx   int
		expected float64
	}{
		{2, 0.9428},
		{3, 0.0},
		{4, 0.4714},
		{5, 0.4714},
		{6, 0.9428},
		{7, 1.6329},
	}

	for _, tt := range tests {
		value, err := manager.ComputeAtBar(ctx, sourceID, tt.barIdx)
		if err != nil {
			t.Fatalf("bar %d: ComputeAtBar failed: %v", tt.barIdx, err)
		}

		if math.Abs(value-tt.expected) > 0.001 {
			t.Errorf("bar %d: expected %.4f, got %.4f", tt.barIdx, tt.expected, value)
		}
	}
}

func TestSTDEVStateManager_DifferentSources(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Open: 10, High: 15, Low: 9, Close: 12},
			{Open: 11, High: 16, Low: 10, Close: 13},
			{Open: 12, High: 17, Low: 11, Close: 14},
		},
	}

	manager := NewSTDEVStateManager("stdev_high_3", 3, 10)
	sourceID := &ast.Identifier{Name: "high"}

	value, err := manager.ComputeAtBar(ctx, sourceID, 2)
	if err != nil {
		t.Fatalf("ComputeAtBar failed: %v", err)
	}

	// high=[15,16,17], mean=16, stdev=sqrt((1+0+1)/3)=0.8165
	expected := 0.8165
	if math.Abs(value-expected) > 0.001 {
		t.Errorf("expected %.4f, got %.4f", expected, value)
	}
}

func TestSTDEVStateManager_StatePreservation(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 105},
			{Close: 110},
			{Close: 95},
			{Close: 100},
			{Close: 115},
			{Close: 90},
		},
	}

	manager := NewSTDEVStateManager("stdev_close_3", 3, 10)
	sourceID := &ast.Identifier{Name: "close"}

	valBar3First, err := manager.ComputeAtBar(ctx, sourceID, 3)
	if err != nil {
		t.Fatalf("ComputeAtBar(3) first call failed: %v", err)
	}

	valBar5, err := manager.ComputeAtBar(ctx, sourceID, 5)
	if err != nil {
		t.Fatalf("ComputeAtBar(5) failed: %v", err)
	}

	valBar3Second, err := manager.ComputeAtBar(ctx, sourceID, 3)
	if err != nil {
		t.Fatalf("ComputeAtBar(3) second call failed: %v", err)
	}

	if valBar3First != valBar3Second {
		t.Errorf("Historical value changed: first=%.4f, after forward=%.4f", valBar3First, valBar3Second)
	}

	if math.Abs(valBar3First-valBar5) < 0.01 {
		t.Errorf("Different bars should produce different STDEV: bar3=%.4f, bar5=%.4f", valBar3First, valBar5)
	}
}
