package codegen

import "github.com/quant5-lab/runner/ast"

type MathCallHandler struct {
	mathHandler *MathHandler
}

func (h *MathCallHandler) CanHandle(funcName string) bool {
	if h.mathHandler == nil {
		h.mathHandler = NewMathHandler()
	}
	return h.mathHandler.CanHandle(funcName)
}

func (h *MathCallHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	if !h.CanHandle(funcName) {
		return "", nil
	}

	if h.mathHandler == nil {
		h.mathHandler = NewMathHandler()
	}

	return h.mathHandler.GenerateMathCall(funcName, call.Arguments, g)
}
