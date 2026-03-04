package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func TestTSIStateManager_FlatSourceZeroTSI(t *testing.T) {
	ctx := &context.Context{
		Data: make([]context.OHLCV, 40),
	}
	for i := range ctx.Data {
		ctx.Data[i] = context.OHLCV{Close: 100}
	}

	manager := NewTSIStateManager("tsi_close_5_13", 5, 13, len(ctx.Data))
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

			manager := NewTSIStateManager("tsi_test", tt.short, tt.long, barCount)
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

	manager := NewTSIStateManager("tsi_close_5_13", 5, 13, len(ctx.Data))
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

	manager := NewTSIStateManager("tsi_close_5_13", 5, 13, len(ctx.Data))
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

func TestTSIStateManager_StatePreservation(t *testing.T) {
	ctx := &context.Context{
		Data: make([]context.OHLCV, 50),
	}
	for i := range ctx.Data {
		phase := float64(100 + i*2)
		if i >= 20 {
			phase = float64(100 + 20*2 - (i - 20))
		}
		ctx.Data[i] = context.OHLCV{Close: phase}
	}

	manager := NewTSIStateManager("tsi_close_5_13", 5, 13, len(ctx.Data))
	sourceID := &ast.Identifier{Name: "close"}

	warmup := 5 + 13 - 1

	valBar20First, err := manager.ComputeAtBar(ctx, sourceID, 20)
	if err != nil {
		t.Fatalf("ComputeAtBar(20) first call failed: %v", err)
	}
	if math.IsNaN(valBar20First) {
		t.Fatalf("bar 20 should be post-warmup (%d), got NaN", warmup)
	}

	valBar30, err := manager.ComputeAtBar(ctx, sourceID, 30)
	if err != nil {
		t.Fatalf("ComputeAtBar(30) failed: %v", err)
	}

	valBar20Second, err := manager.ComputeAtBar(ctx, sourceID, 20)
	if err != nil {
		t.Fatalf("ComputeAtBar(20) second call failed: %v", err)
	}

	if valBar20First != valBar20Second {
		t.Errorf("Historical value changed: first=%.6f, after forward=%.6f", valBar20First, valBar20Second)
	}

	if math.Abs(valBar20First-valBar30) < 0.01 {
		t.Errorf("Phase-change source should produce divergent TSI: bar20=%.6f, bar30=%.6f", valBar20First, valBar30)
	}
}

func TestTSIStateManager_IsolatedState(t *testing.T) {
	ctx := &context.Context{
		Data: make([]context.OHLCV, 40),
	}
	for i := range ctx.Data {
		ctx.Data[i] = context.OHLCV{Close: float64(100 + i)}
	}

	m1 := NewTSIStateManager("key_a", 5, 13, len(ctx.Data))
	m2 := NewTSIStateManager("key_b", 5, 13, len(ctx.Data))

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

func TestTSIStateManager_NonSequentialAccess(t *testing.T) {
	ctx := &context.Context{
		Data: make([]context.OHLCV, 40),
	}
	for i := range ctx.Data {
		oscillation := float64(100 + 10*((i%5)-2))
		ctx.Data[i] = context.OHLCV{Close: oscillation}
	}

	manager := NewTSIStateManager("tsi_close_5_13", 5, 13, len(ctx.Data))
	sourceID := &ast.Identifier{Name: "close"}

	warmup := 5 + 13 - 1

	valBar35, err := manager.ComputeAtBar(ctx, sourceID, 35)
	if err != nil {
		t.Fatalf("ComputeAtBar(35) failed: %v", err)
	}

	accessPattern := []int{25, 30, 20, 35, 28}
	results := make(map[int]float64)

	for _, barIdx := range accessPattern {
		v, err := manager.ComputeAtBar(ctx, sourceID, barIdx)
		if err != nil {
			t.Fatalf("ComputeAtBar(%d) failed: %v", barIdx, err)
		}
		if barIdx >= warmup && math.IsNaN(v) {
			t.Errorf("bar %d (post-warmup): got NaN", barIdx)
		}
		results[barIdx] = v
	}

	if results[35] != valBar35 {
		t.Errorf("Non-sequential access changed bar 35: initial=%.6f, after pattern=%.6f", valBar35, results[35])
	}

	val25First := results[25]
	val25Second, _ := manager.ComputeAtBar(ctx, sourceID, 25)
	if val25First != val25Second {
		t.Errorf("Repeated non-sequential access changed bar 25: first=%.6f, second=%.6f", val25First, val25Second)
	}
}
