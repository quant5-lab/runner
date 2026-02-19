package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* MomHandler generates ta.mom(source, length) = source - source[length] */
type MomHandler struct{}

func (h *MomHandler) CanHandle(funcName string) bool {
	return funcName == "ta.mom" || funcName == "mom"
}

func (h *MomHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.Extract(call, "ta.mom")
	if err != nil {
		return "", err
	}

	baseOffset := comp.AccessGen.GetBaseOffset()
	warmup := comp.Period + baseOffset

	current := comp.AccessGen.GenerateLoopValueAccess("0")
	past := comp.AccessGen.GenerateLoopValueAccess(fmt.Sprintf("%d", comp.Period))

	code := g.ind() + fmt.Sprintf("if ctx.BarIndex < %d {\n", warmup)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(%s - %s)\n", varName, current, past)
	g.indent--
	code += g.ind() + "}\n"

	return comp.Preamble + code, nil
}
