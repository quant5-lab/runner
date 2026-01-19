package codegen

import "fmt"

/*
ArrowAwareSeriesAccessor adapts standard Series access to arrow function context.

Responsibility (SRP):
  - Generate Series.GetCurrent() access code for loop values
  - Generate Series.Get(offset) access code for historical values
  - No knowledge of identifier resolution (delegates to ArrowIdentifierResolver)

Design:
  - Implements AccessGenerator interface for compatibility with inline TA generators
  - Uses composition: wraps identifier resolver for proper Series access
  - KISS: Simple delegation, no complex logic
*/
type ArrowAwareSeriesAccessor struct {
	seriesName string
}

func NewArrowAwareSeriesAccessor(seriesName string) *ArrowAwareSeriesAccessor {
	return &ArrowAwareSeriesAccessor{
		seriesName: seriesName,
	}
}

func (a *ArrowAwareSeriesAccessor) GenerateLoopValueAccess(loopVar string) string {
	return fmt.Sprintf("%sSeries.Get(%s)", a.seriesName, loopVar)
}

func (a *ArrowAwareSeriesAccessor) GenerateInitialValueAccess(period int) string {
	return fmt.Sprintf("%sSeries.Get(%d-1)", a.seriesName, period)
}

func (a *ArrowAwareSeriesAccessor) GenerateCurrentValueAccess() string {
	return fmt.Sprintf("%sSeries.GetCurrent()", a.seriesName)
}

/*
GetPreamble returns any setup code needed before the accessor is used.
*/
func (a *ArrowAwareSeriesAccessor) GetPreamble() string {
	return ""
}

/* GetBaseOffset returns 0 - arrow-aware series access is current bar relative */
func (a *ArrowAwareSeriesAccessor) GetBaseOffset() int {
	return 0
}
