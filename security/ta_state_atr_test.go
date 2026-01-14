package security

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func TestATRStateManager_WarmupPeriod(t *testing.T) {
	ctx := context.New("TEST", "1D", 20)

	for i := 0; i < 20; i++ {
		ctx.AddBar(context.OHLCV{
			Open:   100.0 + float64(i),
			High:   110.0 + float64(i),
			Low:    95.0 + float64(i),
			Close:  105.0 + float64(i),
			Volume: 1000,
		})
	}

	manager := NewATRStateManager("atr_test", 14)
	dummyID := &ast.Identifier{Name: "close"}

	for i := 0; i < 13; i++ {
		result, err := manager.ComputeAtBar(ctx, dummyID, i)
		if err != nil {
			t.Fatalf("ComputeAtBar failed at bar %d: %v", i, err)
		}
		if result != 0.0 {
			t.Errorf("Bar %d: expected 0 during warmup, got %.4f", i, result)
		}
	}

	for i := 13; i < 16; i++ {
		result, err := manager.ComputeAtBar(ctx, dummyID, i)
		if err != nil {
			t.Fatalf("ComputeAtBar failed at bar %d: %v", i, err)
		}
		if result <= 0.0 {
			t.Errorf("Bar %d: expected positive ATR, got %.4f", i, result)
		}
	}
}

func TestATRStateManager_ConsecutiveCalls(t *testing.T) {
	ctx := context.New("TEST", "1D", 20)

	for i := 0; i < 20; i++ {
		ctx.AddBar(context.OHLCV{
			Open:   100.0 + float64(i),
			High:   110.0 + float64(i),
			Low:    95.0 + float64(i),
			Close:  105.0 + float64(i),
			Volume: 1000,
		})
	}

	manager := NewATRStateManager("atr_test", 14)
	dummyID := &ast.Identifier{Name: "close"}

	result1, err := manager.ComputeAtBar(ctx, dummyID, 15)
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}

	result2, err := manager.ComputeAtBar(ctx, dummyID, 15)
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}

	if result1 != result2 {
		t.Errorf("Consecutive calls returned different values: %.4f vs %.4f", result1, result2)
	}
}

func TestATRStateManager_IncreasingVolatility(t *testing.T) {
	ctx := context.New("TEST", "1D", 30)

	for i := 0; i < 15; i++ {
		ctx.AddBar(context.OHLCV{
			Open:   100.0,
			High:   101.0,
			Low:    99.0,
			Close:  100.0,
			Volume: 1000,
		})
	}

	for i := 15; i < 30; i++ {
		ctx.AddBar(context.OHLCV{
			Open:   100.0,
			High:   110.0,
			Low:    90.0,
			Close:  100.0,
			Volume: 1000,
		})
	}

	manager := NewATRStateManager("atr_test", 14)
	dummyID := &ast.Identifier{Name: "close"}

	lowVolATR, _ := manager.ComputeAtBar(ctx, dummyID, 14)
	highVolATR, _ := manager.ComputeAtBar(ctx, dummyID, 28)

	if highVolATR <= lowVolATR {
		t.Errorf("ATR should increase with volatility: low=%.4f, high=%.4f", lowVolATR, highVolATR)
	}
}
