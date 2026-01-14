package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

/* UserDefinedFunctionHandler generates calls to user-defined arrow functions */
type UserDefinedFunctionHandler struct{}

func (h *UserDefinedFunctionHandler) CanHandle(funcName string) bool {
	return false
}

func (h *UserDefinedFunctionHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	if g.inArrowFunctionBody {
		if h.isUnprefixedTAFunction(funcName) {
			taHandler := &TAIndicatorCallHandler{}
			return taHandler.generateArrowFunctionTACall(g, call)
		}
	}

	detector := NewUserDefinedFunctionDetector(g.variables)
	if !detector.IsUserDefinedFunction(funcName) {
		return "", nil
	}

	argumentList, err := h.buildArgumentList(g, funcName, call.Arguments)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s(%s)", funcName, argumentList), nil
}

func (h *UserDefinedFunctionHandler) buildArgumentList(g *generator, funcName string, args []ast.Expression) (string, error) {
	contextArg := h.selectContextArgument(g)
	argStrings := []string{contextArg}

	for idx, arg := range args {
		argGen := NewArgumentExpressionGenerator(g, funcName, idx)
		argCode, err := argGen.Generate(arg)
		if err != nil {
			return "", fmt.Errorf("failed to generate argument %d: %w", idx, err)
		}
		argStrings = append(argStrings, argCode)
	}

	return strings.Join(argStrings, ", "), nil
}

func (h *UserDefinedFunctionHandler) selectContextArgument(g *generator) string {
	if g.inArrowFunctionBody {
		return "arrowCtx"
	}
	return "ctx"
}

func (h *UserDefinedFunctionHandler) isUnprefixedTAFunction(funcName string) bool {
	switch funcName {
	case "sma", "ema", "stdev", "rma", "wma":
		return true
	default:
		return false
	}
}
