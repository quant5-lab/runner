package codegen

import "testing"

/* Tests Series access pattern detection across all variations */
func TestSeriesAccessPattern_Comprehensive(t *testing.T) {
	matcher := NewSeriesAccessPattern()

	tests := []struct {
		name        string
		code        string
		expected    bool
		description string
	}{
		// Current value access patterns
		{
			name:        "GetCurrent standard",
			code:        "priceSeries.GetCurrent()",
			expected:    true,
			description: "Standard current value access",
		},
		{
			name:        "GetCurrent different variable",
			code:        "volumeSeries.GetCurrent()",
			expected:    true,
			description: "Pattern matches any Series variable name",
		},
		{
			name:        "GetCurrent with prefix",
			code:        "bb_topSeries.GetCurrent()",
			expected:    true,
			description: "Series variables can have underscores",
		},

		// Historical access patterns
		{
			name:        "Get(1) one bar ago",
			code:        "closeSeries.Get(1)",
			expected:    true,
			description: "Historical access 1 bar lookback",
		},
		{
			name:        "Get(2) two bars ago",
			code:        "has_active_tradeSeries.Get(2)",
			expected:    true,
			description: "Historical access 2 bars lookback (BB9 pattern)",
		},
		{
			name:        "Get(N) deep history",
			code:        "indicatorSeries.Get(10)",
			expected:    true,
			description: "Deep historical lookback",
		},
		{
			name:        "Get(0) explicit current",
			code:        "valueSeries.Get(0)",
			expected:    true,
			description: "Explicit current value with Get(0)",
		},

		// Negative cases - not Series patterns
		{
			name:        "identifier only",
			code:        "price",
			expected:    false,
			description: "Plain identifier is not Series access",
		},
		{
			name:        "Series without method",
			code:        "priceSeries",
			expected:    false,
			description: "Series variable without method call",
		},
		{
			name:        "GetCurrent without Series",
			code:        "GetCurrent()",
			expected:    false,
			description: "Method name alone is not Series pattern",
		},
		{
			name:        "different method",
			code:        "priceSeries.Set(1.0)",
			expected:    false,
			description: "Other Series methods not matched",
		},
		{
			name:        "lowercase getcurrent",
			code:        "priceseries.getcurrent()",
			expected:    false,
			description: "Method name is case sensitive",
		},

		// Nested and complex patterns
		{
			name:        "nested in function call",
			code:        "ta.Sma(closeSeries.GetCurrent(), 20)",
			expected:    true,
			description: "Series access within function arguments",
		},
		{
			name:        "nested in arithmetic",
			code:        "highSeries.GetCurrent() - lowSeries.Get(1)",
			expected:    true,
			description: "Multiple Series access in expression",
		},
		{
			name:        "nested in comparison",
			code:        "priceSeries.GetCurrent() > bbTopSeries.Get(1)",
			expected:    true,
			description: "Series access in both sides of comparison",
		},
		{
			name:        "deeply nested",
			code:        "ta.Ema(ta.Sma(closeSeries.GetCurrent(), 10), 5)",
			expected:    true,
			description: "Series access in nested function calls",
		},

		// Edge cases
		{
			name:        "empty string",
			code:        "",
			expected:    false,
			description: "Empty code has no pattern",
		},
		{
			name:        "whitespace only",
			code:        "   \t\n  ",
			expected:    false,
			description: "Whitespace-only code has no pattern",
		},
		{
			name:        "Series in string literal",
			code:        `"priceSeries.GetCurrent()"`,
			expected:    true,
			description: "Pattern matches even in strings (string search)",
		},
		{
			name:        "Series in comment (if included)",
			code:        "// priceSeries.GetCurrent()",
			expected:    true,
			description: "Pattern matches in comments (string search)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.Matches(tt.code)
			if result != tt.expected {
				t.Errorf("%s\ncode=%q\nexpected: %v\ngot:      %v",
					tt.description, tt.code, tt.expected, result)
			}
		})
	}
}

