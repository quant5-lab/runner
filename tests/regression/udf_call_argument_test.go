package regression

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const udfArgumentSymbol = "UDFARG"

/* Every call form that can appear as a user-defined function argument must reach the callee as a Go
 * expression: an identifier and a conditional as controls, a TA call plain and nested in an
 * expression and inside a conditional, a boolean-producing TA call, an undefined callee, an input
 * call, a barssince condition, a nested user-defined call, and the same shapes inside
 * request.security, the last two also for a series-source parameter. */
func TestUDFCallArgument_InlineCallFormsCompile(t *testing.T) {
	testCases := []struct {
		name   string
		source string
	}{
		{
			name: "identifier argument",
			source: `//@version=5
indicator("udf identifier argument")
f(x) => x > 50
sel = f(close)
plot(sel ? 1 : 0, "sel")
`,
		},
		{
			name: "conditional expression argument",
			source: `//@version=5
indicator("udf conditional argument")
f(x) => x > 50
sel = f(close > 50 ? 1.0 : 0.0)
plot(sel ? 1 : 0, "sel")
`,
		},
		{
			name: "supported ta call argument",
			source: `//@version=5
indicator("udf ta argument")
f(x) => x > 50
sel = f(ta.sma(close, 20))
plot(sel ? 1 : 0, "sel")
`,
		},
		{
			name: "ta call argument inside a nested expression",
			source: `//@version=5
indicator("udf nested ta argument")
f(x) => x > 50
sel = f(ta.ema(close, 10) + 1)
plot(sel ? 1 : 0, "sel")
`,
		},
		{
			name: "ta call argument inside a conditional",
			source: `//@version=5
indicator("udf conditional ta argument")
f(x) => x > 50
sel = f(ta.rsi(close, 14) > 50 ? 1.0 : 0.0)
plot(sel ? 1 : 0, "sel")
`,
		},
		{
			name: "undefined callee argument",
			source: `//@version=5
indicator("udf undefined callee argument")
f(x) => x > 50
sel = f(qa_gap_b(close))
plot(sel ? 1 : 0, "sel")
`,
		},
		{
			name: "input call argument",
			source: `//@version=5
indicator("udf input argument")
f(x) => x > 50
sel = f(input.int(50, "lvl"))
plot(sel ? 1 : 0, "sel")
`,
		},
		{
			name: "barssince comparison argument",
			source: `//@version=5
indicator("udf barssince argument")
f(x) => x > 50
sel = f(ta.barssince(close > 0))
plot(sel ? 1 : 0, "sel")
`,
		},
		{
			name: "user-defined call argument",
			source: `//@version=5
indicator("udf nested user-defined argument")
f(x) => x > 50
g(y) => y + 1
sel = f(g(close))
plot(sel ? 1 : 0, "sel")
`,
		},
		{
			name: "user-defined and ta call argument",
			source: `//@version=5
indicator("udf nested ta inside user-defined argument")
f(x) => x > 50
f2(x) => x
sel = f(f2(ta.sma(close, 20)))
plot(sel ? 1 : 0, "sel")
`,
		},
		{
			name: "security call argument",
			source: `//@version=5
indicator("udf security argument")
f(x) => x > 50
sel = request.security(syminfo.tickerid, "1D", f(ta.sma(close, 20)))
plot(sel ? 1 : 0, "sel")
`,
		},
		{
			name: "user-defined and ta call argument inside security",
			source: `//@version=5
indicator("udf nested argument inside security")
f(x) => x > 50
g(y) => y + 1
sel = request.security(syminfo.tickerid, "1D", f(g(ta.sma(close, 20))))
plot(sel ? 1 : 0, "sel")
`,
		},
		{
			name: "series source parameter with ta call argument",
			source: `//@version=5
indicator("udf series source argument")
f(src) => ta.sma(src, 3)
sel = f(ta.ema(close, 5))
plot(sel, "sel")
`,
		},
		{
			name: "series source parameter with boolean ta call argument",
			source: `//@version=5
indicator("udf series source boolean argument")
f(src) => ta.sma(src, 3)
sel = f(ta.crossover(close, open))
plot(sel, "sel")
`,
		},
	}

	projectRoot := projectRootFromCwd()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, ok := codegenAndBuild(t, t.TempDir(), "udf_argument", tc.source, projectRoot); !ok {
				t.Fatal("generated Go did not compile")
			}
		})
	}
}

