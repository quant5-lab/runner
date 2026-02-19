package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* CmoHandler generates Chande Momentum Oscillator:
 * (sumUpMoves - sumDownMoves) / (sumUpMoves + sumDownMoves) * 100 over length bars */
type CmoHandler struct{}

func (h *CmoHandler) CanHandle(funcName string) bool {
	return funcName == "ta.cmo" || funcName == "cmo"
}

func (h *CmoHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.Extract(call, "ta.cmo")
	if err != nil {
		return "", err
	}

	baseOffset := comp.AccessGen.GetBaseOffset()
	warmup := comp.Period + baseOffset

	upVar := fmt.Sprintf("_%s_up", varName)
	downVar := fmt.Sprintf("_%s_down", varName)
	totalVar := fmt.Sprintf("_%s_total", varName)
	currVar := fmt.Sprintf("_%s_curr", varName)
	prevVar := fmt.Sprintf("_%s_prev", varName)

	code := g.ind() + fmt.Sprintf("if ctx.BarIndex < %d {\n", warmup)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%s, %s := 0.0, 0.0\n", upVar, downVar)
	code += g.ind() + fmt.Sprintf("for j := 0; j < %d; j++ {\n", comp.Period)
	g.indent++
	code += g.ind() + fmt.Sprintf("%s := %s\n", currVar, comp.AccessGen.GenerateLoopValueAccess("j"))
	code += g.ind() + fmt.Sprintf("%s := %s\n", prevVar, comp.AccessGen.GenerateLoopValueAccess("j+1"))
	code += g.ind() + fmt.Sprintf("if !math.IsNaN(%s) && !math.IsNaN(%s) {\n", currVar, prevVar)
	g.indent++
	code += g.ind() + fmt.Sprintf("if %s > %s { %s += %s - %s } else { %s += %s - %s }\n",
		currVar, prevVar, upVar, currVar, prevVar, downVar, prevVar, currVar)
	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"
	code += g.ind() + fmt.Sprintf("%s := %s + %s\n", totalVar, upVar, downVar)
	code += g.ind() + fmt.Sprintf("if %s == 0.0 {\n", totalVar)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(0.0)\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set((%s - %s) / %s * 100.0)\n", varName, upVar, downVar, totalVar)
	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"

	return comp.Preamble + code, nil
}
