package regression

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type compatibilityResult struct {
	ScriptCompatibility struct {
		Status      string `json:"status"`
		Diagnostics []struct {
			FeatureID    string   `json:"featureId"`
			Phase        string   `json:"phase"`
			Impact       string   `json:"impact"`
			Substitution string   `json:"substitution"`
			Source       string   `json:"source"`
			File         string   `json:"file"`
			Line         int      `json:"line"`
			Column       int      `json:"column"`
			Sinks        []string `json:"affectedSinks"`
		} `json:"diagnostics"`
	} `json:"scriptCompatibility"`
	Backtest struct {
		Status        string   `json:"status"`
		AffectedSinks []string `json:"affectedSinks"`
	} `json:"backtest"`
	Strategy *struct {
		Trades    []json.RawMessage `json:"trades"`
		Equity    float64           `json:"equity"`
		NetProfit float64           `json:"netProfit"`
	} `json:"strategy"`
}

func TestCompatibilityStatus_BacktestCriticalUnknownSuppressesStrategyMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("Gap reaches entry")
x = unsupported_entry_source(close)
if nz(x, 1) > 0
    strategy.entry("L", strategy.long)
plot(close, "c")
`
	result := runCompatibilityFixture(t, script)

	if result.ScriptCompatibility.Status != "degraded" {
		t.Fatalf("scriptCompatibility.status = %q, want degraded", result.ScriptCompatibility.Status)
	}
	if result.Backtest.Status != "unsupported_dependency" {
		t.Fatalf("backtest.status = %q, want unsupported_dependency", result.Backtest.Status)
	}
	if result.Strategy != nil {
		t.Fatalf("strategy payload must be omitted for unsupported backtest dependency: %+v", result.Strategy)
	}
	assertDiagnostic(t, result, "unsupported_entry_source", "backtest-critical", "strategy.entry")
}

func TestCompatibilityStatus_LineGetterFlowingToEntryIsCalculationBearing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("Line getter reaches entry")
l = line.new(bar_index, close, bar_index + 1, close)
p = line.get_price(l, bar_index)
if nz(p, 1) > 0
    strategy.entry("L", strategy.long)
plot(close, "c")
`
	result := runCompatibilityFixture(t, script)

	if result.Backtest.Status != "unsupported_dependency" {
		t.Fatalf("backtest.status = %q, want unsupported_dependency", result.Backtest.Status)
	}
	if result.Strategy != nil {
		t.Fatalf("strategy payload must be omitted for line getter dependency: %+v", result.Strategy)
	}
	assertDiagnostic(t, result, "line.get_price", "backtest-critical", "strategy.entry")
}

func TestCompatibilityStatus_UnsupportedPlotKeepsBacktestComplete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("Plot-only gap")
plot(plot_only_gap(close), "gap")
strategy.entry("L", strategy.long)
`
	result := runCompatibilityFixture(t, script)

	if result.ScriptCompatibility.Status != "degraded" {
		t.Fatalf("scriptCompatibility.status = %q, want degraded", result.ScriptCompatibility.Status)
	}
	if result.Backtest.Status != "complete" {
		t.Fatalf("backtest.status = %q, want complete", result.Backtest.Status)
	}
	if result.Strategy == nil {
		t.Fatal("strategy payload must remain present when unsupported value is plot-only")
	}
	assertDiagnostic(t, result, "plot_only_gap", "observable-non-backtest", "")
}

func TestCompatibilityStatus_UnsupportedStatementRecordsDiagnostic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("Statement gap")
unsupported_statement(close)
strategy.entry("L", strategy.long)
plot(close, "c")
`
	result := runCompatibilityFixture(t, script)

	if result.ScriptCompatibility.Status != "degraded" {
		t.Fatalf("scriptCompatibility.status = %q, want degraded", result.ScriptCompatibility.Status)
	}
	if result.Backtest.Status != "complete" {
		t.Fatalf("backtest.status = %q, want complete", result.Backtest.Status)
	}
	assertDiagnostic(t, result, "unsupported_statement", "observable-non-backtest", "")
}

func TestCompatibilityStatus_DiagnosticCarriesSourceLocation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("Source location gap")
x = unsupported_entry_source(close)
if x > 0
    strategy.entry("L", strategy.long)
