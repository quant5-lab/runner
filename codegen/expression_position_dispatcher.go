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
	code, _, err := d.dispatch(g, call, "float64", "generator.variable_init_unknown")
	return code, err
}

func (d *ExpressionPositionDispatcher) DispatchString(g *generator, call *ast.CallExpression, source string) (string, error) {
	code, _, err := d.dispatch(g, call, "string", source)
	return code, err
}

// DispatchStringAnnotation reports whether the returned string expression is a
// feature-gap sentinel standing in for an unsupported function rather than a
// carried value.
func (d *ExpressionPositionDispatcher) DispatchStringAnnotation(g *generator, call *ast.CallExpression, source string) (string, bool, error) {
	return d.dispatch(g, call, "string", source)
}

func (d *ExpressionPositionDispatcher) dispatch(g *generator, call *ast.CallExpression, valueType, source string) (string, bool, error) {
	funcName := g.extractFunctionName(call.Callee)

	if d.router == nil {
		return "", false, fmt.Errorf("unsupported function in expression position: %s", funcName)
	}

	code, err := d.router.RouteCall(g, call)
	if err != nil {
		return "", false, err
	}

	if d.isActionableCode(code) {
		return strings.TrimSpace(code), false, nil
	}

	// Empty output means a known handler declined to produce an expression value.
	if code == "" {
		return "", false, fmt.Errorf("unsupported function in expression position: %s (no actionable code generated)", funcName)
	}

	// Namespaced unknown functions (ta.x, request.x) indicate an unsupported specific feature — error.
	if strings.Contains(funcName, ".") {
		return "", false, fmt.Errorf("unsupported function in expression position: %s", funcName)
	}

	return featureGapExpression(funcName, source, valueType, call.Location), true, nil
}

func (d *ExpressionPositionDispatcher) isActionableCode(code string) bool {
	trimmed := strings.TrimSpace(code)
	if trimmed == "" {
		return false
	}
	return !strings.HasPrefix(trimmed, "//")
}

func featureGapExpression(funcName, source, valueType string, location ast.SourceLocation) string {
	if valueType == "string" {
		return runtimeFeatureGapString(funcName, source, location)
	}
	return runtimeFeatureGap(funcName, source, location)
}
