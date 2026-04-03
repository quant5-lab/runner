package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

/* TestInlineFunctionsInConditionals validates inline TA functions in various conditional contexts */
func TestInlineFunctionsInConditionals(t *testing.T) {
	tests := []struct {
		name           string
		script         string
		mustContain    []string
		mustNotContain []string
		description    string
	}{
		{
			name: "numeric inline function in ternary test",
			script: `//@version=4
study("Test", overlay=true)
len = 5
result = dev(close, len) ? 1 : 0
plot(result)`,
			mustContain: []string{
				"value.IsTrue",
				"devSum := 0.0",
				"resultSeries.Set",
			},
			mustNotContain: []string{
				"undefined:",
			},
			description: "dev() in ternary with != 0 conversion",
		},
		{
			name: "boolean function in ternary test",
			script: `//@version=4
study("Test", overlay=true)
fast = sma(close, 10)
slow = sma(close, 20)
signal = close > fast ? 1 : 0
plot(signal)`,
			mustContain: []string{
				"signalSeries.Set",
			},
			mustNotContain: []string{
				"undefined:",
			},
			description: "Comparison in ternary without != 0",
		},
		{
			name: "numeric function in if condition",
			script: `//@version=4
study("Test", overlay=true)
len = 5
signal = 0.0
if dev(close, len)
    signal := 1
plot(signal)`,
			mustContain: []string{
				"value.IsTrue",
				"devSum := 0.0",
				"GetCurrent()",
			},
			mustNotContain: []string{
				"undefined:",
			},
			description: "dev() in if with != 0 conversion",
		},
		{
			name: "multiple inline functions with mixed types",
			script: `//@version=4
study("Test", overlay=true)
len = 5
avg = sma(close, 20)
dev_signal = dev(close, len) ? 1 : 0
comp_signal = close > avg ? 1 : 0
plot(dev_signal + comp_signal)`,
			mustContain: []string{
				"dev_signalSeries.Set",
				"comp_signalSeries.Set",
				"devSum := 0.0",
				"value.IsTrue",
			},
			mustNotContain: []string{
				"undefined:",
			},
			description: "Mixed numeric inline and comparison functions handled correctly",
		},
		{
			name: "nested ternary with inline functions",
			script: `//@version=4
study("Test", overlay=true)
len = 5
v1 = dev(close, len)
v2 = dev(open, len)
result = dev(close, len) ? (dev(open, len) ? 1 : 2) : 3
plot(result)`,
			mustContain: []string{
				"resultSeries.Set",
				"value.IsTrue",
				"devSum := 0.0",
			},
			mustNotContain: []string{
				"undefined:",
			},
			description: "Nested ternaries with dev() get != 0 at each level",
		},
		{
			name: "inline function with comparison operators in body",
			script: `//@version=4
study("Test", overlay=true)
len = 2
h = highest(len)
h1 = dev(h, len) ? na : h
plot(h1)`,
			mustContain: []string{
				"h1Series.Set",
				"value.IsTrue",
				"dev_",
				"GetCurrent()",
				"math.NaN()",
			},
			mustNotContain: []string{
				"undefined:",
			},
			description: "dev() function in ternary generates valid code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			parseResult, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(parseResult)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Code generation failed: %v", err)
			}

			code := result.FunctionBody

			for _, pattern := range tt.mustContain {
				if !contains(code, pattern) {
					t.Errorf("%s\nMissing pattern: %q\nGenerated code length: %d bytes",
						tt.description, pattern, len(code))
				}
			}

			for _, pattern := range tt.mustNotContain {
				if contains(code, pattern) {
					t.Errorf("%s\nFound forbidden pattern: %q",
						tt.description, pattern)
				}
			}
		})
	}
}

/* TestInlineFunctionsWithSeriesAccess validates inline functions containing Series.Get() patterns */
func TestInlineFunctionsWithSeriesAccess(t *testing.T) {
	script := `//@version=4
study("Test", overlay=true)
len = 10
avg = sma(close, len)
signal = avg > close ? 1 : 0
plot(signal)`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	parseResult, err := p.ParseBytes("test.pine", []byte(script))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := parser.NewConverter()
	program, err := converter.ToESTree(parseResult)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	result, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}

	code := result.FunctionBody

	// sma() inline generation should use Series.Get(j) for historical access
	if !strings.Contains(code, "Series.Get(") {
		t.Error("Expected Series.Get() pattern for historical access in inline sma")
	}

	// avg variable should be stored in Series
	if !strings.Contains(code, "avgSeries") {
		t.Error("Expected avgSeries variable declaration")
	}

	// Comparison should not add != 0 (already boolean)
	if strings.Contains(code, "value.IsTrue(avgSeries.GetCurrent() > bar.Close)") {
		t.Error("Comparison expression should not get != 0 conversion")
	}
}

