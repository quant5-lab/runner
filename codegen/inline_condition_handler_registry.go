package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* InlineConditionHandlerRegistry dispatches inline function generation to specialized handlers.
 * Replaces large switch statements with handler pattern following SOLID/DRY/KISS principles.
 */
type InlineConditionHandlerRegistry struct {
	handlers []InlineConditionHandler
}

func NewInlineConditionHandlerRegistry() *InlineConditionHandlerRegistry {
	return &InlineConditionHandlerRegistry{
		handlers: []InlineConditionHandler{
			NewValueHandler(),
			NewMathHandler(),
			NewTimeHandler(""),
			NewDevInlineHandler(),
			NewCrossoverInlineHandler(),
			NewCrossunderInlineHandler(),
			NewChangeInlineHandler(),
			NewSecurityInlineHandler(),
			NewLowestInlineHandler(),
			NewHighestInlineHandler(),
		},
	}
}

/* GenerateInline finds a handler that can handle funcName and generates inline expression */
func (r *InlineConditionHandlerRegistry) GenerateInline(funcName string, expr *ast.CallExpression, g *generator) (string, error) {
	for _, handler := range r.handlers {
		if handler.CanHandle(funcName) {
			return handler.GenerateInline(expr, g)
		}
	}
	return "", fmt.Errorf("unsupported inline function in condition: %s", funcName)
}

/* CanHandle checks if any handler supports the given function name */
func (r *InlineConditionHandlerRegistry) CanHandle(funcName string) bool {
	for _, handler := range r.handlers {
		if handler.CanHandle(funcName) {
			return true
		}
	}
	return false
}
