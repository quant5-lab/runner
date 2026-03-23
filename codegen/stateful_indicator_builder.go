package codegen

import (
	"fmt"
)

// StatefulIndicatorBuilder generates code for TA indicators that maintain state
// across bars by referencing their own previous values (RMA, EMA, etc.)
// Unlike window-based indicators, these use recursive formulas with previous results
type StatefulIndicatorBuilder struct {
	indicatorName          string
	varName                string
	period                 PeriodExpression
	accessor               AccessGenerator
	needsNaN               bool
	recoverFromNaNPrevious bool
	indenter               CodeIndenter
	context                StatefulIndicatorContext
}

// NewStatefulIndicatorBuilder creates builder for stateful indicators.
//
// recoverFromNaNPrevious is derived from indicatorName: EMA restarts from the
// current source when the previous indicator value is NaN (e.g. after a gap);
// RMA propagates NaN because its recursive formula alpha*src+(1-alpha)*prev
// has no defined recovery — if prev is NaN the result must also be NaN.
func NewStatefulIndicatorBuilder(
	indicatorName string,
	varName string,
	period PeriodExpression,
	accessor AccessGenerator,
	needsNaN bool,
	context StatefulIndicatorContext,
) *StatefulIndicatorBuilder {
	return &StatefulIndicatorBuilder{
		indicatorName:          indicatorName,
		varName:                varName,
		period:                 period,
		accessor:               accessor,
		needsNaN:               needsNaN,
		recoverFromNaNPrevious: isEMAVariant(indicatorName),
		indenter:               NewCodeIndenter(),
		context:                context,
	}
}

// isEMAVariant returns true for EMA-family indicators that restart from the current
// source value when the previous indicator value is NaN.
func isEMAVariant(indicatorName string) bool {
	return indicatorName == "ta.ema" || indicatorName == "ema"
}

// RMA formula: rma[i] = alpha * source[i] + (1-alpha) * rma[i-1], alpha = 1/period
func (b *StatefulIndicatorBuilder) BuildRMA() string {
	b.indenter.IncreaseIndent()

	code := b.buildHeader("RMA")
	code += b.buildWarmupPeriod()

	b.indenter.IncreaseIndent()
	code += b.buildInitializationPhase()
	code += b.buildRecursivePhase(b.rmaFormula)
	b.indenter.DecreaseIndent()

	code += b.closeBlock()
	return code
}

// EMA formula: ema[i] = alpha * source[i] + (1-alpha) * ema[i-1], alpha = 2/(period+1)
func (b *StatefulIndicatorBuilder) BuildEMA() string {
	b.indenter.IncreaseIndent()

	code := b.buildHeader("EMA")
	code += b.buildWarmupPeriod()

	b.indenter.IncreaseIndent()
	code += b.buildInitializationPhase()
	code += b.buildRecursivePhase(b.emaFormula)
	b.indenter.DecreaseIndent()

	code += b.closeBlock()
	return code
}

func (b *StatefulIndicatorBuilder) buildHeader(indicatorType string) string {
	warmupBarsExpr := ""
	if b.period.IsConstant() {
		warmupBarsExpr = fmt.Sprintf("%d", b.period.AsInt()-1)
	} else {
		warmupBarsExpr = fmt.Sprintf("%s-1", b.period.AsIntCast())
	}

	return b.indenter.Line(fmt.Sprintf("/* Inline %s(%s) - Stateful recursive calculation */", indicatorType, b.period.AsGoExpr())) +
		b.indenter.Line(fmt.Sprintf("if ctx.BarIndex < %s {", warmupBarsExpr))
}

func (b *StatefulIndicatorBuilder) buildWarmupPeriod() string {
	b.indenter.IncreaseIndent()
	code := b.indenter.Line(b.context.GenerateSeriesUpdate(b.varName, "math.NaN()"))
	b.indenter.DecreaseIndent()
	return code + b.indenter.Line("} else {")
}

