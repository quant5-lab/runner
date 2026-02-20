package codegen

import "github.com/quant5-lab/runner/ast"

/* CallExpressionHandler processes Pine Script function calls via strategy pattern. */
type CallExpressionHandler interface {
	CanHandle(funcName string) bool
	GenerateCode(g *generator, call *ast.CallExpression) (string, error)
}

/* CallExpressionRouter dispatches call expressions to the first matching handler. */
type CallExpressionRouter struct {
	handlers []CallExpressionHandler
}

func NewCallExpressionRouter() *CallExpressionRouter {
	router := &CallExpressionRouter{
		handlers: make([]CallExpressionHandler, 0),
	}

	router.RegisterHandler(NewMetaFunctionHandler())
	router.RegisterHandler(&PlotFunctionHandler{})
	router.RegisterHandler(NewStrategyActionHandler())
	router.RegisterHandler(NewTradeCollectionCallHandler())
	router.RegisterHandler(&MathCallHandler{})
	router.RegisterHandler(NewValueCallHandler())
	router.RegisterHandler(&TAIndicatorCallHandler{})
	router.RegisterHandler(NewTickerFunctionHandler())
	router.RegisterHandler(&ColorCallHandler{})
	router.RegisterHandler(NewCalendarCallHandler())
	router.RegisterHandler(NewTimeframeFuncCallHandler())
	router.RegisterHandler(&UserDefinedFunctionHandler{})
	router.RegisterHandler(NewStringNamespaceHandler())
	router.RegisterHandler(&UnknownFunctionHandler{})

	return router
}

func (r *CallExpressionRouter) RegisterHandler(handler CallExpressionHandler) {
	r.handlers = append(r.handlers, handler)
}

func (r *CallExpressionRouter) RouteCall(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	for _, handler := range r.handlers {
		canHandle := handler.CanHandle(funcName)
		code, err := handler.GenerateCode(g, call)
		if err != nil {
			return "", err
		}
		if code != "" {
			return code, nil
		}
		if canHandle {
			return "", nil
		}
	}

	return "", nil
}

func extractCallFunctionName(call *ast.CallExpression) string {
	switch callee := call.Callee.(type) {
	case *ast.Identifier:
		return callee.Name
	case *ast.MemberExpression:
		return extractMemberExpressionFullPath(callee)
	}
	return ""
}

func extractMemberExpressionFullPath(expr *ast.MemberExpression) string {
	prop := extractIdentifierName(expr.Property)
	if prop == "" {
		return ""
	}

	switch obj := expr.Object.(type) {
	case *ast.Identifier:
		return obj.Name + "." + prop
	case *ast.MemberExpression:
		objPath := extractMemberExpressionFullPath(obj)
		if objPath == "" {
			return ""
		}
		return objPath + "." + prop
	default:
		return ""
	}
}

func extractIdentifierName(expr ast.Expression) string {
	if id, ok := expr.(*ast.Identifier); ok {
		return id.Name
	}
	return ""
}
