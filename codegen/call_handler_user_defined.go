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

	detector := NewUserDefinedFunctionDetector(g.variables)
	if !detector.IsUserDefinedFunction(funcName) {
		return "", nil
	}

	if g.chartOnlyUDFs[funcName] {
		return "math.NaN()", nil
	}

	argumentList, err := h.buildArgumentList(g, funcName, call.Arguments)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s(%s)", funcName, argumentList), nil
}

func (h *UserDefinedFunctionHandler) buildArgumentList(g *generator, funcName string, args []ast.Expression) (string, error) {
	var contextArg string
	if g.inArrowFunctionBody {
		contextArg = "arrowCtx"
	} else {
		contextArg = g.arrowContextLifecycle.AllocateContextVariable(funcName)
	}
	argStrings := []string{contextArg}

	for idx, arg := range args {
		argGen := NewArgumentExpressionGenerator(g, funcName, idx)
		argCode, err := argGen.Generate(arg)
		if err != nil {
			return "", fmt.Errorf("failed to generate argument %d: %w", idx, err)
		}
		argStrings = append(argStrings, argCode)
	}

	if g.arrowCaptureRegistry != nil {
		for _, cap := range g.arrowCaptureRegistry.Get(funcName) {
			argStrings = append(argStrings, cap.GoCallSiteExpression(g.constants))
		}
	}

	return strings.Join(argStrings, ", "), nil
}
