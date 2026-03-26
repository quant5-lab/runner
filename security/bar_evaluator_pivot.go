package security

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func (e *StreamingBarEvaluator) evaluatePivotHighAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, leftBars, rightBars, err := extractPivotArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	cacheKey := buildPivotCacheKey("pivothigh", expressionKey(sourceExpr), leftBars, rightBars)
	state := e.getOrCreatePivotState(cacheKey, pivotKindHigh, leftBars, rightBars, sourceExpr, secCtx)
	return state.ComputeAtBar(secCtx, barIdx)
}

func (e *StreamingBarEvaluator) evaluatePivotLowAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, leftBars, rightBars, err := extractPivotArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	cacheKey := buildPivotCacheKey("pivotlow", expressionKey(sourceExpr), leftBars, rightBars)
	state := e.getOrCreatePivotState(cacheKey, pivotKindLow, leftBars, rightBars, sourceExpr, secCtx)
	return state.ComputeAtBar(secCtx, barIdx)
}

func (e *StreamingBarEvaluator) getOrCreatePivotState(cacheKey string, kind pivotKind, leftBars, rightBars int, sourceExpr ast.Expression, secCtx *context.Context) *PivotStateManager {
	if state, exists := e.pivotStateCache[cacheKey]; exists {
		return state
	}
	state := newPivotStateManager(kind, leftBars, rightBars, len(secCtx.Data), sourceExpr, e)
	e.pivotStateCache[cacheKey] = state
	return state
}

func buildPivotCacheKey(funcName, sourceKey string, leftBars, rightBars int) string {
	return fmt.Sprintf("%s_%s_%d_%d", funcName, sourceKey, leftBars, rightBars)
}

func init() {
	registerCallHandler("ta.pivothigh", (*StreamingBarEvaluator).evaluatePivotHighAtBar)
	registerCallHandler("ta.pivotlow", (*StreamingBarEvaluator).evaluatePivotLowAtBar)
}
