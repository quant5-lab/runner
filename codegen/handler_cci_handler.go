package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* CciHandler generates Commodity Channel Index:
 * (source - sma(source, length)) / (0.015 * mean(|source - sma|, length)) */
type CciHandler struct{}

func (h *CciHandler) CanHandle(funcName string) bool {
	return funcName == "ta.cci" || funcName == "cci"
}

func (h *CciHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractWithDynamic(call, "ta.cci")
	if err != nil {
		return "", err
	}

	if comp.PeriodResult.IsRuntimeDynamic() {
		dynamicGen := NewDynamicPeriodTAGenerator(g)
		code, err := dynamicGen.Generate(varName, "ta.cci", comp.SourceExpr, comp.PeriodResult)
		if err != nil {
			return "", err
		}
		return g.indentCode(comp.Preamble + code), nil
	}

	period := comp.PeriodResult.StaticValue
	warmup := period + comp.AccessGen.GetBaseOffset()
	smaVar := fmt.Sprintf("_%s_sma", varName)
	devVar := fmt.Sprintf("_%s_dev", varName)

	code := g.ind() + fmt.Sprintf("if ctx.BarIndex < %d {\n", warmup)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%s := 0.0\n", smaVar)
	code += g.ind() + fmt.Sprintf("for j := 0; j < %d; j++ { %s += %s }\n", period, smaVar, comp.AccessGen.GenerateLoopValueAccess("j"))
	code += g.ind() + fmt.Sprintf("%s /= %d.0\n", smaVar, period)
	code += g.ind() + fmt.Sprintf("%s := 0.0\n", devVar)
	code += g.ind() + fmt.Sprintf("for j := 0; j < %d; j++ {\n", period)
	g.indent++
	code += g.ind() + fmt.Sprintf("v := %s\n", comp.AccessGen.GenerateLoopValueAccess("j"))
	code += g.ind() + fmt.Sprintf("if v > %s { %s += v - %s } else { %s += %s - v }\n", smaVar, devVar, smaVar, devVar, smaVar)
	g.indent--
	code += g.ind() + "}\n"
	code += g.ind() + fmt.Sprintf("%s /= %d.0\n", devVar, period)
	code += g.ind() + fmt.Sprintf("if %s == 0.0 {\n", devVar)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(0.0)\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set((%s - %s) / (0.015 * %s))\n", varName, comp.AccessGen.GenerateLoopValueAccess("0"), smaVar, devVar)
	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"

	return comp.Preamble + code, nil
}
