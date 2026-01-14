package codegen

import (
	"strings"
	"testing"
)

/*
Validates dual-access pattern for arrow function local variables.

ForwardSeriesBuffer paradigm requires scalar declaration (up := expr) before Series.Set(up),
ensuring temporal correctness. Tests validate this ordering and scalar-only access within
current bar computations across all code paths. Generalized tests for algorithmic behavior,
not specific bug verification.
*/

/* TestArrowDualAccess_ScalarDeclarationWithSeriesStorage validates scalar+Series pattern */
func TestArrowDualAccess_ScalarDeclarationWithSeriesStorage(t *testing.T) {
	tests := []struct {
		name              string
		pine              string
		expectedScalar    []string // Scalar declarations that MUST exist
		expectedSeriesSet []string // Series.Set() calls that MUST exist
		forbiddenPattern  []string // Patterns that should NOT exist
		description       string
	}{
		{
			name: "single variable simple assignment",
			pine: `
//@version=5
indicator("Test")
calc(len) =>
    result = close * len
    result
plot(calc(20))
`,
			expectedScalar: []string{
				"result := (bar.Close * len)",
			},
			expectedSeriesSet: []string{
				"resultSeries.Set(result)",
			},
			forbiddenPattern: []string{
				"resultSeries.Set((bar.Close * len))", // Should use scalar, not inline expr
			},
			description: "single variable uses scalar declaration then Series.Set()",
		},
		{
			name: "multiple variables sequential",
			pine: `
//@version=5
indicator("Test")
compute(factor) =>
    a = close * factor
    b = open * factor
    c = high * factor
    c
plot(compute(2))
`,
			expectedScalar: []string{
				"a := (bar.Close * factor)",
				"b := (bar.Open * factor)",
				"c := (bar.High * factor)",
			},
			expectedSeriesSet: []string{
				"aSeries.Set(a)",
				"bSeries.Set(b)",
				"cSeries.Set(c)",
			},
			forbiddenPattern: nil,
			description:      "multiple variables each get scalar declaration and Series.Set()",
		},
		{
			name: "tuple destructuring",
			pine: `
//@version=5
indicator("Test")
pair(multiplier) =>
    first = close * multiplier
    second = open * multiplier
    [first, second]
[x, y] = pair(1.5)
`,
			expectedScalar: []string{
				"first := (bar.Close * multiplier)",
				"second := (bar.Open * multiplier)",
			},
			expectedSeriesSet: []string{
				"firstSeries.Set(first)",
				"secondSeries.Set(second)",
			},
			forbiddenPattern: nil,
			description:      "tuple return values use scalar declarations",
		},
		{
			name: "variable with complex expression",
			pine: `
//@version=5
indicator("Test")
average(period) =>
    avg = (close + open + high + low) / 4
    avg
plot(average(10))
`,
			expectedScalar: []string{
				"avg := ((bar.Close + (bar.Open + (bar.High + bar.Low))) / 4)",
			},
			expectedSeriesSet: []string{
				"avgSeries.Set(avg)",
			},
			forbiddenPattern: nil,
			description:      "complex expressions stored in scalars before Series.Set()",
		},
		{
			name: "variable with conditional expression",
			pine: `
//@version=5
indicator("Test")
select(threshold) =>
    value = close > threshold ? high : low
    value
plot(select(100))
`,
			expectedScalar: []string{
				"value := func() float64 { if (bar.Close > threshold)",
			},
			expectedSeriesSet: []string{
				"valueSeries.Set(value)",
			},
			forbiddenPattern: nil,
			description:      "conditional expressions stored in scalars",
		},
		{
			name: "variable with unary expression",
			pine: `
//@version=5
indicator("Test")
negate(val) =>
    negative = -val
    negative
plot(negate(100))
`,
			expectedScalar: []string{
				"negative := -val",
			},
			expectedSeriesSet: []string{
				"negativeSeries.Set(negative)",
			},
			forbiddenPattern: nil,
			description:      "unary expressions stored in scalars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, expected := range tt.expectedScalar {
				if !strings.Contains(code, expected) {
					t.Errorf("%s: Missing scalar declaration:\n  %s\n\nGenerated:\n%s",
						tt.description, expected, code)
				}
			}

			for _, expected := range tt.expectedSeriesSet {
				if !strings.Contains(code, expected) {
					t.Errorf("%s: Missing Series.Set() call:\n  %s\n\nGenerated:\n%s",
						tt.description, expected, code)
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(code, forbidden) {
					t.Errorf("%s: Found forbidden pattern:\n  %s\n\nGenerated:\n%s",
						tt.description, forbidden, code)
				}
			}
		})
	}
}

