package codegen

import "testing"

func TestSeriesAccessPattern_Matches(t *testing.T) {
	matcher := NewSeriesAccessPattern()

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"simple Series access", "priceSeries.GetCurrent()", true},
		{"different variable name", "enabledSeries.GetCurrent()", true},
		{"with whitespace", "price Series . GetCurrent ( )", false},
		{"historical access Get(N)", "varSeries.Get(1)", false},
		{"partial match Series only", "priceSeries", false},
		{"partial match method only", "GetCurrent()", false},
		{"identifier without Series", "price", false},
		{"comparison expression", "price > 100", false},
		{"empty string", "", false},
		{"nested in expression", "ta.sma(closeSeries.GetCurrent(), 20)", true},
		{"multiple Series access", "aSeries.GetCurrent() + bSeries.GetCurrent()", true},
		{"case sensitive", "priceseries.getcurrent()", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := matcher.Matches(tt.code); result != tt.expected {
				t.Errorf("code=%q: expected %v, got %v", tt.code, tt.expected, result)
			}
		})
	}
}

func TestComparisonPattern_Matches(t *testing.T) {
	matcher := NewComparisonPattern()

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"greater than", "price > 100", true},
		{"less than", "price < 100", true},
		{"equal", "price == 100", true},
		{"not equal", "price != 100", true},
		{"greater or equal", "price >= 100", true},
		{"less or equal", "price <= 100", true},
		{"no operator", "priceSeries.GetCurrent()", false},
		{"arithmetic operator", "price + 100", false},
		{"multiplication", "price * 2", false},
		{"division", "price / 2", false},
		{"empty string", "", false},
		{"multiple comparisons", "a > 10 && b < 20", true},
		{"partial operator >", ">", true},
		{"assignment operator", "x = 5", false},
		{"combined with Series", "priceSeries.GetCurrent() > 100", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := matcher.Matches(tt.code); result != tt.expected {
				t.Errorf("code=%q: expected %v, got %v", tt.code, tt.expected, result)
			}
		})
	}
}

func TestPatternMatcher_BoundaryConditions(t *testing.T) {
	seriesMatcher := NewSeriesAccessPattern()
	comparisonMatcher := NewComparisonPattern()

	tests := []struct {
		name             string
		code             string
		expectSeries     bool
		expectComparison bool
	}{
		{
			name:             "empty code",
			code:             "",
			expectSeries:     false,
			expectComparison: false,
		},
		{
			name:             "only whitespace",
			code:             "   \t\n  ",
			expectSeries:     false,
			expectComparison: false,
		},
		{
			name:             "Series and comparison",
			code:             "priceSeries.GetCurrent() > 100",
			expectSeries:     true,
			expectComparison: true,
		},
		{
			name:             "neither pattern",
			code:             "bar.Close",
			expectSeries:     false,
			expectComparison: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := seriesMatcher.Matches(tt.code); result != tt.expectSeries {
				t.Errorf("Series pattern: expected %v, got %v", tt.expectSeries, result)
			}
			if result := comparisonMatcher.Matches(tt.code); result != tt.expectComparison {
				t.Errorf("Comparison pattern: expected %v, got %v", tt.expectComparison, result)
			}
		})
	}
}
