package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// PercentrankHandler generates code for ta.percentrank(source, length).
type PercentrankHandler struct{}

func (h *PercentrankHandler) CanHandle(funcName string) bool {
	return funcName == "ta.percentrank" || funcName == "percentrank"
}

func (h *PercentrankHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractWithDynamic(call, "ta.percentrank")
	if err != nil {
		return "", err
	}

	if comp.PeriodResult.IsRuntimeDynamic() {
		dynamicGen := NewDynamicPeriodTAGenerator(g)
		code, err := dynamicGen.Generate(varName, "ta.percentrank", comp.SourceExpr, comp.PeriodResult)
		if err != nil {
			return "", err
		}
		return g.indentCode(comp.Preamble + code), nil
	}

	period := comp.PeriodResult.StaticValue
	warmup := period + comp.AccessGen.GetBaseOffset()
	countVar := fmt.Sprintf("_%s_count", varName)
	currentVar := fmt.Sprintf("_%s_cur", varName)

	code := g.ind() + fmt.Sprintf("if ctx.BarIndex < %d {\n", warmup)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%s := %s\n", currentVar, comp.AccessGen.GenerateLoopValueAccess("0"))
	code += g.ind() + fmt.Sprintf("%s := 0\n", countVar)
	code += g.ind() + fmt.Sprintf("for j := 0; j < %d; j++ {\n", period)
	g.indent++
	code += g.ind() + fmt.Sprintf("if %s < %s { %s++ }\n",
		comp.AccessGen.GenerateLoopValueAccess("j"), currentVar, countVar)
	g.indent--
	code += g.ind() + "}\n"
	code += g.ind() + fmt.Sprintf("%sSeries.Set(float64(%s) / %d.0 * 100.0)\n", varName, countVar, period)
	g.indent--
	code += g.ind() + "}\n"

	return comp.Preamble + code, nil
}
