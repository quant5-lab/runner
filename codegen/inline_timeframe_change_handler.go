package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type TimeframeChangeInlineHandler struct{}

func NewTimeframeChangeInlineHandler() *TimeframeChangeInlineHandler {
	return &TimeframeChangeInlineHandler{}
}

func (h *TimeframeChangeInlineHandler) CanHandle(funcName string) bool {
	return funcName == "timeframe.change"
}

func (h *TimeframeChangeInlineHandler) GenerateInline(expr *ast.CallExpression, g *generator) (string, error) {
	if len(expr.Arguments) == 0 {
		return "", fmt.Errorf("timeframe.change requires 1 argument (timeframe string)")
	}

	tfArg, err := g.generateConditionExpression(expr.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("timeframe.change arg: %w", err)
	}

	return buildTimeframeChangeIIFE(tfArg, timeframeChangeBool), nil
}
