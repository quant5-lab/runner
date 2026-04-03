package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

/*
	TAIndicatorCallHandler routes TA indicator calls based on context.

Series context: handled in variable declarations via TAIndicatorBuilder.
Arrow context: delegated to ArrowFunctionTACallGenerator for IIFE/inline patterns.
*/
type TAIndicatorCallHandler struct{}

func (h *TAIndicatorCallHandler) CanHandle(funcName string) bool {
	return sharedTASignatures.Contains(funcName) && !sharedTASignatures.IsTupleFunction(funcName)
}

func (h *TAIndicatorCallHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	if varType, exists := g.variables[funcName]; exists && varType == "function" {
		return "", nil
	}

	if g.inArrowFunctionBody {
		return h.generateArrowFunctionTACall(g, call)
	}

	return "", nil
}

func (h *TAIndicatorCallHandler) generateArrowFunctionTACall(g *generator, call *ast.CallExpression) (string, error) {
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
