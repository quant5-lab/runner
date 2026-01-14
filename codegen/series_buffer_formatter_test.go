package codegen

import (
	"strings"
	"testing"
)

/* TestSeriesBufferFormatter_TopLevelAccess validates top-level series buffer access formatting
 *
 * Tests formatSeriesGet and formatSeriesSet generate correct code for top-level scope:
 * - Pattern: {varName}Series.Get({offset})
 * - Pattern: {varName}Series.Set({value})
 *
 * Generalized test for any series buffer access in main execution scope
 */
func TestSeriesBufferFormatter_TopLevelAccess(t *testing.T) {
	testCases := []struct {
		name        string
		varName     string
		offset      int
		value       string
		expectedGet string
		expectedSet string
	}{
		{
			name:        "Simple variable, offset 0",
			varName:     "rma14",
			offset:      0,
			value:       "newValue",
			expectedGet: "rma14Series.Get(0)",
			expectedSet: "rma14Series.Set(newValue)",
		},
		{
			name:        "Simple variable, offset 1 (previous)",
			varName:     "ema20",
			offset:      1,
			value:       "result",
			expectedGet: "ema20Series.Get(1)",
			expectedSet: "ema20Series.Set(result)",
		},
		{
			name:        "Variable with underscores, large offset",
			varName:     "my_indicator",
			offset:      50,
			value:       "calculated",
			expectedGet: "my_indicatorSeries.Get(50)",
			expectedSet: "my_indicatorSeries.Set(calculated)",
		},
		{
			name:        "Short variable name",
			varName:     "x",
			offset:      5,
			value:       "val",
			expectedGet: "xSeries.Get(5)",
			expectedSet: "xSeries.Set(val)",
		},
		{
			name:        "Complex value expression",
			varName:     "sma",
			offset:      0,
			value:       "sum / float64(period)",
			expectedGet: "smaSeries.Get(0)",
			expectedSet: "smaSeries.Set(sum / float64(period))",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatSeriesGet(tc.varName, tc.offset)
			if got != tc.expectedGet {
				t.Errorf("formatSeriesGet mismatch\nExpected: %s\nGot:      %s", tc.expectedGet, got)
			}

			got = formatSeriesSet(tc.varName, tc.value)
			if got != tc.expectedSet {
				t.Errorf("formatSeriesSet mismatch\nExpected: %s\nGot:      %s", tc.expectedSet, got)
			}
		})
	}
}

/* TestSeriesBufferFormatter_ArrowAccess validates arrow function series buffer access formatting
 *
 * Tests formatArrowSeriesGet and formatArrowSeriesSet generate correct code for arrow scope:
 * - Pattern: arrowCtx.GetOrCreateSeries("{varName}").Get({offset})
 * - Pattern: arrowCtx.GetOrCreateSeries("{varName}").Set({value})
 *
 * Generalized test for any series buffer access within arrow functions
 */
func TestSeriesBufferFormatter_ArrowAccess(t *testing.T) {
	testCases := []struct {
		name        string
		varName     string
		offset      int
		value       string
		expectedGet string
		expectedSet string
	}{
		{
			name:        "Simple variable, offset 0",
			varName:     "truerange",
			offset:      0,
			value:       "tr",
			expectedGet: "arrowCtx.GetOrCreateSeries(\"truerange\").Get(0)",
			expectedSet: "arrowCtx.GetOrCreateSeries(\"truerange\").Set(tr)",
		},
		{
			name:        "Simple variable, offset 1 (previous)",
			varName:     "plus",
			offset:      1,
			value:       "newPlus",
			expectedGet: "arrowCtx.GetOrCreateSeries(\"plus\").Get(1)",
			expectedSet: "arrowCtx.GetOrCreateSeries(\"plus\").Set(newPlus)",
		},
		{
			name:        "Variable with underscores, large offset",
			varName:     "adx_smoothed",
			offset:      100,
			value:       "smoothedValue",
			expectedGet: "arrowCtx.GetOrCreateSeries(\"adx_smoothed\").Get(100)",
			expectedSet: "arrowCtx.GetOrCreateSeries(\"adx_smoothed\").Set(smoothedValue)",
		},
		{
			name:        "Short variable name",
			varName:     "dx",
			offset:      10,
			value:       "directional",
			expectedGet: "arrowCtx.GetOrCreateSeries(\"dx\").Get(10)",
			expectedSet: "arrowCtx.GetOrCreateSeries(\"dx\").Set(directional)",
		},
		{
			name:        "Complex value expression",
			varName:     "ratio",
			offset:      0,
			value:       "math.Abs(plus - minus) / sum",
			expectedGet: "arrowCtx.GetOrCreateSeries(\"ratio\").Get(0)",
			expectedSet: "arrowCtx.GetOrCreateSeries(\"ratio\").Set(math.Abs(plus - minus) / sum)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatArrowSeriesGet(tc.varName, tc.offset)
			if got != tc.expectedGet {
				t.Errorf("formatArrowSeriesGet mismatch\nExpected: %s\nGot:      %s", tc.expectedGet, got)
			}

			got = formatArrowSeriesSet(tc.varName, tc.value)
			if got != tc.expectedSet {
				t.Errorf("formatArrowSeriesSet mismatch\nExpected: %s\nGot:      %s", tc.expectedSet, got)
			}
		})
	}
}

/* TestSeriesBufferFormatter_QuotingConsistency validates string escaping in arrow context
 *
 * Tests that variable names are properly quoted in arrowCtx.GetOrCreateSeries() calls:
 * - Double quotes around variable name
 * - Proper escaping if variable name contains special characters
 *
 * Ensures generated code is syntactically valid Go
 */
func TestSeriesBufferFormatter_QuotingConsistency(t *testing.T) {
	testCases := []struct {
		name          string
		varName       string
		shouldContain []string
	}{
		{
			name:          "Simple alphanumeric variable",
			varName:       "myvar",
			shouldContain: []string{"\"myvar\"", "arrowCtx.GetOrCreateSeries(\"myvar\")"},
		},
		{
			name:          "Variable with underscores",
			varName:       "my_var_name",
			shouldContain: []string{"\"my_var_name\"", "arrowCtx.GetOrCreateSeries(\"my_var_name\")"},
		},
		{
			name:          "Variable with numbers",
			varName:       "var123",
			shouldContain: []string{"\"var123\"", "arrowCtx.GetOrCreateSeries(\"var123\")"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatArrowSeriesGet(tc.varName, 0)
			for _, expected := range tc.shouldContain {
				if !strings.Contains(got, expected) {
					t.Errorf("Missing expected substring\nExpected substring: %s\nGot:                %s", expected, got)
				}
			}

			got = formatArrowSeriesSet(tc.varName, "value")
			for _, expected := range tc.shouldContain {
				if !strings.Contains(got, expected) {
					t.Errorf("Missing expected substring\nExpected substring: %s\nGot:                %s", expected, got)
				}
			}
		})
	}
}
