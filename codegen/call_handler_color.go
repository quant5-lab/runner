package codegen

import "github.com/quant5-lab/runner/ast"

/* Positioned before UserDefinedFunctionHandler and UnknownFunctionHandler */
type ColorCallHandler struct {
	colorHandler *ColorHandler
}

func (h *ColorCallHandler) CanHandle(funcName string) bool {
	if h.colorHandler == nil {
		h.colorHandler = NewColorHandler()
	}
	return h.colorHandler.CanHandle(funcName)
}

func (h *ColorCallHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	if !h.CanHandle(funcName) {
		return "", nil
	}

	if h.colorHandler == nil {
		h.colorHandler = NewColorHandler()
	}

	return h.colorHandler.GenerateColorCall(funcName, call.Arguments, g)
}
