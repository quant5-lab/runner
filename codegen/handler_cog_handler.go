package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* CogHandler generates Center of Gravity oscillator:
 * -sum(source[i] * (i+1), length) / sum(source[i], length) */
type CogHandler struct{}

func (h *CogHandler) CanHandle(funcName string) bool {
	return funcName == "ta.cog" || funcName == "cog"
}

func (h *CogHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractWithDynamic(call, "ta.cog")
	if err != nil {
		return "", err
	}

	if comp.PeriodResult.IsRuntimeDynamic() {
		dynamicGen := NewDynamicPeriodTAGenerator(g)
		code, err := dynamicGen.Generate(varName, "ta.cog", comp.SourceExpr, comp.PeriodResult)
		if err != nil {
			return "", err
		}
		return g.indentCode(comp.Preamble + code), nil
	}

	period := comp.PeriodResult.StaticValue
	warmup := period + comp.AccessGen.GetBaseOffset()
	numVar := fmt.Sprintf("_%s_num", varName)
	denVar := fmt.Sprintf("_%s_den", varName)

	code := g.ind() + fmt.Sprintf("if ctx.BarIndex < %d {\n", warmup)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%s, %s := 0.0, 0.0\n", numVar, denVar)
	code += g.ind() + fmt.Sprintf("for j := 0; j < %d; j++ {\n", period)
	g.indent++
	code += g.ind() + fmt.Sprintf("v := %s\n", comp.AccessGen.GenerateLoopValueAccess("j"))
	code += g.ind() + fmt.Sprintf("%s += v * float64(j+1)\n", numVar)
	code += g.ind() + fmt.Sprintf("%s += v\n", denVar)
	g.indent--
	code += g.ind() + "}\n"
	code += g.ind() + fmt.Sprintf("if %s == 0.0 {\n", denVar)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(0.0)\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(-%s / %s)\n", varName, numVar, denVar)
	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"

	return comp.Preamble + code, nil
}
