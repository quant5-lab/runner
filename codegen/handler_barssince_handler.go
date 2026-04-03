package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* BarsSinceHandler generates inline code for ta.barssince(condition).
 * Returns the number of bars since condition was last true, or NaN if never true.
 * Uses ForwardSeriesBuffer for state: previous counter is read from Get(1). */
type BarsSinceHandler struct{}

func (h *BarsSinceHandler) CanHandle(funcName string) bool {
	return funcName == "ta.barssince" || funcName == "barssince"
}

func (h *BarsSinceHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 {
		return "", fmt.Errorf("ta.barssince requires 1 argument (condition)")
	}

	conditionExpr := g.extractSeriesExpression(call.Arguments[0])

	code := g.ind() + fmt.Sprintf("if value.IsTrue(%s) {\n", conditionExpr)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(0.0)\n", varName)
	g.indent--
	code += g.ind() + fmt.Sprintf("} else if i > 0 && !math.IsNaN(%sSeries.Get(1)) {\n", varName)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(%sSeries.Get(1) + 1.0)\n", varName, varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "}\n"

	return code, nil
}