/* Tests comparison pattern detection with all operators */
func TestComparisonPattern_AllOperators(t *testing.T) {
	matcher := NewComparisonPattern()

	tests := []struct {
		name        string
		code        string
		expected    bool
		description string
	}{
		// All comparison operators
		{
			name:        "greater than >",
			code:        "price > 100",
			expected:    true,
			description: "Greater than operator",
		},
		{
			name:        "less than <",
			code:        "volume < 1000",
			expected:    true,
			description: "Less than operator",
		},
		{
			name:        "greater or equal >=",
			code:        "close >= open",
			expected:    true,
			description: "Greater than or equal operator",
		},
		{
			name:        "less or equal <=",
			code:        "low <= support",
			expected:    true,
			description: "Less than or equal operator",
		},
		{
			name:        "equality ==",
			code:        "status == 1",
			expected:    true,
			description: "Equality operator",
		},
		{
			name:        "not equal !=",
			code:        "signal != 0",
			expected:    true,
			description: "Not equal operator",
		},

		// Operator variations
		{
			name:        "operator only >",
			code:        ">",
			expected:    true,
			description: "Bare operator matches",
		},
		{
			name:        "operator only >=",
			code:        ">=",
			expected:    true,
			description: "Compound operator alone matches",
		},

		// Complex expressions with comparisons
		{
			name:        "comparison in logical AND",
			code:        "price > 100 && volume > 1000",
			expected:    true,
			description: "Multiple comparisons in logical expression",
		},
		{
			name:        "comparison in logical OR",
			code:        "close < low || close > high",
			expected:    true,
			description: "Comparison in disjunction",
		},
		{
			name:        "nested comparison",
			code:        "(price > 100) && (volume < 1000)",
			expected:    true,
			description: "Parenthesized comparisons",
		},
		{
			name:        "comparison with Series",
			code:        "closeSeries.GetCurrent() > openSeries.Get(1)",
			expected:    true,
			description: "Comparison between Series values",
		},

		// Negative cases - not comparison operators
		{
			name:        "addition +",
			code:        "price + 10",
			expected:    false,
			description: "Arithmetic addition is not comparison",
		},
		{
			name:        "subtraction -",
			code:        "high - low",
			expected:    false,
			description: "Arithmetic subtraction is not comparison",
		},
		{
			name:        "multiplication *",
			code:        "price * 2",
			expected:    false,
			description: "Multiplication is not comparison",
		},
		{
			name:        "division /",
			code:        "total / count",
			expected:    false,
			description: "Division is not comparison",
		},
		{
			name:        "modulo %",
			code:        "value % 10",
			expected:    false,
			description: "Modulo is not comparison",
		},
		{
			name:        "assignment =",
			code:        "x = 5",
			expected:    false,
			description: "Assignment is not comparison (single =)",
		},
		{
			name:        "identifier only",
			code:        "enabled",
			expected:    false,
			description: "Plain identifier has no comparison",
		},
		{
			name:        "function call",
			code:        "ta.Sma(close, 20)",
			expected:    false,
			description: "Function call without comparison",
		},

		// Edge cases
		{
			name:        "empty string",
			code:        "",
			expected:    false,
			description: "Empty code has no operators",
		},
		{
			name:        "comparison in string",
			code:        `"price > 100"`,
			expected:    true,
			description: "Operator in string still matches (string search)",
		},
		{
			name:        "comparison symbol in identifier",
			code:        "var_gt_100",
			expected:    false,
			description: "Text 'gt' is not > operator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.Matches(tt.code)
			if result != tt.expected {
				t.Errorf("%s\ncode=%q\nexpected: %v\ngot:      %v",
					tt.description, tt.code, tt.expected, result)
			}
		})
	}
}

/* Tests pattern matcher behavior with boundary conditions */
func TestPatternMatcher_BoundaryAndEdgeCases(t *testing.T) {
	seriesMatcher := NewSeriesAccessPattern()
	comparisonMatcher := NewComparisonPattern()

	tests := []struct {
		name             string
		code             string
		expectSeries     bool
		expectComparison bool
		description      string
	}{
		{
			name:             "empty string",
			code:             "",
			expectSeries:     false,
			expectComparison: false,
			description:      "Empty code matches no patterns",
		},
		{
			name:             "whitespace only",
			code:             "   \t\n  ",
			expectSeries:     false,
			expectComparison: false,
			description:      "Whitespace-only code matches no patterns",
		},
		{
			name:             "very long code",
			code:             "ta.Ema(ta.Sma(ta.Wma(closeSeries.GetCurrent(), 5), 10), 20) > bbTopSeries.Get(1) && volumeSeries.GetCurrent() > avgVolumeSeries.Get(10)",
			expectSeries:     true,
			expectComparison: true,
			description:      "Long complex expression matches both patterns",
		},
		{
			name:             "Series only no comparison",
			code:             "priceSeries.GetCurrent()",
			expectSeries:     true,
			expectComparison: false,
			description:      "Series access without comparison",
		},
		{
			name:             "comparison only no Series",
			code:             "value > 100",
			expectSeries:     false,
			expectComparison: true,
			description:      "Comparison without Series access",
		},
		{
			name:             "both Series and comparison",
			code:             "priceSeries.GetCurrent() > 100",
			expectSeries:     true,
			expectComparison: true,
			description:      "Code with both patterns",
		},
		{
			name:             "neither pattern",
			code:             "bar.Close",
			expectSeries:     false,
			expectComparison: false,
			description:      "Simple member access with no patterns",
		},
		{
			name:             "special characters",
			code:             "$$priceSeries.GetCurrent()##",
			expectSeries:     true,
			expectComparison: false,
			description:      "Special characters don't prevent matching",
		},
		{
			name:             "unicode characters",
			code:             "価格Series.GetCurrent() > 100",
			expectSeries:     true,
			expectComparison: true,
			description:      "Unicode identifiers work with patterns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seriesResult := seriesMatcher.Matches(tt.code)
			comparisonResult := comparisonMatcher.Matches(tt.code)

			if seriesResult != tt.expectSeries {
				t.Errorf("%s\nSeries pattern: expected %v, got %v",
					tt.description, tt.expectSeries, seriesResult)
			}
			if comparisonResult != tt.expectComparison {
				t.Errorf("%s\nComparison pattern: expected %v, got %v",
					tt.description, tt.expectComparison, comparisonResult)
			}
		})
	}
}
