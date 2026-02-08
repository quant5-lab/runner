package codegen

import "github.com/quant5-lab/runner/ast"

/*
	ValueCallHandler adapts ValueHandler to CallExpressionHandler interface.

Ensures nz/na functions route through ValueHandler before reaching
TAIndicatorCallHandler in the router chain.
*/
type ValueCallHandler struct {
	valueHandler *ValueHandler
}

func NewValueCallHandler() *ValueCallHandler {
	return &ValueCallHandler{valueHandler: NewValueHandler()}
}

func (h *ValueCallHandler) CanHandle(funcName string) bool {
	return h.valueHandler.CanHandle(funcName)
}

func (h *ValueCallHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)
	if !h.valueHandler.CanHandle(funcName) {
		return "", nil
	}
	return h.valueHandler.GenerateInlineCall(funcName, call.Arguments, g)
}
