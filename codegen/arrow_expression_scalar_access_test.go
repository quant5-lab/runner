package codegen

import (
	"strings"
	"testing"
)

/*
Validates scalar access in arrow function expressions across all contexts.

ForwardSeriesBuffer paradigm: local variables resolve to scalars for current bar,
Series.Get(offset) for historical access. These tests ensure no .GetCurrent() leaks
into current bar computations, validating algorithm behavior not specific bugs.
*/

/* TestArrowExpressionScalarAccess_ConditionalExpressions validates ternary operators use scalars */
func TestArrowExpressionScalarAccess_ConditionalExpressions(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string // ALL patterns must exist
		forbiddenPattern []string // NONE of these patterns should exist
		description      string
	}{
		{
			name: "simple ternary with local variables",
			pine: `
//@version=5
indicator("Test")
calc(threshold) =>
    x = close + 10
    y = open - 5
    result = x > y ? x : y
    result
plot(calc(100))
`,
			mustContainAll: []string{
				"x := (bar.Close + 10)",
				"y := (bar.Open - 5)",
				"if (x > y)", // Ternary test uses scalar
				"return x",   // Ternary consequent uses scalar
				"return y",   // Ternary alternate uses scalar
			},
			forbiddenPattern: []string{
				"xSeries.GetCurrent()",
				"ySeries.GetCurrent()",
			},
			description: "ternary test, consequent, and alternate all use scalar variables",
		},
		{
			name: "nested ternary expressions",
			pine: `
//@version=5
indicator("Test")
select(val1, val2, val3) =>
    a = val1 * 2
    b = val2 * 2
    c = val3 * 2
    result = a > b ? (a > c ? a : c) : (b > c ? b : c)
    result
plot(select(10, 20, 30))
`,
			mustContainAll: []string{
				"a := (val1 * 2)",
				"b := (val2 * 2)",
				"c := (val3 * 2)",
				"if (a > b)", // Outer test
				"if (a > c)", // Inner test 1
				"return a",   // Multiple scalar returns
				"return c",
				"if (b > c)", // Inner test 2
				"return b",
			},
			forbiddenPattern: []string{
				"aSeries.GetCurrent()",
				"bSeries.GetCurrent()",
				"cSeries.GetCurrent()",
			},
			description: "nested ternaries maintain scalar access at all levels",
		},
		{
			name: "ternary with parameter and local variable",
			pine: `
//@version=5
indicator("Test")
compare(threshold) =>
    value = close * 1.1
    result = value > threshold ? value : threshold
    result
plot(compare(100))
`,
			mustContainAll: []string{
				"value := (bar.Close * 1.1)",
				"if (value > threshold)", // Both scalar
				"return value",           // Scalar return
				"return threshold",       // Parameter remains scalar
			},
			forbiddenPattern: []string{
				"valueSeries.GetCurrent()",
				"thresholdSeries",
			},
			description: "ternary with mixed local variable and parameter uses scalars",
		},
		{
			name: "ternary in TA function source",
			pine: `
//@version=5
indicator("Test")
smoothed(len) =>
    up_diff = high - high[1]
    down_diff = low[1] - low
    source = up_diff > down_diff ? up_diff : down_diff
    ta.sma(source, len)
plot(smoothed(14))
`,
			mustContainAll: []string{
				"up_diff :=",
				"down_diff :=",
				"if (up_diff > down_diff)",
				"return up_diff",
				"return down_diff",
			},
			forbiddenPattern: []string{
				"up_diffSeries.GetCurrent()",
				"down_diffSeries.GetCurrent()",
			},
			description: "ternary used as TA source maintains scalar resolution",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, required := range tt.mustContainAll {
				if !strings.Contains(code, required) {
					t.Errorf("%s: Missing required pattern:\n  %s\n\nGenerated code:\n%s",
						tt.description, required, truncateCode(code, 1000))
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(code, forbidden) {
					t.Errorf("%s: Found forbidden Series.GetCurrent() pattern:\n  %s\n\nGenerated code:\n%s",
						tt.description, forbidden, truncateCode(code, 1000))
				}
			}
		})
	}
}

