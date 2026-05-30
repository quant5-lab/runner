package codegen

import (
	"slices"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

func TestDeduplicateFeatureGaps(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{name: "nil input returns nil", input: nil, want: nil},
		{name: "empty slice returns nil", input: []string{}, want: nil},
		{name: "single entry", input: []string{"ta.kc"}, want: []string{"ta.kc"}},
		{name: "all duplicates collapsed to one", input: []string{"foo", "foo", "foo"}, want: []string{"foo"}},
		{name: "all unique preserved", input: []string{"a", "b", "c"}, want: []string{"a", "b", "c"}},
		{name: "first-seen order preserved", input: []string{"z", "a", "m", "a", "z"}, want: []string{"z", "a", "m"}},
		{
			name:  "namespaced functions deduplicated",
			input: []string{"request.dividends", "ta.kc", "request.dividends", "ta.kc"},
			want:  []string{"request.dividends", "ta.kc"},
		},
		{
			name:  "case-sensitive deduplication",
			input: []string{"Foo", "foo", "FOO"},
			want:  []string{"Foo", "foo", "FOO"},
		},
		{
			name:  "large input with many duplicates",
			input: repeat([]string{"x", "y", "z"}, 100),
			want:  []string{"x", "y", "z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deduplicateFeatureGaps(tt.input)
			if tt.want == nil {
				if got != nil {
					t.Errorf("deduplicateFeatureGaps(%v) = %v, want nil", tt.input, got)
				}
				return
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("deduplicateFeatureGaps(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestStrategyCode_FeatureGapsEmpty(t *testing.T) {
	script := `//@version=5
strategy("Test", overlay=true)
x = ta.sma(close, 14)
plot(x, "sma")
`
	code := mustGenerateFromPine(t, script)
	if len(code.FeatureGaps) != 0 {
		t.Errorf("FeatureGaps = %v for fully-implemented script, want empty", code.FeatureGaps)
	}
}

func TestStrategyCode_FeatureGapsContainsUnknown(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantGap string
	}{
		{
			name: "bare unknown in statement position",
			script: `//@version=5
indicator("Test")
unimplemented_func(close)
plot(close, "c")
`,
			wantGap: "unimplemented_func",
		},
		{
			name: "namespaced unknown in series assignment",
			script: `//@version=5
indicator("Test")
x = request.dividends("AAPL")
plot(x, "d")
`,
			wantGap: "request.dividends",
		},
		{
			name: "unknown in if-condition expression position",
			script: `//@version=5
indicator("Test")
result = 0.0
if exprPosGapFunc(close)
    result := 1.0
plot(result, "r")
`,
			wantGap: "exprPosGapFunc",
		},
		{
			name: "unknown in arrow function body expression position",
			script: `//@version=5
indicator("Test")
f(src) => arrowPosGapFunc(src)
plot(f(close), "f")
`,
			wantGap: "arrowPosGapFunc",
		},
		{
			name: "unknown as direct plot() argument",
			script: `//@version=5
indicator("Test")
plot(directPlotGapFunc(close), "x")
`,
			wantGap: "directPlotGapFunc",
		},
		{
			name: "unknown as binary expression operand in variable init",
			script: `//@version=5
indicator("Test")
x = close + featureGapsBinaryOpFunc(close)
plot(x, "x")
`,
			wantGap: "featureGapsBinaryOpFunc",
		},
		{
			name: "unknown as TA indicator source argument",
			script: `//@version=5
indicator("Test")
x = ta.sma(featureGapsTaSourceFunc(close), 14)
plot(x, "x")
`,
			wantGap: "featureGapsTaSourceFunc",
		},
		{
			name: "unknown nested inside binary expression used as TA source",
			script: `//@version=5
indicator("Test")
x = ta.rsi(featureGapsBinaryTaSourceFunc(close) + close, 14)
plot(x, "x")
`,
			wantGap: "featureGapsBinaryTaSourceFunc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := mustGenerateFromPine(t, tt.script)
			if !containsGap(code.FeatureGaps, tt.wantGap) {
				t.Errorf("FeatureGaps = %v, want entry %q", code.FeatureGaps, tt.wantGap)
			}
		})
	}
}

func TestStrategyCode_FeatureGapsDeduplicated(t *testing.T) {
	script := `//@version=5
indicator("Test")
a = unimplemented_func(close)
b = unimplemented_func(open)
c = unimplemented_func(high)
plot(close, "c")
`
	code := mustGenerateFromPine(t, script)

	count := 0
	for _, g := range code.FeatureGaps {
		if g == "unimplemented_func" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("FeatureGaps contains %q %d time(s), want exactly 1; gaps=%v", "unimplemented_func", count, code.FeatureGaps)
	}
}

func TestStrategyCode_FeatureGapsMultipleDistinct(t *testing.T) {
	script := `//@version=5
indicator("Test")
a = func_alpha(close)
b = func_beta(open)
c = func_gamma(high)
plot(close, "c")
`
	code := mustGenerateFromPine(t, script)

	for _, want := range []string{"func_alpha", "func_beta", "func_gamma"} {
		if !containsGap(code.FeatureGaps, want) {
			t.Errorf("FeatureGaps = %v, missing %q", code.FeatureGaps, want)
		}
	}
}

func TestStrategyCode_FeatureGapsKnownFunctionsExcluded(t *testing.T) {
	knownFunctions := []string{"ta.sma", "ta.ema", "ta.rsi", "plot", "math.abs"}

	script := `//@version=5
indicator("Test")
s = ta.sma(close, 14)
e = ta.ema(close, 14)
r = ta.rsi(close, 14)
a = math.abs(close - open)
plot(s, "sma")
plot(e, "ema")
`
	code := mustGenerateFromPine(t, script)

	for _, known := range knownFunctions {
		if containsGap(code.FeatureGaps, known) {
			t.Errorf("FeatureGaps contains known-implemented function %q; gaps=%v", known, code.FeatureGaps)
		}
	}
}

func TestStrategyCode_VoidBuiltinsExcludedFromFeatureGaps(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{
			name: "alert in top-level statement position",
			script: `//@version=5
strategy("Test", overlay=true)
alert("signal text")
`,
		},
		{
			name: "alert with freq argument in if-branch",
			script: `//@version=5
strategy("Test", overlay=true)
if close > open
    alert("long signal", alert.freq_once_per_bar_close)
`,
		},
		{
			name: "alert with concatenated string argument in if-branch",
			script: `//@version=5
strategy("Test", overlay=true)
if close > open
    strategy.entry("Long", strategy.long)
    alert(syminfo.tickerid + " Long Signal", alert.freq_once_per_bar_close)
`,
		},
		{
			name: "alertcondition in top-level statement position",
			script: `//@version=5
indicator("Test", overlay=true)
alertcondition(close > open, "title", "msg")
`,
		},
		{
			name: "alertcondition with complex condition",
			script: `//@version=5
indicator("Test", overlay=true)
s = ta.sma(close, 14)
alertcondition(close > s, "SMA Cross", "Price crossed SMA")
plot(s, "sma")
`,
		},
		{
			name: "alert and alertcondition coexist without gaps",
			script: `//@version=5
indicator("Test", overlay=true)
alertcondition(close > open, "up", "price up")
if close > open
    alert("price up signal")
plot(close, "c")
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := mustGenerateFromPine(t, tt.script)
			for _, gap := range code.FeatureGaps {
				if gap == "alert" || gap == "alertcondition" {
					t.Errorf("FeatureGaps contains void builtin %q; all gaps=%v", gap, code.FeatureGaps)
				}
			}
		})
	}
}

func mustGenerateFromPine(t *testing.T, source string) *StrategyCode {
	t.Helper()
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}
	script, err := p.ParseBytes("test.pine", []byte(source))
	if err != nil {
		t.Fatalf("ParseBytes: %v", err)
	}
	conv := parser.NewConverter()
	program, err := conv.ToESTree(script)
	if err != nil {
		t.Fatalf("ToESTree: %v", err)
	}
	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST: %v", err)
	}
	return code
}

func containsGap(gaps []string, name string) bool {
	for _, g := range gaps {
		if g == name {
			return true
		}
	}
	return false
}

func repeat(items []string, times int) []string {
	out := make([]string, 0, len(items)*times)
	for range times {
		out = append(out, items...)
	}
	return out
}
