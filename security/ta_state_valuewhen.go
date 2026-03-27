package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/series"
)

type ValuewhenStateManager struct {
	cacheKey      string
	occurrence    int
	conditionExpr ast.Expression
	sourceExpr    ast.Expression
	matchRing     *MatchIndexRing
	computed      int
	evaluator     BarEvaluator
	resultBuf     *series.Series
}

func NewValuewhenStateManager(
	cacheKey string,
	occurrence int,
	conditionExpr ast.Expression,
	sourceExpr ast.Expression,
	capacity int,
	evaluator BarEvaluator,
) *ValuewhenStateManager {
	return &ValuewhenStateManager{
		cacheKey:      cacheKey,
		occurrence:    occurrence,
		conditionExpr: conditionExpr,
		sourceExpr:    sourceExpr,
		matchRing:     NewMatchIndexRing(occurrence),
		evaluator:     evaluator,
		resultBuf:     series.NewSeries(max(capacity, 1)),
	}
}

func (s *ValuewhenStateManager) ComputeAtBar(secCtx *context.Context, _ ast.Expression, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		if s.computed > 0 {
			s.resultBuf.Next()
		}

		condVal, err := s.evaluator.EvaluateAtBar(s.conditionExpr, secCtx, s.computed)
		if err != nil {
			return math.NaN(), err
		}

		if condVal != 0.0 {
			s.matchRing.Push(s.computed)
		}

		targetIdx, found := s.matchRing.GetNthMostRecent(s.occurrence)
		if !found {
			s.resultBuf.Set(math.NaN())
		} else {
			sourceVal, err := s.evaluator.EvaluateAtBar(s.sourceExpr, secCtx, targetIdx)
			if err != nil {
				return math.NaN(), err
			}
			s.resultBuf.Set(sourceVal)
		}

		s.computed++
	}

	return s.resultBuf.Get(s.resultBuf.Position() - barIdx), nil
}
