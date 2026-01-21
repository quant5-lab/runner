package codegen

import (
	"strings"
	"testing"
)

/*
Validates for-loop code generation in arrow functions across all patterns.

Tests ensure proper bounds expression generation, subscript access, parameter resolution,
and loop-modified variable tracking. Generalized for algorithmic behavior validation.
*/

/* TestArrowForLoopBoundsGeneration validates for-loop bound expressions use correct context */
func TestArrowForLoopBoundsGeneration(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "parameter in loop bounds scalar access",
			pine: `
//@version=5
indicator("Test")
sum(len) =>
    result = 0.0
    for i = 0 to len - 1
        result := result + 1
    result
plot(sum(10))
`,
			mustContainAll: []string{
				"_to := int((len - 1))", // Parameter uses scalar, not lenSeries.GetCurrent()
				"resultSeries.Set(",
			},
			forbiddenPattern: []string{
				"lenSeries.GetCurrent()",
				"lenSeries.Get(",
			},
			description: "parameter in bounds expression accesses scalar value",
		},
		{
			name: "local variable in loop bounds",
			pine: `
//@version=5
indicator("Test")
iterate(count) =>
    limit = count * 2
    sum = 0.0
    for i = 0 to limit - 1
        sum := sum + 1
    sum
plot(iterate(5))
`,
			mustContainAll: []string{
				"limit := (count * 2)",
				"limitSeries.Set(limit)",
				"_to := int((limit - 1))", // Local var uses scalar
			},
			forbiddenPattern: []string{
				"limitSeries.GetCurrent()",
			},
			description: "local variable in bounds uses scalar from current bar",
		},
		{
			name: "literal bounds",
			pine: `
//@version=5
indicator("Test")
fixed() =>
    sum = 0.0
    for i = 0 to 9
        sum := sum + 1
    sum
plot(fixed())
`,
			mustContainAll: []string{
				"i := int(0)",
				"_to := int(9)",
			},
			forbiddenPattern: nil,
			description:      "literal bounds generate correctly",
		},
		{
			name: "complex expression in bounds",
			pine: `
//@version=5
indicator("Test")
calc(a, b) =>
    result = 0.0
    for i = 0 to (a + b) * 2 - 1
        result := result + 1
    result
plot(calc(3, 4))
`,
			mustContainAll: []string{
				"_to := int((((a + b) * 2) - 1))",
			},
			forbiddenPattern: []string{
				"aSeries.GetCurrent()",
				"bSeries.GetCurrent()",
			},
			description: "complex bounds expression evaluates parameters as scalars",
		},
		{
			name: "step expression with parameter",
			pine: `
//@version=5
indicator("Test")
stride(len, step) =>
    sum = 0.0
    for i = 0 to len by step
        sum := sum + 1
    sum
plot(stride(20, 2))
`,
			mustContainAll: []string{
				"_to := int(len)",
				"_step := int(step)",
			},
			forbiddenPattern: []string{
				"stepSeries.GetCurrent()",
			},
			description: "step expression uses parameter as scalar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goCode, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, pattern := range tt.mustContainAll {
				if !strings.Contains(goCode, pattern) {
					t.Errorf("Missing required pattern: %q\nDescription: %s\nGenerated code:\n%s",
						pattern, tt.description, goCode)
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(goCode, forbidden) {
					t.Errorf("Found forbidden pattern: %q\nDescription: %s\nGenerated code:\n%s",
						forbidden, tt.description, goCode)
				}
			}
		})
	}
}

