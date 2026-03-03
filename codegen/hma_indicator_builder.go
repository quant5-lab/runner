package codegen

import (
	"fmt"
	"math"
)

// HMAIndicatorBuilder generates Hull Moving Average code:
// HMA(source, n) = WMA(2*WMA(source, n/2) - WMA(source, n), round(sqrt(n)))
//
// Three intermediate series:
//   - _varName_wma1: WMA(source, halfPeriod)
//   - _varName_wma2: WMA(source, period)
//   - _varName_diff: 2*wma1 - wma2
type HMAIndicatorBuilder struct {
	resultVarName  string
	period         PeriodExpression
	sourceAccessor AccessGenerator
	context        StatefulIndicatorContext
	indenter       CodeIndenter
	internalSeries []string
}

func NewHMAIndicatorBuilder(
	resultVarName string,
	period PeriodExpression,
	sourceAccessor AccessGenerator,
	context StatefulIndicatorContext,
) *HMAIndicatorBuilder {
	return &HMAIndicatorBuilder{
		resultVarName:  resultVarName,
		period:         period,
		sourceAccessor: sourceAccessor,
		context:        context,
		indenter:       NewCodeIndenter(),
		internalSeries: make([]string, 0),
	}
}

func (b *HMAIndicatorBuilder) trackedInternal(component string) string {
	name := fmt.Sprintf("_%s_%s", b.resultVarName, component)
	b.internalSeries = append(b.internalSeries, name)
	return name
}

func (b *HMAIndicatorBuilder) GetInternalSeriesNames() []string {
	return b.internalSeries
}

func (b *HMAIndicatorBuilder) Build() string {
	periodInt := b.period.AsInt()
	halfPeriod := NewConstantPeriod(periodInt / 2)
	sqrtPeriod := NewConstantPeriod(int(math.Round(math.Sqrt(float64(periodInt)))))

	wma1Name := b.trackedInternal("wma1")
	wma2Name := b.trackedInternal("wma2")
	diffName := b.trackedInternal("diff")

	code := b.generateWMAStep(wma1Name, b.sourceAccessor, halfPeriod)
	code += b.generateWMAStep(wma2Name, b.sourceAccessor, b.period)
	code += b.generateDiffStep(diffName, wma1Name, wma2Name, b.period, halfPeriod)
	code += b.generateFinalWMAStep(diffName, sqrtPeriod, b.period)

	return code
}

func (b *HMAIndicatorBuilder) generateWMAStep(seriesName string, accessor AccessGenerator, period PeriodExpression) string {
	periodInt := period.AsInt()
	warmup := periodInt + accessor.GetBaseOffset()

	b.indenter.IncreaseIndent()
	code := b.indenter.Line(fmt.Sprintf("if ctx.BarIndex < %d {", warmup))
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(seriesName, "math.NaN()"))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("} else {")
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("_%s_sum, _%s_wsum := 0.0, 0.0", seriesName, seriesName))
	code += b.indenter.Line(fmt.Sprintf("for j := 0; j < %d; j++ {", periodInt))
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("_%s_w := float64(%d - j)", seriesName, periodInt))
	code += b.indenter.Line(fmt.Sprintf("_%s_sum += _%s_w * %s", seriesName, seriesName, accessor.GenerateLoopValueAccess("j")))
	code += b.indenter.Line(fmt.Sprintf("_%s_wsum += _%s_w", seriesName, seriesName))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(seriesName, fmt.Sprintf("_%s_sum / _%s_wsum", seriesName, seriesName)))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")
	b.indenter.DecreaseIndent()

	return code
}

func (b *HMAIndicatorBuilder) generateDiffStep(diffName, wma1Name, wma2Name string, period, halfPeriod PeriodExpression) string {
	wma1Val := b.context.GenerateSeriesAccess(wma1Name, 0)
	wma2Val := b.context.GenerateSeriesAccess(wma2Name, 0)

	b.indenter.IncreaseIndent()
	code := b.indenter.Line(fmt.Sprintf("_%s_w1 := %s", diffName, wma1Val))
	code += b.indenter.Line(fmt.Sprintf("_%s_w2 := %s", diffName, wma2Val))
	code += b.indenter.Line(fmt.Sprintf("if math.IsNaN(_%s_w1) || math.IsNaN(_%s_w2) {", diffName, diffName))
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(diffName, "math.NaN()"))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("} else {")
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(diffName, fmt.Sprintf("2*_%s_w1 - _%s_w2", diffName, diffName)))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")
	b.indenter.DecreaseIndent()

	return code
}

func (b *HMAIndicatorBuilder) generateFinalWMAStep(diffName string, sqrtPeriod, fullPeriod PeriodExpression) string {
	sqrtInt := sqrtPeriod.AsInt()
	fullInt := fullPeriod.AsInt()
	totalWarmup := fullInt + sqrtInt - 1

	diffAccessor := NewInternalSeriesAccessor(diffName, b.context)

	b.indenter.IncreaseIndent()
	code := b.indenter.Line(fmt.Sprintf("if ctx.BarIndex < %d {", totalWarmup))
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.resultVarName, "math.NaN()"))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("} else {")
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("_%s_fsum, _%s_fwsum := 0.0, 0.0", b.resultVarName, b.resultVarName))
	code += b.indenter.Line(fmt.Sprintf("for j := 0; j < %d; j++ {", sqrtInt))
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("_%s_fw := float64(%d - j)", b.resultVarName, sqrtInt))
	code += b.indenter.Line(fmt.Sprintf("_%s_fsum += _%s_fw * %s", b.resultVarName, b.resultVarName, diffAccessor.GenerateLoopValueAccess("j")))
	code += b.indenter.Line(fmt.Sprintf("_%s_fwsum += _%s_fw", b.resultVarName, b.resultVarName))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.resultVarName, fmt.Sprintf("_%s_fsum / _%s_fwsum", b.resultVarName, b.resultVarName)))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")
	b.indenter.DecreaseIndent()

	return code
}
