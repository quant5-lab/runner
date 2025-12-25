package codegen

import "fmt"

/*
SeriesBackedComplexExpressionAccessor enables historical access to complex expressions.

Generates temp series for complex expressions (ternary, binary, call) used in TA functions.
Provides both current value access and historical offset access via series.Get().
*/
type SeriesBackedComplexExpressionAccessor struct {
	seriesName     string
	expressionCode string
	needsAdvance   bool
}

func NewSeriesBackedComplexExpressionAccessor(seriesName, expressionCode string) *SeriesBackedComplexExpressionAccessor {
	return &SeriesBackedComplexExpressionAccessor{
		seriesName:     seriesName,
		expressionCode: expressionCode,
		needsAdvance:   true,
	}
}

func (a *SeriesBackedComplexExpressionAccessor) GenerateLoopValueAccess(loopVar string) string {
	return fmt.Sprintf("%s.Get(%s)", a.seriesName, loopVar)
}

func (a *SeriesBackedComplexExpressionAccessor) GenerateInitialValueAccess(period int) string {
	return fmt.Sprintf("%s.Get(0)", a.seriesName)
}

func (a *SeriesBackedComplexExpressionAccessor) GetPreamble() string {
	baseName := a.seriesName
	if len(baseName) > 6 && baseName[len(baseName)-6:] == "Series" {
		baseName = baseName[:len(baseName)-6]
	}

	init := fmt.Sprintf("%s := arrowCtx.GetOrCreateSeries(%q); ", a.seriesName, baseName)
	set := fmt.Sprintf("%s.Set(%s)", a.seriesName, a.expressionCode)

	return init + set
}

func (a *SeriesBackedComplexExpressionAccessor) GetSeriesName() string {
	return a.seriesName
}

func (a *SeriesBackedComplexExpressionAccessor) NeedsAdvance() bool {
	return a.needsAdvance
}
