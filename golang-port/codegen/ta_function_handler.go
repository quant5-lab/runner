package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// TAFunctionHandler defines the interface for handling TA function code generation.
// Each TA function (sma, ema, stdev, atr, etc.) has its own handler implementation
// that knows how to generate the appropriate inline code.
//
// This follows the Strategy pattern, replacing switch-case branching with polymorphism.
type TAFunctionHandler interface {
	// CanHandle returns true if this handler can process the given function name.
	// Supports both Pine v4 (e.g., "sma") and v5 (e.g., "ta.sma") syntax.
	CanHandle(funcName string) bool

	// GenerateCode produces the inline calculation code for this TA function.
	// Returns the generated code string or an error if generation fails.
	GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error)
}

// TAFunctionRegistry manages all TA function handlers and routes function calls
// to the appropriate handler based on function name.
//
// This centralizes TA function routing logic, making it trivial to add new
// indicators without modifying existing code (Open/Closed Principle).
type TAFunctionRegistry struct {
	handlers []TAFunctionHandler
}

// NewTAFunctionRegistry creates a registry with all standard TA function handlers.
func NewTAFunctionRegistry() *TAFunctionRegistry {
	return &TAFunctionRegistry{
		handlers: []TAFunctionHandler{
			&SMAHandler{},
			&EMAHandler{},
			&STDEVHandler{},
			&ATRHandler{},
			&RMAHandler{},
			&RSIHandler{},
			&ChangeHandler{},
			&PivotHighHandler{},
			&PivotLowHandler{},
			&CrossoverHandler{},
			&CrossunderHandler{},
			&FixnanHandler{},
		},
	}
}

// FindHandler locates the appropriate handler for the given function name.
// Returns nil if no handler can process this function.
func (r *TAFunctionRegistry) FindHandler(funcName string) TAFunctionHandler {
	for _, handler := range r.handlers {
		if handler.CanHandle(funcName) {
			return handler
		}
	}
	return nil
}

// IsSupported checks if a function name has a registered handler.
func (r *TAFunctionRegistry) IsSupported(funcName string) bool {
	return r.FindHandler(funcName) != nil
}

// GenerateInlineTA generates inline TA calculation code by delegating to
// the appropriate handler. This is the main entry point replacing the old
// switch-case logic.
func (r *TAFunctionRegistry) GenerateInlineTA(g *generator, varName string, funcName string, call *ast.CallExpression) (string, error) {
	handler := r.FindHandler(funcName)
	if handler == nil {
		return "", fmt.Errorf("no handler found for TA function: %s", funcName)
	}
	return handler.GenerateCode(g, varName, call)
}

// normalizeFunctionName converts Pine v4 syntax to v5 (e.g., "sma" -> "ta.sma").
// This ensures consistent function naming across different Pine versions.
func normalizeFunctionName(funcName string) string {
	// Already normalized (ta.xxx format)
	if len(funcName) > 3 && funcName[:3] == "ta." {
		return funcName
	}

	// Known v4 functions that need ta. prefix
	v4Functions := map[string]bool{
		"sma": true, "ema": true, "rma": true, "rsi": true,
		"atr": true, "stdev": true, "change": true,
		"pivothigh": true, "pivotlow": true,
	}

	if v4Functions[funcName] {
		return "ta." + funcName
	}

	return funcName
}
