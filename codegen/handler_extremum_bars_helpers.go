package codegen

import "fmt"

/* generateBarsToExtremum emits code finding the negative offset to the bar holding the window extremum.
 * compareOp is ">" for highestbars or "<" for lowestbars. */
func generateBarsToExtremum(g *generator, varName string, accessor AccessGenerator, period int, compareOp string) string {
	baseOffset := accessor.GetBaseOffset()
	warmup := period - 1 + baseOffset

	code := g.ind() + fmt.Sprintf("if ctx.BarIndex < %d {\n", warmup)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	extremumIdx := fmt.Sprintf("_%s_extIdx", varName)
	extremumVal := fmt.Sprintf("_%s_extVal", varName)
	code += g.ind() + fmt.Sprintf("%s := 0\n", extremumIdx)
	code += g.ind() + fmt.Sprintf("%s := %s\n", extremumVal, accessor.GenerateLoopValueAccess("0"))
	code += g.ind() + fmt.Sprintf("for j := 1; j < %d; j++ {\n", period)
	g.indent++
	loopVal := fmt.Sprintf("_%s_v", varName)
	code += g.ind() + fmt.Sprintf("%s := %s\n", loopVal, accessor.GenerateLoopValueAccess("j"))
	code += g.ind() + fmt.Sprintf("if !math.IsNaN(%s) && %s %s %s {\n", loopVal, loopVal, compareOp, extremumVal)
	g.indent++
	code += g.ind() + fmt.Sprintf("%s = %s\n", extremumVal, loopVal)
	code += g.ind() + fmt.Sprintf("%s = j\n", extremumIdx)
	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"
	code += g.ind() + fmt.Sprintf("%sSeries.Set(float64(-%s))\n", varName, extremumIdx)
	g.indent--
	code += g.ind() + "}\n"
	return code
}
