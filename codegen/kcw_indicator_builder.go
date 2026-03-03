package codegen

import "fmt"

// KCWIndicatorBuilder generates Keltner Channel Width code:
// KCW = 2 * mult * ATR(length) / EMA(source, length)
//
// Requires two intermediate series:
//   - _varName_ema: EMA(source, length)
//   - _varName_atr: ATR(length)
type KCWIndicatorBuilder struct {
	resultVarName  string
	period         PeriodExpression
	sourceAccessor AccessGenerator
	multExpr       string
	context        StatefulIndicatorContext
	indenter       CodeIndenter
	internalSeries []string
}

func NewKCWIndicatorBuilder(
	resultVarName string,
	period PeriodExpression,
	sourceAccessor AccessGenerator,
	multExpr string,
	context StatefulIndicatorContext,
) *KCWIndicatorBuilder {
	return &KCWIndicatorBuilder{
		resultVarName:  resultVarName,
		period:         period,
		sourceAccessor: sourceAccessor,
		multExpr:       multExpr,
		context:        context,
		indenter:       NewCodeIndenter(),
		internalSeries: make([]string, 0),
	}
}

func (b *KCWIndicatorBuilder) trackedInternal(component string) string {
	name := fmt.Sprintf("_%s_%s", b.resultVarName, component)
	b.internalSeries = append(b.internalSeries, name)
	return name
}

func (b *KCWIndicatorBuilder) GetInternalSeriesNames() []string {
	return b.internalSeries
}

func (b *KCWIndicatorBuilder) Build() string {
	emaName := b.trackedInternal("ema")
	atrName := b.trackedInternal("atr")

	emaBuilder := NewStatefulIndicatorBuilder("ema", emaName, b.period, b.sourceAccessor, false, b.context)
	atrBuilder := NewStatefulIndicatorBuilder("rma", atrName, b.period, NewTrueRangeAccessGenerator(), false, b.context)

	code := emaBuilder.BuildEMA()
	code += atrBuilder.BuildRMA()
	code += b.generateFinalFormula(emaName, atrName)

	return code
}

func (b *KCWIndicatorBuilder) generateFinalFormula(emaName, atrName string) string {
	emaAccess := b.context.GenerateSeriesAccess(emaName, 0)
	atrAccess := b.context.GenerateSeriesAccess(atrName, 0)

	warmupExpr := ""
	if b.period.IsConstant() {
		warmupExpr = fmt.Sprintf("%d", b.period.AsInt()-1)
	} else {
		warmupExpr = fmt.Sprintf("%s-1", b.period.AsIntCast())
	}

	b.indenter.IncreaseIndent()
	code := b.indenter.Line(fmt.Sprintf("if ctx.BarIndex < %s {", warmupExpr))
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.resultVarName, "math.NaN()"))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("} else {")
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("_%s_ema := %s", b.resultVarName, emaAccess))
	code += b.indenter.Line(fmt.Sprintf("_%s_atr := %s", b.resultVarName, atrAccess))
	code += b.indenter.Line(fmt.Sprintf("if _%s_ema == 0 {", b.resultVarName))
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.resultVarName, "0.0"))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("} else {")
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.resultVarName,
		fmt.Sprintf("2.0 * %s * _%s_atr / _%s_ema", b.multExpr, b.resultVarName, b.resultVarName)))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")
	b.indenter.DecreaseIndent()

	return code
}
