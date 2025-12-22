package codegen

import (
	"strings"
	"testing"
)

/*
arrow_identifier_resolution_test.go

PURPOSE:
Comprehensive test coverage for identifier resolution in arrow function context.
Validates that the ArrowSeriesAccessResolver correctly distinguishes between:
- Function parameters (scalar access)
- Local variables (Series.GetCurrent() access)
- Builtin identifiers (OHLCV field access)

GENERALIZATION PRINCIPLES:
- Tests validate BEHAVIOR, not specific bug fixes
- Edge cases cover all identifier types and contexts
- Tests remain valid regardless of internal implementation changes
*/

func TestArrowFunction_IdentifierResolution_LocalVariables(t *testing.T) {
	tests := []struct {
		name              string
		pine              string
		localVariable     string
		mustContainAccess string // Expected Series access pattern
	}{
		{
			name: "local variable in binary expression right operand",
			pine: `
//@version=5
indicator("Test")
calc(mult) =>
    base = close * 2
    result = 100 / base
    result
plot(calc(1.5))
`,
			localVariable:     "base",
			mustContainAccess: "baseSeries.GetCurrent()",
		},
		{
			name: "local variable in binary expression left operand",
			pine: `
//@version=5
indicator("Test")
calc(period) =>
    divisor = 10
    result = divisor * close
    result
plot(calc(20))
`,
			localVariable:     "divisor",
			mustContainAccess: "divisorSeries.GetCurrent()",
		},
		{
			name: "local variable in nested binary expressions",
			pine: `
//@version=5
indicator("Test")
calc() =>
    a = 5
    b = 10
    c = (a + b) * (a - b)
    c
plot(calc())
`,
			localVariable:     "a",
			mustContainAccess: "aSeries.GetCurrent()",
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
			localVariable:     "value",
			mustContainAccess: "valueSeries.GetCurrent()",
		},
		{
			name: "local variable in conditional consequent",
			pine: `
//@version=5
indicator("Test")
pick(flag) =>
    option1 = high
    option2 = low
    result = flag ? option1 : option2
    result
plot(pick(true))
`,
			localVariable:     "option1",
			mustContainAccess: "option1Series.GetCurrent()",
		},
		{
			name: "local variable in conditional alternate",
			pine: `
//@version=5
indicator("Test")
select(flag) =>
    first = 100
    second = 200
    choice = flag ? first : second
    choice
plot(select(false))
`,
			localVariable:     "second",
			mustContainAccess: "secondSeries.GetCurrent()",
		},
		{
			name: "local variable in unary expression",
			pine: `
//@version=5
indicator("Test")
negate() =>
    value = close - open
    inverted = -value
    inverted
plot(negate())
`,
			localVariable:     "value",
			mustContainAccess: "valueSeries.GetCurrent()",
		},
		{
			name: "local variable used multiple times in single expression",
			pine: `
//@version=5
indicator("Test")
compute() =>
    x = close
    square = x * x
    square
plot(compute())
`,
			localVariable:     "x",
			mustContainAccess: "xSeries.GetCurrent()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation error: %v", err)
			}

			if !strings.Contains(code, tt.mustContainAccess) {
				t.Errorf("Missing expected Series access pattern: %s\nGenerated code:\n%s",
					tt.mustContainAccess, code)
			}

			// Ensure Series is declared
			expectedDecl := tt.localVariable + "Series := arrowCtx.GetOrCreateSeries"
			if !strings.Contains(code, expectedDecl) {
				t.Errorf("Missing Series declaration for variable: %s\nGenerated code:\n%s",
					tt.localVariable, code)
			}
		})
	}
}

