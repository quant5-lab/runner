package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* BbwHandler generates Bollinger Bands Width:
 * 2 * mult * stdev(source, length) / sma(source, length) */
type BbwHandler struct{}

func (h *BbwHandler) CanHandle(funcName string) bool {
	return funcName == "ta.bbw" || funcName == "bbw"
}

func (h *BbwHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractWithDynamic(call, "ta.bbw")
	if err != nil {
		return "", err
	}

	mult := 2.0
	if len(call.Arguments) >= 3 {
		if lit, ok := call.Arguments[2].(*ast.Literal); ok {
			if v, ok := lit.Value.(float64); ok {
				mult = v
			}
		}
	}

	if comp.PeriodResult.IsRuntimeDynamic() {
		dynamicGen := NewDynamicPeriodTAGenerator(g)
		code, err := dynamicGen.GenerateWithEmitter(varName, DynamicBBWEmitter{mult: mult}, comp.SourceExpr, comp.PeriodResult)
		if err != nil {
			return "", err
		}
		return g.indentCode(comp.Preamble + code), nil
	}

	period := comp.PeriodResult.StaticValue
	warmup := period + comp.AccessGen.GetBaseOffset()
	smaVar := fmt.Sprintf("_%s_sma", varName)
	varV := fmt.Sprintf("_%s_v", varName)
	sdVar := fmt.Sprintf("_%s_sd", varName)

	code := g.ind() + fmt.Sprintf("if ctx.BarIndex < %d {\n", warmup)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%s := 0.0\n", smaVar)
	code += g.ind() + fmt.Sprintf("for j := 0; j < %d; j++ { %s += %s }\n", period, smaVar, comp.AccessGen.GenerateLoopValueAccess("j"))
	code += g.ind() + fmt.Sprintf("%s /= %d.0\n", smaVar, period)
	code += g.ind() + fmt.Sprintf("%s := 0.0\n", sdVar)
	code += g.ind() + fmt.Sprintf("for j := 0; j < %d; j++ {\n", period)
	g.indent++
	code += g.ind() + fmt.Sprintf("%s := %s - %s\n", varV, comp.AccessGen.GenerateLoopValueAccess("j"), smaVar)
	code += g.ind() + fmt.Sprintf("%s += %s * %s\n", sdVar, varV, varV)
	g.indent--
	code += g.ind() + "}\n"
	code += g.ind() + fmt.Sprintf("%s = math.Sqrt(%s / %d.0)\n", sdVar, sdVar, period)
	code += g.ind() + fmt.Sprintf("if %s == 0.0 {\n", smaVar)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(0.0)\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(2.0 * %g * %s / %s)\n", varName, mult, sdVar, smaVar)
	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"

	return comp.Preamble + code, nil
}