`
	result := runCompatibilityFixture(t, script)

	for _, diag := range result.ScriptCompatibility.Diagnostics {
		if diag.FeatureID != "unsupported_entry_source" {
			continue
		}
		if diag.File == "" || diag.Line != 3 || diag.Column != 5 {
			t.Fatalf("location = %s:%d:%d, want generated strategy file line 3 column 5", diag.File, diag.Line, diag.Column)
		}
		return
	}
	t.Fatalf("missing unsupported_entry_source diagnostic in %+v", result.ScriptCompatibility.Diagnostics)
}

func TestCompatibilityStatus_UDFMediatedGapSuppressesStrategyMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("UDF gap")
f(v) => unsupported_entry_source(v)
x = f(close)
if nz(x, 1) > 0
    strategy.entry("L", strategy.long)
plot(close, "c")
`
	result := runCompatibilityFixture(t, script)

	if result.Backtest.Status != "unsupported_dependency" {
		t.Fatalf("backtest.status = %q, want unsupported_dependency", result.Backtest.Status)
	}
	if result.Strategy != nil {
		t.Fatalf("strategy payload must be omitted for UDF-mediated unsupported dependency: %+v", result.Strategy)
	}
	assertDiagnostic(t, result, "unsupported_entry_source", "backtest-critical", "strategy.entry")
}

func TestCompatibilityStatus_AllControlFlowGapsMarkedBacktestCritical(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("Multiple gaps")
a = gap_a(close)
b = gap_b(close)
if nz(a, 1) > 0 and nz(b, 1) > 0
    strategy.entry("L", strategy.long)
plot(close, "c")
`
	result := runCompatibilityFixture(t, script)

	assertDiagnostic(t, result, "gap_a", "backtest-critical", "strategy.entry")
	assertDiagnostic(t, result, "gap_b", "backtest-critical", "strategy.entry")
}

func TestPineGen_UnrecoverableParseReportsBlockedCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	projectRoot := projectRootFromCwd()
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "broken.pine")
	if err := os.WriteFile(strategyPath, []byte(`//@version=5
strategy("Broken grammar")
if close >
    strategy.entry("L", strategy.long)
`), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "run",
		filepath.Join(projectRoot, "cmd", "pine-gen", "main.go"),
		"-input", strategyPath,
		"-output", filepath.Join(testDir, "broken"),
		"-template", filepath.Join(projectRoot, "template", "main.go.tmpl"),
	)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("pine-gen must keep non-zero exit for unrecoverable parse failure")
	}

	firstLine := strings.SplitN(string(out), "\n", 2)[0]
	var result compatibilityResult
	if jsonErr := json.Unmarshal([]byte(firstLine), &result); jsonErr != nil {
		t.Fatalf("parse failure must emit structured compatibility JSON, got %q: %v", string(out), jsonErr)
	}
	if result.ScriptCompatibility.Status != "blocked" {
		t.Fatalf("scriptCompatibility.status = %q, want blocked", result.ScriptCompatibility.Status)
	}
	if result.Backtest.Status != "blocked" {
		t.Fatalf("backtest.status = %q, want blocked", result.Backtest.Status)
	}
	assertDiagnostic(t, result, "parser.unrecoverable", "structural", "")
	if len(result.ScriptCompatibility.Diagnostics) != 1 {
		t.Fatalf("diagnostics = %+v, want exactly one parser diagnostic", result.ScriptCompatibility.Diagnostics)
	}
	diag := result.ScriptCompatibility.Diagnostics[0]
	if diag.File == "" || diag.Line == 0 || diag.Column == 0 {
		t.Fatalf("parser diagnostic location = %s:%d:%d, want non-zero source position", diag.File, diag.Line, diag.Column)
	}
}

func runCompatibilityFixture(t *testing.T, script string) compatibilityResult {
	t.Helper()

	projectRoot := projectRootFromCwd()
	testDir := t.TempDir()
	build, ok := codegenAndBuild(t, testDir, "strategy", script, projectRoot)
	if !ok {
		t.Fatal("codegen and build failed")
	}

	dataPath := filepath.Join(testDir, "data.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(4, 86400)), 0644); err != nil {
		t.Fatal(err)
	}
	resultPath := filepath.Join(testDir, "result.json")
	runCmd := exec.Command(build.BinaryPath, "-symbol", "TEST", "-data", dataPath, "-output", resultPath)
	if out, err := runCmd.CombinedOutput(); err != nil {
		t.Fatalf("generated binary failed: %v\n%s", err, out)
	}

	data, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	var result compatibilityResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal result: %v\n%s", err, data)
	}
	return result
}

func assertDiagnostic(t *testing.T, result compatibilityResult, featureID, impact, sink string) {
	t.Helper()
	for _, diag := range result.ScriptCompatibility.Diagnostics {
		if diag.FeatureID != featureID || diag.Impact != impact {
			continue
		}
		if sink == "" || containsString(diag.Sinks, sink) {
			return
		}
	}
	t.Fatalf("missing diagnostic feature=%s impact=%s sink=%s in %+v", featureID, impact, sink, result.ScriptCompatibility.Diagnostics)
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
