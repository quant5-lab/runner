package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* ValuewhenHandler generates inline code for valuewhen calculations */
type ValuewhenHandler struct{}

func (h *ValuewhenHandler) CanHandle(funcName string) bool {
	return funcName == "ta.valuewhen" || funcName == "valuewhen"
}

func (h *ValuewhenHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	resolver := newNamedArgResolver("condition", "source", "occurrence")
	args := resolver.resolve(call.Arguments)

	if len(args) < 3 {
		return "", fmt.Errorf("valuewhen requires 3 arguments (condition, source, occurrence)")
	}

	conditionExpr := g.extractSeriesExpression(args[0])
	sourceExpr := g.extractSeriesExpression(args[1])

	occurrenceArg, ok := args[2].(*ast.Literal)
	if !ok {
		return "", fmt.Errorf("valuewhen occurrence must be literal")
	}

	occurrence, err := extractIntegerValue(occurrenceArg)
	if err != nil {
		return "", fmt.Errorf("valuewhen: %w", err)
	}

	return g.generateValuewhen(varName, conditionExpr, sourceExpr, occurrence)
}
