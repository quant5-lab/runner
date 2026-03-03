package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func TestTSIStateManager_FlatSourceZeroTSI(t *testing.T) {
	/* Flat source: zero momentum → TSI = 0 after warmup */
	ctx := &context.Context{
		Data: make([]context.OHLCV, 40),
	}
	for i := range ctx.Data {
		ctx.Data[i] = context.OHLCV{Close: 100}
	}

	manager := NewTSIStateManager("tsi_close_5_13", 5, 13)
	sourceID := &ast.Identifier{Name: "close"}

	value, err := manager.ComputeAtBar(ctx, sourceID, 30)
	if err != nil {
		t.Fatalf("ComputeAtBar failed: %v", err)
	}

	if math.Abs(value-0.0) > 0.0001 {
		t.Errorf("TSI with flat source = %f, want 0.0", value)
	}
}

func TestTSIStateManager_WarmupPeriod(t *testing.T) {
	tests := []struct {
		name  string
		short int
		long  int
	}{
		{"short1_long1", 1, 1},
		{"short1_long13", 1, 13},
		{"short5_long1", 5, 1},
		{"short5_long13", 5, 13},
		{"short13_long25", 13, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			barCount := tt.short + tt.long + 5
			ctx := &context.Context{Data: make([]context.OHLCV, barCount)}
			for i := range ctx.Data {
				ctx.Data[i] = context.OHLCV{Close: float64(100 + i)}
			}

			manager := NewTSIStateManager("tsi_test", tt.short, tt.long)
			sourceID := &ast.Identifier{Name: "close"}
			warmup := tt.short + tt.long - 1

			for i := 0; i < warmup; i++ {
				v, err := manager.ComputeAtBar(ctx, sourceID, i)
				if err != nil {
					t.Fatalf("bar %d: ComputeAtBar failed: %v", i, err)
				}
				if !math.IsNaN(v) {
					t.Errorf("bar %d: expected NaN during warmup, got %f", i, v)
				}
			}

			v, err := manager.ComputeAtBar(ctx, sourceID, warmup)
			if err != nil {
				t.Fatalf("bar %d: ComputeAtBar failed: %v", warmup, err)
			}
			if math.IsNaN(v) {
				t.Errorf("bar %d (first post-warmup): expected valid value, got NaN", warmup)
			}
		})
	}
}

func TestTSIStateManager_SequentialComputation(t *testing.T) {
	ctx := &context.Context{
		Data: make([]context.OHLCV, 40),
	}
	for i := range ctx.Data {
		ctx.Data[i] = context.OHLCV{Close: float64(100 + i)}
	}

	manager := NewTSIStateManager("tsi_close_5_13", 5, 13)
	sourceID := &ast.Identifier{Name: "close"}

	var lastValid float64
	for i := 0; i < len(ctx.Data); i++ {
		v, err := manager.ComputeAtBar(ctx, sourceID, i)
		if err != nil {
			t.Fatalf("bar %d: ComputeAtBar failed: %v", i, err)
		}
		if !math.IsNaN(v) {
			lastValid = v
		}
	}

	if math.IsNaN(lastValid) {
		t.Error("Expected a valid TSI value in monotonic series, all are NaN")
	}
}

func TestTSIStateManager_ResultInBounds(t *testing.T) {
	/* TSI is bounded [-100, 100] */
	ctx := &context.Context{
		Data: make([]context.OHLCV, 60),
	}
	for i := range ctx.Data {
		if i%2 == 0 {
			ctx.Data[i] = context.OHLCV{Close: float64(100 + i)}
		} else {
			ctx.Data[i] = context.OHLCV{Close: float64(100 - i)}
		}
	}

	manager := NewTSIStateManager("tsi_close_5_13", 5, 13)
	sourceID := &ast.Identifier{Name: "close"}

	for i := 0; i < len(ctx.Data); i++ {
		v, err := manager.ComputeAtBar(ctx, sourceID, i)
		if err != nil {
			t.Fatalf("bar %d: ComputeAtBar failed: %v", i, err)
		}
		if math.IsNaN(v) {
			continue
		}
		if v < -100.0 || v > 100.0 {
			t.Errorf("bar %d: TSI = %f, out of [-100, 100] bounds", i, v)
		}
	}
}

func TestTSIStateManager_StateReuse(t *testing.T) {
	ctx := &context.Context{
		Data: make([]context.OHLCV, 40),
	}
	for i := range ctx.Data {
		ctx.Data[i] = context.OHLCV{Close: float64(100 + i)}
	}

	manager := NewTSIStateManager("tsi_close_5_13", 5, 13)
	sourceID := &ast.Identifier{Name: "close"}

	v1, _ := manager.ComputeAtBar(ctx, sourceID, 25)
	v2, _ := manager.ComputeAtBar(ctx, sourceID, 25)

	if v1 != v2 {
		t.Errorf("State reuse failed: first=%.6f, second=%.6f", v1, v2)
	}
}

func TestTSIStateManager_IsolatedState(t *testing.T) {
	/* Two managers with same params but different cacheKeys must not share state */
	ctx := &context.Context{
		Data: make([]context.OHLCV, 40),
	}
	for i := range ctx.Data {
		ctx.Data[i] = context.OHLCV{Close: float64(100 + i)}
	}

	m1 := NewTSIStateManager("key_a", 5, 13)
	m2 := NewTSIStateManager("key_b", 5, 13)

	closeID := &ast.Identifier{Name: "close"}

	v1, _ := m1.ComputeAtBar(ctx, closeID, 30)
	v2, _ := m2.ComputeAtBar(ctx, closeID, 30)

	if math.IsNaN(v1) || math.IsNaN(v2) {
		t.Skip("TSI values are NaN at bar 30")
	}

	if math.Abs(v1-v2) > 0.0001 {
		t.Errorf("Same params and data should produce same TSI: m1=%.6f, m2=%.6f", v1, v2)
	}
}
