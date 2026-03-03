package codegen

import "fmt"

// SARIndicatorBuilder generates Parabolic SAR code:
// ta.sar(start, inc, max) — iterative state machine tracking trend, EP, and AF.
//
// Requires 3 internal state series (beyond the result series):
//   - _varName_ep:    extreme point (highest high in uptrend / lowest low in downtrend)
//   - _varName_af:    acceleration factor
//   - _varName_trend: 1.0 = uptrend, -1.0 = downtrend
type SARIndicatorBuilder struct {
	resultVarName  string
	start          float64
	inc            float64
	maxAF          float64
	context        StatefulIndicatorContext
	indenter       CodeIndenter
	internalSeries []string
}

func NewSARIndicatorBuilder(
	resultVarName string,
	start, inc, maxAF float64,
	context StatefulIndicatorContext,
) *SARIndicatorBuilder {
	return &SARIndicatorBuilder{
		resultVarName:  resultVarName,
		start:          start,
		inc:            inc,
		maxAF:          maxAF,
		context:        context,
		indenter:       NewCodeIndenter(),
		internalSeries: make([]string, 0),
	}
}

func (b *SARIndicatorBuilder) trackedInternal(component string) string {
	name := fmt.Sprintf("_%s_%s", b.resultVarName, component)
	b.internalSeries = append(b.internalSeries, name)
	return name
}

func (b *SARIndicatorBuilder) GetInternalSeriesNames() []string {
	return b.internalSeries
}

func (b *SARIndicatorBuilder) Build() string {
	epName := b.trackedInternal("ep")
	afName := b.trackedInternal("af")
	trendName := b.trackedInternal("trend")

	return b.generateFirstBarInit(epName, afName, trendName)
}

func (b *SARIndicatorBuilder) generateFirstBarInit(epName, afName, trendName string) string {
	n := b.resultVarName

	b.indenter.IncreaseIndent()
	code := b.indenter.Line("if ctx.BarIndex == 0 {")
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.resultVarName, "math.NaN()"))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(epName, "math.NaN()"))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(afName, "math.NaN()"))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(trendName, "math.NaN()"))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("} else if ctx.BarIndex == 1 {")
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("_%s_isUp := ctx.Data[1].High >= ctx.Data[0].High", n))
	code += b.indenter.Line(fmt.Sprintf("_%s_trend := 1.0", n))
	code += b.indenter.Line(fmt.Sprintf("_%s_sarInit := ctx.Data[0].Low", n))
	code += b.indenter.Line(fmt.Sprintf("_%s_epInit := ctx.Data[0].High", n))
	code += b.indenter.Line(fmt.Sprintf("if !_%s_isUp {", n))
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("_%s_trend = -1.0", n))
	code += b.indenter.Line(fmt.Sprintf("_%s_sarInit = ctx.Data[0].High", n))
	code += b.indenter.Line(fmt.Sprintf("_%s_epInit = ctx.Data[0].Low", n))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.resultVarName, fmt.Sprintf("_%s_sarInit", n)))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(epName, fmt.Sprintf("_%s_epInit", n)))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(afName, fmt.Sprintf("%g", b.start)))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(trendName, fmt.Sprintf("_%s_trend", n)))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("} else {")
	b.indenter.IncreaseIndent()
	code += b.generateBarUpdate(epName, afName, trendName)
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")
	b.indenter.DecreaseIndent()

	return code
}

