package codegen

import (
	"os"
	"strings"
	"testing"
)

/* TestPlotExpressionHandler_MathBeforeTADispatch verifies math handler takes priority over IIFE TA
 * registry when bare names like max/min appear in plot() expressions.
 *
 * Regression guard for the dispatch ordering fix in handleCallExpression:
 * RegisterWithBareAlias registers both "ta.max" and "max" in InlineTAIIFERegistry, so without
 * explicit math-first ordering, plot(max(high,low)) would misroute to the TA series handler. */
func TestPlotExpressionHandler_MathBeforeTADispatch(t *testing.T) {
	tests := []struct {
		name        string
		pine        string
		mustContain []string
		mustExclude []string
	}{
		{
			name: "bare max routes to math not TA series",
			pine: `
//@version=5
indicator("Test")
plot(max(high, low), "m")
`,
			mustContain: []string{"math.Max(bar.High, bar.Low)"},
			mustExclude: []string{"maxSeries", "runtime dynamic period"},
		},
		{
			name: "bare min routes to math not TA series",
			pine: `
//@version=5
indicator("Test")
plot(min(high, low), "m")
`,
			mustContain: []string{"math.Min(bar.High, bar.Low)"},
			mustExclude: []string{"minSeries", "runtime dynamic period"},
		},
		{
			name: "bare abs routes to math",
			pine: `
//@version=5
indicator("Test")
plot(abs(close - open), "m")
`,
			mustContain: []string{"math.Abs("},
			mustExclude: []string{"absSeries"},
		},
		{
			name: "ta.max with period routes to IIFE TA handler",
			pine: `
//@version=5
indicator("Test")
x = ta.max(close, 10)
plot(x, "m")
`,
			mustContain: []string{"xSeries.Set(", "maxVal", "for j := 0"},
			mustExclude: []string{"math.Max(bar.Close, 10)"},
		},
		{
			name: "ta.min with period routes to IIFE TA handler",
			pine: `
//@version=5
indicator("Test")
x = ta.min(close, 10)
plot(x, "m")
`,
			mustContain: []string{"xSeries.Set(", "minVal", "for j := 0"},
			mustExclude: []string{"math.Min(bar.Close, 10)"},
		},
		{
			name: "math and ta coexist in same script without conflict",
			pine: `
//@version=5
indicator("Test")
taMax = ta.max(close, 10)
plot(max(high, low), "math")
plot(taMax, "ta")
`,
			mustContain: []string{
				"math.Max(bar.High, bar.Low)",
				"taMaxSeries.Set(",
				"maxVal",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Missing expected pattern %q\nGenerated:\n%s", pattern, code)
				}
			}

			for _, pattern := range tt.mustExclude {
				if strings.Contains(code, pattern) {
					t.Errorf("Unexpected pattern %q found\nGenerated:\n%s", pattern, code)
				}
			}
		})
	}
}

/* TestPlotExpressionHandler_MomentumFunctionsInPlot verifies all 9 new momentum TA functions
 * compile and route through the TA handler (not math handler) in plot-adjacent variable assignments. */
func TestPlotExpressionHandler_MomentumFunctionsInPlot(t *testing.T) {
	tests := []struct {
		name        string
		pine        string
		mustContain []string
		mustExclude []string
	}{
		{
			name: "ta.rising compiles and emits series assignment",
			pine: `
//@version=5
indicator("Test")
r = ta.rising(close, 3)
plot(r, "r")
`,
			mustContain: []string{"rSeries.Set(", "for j := 0; j < 3", "<= prev", "return 1.0"},
		},
		{
			name: "ta.falling compiles and emits series assignment",
			pine: `
//@version=5
indicator("Test")
f = ta.falling(close, 3)
plot(f, "f")
`,
			mustContain: []string{"fSeries.Set(", "for j := 0; j < 3", ">= prev", "return 1.0"},
		},
		{
			name: "ta.cross compiles with two series arguments",
			pine: `
//@version=5
indicator("Test")
s = ta.sma(close, 14)
c = ta.cross(close, s)
plot(c, "c")
`,
			mustContain: []string{"cSeries.Set("},
		},
		{
			name: "ta.highestbars returns negative offset",
			pine: `
//@version=5
indicator("Test")
hb = ta.highestbars(close, 5)
plot(hb, "hb")
`,
			mustContain: []string{"float64(-_hb_extIdx)", "hbSeries.Set("},
		},
		{
			name: "ta.lowestbars returns negative offset",
			pine: `
//@version=5
indicator("Test")
lb = ta.lowestbars(close, 5)
plot(lb, "lb")
`,
			mustContain: []string{"float64(-_lb_extIdx)", "lbSeries.Set("},
		},
		{
			name: "ta.mom emits difference formula",
			pine: `
//@version=5
indicator("Test")
m = ta.mom(close, 10)
plot(m, "m")
`,
			mustContain: []string{"mSeries.Set(closeSeries.Get(0) - closeSeries.Get(10))"},
		},
		{
			name: "ta.roc emits percent change with zero guard",
			pine: `
//@version=5
indicator("Test")
r = ta.roc(close, 12)
plot(r, "r")
`,
			mustContain: []string{"past == 0.0", "100.0", "rSeries.Set("},
		},
		{
			name: "ta.cmo emits up/down accumulation",
			pine: `
//@version=5
indicator("Test")
c = ta.cmo(close, 9)
plot(c, "c")
`,
			mustContain: []string{"_c_up", "_c_down", "100.0", "cSeries.Set("},
		},
		{
			name: "ta.wpr uses OHLC directly, not source series",
			pine: `
//@version=5
indicator("Test")
w = ta.wpr(14)
plot(w, "w")
`,
			mustContain: []string{"ctx.Data[ctx.BarIndex]", "100.0", "wSeries.Set("},
			mustExclude: []string{"closeSeries.Get(0) - closeSeries.Get(14)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Missing expected pattern %q\nGenerated:\n%s", pattern, code)
				}
			}

			for _, pattern := range tt.mustExclude {
				if strings.Contains(code, pattern) {
					t.Errorf("Unexpected pattern %q found\nGenerated:\n%s", pattern, code)
				}
			}
		})
	}
}

