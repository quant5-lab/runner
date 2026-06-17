package context

// BarOpenTimeAtTimeframe returns the Unix timestamp in seconds of the period
// boundary that contains barTimeSec at the scale of timeframe (UTC origin).
func BarOpenTimeAtTimeframe(barTimeSec int64, timeframe string) int64 {
	return boundaryAligner.AlignToPeriod(barTimeSec, timeframe)
}

// BarOpenTimeAtTimeframeWithAnchor returns the period-boundary timestamp for
// barTimeSec at the scale of timeframe, tiling from anchor.SessionOpenMinute
// within anchor.Timezone.  This is the session-aware form used by generated
// strategy code so that time(tf) boundaries match TradingView for non-UTC
// exchanges such as MOEX and NYSE.
func BarOpenTimeAtTimeframeWithAnchor(barTimeSec int64, timeframe string, anchor PeriodAnchor) int64 {
	return boundaryAligner.AlignToPeriodWithAnchor(barTimeSec, timeframe, anchor)
}
