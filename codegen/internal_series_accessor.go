package codegen

import "fmt"

/* InternalSeriesAccessor generates access code for series created and managed
 * within composite indicators (RSI, DMI, etc.).
 *
 * Unlike OHLCVFieldAccessGenerator (external OHLC data) or SeriesVariableAccessGenerator
 * (user-defined variables), this accessor targets intermediate computed series that
 * exist only within the indicator's scope.
 *
 * Use Case Example (RSI):
 *   RSI needs two internal series: gains and losses
 *   accessor := NewInternalSeriesAccessor("_rsi_gains_abc123", context)
 *   accessor.GenerateLoopValueAccess("j") → arrowCtx.GetOrCreateSeries("_rsi_gains_abc123").Get(j)
 */
type InternalSeriesAccessor struct {
	seriesName string
	context    StatefulIndicatorContext
}

func NewInternalSeriesAccessor(seriesName string, context StatefulIndicatorContext) *InternalSeriesAccessor {
	return &InternalSeriesAccessor{
		seriesName: seriesName,
		context:    context,
	}
}

func (a *InternalSeriesAccessor) GenerateLoopValueAccess(loopVar string) string {
	if a.context.IsWithinArrowFunction() {
		return fmt.Sprintf("arrowCtx.GetOrCreateSeries(%q).Get(%s)", a.seriesName, loopVar)
	}
	return fmt.Sprintf("%sSeries.Get(%s)", a.seriesName, loopVar)
}

func (a *InternalSeriesAccessor) GenerateInitialValueAccess(period int) string {
	return a.context.GenerateSeriesAccess(a.seriesName, period-1)
}

func (a *InternalSeriesAccessor) GenerateCurrentValueAccess() string {
	return a.context.GenerateSeriesAccess(a.seriesName, 0)
}

func (a *InternalSeriesAccessor) GetBaseOffset() int {
	return 0
}
