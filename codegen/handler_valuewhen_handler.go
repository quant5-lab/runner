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
	if len(call.Arguments) < 3 {
		return "", fmt.Errorf("valuewhen requires 3 arguments (condition, source, occurrence)")
	}

	conditionExpr := g.extractSeriesExpression(call.Arguments[0])
	sourceExpr := g.extractSeriesExpression(call.Arguments[1])

	occurrenceArg, ok := call.Arguments[2].(*ast.Literal)
	if !ok {
		return "", fmt.Errorf("valuewhen occurrence must be literal")
	}

	occurrence, err := extractIntegerValue(occurrenceArg)
	if err != nil {
		return "", fmt.Errorf("valuewhen: %w", err)
	}

	return g.generateValuewhen(varName, conditionExpr, sourceExpr, occurrence)
}
