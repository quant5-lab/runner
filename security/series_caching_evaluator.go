package security

import (
	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

type SeriesCachingEvaluator struct {
	delegate BarEvaluator
}

func NewSeriesCachingEvaluator(delegate BarEvaluator) *SeriesCachingEvaluator {
	return &SeriesCachingEvaluator{
		delegate: delegate,
	}
}

func (e *SeriesCachingEvaluator) EvaluateAtBar(expr ast.Expression, secCtx *context.Context, barIdx int) (float64, error) {
	return e.delegate.EvaluateAtBar(expr, secCtx, barIdx)
}

func (e *SeriesCachingEvaluator) Unwrap() BarEvaluator {
	return e.delegate
}
