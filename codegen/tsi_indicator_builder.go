package codegen

import "fmt"

/* TSIIndicatorBuilder generates double-smoothed momentum/|momentum| × 100 calculation code. */
type TSIIndicatorBuilder struct {
	resultVarName  string
	shortPeriod    PeriodExpression
	longPeriod     PeriodExpression
	sourceAccessor AccessGenerator
	context        StatefulIndicatorContext
	indenter       CodeIndenter
	internalSeries []string
}

func NewTSIIndicatorBuilder(
	resultVarName string,
	shortPeriod PeriodExpression,
	longPeriod PeriodExpression,
	sourceAccessor AccessGenerator,
	context StatefulIndicatorContext,
) *TSIIndicatorBuilder {
	return &TSIIndicatorBuilder{
		resultVarName:  resultVarName,
		shortPeriod:    shortPeriod,
		longPeriod:     longPeriod,
		sourceAccessor: sourceAccessor,
		context:        context,
		indenter:       NewCodeIndenter(),
		internalSeries: make([]string, 0),
	}
}

func (b *TSIIndicatorBuilder) generateInternalName(component string) string {
	name := fmt.Sprintf("_%s_%s", b.resultVarName, component)
	b.internalSeries = append(b.internalSeries, name)
	return name
}

func (b *TSIIndicatorBuilder) GetInternalSeriesNames() []string {
	return b.internalSeries
}

func (b *TSIIndicatorBuilder) Build() string {
	momName := b.generateInternalName("mom")
	momAbsName := b.generateInternalName("mom_abs")
	ema1MomName := b.generateInternalName("ema1_mom")
	ema1AbsName := b.generateInternalName("ema1_abs")
	ema2MomName := b.generateInternalName("ema2_mom")
	ema2AbsName := b.generateInternalName("ema2_abs")

	code := b.generateMomentumStep(momName, momAbsName)
	code += b.generateFirstEMAStep(momName, momAbsName, ema1MomName, ema1AbsName)
	code += b.generateSecondEMAStep(ema1MomName, ema1AbsName, ema2MomName, ema2AbsName)
	code += b.generateFinalFormula(ema2MomName, ema2AbsName)

	return code
}

func (b *TSIIndicatorBuilder) generateMomentumStep(momName, momAbsName string) string {
	changeTempVar := fmt.Sprintf("_%s_change", b.resultVarName)
	changeCalc := NewChangeCalculator(b.sourceAccessor)
	code := changeCalc.GenerateChangeCode(changeTempVar)

	b.indenter.IncreaseIndent()
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(momName, changeTempVar))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(momAbsName, fmt.Sprintf("math.Abs(%s)", changeTempVar)))
	b.indenter.DecreaseIndent()

	return code
}

func (b *TSIIndicatorBuilder) generateFirstEMAStep(momName, momAbsName, ema1MomName, ema1AbsName string) string {
	momAccessor := NewInternalSeriesAccessor(momName, b.context)
	momAbsAccessor := NewInternalSeriesAccessor(momAbsName, b.context)

	ema1MomBuilder := NewStatefulIndicatorBuilder("ema", ema1MomName, b.longPeriod, momAccessor, false, b.context)
	ema1AbsBuilder := NewStatefulIndicatorBuilder("ema", ema1AbsName, b.longPeriod, momAbsAccessor, false, b.context)

	return ema1MomBuilder.BuildEMA() + ema1AbsBuilder.BuildEMA()
}

func (b *TSIIndicatorBuilder) generateSecondEMAStep(ema1MomName, ema1AbsName, ema2MomName, ema2AbsName string) string {
	ema1MomAccessor := NewInternalSeriesAccessor(ema1MomName, b.context)
	ema1AbsAccessor := NewInternalSeriesAccessor(ema1AbsName, b.context)

	ema2MomBuilder := NewStatefulIndicatorBuilder("ema", ema2MomName, b.shortPeriod, ema1MomAccessor, false, b.context)
	ema2AbsBuilder := NewStatefulIndicatorBuilder("ema", ema2AbsName, b.shortPeriod, ema1AbsAccessor, false, b.context)

	return ema2MomBuilder.BuildEMA() + ema2AbsBuilder.BuildEMA()
}

func (b *TSIIndicatorBuilder) generateFinalFormula(ema2MomName, ema2AbsName string) string {
	ema2MomAccess := b.context.GenerateSeriesAccess(ema2MomName, 0)
	ema2AbsAccess := b.context.GenerateSeriesAccess(ema2AbsName, 0)

	warmupExpr := ""
	if b.longPeriod.IsConstant() && b.shortPeriod.IsConstant() {
		warmupExpr = fmt.Sprintf("%d", b.longPeriod.AsInt()+b.shortPeriod.AsInt()-1)
	} else {
		warmupExpr = fmt.Sprintf("%s+%s-1", b.longPeriod.AsIntCast(), b.shortPeriod.AsIntCast())
	}

	b.indenter.IncreaseIndent()
	code := b.indenter.Line(fmt.Sprintf("if ctx.BarIndex < %s {", warmupExpr))
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.resultVarName, "math.NaN()"))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("} else {")
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("if %s == 0.0 {", ema2AbsAccess))
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.resultVarName, "0.0"))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("} else {")
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.resultVarName, fmt.Sprintf("100.0 * %s / %s", ema2MomAccess, ema2AbsAccess)))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")
	b.indenter.DecreaseIndent()

	return code
}
