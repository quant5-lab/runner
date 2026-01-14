package request

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/series"
)

type ExpressionSeriesBuilder struct {
	evaluator BarEvaluator
}

func NewExpressionSeriesBuilder(evaluator BarEvaluator) *ExpressionSeriesBuilder {
	return &ExpressionSeriesBuilder{
		evaluator: evaluator,
	}
}

func (b *ExpressionSeriesBuilder) BuildSeries(expr ast.Expression, secCtx *context.Context) (*series.Series, error) {
	if len(secCtx.Data) == 0 {
		return nil, fmt.Errorf("cannot build series from empty context")
	}

	seriesBuffer := series.NewSeries(len(secCtx.Data))

	for barIdx := 0; barIdx < len(secCtx.Data); barIdx++ {
		value, err := b.evaluator.EvaluateAtBar(expr, secCtx, barIdx)
		if err != nil {
			return nil, err
		}

		seriesBuffer.Set(value)

		if barIdx < len(secCtx.Data)-1 {
			seriesBuffer.Next()
		}
	}

	return seriesBuffer, nil
}
