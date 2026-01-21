package codegen

import "fmt"

/* CompositeIndicatorBuilder orchestrates multi-step indicator calculations.
 *
 * Unlike StatefulIndicatorBuilder (single recursive formula), composite indicators
 * require multiple intermediate series and sequential computations.
 *
 * Responsibilities:
 * - Internal series name generation (collision-free)
 * - Context-aware series lifecycle (TopLevel declarations vs Arrow lazy-init)
 * - Composition orchestration (delegates to primitives)
 *
 * Use: Base for RSI, DMI, MACD, Stochastic, and other multi-component indicators
 */
type CompositeIndicatorBuilder struct {
	resultVarName  string
	period         PeriodExpression
	sourceAccessor AccessGenerator
	context        StatefulIndicatorContext
	indenter       CodeIndenter
	internalSeries []string
}

func NewCompositeIndicatorBuilder(
	resultVarName string,
	period PeriodExpression,
	sourceAccessor AccessGenerator,
	context StatefulIndicatorContext,
) *CompositeIndicatorBuilder {
	return &CompositeIndicatorBuilder{
		resultVarName:  resultVarName,
		period:         period,
		sourceAccessor: sourceAccessor,
		context:        context,
		indenter:       NewCodeIndenter(),
		internalSeries: make([]string, 0),
	}
}

/* GenerateInternalSeriesName creates unique internal series identifiers */
func (b *CompositeIndicatorBuilder) GenerateInternalSeriesName(component string) string {
	name := fmt.Sprintf("_%s_%s", b.resultVarName, component)
	b.internalSeries = append(b.internalSeries, name)
	return name
}

/* GenerateSeriesDeclarations generates TopLevel series declarations for internal series */
func (b *CompositeIndicatorBuilder) GenerateSeriesDeclarations() string {
	if b.context.IsWithinArrowFunction() {
		return ""
	}

	code := ""
	for _, seriesName := range b.internalSeries {
		code += b.indenter.Line(fmt.Sprintf("var %sSeries *series.Series", seriesName))
	}
	return code
}

/* GenerateSeriesInitializations generates TopLevel series buffer allocations */
func (b *CompositeIndicatorBuilder) GenerateSeriesInitializations(dataLengthExpr string) string {
	if b.context.IsWithinArrowFunction() {
		return ""
	}

	code := ""
	for _, seriesName := range b.internalSeries {
		code += b.indenter.Line(fmt.Sprintf("%sSeries = series.NewSeries(%s)", seriesName, dataLengthExpr))
	}
	return code
}
