package security

import (
	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

type StatefulForwardFill interface {
	ForwardFill(value float64) float64
}

type WarmupStrategy interface {
	Warmup(evaluator BarEvaluator, expr ast.Expression, ctx *context.Context, targetBar int, state StatefulForwardFill) error
}

type SequentialWarmupStrategy struct{}

func NewSequentialWarmupStrategy() *SequentialWarmupStrategy {
	return &SequentialWarmupStrategy{}
}

func (s *SequentialWarmupStrategy) Warmup(evaluator BarEvaluator, expr ast.Expression, ctx *context.Context, targetBar int, state StatefulForwardFill) error {
	for barIdx := 0; barIdx < targetBar; barIdx++ {
		value, err := evaluator.EvaluateAtBar(expr, ctx, barIdx)
		if err != nil {
			continue
		}
		state.ForwardFill(value)
	}
	return nil
}