func (b *SARIndicatorBuilder) generateBarUpdate(epName, afName, trendName string) string {
	n := b.resultVarName
	prevSAR := b.context.GenerateSeriesAccess(b.resultVarName, 1)
	prevEP := b.context.GenerateSeriesAccess(epName, 1)
	prevAF := b.context.GenerateSeriesAccess(afName, 1)
	prevTrend := b.context.GenerateSeriesAccess(trendName, 1)

	code := b.indenter.Line(fmt.Sprintf("_%s_pSAR := %s", n, prevSAR))
	code += b.indenter.Line(fmt.Sprintf("_%s_pEP := %s", n, prevEP))
	code += b.indenter.Line(fmt.Sprintf("_%s_pAF := %s", n, prevAF))
	code += b.indenter.Line(fmt.Sprintf("_%s_isUp := %s > 0", n, prevTrend))
	code += b.indenter.Line(fmt.Sprintf("_%s_high := ctx.Data[ctx.BarIndex].High", n))
	code += b.indenter.Line(fmt.Sprintf("_%s_low := ctx.Data[ctx.BarIndex].Low", n))
	code += b.indenter.Line(fmt.Sprintf("_%s_proj := _%s_pSAR + _%s_pAF*(_%s_pEP-_%s_pSAR)", n, n, n, n, n))

	code += b.indenter.Line(fmt.Sprintf("if _%s_isUp {", n))
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("if ctx.BarIndex >= 2 && _%s_proj > ctx.Data[ctx.BarIndex-2].Low { _%s_proj = ctx.Data[ctx.BarIndex-2].Low }", n, n))
	code += b.indenter.Line(fmt.Sprintf("if _%s_proj > ctx.Data[ctx.BarIndex-1].Low { _%s_proj = ctx.Data[ctx.BarIndex-1].Low }", n, n))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("} else {")
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("if ctx.BarIndex >= 2 && _%s_proj < ctx.Data[ctx.BarIndex-2].High { _%s_proj = ctx.Data[ctx.BarIndex-2].High }", n, n))
	code += b.indenter.Line(fmt.Sprintf("if _%s_proj < ctx.Data[ctx.BarIndex-1].High { _%s_proj = ctx.Data[ctx.BarIndex-1].High }", n, n))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")

	code += b.indenter.Line(fmt.Sprintf("_%s_newSAR := _%s_proj", n, n))
	code += b.indenter.Line(fmt.Sprintf("_%s_newEP := _%s_pEP", n, n))
	code += b.indenter.Line(fmt.Sprintf("_%s_newAF := _%s_pAF", n, n))
	code += b.indenter.Line(fmt.Sprintf("_%s_newTrend := 1.0", n))
	code += b.indenter.Line(fmt.Sprintf("if !_%s_isUp { _%s_newTrend = -1.0 }", n, n))

	code += b.indenter.Line(fmt.Sprintf("if _%s_isUp {", n))
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("if _%s_low < _%s_proj {", n, n))
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("_%s_newTrend = -1.0", n))
	code += b.indenter.Line(fmt.Sprintf("_%s_newSAR = _%s_pEP", n, n))
	code += b.indenter.Line(fmt.Sprintf("_%s_newEP = _%s_low", n, n))
	code += b.indenter.Line(fmt.Sprintf("_%s_newAF = %g", n, b.start))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("} else {")
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("if _%s_high > _%s_pEP {", n, n))
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("_%s_newEP = _%s_high", n, n))
	code += b.indenter.Line(fmt.Sprintf("_%s_newAF = math.Min(_%s_pAF+%g, %g)", n, n, b.inc, b.maxAF))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("} else {")
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("if _%s_high > _%s_proj {", n, n))
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("_%s_newTrend = 1.0", n))
	code += b.indenter.Line(fmt.Sprintf("_%s_newSAR = _%s_pEP", n, n))
	code += b.indenter.Line(fmt.Sprintf("_%s_newEP = _%s_high", n, n))
	code += b.indenter.Line(fmt.Sprintf("_%s_newAF = %g", n, b.start))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("} else {")
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("if _%s_low < _%s_pEP {", n, n))
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("_%s_newEP = _%s_low", n, n))
	code += b.indenter.Line(fmt.Sprintf("_%s_newAF = math.Min(_%s_pAF+%g, %g)", n, n, b.inc, b.maxAF))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")

	code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.resultVarName, fmt.Sprintf("_%s_newSAR", n)))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(epName, fmt.Sprintf("_%s_newEP", n)))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(afName, fmt.Sprintf("_%s_newAF", n)))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(trendName, fmt.Sprintf("_%s_newTrend", n)))

	return code
}
