package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func (e *StreamingBarEvaluator) evaluateChangeAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, length, err := extractChangeArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length {
		return math.NaN(), nil
	}
	current, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	previous, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx-length)
	if err != nil {
		return math.NaN(), err
	}
	return current - previous, nil
}

func (e *StreamingBarEvaluator) evaluateMomAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, length, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length {
		return math.NaN(), nil
	}
	current, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	previous, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx-length)
	if err != nil {
		return math.NaN(), err
	}
	return current - previous, nil
}

func (e *StreamingBarEvaluator) evaluateRocAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, length, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length {
		return math.NaN(), nil
	}
	current, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	previous, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx-length)
	if err != nil {
		return math.NaN(), err
	}
	if previous == 0.0 {
		return math.NaN(), nil
	}
	return 100.0 * (current - previous) / math.Abs(previous), nil
}

func (e *StreamingBarEvaluator) evaluateCrossoverAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	arg1, arg2, err := extractTwoExpressionArguments(call, "crossover")
	if err != nil {
		return 0.0, err
	}
	if barIdx < 1 {
		return 0.0, nil
	}
	curr1, err := e.EvaluateAtBar(arg1, secCtx, barIdx)
	if err != nil {
		return 0.0, err
	}
	curr2, err := e.EvaluateAtBar(arg2, secCtx, barIdx)
	if err != nil {
		return 0.0, err
	}
	prev1, err := e.EvaluateAtBar(arg1, secCtx, barIdx-1)
	if err != nil {
		return 0.0, err
	}
	prev2, err := e.EvaluateAtBar(arg2, secCtx, barIdx-1)
	if err != nil {
		return 0.0, err
	}
	if curr1 > curr2 && prev1 <= prev2 {
		return 1.0, nil
	}
	return 0.0, nil
}

func (e *StreamingBarEvaluator) evaluateCrossunderAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	arg1, arg2, err := extractTwoExpressionArguments(call, "crossunder")
	if err != nil {
		return 0.0, err
	}
	if barIdx < 1 {
		return 0.0, nil
	}
	curr1, err := e.EvaluateAtBar(arg1, secCtx, barIdx)
	if err != nil {
		return 0.0, err
	}
	curr2, err := e.EvaluateAtBar(arg2, secCtx, barIdx)
	if err != nil {
		return 0.0, err
	}
	prev1, err := e.EvaluateAtBar(arg1, secCtx, barIdx-1)
	if err != nil {
		return 0.0, err
	}
	prev2, err := e.EvaluateAtBar(arg2, secCtx, barIdx-1)
	if err != nil {
		return 0.0, err
	}
	if curr1 < curr2 && prev1 >= prev2 {
		return 1.0, nil
	}
	return 0.0, nil
}

func (e *StreamingBarEvaluator) evaluateCrossAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	over, err := e.evaluateCrossoverAtBar(call, secCtx, barIdx)
	if err != nil {
		return 0.0, err
	}
	if over != 0.0 {
		return 1.0, nil
	}
	return e.evaluateCrossunderAtBar(call, secCtx, barIdx)
}

func (e *StreamingBarEvaluator) evaluateFallingAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, length, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length {
		return 0.0, nil
	}
	for i := 0; i < length; i++ {
		curr, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx-i)
		if err != nil {
			return 0.0, err
		}
		prev, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx-i-1)
		if err != nil {
			return 0.0, err
		}
		if curr >= prev {
			return 0.0, nil
		}
	}
	return 1.0, nil
}

func (e *StreamingBarEvaluator) evaluateRisingAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, length, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length {
		return 0.0, nil
	}
	for i := 0; i < length; i++ {
		curr, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx-i)
		if err != nil {
			return 0.0, err
		}
		prev, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx-i-1)
		if err != nil {
			return 0.0, err
		}
		if curr <= prev {
			return 0.0, nil
		}
	}
	return 1.0, nil
}

func (e *StreamingBarEvaluator) evaluateBarsSinceAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	condExpr, err := extractSingleExpressionArgument(call, "barssince")
	if err != nil {
		return 0.0, err
	}

	state, cached := e.barsSinceCache[call]
	if !cached {
		state = NewBarsSinceStateManager(condExpr, e, len(secCtx.Data))
		e.barsSinceCache[call] = state
	}

	return state.ComputeAtBar(secCtx, barIdx)
}

func (e *StreamingBarEvaluator) evaluateCumAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, err := extractSourceOnlyArgument(call, "cum")
	if err != nil {
		return 0.0, err
	}
	cacheKey := buildTACacheKey("cum", expressionKey(sourceExpr), 0)
	if state, exists := e.taStateCache[cacheKey]; exists {
		return state.ComputeAtBar(secCtx, sourceExpr, barIdx)
	}
	state := NewCUMStateManager(len(secCtx.Data), e)
	e.taStateCache[cacheKey] = state
	return state.ComputeAtBar(secCtx, sourceExpr, barIdx)
}

func (e *StreamingBarEvaluator) evaluateMathMaxAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	if len(call.Arguments) == 0 {
		return 0.0, newInsufficientArgumentsError("math.max", 1, 0)
	}
	result := math.Inf(-1)
	for _, arg := range call.Arguments {
		val, err := e.EvaluateAtBar(arg, secCtx, barIdx)
		if err != nil {
			return 0.0, err
		}
		if math.IsNaN(val) {
			return math.NaN(), nil
		}
		if val > result {
			result = val
		}
	}
	return result, nil
}

func (e *StreamingBarEvaluator) evaluateMathMinAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	if len(call.Arguments) == 0 {
		return 0.0, newInsufficientArgumentsError("math.min", 1, 0)
	}
	result := math.Inf(1)
	for _, arg := range call.Arguments {
		val, err := e.EvaluateAtBar(arg, secCtx, barIdx)
		if err != nil {
			return 0.0, err
		}
		if math.IsNaN(val) {
			return math.NaN(), nil
		}
		if val < result {
			result = val
		}
	}
	return result, nil
}

func (e *StreamingBarEvaluator) evaluateMathAbsAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "math.abs")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return 0.0, err
	}
	return math.Abs(val), nil
}
