package codegen

import (
	"strings"
	"testing"
)

/*
Validates scalar access resolution through delegated call handlers in arrow functions.

ForwardSeriesBuffer paradigm: when arrow function bodies contain non-TA function calls
(math.*, user-defined), arguments must resolve parameters and locals as scalars,
not as %sSeries.GetCurrent(). This complements arrow_expression_scalar_access_test.go
which validates scalar access in direct expressions (conditions, logic, binary ops).
*/

/*
TestArrowDelegatedCall_MathScalarResolution validates math function arguments

	use scalar access for arrow parameters and locals
*/
func TestArrowDelegatedCall_MathScalarResolution(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "math.abs with parameter argument",
			pine: `
//@version=5
indicator("Test")
normalize(src) =>
    math.abs(src)
plot(normalize(close))
`,
			mustContainAll: []string{
				"math.Abs(src)",
			},
			forbiddenPattern: []string{
				"srcSeries.GetCurrent()",
			},
			description: "math.abs resolves parameter as scalar",
		},
		{
			name: "math.max with two parameters",
			pine: `
//@version=5
indicator("Test")
clamp_low(val, floor_val) =>
    math.max(val, floor_val)
plot(clamp_low(close, 0))
`,
			mustContainAll: []string{
				"math.Max(val, floor_val)",
			},
			forbiddenPattern: []string{
				"valSeries.GetCurrent()",
				"floor_valSeries.GetCurrent()",
			},
			description: "math.max resolves both parameters as scalars",
		},
		{
			name: "math.min with parameter and literal",
			pine: `
//@version=5
indicator("Test")
cap(src) =>
    math.min(src, 100)
plot(cap(close))
`,
			mustContainAll: []string{
				"math.Min(src,",
			},
			forbiddenPattern: []string{
				"srcSeries.GetCurrent()",
			},
			description: "math.min resolves parameter as scalar alongside literal",
		},
		{
			name: "math.abs with local variable argument",
			pine: `
//@version=5
indicator("Test")
spread(src) =>
    diff = src - close
    math.abs(diff)
plot(spread(open))
`,
			mustContainAll: []string{
				"diff := (src - bar.Close)",
				"math.Abs(diff)",
			},
			forbiddenPattern: []string{
				"diffSeries.GetCurrent()",
				"srcSeries.GetCurrent()",
			},
			description: "math.abs resolves local variable as scalar",
		},
		{
			name: "math.pow with parameter arguments",
			pine: `
//@version=5
indicator("Test")
power(base, exp) =>
    math.pow(base, exp)
plot(power(close, 2))
`,
			mustContainAll: []string{
				"math.Pow(base,",
			},
			forbiddenPattern: []string{
				"baseSeries.GetCurrent()",
				"expSeries.GetCurrent()",
			},
			description: "math.pow resolves both parameters as scalars",
		},
		{
			name: "math.sqrt with local variable",
			pine: `
//@version=5
indicator("Test")
vol_adjusted(src, factor) =>
    adjusted = src * factor
    math.sqrt(adjusted)
plot(vol_adjusted(close, 2))
`,
			mustContainAll: []string{
				"adjusted := (src * factor)",
				"math.Sqrt(adjusted)",
			},
			forbiddenPattern: []string{
				"adjustedSeries.GetCurrent()",
				"srcSeries.GetCurrent()",
				"factorSeries.GetCurrent()",
			},
			description: "math.sqrt resolves local variable derived from parameters as scalar",
		},
		{
			name: "multiple math calls in same arrow body",
			pine: `
//@version=5
indicator("Test")
normalize(src, scale) =>
    abs_val = math.abs(src)
    capped = math.min(abs_val, scale)
    capped
plot(normalize(close, 100))
`,
			mustContainAll: []string{
				"math.Abs(src)",
				"math.Min(abs_val,",
			},
			forbiddenPattern: []string{
				"srcSeries.GetCurrent()",
				"abs_valSeries.GetCurrent()",
			},
			description: "sequential math calls each resolve identifiers as scalars",
		},
		{
			name: "math with binary expression argument containing parameters",
			pine: `
//@version=5
indicator("Test")
range_ratio(hi, lo) =>
    math.abs(hi - lo)
plot(range_ratio(high, low))
`,
			mustContainAll: []string{
				"math.Abs(",
				"hi - lo",
			},
			forbiddenPattern: []string{
				"hiSeries.GetCurrent()",
				"loSeries.GetCurrent()",
			},
			description: "binary expression inside math resolves parameters as scalars",
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
					t.Errorf("%s: missing required pattern:\n  %s\n\nGenerated code:\n%s",
						tt.description, required, truncateCode(code, 1500))
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(code, forbidden) {
					t.Errorf("%s: found forbidden Series.GetCurrent() pattern:\n  %s\n\nGenerated code:\n%s",
						tt.description, forbidden, truncateCode(code, 1500))
				}
			}
		})
	}
}

