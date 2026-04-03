package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

/* TestArrowFunctionBBW_MultParameterHandling verifies BBW mult argument extraction and code generation */
func TestArrowFunctionBBW_MultParameterHandling(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		mustContain    []string
		mustNotContain []string
	}{
		{
			name: "mult as runtime parameter uses identifier",
			source: `
calcBBW(src, len, mult) => ta.bbw(src, len, mult)
result = calcBBW(close, 20, 3.0)
`,
			mustContain: []string{
				"srcSeries *series.Series",
				"2.0 * mult * sd / sma",
			},
			mustNotContain: []string{
				"2.0 * 2.0",
				"2.0 * 3",
			},
		},
		{
			name: "mult as literal constant inlines value",
			source: `
calcBBW(src, len) => ta.bbw(src, len, 2.0)
result = calcBBW(close, 20)
`,
			mustContain: []string{
				"srcSeries *series.Series",
				"2.0 * 2",
			},
			mustNotContain: []string{
				"2.0 * mult",
			},
		},
		{
			name: "fractional mult literal preserved",
			source: `
calcBBW(src, len) => ta.bbw(src, len, 2.5)
result = calcBBW(close, 14)
`,
			mustContain: []string{
				"2.0 * 2.5",
			},
			mustNotContain: []string{
				"2.0 * 2 ",
			},
		},
		{
			name: "bare alias 2-arg extracts mult from arg[1]",
			source: `
calcBBW(len, mult) => bbw(len, mult)
result = calcBBW(10, 1.5)
`,
			mustContain: []string{
				"2.0 * mult * sd / sma",
				"ctx.Data[ctx.BarIndex-j].Close",
			},
		},
		{
			name: "bare alias 3-arg explicit source",
			source: `
width(src, period, mult) => bbw(src, period, mult)
w = width(close, 20, 2.8)
`,
			mustContain: []string{
				"2.0 * mult * sd / sma",
				"srcSeries.Get(j)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("parser creation failed: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("codegen failed: %v", err)
			}

			arrowCode := result.UserDefinedFunctions

			for _, pattern := range tt.mustContain {
				if !strings.Contains(arrowCode, pattern) {
					t.Errorf("arrow code missing pattern %q\nArrow functions:\n%s", pattern, arrowCode)
				}
			}

			for _, pattern := range tt.mustNotContain {
				if strings.Contains(arrowCode, pattern) {
					t.Errorf("arrow code should not contain pattern %q\nArrow functions:\n%s", pattern, arrowCode)
				}
			}
		})
	}
}

/* TestArrowFunctionBBW_AlgorithmInvariants verifies BBW formula structure */
func TestArrowFunctionBBW_AlgorithmInvariants(t *testing.T) {
	source := `
width(src, period) => ta.bbw(src, period, 3.0)
result = width(close, 20)
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("parser creation failed: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(source))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	converter := parser.NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	result, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}

	code := result.UserDefinedFunctions

	t.Run("SMA_calculation", func(t *testing.T) {
		if !strings.Contains(code, "sma := 0.0") {
			t.Error("BBW must initialize SMA accumulator")
		}
		if !strings.Contains(code, "sma /= float64(period)") {
			t.Error("BBW must compute mean for SMA")
		}
	})

	t.Run("standard_deviation_calculation", func(t *testing.T) {
		if !strings.Contains(code, "sd := 0.0") {
			t.Error("BBW must initialize SD accumulator")
		}
		if !strings.Contains(code, "d := ") && !strings.Contains(code, "- sma") {
			t.Error("BBW must compute deviations from SMA")
		}
		if !strings.Contains(code, "math.Sqrt") {
			t.Error("BBW must use sqrt for standard deviation")
		}
	})

	t.Run("zero_SMA_guard", func(t *testing.T) {
		if !strings.Contains(code, "if sma == 0.0 { return 0.0 }") {
			t.Error("BBW must guard against division by zero SMA")
		}
	})

	t.Run("bandwidth_formula", func(t *testing.T) {
		if !strings.Contains(code, "2.0 * 3 * sd / sma") {
			t.Error("BBW bandwidth must be: 2 * mult * stdev / sma")
		}
	})

	t.Run("warmup_check", func(t *testing.T) {
		if !strings.Contains(code, "ctx.BarIndex < int(period)") {
			t.Error("BBW must check warmup period")
		}
		if !strings.Contains(code, "return math.NaN()") {
			t.Error("BBW must return NaN during warmup")
		}
	})
}

/* TestArrowFunctionBBW_EdgeCases verifies boundary conditions */
func TestArrowFunctionBBW_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		mustContain []string
	}{
		{
			name: "period 1 no warmup needed",
			source: `
instant(src) => ta.bbw(src, 1, 2.0)
r = instant(close)
`,
			mustContain: []string{
				"for j := 0; j < 1; j++",
				"float64(1)",
			},
		},
		{
			name: "multiple BBW calls with different mults",
			source: `
compare(src) =>
    narrow = ta.bbw(src, 20, 1.0)
    wide = ta.bbw(src, 20, 3.0)
    [narrow, wide]
[n, w] = compare(close)
`,
			mustContain: []string{
				"narrowSeries := arrowCtx.GetOrCreateSeries(\"narrow\")",
				"wideSeries := arrowCtx.GetOrCreateSeries(\"wide\")",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("parser creation failed: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("codegen failed: %v", err)
			}

			code := result.UserDefinedFunctions

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("code missing pattern %q\nFull code:\n%s", pattern, code)
				}
			}
		})
	}
}
