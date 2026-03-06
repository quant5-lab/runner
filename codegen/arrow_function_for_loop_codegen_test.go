package codegen

import (
	"strings"
	"testing"
)

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
				"_to := int((len - 1))",
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
				"_to := int((limit - 1))",
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
				"ctx.BarIndex-int(float64(i))",
				"ctx.Data[barIdx].Close",
				"totalSeries.Set(",
			},
			forbiddenPattern: []string{
				"bar.Close[i]",
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
				"ctx.BarIndex-int(float64(j))",
				"ctx.Data[barIdx].Close",
				"outerSeries.Set(",
			},
			forbiddenPattern: []string{
				"closeSeries.Get(",
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
				"ctx.BarIndex-int(1)",
				"ctx.BarIndex-int(0)",
				"ctx.Data[barIdx].Close",
			},
			forbiddenPattern: []string{
				"closeSeries.Get(",
			},
			description: "literal subscripts in arrow functions resolve via ctx.Data with type-safe index",
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
				"if (func() float64",
			},
			forbiddenPattern: []string{
				"closeSeries.Get(",
				"openSeries.Get(",
			},
			description: "subscript in loop conditional resolves correctly",
		},
		{
			name: "all five builtin categories in same loop body",
			pine: `
//@version=5
indicator("Test")
mixed(len) =>
    total = 0.0
    for i = 0 to len - 1
        total := total + close[i] + time[i] + dayofweek[i] + bar_index[i] + hl2[i]
    total
plot(mixed(5))
`,
			mustContainAll: []string{
				"ctx.Data[barIdx].Close",
				"float64(ctx.Data[barIdx].Time * 1000)",
				"time.LoadLocation(ctx.Timezone)",
				"Weekday()",
				"float64(barIdx)",
				"ctx.Data[barIdx].High",
				"ctx.Data[barIdx].Low",
			},
			forbiddenPattern: []string{
				"closeSeries.Get(",
				"timeSeries.Get(",
				"dayofweekSeries.Get(",
				"bar_indexSeries.Get(",
				"hl2Series.Get(",
			},
			description: "ohlcv + time + calendar + bar_index + derived categories each route via ctx.Data IIFE",
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
				"total := float64(0)",
				"totalSeries.Set(total)",
			},
			forbiddenPattern: []string{
				"total := 0\n",
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
				"sum := float64(0)",
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
				"offset := float64(-5)",
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
				"value := (ctx.Data[ctx.BarIndex].Close * multiplier)",
			},
			forbiddenPattern: []string{
				"float64((ctx.Data[ctx.BarIndex].Close * multiplier))", // Should not double-wrap
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
				"count = (countSeries.GetCurrent() + 1)",
				"countSeries.Set(count)",
				"return countSeries.GetCurrent()",
			},
			forbiddenPattern: []string{
				"count := (count + 1)", // := reserved for initial declaration, not reassignment
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
				"upsSeries.Set((upsSeries.GetCurrent() + 1))",
				"downs = (downsSeries.GetCurrent() + 1)",
				"downsSeries.Set(downs)",
			},
			forbiddenPattern: []string{
				"ups := (ups + 1)",
				"downs := (downs + ",
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
				"sumSeries.GetCurrent()",
				"* multiplier)",
			},
			forbiddenPattern: []string{
				"multiplierSeries.GetCurrent()",
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
				"innerSeries := arrowCtx.GetOrCreateSeries(\"inner\")",
				"inner := float64(0)",
				"innerSeries.Set(inner)",
				"inner = (innerSeries.GetCurrent() + 1)",
				"innerSeries.Set(inner)",
				"outer = (outerSeries.GetCurrent()",
				"outerSeries.Set(outer)",
			},
			forbiddenPattern: []string{
				"inner := (inner + 1)",
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

/*
TestArrowForLoopCounterArithmetic validates that loop counter variables (Go int) are correctly
cast to float64 when used in arithmetic expressions, math calls, and conditional expressions
inside arrow-function for-loops. This is necessary because Pine counters are logically numeric
but the generated Go for-loop counter is typed int.
*/
func TestArrowForLoopCounterArithmetic(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "counter in math.abs call is float64-cast",
			pine: `
//@version=5
indicator("Test")
absSum(len) =>
    total = 0.0
    for i = 0 to len - 1
        total := total + math.abs(i)
    total
plot(absSum(5))
`,
			mustContainAll: []string{
				"math.Abs(float64(i))",
			},
			forbiddenPattern: []string{
				"iSeries",
				"math.Abs(i)",
			},
			description: "loop counter passed to math.abs() receives float64() cast",
		},
		{
			name: "counter in math.sin call is float64-cast",
			pine: `
//@version=5
indicator("Test")
sinSum(len) =>
    total = 0.0
    for i = 0 to len - 1
        total := total + math.sin(i)
    total
plot(sinSum(5))
`,
			mustContainAll: []string{
				"math.Sin(float64(i))",
			},
			forbiddenPattern: []string{
				"iSeries",
			},
			description: "loop counter passed to math.sin() receives float64() cast",
		},
		{
			name: "counter multiplied with parameter is float64-cast",
			pine: `
//@version=5
indicator("Test")
weightedSum(src, len) =>
    total = 0.0
    for i = 0 to len - 1
        total := total + src * i
    total
plot(weightedSum(close, 5))
`,
			mustContainAll: []string{
				"float64(i)",
				"src * float64(i)",
			},
			forbiddenPattern: []string{
				"iSeries",
			},
			description: "loop counter in binary expression with scalar parameter is float64-cast",
		},
		{
			name: "nested loop inner counter float64-cast, outer counter float64-cast independently",
			pine: `
//@version=5
indicator("Test")
matrix(rows, cols) =>
    total = 0.0
    for r = 0 to rows - 1
        for c = 0 to cols - 1
            total := total + r + c
    total
plot(matrix(3, 4))
`,
			mustContainAll: []string{
				"float64(r)",
				"float64(c)",
			},
			forbiddenPattern: []string{
				"var rSeries",
				"var cSeries",
			},
			description: "each loop counter in nested loops receives independent float64 cast",
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
