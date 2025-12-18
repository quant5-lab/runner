package codegen

import "github.com/quant5-lab/runner/ast"

// CallExpressionHandler processes specific Pine Script function calls.
//
// Design: Strategy pattern for call expression handling
// - Each handler type implements this interface
// - Router delegates to appropriate handler
// - Open/Closed: Add handlers without modifying existing code
type CallExpressionHandler interface {
	// CanHandle returns true if this handler processes the given function name
	CanHandle(funcName string) bool

	// GenerateCode produces Go code for the call expression
	// Returns: (generated code, error)
	// Empty string = handled but produces no immediate code (e.g., declarations)
	GenerateCode(g *generator, call *ast.CallExpression) (string, error)
}

// CallExpressionRouter delegates call expressions to registered handlers.
//
// Responsibilities:
//   - Extract function name from CallExpression
//   - Find appropriate handler via CanHandle()
//   - Delegate code generation to handler
//
// Design: Chain of Responsibility + Registry pattern
type CallExpressionRouter struct {
	handlers []CallExpressionHandler
}

// NewCallExpressionRouter creates router with standard handlers
func NewCallExpressionRouter() *CallExpressionRouter {
	router := &CallExpressionRouter{
		handlers: make([]CallExpressionHandler, 0),
	}

	// Register handlers in priority order
	router.RegisterHandler(&MetaFunctionHandler{})
	router.RegisterHandler(&PlotFunctionHandler{})
	router.RegisterHandler(&StrategyActionHandler{})
	router.RegisterHandler(&TAIndicatorCallHandler{})
	router.RegisterHandler(&UnknownFunctionHandler{})

	return router
}

// RegisterHandler adds a handler to the chain
func (r *CallExpressionRouter) RegisterHandler(handler CallExpressionHandler) {
	r.handlers = append(r.handlers, handler)
}

// RouteCall finds appropriate handler and generates code
func (r *CallExpressionRouter) RouteCall(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	for _, handler := range r.handlers {
		if handler.CanHandle(funcName) {
			return handler.GenerateCode(g, call)
		}
	}

	// Should never reach here if UnknownFunctionHandler is registered
	return "", nil
}

// extractCallFunctionName extracts function name from CallExpression.Callee
//
// Examples:
//   - Identifier "plot" → "plot"
//   - MemberExpression "ta.sma" → "ta.sma"
//   - MemberExpression "strategy.entry" → "strategy.entry"
func extractCallFunctionName(call *ast.CallExpression) string {
	switch callee := call.Callee.(type) {
	case *ast.Identifier:
		return callee.Name
	case *ast.MemberExpression:
		obj := extractIdentifierName(callee.Object)
		prop := extractIdentifierName(callee.Property)
		if obj != "" && prop != "" {
			return obj + "." + prop
		}
	}
	return ""
}

func extractIdentifierName(expr ast.Expression) string {
	if id, ok := expr.(*ast.Identifier); ok {
		return id.Name
	}
	return ""
}