/*
TestArrowDelegatedCall_MathWithMixedIdentifierCategories validates correct resolution

	when math arguments span parameters, locals, loop-modified, and outer scope
*/
func TestArrowDelegatedCall_MathWithMixedIdentifierCategories(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "math with parameter and computed local",
			pine: `
//@version=5
indicator("Test")
safe_div(numerator, divisor) =>
    safe_d = math.max(divisor, 1)
    numerator / safe_d
plot(safe_div(close, volume))
`,
			mustContainAll: []string{
				"math.Max(divisor,",
			},
			forbiddenPattern: []string{
				"divisorSeries.GetCurrent()",
				"numeratorSeries.GetCurrent()",
			},
			description: "parameter used in math.max then result used in division",
		},
		{
			name: "math.max chained with math.min for clamping",
			pine: `
//@version=5
indicator("Test")
clamp(val, lo, hi) =>
    math.min(math.max(val, lo), hi)
plot(clamp(close, 50, 150))
`,
			mustContainAll: []string{
				"math.Min(",
				"math.Max(val, lo)",
			},
			forbiddenPattern: []string{
				"valSeries.GetCurrent()",
				"loSeries.GetCurrent()",
				"hiSeries.GetCurrent()",
			},
			description: "nested math calls all resolve parameters as scalars",
		},
		{
			name: "math.abs in condition with parameters",
			pine: `
//@version=5
indicator("Test")
is_close(a, b, tolerance) =>
    math.abs(a - b) < tolerance ? 1 : 0
plot(is_close(close, open, 0.5))
`,
			mustContainAll: []string{
				"math.Abs(",
				"a - b",
			},
			forbiddenPattern: []string{
				"aSeries.GetCurrent()",
				"bSeries.GetCurrent()",
				"toleranceSeries.GetCurrent()",
			},
			description: "math inside ternary condition resolves parameters as scalars",
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
					t.Errorf("%s: missing required pattern:\n  %s\n\nGenerated code:\n%s",
						tt.description, required, truncateCode(code, 1500))
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(code, forbidden) {
					t.Errorf("%s: found forbidden Series.GetCurrent() pattern:\n  %s\n\nGenerated code:\n%s",
						tt.description, forbidden, truncateCode(code, 1500))
				}
			}
		})
	}
}

/*
TestArrowDelegatedCall_UserDefinedFunctionScalarResolution validates that arrow functions

	calling other user-defined functions resolve arguments correctly
*/
func TestArrowDelegatedCall_UserDefinedFunctionScalarResolution(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "arrow calling another arrow with parameter forwarding",
			pine: `
//@version=5
indicator("Test")
helper(x) =>
    x * 2
main_calc(src) =>
    helper(src)
plot(main_calc(close))
`,
			mustContainAll: []string{
				"helper(",
			},
			forbiddenPattern: []string{
				"srcSeries.GetCurrent()",
			},
			description: "arrow calling another arrow forwards parameter as scalar",
		},
		{
			name: "arrow calling another arrow with local variable",
			pine: `
//@version=5
indicator("Test")
double(x) =>
    x * 2
process(src) =>
    intermediate = src + 10
    double(intermediate)
plot(process(close))
`,
			mustContainAll: []string{
				"intermediate := (src + 10)",
				"double(",
			},
			forbiddenPattern: []string{
				"intermediateSeries.GetCurrent()",
				"srcSeries.GetCurrent()",
			},
			description: "arrow calling another arrow with local variable uses scalar",
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
					t.Errorf("%s: missing required pattern:\n  %s\n\nGenerated code:\n%s",
						tt.description, required, truncateCode(code, 1500))
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(code, forbidden) {
					t.Errorf("%s: found forbidden Series.GetCurrent() pattern:\n  %s\n\nGenerated code:\n%s",
						tt.description, forbidden, truncateCode(code, 1500))
				}
			}
		})
	}
}
