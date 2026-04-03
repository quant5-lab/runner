package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/value"
)

type BarsSinceStateManager struct {
	buf       forwardBufferE
	condExpr  ast.Expression
	evaluator BarEvaluator
}

func NewBarsSinceStateManager(condExpr ast.Expression, evaluator BarEvaluator, capacity int) *BarsSinceStateManager {
	return &BarsSinceStateManager{
		buf:       newForwardBufferE(capacity),
		condExpr:  condExpr,
		evaluator: evaluator,
	}
}

func (s *BarsSinceStateManager) ComputeAtBar(secCtx *context.Context, barIdx int) (float64, error) {
	if s.buf.growsFor(len(secCtx.Data)) {
		s.buf.reallocate(len(secCtx.Data))
	}
	if err := s.buf.advanceTo(barIdx, func(bar int) (float64, error) {
		condVal, err := s.evaluator.EvaluateAtBar(s.condExpr, secCtx, bar)
		if err != nil {
			return math.NaN(), err
		}
		switch {
		case value.IsTrue(condVal):
			return 0.0, nil
		case bar > 0 && !math.IsNaN(s.buf.prev()):
			return s.buf.prev() + 1.0, nil
		default:
			return math.NaN(), nil
		}
	}); err != nil {
		return math.NaN(), err
	}
	return s.buf.at(barIdx), nil
}