/* TestArrowDualAccess_CurrentBarScalarReferences validates scalar access in expressions */
func TestArrowDualAccess_CurrentBarScalarReferences(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		expectedScalar   []string // Scalar references in expressions
		forbiddenPattern []string // Series.GetCurrent() should NOT appear
		description      string
	}{
		{
			name: "local variable in binary expression",
			pine: `
//@version=5
indicator("Test")
calc(multiplier) =>
    base = close * 2
    result = base + multiplier
    result
plot(calc(5))
`,
			expectedScalar: []string{
				"base := (bar.Close * 2)",
				"result := (base + multiplier)",
			},
			forbiddenPattern: []string{
				"baseSeries.GetCurrent()",
			},
			description: "local variable in binary expression uses scalar",
		},
		{
			name: "local variable in conditional test",
			pine: `
//@version=5
indicator("Test")
check(threshold) =>
    value = close + open
    signal = value > threshold ? 1 : 0
    signal
plot(check(100))
`,
			expectedScalar: []string{
				"value := (bar.Close + bar.Open)",
				"if (value > threshold)",
			},
			forbiddenPattern: []string{
				"valueSeries.GetCurrent() > threshold",
			},
			description: "local variable in conditional uses scalar",
		},
		{
			name: "local variable in conditional consequent",
			pine: `
//@version=5
indicator("Test")
select(threshold) =>
    high_val = high * 1.1
    low_val = low * 0.9
    result = close > threshold ? high_val : low_val
    result
plot(select(100))
`,
			expectedScalar: []string{
				"high_val := (bar.High * 1.1)",
				"low_val := (bar.Low * 0.9)",
				"return high_val",
				"return low_val",
			},
			forbiddenPattern: []string{
				"high_valSeries.GetCurrent()",
				"low_valSeries.GetCurrent()",
			},
			description: "local variables in conditional branches use scalars",
		},
		{
			name: "multiple local variables in expression",
			pine: `
//@version=5
indicator("Test")
combine(factor) =>
    a = close * factor
    b = open * factor
    sum = a + b
    sum
plot(combine(2))
`,
			expectedScalar: []string{
				"a := (bar.Close * factor)",
				"b := (bar.Open * factor)",
				"sum := (a + b)",
			},
			forbiddenPattern: []string{
				"aSeries.GetCurrent()",
				"bSeries.GetCurrent()",
			},
			description: "multiple local variables use scalars",
		},
		{
			name: "local variable in nested expression",
			pine: `
//@version=5
indicator("Test")
nested(threshold) =>
    base = close + open
    adjusted = (base * 2) / threshold
    adjusted
plot(nested(100))
`,
			expectedScalar: []string{
				"base := (bar.Close + bar.Open)",
				"adjusted := ((base * 2) / threshold)",
			},
			forbiddenPattern: []string{
				"baseSeries.GetCurrent()",
			},
			description: "local variable in nested parentheses uses scalar",
		},
		{
			name: "local variable in unary expression",
			pine: `
//@version=5
indicator("Test")
invert(multiplier) =>
    value = close * multiplier
    inverted = -value
    inverted
plot(invert(2))
`,
			expectedScalar: []string{
				"value := (bar.Close * multiplier)",
				"inverted := -value",
			},
			forbiddenPattern: []string{
				"-valueSeries.GetCurrent()",
			},
			description: "local variable in unary expression uses scalar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, expected := range tt.expectedScalar {
				if !strings.Contains(code, expected) {
					t.Errorf("%s: Missing scalar access:\n  %s\n\nGenerated:\n%s",
						tt.description, expected, code)
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(code, forbidden) {
					t.Errorf("%s: Found forbidden Series.GetCurrent():\n  %s\n\nGenerated:\n%s",
						tt.description, forbidden, code)
				}
			}
		})
	}
}

