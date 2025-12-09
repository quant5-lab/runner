package codegen

import "strings"

type PatternMatcher interface {
	Matches(code string) bool
}

type seriesAccessPattern struct{}

func (p *seriesAccessPattern) Matches(code string) bool {
	return strings.Contains(code, ".GetCurrent()")
}

type comparisonPattern struct{}

func (p *comparisonPattern) Matches(code string) bool {
	operators := []string{">", "<", "==", "!=", ">=", "<="}
	for _, op := range operators {
		if strings.Contains(code, op) {
			return true
		}
	}
	return false
}

func NewSeriesAccessPattern() PatternMatcher {
	return &seriesAccessPattern{}
}

func NewComparisonPattern() PatternMatcher {
	return &comparisonPattern{}
}
