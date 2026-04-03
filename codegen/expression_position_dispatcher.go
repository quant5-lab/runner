package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

type ExpressionPositionDispatcher struct {
	router *CallExpressionRouter
}

func NewExpressionPositionDispatcher(router *CallExpressionRouter) *ExpressionPositionDispatcher {
	return &ExpressionPositionDispatcher{router: router}
}

func (d *ExpressionPositionDispatcher) Dispatch(g *generator, call *ast.CallExpression) (string, error) {
	funcName := g.extractFunctionName(call.Callee)

	if d.router == nil {
		return "", fmt.Errorf("unsupported function in expression position: %s", funcName)
	}

	code, err := d.router.RouteCall(g, call)
	if err != nil {
		return "", err
	}

	if d.isActionableCode(code) {
		return strings.TrimSpace(code), nil
	}

	return "", fmt.Errorf("unsupported function in expression position: %s", funcName)
}

func (d *ExpressionPositionDispatcher) isActionableCode(code string) bool {
	trimmed := strings.TrimSpace(code)
	if trimmed == "" {
		return false
	}
	return !strings.HasPrefix(trimmed, "//")
}
