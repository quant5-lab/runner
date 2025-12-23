package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

// TAIndicatorCallHandler handles TA indicator calls in expression context.
//
// Handles: ta.sma(), ta.ema(), ta.stdev(), ta.crossover(), etc.
// Behavior: These are handled in variable declarations, not as statements
//
// Note: This is separate from TAIndicatorBuilder which generates declaration code
type TAIndicatorCallHandler struct{}

func (h *TAIndicatorCallHandler) CanHandle(funcName string) bool {
	switch funcName {
	case "ta.sma", "ta.ema", "ta.stdev", "ta.rma", "ta.wma",
		"ta.crossover", "ta.crossunder",
		"ta.change", "ta.pivothigh", "ta.pivotlow",
		"fixnan", "valuewhen":
		return true
	default:
		return false
	}
}

func (h *TAIndicatorCallHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	// Check if this is actually a user-defined function (not a TA function)
	if varType, exists := g.variables[funcName]; exists && varType == "function" {
		return "", nil // Let UserDefinedFunctionHandler handle it
	}

	// Arrow function context: Generate function call expression
	if g.inArrowFunctionBody {
		return h.generateArrowFunctionTACall(g, call)
	}

	// Series context: TA indicator calls are handled in variable declarations
	return "", nil
}

func (h *TAIndicatorCallHandler) generateArrowFunctionTACall(g *generator, call *ast.CallExpression) (string, error) {
	// Create a simple expression generator that uses the OLD generator methods
	// This is a fallback for cases where arrow-aware context is not available
	exprGen := &legacyArrowExpressionGenerator{gen: g}
	generator := NewArrowFunctionTACallGenerator(g, exprGen)
	return generator.Generate(call)
}

type legacyArrowExpressionGenerator struct {
	gen *generator
}

func (e *legacyArrowExpressionGenerator) Generate(expr ast.Expression) (string, error) {
	return e.gen.generateArrowFunctionExpression(expr)
}
