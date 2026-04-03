package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type ArrowValueFunctionGenerator struct {
	exprGenerator ArrowExpressionGenerator
}

func NewArrowValueFunctionGenerator(exprGen ArrowExpressionGenerator) *ArrowValueFunctionGenerator {
	return &ArrowValueFunctionGenerator{
		exprGenerator: exprGen,
	}
}

func (g *ArrowValueFunctionGenerator) CanHandle(funcName string) bool {
	switch funcName {
	case "nz", "na", "fixnan":
		return true
	default:
		return false
	}
}

func (g *ArrowValueFunctionGenerator) Generate(call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	switch funcName {
	case "nz":
		return g.generateNz(call)
	case "na":
		return g.generateNa(call)
	case "fixnan":
		return g.generateFixnan(call)
	default:
		return "", fmt.Errorf("unsupported value function: %s", funcName)
	}
}

func (g *ArrowValueFunctionGenerator) generateNz(call *ast.CallExpression) (string, error) {
	if len(call.Arguments) == 0 {
		return "0", nil
	}

	argCode, err := g.exprGenerator.Generate(call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("nz() argument failed: %w", err)
	}

	replacement := "0"
	if len(call.Arguments) >= 2 {
		replCode, err := g.exprGenerator.Generate(call.Arguments[1])
		if err != nil {
			return "", fmt.Errorf("nz() replacement failed: %w", err)
		}
		replacement = replCode
	}

	return fmt.Sprintf("value.Nz(%s, %s)", argCode, replacement), nil
}

func (g *ArrowValueFunctionGenerator) generateNa(call *ast.CallExpression) (string, error) {
	if len(call.Arguments) == 0 {
		return "true", nil
	}

	argCode, err := g.exprGenerator.Generate(call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("na() argument failed: %w", err)
	}

	return fmt.Sprintf("math.IsNaN(%s)", argCode), nil
}

func (g *ArrowValueFunctionGenerator) generateFixnan(call *ast.CallExpression) (string, error) {
	if len(call.Arguments) == 0 {
		return "", fmt.Errorf("fixnan() requires 1 argument")
	}

	sourceCode, err := g.exprGenerator.Generate(call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("fixnan() source failed: %w", err)
	}

	return fmt.Sprintf("func() float64 { val := %s; if math.IsNaN(val) { return 0.0 }; return val }()", sourceCode), nil
}
