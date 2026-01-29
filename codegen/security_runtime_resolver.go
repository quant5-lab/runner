package codegen

func isRuntimeSymbol(symbol string) bool {
	return symbol == "syminfo.tickerid" || symbol == "syminfo.ticker" || symbol == "tickerid" || symbol == "ticker"
}

func isRuntimeTimeframe(timeframe string) bool {
	return timeframe == "timeframe.period"
}

func runtimePlaceholder() string {
	return "%s"
}
