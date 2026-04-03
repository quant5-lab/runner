package codegen

import "fmt"

// KCIndicatorBuilder: [upper, basis, lower] = EMA ± mult * range(length)
// where range = ATR(length) when useTrueRange=true, or SMA(high-low, length) otherwise.
//
// Requires two intermediate series (named after the first output variable):
//   - _upper_ema: EMA(source, length)
//   - _upper_atr: ATR(length) or SMA(high-low, length)
type KCIndicatorBuilder struct {
	outputVarNames [3]string // [upper, basis, lower]
	period         PeriodExpression
	sourceAccessor AccessGenerator
	multExpr       string
	useTrueRange   bool
	context        StatefulIndicatorContext
	indenter       CodeIndenter
}

func NewKCIndicatorBuilder(
	outputVarNames [3]string,
	period PeriodExpression,
	sourceAccessor AccessGenerator,
	multExpr string,
	useTrueRange bool,
	context StatefulIndicatorContext,
) *KCIndicatorBuilder {
	return &KCIndicatorBuilder{
		outputVarNames: outputVarNames,
		period:         period,
		sourceAccessor: sourceAccessor,
		multExpr:       multExpr,
		useTrueRange:   useTrueRange,
		context:        context,
		indenter:       NewCodeIndenter(),
	}
}

func (b *KCIndicatorBuilder) InternalSeriesNames() []string {
	prefix := b.outputVarNames[0]
	return []string{
		fmt.Sprintf("_%s_ema", prefix),
		fmt.Sprintf("_%s_atr", prefix),
	}
}

func (b *KCIndicatorBuilder) Build() string {
	prefix := b.outputVarNames[0]
	emaName := fmt.Sprintf("_%s_ema", prefix)
	atrName := fmt.Sprintf("_%s_atr", prefix)

	emaBuilder := NewStatefulIndicatorBuilder("ema", emaName, b.period, b.sourceAccessor, false, b.context)
	code := emaBuilder.BuildEMA()

	if b.useTrueRange {
		atrBuilder := NewStatefulIndicatorBuilder("rma", atrName, b.period, NewTrueRangeAccessGenerator(), false, b.context)
		code += atrBuilder.BuildRMA()
	} else {
		code += buildHighLowSMA(atrName, b.period, b.context)
	}

	code += b.buildOutputBands(emaName, atrName)
	return code
}

func (b *KCIndicatorBuilder) buildOutputBands(emaName, atrName string) string {
	upperVar := b.outputVarNames[0]
	basisVar := b.outputVarNames[1]
	lowerVar := b.outputVarNames[2]

	emaAccess := b.context.GenerateSeriesAccess(emaName, 0)
	atrAccess := b.context.GenerateSeriesAccess(atrName, 0)

	b.indenter.IncreaseIndent()
	code := b.indenter.Line(fmt.Sprintf("if ctx.BarIndex < %s {", warmupBarExpr(b.period)))
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(upperVar, "math.NaN()"))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(basisVar, "math.NaN()"))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(lowerVar, "math.NaN()"))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("} else {")
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("_kc_ema := %s", emaAccess))
	code += b.indenter.Line(fmt.Sprintf("_kc_atr := %s", atrAccess))
	code += b.indenter.Line(fmt.Sprintf("_kc_band := %s * _kc_atr", b.multExpr))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(upperVar, "_kc_ema + _kc_band"))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(basisVar, "_kc_ema"))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(lowerVar, "_kc_ema - _kc_band"))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")
	b.indenter.DecreaseIndent()

	return code
}
