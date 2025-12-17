package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func TestStreamingBarEvaluator_ATR(t *testing.T) {
	ctx := context.New("TEST", "1D", 20)

	bars := []context.OHLCV{
		{Open: 100, High: 105, Low: 95, Close: 102, Volume: 1000},
		{Open: 102, High: 108, Low: 101, Close: 106, Volume: 1100},
		{Open: 106, High: 110, Low: 104, Close: 107, Volume: 1200},
		{Open: 107, High: 112, Low: 106, Close: 111, Volume: 1300},
		{Open: 111, High: 115, Low: 109, Close: 113, Volume: 1400},
		{Open: 113, High: 118, Low: 112, Close: 116, Volume: 1500},
		{Open: 116, High: 120, Low: 114, Close: 118, Volume: 1600},
		{Open: 118, High: 122, Low: 116, Close: 120, Volume: 1700},
		{Open: 120, High: 125, Low: 119, Close: 123, Volume: 1800},
		{Open: 123, High: 128, Low: 122, Close: 126, Volume: 1900},
		{Open: 126, High: 130, Low: 124, Close: 128, Volume: 2000},
		{Open: 128, High: 132, Low: 126, Close: 130, Volume: 2100},
		{Open: 130, High: 135, Low: 129, Close: 133, Volume: 2200},
		{Open: 133, High: 138, Low: 132, Close: 136, Volume: 2300},
		{Open: 136, High: 140, Low: 134, Close: 138, Volume: 2400},
	}

	for _, bar := range bars {
		ctx.AddBar(bar)
	}

	atrCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "atr"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: float64(14)},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	result12, err := evaluator.EvaluateAtBar(atrCall, ctx, 12)
	if err != nil {
		t.Fatalf("EvaluateAtBar at bar 12 failed: %v", err)
	}
	if result12 != 0.0 {
		t.Errorf("bar 12 warmup: expected 0, got %.2f", result12)
	}

	result13, err := evaluator.EvaluateAtBar(atrCall, ctx, 13)
	if err != nil {
		t.Fatalf("EvaluateAtBar at bar 13 failed: %v", err)
	}
	if result13 == 0.0 {
		t.Error("bar 13: expected valid ATR, got 0")
	}
	if result13 <= 0 {
		t.Errorf("bar 13: expected positive ATR, got %.2f", result13)
	}

	result14, err := evaluator.EvaluateAtBar(atrCall, ctx, 14)
	if err != nil {
		t.Fatalf("EvaluateAtBar at bar 14 failed: %v", err)
	}
	if math.IsNaN(result14) {
		t.Error("bar 14: expected valid ATR, got NaN")
	}
	if result14 <= 0 {
		t.Errorf("bar 14: expected positive ATR, got %.2f", result14)
	}
}

func TestStreamingBarEvaluator_ATRCaching(t *testing.T) {
	ctx := context.New("TEST", "1D", 20)

	bars := []context.OHLCV{
		{Open: 100, High: 105, Low: 95, Close: 102, Volume: 1000},
		{Open: 102, High: 108, Low: 101, Close: 106, Volume: 1100},
		{Open: 106, High: 110, Low: 104, Close: 107, Volume: 1200},
		{Open: 107, High: 112, Low: 106, Close: 111, Volume: 1300},
		{Open: 111, High: 115, Low: 109, Close: 113, Volume: 1400},
		{Open: 113, High: 118, Low: 112, Close: 116, Volume: 1500},
		{Open: 116, High: 120, Low: 114, Close: 118, Volume: 1600},
		{Open: 118, High: 122, Low: 116, Close: 120, Volume: 1700},
		{Open: 120, High: 125, Low: 119, Close: 123, Volume: 1800},
		{Open: 123, High: 128, Low: 122, Close: 126, Volume: 1900},
		{Open: 126, High: 130, Low: 124, Close: 128, Volume: 2000},
		{Open: 128, High: 132, Low: 126, Close: 130, Volume: 2100},
		{Open: 130, High: 135, Low: 129, Close: 133, Volume: 2200},
		{Open: 133, High: 138, Low: 132, Close: 136, Volume: 2300},
		{Open: 136, High: 140, Low: 134, Close: 138, Volume: 2400},
	}

	for _, bar := range bars {
		ctx.AddBar(bar)
	}

	atrCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "atr"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: float64(14)},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	result1, err := evaluator.EvaluateAtBar(atrCall, ctx, 14)
	if err != nil {
		t.Fatalf("First evaluation failed: %v", err)
	}

	result2, err := evaluator.EvaluateAtBar(atrCall, ctx, 14)
	if err != nil {
		t.Fatalf("Second evaluation failed: %v", err)
	}

	if result1 != result2 {
		t.Errorf("ATR values differ between calls: %.4f vs %.4f", result1, result2)
	}
}
