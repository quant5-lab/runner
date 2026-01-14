package codegen

import (
	"fmt"
)

// StatefulIndicatorBuilder generates code for TA indicators that maintain state
// across bars by referencing their own previous values (RMA, EMA, etc.)
// Unlike window-based indicators, these use recursive formulas with previous results
type StatefulIndicatorBuilder struct {
	indicatorName string
	varName       string
	period        PeriodExpression
	accessor      AccessGenerator
	needsNaN      bool
	indenter      CodeIndenter
	context       StatefulIndicatorContext
}

// NewStatefulIndicatorBuilder creates builder for stateful indicators
func NewStatefulIndicatorBuilder(
	indicatorName string,
	varName string,
	period PeriodExpression,
	accessor AccessGenerator,
	needsNaN bool,
	context StatefulIndicatorContext,
) *StatefulIndicatorBuilder {
	return &StatefulIndicatorBuilder{
		indicatorName: indicatorName,
		varName:       varName,
		period:        period,
		accessor:      accessor,
		needsNaN:      needsNaN,
		indenter:      NewCodeIndenter(),
		context:       context,
	}
}

// BuildRMA generates stateful RMA calculation using previous RMA values
// RMA formula: rma[i] = alpha * source[i] + (1-alpha) * rma[i-1]
// where alpha = 1/period
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

// BuildEMA generates stateful EMA calculation using previous EMA values
// EMA formula: ema[i] = alpha * source[i] + (1-alpha) * ema[i-1]
// where alpha = 2/(period+1)
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
	/* For warmup check, use evaluated constant if possible, otherwise expression */
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
	/* For initialization check, use evaluated constant if possible, otherwise expression */
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

	/* For loop bound, use literal for constants */
	loopBound := ""
	if b.period.IsConstant() {
		loopBound = fmt.Sprintf("%d", b.period.AsInt())
	} else {
		loopBound = b.period.AsIntCast()
	}
	code += b.indenter.Line(fmt.Sprintf("for j := 0; j < %s; j++ {", loopBound))
	b.indenter.IncreaseIndent()

	valueAccess := b.accessor.GenerateLoopValueAccess("j")
	if b.needsNaN {
		code += b.indenter.Line(fmt.Sprintf("val := %s", valueAccess))
		code += b.indenter.Line("if math.IsNaN(val) {")
		b.indenter.IncreaseIndent()
		code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.varName, "math.NaN()"))
		code += b.indenter.Line("break")
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("}")
		code += b.indenter.Line("_sma_accumulator += val")
	} else {
		code += b.indenter.Line(fmt.Sprintf("_sma_accumulator += %s", valueAccess))
	}

	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")

	/* For SMA division, optimize constant periods */
	smaDiv := ""
	if b.period.IsConstant() {
		smaDiv = fmt.Sprintf("float64(%d)", b.period.AsInt())
	} else {
		smaDiv = b.period.AsFloat64Cast()
	}
	code += b.indenter.Line(fmt.Sprintf("initialValue := _sma_accumulator / %s", smaDiv))
	code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.varName, "initialValue"))

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
		code += b.indenter.Line("if math.IsNaN(currentSource) || math.IsNaN(previousValue) {")
		b.indenter.IncreaseIndent()
		code += b.indenter.Line(b.context.GenerateSeriesUpdate(b.varName, "math.NaN()"))
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
	/* For RMA alpha, optimize constant periods to avoid redundant cast */
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
	/* For EMA alpha, optimize constant periods to avoid redundant cast */
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