func (b *StatefulIndicatorBuilder) buildInitializationPhase() string {
	initBarExpr := ""
	if b.period.IsConstant() {
		initBarExpr = fmt.Sprintf("%d", b.period.AsInt()-1)
	} else {
		initBarExpr = fmt.Sprintf("%s-1", b.period.AsIntCast())
	}

	code := b.indenter.Line(fmt.Sprintf("if ctx.BarIndex == %s {", initBarExpr))
	b.indenter.IncreaseIndent()

	code += b.indenter.Line("/* First valid value: calculate SMA as initial state */")
	code += b.indenter.Line("_sma_accumulator := 0.0")

	loopBound := ""
	if b.period.IsConstant() {
		loopBound = fmt.Sprintf("%d", b.period.AsInt())
	} else {
		loopBound = b.period.AsIntCast()
	}
	if b.needsNaN {
		code += b.indenter.Line("_sma_has_nan := false")
	}
	code += b.indenter.Line(fmt.Sprintf("for j := 0; j < %s; j++ {", loopBound))
	b.indenter.IncreaseIndent()

	valueAccess := b.accessor.GenerateLoopValueAccess("j")
	if b.needsNaN {
		code += b.indenter.Line(fmt.Sprintf("val := %s", valueAccess))
		code += b.indenter.Line("if math.IsNaN(val) {")
		b.indenter.IncreaseIndent()
		code += b.indenter.Line("_sma_has_nan = true")
		code += b.indenter.Line("break")
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("}")
		code += b.indenter.Line("_sma_accumulator += val")
	} else {
		code += b.indenter.Line(fmt.Sprintf("_sma_accumulator += %s", valueAccess))
	}

	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")

	smaDiv := ""
	if b.period.IsConstant() {
		smaDiv = fmt.Sprintf("float64(%d)", b.period.AsInt())
	} else {
		smaDiv = b.period.AsFloat64Cast()
	}
	if b.needsNaN {
		code += b.indenter.Line("if _sma_has_nan {")
		b.indenter.IncreaseIndent()
		code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.varName, "math.NaN()"))
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("} else {")
		b.indenter.IncreaseIndent()
		code += b.indenter.Line(fmt.Sprintf("initialValue := _sma_accumulator / %s", smaDiv))
		code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.varName, "initialValue"))
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("}")
	} else {
		code += b.indenter.Line(fmt.Sprintf("initialValue := _sma_accumulator / %s", smaDiv))
		code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.varName, "initialValue"))
	}

	b.indenter.DecreaseIndent()
	code += b.indenter.Line("} else {")

	return code
}

type recursiveFormula func() string

func (b *StatefulIndicatorBuilder) buildRecursivePhase(formula recursiveFormula) string {
	b.indenter.IncreaseIndent()

	code := b.indenter.Line("/* Recursive phase: use previous indicator value */")
	code += b.indenter.Line(fmt.Sprintf("previousValue := %s", b.context.GenerateSeriesAccess(b.varName, 1)))

	currentSourceAccess := b.accessor.GenerateLoopValueAccess("0")
	code += b.indenter.Line(fmt.Sprintf("currentSource := %s", currentSourceAccess))

	if b.needsNaN {
		code += b.indenter.Line("if math.IsNaN(currentSource) {")
		b.indenter.IncreaseIndent()
		code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.varName, "math.NaN()"))
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("} else if math.IsNaN(previousValue) {")
		b.indenter.IncreaseIndent()
		code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.varName, "currentSource"))
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("} else {")
		b.indenter.IncreaseIndent()
	}

	code += formula()

	if b.needsNaN {
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("}")
	}

	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")

	return code
}

func (b *StatefulIndicatorBuilder) rmaFormula() string {
	alphaExpr := ""
	if b.period.IsConstant() {
		alphaExpr = fmt.Sprintf("1.0 / float64(%d)", b.period.AsInt())
	} else {
		alphaExpr = fmt.Sprintf("1.0 / %s", b.period.AsFloat64Cast())
	}

	code := b.indenter.Line(fmt.Sprintf("alpha := %s", alphaExpr))
	code += b.indenter.Line("newValue := alpha*currentSource + (1-alpha)*previousValue")
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.varName, "newValue"))
	return code
}

func (b *StatefulIndicatorBuilder) emaFormula() string {
	alphaExpr := ""
	if b.period.IsConstant() {
		alphaExpr = fmt.Sprintf("2.0 / float64(%d+1)", b.period.AsInt())
	} else {
		alphaExpr = fmt.Sprintf("2.0 / (%s+1)", b.period.AsFloat64Cast())
	}

	code := b.indenter.Line(fmt.Sprintf("alpha := %s", alphaExpr))
	code += b.indenter.Line("newValue := alpha*currentSource + (1-alpha)*previousValue")
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.varName, "newValue"))
	return code
}

func (b *StatefulIndicatorBuilder) closeBlock() string {
	return b.indenter.Line("}")
}
