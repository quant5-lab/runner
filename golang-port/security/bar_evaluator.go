package security

import (
	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

type BarEvaluator interface {
	EvaluateAtBar(expr ast.Expression, secCtx *context.Context, barIdx int) (float64, error)
}

type StreamingBarEvaluator struct {
	taStateCache map[string]TAStateManager
}

func NewStreamingBarEvaluator() *StreamingBarEvaluator {
	return &StreamingBarEvaluator{
		taStateCache: make(map[string]TAStateManager),
	}
}

func (e *StreamingBarEvaluator) EvaluateAtBar(expr ast.Expression, secCtx *context.Context, barIdx int) (float64, error) {
	switch exp := expr.(type) {
	case *ast.Identifier:
		return evaluateOHLCVAtBar(exp, secCtx, barIdx)
	case *ast.CallExpression:
		return e.evaluateTACallAtBar(exp, secCtx, barIdx)
	case *ast.MemberExpression:
		return 0.0, newUnsupportedExpressionError(exp)
	default:
		return 0.0, newUnsupportedExpressionError(exp)
	}
}

func evaluateOHLCVAtBar(id *ast.Identifier, secCtx *context.Context, barIdx int) (float64, error) {
	if barIdx < 0 || barIdx >= len(secCtx.Data) {
		return 0.0, newBarIndexOutOfRangeError(barIdx, len(secCtx.Data))
	}

	bar := secCtx.Data[barIdx]

	switch id.Name {
	case "close":
		return bar.Close, nil
	case "open":
		return bar.Open, nil
	case "high":
		return bar.High, nil
	case "low":
		return bar.Low, nil
	case "volume":
		return bar.Volume, nil
	default:
		return 0.0, newUnknownIdentifierError(id.Name)
	}
}

func (e *StreamingBarEvaluator) evaluateTACallAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	funcName := extractCallFunctionName(call.Callee)

	switch funcName {
	case "ta.sma":
		return e.evaluateSMAAtBar(call, secCtx, barIdx)
	case "ta.ema":
		return e.evaluateEMAAtBar(call, secCtx, barIdx)
	case "ta.rma":
		return e.evaluateRMAAtBar(call, secCtx, barIdx)
	case "ta.rsi":
		return e.evaluateRSIAtBar(call, secCtx, barIdx)
	default:
		return 0.0, newUnsupportedFunctionError(funcName)
	}
}

func (e *StreamingBarEvaluator) evaluateSMAAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, period, err := extractTAArguments(call)
	if err != nil {
		return 0.0, err
	}

	cacheKey := buildTACacheKey("sma", sourceID.Name, period)
	stateManager := e.getOrCreateTAState(cacheKey, period, secCtx)

	return stateManager.ComputeAtBar(secCtx, sourceID, barIdx)
}

func (e *StreamingBarEvaluator) evaluateEMAAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, period, err := extractTAArguments(call)
	if err != nil {
		return 0.0, err
	}

	cacheKey := buildTACacheKey("ema", sourceID.Name, period)
	stateManager := e.getOrCreateTAState(cacheKey, period, secCtx)

	return stateManager.ComputeAtBar(secCtx, sourceID, barIdx)
}

func (e *StreamingBarEvaluator) evaluateRMAAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, period, err := extractTAArguments(call)
	if err != nil {
		return 0.0, err
	}

	cacheKey := buildTACacheKey("rma", sourceID.Name, period)
	stateManager := e.getOrCreateTAState(cacheKey, period, secCtx)

	return stateManager.ComputeAtBar(secCtx, sourceID, barIdx)
}

func (e *StreamingBarEvaluator) evaluateRSIAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, period, err := extractTAArguments(call)
	if err != nil {
		return 0.0, err
	}

	cacheKey := buildTACacheKey("rsi", sourceID.Name, period)
	stateManager := e.getOrCreateTAState(cacheKey, period, secCtx)

	return stateManager.ComputeAtBar(secCtx, sourceID, barIdx)
}

func (e *StreamingBarEvaluator) getOrCreateTAState(cacheKey string, period int, secCtx *context.Context) TAStateManager {
	if state, exists := e.taStateCache[cacheKey]; exists {
		return state
	}

	state := NewTAStateManager(cacheKey, period, len(secCtx.Data))
	e.taStateCache[cacheKey] = state
	return state
}
