package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

type CUMStateManager struct {
	buf       forwardBufferE
	evaluator BarEvaluator
}

func NewCUMStateManager(capacity int, evaluator BarEvaluator) *CUMStateManager {
	return &CUMStateManager{
		buf:       newForwardBufferE(capacity),
		evaluator: evaluator,
	}
}

func (s *CUMStateManager) ComputeAtBar(secCtx *context.Context, sourceExpr ast.Expression, barIdx int) (float64, error) {
	if s.buf.growsFor(len(secCtx.Data)) {
		s.buf.reallocate(len(secCtx.Data))
	}
	if err := s.buf.advanceTo(barIdx, func(bar int) (float64, error) {
		v, err := s.evaluator.EvaluateAtBar(sourceExpr, secCtx, bar)
		if err != nil {
			return math.NaN(), err
		}
		prev := 0.0
		if bar > 0 {
			prev = s.buf.prev()
		}
		return prev + v, nil
	}); err != nil {
		return math.NaN(), err
	}
	return s.buf.at(barIdx), nil
}
