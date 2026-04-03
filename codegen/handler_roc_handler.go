package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* RocHandler generates ta.roc(source, length) = (source - source[length]) / source[length] * 100 */
type RocHandler struct{}

func (h *RocHandler) CanHandle(funcName string) bool {
	return funcName == "ta.roc" || funcName == "roc"
}

func (h *RocHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.Extract(call, "ta.roc")
	if err != nil {
		return "", err
	}

	baseOffset := comp.AccessGen.GetBaseOffset()
	warmup := comp.Period + baseOffset

	current := comp.AccessGen.GenerateLoopValueAccess("0")
	past := comp.AccessGen.GenerateLoopValueAccess(fmt.Sprintf("%d", comp.Period))
	pastVar := fmt.Sprintf("_%s_past", varName)

	code := g.ind() + fmt.Sprintf("if ctx.BarIndex < %d {\n", warmup)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%s := %s\n", pastVar, past)
	code += g.ind() + fmt.Sprintf("if math.IsNaN(%s) || %s == 0.0 {\n", pastVar, pastVar)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set((%s - %s) / %s * 100.0)\n", varName, current, pastVar, pastVar)
	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"

	return comp.Preamble + code, nil
}
