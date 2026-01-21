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
		return a.GenerateFormulaAtOffset(fmt.Sprintf("ctx.BarIndex-%s", loopVar))
	}
	return a.GenerateFormulaAtOffset(fmt.Sprintf("ctx.BarIndex-(%s+%d)", loopVar, a.baseOffset))
}

func (a *DerivedPriceAccessor) GenerateInitialValueAccess(period int) string {
	totalOffset := period - 1 + a.baseOffset
	return a.GenerateFormulaAtOffset(fmt.Sprintf("ctx.BarIndex-%d", totalOffset))
}

func (a *DerivedPriceAccessor) GenerateCurrentValueAccess() string {
	if a.baseOffset == 0 {
		return a.GenerateFormulaAtOffset("ctx.BarIndex")
	}
	return a.GenerateFormulaAtOffset(fmt.Sprintf("ctx.BarIndex-%d", a.baseOffset))
}

func (a *DerivedPriceAccessor) GetPreamble() string {
	return ""
}

func (a *DerivedPriceAccessor) GetBaseOffset() int {
	return a.baseOffset
}

func (a *DerivedPriceAccessor) GenerateFormulaAtOffset(indexExpr string) string {
	switch a.priceName {
	case "hl2":
		return fmt.Sprintf("((ctx.Data[%s].High + ctx.Data[%s].Low) / 2)", indexExpr, indexExpr)
	case "hlc3":
		return fmt.Sprintf("((ctx.Data[%s].High + ctx.Data[%s].Low + ctx.Data[%s].Close) / 3)", indexExpr, indexExpr, indexExpr)
	case "ohlc4":
		return fmt.Sprintf("((ctx.Data[%s].Open + ctx.Data[%s].High + ctx.Data[%s].Low + ctx.Data[%s].Close) / 4)", indexExpr, indexExpr, indexExpr, indexExpr)
	case "hlcc4":
		return fmt.Sprintf("((ctx.Data[%s].High + ctx.Data[%s].Low + ctx.Data[%s].Close + ctx.Data[%s].Close) / 4)", indexExpr, indexExpr, indexExpr, indexExpr)
	default:
		return "math.NaN()"
	}
}
