package codegen

import (
	"strings"
	"testing"
)

/* TestArrowFunctionForLoopReturn validates return statement uses Series.GetCurrent() when variable modified in loop */
func TestArrowFunctionForLoopReturn(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		expectedReturn   string
		forbiddenPattern string
		description      string
	}{
		{
			name: "for-loop accumulator returns series",
			pine: `
//@version=5
indicator("Test")
countBullish(len) =>
    count = 0
    for i = 0 to len - 1
        count := count + 1
    count
plot(countBullish(10))
`,
			expectedReturn:   "return countSeries.GetCurrent()",
			forbiddenPattern: "return count\n",
			description:      "for-loop modified variable returns from series",
		},
		{
			name: "for-loop conditional accumulator",
			pine: `
//@version=5
indicator("Test")
countIf(len) =>
    count = 0
    for i = 0 to len - 1
        if close[i] > open[i]
            count := count + 1
    count
plot(countIf(20))
`,
			expectedReturn:   "return countSeries.GetCurrent()",
			forbiddenPattern: "return count\n",
			description:      "conditional accumulation in loop returns from series",
		},
		{
			name: "no loop modification returns scalar",
			pine: `
//@version=5
indicator("Test")
compute(len) =>
    result = close * len
    result
plot(compute(10))
`,
			expectedReturn:   "return result\n",
			forbiddenPattern: "return resultSeries.GetCurrent()",
			description:      "non-loop variable returns scalar (optimization)",
		},
		{
			name: "sum accumulator in loop",
			pine: `
//@version=5
indicator("Test")
sumValues(len) =>
    sum = 0.0
    for i = 0 to len - 1
        sum := sum + close[i]
    sum
plot(sumValues(5))
`,
			expectedReturn:   "return sumSeries.GetCurrent()",
			forbiddenPattern: "return sum\n",
			description:      "sum accumulator returns from series",
		},
		{
			name: "multiple variables one modified",
			pine: `
//@version=5
indicator("Test")
calc(len) =>
    multiplier = 2.0
    count = 0
    for i = 0 to len - 1
        count := count + 1
    count * multiplier
plot(calc(10))
`,
			expectedReturn:   "countSeries.GetCurrent()",
			forbiddenPattern: "",
			description:      "loop-modified variable uses series in expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			if !strings.Contains(code, tt.expectedReturn) {
				t.Errorf("%s: Missing expected return pattern:\n  %q\n\nGenerated:\n%s",
					tt.description, tt.expectedReturn, code)
			}

			if tt.forbiddenPattern != "" && strings.Contains(code, tt.forbiddenPattern) {
				t.Errorf("%s: Found forbidden pattern:\n  %q\n\nGenerated:\n%s",
					tt.description, tt.forbiddenPattern, code)
			}
		})
	}
}
