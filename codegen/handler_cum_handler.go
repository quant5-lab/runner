package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type CumHandler struct{}

func (h *CumHandler) CanHandle(funcName string) bool {
	return funcName == "ta.cum" || funcName == "cum"
}

func (h *CumHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 1 {
		return "", fmt.Errorf("ta.cum requires exactly 1 argument")
	}

	sourceExpr := g.extractSeriesExpression(call.Arguments[0])

	return g.generateCum(varName, sourceExpr)
}
