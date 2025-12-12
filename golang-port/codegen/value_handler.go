package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* ValueHandler generates inline code for Pine Script value functions (na, nz, fixnan) */
type ValueHandler struct{}

func NewValueHandler() *ValueHandler {
	return &ValueHandler{}
}

func (vh *ValueHandler) CanHandle(funcName string) bool {
	switch funcName {
	case "na", "nz", "fixnan":
		return true
	default:
		return false
	}
}

func (vh *ValueHandler) GenerateInlineCall(funcName string, args []ast.Expression, g *generator) (string, error) {
	switch funcName {
	case "na":
		return vh.generateNa(args, g)
	case "nz":
		return vh.generateNz(args, g)
	default:
		return "", fmt.Errorf("unsupported value function: %s", funcName)
	}
}

func (vh *ValueHandler) generateNa(args []ast.Expression, g *generator) (string, error) {
	if len(args) == 0 {
		return "true", nil
	}

	argCode, err := g.generateConditionExpression(args[0])
	if err != nil {
		return "", fmt.Errorf("na() argument generation failed: %w", err)
	}

	return fmt.Sprintf("math.IsNaN(%s)", argCode), nil
}

func (vh *ValueHandler) generateNz(args []ast.Expression, g *generator) (string, error) {
	if len(args) == 0 {
		return "0", nil
	}

	argCode, err := g.generateConditionExpression(args[0])
	if err != nil {
		return "", fmt.Errorf("nz() argument generation failed: %w", err)
	}

	replacement := "0"
	if len(args) >= 2 {
		replCode, err := g.generateConditionExpression(args[1])
		if err != nil {
			return "", fmt.Errorf("nz() replacement generation failed: %w", err)
		}
		replacement = replCode
	}

	return fmt.Sprintf("value.Nz(%s, %s)", argCode, replacement), nil
}
