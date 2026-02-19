package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* CrossHandler detects any crossing: crossover OR crossunder. */
type CrossHandler struct{}

func (h *CrossHandler) CanHandle(funcName string) bool {
	return funcName == "ta.cross" || funcName == "cross"
}

func (h *CrossHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 2 {
		return "", fmt.Errorf("ta.cross requires 2 arguments")
	}

	series1 := g.extractSeriesExpression(call.Arguments[0])
	series2 := g.extractSeriesExpression(call.Arguments[1])

	prev1Var := varName + "_prev1"
	prev2Var := varName + "_prev2"
	prev2Value := ensureFloat64Literal(g.convertSeriesAccessToPrev(series2))

	code := g.ind() + "if i > 0 {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%s := %s\n", prev1Var, g.convertSeriesAccessToPrev(series1))
	code += g.ind() + fmt.Sprintf("%s := %s\n", prev2Var, prev2Value)
	crossover := fmt.Sprintf("%s > %s && %s <= %s", series1, series2, prev1Var, prev2Var)
	crossunder := fmt.Sprintf("%s < %s && %s >= %s", series1, series2, prev1Var, prev2Var)
	code += g.ind() + fmt.Sprintf("%sSeries.Set(func() float64 { if (%s) || (%s) { return 1.0 }; return 0.0 }())\n", varName, crossover, crossunder)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(0.0)\n", varName)
	g.indent--
	code += g.ind() + "}\n"

	return code, nil
}
