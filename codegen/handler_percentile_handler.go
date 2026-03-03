package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// PercentileNearestRankHandler generates code for ta.percentile_nearest_rank(source, length, percentage).
type PercentileNearestRankHandler struct{}

func (h *PercentileNearestRankHandler) CanHandle(funcName string) bool {
	return funcName == "ta.percentile_nearest_rank" || funcName == "percentile_nearest_rank"
}

func (h *PercentileNearestRankHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractWithDynamic(call, "ta.percentile_nearest_rank")
	if err != nil {
		return "", err
	}

	pct := 50.0
	if len(call.Arguments) >= 3 {
		if lit, ok := call.Arguments[2].(*ast.Literal); ok {
			if v, ok := lit.Value.(float64); ok {
				pct = v
			}
		}
	}

	if comp.PeriodResult.IsRuntimeDynamic() {
		return "", fmt.Errorf("ta.percentile_nearest_rank does not support runtime dynamic period")
	}

	period := comp.PeriodResult.StaticValue
	warmup := period + comp.AccessGen.GetBaseOffset()
	windowVar := fmt.Sprintf("_%s_w", varName)

	code := g.ind() + fmt.Sprintf("if ctx.BarIndex < %d {\n", warmup)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%s := make([]float64, %d)\n", windowVar, period)
	code += g.ind() + fmt.Sprintf("for j := 0; j < %d; j++ { %s[j] = %s }\n",
		period, windowVar, comp.AccessGen.GenerateLoopValueAccess("j"))
	code += g.ind() + fmt.Sprintf("sort.Float64s(%s)\n", windowVar)
	idx := fmt.Sprintf("int(math.Ceil(%g/100.0*%d.0))-1", pct, period)
	code += g.ind() + fmt.Sprintf("_%s_idx := %s\n", varName, idx)
	code += g.ind() + fmt.Sprintf("if _%s_idx < 0 { _%s_idx = 0 }\n", varName, varName)
	code += g.ind() + fmt.Sprintf("if _%s_idx >= %d { _%s_idx = %d }\n", varName, period, varName, period-1)
	code += g.ind() + fmt.Sprintf("%sSeries.Set(%s[_%s_idx])\n", varName, windowVar, varName)
	g.indent--
	code += g.ind() + "}\n"

	return comp.Preamble + code, nil
}

// PercentileLinearInterpolationHandler generates code for ta.percentile_linear_interpolation(source, length, percentage).
type PercentileLinearInterpolationHandler struct{}

func (h *PercentileLinearInterpolationHandler) CanHandle(funcName string) bool {
	return funcName == "ta.percentile_linear_interpolation" || funcName == "percentile_linear_interpolation"
}

func (h *PercentileLinearInterpolationHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractWithDynamic(call, "ta.percentile_linear_interpolation")
	if err != nil {
		return "", err
	}

	pct := 50.0
	if len(call.Arguments) >= 3 {
		if lit, ok := call.Arguments[2].(*ast.Literal); ok {
			if v, ok := lit.Value.(float64); ok {
				pct = v
			}
		}
	}

	if comp.PeriodResult.IsRuntimeDynamic() {
		return "", fmt.Errorf("ta.percentile_linear_interpolation does not support runtime dynamic period")
	}

	period := comp.PeriodResult.StaticValue
	warmup := period + comp.AccessGen.GetBaseOffset()
	windowVar := fmt.Sprintf("_%s_w", varName)

	code := g.ind() + fmt.Sprintf("if ctx.BarIndex < %d {\n", warmup)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%s := make([]float64, %d)\n", windowVar, period)
	code += g.ind() + fmt.Sprintf("for j := 0; j < %d; j++ { %s[j] = %s }\n",
		period, windowVar, comp.AccessGen.GenerateLoopValueAccess("j"))
	code += g.ind() + fmt.Sprintf("sort.Float64s(%s)\n", windowVar)
	code += g.ind() + fmt.Sprintf("_%s_rank := %g / 100.0 * float64(%d-1)\n", varName, pct, period)
	code += g.ind() + fmt.Sprintf("_%s_lower := int(math.Floor(_%s_rank))\n", varName, varName)
	code += g.ind() + fmt.Sprintf("_%s_upper := _%s_lower + 1\n", varName, varName)
	code += g.ind() + fmt.Sprintf("_%s_frac := _%s_rank - float64(_%s_lower)\n", varName, varName, varName)
	code += g.ind() + fmt.Sprintf("if _%s_upper >= %d {\n", varName, period)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(%s[%d-1])\n", varName, windowVar, period)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(%s[_%s_lower] + _%s_frac*(%s[_%s_upper]-%s[_%s_lower]))\n",
		varName, windowVar, varName, varName, windowVar, varName, windowVar, varName)
	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"

	return comp.Preamble + code, nil
}
