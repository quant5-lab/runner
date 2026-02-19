package codegen

import "fmt"

/* generateMonotoneSequence emits code checking that source is strictly monotone over length bars.
 * comparisonOp is "<=" for rising (violation) or ">=" for falling (violation). */
func generateMonotoneSequence(g *generator, varName string, accessor AccessGenerator, period int, violationOp string) string {
	baseOffset := accessor.GetBaseOffset()
	warmup := period + baseOffset

	code := g.ind() + fmt.Sprintf("if ctx.BarIndex < %d {\n", warmup)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(func() float64 {\n", varName)
	g.indent++
	code += g.ind() + fmt.Sprintf("for j := 0; j < %d; j++ {\n", period)
	g.indent++
	code += g.ind() + fmt.Sprintf("curr := %s\n", accessor.GenerateLoopValueAccess("j"))
	code += g.ind() + fmt.Sprintf("prev := %s\n", accessor.GenerateLoopValueAccess("j+1"))
	code += g.ind() + fmt.Sprintf("if curr %s prev { return 0.0 }\n", violationOp)
	g.indent--
	code += g.ind() + "}\n"
	code += g.ind() + "return 1.0\n"
	g.indent--
	code += g.ind() + "}())\n"
	g.indent--
	code += g.ind() + "}\n"
	return code
}
