package codegen

import "fmt"

/* RSIIndicatorBuilder generates Relative Strength Index calculation code.
 *
 * Architecture: Composes generic primitives for RSI-specific formula
 * - ChangeCalculator: source[0] - source[1]
 * - DirectionalSplitGenerator: gains/losses separation
 * - StatefulIndicatorBuilder (2x): RMA smoothing of gains and losses
 * - Final formula: RSI = 100 - 100/(1 + rmaGains/rmaLosses)
 *
 * Context awareness: Works in both TopLevel and Arrow function contexts
 * Internal series: Manages gains, losses, rmaGains, rmaLosses automatically
 */
type RSIIndicatorBuilder struct {
	base     *CompositeIndicatorBuilder
	needsNaN bool
}

func NewRSIIndicatorBuilder(
	resultVarName string,
	period PeriodExpression,
	sourceAccessor AccessGenerator,
	needsNaN bool,
	context StatefulIndicatorContext,
) *RSIIndicatorBuilder {
	return &RSIIndicatorBuilder{
		base:     NewCompositeIndicatorBuilder(resultVarName, period, sourceAccessor, context),
		needsNaN: needsNaN,
	}
}

/* Build generates complete RSI calculation code */
func (b *RSIIndicatorBuilder) Build() string {
	gainsName := b.base.GenerateInternalSeriesName("gains")
	lossesName := b.base.GenerateInternalSeriesName("losses")
	rmaGainsName := b.base.GenerateInternalSeriesName("rma_gains")
	rmaLossesName := b.base.GenerateInternalSeriesName("rma_losses")

	code := b.generateChangeStep()
	code += b.generateDirectionalSplitStep(gainsName, lossesName)
	code += b.generateRMASmoothingSteps(gainsName, lossesName, rmaGainsName, rmaLossesName)
	code += b.generateFinalFormula(rmaGainsName, rmaLossesName)

	return code
}

/* generateChangeStep produces bar-to-bar change calculation */
func (b *RSIIndicatorBuilder) generateChangeStep() string {
	changeCalc := NewChangeCalculator(b.base.sourceAccessor)
	return changeCalc.GenerateChangeCode(b.changeVarName())
}

/* generateDirectionalSplitStep separates change into gains and losses */
func (b *RSIIndicatorBuilder) generateDirectionalSplitStep(gainsName, lossesName string) string {
	splitGen := NewDirectionalSplitGenerator(gainsName, lossesName, b.base.context)
	return splitGen.GenerateSplitCode(b.changeVarName())
}

/* generateRMASmoothingSteps applies RMA to both gains and losses */
func (b *RSIIndicatorBuilder) generateRMASmoothingSteps(
	gainsName, lossesName, rmaGainsName, rmaLossesName string,
) string {
	gainsAccessor := NewInternalSeriesAccessor(gainsName, b.base.context)
	gainsRMABuilder := NewStatefulIndicatorBuilder(
		"rma", rmaGainsName, b.base.period, gainsAccessor, false, b.base.context,
	)

	lossesAccessor := NewInternalSeriesAccessor(lossesName, b.base.context)
	lossesRMABuilder := NewStatefulIndicatorBuilder(
		"rma", rmaLossesName, b.base.period, lossesAccessor, false, b.base.context,
	)

	return gainsRMABuilder.BuildRMA() + lossesRMABuilder.BuildRMA()
}

/* generateFinalFormula calculates RSI from smoothed gains/losses */
func (b *RSIIndicatorBuilder) generateFinalFormula(rmaGainsName, rmaLossesName string) string {
	rmaGainsAccess := b.base.context.GenerateSeriesAccess(rmaGainsName, 0)
	rmaLossesAccess := b.base.context.GenerateSeriesAccess(rmaLossesName, 0)

	code := ""

	warmupExpr := ""
	if b.base.period.IsConstant() {
		warmupExpr = fmt.Sprintf("%d", b.base.period.AsInt())
	} else {
		warmupExpr = b.base.period.AsIntCast()
	}

	code += b.base.indenter.Line(fmt.Sprintf("if ctx.BarIndex < %s {", warmupExpr))
	b.base.indenter.IncreaseIndent()
	code += b.base.indenter.Line(b.base.context.GenerateSeriesUpdate(b.base.resultVarName, "math.NaN()"))
	b.base.indenter.DecreaseIndent()
	code += b.base.indenter.Line("} else {")
	b.base.indenter.IncreaseIndent()

	code += b.base.indenter.Line(fmt.Sprintf("rs := %s / %s", rmaGainsAccess, rmaLossesAccess))
	code += b.base.indenter.Line("rsi := 100.0 - (100.0 / (1.0 + rs))")
	code += b.base.indenter.Line(b.base.context.GenerateSeriesUpdate(b.base.resultVarName, "rsi"))

	b.base.indenter.DecreaseIndent()
	code += b.base.indenter.Line("}")

	return code
}

/* changeVarName generates consistent temporary variable name for change */
func (b *RSIIndicatorBuilder) changeVarName() string {
	return fmt.Sprintf("_%s_change", b.base.resultVarName)
}

/* GetInternalSeriesNames exposes internal series for declaration generation */
func (b *RSIIndicatorBuilder) GetInternalSeriesNames() []string {
	return b.base.internalSeries
}