/* TestArrowExpressionScalarAccess_LogicalExpressions validates and/or operators use scalars */
func TestArrowExpressionScalarAccess_LogicalExpressions(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "logical AND with local variables",
			pine: `
//@version=5
indicator("Test")
check(threshold) =>
    above_threshold = close > threshold
    below_high = close < high
    signal = above_threshold and below_high ? 1 : 0
    signal
plot(check(100))
`,
			mustContainAll: []string{
				"above_threshold := (bar.Close > threshold)",
				"below_high := (bar.Close < bar.High)",
				"(above_threshold && below_high)", // Logical AND uses scalars
			},
			forbiddenPattern: []string{
				"above_thresholdSeries.GetCurrent()",
				"below_highSeries.GetCurrent()",
			},
			description: "logical AND with local variables uses scalar access",
		},
		{
			name: "logical OR with nested comparisons",
			pine: `
//@version=5
indicator("Test")
validate(min_val, max_val) =>
    current = close + open
    too_low = current < min_val
    too_high = current > max_val
    invalid = too_low or too_high
    invalid
plot(validate(10, 100))
`,
			mustContainAll: []string{
				"current := (bar.Close + bar.Open)",
				"too_low := (current < min_val)",
				"too_high := (current > max_val)",
				"invalid := (too_low || too_high)", // Logical OR uses scalars
			},
			forbiddenPattern: []string{
				"currentSeries.GetCurrent()",
				"too_lowSeries.GetCurrent()",
				"too_highSeries.GetCurrent()",
			},
			description: "logical OR with local variables uses scalar access",
		},
		{
			name: "complex logical expression in ternary test",
			pine: `
//@version=5
indicator("Test")
select_value(threshold) =>
    up_move = high - low
    down_move = low - open
    condition = (up_move > down_move) and (up_move > threshold)
    result = condition ? up_move : 0
    result
plot(select_value(5))
`,
			mustContainAll: []string{
				"up_move := (bar.High - bar.Low)",
				"down_move := (bar.Low - bar.Open)",
				"condition := ((up_move > down_move) && (up_move > threshold))",
				"if condition",   // Ternary test uses scalar boolean
				"return up_move", // Scalar return
			},
			forbiddenPattern: []string{
				"up_moveSeries.GetCurrent()",
				"down_moveSeries.GetCurrent()",
				"conditionSeries.GetCurrent()",
			},
			description: "complex logical expression in ternary uses scalar variables",
		},
		{
			name: "logical expression in TA source - DMI pattern",
			pine: `
//@version=5
indicator("Test")
dmi_calc(len) =>
    up_diff = high - high[1]
    down_diff = low[1] - low
    up_valid = (up_diff > down_diff) and (up_diff > 0)
    up_value = up_valid ? up_diff : 0
    ta.rma(up_value, len)
plot(dmi_calc(14))
`,
			mustContainAll: []string{
				"up_diff :=",
				"down_diff :=",
				"((up_diff > down_diff) && (up_diff > 0))", // Logical AND in ternary test
				"if", // Ternary IIFE
				"return up_diff",
			},
			forbiddenPattern: []string{
				"up_diffSeries.GetCurrent()",
				"down_diffSeries.GetCurrent()",
			},
			description: "DMI-style logical expression in TA source uses scalar access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, required := range tt.mustContainAll {
				if !strings.Contains(code, required) {
					t.Errorf("%s: Missing required pattern:\n  %s\n\nGenerated code:\n%s",
						tt.description, required, truncateCode(code, 1000))
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(code, forbidden) {
					t.Errorf("%s: Found forbidden Series.GetCurrent() pattern:\n  %s\n\nGenerated code:\n%s",
						tt.description, forbidden, truncateCode(code, 1000))
				}
			}
		})
	}
}

