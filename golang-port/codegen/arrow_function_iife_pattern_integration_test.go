package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

/* TestArrowFunctionIIFEPattern_Integration validates IIFE generation for TA calls in arrow functions */
func TestArrowFunctionIIFEPattern_Integration(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		mustContain    []string
		mustNotContain []string
	}{
		{
			name: "change() with default offset",
			source: `
dirmov(len) =>
    up = change(high)
    down = -change(low)
    [up, down]

[x, y] = dirmov(5)
`,
			mustContain: []string{
				"func dirmov(ctx *Context, len float64)",
				"func() float64",
				"current := ",
				"previous := ",
				"return current - previous",
				"ctx.BarIndex < 1",
			},
		},
		{
			name: "change() with custom offset",
			source: `
customChange(src) =>
    change(src, 2)

result = customChange(close)
`,
			mustContain: []string{
				"func customChange(ctx *Context, src float64)",
				"ctx.BarIndex < 2",
				"func() float64",
			},
		},
		{
			name: "unary expression with change()",
			source: `
negChange(src) =>
    -change(src)

result = negChange(low)
`,
			mustContain: []string{
				"func negChange(ctx *Context, src float64)",
				"-func() float64",
				"return current - previous",
			},
		},
		{
			name: "multiple change() calls",
			source: `
spread(len) =>
    highChange = change(high)
    lowChange = change(low)
    highChange - lowChange

result = spread(10)
`,
			mustContain: []string{
				"func spread(ctx *Context, len float64)",
				"highChange :=",
				"lowChange :=",
			},
		},
		{
			name: "ta.change prefix",
			source: `
indicator(period) =>
    ta.change(close, period)

result = indicator(14)
`,
			mustContain: []string{
				"func indicator(ctx *Context, period float64)",
				"func() float64",
				"current := ",
				"previous := ",
			},
		},
		{
			name: "change() in complex expression",
			source: `
momentum(len) =>
    upMove = change(high) > 0 ? change(high) : 0
    upMove

result = momentum(14)
`,
			mustContain: []string{
				"func momentum(ctx *Context, len float64)",
				"upMove :=",
			},
		},
		{
			name: "mixed TA functions",
			source: `
composite(len) =>
    chg = change(close)
    avg = sma(close, len)
    chg + avg

result = composite(20)
`,
			mustContain: []string{
				"func composite(ctx *Context, len float64)",
				"chg :=",
				"avg :=",
				"sum := 0.0",
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

			for _, want := range tt.mustContain {
				if !strings.Contains(code.UserDefinedFunctions, want) {
					t.Errorf("Missing pattern %q", want)
				}
			}

			for _, notWant := range tt.mustNotContain {
				if strings.Contains(code.UserDefinedFunctions, notWant) {
					t.Errorf("Unexpected pattern %q", notWant)
				}
			}
		})
	}
}

/* TestArrowFunctionIIFEPattern_WarmupBehavior validates warmup check generation */
func TestArrowFunctionIIFEPattern_WarmupBehavior(t *testing.T) {
	tests := []struct {
		name            string
		source          string
		expectedWarmup  string
		unexpectedCheck string
	}{
		{
			name: "offset 1 warmup",
			source: `
indicator() =>
    change(close, 1)

result = indicator()
`,
			expectedWarmup: "ctx.BarIndex < 1",
		},
		{
			name: "offset 10 warmup",
			source: `
indicator() =>
    change(close, 10)

result = indicator()
`,
			expectedWarmup: "ctx.BarIndex < 10",
		},
		{
			name: "default offset warmup",
			source: `
indicator() =>
    change(high)

result = indicator()
`,
			expectedWarmup:  "ctx.BarIndex < 1",
			unexpectedCheck: "ctx.BarIndex < 0",
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

			if !strings.Contains(code.UserDefinedFunctions, tt.expectedWarmup) {
				t.Errorf("Missing expected warmup check %q", tt.expectedWarmup)
			}

			if tt.unexpectedCheck != "" && strings.Contains(code.UserDefinedFunctions, tt.unexpectedCheck) {
				t.Errorf("Found unexpected warmup check %q", tt.unexpectedCheck)
			}

			if !strings.Contains(code.UserDefinedFunctions, "math.NaN()") {
				t.Error("Missing NaN return for warmup period")
			}
		})
	}
}