/* TestArrowForLoopSubscriptAccess validates subscript expressions in loop bodies */
func TestArrowForLoopSubscriptAccess(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "builtin series subscript with loop counter",
			pine: `
//@version=5
indicator("Test")
sum(len) =>
    total = 0.0
    for i = 0 to len - 1
        total := total + close[i]
    total
plot(sum(5))
`,
			mustContainAll: []string{
				"barIdx := ctx.BarIndex-float64(i)", // Loop counter cast to float64
				"ctx.Data[barIdx].Close",            // Proper builtin access
				"totalSeries.Set(",                  // Loop-modified variable uses series
			},
			forbiddenPattern: []string{
				"bar.Close[i]",       // Should not use bar struct subscript
				"closeSeries.Get(i)", // Should not use Series.Get for builtins in arrow
			},
			description: "loop counter subscript accesses builtin via ctx.Data",
		},
		{
			name: "nested loop with different counters",
			pine: `
//@version=5
indicator("Test")
nested(len) =>
    outer = 0.0
    for i = 0 to len - 1
        inner = 0.0
        for j = 0 to 2
            inner := inner + close[j]
        outer := outer + inner
    outer
plot(nested(3))
`,
			mustContainAll: []string{
				"barIdx := ctx.BarIndex-float64(j)", // Inner loop counter cast to float64
				"ctx.Data[barIdx].Close",
				"outerSeries.Set(",
			},
			forbiddenPattern: []string{
				"closeSeries.Get(j)",
			},
			description: "nested loops maintain proper subscript resolution per counter",
		},
		{
			name: "multiple builtin subscripts in same expression",
			pine: `
//@version=5
indicator("Test")
range(len) =>
    maxRange = 0.0
    for i = 0 to len - 1
        r = high[i] - low[i]
        maxRange := r > maxRange ? r : maxRange
    maxRange
plot(range(10))
`,
			mustContainAll: []string{
				"ctx.Data[barIdx].High",
				"ctx.Data[barIdx].Low",
			},
			forbiddenPattern: []string{
				"highSeries.Get(",
				"lowSeries.Get(",
			},
			description: "multiple builtin subscripts in expression resolve correctly",
		},
		{
			name: "literal subscript constant offset",
			pine: `
//@version=5
indicator("Test")
compare() =>
    prev = close[1]
    curr = close[0]
    curr - prev
plot(compare())
`,
			mustContainAll: []string{
				"barIdx := ctx.BarIndex-1", // Arrow functions use ctx.BarIndex for subscript
				"barIdx := ctx.BarIndex-0", // close[0] still uses IIFE pattern in arrow context
				"ctx.Data[barIdx].Close",   // Both subscripts access ctx.Data
			},
			forbiddenPattern: nil,
			description:      "literal subscripts in arrow functions resolve via ctx.BarIndex",
		},
		{
			name: "subscript in conditional within loop",
			pine: `
//@version=5
indicator("Test")
countBullish(len) =>
    count = 0.0
    for i = 0 to len - 1
        if close[i] > open[i]
            count := count + 1
    count
plot(countBullish(20))
`,
			mustContainAll: []string{
				"ctx.Data[barIdx].Close",
				"ctx.Data[barIdx].Open",
				"if (func() float64", // Subscript in conditional generates IIFE
			},
			forbiddenPattern: []string{
				"closeSeries.Get(",
				"openSeries.Get(",
			},
			description: "subscript in loop conditional resolves correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goCode, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, pattern := range tt.mustContainAll {
				if !strings.Contains(goCode, pattern) {
					t.Errorf("Missing required pattern: %q\nDescription: %s\nGenerated code:\n%s",
						pattern, tt.description, goCode)
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(goCode, forbidden) {
					t.Errorf("Found forbidden pattern: %q\nDescription: %s\nGenerated code:\n%s",
						forbidden, tt.description, goCode)
				}
			}
		})
	}
}

/* TestArrowForLoopVariableTypes validates proper type handling in loop contexts */
func TestArrowForLoopVariableTypes(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "integer literal initialized variable",
			pine: `
//@version=5
indicator("Test")
count(len) =>
    total = 0
    for i = 0 to len - 1
        total := total + 1
    total
plot(count(10))
`,
			mustContainAll: []string{
				"total := float64(0)", // Integer literal converted to float64
				"totalSeries.Set(total)",
			},
			forbiddenPattern: []string{
				"total := 0\n", // Raw integer should not exist
			},
			description: "integer literal 0 converted to float64 for Series compatibility",
		},
		{
			name: "float literal no conversion",
			pine: `
//@version=5
indicator("Test")
calc(len) =>
    sum = 0.0
    for i = 0 to len - 1
        sum := sum + 1.0
    sum
plot(calc(5))
`,
			mustContainAll: []string{
				"sum := float64(0)", // isSimpleInteger converts numeric literals to float64
			},
			forbiddenPattern: nil,
			description:      "float literal converted by isSimpleInteger helper",
		},
		{
			name: "negative integer literal",
			pine: `
//@version=5
indicator("Test")
adjust(len) =>
    offset = -5
    result = 0.0
    for i = 0 to len - 1
        result := result + 1
    result + offset
plot(adjust(10))
`,
			mustContainAll: []string{
				"offset := float64(-5)", // Negative integer converted
			},
			forbiddenPattern: []string{
				"offset := -5\n",
			},
			description: "negative integer literals converted to float64",
		},
		{
			name: "non-literal expression no conversion",
			pine: `
//@version=5
indicator("Test")
calc(multiplier) =>
    value = close * multiplier
    sum = 0.0
    for i = 0 to 5
        sum := sum + value
    sum
plot(calc(2))
`,
			mustContainAll: []string{
				"value := (bar.Close * multiplier)", // Expression not wrapped
			},
			forbiddenPattern: []string{
				"float64((bar.Close * multiplier))", // Should not double-wrap
			},
			description: "non-literal expressions not wrapped in float64()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goCode, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, pattern := range tt.mustContainAll {
				if !strings.Contains(goCode, pattern) {
					t.Errorf("Missing required pattern: %q\nDescription: %s\nGenerated code:\n%s",
						pattern, tt.description, goCode)
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(goCode, forbidden) {
					t.Errorf("Found forbidden pattern: %q\nDescription: %s\nGenerated code:\n%s",
						forbidden, tt.description, goCode)
				}
			}
		})
	}
}

