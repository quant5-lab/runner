package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

func TestArrowFunctionIIFE_RuntimePeriodWarmupGuards(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		mustContain    []string
		mustNotContain []string
		minCount       map[string]int
	}{
		{
			name: "SMA with runtime period parameter",
			source: `
//@version=5
strategy("SMA Runtime")
avg(len) => ta.sma(close, len)
result = avg(14)
`,
			mustContain: []string{
				"func avg(arrowCtx *context.ArrowContext, len float64) float64",
				"if ctx.BarIndex < int(len)-1 { return math.NaN() }",
				"for j := 0; j < int(len); j++",
			},
			minCount: map[string]int{
				"if ctx.BarIndex < int(len)-1": 1,
			},
		},
		{
			name: "Multiple TA functions with same runtime period",
			source: `
//@version=5
strategy("Multi TA Runtime")
indicator(length) =>
    avg = ta.sma(close, length)
    dev = ta.stdev(close, length)
    avg + dev
result = indicator(20)
`,
			mustContain: []string{
				"func indicator(arrowCtx *context.ArrowContext, length float64) float64",
				"if ctx.BarIndex < int(length)-1 { return math.NaN() }",
			},
			minCount: map[string]int{
				"if ctx.BarIndex < int(length)-1": 2,
			},
		},
		{
			name: "Highest/Lowest with runtime period",
			source: `
//@version=5
strategy("Range Runtime")
calcRange(bars) =>
    high_val = ta.highest(high, bars)
    low_val = ta.lowest(low, bars)
    high_val - low_val
r = calcRange(10)
`,
			mustContain: []string{
				"func calcRange(arrowCtx *context.ArrowContext, bars float64) float64",
				"if ctx.BarIndex < int(bars)-1 { return math.NaN() }",
				"periodVal := int(bars)",
			},
			minCount: map[string]int{
				"if ctx.BarIndex < int(bars)-1": 2,
				"periodVal := int(bars)":        2,
			},
		},
		{
			name: "Change with runtime offset",
			source: `
//@version=5
strategy("Change Runtime")
delta(offset) => ta.change(close, offset)
d = delta(5)
`,
			mustContain: []string{
				"func delta(arrowCtx *context.ArrowContext, offset float64) float64",
				"if ctx.BarIndex < (int(offset)+1)-1 { return math.NaN() }",
			},
		},
		{
			name: "Linreg with runtime period",
			source: `
//@version=5
strategy("Linreg Runtime")
trend(len) => ta.linreg(close, len)
t = trend(20)
`,
			mustContain: []string{
				"func trend(arrowCtx *context.ArrowContext, len float64) float64",
				"if ctx.BarIndex < int(len)-1 { return math.NaN() }",
				"periodVal := int(len)",
			},
		},
		{
			name: "WMA with runtime period",
			source: `
//@version=5
strategy("WMA Runtime")
wavg(period) => ta.wma(close, period)
w = wavg(10)
`,
			mustContain: []string{
				"func wavg(arrowCtx *context.ArrowContext, period float64) float64",
				"if ctx.BarIndex < int(period)-1 { return math.NaN() }",
				"for j := 0; j < int(period); j++",
			},
		},
		{
			name: "Multiple parameters with runtime period",
			source: `
//@version=5
strategy("Multi Param")
bband(length, mult) =>
    basis = ta.sma(close, length)
    dev = ta.stdev(close, length)
    basis + mult * dev
upper = bband(20, 2)
`,
			mustContain: []string{
				"func bband(arrowCtx *context.ArrowContext, length float64, mult float64) float64",
				"if ctx.BarIndex < int(length)-1 { return math.NaN() }",
			},
			minCount: map[string]int{
				"if ctx.BarIndex < int(length)-1": 2,
			},
		},
		{
			name: "Keltner-Squeeze pattern (Issue #22 original)",
			source: `
//@version=5
strategy("Keltner Squeeze")

bband(length, mult) =>
    basis = ta.sma(close, length)
    dev = ta.stdev(close, length)
    [basis - mult * dev, basis, basis + mult * dev]

keltner(length, mult) =>
    ema_val = ta.ema(close, length)
    ema_tr = ta.ema(ta.tr, length)
    [ema_val - mult * ema_tr, ema_val, ema_val + mult * ema_tr]

[bbLower, bbBasis, bbUpper] = bband(20, 2)
[kLower, kBasis, kUpper] = keltner(20, 1.5)
`,
			mustContain: []string{
				"func bband(arrowCtx *context.ArrowContext, length float64, mult float64)",
				"func keltner(arrowCtx *context.ArrowContext, length float64, mult float64)",
				"if ctx.BarIndex < int(length)-1 { return math.NaN() }",
			},
			minCount: map[string]int{
				"if ctx.BarIndex < int(length)-1": 4,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			code, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code.UserDefinedFunctions, pattern) {
					maxLen := len(code.UserDefinedFunctions)
					if maxLen > 2000 {
						maxLen = 2000
					}
					t.Errorf("Missing pattern: %q\nGenerated code:\n%s",
						pattern, code.UserDefinedFunctions[:maxLen])
				}
			}

			for _, pattern := range tt.mustNotContain {
				if strings.Contains(code.UserDefinedFunctions, pattern) {
					t.Errorf("Should NOT contain: %q", pattern)
				}
			}

			for pattern, minCount := range tt.minCount {
				count := strings.Count(code.UserDefinedFunctions, pattern)
				if count < minCount {
					t.Errorf("Pattern %q found %d times, expected at least %d",
						pattern, count, minCount)
				}
			}
		})
	}
}