/* TestInlineFunctionsEdgeCases validates boundary and error conditions */
func TestInlineFunctionsEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		shouldError bool
		description string
	}{
		{
			name: "minimal period handled",
			script: `//@version=4
study("Test", overlay=true)
result = dev(close, 1) ? 1 : 0
plot(result)`,
			shouldError: false,
			description: "Minimal period (1) should work correctly",
		},
		{
			name: "positive period handled",
			script: `//@version=4
study("Test", overlay=true)
result = dev(close, 10) ? 1 : 0
plot(result)`,
			shouldError: false,
			description: "Positive period should work correctly",
		},
		{
			name: "inline function with variable period",
			script: `//@version=4
study("Test", overlay=true)
len = input(5, title="Length")
result = dev(close, len) ? 1 : 0
plot(result)`,
			shouldError: true,
			description: "Variable period should fail with inline dev() - requires compile-time constant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			parseResult, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				if !tt.shouldError {
					t.Fatalf("Parse failed unexpectedly: %v", err)
				}
				return
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(parseResult)
			if err != nil {
				if !tt.shouldError {
					t.Fatalf("Conversion failed unexpectedly: %v", err)
				}
				return
			}

			_, err = GenerateStrategyCodeFromAST(program)
			if tt.shouldError && err == nil {
				t.Error("Expected error but code generation succeeded")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Code generation failed unexpectedly: %v", err)
			}
		})
	}
}

/* TestInlineFunctionsCompilability validates generated code compiles */
func TestInlineFunctionsCompilability(t *testing.T) {
	script := `//@version=4
strategy(title="Inline Functions Test", overlay=true)

len = 5
h = highest(len)
l = lowest(len)
h1 = dev(h, len) ? na : h
l1 = dev(l, len) ? na : l

fast = sma(close, 10)
slow = sma(close, 20)
cross = ta.crossover(fast, slow)

signal = cross and not na(h1) ? 1 : 0

if signal == 1
    strategy.entry("Long", strategy.long)

plot(h1, title="H1", color=color.red)
plot(l1, title="L1", color=color.blue)`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	parseResult, err := p.ParseBytes("test.pine", []byte(script))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := parser.NewConverter()
	program, err := converter.ToESTree(parseResult)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	result, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}

	code := result.FunctionBody

	criticalPatterns := []string{
		"h1Series.Set",
		"l1Series.Set",
		"signalSeries.Set",
		"value.IsTrue",
		"math.NaN()",
		"devSum := 0.0",
		"Series.Get(",
		"Crossover",
		"strat.Entry",
	}

	for _, pattern := range criticalPatterns {
		if !strings.Contains(code, pattern) {
			t.Errorf("Generated code missing critical pattern: %q", pattern)
		}
	}

	invalidPatterns := []string{
		"undefined:",
		"compile error",
		"syntax error",
	}

	for _, pattern := range invalidPatterns {
		if strings.Contains(code, pattern) {
			t.Errorf("Generated code contains invalid pattern: %q", pattern)
		}
	}
}

func TestInlineFunctionsTypeConsistency(t *testing.T) {
	tests := []struct {
		name               string
		script             string
		mustHaveConversion bool
		functionType       string
	}{
		{
			name: "numeric inline functions need conversion",
			script: `//@version=4
study("Test", overlay=true)
len = 5
s1 = dev(close, len) ? 1 : 0
s2 = change(close) ? 1 : 0
plot(s1 + s2)`,
			mustHaveConversion: true,
			functionType:       "numeric",
		},
		{
			name: "boolean functions skip conversion",
			script: `//@version=4
study("Test", overlay=true)
fast = sma(close, 10)
slow = sma(close, 20)
cross_up = ta.crossover(fast, slow)
cross_down = ta.crossunder(fast, slow)
b1 = cross_up ? 1 : 0
b2 = cross_down ? 1 : 0
b3 = na(close) ? 1 : 0
plot(b1 + b2 + b3)`,
			mustHaveConversion: true,
			functionType:       "boolean",
		},
		{
			name: "comparison expressions skip conversion",
			script: `//@version=4
study("Test", overlay=true)
c1 = close > open ? 1 : 0
c2 = high < low ? 1 : 0
c3 = close == open ? 1 : 0
c4 = high != low ? 1 : 0
plot(c1 + c2 + c3 + c4)`,
			mustHaveConversion: false,
			functionType:       "comparison",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			parseResult, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(parseResult)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Code generation failed: %v", err)
			}

			code := result.FunctionBody
			hasConversion := strings.Contains(code, "value.IsTrue")

			if tt.mustHaveConversion && !hasConversion {
				t.Errorf("%s functions should have != 0 conversion but none found",
					tt.functionType)
			}

			if !tt.mustHaveConversion && tt.functionType == "boolean" {
				if strings.Contains(code, "value.IsTrue(cross_upSeries.GetCurrent())") ||
					strings.Contains(code, "value.IsTrue(cross_downSeries.GetCurrent())") {
					t.Errorf("boolean variables should not have != 0 conversion")
				}
			}
		})
	}
}

