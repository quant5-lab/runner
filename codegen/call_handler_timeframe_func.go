package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type TimeframeFuncCallHandler struct{}

func NewTimeframeFuncCallHandler() *TimeframeFuncCallHandler {
	return &TimeframeFuncCallHandler{}
}

func (h *TimeframeFuncCallHandler) CanHandle(funcName string) bool {
	return funcName == "timeframe.in_seconds" ||
		funcName == "timeframe.from_seconds" ||
		funcName == "timeframe.change"
}

func (h *TimeframeFuncCallHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	switch funcName {
	case "timeframe.in_seconds":
		return h.generateInSeconds(call.Arguments, g)
	case "timeframe.from_seconds":
		return h.generateFromSeconds(call.Arguments, g)
	case "timeframe.change":
		return h.generateChange(call.Arguments, g)
	default:
		return "", nil
	}
}

func (h *TimeframeFuncCallHandler) generateInSeconds(args []ast.Expression, g *generator) (string, error) {
	if len(args) == 0 {
		return "float64(context.TimeframeToSeconds(ctx.Timeframe))", nil
	}

	tfExpr, err := g.generateConditionExpression(args[0])
	if err != nil {
		return "", fmt.Errorf("timeframe.in_seconds arg: %w", err)
	}

	return fmt.Sprintf("float64(context.TimeframeToSeconds(%s))", tfExpr), nil
}

func (h *TimeframeFuncCallHandler) generateFromSeconds(args []ast.Expression, g *generator) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("timeframe.from_seconds requires 1 argument")
	}

	secExpr, err := g.generateConditionExpression(args[0])
	if err != nil {
		return "", fmt.Errorf("timeframe.from_seconds arg: %w", err)
	}

	return fmt.Sprintf("context.TimeframeFromSeconds(int64(%s))", secExpr), nil
}

func (h *TimeframeFuncCallHandler) generateChange(args []ast.Expression, g *generator) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("timeframe.change requires 1 argument")
	}

	tfArg, err := g.generateConditionExpression(args[0])
	if err != nil {
		return "", fmt.Errorf("timeframe.change arg: %w", err)
	}

	return buildTimeframeChangeIIFE(tfArg, timeframeChangeFloat), nil
}
