package codegen

import "fmt"

/* Backs a computed expression with an arrowCtx series so outer TA calls can iterate historical values */
type ArrowCtxSeriesAccessor struct {
	seriesName string
	callCode   string
}

func NewArrowCtxSeriesAccessor(seriesName, callCode string) *ArrowCtxSeriesAccessor {
	return &ArrowCtxSeriesAccessor{
		seriesName: seriesName,
		callCode:   callCode,
	}
}

func (a *ArrowCtxSeriesAccessor) GetPreamble() string {
	return fmt.Sprintf("%sSeries := arrowCtx.GetOrCreateSeries(%q); %sSeries.Set(%s)",
		a.seriesName, a.seriesName, a.seriesName, a.callCode)
}

func (a *ArrowCtxSeriesAccessor) GenerateLoopValueAccess(loopVar string) string {
	return fmt.Sprintf("%sSeries.Get(%s)", a.seriesName, loopVar)
}

func (a *ArrowCtxSeriesAccessor) GenerateInitialValueAccess(period int) string {
	return fmt.Sprintf("%sSeries.Get(%d-1)", a.seriesName, period)
}

func (a *ArrowCtxSeriesAccessor) GenerateCurrentValueAccess() string {
	return fmt.Sprintf("%sSeries.Get(0)", a.seriesName)
}

func (a *ArrowCtxSeriesAccessor) GetBaseOffset() int {
	return 0
}
