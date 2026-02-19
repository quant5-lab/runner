package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* WprHandler generates Williams %R:
 * (highest(high, length) - close) / (highest(high, length) - lowest(low, length)) * -100 */
type WprHandler struct{}

func (h *WprHandler) CanHandle(funcName string) bool {
	return funcName == "ta.wpr" || funcName == "wpr"
}

func (h *WprHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	period, err := extractSinglePeriodArgument(g, call, "ta.wpr")
	if err != nil {
		return "", err
	}

	warmup := period - 1
	highVar := fmt.Sprintf("_%s_hh", varName)
	lowVar := fmt.Sprintf("_%s_ll", varName)
	denomVar := fmt.Sprintf("_%s_denom", varName)

	code := g.ind() + fmt.Sprintf("if ctx.BarIndex < %d {\n", warmup)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%s := ctx.Data[ctx.BarIndex].High\n", highVar)
	code += g.ind() + fmt.Sprintf("%s := ctx.Data[ctx.BarIndex].Low\n", lowVar)
	code += g.ind() + fmt.Sprintf("for j := 1; j < %d; j++ {\n", period)
	g.indent++
	code += g.ind() + "bar := ctx.Data[ctx.BarIndex-j]\n"
	code += g.ind() + fmt.Sprintf("if bar.High > %s { %s = bar.High }\n", highVar, highVar)
	code += g.ind() + fmt.Sprintf("if bar.Low < %s { %s = bar.Low }\n", lowVar, lowVar)
	g.indent--
	code += g.ind() + "}\n"
	code += g.ind() + fmt.Sprintf("%s := %s - %s\n", denomVar, highVar, lowVar)
	code += g.ind() + fmt.Sprintf("if %s == 0.0 {\n", denomVar)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(0.0)\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set((ctx.Data[ctx.BarIndex].Close - %s) / %s * 100.0)\n", varName, highVar, denomVar)
	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"

	return code, nil
}
