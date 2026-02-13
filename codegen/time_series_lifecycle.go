package codegen

import "fmt"

type TimeSeriesLifecycle struct {
	hasTimeClose      bool
	hasTimeTradingday bool
}

func NewTimeSeriesLifecycle(hasTimeClose, hasTimeTradingday bool) *TimeSeriesLifecycle {
	return &TimeSeriesLifecycle{
		hasTimeClose:      hasTimeClose,
		hasTimeTradingday: hasTimeTradingday,
	}
}

func (l *TimeSeriesLifecycle) HasUsage() bool {
	return l != nil && (l.hasTimeClose || l.hasTimeTradingday)
}

func (l *TimeSeriesLifecycle) NeedsTimezone() bool {
	return l != nil && l.hasTimeTradingday
}

func (l *TimeSeriesLifecycle) GenerateDeclarations(indent string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	if l.hasTimeClose {
		code += indent + "var time_closeSeries *series.Series\n"
	}
	if l.hasTimeTradingday {
		code += indent + "var time_tradingdaySeries *series.Series\n"
	}
	return code
}

func (l *TimeSeriesLifecycle) GenerateInitializations(indent string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	if l.hasTimeClose {
		code += indent + "time_closeSeries = series.NewSeries(len(ctx.Data))\n"
	}
	if l.hasTimeTradingday {
		code += indent + "time_tradingdaySeries = series.NewSeries(len(ctx.Data))\n"
	}
	return code
}

/* time_close falls back to timeframe-based computation on the last bar */
func (l *TimeSeriesLifecycle) GenerateBarPopulation(indent, iterVar string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	if l.hasTimeClose {
		code += indent + fmt.Sprintf("if %s < barCount-1 {\n", iterVar)
		code += indent + fmt.Sprintf("\ttime_closeSeries.Set(float64(ctx.Data[%s+1].Time * 1000))\n", iterVar)
		code += indent + "} else {\n"
		code += indent + "\ttime_closeSeries.Set(float64(bar.Time*1000 + context.TimeframeToSeconds(ctx.Timeframe)*1000))\n"
		code += indent + "}\n"
	}
	if l.hasTimeTradingday {
		code += indent + "func() {\n"
		code += indent + "\tbarTime := time.Unix(bar.Time, 0).In(exchangeLoc)\n"
		code += indent + "\ttradingDayStart := time.Date(barTime.Year(), barTime.Month(), barTime.Day(), 0, 0, 0, 0, exchangeLoc)\n"
		code += indent + "\ttime_tradingdaySeries.Set(float64(tradingDayStart.Unix() * 1000))\n"
		code += indent + "}()\n"
	}
	return code
}

func (l *TimeSeriesLifecycle) GenerateAdvancement(indent, iterVar string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	if l.hasTimeClose {
		code += indent + fmt.Sprintf("if %s < barCount-1 { time_closeSeries.Next() }\n", iterVar)
	}
	if l.hasTimeTradingday {
		code += indent + fmt.Sprintf("if %s < barCount-1 { time_tradingdaySeries.Next() }\n", iterVar)
	}
	return code
}

func (l *TimeSeriesLifecycle) GenerateRegistrations(indent string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	if l.hasTimeClose {
		code += indent + `ctx.RegisterSeries("time_closeSeries", time_closeSeries)` + "\n"
	}
	if l.hasTimeTradingday {
		code += indent + `ctx.RegisterSeries("time_tradingdaySeries", time_tradingdaySeries)` + "\n"
	}
	return code
}

func (l *TimeSeriesLifecycle) GenerateSuppressUnused(indent string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	if l.hasTimeClose {
		code += indent + "_ = time_closeSeries\n"
	}
	if l.hasTimeTradingday {
		code += indent + "_ = time_tradingdaySeries\n"
	}
	return code
}

func (l *TimeSeriesLifecycle) GenerateSymbolTableRegistrations(symbolTable SymbolTable) {
	if l == nil {
		return
	}
	if l.hasTimeClose && symbolTable != nil {
		symbolTable.Register("time_close", VariableTypeSeries)
	}
	if l.hasTimeTradingday && symbolTable != nil {
		symbolTable.Register("time_tradingday", VariableTypeSeries)
	}
}
