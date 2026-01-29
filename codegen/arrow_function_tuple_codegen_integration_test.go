package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

/*
TestArrowFunctionTupleArgumentHandling_Integration validates complete codegen
for tuple indicators in arrow context, ensuring correct argument passing patterns:
- Series arguments receive accessor methods (GetCurrent())
- Period arguments pass through as literals or runtime variables
- Mixed argument types handled correctly
- All registered tuple functions generate valid code
*/
func TestArrowFunctionTupleArgumentHandling_Integration(t *testing.T) {
	tests := []struct {
		name           string
		pineScript     string
		mustContain    []string
		mustNotContain []string
		desc           string
	}{
		{
			name: "series_argument_accessor_generation",
			pineScript: `//@version=5
indicator('test')
f(src) => ta.macd(src, 12, 26, 9)
[macdLine, signal, hist] = f(close)
plot(macdLine)`,
			mustContain: []string{
				"func() (float64, float64, float64)",
				"ta.Macd(",
				"srcSeries.GetCurrent()",
				"12", "26", "9",
			},
			mustNotContain: []string{
				"ta.Macd(src,",
				"ta.Macd(close,",
			},
			desc: "Series parameter uses accessor method in IIFE",
		},
		{
			name: "bollinger_bands_argument_pattern",
			pineScript: `//@version=5
indicator('test')
f(src) => ta.bb(src, 20, 2)
[upper, basis, lower] = f(close)
plot(upper)`,
			mustContain: []string{
				"func() (float64, float64, float64)",
				"ta.BBands(",
				"srcSeries.GetCurrent()",
				"20", "2",
			},
			mustNotContain: []string{
				"ta.BBands(src,",
				"ta.BBands(close,",
			},
			desc: "Bollinger Bands series and period arguments handled correctly",
		},
		{
			name: "stochastic_bar_field_access",
			pineScript: `//@version=5
indicator('test')
f() => ta.stoch(close, high, low, 14)
[k, d] = f()
plot(k)`,
			mustContain: []string{
				"func() (float64, float64)",
				"ta.Stoch(",
				"bar.Close",
				"bar.High",
				"bar.Low",
				"14",
			},
			mustNotContain: []string{
				"closeSeries.GetCurrent()",
			},
			desc: "Functions without explicit source use bar field access",
		},
		{
			name: "period_only_arguments",
			pineScript: `//@version=5
indicator('test')
f() => ta.dmi(14, 14)
[plus, minus, adx] = f()
plot(plus)`,
			mustContain: []string{
				"func() (float64, float64, float64)",
				"ta.Dmi(",
				"14, 14",
			},
			mustNotContain: []string{
				"GetCurrent()",
			},
			desc: "Period-only functions pass literal arguments",
		},
		{
			name: "keltner_channels_mixed_arguments",
			pineScript: `//@version=5
indicator('test')
f(src) => ta.kc(src, 20, 2)
[upper, basis, lower] = f(close)
plot(upper)`,
			mustContain: []string{
				"func() (float64, float64, float64)",
				"ta.KeltnerChannels(",
				"srcSeries.GetCurrent()",
				"20", "2",
			},
			mustNotContain: []string{
				"ta.KeltnerChannels(src,",
			},
			desc: "Mixed series and period arguments distinguished",
		},
		{
			name: "supertrend_two_output_function",
			pineScript: `//@version=5
indicator('test')
f() => ta.supertrend(10, 3)
[supertrend, direction] = f()
plot(supertrend)`,
			mustContain: []string{
				"func() (float64, float64)",
				"ta.Supertrend(",
				"10, 3",
			},
			mustNotContain: []string{
				"GetCurrent()",
			},
			desc: "Two-output functions generate correct signature",
		},
		{
			name: "function_alias_without_namespace",
			pineScript: `//@version=5
indicator('test')
f(src) => macd(src, 12, 26, 9)
[macdLine, signal, hist] = f(close)
plot(macdLine)`,
			mustContain: []string{
				"func() (float64, float64, float64)",
				"ta.Macd(",
				"srcSeries.GetCurrent()",
			},
			mustNotContain: []string{
				"macd(src,",
			},
			desc: "Function aliases map to correct runtime calls",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			parseResult, err := p.ParseBytes("test.pine", []byte(tt.pineScript))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(parseResult)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			code, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Generation failed: %v", err)
			}

			generatedCode := code.UserDefinedFunctions + "\n" + code.FunctionBody

			for _, pattern := range tt.mustContain {
				if !strings.Contains(generatedCode, pattern) {
					t.Errorf("%s: missing expected pattern %q\nGenerated code:\n%s", tt.desc, pattern, generatedCode)
				}
			}

			for _, pattern := range tt.mustNotContain {
				if strings.Contains(generatedCode, pattern) {
					t.Errorf("%s: found forbidden pattern %q\nGenerated code:\n%s", tt.desc, pattern, generatedCode)
				}
			}
		})
	}
}