/* A series-typed parameter receives the same value whether the caller hoists the argument into a
 * variable or passes the inline call, because both resolve to the same series. */
func TestUDFCallArgument_SeriesSourceMatchesVariableForm(t *testing.T) {
	source := `//@version=5
indicator("udf series source equivalence")
f(src) => ta.sma(src, 3)
v = ta.ema(close, 5)
a = f(v)
b = f(ta.ema(close, 5))
plot(a - b, "diff")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "udf-series-source.pine")
	if err := os.WriteFile(strategyPath, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, udfArgumentSymbol+"_1h.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(60, 3600)), 0644); err != nil {
		t.Fatal(err)
	}

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRootFromCwd(), udfArgumentSymbol, testDir)

	diffSeries, ok := result.Indicators["diff"]
	if !ok {
		t.Fatalf("indicator \"diff\" absent from output: %v", getIndicatorNames(result.Indicators))
	}

	compared := 0
	for i, value := range extractValues(diffSeries.Data) {
		if math.IsNaN(value) {
			continue
		}
		compared++
		if value != 0 {
			t.Errorf("bar %d: inline call argument differs from the variable form by %v", i, value)
		}
	}

	if compared < 50 {
		t.Errorf("only %d bars compared the two argument forms, want at least 50", compared)
	}
}

/* A security() expression that is a user-defined call is compiled at the call site, so its inline
 * call argument delivers the same value as passing a variable that holds the expression. Distinct
 * user-defined functions are used so the two security calls cannot share one evaluator. */
func TestUDFCallArgument_SecurityInlineCallMatchesVariableForm(t *testing.T) {
	source := `//@version=5
indicator("udf security argument equivalence")
f(x) => x
g(x) => x
v = ta.sma(close, 20)
a = request.security(syminfo.tickerid, "1D", f(v))
b = request.security(syminfo.tickerid, "1D", g(ta.sma(close, 20)))
plot(a - b, "diff")
plot(a, "a")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "udf-security-argument.pine")
	if err := os.WriteFile(strategyPath, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(testDir, udfArgumentSymbol+"_1h.json"), []byte(generateTestOHLCV(240, 3600)), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(testDir, udfArgumentSymbol+"_1D.json"), []byte(generateTestOHLCV(10, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	result := compileAndRun(t, strategyPath, filepath.Join(testDir, udfArgumentSymbol+"_1h.json"), testDir, projectRootFromCwd(), udfArgumentSymbol, testDir)

	securityValue, ok := result.Indicators["a"]
	if !ok {
		t.Fatalf("indicator \"a\" absent from output: %v", getIndicatorNames(result.Indicators))
	}
	if countNonNull(securityValue.Data) < 200 {
		t.Errorf("security argument delivered only %d values, want at least 200", countNonNull(securityValue.Data))
	}

	diff, ok := result.Indicators["diff"]
	if !ok {
		t.Fatalf("indicator \"diff\" absent from output: %v", getIndicatorNames(result.Indicators))
	}

	compared := 0
	for i, value := range extractValues(diff.Data) {
		if math.IsNaN(value) {
			continue
		}
		compared++
		if value != 0 {
			t.Errorf("bar %d: inline call argument differs from the variable form by %v", i, value)
		}
	}

	if compared < 200 {
		t.Errorf("only %d bars compared the two security argument forms, want at least 200", compared)
	}
}

/* An input call used directly as an argument delivers its value, so the callee sees 75 and not a
 * substituted sentinel. */
func TestUDFCallArgument_InputCallDeliversValue(t *testing.T) {
	source := `//@version=5
indicator("udf input argument value")
f(x) => x
plot(f(input.int(75, "lvl")), "sel")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "udf-input-argument.pine")
	if err := os.WriteFile(strategyPath, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, udfArgumentSymbol+"_1h.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 3600)), 0644); err != nil {
		t.Fatal(err)
	}

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRootFromCwd(), udfArgumentSymbol, testDir)

	sel, ok := result.Indicators["sel"]
	if !ok {
		t.Fatalf("indicator \"sel\" absent from output: %v", getIndicatorNames(result.Indicators))
	}

	for i, value := range extractValues(sel.Data) {
		if value != 75 {
			t.Errorf("bar %d: f(input.int(75)) = %v, want 75", i, value)
		}
	}
}

/* ta.barssince takes a condition. A comparison condition already lowers to a Go bool, so the
 * numeric-truthiness helper must not wrap it; the counter resets to 0 on the bar the condition
 * holds and increments while it stays false. */
func TestBarssinceComparisonCondition_CountsBarsSinceCondition(t *testing.T) {
	/* The condition holds on every odd bar, so the counter covers all three branches: no bar has
	 * held it yet, it holds now, and it last held one bar ago. */
	source := `//@version=5
indicator("barssince comparison condition")
n = ta.barssince(bar_index % 2 == 1)
plot(n, "n")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "barssince-comparison.pine")
	if err := os.WriteFile(strategyPath, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, udfArgumentSymbol+"_1h.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(12, 3600)), 0644); err != nil {
		t.Fatal(err)
	}

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRootFromCwd(), udfArgumentSymbol, testDir)

	counter, ok := result.Indicators["n"]
	if !ok {
		t.Fatalf("indicator \"n\" absent from output: %v", getIndicatorNames(result.Indicators))
	}

	values := extractValues(counter.Data)
	for i, value := range values {
		if i == 0 {
			if !math.IsNaN(value) {
				t.Errorf("bar 0: condition has not held yet, got %v, want NaN", value)
			}
			continue
		}

		want := float64((i - 1) % 2)
		if value != want {
			t.Errorf("bar %d: barssince = %v, want %v", i, value, want)
		}
	}
}

