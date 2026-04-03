package codegen

import "fmt"

/* DirectionalSplitGenerator separates signed change into positive/negative components.
 *
 * Algorithm: gain = max(change, 0), loss = max(-change, 0)
 * Sign convention: Both gains and losses stored as positive values
 * NaN handling: Converts NaN change to zero gain/loss (prevents propagation)
 *
 * Use: Required by indicators needing directional decomposition (RSI, DMI, MFI, etc.)
 */
type DirectionalSplitGenerator struct {
	gainsSeriesName  string
	lossesSeriesName string
	context          StatefulIndicatorContext
	indenter         CodeIndenter
	prefix           string
}

func NewDirectionalSplitGenerator(
	gainsSeriesName string,
	lossesSeriesName string,
	context StatefulIndicatorContext,
) *DirectionalSplitGenerator {
	/* Extract prefix from gains series name for unique variable naming */
	prefix := extractPrefix(gainsSeriesName)
	return &DirectionalSplitGenerator{
		gainsSeriesName:  gainsSeriesName,
		lossesSeriesName: lossesSeriesName,
		context:          context,
		indenter:         NewCodeIndenter(),
		prefix:           prefix,
	}
}

/* extractPrefix gets unique identifier from series name (_rsi_gains → _rsi_, _rsi14_gains → _rsi14_) */
func extractPrefix(seriesName string) string {
	/* Count underscores to determine naming strategy */
	/* Examples:
	   _rsi_gains (2 underscores) → _rsi_
	   _rsi14_gains (2 underscores) → _rsi14_
	   _gain (1 underscore) → "" (no prefix needed, use bare names)
	*/
	underscoreCount := 0

	for i := 0; i < len(seriesName); i++ {
		if seriesName[i] == '_' {
			underscoreCount++
			/* Found second underscore - extract prefix */
			if underscoreCount == 2 {
				return seriesName[:i+1]
			}
		}
	}

	/* Only 1 underscore (or 0) - no meaningful prefix, return empty */
	return ""
}

/* GenerateSplitCode generates gain/loss split with NaN handling */
func (g *DirectionalSplitGenerator) GenerateSplitCode(changeVarName string) string {
	g.indenter.IncreaseIndent()

	gainVar := g.prefix + "gain"
	lossVar := g.prefix + "loss"
	code := g.indenter.Line(fmt.Sprintf("var %s, %s float64", gainVar, lossVar))

	code += g.indenter.Line(fmt.Sprintf("if math.IsNaN(%s) {", changeVarName))
	g.indenter.IncreaseIndent()
	code += g.indenter.Line(fmt.Sprintf("%s = 0.0", gainVar))
	code += g.indenter.Line(fmt.Sprintf("%s = 0.0", lossVar))
	g.indenter.DecreaseIndent()

	code += g.indenter.Line(fmt.Sprintf("} else if %s > 0 {", changeVarName))
	g.indenter.IncreaseIndent()
	code += g.indenter.Line(fmt.Sprintf("%s = %s", gainVar, changeVarName))
	code += g.indenter.Line(fmt.Sprintf("%s = 0.0", lossVar))
	g.indenter.DecreaseIndent()

	/* Negate to store loss as positive (sign convention for smoothing) */
	code += g.indenter.Line("} else {")
	g.indenter.IncreaseIndent()
	code += g.indenter.Line(fmt.Sprintf("%s = 0.0", gainVar))
	code += g.indenter.Line(fmt.Sprintf("%s = -%s", lossVar, changeVarName))
	g.indenter.DecreaseIndent()
	code += g.indenter.Line("}")

	code += g.indenter.Line(g.context.GenerateSeriesUpdate(g.gainsSeriesName, gainVar))
	code += g.indenter.Line(g.context.GenerateSeriesUpdate(g.lossesSeriesName, lossVar))

	return code
}