/* TestArrowExpressionScalarAccess_TernarySources validates ternary expressions as TA sources */
func TestArrowExpressionScalarAccess_TernarySources(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "inline ternary as RMA source",
			pine: `
//@version=5
indicator("Test")
calc(len) =>
    up = high - low
    down = low - open
    source = up > down ? up : down
    ta.rma(source, len)
plot(calc(14))
`,
			mustContainAll: []string{
				"up := (bar.High - bar.Low)",
				"down := (bar.Low - bar.Open)",
				"source := func() float64 { if (up > down)", // Variable assigned to IIFE
				"return up",
				"return down",
			},
			forbiddenPattern: []string{
				"upSeries.GetCurrent()",
				"downSeries.GetCurrent()",
			},
			description: "ternary as TA source generates IIFE with scalar access",
		},
		{
			name: "ternary with zero default - ADX pattern",
			pine: `
//@version=5
indicator("Test")
plus_di(len) =>
    up = high - high[1]
    down = low[1] - low
    plus = (up > down) and (up > 0) ? up : 0
    ta.rma(plus, len)
plot(plus_di(14))
`,
			mustContainAll: []string{
				"up :=",
				"down :=",
				"((up > down) && (up > 0))", // Logical expression in test
				"return up",
				"return 0",
			},
			forbiddenPattern: []string{
				"upSeries.GetCurrent()",
				"downSeries.GetCurrent()",
			},
			description: "ADX-style ternary with zero uses scalar variables",
		},
		{
			name: "multiple ternary sources in sequence",
			pine: `
//@version=5
indicator("Test")
dual_smooth(len) =>
    trend_up = close > open ? close - open : 0
    trend_down = open > close ? open - close : 0
    smooth_up = ta.sma(trend_up, len)
    smooth_down = ta.sma(trend_down, len)
    smooth_up - smooth_down
plot(dual_smooth(10))
`,
			mustContainAll: []string{
				"if (bar.Close > bar.Open)",
				"return (bar.Close - bar.Open)",
				"if (bar.Open > bar.Close)",
				"return (bar.Open - bar.Close)",
			},
			forbiddenPattern: []string{
				"trend_upSeries.GetCurrent()",
				"trend_downSeries.GetCurrent()",
			},
			description: "multiple ternary sources each use scalar resolution",
		},
		{
			name: "nested ternary as TA source",
			pine: `
//@version=5
indicator("Test")
adaptive(len, threshold) =>
    range_val = high - low
    volatility = range_val > threshold ? 
        (range_val > threshold * 2 ? range_val * 1.5 : range_val) : 
        threshold
    ta.ema(volatility, len)
plot(adaptive(14, 10))
`,
			mustContainAll: []string{
				"range_val := (bar.High - bar.Low)",
				"if (range_val > threshold)",
				"if (range_val > (threshold * 2))",
				"return (range_val * 1.5)",
				"return range_val",
				"return threshold",
			},
			forbiddenPattern: []string{
				"range_valSeries.GetCurrent()",
				"volatilitySeries.GetCurrent()",
			},
			description: "nested ternary as TA source uses scalar at all levels",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, required := range tt.mustContainAll {
				if !strings.Contains(code, required) {
					t.Errorf("%s: Missing required pattern:\n  %s\n\nGenerated code:\n%s",
						tt.description, required, truncateCode(code, 1000))
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(code, forbidden) {
					t.Errorf("%s: Found forbidden Series.GetCurrent() pattern:\n  %s\n\nGenerated code:\n%s",
						tt.description, forbidden, truncateCode(code, 1000))
				}
			}
		})
	}
}