/* TestArrowForLoopModifiedVariableTracking validates loop-modified variables use Series */
func TestArrowForLoopModifiedVariableTracking(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "reassigned accumulator",
			pine: `
//@version=5
indicator("Test")
accumulate(len) =>
    count = 0.0
    for i = 0 to len - 1
        count := count + 1
    count
plot(accumulate(10))
`,
			mustContainAll: []string{
				"count = (countSeries.GetCurrent() + 1)", // Reassignment uses =
				"countSeries.Set(count)",
				"return countSeries.GetCurrent()", // Returns from series
			},
			forbiddenPattern: []string{
				"count := (count + 1)", // Should NOT use := for reassignment
			},
			description: "loop-modified variable uses Series.GetCurrent() for reads",
		},
		{
			name: "multiple modified variables",
			pine: `
//@version=5
indicator("Test")
tally(len) =>
    ups = 0.0
    downs = 0.0
    for i = 0 to len - 1
        if close[i] > open[i]
            ups := ups + 1
        else
            downs := downs + 1
    ups - downs
plot(tally(20))
`,
			mustContainAll: []string{
				"upsSeries.Set((upsSeries.GetCurrent() + 1))", // Inline pattern in if-body
				"downs = (downsSeries.GetCurrent() + 1)",      // Reassignment uses =
				"downsSeries.Set(downs)",
			},
			forbiddenPattern: []string{
				"ups := (ups + 1)",   // Should NOT use := for reassignment
				"downs := (downs + ", // Should NOT use := for reassignment
			},
			description: "multiple loop-modified variables tracked independently",
		},
		{
			name: "unmodified variable stays scalar",
			pine: `
//@version=5
indicator("Test")
mixed(len) =>
    multiplier = 2.0
    sum = 0.0
    for i = 0 to len - 1
        sum := sum + 1
    sum * multiplier
plot(mixed(10))
`,
			mustContainAll: []string{
				"sumSeries.GetCurrent()", // Modified variable uses series
				"* multiplier)",          // Unmodified uses scalar
			},
			forbiddenPattern: []string{
				"multiplierSeries.GetCurrent()", // Should not use series
			},
			description: "unmodified variables maintain scalar access in expressions",
		},
		{
			name: "nested loop variable tracking",
			pine: `
//@version=5
indicator("Test")
nested(len) =>
    outer = 0.0
    for i = 0 to len - 1
        inner = 0.0
        for j = 0 to 3
            inner := inner + 1
        outer := outer + inner
    outer
plot(nested(5))
`,
			mustContainAll: []string{
				"innerSeries := arrowCtx.GetOrCreateSeries(\"inner\")", // Inner var gets Series storage
				"inner := float64(0)",                    // Scalar declaration
				"innerSeries.Set(inner)",                 // Series storage
				"inner = (innerSeries.GetCurrent() + 1)", // Reassignment uses =
				"innerSeries.Set(inner)",                 // Update series
				"outer = (outerSeries.GetCurrent()",      // Reassignment uses =
				"outerSeries.Set(outer)",
			},
			forbiddenPattern: []string{
				"inner := (inner + 1)", // Should NOT use := for reassignment
			},
			description: "nested loop variables use Series when modified in inner loops",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goCode, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, pattern := range tt.mustContainAll {
				if !strings.Contains(goCode, pattern) {
					t.Errorf("Missing required pattern: %q\nDescription: %s\nGenerated code:\n%s",
						pattern, tt.description, goCode)
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(goCode, forbidden) {
					t.Errorf("Found forbidden pattern: %q\nDescription: %s\nGenerated code:\n%s",
						forbidden, tt.description, goCode)
				}
			}
		})
	}
}
