package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

// FixnanStateManager ensures forward-fill semantics are order-independent: any bar
// index query returns the value a strictly sequential 0…N scan would produce at that bar.
type FixnanStateManager struct {
	innerExpr ast.Expression
	buf       forwardBufferE
	evaluator BarEvaluator
}

func newFixnanStateManager(innerExpr ast.Expression, capacity int, evaluator BarEvaluator) *FixnanStateManager {
	return &FixnanStateManager{
		innerExpr: innerExpr,
		buf:       newForwardBufferE(capacity),
		evaluator: evaluator,
	}
}

func (s *FixnanStateManager) ComputeAtBar(secCtx *context.Context, barIdx int) (float64, error) {
	if s.buf.growsFor(len(secCtx.Data)) {
		s.buf.reallocate(len(secCtx.Data))
	}
	if err := s.buf.advanceTo(barIdx, func(bar int) (float64, error) {
		val, err := s.evaluator.EvaluateAtBar(s.innerExpr, secCtx, bar)
		if err != nil || math.IsNaN(val) {
			return s.buf.prev(), nil
		}
		return val, nil
	}); err != nil {
		return math.NaN(), err
	}
	return s.buf.at(barIdx), nil
}

func evaluateFixnanAtBar(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	if len(call.Arguments) < 1 {
		return 0.0, newInsufficientArgumentsError("fixnan", 1, len(call.Arguments))
	}
	innerExpr := call.Arguments[0]
	cacheKey := "fixnan:" + expressionKey(innerExpr)
	state, exists := e.fixnanCache[cacheKey]
	if !exists {
		state = newFixnanStateManager(innerExpr, len(secCtx.Data), e)
		e.fixnanCache[cacheKey] = state
	}
	return state.ComputeAtBar(secCtx, barIdx)
}

func init() {
	registerCallHandlerAliases(evaluateFixnanAtBar, "fixnan", "ta.fixnan")
}
