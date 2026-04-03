package codegen

import "github.com/quant5-lab/runner/ast"

/* Adapts CalendarHandler to CallExpressionHandler interface */
type CalendarCallHandler struct {
	handler *CalendarHandler
}

func NewCalendarCallHandler() *CalendarCallHandler {
	return &CalendarCallHandler{handler: NewCalendarHandler()}
}

func (h *CalendarCallHandler) CanHandle(funcName string) bool {
	return h.handler.CanHandle(funcName)
}

func (h *CalendarCallHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)
	if !h.handler.CanHandle(funcName) {
		return "", nil
	}
	return h.handler.GenerateCalendarCall(funcName, call.Arguments, g)
}
