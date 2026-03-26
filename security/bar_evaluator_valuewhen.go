package security

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func (e *StreamingBarEvaluator) evaluateValuewhenAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	condExpr, srcExpr, occurrence, err := extractValuewhenArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}

	cacheKey := fmt.Sprintf("valuewhen_%s_%s_%d", expressionKey(condExpr), expressionKey(srcExpr), occurrence)
	state, exists := e.valuewhenCache[cacheKey]
	if !exists {
		state = newValuewhenStateManager(condExpr, srcExpr, occurrence, len(secCtx.Data), e)
		e.valuewhenCache[cacheKey] = state
	}

	return state.ComputeAtBar(secCtx, barIdx)
}

func init() {
	registerCallHandlerAliases((*StreamingBarEvaluator).evaluateValuewhenAtBar, "ta.valuewhen", "valuewhen")
}
