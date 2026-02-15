package context

var (
	timestampMatcher   = NewTimestampBarMatcher()
	valueRetriever     = NewSecurityValueRetriever()
	timeframeConverter = NewTimeframeConverter()
	timestampAligner   = NewTimestampAligner()
	boundaryAligner    = NewTimeframeBoundaryAligner()
)

func FindBarIndexByTimestamp(secCtx *Context, targetTimestamp int64) int {
	return timestampMatcher.MatchBarForTimestamp(secCtx, targetTimestamp)
}

func FindBarIndexByTimestampWithLookahead(secCtx *Context, targetTimestamp int64) int {
	return timestampMatcher.MatchBarWithLookahead(secCtx, targetTimestamp)
}

func GetSecurityValue(secCtx *Context, targetTimestamp int64, getValue func(*Context, int) float64) float64 {
	return valueRetriever.RetrieveValue(secCtx, targetTimestamp, getValue)
}

func TimeframeToSeconds(tf string) int64 {
	return timeframeConverter.ToSeconds(tf)
}

func TimeframeMultiplier(tf string) int64 {
	return timeframeConverter.extractNumericPart(tf)
}

func TimeframeFromSeconds(seconds int64) string {
	return timeframeConverter.FromSeconds(seconds)
}

func AlignTimestampToTimeframe(timestamp int64, timeframeSeconds int64) int64 {
	return timestampAligner.AlignToTimeframe(timestamp, timeframeSeconds)
}

/* Rounds timestamp to calendar-aware period boundary (weeks start Monday, months use actual boundaries) */
func AlignTimestampToPeriod(timestamp int64, timeframe string) int64 {
	return boundaryAligner.AlignToPeriod(timestamp, timeframe)
}

func GetAlignedTimestamp(ctx *Context, secTimeframe string) int64 {
	return timestampAligner.GetAlignedTimestamp(ctx, secTimeframe, timeframeConverter)
}
