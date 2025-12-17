package context

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
	return m.MatchBarForTimestamp(securityContext, targetTimestamp)
}