/* TestPlotExpressionHandler_FixtureCompilation validates the .pine fixture files for new momentum
 * and dispatch-safety functions compile through the full codegen pipeline. */
func TestPlotExpressionHandler_FixtureCompilation(t *testing.T) {
	fixturesDir := "../e2e/fixtures/strategies"

	tests := []struct {
		name        string
		fixture     string
		mustContain []string
	}{
		{
			name:    "momentum functions fixture",
			fixture: "test-ta-momentum.pine",
			mustContain: []string{
				"r3Series.Set(", "f3Series.Set(",
				"<= prev", ">= prev",
				"float64(-_hb5_extIdx)", "float64(-_lb5_extIdx)",
				"m10Series.Set(closeSeries.Get(0) - closeSeries.Get(10))",
				"100.0",
				"ctx.Data[ctx.BarIndex]",
			},
		},
		{
			name:    "math vs TA plot dispatch fixture",
			fixture: "test-ta-math-plot-dispatch.pine",
			mustContain: []string{
				"math.Max(bar.High, bar.Low)",
				"math.Min(bar.High, bar.Low)",
				"math.Abs(",
				"taMaxSeries.Set(",
				"taMinSeries.Set(",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := os.ReadFile(fixturesDir + "/" + tt.fixture)
			if err != nil {
				t.Fatalf("Failed to read fixture %s: %v", tt.fixture, err)
			}

			code, err := compilePineScript(string(content))
			if err != nil {
				t.Fatalf("Compilation failed for %s: %v", tt.fixture, err)
			}

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Missing expected pattern %q\nGenerated:\n%s", pattern, code)
				}
			}
		})
	}
}

/* TestPlotExpressionHandler_BareNameCollisionIsolation verifies each bare name that exists in both
 * MathHandler and InlineTAIIFERegistry dispatches to math when used as a 2-arg call in plot(). */
func TestPlotExpressionHandler_BareNameCollisionIsolation(t *testing.T) {
	collisionNames := []struct {
		bare      string
		mathCall  string
		goPattern string
	}{
		{"max", "max(high, low)", "math.Max(bar.High, bar.Low)"},
		{"min", "min(high, low)", "math.Min(bar.High, bar.Low)"},
	}

	for _, tc := range collisionNames {
		t.Run(tc.bare+"_in_plot_is_math", func(t *testing.T) {
			pine := "//@version=5\nindicator(\"Test\")\nplot(" + tc.mathCall + ", \"x\")\n"
			code, err := compilePineScript(pine)
			if err != nil {
				t.Fatalf("Compilation failed for bare %s in plot: %v", tc.bare, err)
			}

			if !strings.Contains(code, tc.goPattern) {
				t.Errorf("Bare %q in plot() must emit %q\nGenerated:\n%s", tc.bare, tc.goPattern, code)
			}

			/* Confirm TA series infrastructure is absent for this 2-arg math call */
			if strings.Contains(code, tc.bare+"Series.Set(") {
				t.Errorf("Bare %q in plot() must not emit TA series.Set — it is a math call\nGenerated:\n%s", tc.bare, code)
			}
		})

		t.Run(tc.bare+"_as_ta_with_period_is_series", func(t *testing.T) {
			pine := "//@version=5\nindicator(\"Test\")\nx = ta." + tc.bare + "(close, 10)\nplot(x, \"x\")\n"
			code, err := compilePineScript(pine)
			if err != nil {
				t.Fatalf("Compilation failed for ta.%s with period: %v", tc.bare, err)
			}

			if !strings.Contains(code, "xSeries.Set(") {
				t.Errorf("ta.%s(close,10) must emit xSeries.Set() — TA series path\nGenerated:\n%s", tc.bare, code)
			}
		})
	}
}
