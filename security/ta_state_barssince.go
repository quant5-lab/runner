package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/series"
	"github.com/quant5-lab/runner/runtime/value"
)

type BarsSinceStateManager struct {
	buf       *series.Series
	computed  int
	condExpr  ast.Expression
	evaluator *StreamingBarEvaluator
}

func NewBarsSinceStateManager(condExpr ast.Expression, evaluator *StreamingBarEvaluator, capacity int) *BarsSinceStateManager {
	return &BarsSinceStateManager{
		buf:       series.NewSeries(max(capacity, 1)),
		condExpr:  condExpr,
		evaluator: evaluator,
	}
}

func (s *BarsSinceStateManager) ComputeAtBar(secCtx *context.Context, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		if s.computed > 0 {
			s.buf.Next()
		}

		condVal, err := s.evaluator.EvaluateAtBar(s.condExpr, secCtx, s.computed)
		if err != nil {
			return math.NaN(), err
		}

		switch {
		case value.IsTrue(condVal):
			s.buf.Set(0.0)
		case s.computed > 0 && !math.IsNaN(s.buf.Get(1)):
			s.buf.Set(s.buf.Get(1) + 1.0)
		default:
			s.buf.Set(math.NaN())
		}

		s.computed++
	}

	return s.buf.Get(s.buf.Position() - barIdx), nil
}