/* TestArrowExpressionScalarAccess_BinaryExpressionSources validates arithmetic in TA sources */
func TestArrowExpressionScalarAccess_BinaryExpressionSources(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "binary expression as TA source",
			pine: `
//@version=5
indicator("Test")
momentum(len) =>
    diff = close - open
    scaled = diff * 100
    ta.sma(scaled, len)
plot(momentum(14))
`,
			mustContainAll: []string{
				"diff := (bar.Close - bar.Open)",
				"scaled := (diff * 100)",
				"scaledSeries.Get(j)", // TA loop uses Series for historical access
			},
			forbiddenPattern: []string{
				"diffSeries.GetCurrent()",
				"scaledSeries.GetCurrent()",
			},
			description: "binary expression as TA source uses scalar variable, Series for historical",
		},
		{
			name: "complex binary with division - ADX pattern",
			pine: `
//@version=5
indicator("Test")
adx_calc(len) =>
    plus_dm = high - high[1]
    minus_dm = low[1] - low
    tr_val = high - low
    plus_di = 100 * ta.rma(plus_dm, len) / ta.rma(tr_val, len)
    plus_di
plot(adx_calc(14))
`,
			mustContainAll: []string{
				"plus_dm :=",
				"minus_dm :=",
				"tr_val :=",
			},
			forbiddenPattern: []string{
				"plus_dmSeries.GetCurrent()",
				"minus_dmSeries.GetCurrent()",
				"tr_valSeries.GetCurrent()",
			},
			description: "ADX-style calculations use scalar variables throughout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, required := range tt.mustContainAll {
				if !strings.Contains(code, required) {
					t.Errorf("%s: Missing required pattern:\n  %s\n\nGenerated code:\n%s",
						tt.description, required, truncateCode(code, 1000))
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(code, forbidden) {
					t.Errorf("%s: Found forbidden Series.GetCurrent() pattern:\n  %s\n\nGenerated code:\n%s",
						tt.description, forbidden, truncateCode(code, 1000))
				}
			}
		})
	}
}

/* TestArrowExpressionScalarAccess_EdgeCases validates boundary conditions */
func TestArrowExpressionScalarAccess_EdgeCases(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "triple nested ternary",
			pine: `
//@version=5
indicator("Test")
select(a, b, c, d) =>
    x = a * 2
    y = b * 2
    z = c * 2
    w = d * 2
    result = x > y ? (x > z ? (x > w ? x : w) : (z > w ? z : w)) : (y > z ? (y > w ? y : w) : (z > w ? z : w))
    result
plot(select(1, 2, 3, 4))
`,
			mustContainAll: []string{
				"x := (a * 2)",
				"y := (b * 2)",
				"z := (c * 2)",
				"w := (d * 2)",
				"if (x > y)",
				"return x",
				"return y",
				"return z",
				"return w",
			},
			forbiddenPattern: []string{
				"xSeries.GetCurrent()",
				"ySeries.GetCurrent()",
				"zSeries.GetCurrent()",
				"wSeries.GetCurrent()",
			},
			description: "deeply nested ternary maintains scalar access at all depths",
		},
		{
			name: "ternary with chained logical operators",
			pine: `
//@version=5
indicator("Test")
complex_condition(threshold) =>
    a = close > threshold
    b = high > open
    c = low < close
    signal = (a and b) or c ? 1 : 0
    signal
plot(complex_condition(100))
`,
			mustContainAll: []string{
				"a := (bar.Close > threshold)",
				"b := (bar.High > bar.Open)",
				"c := (bar.Low < bar.Close)",
				"((a && b) || c)", // Chained logical with scalars
			},
			forbiddenPattern: []string{
				"aSeries.GetCurrent()",
				"bSeries.GetCurrent()",
				"cSeries.GetCurrent()",
			},
			description: "chained logical operators in ternary use scalar variables",
		},
		{
			name: "single-character variable names",
			pine: `
//@version=5
indicator("Test")
calc(p) =>
    x = close * p
    y = x > 100 ? x : 0
    ta.sma(y, 10)
plot(calc(2))
`,
			mustContainAll: []string{
				"x := (bar.Close * p)",
				"if (x > 100)",
				"return x",
			},
			forbiddenPattern: []string{
				"xSeries.GetCurrent()",
				"pSeries",
			},
			description: "single-character variables use scalar access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, required := range tt.mustContainAll {
				if !strings.Contains(code, required) {
					t.Errorf("%s: Missing required pattern:\n  %s\n\nGenerated code:\n%s",
						tt.description, required, truncateCode(code, 1000))
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(code, forbidden) {
					t.Errorf("%s: Found forbidden Series.GetCurrent() pattern:\n  %s\n\nGenerated code:\n%s",
						tt.description, forbidden, truncateCode(code, 1000))
				}
			}
		})
	}
}

/* truncateCode limits code output for readable error messages */
func truncateCode(code string, maxLen int) string {
	if len(code) <= maxLen {
		return code
	}
	return code[:maxLen] + "\n... (truncated)"
}
