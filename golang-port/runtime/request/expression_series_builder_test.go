package request

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

type mockBarEvaluator struct {
	values []float64
}

func (m *mockBarEvaluator) EvaluateAtBar(expr ast.Expression, secCtx *context.Context, barIdx int) (float64, error) {
	if barIdx < 0 || barIdx >= len(m.values) {
		return 0.0, nil
	}
	return m.values[barIdx], nil
}

func TestExpressionSeriesBuilder_BuildSeries(t *testing.T) {
	evaluator := &mockBarEvaluator{
		values: []float64{10.0, 20.0, 30.0, 40.0, 50.0},
	}

	builder := NewExpressionSeriesBuilder(evaluator)

	secCtx := &context.Context{
		Data: make([]context.OHLCV, 5),
	}

	expr := &ast.Identifier{Name: "close"}

	seriesBuffer, err := builder.BuildSeries(expr, secCtx)
	if err != nil {
		t.Fatalf("BuildSeries failed: %v", err)
	}

	if seriesBuffer == nil {
		t.Fatal("Expected series buffer, got nil")
	}

	if seriesBuffer.GetCurrent() != 50.0 {
		t.Errorf("Expected current value 50.0, got %f", seriesBuffer.GetCurrent())
	}

	if seriesBuffer.Get(1) != 40.0 {
		t.Errorf("Expected Get(1) = 40.0, got %f", seriesBuffer.Get(1))
	}

	if seriesBuffer.Get(4) != 10.0 {
		t.Errorf("Expected Get(4) = 10.0, got %f", seriesBuffer.Get(4))
	}
}

func TestExpressionSeriesBuilder_EmptyContext(t *testing.T) {
	evaluator := &mockBarEvaluator{
		values: []float64{},
	}

	builder := NewExpressionSeriesBuilder(evaluator)

	secCtx := &context.Context{
		Data: make([]context.OHLCV, 0),
	}

	expr := &ast.Identifier{Name: "close"}

	seriesBuffer, err := builder.BuildSeries(expr, secCtx)
	if err == nil {
		t.Error("Expected error for empty context, got nil")
	}

	if seriesBuffer != nil {
		t.Error("Expected nil series for empty context")
	}
}
