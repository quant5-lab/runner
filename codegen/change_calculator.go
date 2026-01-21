package codegen

import "fmt"

/* ChangeCalculator generates bar-to-bar change calculation code.
 *
 * Critical: NO FUTURE PEEK - accesses previous bar via offset, never forward.
 * Warmup: Returns NaN for bar index 0 (no previous bar exists).
 *
 * Use: Required by composite indicators (RSI, momentum, rate of change, etc.)
 *      that need delta between consecutive bars.
 */
type ChangeCalculator struct {
	sourceAccessor AccessGenerator
	indenter       CodeIndenter
}

func NewChangeCalculator(sourceAccessor AccessGenerator) *ChangeCalculator {
	return &ChangeCalculator{
		sourceAccessor: sourceAccessor,
		indenter:       NewCodeIndenter(),
	}
}

/* GenerateChangeCode returns code that calculates change with warmup guard */
func (c *ChangeCalculator) GenerateChangeCode(changeVarName string) string {
	c.indenter.IncreaseIndent()

	code := c.indenter.Line(fmt.Sprintf("var %s float64", changeVarName))
	code += c.indenter.Line("if ctx.BarIndex < 1 {")
	c.indenter.IncreaseIndent()
	code += c.indenter.Line(fmt.Sprintf("%s = math.NaN()", changeVarName))
	c.indenter.DecreaseIndent()
	code += c.indenter.Line("} else {")
	c.indenter.IncreaseIndent()

	currentAccess := c.sourceAccessor.GenerateCurrentValueAccess()

	/* Use offset 1 for previous bar (baseOffset determines accessor behavior) */
	previousAccess := ""
	baseOffset := c.sourceAccessor.GetBaseOffset()
	if baseOffset == 0 {
		previousAccess = c.generateOffsetAccess(1)
	} else {
		previousAccess = c.sourceAccessor.GenerateCurrentValueAccess()
	}

	code += c.indenter.Line(fmt.Sprintf("%s = %s - %s", changeVarName, currentAccess, previousAccess))
	c.indenter.DecreaseIndent()
	code += c.indenter.Line("}")

	return code
}

/* generateOffsetAccess wraps offset as string for GenerateLoopValueAccess */
func (c *ChangeCalculator) generateOffsetAccess(offset int) string {
	return c.sourceAccessor.GenerateLoopValueAccess(fmt.Sprintf("%d", offset))
}
