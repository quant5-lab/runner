package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type SwmaHandler struct{}

func (h *SwmaHandler) CanHandle(funcName string) bool {
	return funcName == "ta.swma" || funcName == "swma"
}

func (h *SwmaHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractSourceOnly(call, "ta.swma")
	if err != nil {
		return "", err
	}

	warmup := 3 + comp.AccessGen.GetBaseOffset()
	a := comp.AccessGen.GenerateLoopValueAccess("3")
	b := comp.AccessGen.GenerateLoopValueAccess("2")
	c := comp.AccessGen.GenerateLoopValueAccess("1")
	d := comp.AccessGen.GenerateLoopValueAccess("0")

	code := g.ind() + fmt.Sprintf("if ctx.BarIndex < %d {\n", warmup)
	g.indent++
	code += g.ind() + varName + "Series.Set(math.NaN())\n"
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + varName + "Series.Set(" + a + "*(1.0/6.0) + " + b + "*(2.0/6.0) + " + c + "*(2.0/6.0) + " + d + "*(1.0/6.0))\n"
	g.indent--
	code += g.ind() + "}\n"
	return comp.Preamble + code, nil
}