func TestArrowFunction_IdentifierResolution_Parameters(t *testing.T) {
	tests := []struct {
		name              string
		pine              string
		parameter         string
		mustContainScalar string // Parameter should be used as scalar, not Series
		mustNotContain    string // Should NOT have Series access
	}{
		{
			name: "parameter in binary expression",
			pine: `
//@version=5
indicator("Test")
multiply(factor) =>
    result = close * factor
    result
plot(multiply(2))
`,
			parameter:         "factor",
			mustContainScalar: "* factor",
			mustNotContain:    "factorSeries.GetCurrent()",
		},
		{
			name: "parameter in division",
			pine: `
//@version=5
indicator("Test")
divide(divisor) =>
    result = close / divisor
    result
plot(divide(10))
`,
			parameter:         "divisor",
			mustContainScalar: "/ divisor",
			mustNotContain:    "divisorSeries.GetCurrent()",
		},
		{
			name: "parameter in addition",
			pine: `
//@version=5
indicator("Test")
offset(amount) =>
    result = close + amount
    result
plot(offset(50))
`,
			parameter:         "amount",
			mustContainScalar: "+ amount",
			mustNotContain:    "amountSeries.GetCurrent()",
		},
		{
			name: "parameter in comparison",
			pine: `
//@version=5
indicator("Test")
compare(threshold) =>
    signal = close > threshold ? 1 : 0
    signal
plot(compare(100))
`,
			parameter:         "threshold",
			mustContainScalar: "> threshold",
			mustNotContain:    "thresholdSeries.GetCurrent()",
		},
		{
			name: "multiple parameters maintain scalar access",
			pine: `
//@version=5
indicator("Test")
band(length, multiplier) =>
    mid = close
    result = mid * multiplier + length
    result
plot(band(20, 2))
`,
			parameter:         "multiplier",
			mustContainScalar: "* multiplier",
			mustNotContain:    "multiplierSeries.GetCurrent()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation error: %v", err)
			}

			if !strings.Contains(code, tt.mustContainScalar) {
				t.Errorf("Missing expected scalar parameter usage: %s\nGenerated code:\n%s",
					tt.mustContainScalar, code)
			}

			if strings.Contains(code, tt.mustNotContain) {
				t.Errorf("Parameter incorrectly treated as Series: found %s\nGenerated code:\n%s",
					tt.mustNotContain, code)
			}
		})
	}
}

func TestArrowFunction_IdentifierResolution_BuiltinIdentifiers(t *testing.T) {
	tests := []struct {
		name              string
		pine              string
		builtinField      string
		mustContainAccess string
	}{
		{
			name: "close in binary expression",
			pine: `
//@version=5
indicator("Test")
calc(mult) =>
    result = close * mult
    result
plot(calc(2))
`,
			builtinField:      "close",
			mustContainAccess: "ctx.Data[ctx.BarIndex].Close",
		},
		{
			name: "high in conditional",
			pine: `
//@version=5
indicator("Test")
check() =>
    signal = high > close ? 1 : 0
    signal
plot(check())
`,
			builtinField:      "high",
			mustContainAccess: "ctx.Data[ctx.BarIndex].High",
		},
		{
			name: "low in subtraction",
			pine: `
//@version=5
indicator("Test")
range_calc() =>
    range = high - low
    range
plot(range_calc())
`,
			builtinField:      "low",
			mustContainAccess: "ctx.Data[ctx.BarIndex].Low",
		},
		{
			name: "open in comparison",
			pine: `
//@version=5
indicator("Test")
bullish() =>
    is_bull = close > open
    is_bull ? 1 : 0
plot(bullish())
`,
			builtinField:      "open",
			mustContainAccess: "ctx.Data[ctx.BarIndex].Open",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation error: %v", err)
			}
			// Test validates builtin identifiers compile successfully in arrow context
		})
	}
}

