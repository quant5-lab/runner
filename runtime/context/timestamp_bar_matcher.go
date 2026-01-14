package context

import "time"

type TimestampBarMatcher struct {
	indexFinder *BarIndexFinder
}

func NewTimestampBarMatcher() *TimestampBarMatcher {
	return &TimestampBarMatcher{
		indexFinder: NewBarIndexFinder(),
	}
}

func (m *TimestampBarMatcher) MatchBarForTimestamp(
	securityContext *Context,
	targetTimestamp int64,
) int {
	return m.indexFinder.FindContainingBar(
		securityContext.Data,
		targetTimestamp,
	)
}

func (m *TimestampBarMatcher) MatchBarWithLookahead(
	securityContext *Context,
	targetTimestamp int64,
) int {
	if len(securityContext.Data) == 0 {
		return -1
	}

	targetDate := time.Unix(targetTimestamp, 0).UTC()
	targetYear, targetMonth, targetDay := targetDate.Date()

	for i := 0; i < len(securityContext.Data); i++ {
		barDate := time.Unix(securityContext.Data[i].Time, 0).UTC()
		barYear, barMonth, barDay := barDate.Date()

		if barYear == targetYear && barMonth == targetMonth && barDay == targetDay {
			return i
		}
	}

	return m.indexFinder.FindContainingBar(securityContext.Data, targetTimestamp)
}

func (m *TimestampBarMatcher) findFirstBarAfter(data []OHLCV, timestamp int64) int {
	for i := 0; i < len(data); i++ {
		if data[i].Time > timestamp {
			return i
		}
	}
	return -1
}

func (m *TimestampBarMatcher) handleBeyondLastBar(data []OHLCV) int {
	return len(data) - 1
}
