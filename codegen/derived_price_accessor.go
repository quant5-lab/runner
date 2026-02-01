package codegen

import "fmt"

type DerivedPriceAccessor struct {
	priceName  string
	baseOffset int
}

func NewDerivedPriceAccessor(priceName string, baseOffset int) *DerivedPriceAccessor {
	return &DerivedPriceAccessor{
		priceName:  priceName,
		baseOffset: baseOffset,
	}
}

func (a *DerivedPriceAccessor) GenerateLoopValueAccess(loopVar string) string {
	if a.baseOffset == 0 {
		return a.GenerateFormulaAtOffset(loopVar)
	}
	return a.GenerateFormulaAtOffset(fmt.Sprintf("%s+%d", loopVar, a.baseOffset))
}

func (a *DerivedPriceAccessor) GenerateInitialValueAccess(period int) string {
	totalOffset := period - 1 + a.baseOffset
	return a.GenerateFormulaAtOffset(fmt.Sprintf("%d", totalOffset))
}

func (a *DerivedPriceAccessor) GenerateCurrentValueAccess() string {
	if a.baseOffset == 0 {
		return a.GenerateFormulaAtOffsetCurrent()
	}
	return a.GenerateFormulaAtOffset(fmt.Sprintf("%d", a.baseOffset))
}

func (a *DerivedPriceAccessor) GetPreamble() string {
	return ""
}

func (a *DerivedPriceAccessor) GetBaseOffset() int {
	return a.baseOffset
}

func (a *DerivedPriceAccessor) GenerateFormulaAtOffset(offset string) string {
	switch a.priceName {
	case "hl2":
		return fmt.Sprintf("((highSeries.Get(%s) + lowSeries.Get(%s)) / 2)", offset, offset)
	case "hlc3":
		return fmt.Sprintf("((highSeries.Get(%s) + lowSeries.Get(%s) + closeSeries.Get(%s)) / 3)", offset, offset, offset)
	case "ohlc4":
		return fmt.Sprintf("((openSeries.Get(%s) + highSeries.Get(%s) + lowSeries.Get(%s) + closeSeries.Get(%s)) / 4)", offset, offset, offset, offset)
	case "hlcc4":
		return fmt.Sprintf("((highSeries.Get(%s) + lowSeries.Get(%s) + closeSeries.Get(%s) + closeSeries.Get(%s)) / 4)", offset, offset, offset, offset)
	default:
		return "math.NaN()"
	}
}

func (a *DerivedPriceAccessor) GenerateFormulaAtOffsetCurrent() string {
	switch a.priceName {
	case "hl2":
		return "((highSeries.GetCurrent() + lowSeries.GetCurrent()) / 2)"
	case "hlc3":
		return "((highSeries.GetCurrent() + lowSeries.GetCurrent() + closeSeries.GetCurrent()) / 3)"
	case "ohlc4":
		return "((openSeries.GetCurrent() + highSeries.GetCurrent() + lowSeries.GetCurrent() + closeSeries.GetCurrent()) / 4)"
	case "hlcc4":
		return "((highSeries.GetCurrent() + lowSeries.GetCurrent() + closeSeries.GetCurrent() + closeSeries.GetCurrent()) / 4)"
	default:
		return "math.NaN()"
	}
}