func TestArrowFunctionIIFE_ConstantVsRuntimePeriod(t *testing.T) {
	tests := []struct {
		name              string
		source            string
		constantWarmup    string
		runtimeWarmup     string
		shouldHaveConst   bool
		shouldHaveRuntime bool
	}{
		{
			name: "Constant period only",
			source: `
//@version=5
strategy("Constant")
avg() => ta.sma(close, 20)
result = avg()
`,
			constantWarmup:    "if ctx.BarIndex < 19 { return math.NaN() }",
			shouldHaveConst:   true,
			shouldHaveRuntime: false,
		},
		{
			name: "Runtime period only",
			source: `
//@version=5
strategy("Runtime")
avg(len) => ta.sma(close, len)
result = avg(14)
`,
			runtimeWarmup:     "if ctx.BarIndex < int(len)-1 { return math.NaN() }",
			shouldHaveConst:   false,
			shouldHaveRuntime: true,
		},
		{
			name: "Mixed constant and runtime",
			source: `
//@version=5
strategy("Mixed")
indicator(dynamic) =>
    constant_sma = ta.sma(close, 10)
    runtime_sma = ta.sma(close, dynamic)
    constant_sma + runtime_sma
result = indicator(20)
`,
			constantWarmup:    "if ctx.BarIndex < 9 { return math.NaN() }",
			runtimeWarmup:     "if ctx.BarIndex < int(dynamic)-1 { return math.NaN() }",
			shouldHaveConst:   true,
			shouldHaveRuntime: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			code, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			hasConst := tt.constantWarmup != "" && strings.Contains(code.UserDefinedFunctions, tt.constantWarmup)
			hasRuntime := tt.runtimeWarmup != "" && strings.Contains(code.UserDefinedFunctions, tt.runtimeWarmup)

			if tt.shouldHaveConst && !hasConst && tt.constantWarmup != "" {
				t.Errorf("Expected constant warmup: %q", tt.constantWarmup)
			}
			if !tt.shouldHaveConst && hasConst {
				t.Errorf("Should NOT have constant warmup: %q", tt.constantWarmup)
			}
			if tt.shouldHaveRuntime && !hasRuntime && tt.runtimeWarmup != "" {
				t.Errorf("Expected runtime warmup: %q", tt.runtimeWarmup)
			}
			if !tt.shouldHaveRuntime && hasRuntime {
				t.Errorf("Should NOT have runtime warmup: %q", tt.runtimeWarmup)
			}
		})
	}
}

func TestArrowFunctionIIFE_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		mustContain []string
	}{
		{
			name: "Minimum period (1)",
			source: `
//@version=5
strategy("Min Period")
single(p) => ta.sma(close, p)
s = single(1)
`,
			mustContain: []string{
				"if ctx.BarIndex < int(p)-1 { return math.NaN() }",
			},
		},
		{
			name: "Large period value",
			source: `
//@version=5
strategy("Large Period")
longAvg(len) => ta.sma(close, len)
l = longAvg(200)
`,
			mustContain: []string{
				"if ctx.BarIndex < int(len)-1 { return math.NaN() }",
				"for j := 0; j < int(len); j++",
			},
		},
		{
			name: "Single character parameter",
			source: `
//@version=5
strategy("Single Char")
f(n) => ta.sma(close, n)
result = f(5)
`,
			mustContain: []string{
				"func f(arrowCtx *context.ArrowContext, n float64) float64",
				"if ctx.BarIndex < int(n)-1 { return math.NaN() }",
			},
		},
		{
			name: "Nested arrow functions with runtime periods",
			source: `
//@version=5
strategy("Nested")
inner(len) => ta.sma(close, len)
outer(period) => inner(period) * 2
result = outer(10)
`,
			mustContain: []string{
				"func inner(arrowCtx *context.ArrowContext, len float64) float64",
				"func outer(arrowCtx *context.ArrowContext, period float64) float64",
				"if ctx.BarIndex < int(len)-1 { return math.NaN() }",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			code, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code.UserDefinedFunctions, pattern) {
					maxLen := len(code.UserDefinedFunctions)
					if maxLen > 1500 {
						maxLen = 1500
					}
					t.Errorf("Missing pattern: %q\nGenerated:\n%s",
						pattern, code.UserDefinedFunctions[:maxLen])
				}
			}
		})
	}
}
