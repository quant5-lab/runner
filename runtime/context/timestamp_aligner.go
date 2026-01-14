package context

type TimestampAligner struct{}

func NewTimestampAligner() *TimestampAligner {
	return &TimestampAligner{}
}

func (a *TimestampAligner) AlignToTimeframe(timestamp int64, timeframeSeconds int64) int64 {
	if timeframeSeconds <= 0 {
		return timestamp
	}
	return (timestamp / timeframeSeconds) * timeframeSeconds
}

func (a *TimestampAligner) GetAlignedTimestamp(ctx *Context, secTimeframe string, converter *TimeframeConverter) int64 {
	if ctx.BarIndex < 0 || ctx.BarIndex >= len(ctx.Data) {
		return 0
	}

	currentBarTime := ctx.Data[ctx.BarIndex].Time
	secTfSeconds := converter.ToSeconds(secTimeframe)

	return a.AlignToTimeframe(currentBarTime, secTfSeconds)
}
