package context

import "math"

type SecurityValueRetriever struct {
	barMatcher *TimestampBarMatcher
}

func NewSecurityValueRetriever() *SecurityValueRetriever {
	return &SecurityValueRetriever{
		barMatcher: NewTimestampBarMatcher(),
	}
}

func (r *SecurityValueRetriever) RetrieveValue(
	securityContext *Context,
	targetTimestamp int64,
	valueExtractor func(*Context, int) float64,
) float64 {
	barIndex := r.barMatcher.MatchBarForTimestamp(securityContext, targetTimestamp)

	if !r.isValidBarIndex(barIndex, securityContext) {
		return math.NaN()
	}

	return r.extractValueWithTemporaryIndex(securityContext, barIndex, valueExtractor)
}

func (r *SecurityValueRetriever) isValidBarIndex(index int, context *Context) bool {
	return index >= 0 && index < len(context.Data)
}

func (r *SecurityValueRetriever) extractValueWithTemporaryIndex(
	context *Context,
	barIndex int,
	extractor func(*Context, int) float64,
) float64 {
	originalIndex := context.BarIndex
	context.BarIndex = barIndex

	value := extractor(context, barIndex)

	context.BarIndex = originalIndex

	return value
}