/* An arrow body keeps its own TA lowering, so a TA call argument inside a user-defined function
 * body must still compile and produce a value. */
func TestUDFCallArgument_TaCallInsideArrowBody(t *testing.T) {
	source := `//@version=5
indicator("ta call inside arrow body")
f(src) => ta.sma(src, 3)
plot(f(close), "sel")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "udf-arrow-body.pine")
	if err := os.WriteFile(strategyPath, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, udfArgumentSymbol+"_1h.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(40, 3600)), 0644); err != nil {
		t.Fatal(err)
	}

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRootFromCwd(), udfArgumentSymbol, testDir)

	sel, ok := result.Indicators["sel"]
	if !ok {
		t.Fatalf("indicator \"sel\" absent from output: %v", getIndicatorNames(result.Indicators))
	}

	values := extractValues(sel.Data)
	nonNaN := 0
	for _, value := range values {
		if !math.IsNaN(value) {
			nonNaN++
		}
	}

	if nonNaN < 30 {
		t.Errorf("only %d non-NaN values from 40 bars: %v", nonNaN, values)
	}
}

/* An undefined callee in argument position must become a recorded compatibility substitution, not
 * the statement-position placeholder comment whose text would swallow the rest of the call. */
func TestUDFCallArgument_UndefinedCalleeBecomesSubstitution(t *testing.T) {
	source := `//@version=5
indicator("udf argument placeholder")
f(x) => x > 50
sel = f(qa_gap_b(close))
plot(sel ? 1 : 0, "sel")
`
	built, ok := codegenAndBuild(t, t.TempDir(), "udf_argument_placeholder", source, projectRootFromCwd())
	if !ok {
		t.Fatal("generated Go did not compile")
	}

	generated, err := os.ReadFile(built.GeneratedPath)
	if err != nil {
		t.Fatalf("read generated code: %v", err)
	}

	if strings.Contains(string(generated), "TODO: implement") {
		t.Errorf("generated code carries a statement-position placeholder inside a call:\n%s", generated)
	}
	if !strings.Contains(string(generated), `featuregap.Record("qa_gap_b"`) {
		t.Errorf("generated code does not record the undefined callee as a substitution:\n%s", generated)
	}
}
