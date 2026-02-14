package codegen

import "fmt"

type ArrowBuiltinAccessGenerator struct {
	registry   *BuiltinIdentifierRegistry
	formulaGen *DerivedPriceFormulaGenerator
}

func NewArrowBuiltinAccessGenerator(
	registry *BuiltinIdentifierRegistry,
	formulaGen *DerivedPriceFormulaGenerator,
) *ArrowBuiltinAccessGenerator {
	return &ArrowBuiltinAccessGenerator{
		registry:   registry,
		formulaGen: formulaGen,
	}
}

func (g *ArrowBuiltinAccessGenerator) GenerateCurrentAccess(name string) string {
	if g == nil || g.registry == nil || g.formulaGen == nil {
		return ""
	}

	if g.registry.IsDerivedPrice(name) {
		return g.formulaGen.Generate(name,
			"ctx.Data[ctx.BarIndex].High",
			"ctx.Data[ctx.BarIndex].Low",
			"ctx.Data[ctx.BarIndex].Close",
			"ctx.Data[ctx.BarIndex].Open")
	}

	if info, ok := g.registry.CalendarInfo(name); ok {
		return SeriesLookupIIFE(info.SeriesName)
	}

	if field, ok := OHLCVFieldName(name); ok {
		return fmt.Sprintf("ctx.Data[ctx.BarIndex].%s", field)
	}

	switch name {
	case "tr":
		return arrowTrueRangeIIFE()
	case "bar_index":
		return "float64(ctx.BarIndex)"
	case "time":
		return "float64(ctx.Data[ctx.BarIndex].Time * 1000)"
	case "time_close":
		return SeriesLookupIIFE("time_closeSeries")
	case "time_tradingday":
		return SeriesLookupIIFE("time_tradingdaySeries")
	case "last_bar_index":
		return "float64(len(ctx.Data) - 1)"
	case "last_bar_time":
		return "float64(ctx.Data[len(ctx.Data)-1].Time * 1000)"
	case "timenow":
		return "float64(ctx.Data[len(ctx.Data)-1].Time * 1000)"
	default:
		return ""
	}
}

func (g *ArrowBuiltinAccessGenerator) GenerateHistoricalAccess(name string, offset int) string {
	if g == nil || g.registry == nil || g.formulaGen == nil {
		return ""
	}

	if name == "tr" {
		return TrueRangeArrowIIFE(fmt.Sprintf("%d", offset))
	}

	if g.registry.IsDerivedPrice(name) {
		accessor := NewDerivedPriceAccessor(name, offset)
		formula := accessor.GenerateFormulaAtOffset(fmt.Sprintf("ctx.BarIndex-%d", offset))
		return arrowBoundsCheckedExpr(offset, formula)
	}

	if info, ok := g.registry.CalendarInfo(name); ok {
		return SeriesLookupWithOffsetIIFE(info.SeriesName, offset)
	}

	if field, ok := OHLCVFieldName(name); ok {
		return arrowBoundsCheckedExpr(offset, fmt.Sprintf("ctx.Data[ctx.BarIndex-%d].%s", offset, field))
	}

	switch name {
	case "bar_index":
		return arrowBoundsCheckedExpr(offset, fmt.Sprintf("float64(ctx.BarIndex-%d)", offset))
	case "time":
		return arrowBoundsCheckedExpr(offset, fmt.Sprintf("float64(ctx.Data[ctx.BarIndex-%d].Time * 1000)", offset))
	case "time_close":
		return SeriesLookupWithOffsetIIFE("time_closeSeries", offset)
	case "time_tradingday":
		return SeriesLookupWithOffsetIIFE("time_tradingdaySeries", offset)
	default:
		return ""
	}
}

func (g *ArrowBuiltinAccessGenerator) GenerateStrategyAccess(property string) string {
	if g == nil {
		return ""
	}

	seriesName := ""
	switch property {
	case "position_avg_price":
		seriesName = StrategyPositionAvgPriceSeriesName
	case "position_size":
		seriesName = StrategyPositionSizeSeriesName
	case "equity":
		seriesName = StrategyEquitySeriesName
	case "netprofit":
		seriesName = StrategyNetProfitSeriesName
	case "closedtrades":
		seriesName = StrategyClosedTradesSeriesName
	case "position_entry_name":
		return `""`
	default:
		return "math.NaN()"
	}
	return SeriesLookupIIFE(seriesName)
}

func arrowBoundsCheckedExpr(offset int, expr string) string {
	return fmt.Sprintf(
		"func() float64 { if ctx.BarIndex-%d < 0 { return math.NaN() }; return %s }()",
		offset, expr)
}

func arrowTrueRangeIIFE() string {
	return "func() float64 { " +
		"curBar := ctx.Data[ctx.BarIndex]; " +
		"if ctx.BarIndex < 1 { return curBar.High - curBar.Low }; " +
		"prevClose := ctx.Data[ctx.BarIndex-1].Close; " +
		"return math.Max(curBar.High - curBar.Low, math.Max(math.Abs(curBar.High - prevClose), math.Abs(curBar.Low - prevClose))) " +
		"}()"
}

func SeriesLookupIIFE(seriesName string) string {
	return fmt.Sprintf(
		"func() float64 { if s, ok := ctx.LookupSeries(%q); ok { return s.GetCurrent() }; return math.NaN() }()",
		seriesName)
}

func SeriesPointerLookupIIFE(seriesName string) string {
	return fmt.Sprintf(
		"func() *series.Series { s, _ := ctx.LookupSeries(%q); return s }()",
		seriesName)
}

func SeriesLookupWithOffsetIIFE(seriesName string, offset int) string {
	if offset == 0 {
		return SeriesLookupIIFE(seriesName)
	}
	return fmt.Sprintf(
		"func() float64 { if s, ok := ctx.LookupSeries(%q); ok { return s.Get(%d) }; return math.NaN() }()",
		seriesName, offset)
}