func TestInlineFunctionsContextVariations(t *testing.T) {
	script := `//@version=4
study("Test", overlay=true)
len = 5
// Same function in different contexts
d = dev(close, len)
// Assignment - stores float64 value
result1 = d
// Ternary test - needs != 0
result2 = dev(close, len) ? 1 : 0
// If condition - needs != 0
result3 = 0
if dev(close, len)
    result3 := 1
// Logical AND - needs != 0
result4 = close > 100 and dev(close, len) ? 1 : 0
// Comparison - value used directly
result5 = dev(close, len) > 0.5 ? 1 : 0
plot(result1 + result2 + result3 + result4 + result5)`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	parseResult, err := p.ParseBytes("test.pine", []byte(script))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := parser.NewConverter()
	program, err := converter.ToESTree(parseResult)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	result, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}

	code := result.FunctionBody

	// Should have multiple != 0 conversions for conditional contexts
	conversionCount := strings.Count(code, "value.IsTrue")
	if conversionCount < 3 {
		t.Errorf("Expected at least 3 != 0 conversions for conditional contexts, found %d", conversionCount)
	}

	// Should have dev() calls
	if !strings.Contains(code, "devSum := 0.0") {
		t.Error("Expected dev() inline function generation")
	}

	// Assignment context should store value without conversion
	if !strings.Contains(code, "dSeries.Set") {
		t.Error("Expected d variable assignment")
	}
}

/* TestTAFunctionsInTernaryBranches validates TA function calls within ternary expression branches */
func TestTAFunctionsInTernaryBranches(t *testing.T) {
	tests := []struct {
		name           string
		script         string
		mustContain    []string
		mustNotContain []string
		description    string
	}{
		{
			name: "single TA function in ternary branches",
			script: `//@version=5
indicator("Test")
period = input.int(10)
mode = input.int(1)
ma = mode == 1 ? ta.sma(close, period) : ta.ema(close, period)
plot(ma)`,
			mustContain: []string{
				"maSeries.Set",
				"func() float64 {",
				".GetCurrent()",
			},
			mustNotContain: []string{
				"unsupported inline function",
				"undefined:",
			},
			description: "TA functions in ternary branches generate successfully",
		},
		{
			name: "nested ternary with multiple TA functions",
			script: `//@version=5
indicator("Test")
period = input.int(10)
mode = input.int(1)
ma = mode == 1 ? ta.sma(close, period) : (mode == 2 ? ta.ema(close, period) : ta.rma(close, period))
plot(ma)`,
			mustContain: []string{
				"maSeries.Set",
				"func() float64 {",
				".GetCurrent()",
			},
			mustNotContain: []string{
				"unsupported inline function",
			},
			description: "Nested ternaries with TA functions handle all branches",
		},
		{
			name: "mixed TA and builtin in ternary",
			script: `//@version=5
indicator("Test")
useTA = input.bool(true)
val = useTA ? ta.sma(close, 10) : close
plot(val)`,
			mustContain: []string{
				"valSeries.Set",
				"func() float64 {",
			},
			mustNotContain: []string{
				"unsupported inline function",
			},
			description: "Ternary with TA function and builtin series",
		},
		{
			name: "TA functions with different argument types",
			script: `//@version=5
indicator("Test")
src = input.source(close)
len1 = input.int(10)
len2 = input.int(20)
mode = input.bool(true)
result = mode ? ta.sma(src, len1) : ta.ema(src, len2)
plot(result)`,
			mustContain: []string{
				"resultSeries.Set",
				".GetCurrent()",
			},
			mustNotContain: []string{
				"unsupported inline function",
			},
			description: "TA functions with input parameters in ternary",
		},
		{
			name: "chained ternaries with TA functions",
			script: `//@version=5
indicator("Test")
period = input.int(10)
ma1 = ta.sma(close, period)
ma2 = ta.ema(close, period)
condition1 = input.bool(true)
condition2 = input.bool(false)
result = condition1 ? ma1 : condition2 ? ma2 : close
plot(result)`,
			mustContain: []string{
				"resultSeries.Set",
				"func() float64 {",
			},
			mustNotContain: []string{
				"undefined:",
			},
			description: "Chained ternaries reference TA-derived series",
		},
		{
			name: "TA function in ternary test condition",
			script: `//@version=5
indicator("Test")
period = input.int(10)
vol_ma = ta.sma(volume, period)
signal = volume > vol_ma ? 1.0 : 0.0
plot(signal)`,
			mustContain: []string{
				"signalSeries.Set",
				"vol_maSeries.GetCurrent()",
			},
			mustNotContain: []string{
				"unsupported inline function",
			},
			description: "TA function result used in ternary test condition",
		},
		{
			name: "TA functions with constants in ternary",
			script: `//@version=5
indicator("Test")
useDefault = input.bool(true)
ma = useDefault ? ta.sma(close, 20) : ta.ema(close, 10)
plot(ma)`,
			mustContain: []string{
				"maSeries.Set",
				".GetCurrent()",
			},
			mustNotContain: []string{
				"unsupported inline function",
			},
			description: "TA functions with constant periods in ternary",
		},
		{
			name: "complex expression with TA in ternary",
			script: `//@version=5
indicator("Test")
period = input.int(10)
multiplier = input.float(2.0)
mode = input.bool(true)
result = mode ? ta.sma(close, period) * multiplier : ta.ema(close, period) / multiplier
plot(result)`,
			mustContain: []string{
				"resultSeries.Set",
				".GetCurrent()",
			},
			mustNotContain: []string{
				"unsupported inline function",
			},
			description: "TA functions in arithmetic expressions within ternary branches",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			parseResult, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(parseResult)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Code generation failed: %v\nScript: %s", err, tt.script)
			}

			code := result.FunctionBody

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("%s\nMissing pattern: %q\nGenerated code length: %d bytes",
						tt.description, pattern, len(code))
				}
			}

			for _, pattern := range tt.mustNotContain {
				if strings.Contains(code, pattern) {
					t.Errorf("%s\nFound forbidden pattern: %q",
						tt.description, pattern)
				}
			}
		})
	}
}