/* TestArrowDualAccess_ReturnScalarValues validates return statements use scalars */
func TestArrowDualAccess_ReturnScalarValues(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		expectedReturn   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "single value return",
			pine: `
//@version=5
indicator("Test")
compute(len) =>
    result = close * len
    result
plot(compute(10))
`,
			expectedReturn: []string{
				"return result",
			},
			forbiddenPattern: []string{
				"return resultSeries.GetCurrent()",
			},
			description: "single return value uses scalar",
		},
		{
			name: "tuple return",
			pine: `
//@version=5
indicator("Test")
pair(multiplier) =>
    a = close * multiplier
    b = open * multiplier
    [a, b]
[x, y] = pair(2)
`,
			expectedReturn: []string{
				"return a, b",
			},
			forbiddenPattern: []string{
				"return aSeries.GetCurrent()",
				"return bSeries.GetCurrent()",
			},
			description: "tuple return uses scalars",
		},
		{
			name: "three element tuple",
			pine: `
//@version=5
indicator("Test")
triple(len) =>
    x = close + len
    y = open + len
    z = high + len
    [x, y, z]
[a, b, c] = triple(5)
`,
			expectedReturn: []string{
				"return x, y, z",
			},
			forbiddenPattern: []string{
				"xSeries.GetCurrent()",
				"ySeries.GetCurrent()",
				"zSeries.GetCurrent()",
			},
			description: "three-element tuple uses scalars",
		},
		{
			name: "immediate return of expression",
			pine: `
//@version=5
indicator("Test")
direct(val) =>
    close * val
plot(direct(2))
`,
			expectedReturn: []string{
				"return (bar.Close * val)",
			},
			forbiddenPattern: nil,
			description:      "immediate expression return is scalar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, expected := range tt.expectedReturn {
				if !strings.Contains(code, expected) {
					t.Errorf("%s: Missing scalar return:\n  %s\n\nGenerated:\n%s",
						tt.description, expected, code)
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(code, forbidden) {
					t.Errorf("%s: Found forbidden Series.GetCurrent() in return:\n  %s\n\nGenerated:\n%s",
						tt.description, forbidden, code)
				}
			}
		})
	}
}

/* TestArrowDualAccess_ParameterScalarAccess validates function parameters remain scalars */
func TestArrowDualAccess_ParameterScalarAccess(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		expectedParam    []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "parameter in binary expression",
			pine: `
//@version=5
indicator("Test")
calc(multiplier) =>
    result = close * multiplier
    result
plot(calc(2))
`,
			expectedParam: []string{
				"(bar.Close * multiplier)",
			},
			forbiddenPattern: []string{
				"multiplierSeries",
			},
			description: "function parameter remains scalar in expressions",
		},
		{
			name: "parameter in conditional",
			pine: `
//@version=5
indicator("Test")
check(threshold) =>
    signal = close > threshold ? 1 : 0
    signal
plot(check(100))
`,
			expectedParam: []string{
				"(bar.Close > threshold)",
			},
			forbiddenPattern: []string{
				"thresholdSeries",
			},
			description: "parameter in conditional remains scalar",
		},
		{
			name: "multiple parameters",
			pine: `
//@version=5
indicator("Test")
combine(factor1, factor2) =>
    result = (close * factor1) + (open * factor2)
    result
plot(combine(2, 3))
`,
			expectedParam: []string{
				"(bar.Close * factor1)",
				"(bar.Open * factor2)",
			},
			forbiddenPattern: []string{
				"factor1Series",
				"factor2Series",
			},
			description: "multiple parameters remain scalars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, expected := range tt.expectedParam {
				if !strings.Contains(code, expected) {
					t.Errorf("%s: Missing scalar parameter access:\n  %s\n\nGenerated:\n%s",
						tt.description, expected, code)
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(code, forbidden) {
					t.Errorf("%s: Found forbidden Series pattern for parameter:\n  %s\n\nGenerated:\n%s",
						tt.description, forbidden, code)
				}
			}
		})
	}
}

