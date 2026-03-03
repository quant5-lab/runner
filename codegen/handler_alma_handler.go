package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// AlmaHandler generates code for ta.alma(source, length, offset=0.85, sigma=6):
// Arnaud Legoux Moving Average with Gaussian weighting.
type AlmaHandler struct{}

func (h *AlmaHandler) CanHandle(funcName string) bool {
	return funcName == "ta.alma" || funcName == "alma"
}

func (h *AlmaHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.Extract(call, "ta.alma")
	if err != nil {
		return "", err
	}

	offset := 0.85
	sigma := 6.0
	if len(call.Arguments) >= 3 {
		if lit, ok := call.Arguments[2].(*ast.Literal); ok {
			if v, ok := lit.Value.(float64); ok {
				offset = v
			}
		}
	}
	if len(call.Arguments) >= 4 {
		if lit, ok := call.Arguments[3].(*ast.Literal); ok {
			if v, ok := lit.Value.(float64); ok {
				sigma = v
			}
		}
	}

	period := comp.Period
	warmup := period + comp.AccessGen.GetBaseOffset()
	m := offset * float64(period-1)
	s := float64(period) / sigma

	wVar := fmt.Sprintf("_%s_w", varName)
	wSumVar := fmt.Sprintf("_%s_wsum", varName)

	code := g.ind() + fmt.Sprintf("if ctx.BarIndex < %d {\n", warmup)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%s := [%d]float64{}\n", wVar, period)
	code += g.ind() + fmt.Sprintf("%s := 0.0\n", wSumVar)
	code += g.ind() + fmt.Sprintf("for j := 0; j < %d; j++ {\n", period)
	g.indent++
	code += g.ind() + fmt.Sprintf("d := float64(j) - %g\n", m)
	code += g.ind() + fmt.Sprintf("%s[j] = math.Exp(-(d*d)/(2*%g*%g))\n", wVar, s, s)
	code += g.ind() + fmt.Sprintf("%s += %s[j]\n", wSumVar, wVar)
	g.indent--
	code += g.ind() + "}\n"
	code += g.ind() + fmt.Sprintf("_%s_val := 0.0\n", varName)
	code += g.ind() + fmt.Sprintf("for j := 0; j < %d; j++ {\n", period)
	g.indent++
	code += g.ind() + fmt.Sprintf("_%s_val += %s[j] * %s\n", varName, wVar,
		comp.AccessGen.GenerateLoopValueAccess(fmt.Sprintf("%d-1-j", period)))
	g.indent--
	code += g.ind() + "}\n"
	code += g.ind() + fmt.Sprintf("%sSeries.Set(_%s_val / %s)\n", varName, varName, wSumVar)
	g.indent--
	code += g.ind() + "}\n"

	return comp.Preamble + code, nil
}