/* TestTAFunctionsInTernaryEdgeCases validates boundary conditions and error handling */
func TestTAFunctionsInTernaryEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		shouldError bool
		description string
	}{
		{
			name: "ternary with same TA function different params",
			script: `//@version=5
indicator("Test")
len1 = input.int(10)
len2 = input.int(20)
mode = input.bool(true)
ma = mode ? ta.sma(close, len1) : ta.sma(close, len2)
plot(ma)`,
			shouldError: false,
			description: "Same TA function with different period parameters",
		},
		{
			name: "ternary with ta.crossover result",
			script: `//@version=5
indicator("Test")
fast = ta.sma(close, 10)
slow = ta.sma(close, 20)
cross_up = ta.crossover(fast, slow)
cross_down = ta.crossunder(fast, slow)
signal = cross_up ? 1.0 : cross_down ? -1.0 : 0.0
plot(signal)`,
			shouldError: false,
			description: "Ternary with boolean TA functions (crossover/crossunder)",
		},
		{
			name: "ternary with ta.change in test",
			script: `//@version=5
indicator("Test")
delta = ta.change(close)
direction = delta > 0 ? 1.0 : delta < 0 ? -1.0 : 0.0
plot(direction)`,
			shouldError: false,
			description: "ta.change result used in nested ternary conditions",
		},
		{
			name: "deeply nested ternary with TA functions",
			script: `//@version=5
indicator("Test")
mode = input.int(1)
period = input.int(10)
ma = mode == 1 ? ta.sma(close, period) : 
     mode == 2 ? ta.ema(close, period) : 
     mode == 3 ? ta.rma(close, period) : 
     ta.wma(close, period)
plot(ma)`,
			shouldError: false,
			description: "Deeply nested ternary (4 levels) with different TA functions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			parseResult, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				if !tt.shouldError {
					t.Fatalf("Parse failed unexpectedly: %v", err)
				}
				return
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(parseResult)
			if err != nil {
				if !tt.shouldError {
					t.Fatalf("Conversion failed unexpectedly: %v", err)
				}
				return
			}

			_, err = GenerateStrategyCodeFromAST(program)
			if tt.shouldError && err == nil {
				t.Errorf("%s: Expected error but code generation succeeded", tt.description)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("%s: Code generation failed unexpectedly: %v", tt.description, err)
			}
		})
	}
}