/* TestArrowDualAccess_TemporalOrdering validates scalar-before-Series execution order */
func TestArrowDualAccess_TemporalOrdering(t *testing.T) {
	tests := []struct {
		name        string
		pine        string
		mustPrecede []struct{ before, after string }
		description string
	}{
		{
			name: "single variable ordering",
			pine: `
//@version=5
indicator("Test")
calc(len) =>
    result = close * len
    result
plot(calc(10))
`,
			mustPrecede: []struct{ before, after string }{
				{
					before: "result := (bar.Close * len)",
					after:  "resultSeries.Set(result)",
				},
			},
			description: "scalar declaration must precede Series.Set()",
		},
		{
			name: "sequential variable ordering",
			pine: `
//@version=5
indicator("Test")
multi(factor) =>
    a = close * factor
    b = a + open
    b
plot(multi(2))
`,
			mustPrecede: []struct{ before, after string }{
				{
					before: "a := (bar.Close * factor)",
					after:  "aSeries.Set(a)",
				},
				{
					before: "aSeries.Set(a)",
					after:  "b := (a + bar.Open)",
				},
			},
			description: "dependent variables maintain temporal order",
		},
		{
			name: "tuple variable ordering",
			pine: `
//@version=5
indicator("Test")
pair(multiplier) =>
    first = close * multiplier
    second = open * multiplier
    [first, second]
[x, y] = pair(2)
`,
			mustPrecede: []struct{ before, after string }{
				{
					before: "first := (bar.Close * multiplier)",
					after:  "firstSeries.Set(first)",
				},
				{
					before: "second := (bar.Open * multiplier)",
					after:  "secondSeries.Set(second)",
				},
			},
			description: "tuple elements maintain scalar-before-Series ordering",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, ordering := range tt.mustPrecede {
				beforeIdx := strings.Index(code, ordering.before)
				afterIdx := strings.Index(code, ordering.after)

				if beforeIdx == -1 {
					t.Errorf("%s: Missing expected pattern:\n  %s", tt.description, ordering.before)
					continue
				}

				if afterIdx == -1 {
					t.Errorf("%s: Missing expected pattern:\n  %s", tt.description, ordering.after)
					continue
				}

				if beforeIdx >= afterIdx {
					t.Errorf("%s: Temporal ordering violated:\n  '%s'\n  should precede\n  '%s'\n\nGenerated:\n%s",
						tt.description, ordering.before, ordering.after, code)
				}
			}
		})
	}
}

/* TestArrowDualAccess_EdgeCases validates boundary conditions and error handling */
func TestArrowDualAccess_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		pine            string
		expectedPattern []string
		description     string
	}{
		{
			name: "single letter variable names",
			pine: `
//@version=5
indicator("Test")
calc(f) =>
    x = close * f
    y = open * f
    x + y
plot(calc(2))
`,
			expectedPattern: []string{
				"x := (bar.Close * f)",
				"xSeries.Set(x)",
				"y := (bar.Open * f)",
				"ySeries.Set(y)",
			},
			description: "single-letter variables work correctly",
		},
		{
			name: "variable with underscores",
			pine: `
//@version=5
indicator("Test")
calc(len) =>
    my_value = close * len
    my_value
plot(calc(10))
`,
			expectedPattern: []string{
				"my_value := (bar.Close * len)",
				"my_valueSeries.Set(my_value)",
			},
			description: "underscore variable names work correctly",
		},
		{
			name: "empty function body with immediate return",
			pine: `
//@version=5
indicator("Test")
identity(x) =>
    x
plot(identity(close))
`,
			expectedPattern: []string{
				"return x",
			},
			description: "immediate parameter return uses scalar",
		},
		{
			name: "variable reassignment",
			pine: `
//@version=5
indicator("Test")
update(initial) =>
    value = initial
    value := value * 2
    value
plot(update(10))
`,
			expectedPattern: []string{
				"value := initial",
				"valueSeries.Set(value)",
				"value := (value * 2)",
				"valueSeries.Set(value)",
			},
			description: "variable reassignment maintains dual-access pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, expected := range tt.expectedPattern {
				if !strings.Contains(code, expected) {
					t.Errorf("%s: Missing expected pattern:\n  %s\n\nGenerated:\n%s",
						tt.description, expected, code)
				}
			}
		})
	}
}

/* TestArrowDualAccess_ConsistencyAcrossContexts validates uniform behavior */
func TestArrowDualAccess_ConsistencyAcrossContexts(t *testing.T) {
	tests := []struct {
		name            string
		pine            string
		expectedPattern []string
		description     string
	}{
		{
			name: "nested function calls maintain dual-access",
			pine: `
//@version=5
indicator("Test")
inner(x) =>
    x * 2

outer(y) =>
    temp = inner(y)
    result = temp + 10
    result

plot(outer(5))
`,
			expectedPattern: []string{
				"temp := inner(",
				"tempSeries.Set(temp)",
				"result := (temp + 10)",
				"resultSeries.Set(result)",
			},
			description: "nested calls maintain dual-access pattern",
		},
		{
			name: "multiple functions same variable names",
			pine: `
//@version=5
indicator("Test")
func1(x) =>
    result = x * 2
    result

func2(y) =>
    result = y * 3
    result

plot(func1(10) + func2(20))
`,
			expectedPattern: []string{
				"result := (x * 2)",
				"resultSeries.Set(result)",
				"result := (y * 3)",
				"resultSeries.Set(result)",
			},
			description: "same variable names in different functions work correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, expected := range tt.expectedPattern {
				if !strings.Contains(code, expected) {
					t.Errorf("%s: Missing expected pattern:\n  %s\n\nGenerated:\n%s",
						tt.description, expected, code)
				}
			}
		})
	}
}
