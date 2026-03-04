package codegen

import "fmt"

// SupertrendIndicatorBuilder: [supertrend, direction] = hl2 ± factor * ATR(atrPeriod)
// with carry-forward band clamping and directional flip detection.
//
// direction = 1  → uptrend  (price above supertrend = lower band = support)
// direction = -1 → downtrend (price below supertrend = upper band = resistance)
//
// Band clamping (PineScript convention):
//
//	clampedLower = rawLower > prevLower || prevClose < prevLower ? rawLower : prevLower
//	clampedUpper = rawUpper < prevUpper || prevClose > prevUpper ? rawUpper : prevUpper
//
// Requires 4 intermediate series (named after the first output variable):
//   - _st_atr:   RMA of True Range
//   - _st_upper: clamped upper band (resistance in downtrend)
//   - _st_lower: clamped lower band (support in uptrend)
//   - _st_dir:   direction carry-forward (1.0 or -1.0)
type SupertrendIndicatorBuilder struct {
	outputVarNames [2]string // [supertrend, direction]
	atrPeriod      int
	factor         string // float expression
	context        StatefulIndicatorContext
	indenter       CodeIndenter
}

func NewSupertrendIndicatorBuilder(
	outputVarNames [2]string,
	atrPeriod int,
	factor string,
	context StatefulIndicatorContext,
) *SupertrendIndicatorBuilder {
	return &SupertrendIndicatorBuilder{
		outputVarNames: outputVarNames,
		atrPeriod:      atrPeriod,
		factor:         factor,
		context:        context,
		indenter:       NewCodeIndenter(),
	}
}

func (b *SupertrendIndicatorBuilder) InternalSeriesNames() []string {
	prefix := b.outputVarNames[0]
	return []string{
		fmt.Sprintf("_%s_atr", prefix),
		fmt.Sprintf("_%s_upper", prefix),
		fmt.Sprintf("_%s_lower", prefix),
		fmt.Sprintf("_%s_dir", prefix),
	}
}

func (b *SupertrendIndicatorBuilder) Build() string {
	prefix := b.outputVarNames[0]
	atrName := fmt.Sprintf("_%s_atr", prefix)
	upperName := fmt.Sprintf("_%s_upper", prefix)
	lowerName := fmt.Sprintf("_%s_lower", prefix)
	dirName := fmt.Sprintf("_%s_dir", prefix)
	stVar := b.outputVarNames[0]
	dirVar := b.outputVarNames[1]

	atrBuilder := NewStatefulIndicatorBuilder("rma", atrName, NewConstantPeriod(b.atrPeriod), NewTrueRangeAccessGenerator(), false, b.context)
	code := atrBuilder.BuildRMA()
	code += b.buildBandsAndDirection(atrName, upperName, lowerName, dirName, stVar, dirVar)
	return code
}

func (b *SupertrendIndicatorBuilder) buildBandsAndDirection(
	atrName, upperName, lowerName, dirName, stVar, dirVar string,
) string {
	warmup := b.atrPeriod - 1

	b.indenter.IncreaseIndent()
	code := b.indenter.Line(fmt.Sprintf("if ctx.BarIndex < %d {", warmup))
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(upperName, "math.NaN()"))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(lowerName, "math.NaN()"))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(dirName, "math.NaN()"))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(stVar, "math.NaN()"))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(dirVar, "math.NaN()"))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("} else {")
	b.indenter.IncreaseIndent()
	code += b.buildActiveBarCode(atrName, upperName, lowerName, dirName, stVar, dirVar)
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")
	b.indenter.DecreaseIndent()

	return code
}

func (b *SupertrendIndicatorBuilder) buildActiveBarCode(
	atrName, upperName, lowerName, dirName, stVar, dirVar string,
) string {
	atrAccess := b.context.GenerateSeriesAccess(atrName, 0)
	prevUpperAccess := b.context.GenerateSeriesAccess(upperName, 1)
	prevLowerAccess := b.context.GenerateSeriesAccess(lowerName, 1)
	prevDirAccess := b.context.GenerateSeriesAccess(dirName, 1)

	ind := b.indenter

	code := ind.Line(fmt.Sprintf("_st_atr := %s", atrAccess))
	code += ind.Line("_st_hl2 := (ctx.Data[ctx.BarIndex].High + ctx.Data[ctx.BarIndex].Low) / 2.0")
	code += ind.Line(fmt.Sprintf("_st_rawUpper := _st_hl2 + %s*_st_atr", b.factor))
	code += ind.Line(fmt.Sprintf("_st_rawLower := _st_hl2 - %s*_st_atr", b.factor))
	code += ind.Line("")

	code += ind.Line(fmt.Sprintf("_st_prevUpper := %s", prevUpperAccess))
	code += ind.Line(fmt.Sprintf("_st_prevLower := %s", prevLowerAccess))
	code += ind.Line("_st_prevClose := closeSeries.Get(1)")
	code += ind.Line("")

	code += ind.Line("_st_lower := _st_rawLower")
	code += ind.Line("if !math.IsNaN(_st_prevLower) && _st_rawLower <= _st_prevLower && _st_prevClose >= _st_prevLower {")
	ind.IncreaseIndent()
	code += ind.Line("_st_lower = _st_prevLower")
	ind.DecreaseIndent()
	code += ind.Line("}")
	code += ind.Line("")

	code += ind.Line("_st_upper := _st_rawUpper")
	code += ind.Line("if !math.IsNaN(_st_prevUpper) && _st_rawUpper >= _st_prevUpper && _st_prevClose <= _st_prevUpper {")
	ind.IncreaseIndent()
	code += ind.Line("_st_upper = _st_prevUpper")
	ind.DecreaseIndent()
	code += ind.Line("}")
	code += ind.Line("")

	code += ind.Line(fmt.Sprintf("_st_prevDir := %s", prevDirAccess))
	code += ind.Line("_st_dir := 1.0")
	code += ind.Line("if !math.IsNaN(_st_prevDir) {")
	ind.IncreaseIndent()
	code += ind.Line("_st_dir = _st_prevDir")
	code += ind.Line("_st_close := ctx.Data[ctx.BarIndex].Close")
	code += ind.Line("if _st_prevDir == -1.0 && _st_close > _st_upper { _st_dir = 1.0 }")
	code += ind.Line("if _st_prevDir == 1.0 && _st_close < _st_lower { _st_dir = -1.0 }")
	ind.DecreaseIndent()
	code += ind.Line("}")
	code += ind.Line("")

	code += ind.Line("_st_value := _st_lower")
	code += ind.Line("if _st_dir == -1.0 { _st_value = _st_upper }")
	code += ind.Line("")

	code += ind.Line(b.context.GenerateSeriesUpdate(upperName, "_st_upper"))
	code += ind.Line(b.context.GenerateSeriesUpdate(lowerName, "_st_lower"))
	code += ind.Line(b.context.GenerateSeriesUpdate(dirName, "_st_dir"))
	code += ind.Line(b.context.GenerateSeriesUpdate(stVar, "_st_value"))
	code += ind.Line(b.context.GenerateSeriesUpdate(dirVar, "_st_dir"))

	return code
}
