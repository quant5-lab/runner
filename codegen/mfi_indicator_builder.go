package codegen

import "fmt"

/* MFIIndicatorBuilder generates Money Flow Index calculation code.
 *
 * Architecture: Composes generic primitives for MFI-specific formula
 * - ChangeCalculator: source[0] - source[1] (price direction)
 * - Volume-weighted directional money flow split
 * - Window sum of positive/negative money flows over period
 * - Final formula: MFI = 100 - 100/(1 + posSum/negSum)
 *
 * Internal series: positive_mf, negative_mf (volume-weighted directional flows) */
type MFIIndicatorBuilder struct {
	base *CompositeIndicatorBuilder
}

func NewMFIIndicatorBuilder(
	resultVarName string,
	period PeriodExpression,
	sourceAccessor AccessGenerator,
	context StatefulIndicatorContext,
) *MFIIndicatorBuilder {
	return &MFIIndicatorBuilder{
		base: NewCompositeIndicatorBuilder(resultVarName, period, sourceAccessor, context),
	}
}

func (b *MFIIndicatorBuilder) Build() string {
	posMFName := b.base.GenerateInternalSeriesName("positive_mf")
	negMFName := b.base.GenerateInternalSeriesName("negative_mf")

	code := b.generateChangeStep()
	code += b.generateMoneyFlowSplit(posMFName, negMFName)
	code += b.generateFinalFormula(posMFName, negMFName)

	return code
}

func (b *MFIIndicatorBuilder) generateChangeStep() string {
	changeCalc := NewChangeCalculator(b.base.sourceAccessor)
	return changeCalc.GenerateChangeCode(b.changeVarName())
}

func (b *MFIIndicatorBuilder) generateMoneyFlowSplit(posMFName, negMFName string) string {
	changeVar := b.changeVarName()
	sourceAccess := b.base.sourceAccessor.GenerateCurrentValueAccess()

	b.base.indenter.IncreaseIndent()

	code := b.base.indenter.Line(fmt.Sprintf("rawMF := %s * bar.Volume", sourceAccess))

	posMFVar := fmt.Sprintf("_%s_posMF", b.base.resultVarName)
	negMFVar := fmt.Sprintf("_%s_negMF", b.base.resultVarName)

	code += b.base.indenter.Line(fmt.Sprintf("var %s, %s float64", posMFVar, negMFVar))

	code += b.base.indenter.Line(fmt.Sprintf("if math.IsNaN(%s) {", changeVar))
	b.base.indenter.IncreaseIndent()
	code += b.base.indenter.Line(fmt.Sprintf("%s = 0.0", posMFVar))
	code += b.base.indenter.Line(fmt.Sprintf("%s = 0.0", negMFVar))
	b.base.indenter.DecreaseIndent()

	code += b.base.indenter.Line(fmt.Sprintf("} else if %s > 0 {", changeVar))
	b.base.indenter.IncreaseIndent()
	code += b.base.indenter.Line(fmt.Sprintf("%s = rawMF", posMFVar))
	code += b.base.indenter.Line(fmt.Sprintf("%s = 0.0", negMFVar))
	b.base.indenter.DecreaseIndent()

	code += b.base.indenter.Line(fmt.Sprintf("} else if %s < 0 {", changeVar))
	b.base.indenter.IncreaseIndent()
	code += b.base.indenter.Line(fmt.Sprintf("%s = 0.0", posMFVar))
	code += b.base.indenter.Line(fmt.Sprintf("%s = rawMF", negMFVar))
	b.base.indenter.DecreaseIndent()

	code += b.base.indenter.Line("} else {")
	b.base.indenter.IncreaseIndent()
	code += b.base.indenter.Line(fmt.Sprintf("%s = 0.0", posMFVar))
	code += b.base.indenter.Line(fmt.Sprintf("%s = 0.0", negMFVar))
	b.base.indenter.DecreaseIndent()
	code += b.base.indenter.Line("}")

	code += b.base.indenter.Line(b.base.context.GenerateSeriesUpdate(posMFName, posMFVar))
	code += b.base.indenter.Line(b.base.context.GenerateSeriesUpdate(negMFName, negMFVar))

	return code
}

func (b *MFIIndicatorBuilder) generateFinalFormula(posMFName, negMFName string) string {
	warmupExpr := ""
	if b.base.period.IsConstant() {
		warmupExpr = fmt.Sprintf("%d", b.base.period.AsInt())
	} else {
		warmupExpr = b.base.period.AsIntCast()
	}

	code := b.base.indenter.Line(fmt.Sprintf("if ctx.BarIndex < %s {", warmupExpr))
	b.base.indenter.IncreaseIndent()
	code += b.base.indenter.Line(b.base.context.GenerateSeriesUpdate(b.base.resultVarName, "math.NaN()"))
	b.base.indenter.DecreaseIndent()
	code += b.base.indenter.Line("} else {")
	b.base.indenter.IncreaseIndent()

	code += b.base.indenter.Line("posSum := 0.0")
	code += b.base.indenter.Line("negSum := 0.0")
	code += b.base.indenter.Line(fmt.Sprintf("for j := 0; j < %s; j++ {", warmupExpr))
	b.base.indenter.IncreaseIndent()

	posLoopAccess := b.base.context.GenerateSeriesDynamicAccess(posMFName, "j")
	negLoopAccess := b.base.context.GenerateSeriesDynamicAccess(negMFName, "j")

	code += b.base.indenter.Line(fmt.Sprintf("posSum += %s", posLoopAccess))
	code += b.base.indenter.Line(fmt.Sprintf("negSum += %s", negLoopAccess))
	b.base.indenter.DecreaseIndent()
	code += b.base.indenter.Line("}")

	code += b.base.indenter.Line("if negSum == 0 {")
	b.base.indenter.IncreaseIndent()
	code += b.base.indenter.Line(b.base.context.GenerateSeriesUpdate(b.base.resultVarName, "100.0"))
	b.base.indenter.DecreaseIndent()
	code += b.base.indenter.Line("} else {")
	b.base.indenter.IncreaseIndent()
	code += b.base.indenter.Line("mfr := posSum / negSum")
	code += b.base.indenter.Line("mfi := 100.0 - (100.0 / (1.0 + mfr))")
	code += b.base.indenter.Line(b.base.context.GenerateSeriesUpdate(b.base.resultVarName, "mfi"))
	b.base.indenter.DecreaseIndent()
	code += b.base.indenter.Line("}")

	b.base.indenter.DecreaseIndent()
	code += b.base.indenter.Line("}")

	return code
}

func (b *MFIIndicatorBuilder) changeVarName() string {
	return fmt.Sprintf("_%s_change", b.base.resultVarName)
}

func (b *MFIIndicatorBuilder) GetInternalSeriesNames() []string {
	return b.base.internalSeries
}
