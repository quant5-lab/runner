package context

var (
	timestampMatcher   = NewTimestampBarMatcher()
	valueRetriever     = NewSecurityValueRetriever()
	timeframeConverter = NewTimeframeConverter()
	timestampAligner   = NewTimestampAligner()
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

/* TimeframeToSeconds converts Pine timeframe string to seconds
 * Examples: "1h" → 3600, "1D" → 86400, "5m" → 300
 */
func TimeframeToSeconds(tf string) int64 {
	return timeframeConverter.ToSeconds(tf)
}

/* AlignTimestampToTimeframe rounds timestamp down to timeframe boundary
 * Example: 2024-01-01 14:30:00 aligned to 1D → 2024-01-01 00:00:00
 */
func AlignTimestampToTimeframe(timestamp int64, timeframeSeconds int64) int64 {
	return timestampAligner.AlignToTimeframe(timestamp, timeframeSeconds)
}

/* GetAlignedTimestamp returns timestamp aligned to security timeframe
 * Used for upsampling: repeat daily value across all hourly bars of that day
 */
func GetAlignedTimestamp(ctx *Context, secTimeframe string) int64 {
	return timestampAligner.GetAlignedTimestamp(ctx, secTimeframe, timeframeConverter)
}