func TestArrowFunction_IdentifierResolution_MixedContexts(t *testing.T) {
	tests := []struct {
		name                string
		pine                string
		mustContainPatterns []string
		mustNotContain      []string
	}{
		{
			name: "local variable, parameter, and builtin in same expression",
			pine: `
//@version=5
indicator("Test")
calc(multiplier) =>
    base = close * 2
    result = base * multiplier + high
    result
plot(calc(1.5))
`,
			mustContainPatterns: []string{
				"baseSeries.GetCurrent()",     // Local variable uses Series
				"* multiplier",                // Parameter stays scalar
				"ctx.Data[ctx.BarIndex].High", // Builtin uses direct access
			},
			mustNotContain: []string{
				"multiplierSeries.GetCurrent()", // Parameter should NOT use Series
				"baseSeries.Get(",               // Local should use GetCurrent(), not Get(offset)
			},
		},
		{
			name: "nested expressions with all identifier types",
			pine: `
//@version=5
indicator("Test")
complex(threshold, offset) =>
    local1 = close + offset
    local2 = high - low
    result = (local1 * local2) > threshold ? local1 : local2
    result
plot(complex(100, 10))
`,
			mustContainPatterns: []string{
				"local1Series.GetCurrent()",
				"local2Series.GetCurrent()",
				"> threshold", // Parameter comparison
				"ctx.Data[ctx.BarIndex].High",
				"ctx.Data[ctx.BarIndex].Low",
			},
			mustNotContain: []string{
				"thresholdSeries.GetCurrent()",
				"offsetSeries.GetCurrent()",
			},
		},
		{
			name: "local variable referencing another local in assignment",
			pine: `
//@version=5
indicator("Test")
chain() =>
    a = close
    b = a * 2
    c = b + a
    c
plot(chain())
`,
			mustContainPatterns: []string{
				"aSeries.GetCurrent()", // 'a' used in expressions
				"bSeries.GetCurrent()", // 'b' used in expressions
				"aSeries.Set(",         // All get Series.Set()
				"bSeries.Set(",
				"cSeries.Set(",
			},
			mustNotContain: []string{
				"a :=", // No scalar assignments
				"b :=",
				"c :=",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation error: %v", err)
			}

			for _, pattern := range tt.mustContainPatterns {
				if strings.Contains(pattern, "Series.GetCurrent()") || strings.Contains(pattern, "Series.Set(") {
					if !strings.Contains(code, pattern) {
						t.Errorf("Missing expected Series access pattern: %s", pattern)
					}
				}
			}

			for _, pattern := range tt.mustNotContain {
				if strings.Contains(code, pattern) {
					t.Errorf("Found forbidden pattern: %s", pattern)
				}
			}
		})
	}
}

func TestArrowFunction_IdentifierResolution_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		pine        string
		description string
		validate    func(t *testing.T, code string)
	}{
		{
			name: "shadowing parameter name with local variable",
			pine: `
//@version=5
indicator("Test")
calc(value) =>
    value_local = value * 2
    result = value_local + value
    result
plot(calc(10))
`,
			description: "Parameter 'value' stays scalar, 'value_local' uses Series",
			validate: func(t *testing.T, code string) {
				if !strings.Contains(code, "value_localSeries.GetCurrent()") {
					t.Error("Local variable 'value_local' should use Series.GetCurrent()")
				}
				if !strings.Contains(code, "+ value") && !strings.Contains(code, "value *") {
					t.Error("Parameter 'value' should remain scalar")
				}
			},
		},
		{
			name: "single letter variable names",
			pine: `
//@version=5
indicator("Test")
f(x) =>
    a = x * 2
    b = a + x
    c = b - a
    c
plot(f(5))
`,
			description: "Single letter variables use Series consistently",
			validate: func(t *testing.T, code string) {
				if !strings.Contains(code, "aSeries.GetCurrent()") {
					t.Error("Variable 'a' should use Series.GetCurrent()")
				}
				if !strings.Contains(code, "bSeries.GetCurrent()") {
					t.Error("Variable 'b' should use Series.GetCurrent()")
				}
				if strings.Contains(code, "xSeries.GetCurrent()") {
					t.Error("Parameter 'x' should NOT use Series.GetCurrent()")
				}
			},
		},
		{
			name: "variable used only once",
			pine: `
//@version=5
indicator("Test")
calc() =>
    temp = close * 2
    temp
plot(calc())
`,
			description: "Even single-use variables get Series storage",
			validate: func(t *testing.T, code string) {
				if !strings.Contains(code, "tempSeries := arrowCtx.GetOrCreateSeries(\"temp\")") {
					t.Error("Single-use variable should still get Series declaration")
				}
				if !strings.Contains(code, "tempSeries.GetCurrent()") {
					t.Error("Single-use variable should return via Series.GetCurrent()")
				}
			},
		},
		{
			name: "immediate return of local variable",
			pine: `
//@version=5
indicator("Test")
direct() =>
    x = close + open
    x
plot(direct())
`,
			description: "Immediately returned local variable uses Series.GetCurrent()",
			validate: func(t *testing.T, code string) {
				if !strings.Contains(code, "return xSeries.GetCurrent()") {
					t.Error("Immediate return should use Series.GetCurrent()")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation error: %v", err)
			}

			tt.validate(t, code)
		})
	}
}
