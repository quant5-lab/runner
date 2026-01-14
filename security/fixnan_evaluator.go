package security

import (
	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

type FixnanEvaluator struct {
	stateStorage StateStorage
	warmup       WarmupStrategy
	identifier   ExpressionIdentifier
}

func NewFixnanEvaluator(storage StateStorage, warmup WarmupStrategy, identifier ExpressionIdentifier) *FixnanEvaluator {
	return &FixnanEvaluator{
		stateStorage: storage,
		warmup:       warmup,
		identifier:   identifier,
	}
}

func (e *FixnanEvaluator) EvaluateAtBar(evaluator BarEvaluator, call *ast.CallExpression, ctx *context.Context, barIdx int) (float64, error) {
	if len(call.Arguments) < 1 {
		return 0.0, newInsufficientArgumentsError("fixnan", 1, len(call.Arguments))
	}

	cacheKey := "fixnan_" + e.identifier.Identify(call.Arguments[0])

	var state *FixnanState
	if cached, exists := e.stateStorage.Get(cacheKey); exists {
		state = cached.(*FixnanState)
	} else {
		state = NewFixnanState()
		e.stateStorage.Set(cacheKey, state)
		if err := e.warmup.Warmup(evaluator, call.Arguments[0], ctx, barIdx, state); err != nil {
			return 0.0, err
		}
	}

	value, err := evaluator.EvaluateAtBar(call.Arguments[0], ctx, barIdx)
	if err != nil {
		return 0.0, err
	}

	return state.ForwardFill(value), nil
}
