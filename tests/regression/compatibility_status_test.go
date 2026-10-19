package regression

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
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
			FirstBar     int      `json:"firstReachedBar"`
			LastBar      int      `json:"lastReachedBar"`
			HitCount     int      `json:"hitCount"`
		} `json:"diagnostics"`
	} `json:"scriptCompatibility"`
	Backtest struct {
		Status        string   `json:"status"`
		AffectedSinks []string `json:"affectedSinks"`
	} `json:"backtest"`
	Indicators map[string]struct {
		Data []json.RawMessage `json:"data"`
	} `json:"indicators"`
	Strategy *struct {
		Trades     []json.RawMessage `json:"trades"`
		OpenTrades []json.RawMessage `json:"openTrades"`
		Equity     float64           `json:"equity"`
		NetProfit  float64           `json:"netProfit"`
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
	assertDiagnostic(t, result, "line.new", "backtest-critical", "strategy.entry")
	assertDiagnosticRuntimeReach(t, result, "line.new", 0, 3, 4)
	assertPlotRetained(t, result, "c")
}

func TestCompatibilityStatus_ChartHandleFlowingToEntryGateIsCalculationBearing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("Chart handle signal")
l = line.new(bar_index, close, bar_index + 1, close)
if l
    strategy.entry("L", strategy.long)
plot(close, "c")
`
	result := runCompatibilityFixture(t, script)

	if result.Backtest.Status != "unsupported_dependency" {
		t.Fatalf("backtest.status = %q, want unsupported_dependency", result.Backtest.Status)
	}
	if result.Strategy != nil {
		t.Fatalf("strategy payload must be omitted for chart handle dependency: %+v", result.Strategy)
	}
	assertDiagnostic(t, result, "line.new", "backtest-critical", "strategy.entry")
	assertDiagnosticRuntimeReach(t, result, "line.new", 0, 3, 4)
	assertPlotRetained(t, result, "c")
}

func TestCompatibilityStatus_InlineChartHandleFlowingToEntryGateIsCalculationBearing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("Direct chart call signal")
if line.new(bar_index, close, bar_index + 1, close)
    strategy.entry("L", strategy.long)
plot(close, "c")
`
	result := runCompatibilityFixture(t, script)

	if result.Backtest.Status != "unsupported_dependency" {
		t.Fatalf("backtest.status = %q, want unsupported_dependency", result.Backtest.Status)
	}
	if result.Strategy != nil {
		t.Fatalf("strategy payload must be omitted for inline chart handle dependency: %+v", result.Strategy)
	}
	assertDiagnostic(t, result, "line.new", "backtest-critical", "strategy.entry")
	assertDiagnosticRuntimeReach(t, result, "line.new", 0, 3, 4)
	assertPlotRetained(t, result, "c")
}

func TestCompatibilityStatus_DirectConditionGapRetainsRuntimeReach(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("If expression gap")
if if_expr_gap(close)
    strategy.entry("L", strategy.long)
plot(close, "c")
`
	result := runCompatibilityFixture(t, script)

	if result.Backtest.Status != "unsupported_dependency" {
		t.Fatalf("backtest.status = %q, want unsupported_dependency", result.Backtest.Status)
	}
	assertDiagnostic(t, result, "if_expr_gap", "backtest-critical", "strategy.entry")
	assertDiagnosticRuntimeReach(t, result, "if_expr_gap", 0, 3, 4)
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
	assertDiagnosticSourceLocation(t, result, "plot_only_gap")
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

func TestCompatibilityStatus_UnsupportedCommentVariableUsesTypedSentinel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("Comment gap")
c = unsupported_comment(close)
strategy.entry("L", strategy.long, comment=c)
plot(close, "c")
`
	result := runCompatibilityFixture(t, script)

	if result.ScriptCompatibility.Status != "degraded" {
		t.Fatalf("scriptCompatibility.status = %q, want degraded", result.ScriptCompatibility.Status)
	}
	if result.Backtest.Status != "complete" {
		t.Fatalf("backtest.status = %q, want complete", result.Backtest.Status)
	}
	if result.Strategy == nil {
		t.Fatal("strategy payload must remain present when unsupported value is only an order comment")
	}
	assertDiagnostic(t, result, "unsupported_comment", "observable-non-backtest", "")
	for _, diag := range result.ScriptCompatibility.Diagnostics {
		if diag.FeatureID == "unsupported_comment" && diag.Substitution != "empty-string" {
			t.Fatalf("substitution = %q, want empty-string", diag.Substitution)
		}
	}
}

func TestCompatibilityStatus_UnsupportedExitCommentVariableUsesTypedSentinel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("Exit comment gap")
c = unsupported_exit_comment(close)
strategy.entry("L", strategy.long)
strategy.exit("X", "L", stop=close * 0.95, comment_profit=c)
plot(close, "c")
`
	result := runCompatibilityFixture(t, script)

	if result.ScriptCompatibility.Status != "degraded" {
		t.Fatalf("scriptCompatibility.status = %q, want degraded", result.ScriptCompatibility.Status)
	}
	if result.Backtest.Status != "complete" {
		t.Fatalf("backtest.status = %q, want complete", result.Backtest.Status)
	}
	if result.Strategy == nil {
		t.Fatal("strategy payload must remain present when unsupported value is only an exit comment")
	}
	assertDiagnostic(t, result, "unsupported_exit_comment", "observable-non-backtest", "")
	for _, diag := range result.ScriptCompatibility.Diagnostics {
		if diag.FeatureID == "unsupported_exit_comment" && diag.Substitution != "empty-string" {
			t.Fatalf("substitution = %q, want empty-string", diag.Substitution)
		}
	}
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

func TestCompatibilityStatus_PositionalCommentAnnotationUsesTypedSentinel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("Positional annotation")
c = unsupported_positional_comment(close)
strategy.entry("L", strategy.long, 1, na, na, na, na, c)
plot(close, "c")
`
	result := runCompatibilityFixture(t, script)

	if result.ScriptCompatibility.Status != "degraded" {
		t.Fatalf("scriptCompatibility.status = %q, want degraded", result.ScriptCompatibility.Status)
	}
	if result.Backtest.Status != "complete" {
		t.Fatalf("backtest.status = %q, want complete", result.Backtest.Status)
	}
	if result.Strategy == nil {
		t.Fatal("strategy payload must remain present when the unsupported value is only a positional order comment")
	}
	assertDiagnostic(t, result, "unsupported_positional_comment", "observable-non-backtest", "")
	for _, diag := range result.ScriptCompatibility.Diagnostics {
		if diag.FeatureID == "unsupported_positional_comment" && diag.Substitution != "empty-string" {
			t.Fatalf("substitution = %q, want empty-string", diag.Substitution)
		}
	}
}

// TestCompatibilityStatus_PositionalAnnotationSinksAcrossStrategyCalls proves
// every comment-bearing strategy call reads a positional `comment` from its real
// Pine argument position, not only strategy.entry.
func TestCompatibilityStatus_PositionalAnnotationSinksAcrossStrategyCalls(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	calls := []struct {
		name string
		call string
	}{
		{"order", `strategy.order("L", strategy.long, 1, na, na, na, na, c)`},
		{"close", `strategy.close("L", c)`},
		{"close_all", `strategy.close_all(c)`},
		{"exit", `strategy.exit("X", "L", na, na, na, na, na, na, na, na, na, na, c)`},
	}

	for _, tc := range calls {
		t.Run(tc.name, func(t *testing.T) {
			featureID := "unsupported_positional_" + tc.name
			script := "//@version=5\n" +
				`strategy("Positional ` + tc.name + ` annotation")` + "\n" +
				"c = " + featureID + "(close)\n" +
				tc.call + "\n" +
				`plot(close, "c")` + "\n"
			result := runCompatibilityFixture(t, script)

			if result.ScriptCompatibility.Status != "degraded" {
				t.Fatalf("scriptCompatibility.status = %q, want degraded", result.ScriptCompatibility.Status)
			}
			if result.Backtest.Status != "complete" {
				t.Fatalf("backtest.status = %q, want complete", result.Backtest.Status)
			}
			if result.Strategy == nil {
				t.Fatalf("strategy payload must remain present when only a positional %s comment is unsupported", tc.name)
			}
			assertDiagnostic(t, result, featureID, "observable-non-backtest", "")
			assertEmptyStringSubstitution(t, result, featureID)
			assertDiagnosticSourceLocation(t, result, featureID)
		})
	}
}

func TestPineGen_ReboundChartHandleTranspiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	projectRoot := projectRootFromCwd()
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "rebind.pine")
	script := `//@version=5
strategy("Rebound chart handle")
if close > open
    handle = line.new(bar_index, close, bar_index + 1, close)
else
    handle = box.new(bar_index, open, bar_index + 1, open)
plot(close, "c")
`
	if err := os.WriteFile(strategyPath, []byte(script), 0644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run",
		filepath.Join(projectRoot, "cmd", "pine-gen", "main.go"),
		"-input", strategyPath,
		"-output", filepath.Join(testDir, "rebind"),
		"-template", filepath.Join(projectRoot, "template", "main.go.tmpl"),
	)
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("transpiling a rebound chart handle did not terminate: %v", ctx.Err())
	}
	if err != nil {
		t.Fatalf("pine-gen failed for a rebound chart handle: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Generated: ") {
		t.Fatalf("pine-gen produced no output file for a rebound chart handle:\n%s", out)
	}
}

func TestCompatibilityStatus_NamedDirectionEntryCarriesTypedSentinel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("Named direction comment")
strategy.entry("L", direction=strategy.long, comment=unsupported_comment(close))
plot(close, "c")
`
	result := runCompatibilityFixture(t, script)

	if result.ScriptCompatibility.Status != "degraded" {
		t.Fatalf("scriptCompatibility.status = %q, want degraded", result.ScriptCompatibility.Status)
	}
	if result.Backtest.Status != "complete" {
		t.Fatalf("backtest.status = %q, want complete", result.Backtest.Status)
	}
	if result.Strategy == nil {
		t.Fatal("strategy payload must remain present when the unsupported value is only a named-direction comment")
	}
	assertDiagnostic(t, result, "unsupported_comment", "observable-non-backtest", "")
	assertEmptyStringSubstitution(t, result, "unsupported_comment")
}

func TestCompatibilityStatus_ExitCommentProfitUsesTypedSentinel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("Exit profit comment")
strategy.entry("L", strategy.long)
strategy.exit("X", "L", stop=close * 0.95, comment_profit=unsupported_profit(close))
plot(close, "c")
`
	result := runCompatibilityFixture(t, script)

	if result.ScriptCompatibility.Status != "degraded" {
		t.Fatalf("scriptCompatibility.status = %q, want degraded", result.ScriptCompatibility.Status)
	}
	if result.Backtest.Status != "complete" {
		t.Fatalf("backtest.status = %q, want complete", result.Backtest.Status)
	}
	if result.Strategy == nil {
		t.Fatal("strategy payload must remain present when the unsupported value is only an exit comment variant")
	}
	assertDiagnostic(t, result, "unsupported_profit", "observable-non-backtest", "")
	assertEmptyStringSubstitution(t, result, "unsupported_profit")
}

func TestCompatibilityStatus_SupportedExitCommentVariantsAreReportedNotDropped(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("Supported exit comment variants")
strategy.entry("L", strategy.long)
strategy.exit("X", "L", stop=close * 0.95, comment="base", comment_profit="TP", comment_loss="SL")
plot(close, "c")
`
	result := runCompatibilityFixture(t, script)

	if result.ScriptCompatibility.Status != "degraded" {
		t.Fatalf("scriptCompatibility.status = %q, want degraded for unrepresented exit comment variants", result.ScriptCompatibility.Status)
	}
	if result.Backtest.Status != "complete" {
		t.Fatalf("backtest.status = %q, want complete", result.Backtest.Status)
	}
	if result.Strategy == nil {
		t.Fatal("strategy payload must remain present when only exit comment variants are unrepresented")
	}
	assertDiagnostic(t, result, "strategy.exit.comment_profit", "observable-non-backtest", "")
	assertDiagnostic(t, result, "strategy.exit.comment_loss", "observable-non-backtest", "")
	assertEmptyStringSubstitution(t, result, "strategy.exit.comment_profit")
	assertDiagnosticSource(t, result, "strategy.exit.comment_profit", "generator.comment_unrepresented")
}

func TestCompatibilityStatus_UnsupportedCommentAnnotationCarriesSourceLocation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("Annotation location")
strategy.entry("L", strategy.long, comment=unsupported_comment(close))
plot(close, "c")
`
	result := runCompatibilityFixture(t, script)

	if result.ScriptCompatibility.Status != "degraded" {
		t.Fatalf("scriptCompatibility.status = %q, want degraded", result.ScriptCompatibility.Status)
	}
	if result.Backtest.Status != "complete" {
		t.Fatalf("backtest.status = %q, want complete", result.Backtest.Status)
	}
	for _, diag := range result.ScriptCompatibility.Diagnostics {
		if diag.FeatureID != "unsupported_comment" {
			continue
		}
		if diag.Impact != "observable-non-backtest" || diag.Substitution != "empty-string" {
			t.Fatalf("diagnostic = %+v, want observable empty-string record", diag)
		}
		if diag.File == "" || diag.Line != 3 || diag.Column != 44 {
			t.Fatalf("location = %s:%d:%d, want generated strategy file line 3 column 44", diag.File, diag.Line, diag.Column)
		}
		return
	}
	t.Fatalf("missing unsupported_comment diagnostic in %+v", result.ScriptCompatibility.Diagnostics)
}

func TestCompatibilityStatus_UnsupportedCommentAnnotationInUnreachedBranchIsStillReported(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("Unreached annotation")
if close < 0
    strategy.entry("L", strategy.long, comment=unsupported_branch_comment(close))
plot(close, "c")
`
	result := runCompatibilityFixture(t, script)

	if result.ScriptCompatibility.Status != "degraded" {
		t.Fatalf("scriptCompatibility.status = %q, want degraded for an annotation the branch never reaches", result.ScriptCompatibility.Status)
	}
	if result.Backtest.Status != "complete" {
		t.Fatalf("backtest.status = %q, want complete", result.Backtest.Status)
	}
	if result.Strategy == nil {
		t.Fatal("strategy payload must remain present when the unsupported value is only a comment")
	}
	for _, diag := range result.ScriptCompatibility.Diagnostics {
		if diag.FeatureID != "unsupported_branch_comment" {
			continue
		}
		if diag.File == "" || diag.Line != 4 || diag.Column != 48 {
			t.Fatalf("location = %s:%d:%d, want generated strategy file line 4 column 48", diag.File, diag.Line, diag.Column)
		}
		return
	}
	t.Fatalf("missing unsupported_branch_comment diagnostic in %+v", result.ScriptCompatibility.Diagnostics)
}

func TestCompatibilityStatus_FullyNamedEntryStillExecutes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
strategy("Fully named entry")
strategy.entry(id="L", direction=strategy.long, comment="signal")
plot(close, "c")
`
	result := runCompatibilityFixture(t, script)

	if result.ScriptCompatibility.Status != "complete" {
		t.Fatalf("scriptCompatibility.status = %q, want complete", result.ScriptCompatibility.Status)
	}
	if result.Strategy == nil {
		t.Fatal("fully named entry must produce a strategy payload")
	}
	if len(result.Strategy.Trades)+len(result.Strategy.OpenTrades) == 0 {
		t.Fatal("fully named entry must still execute an entry order")
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

func assertEmptyStringSubstitution(t *testing.T, result compatibilityResult, featureID string) {
	t.Helper()
	for _, diag := range result.ScriptCompatibility.Diagnostics {
		if diag.FeatureID != featureID {
			continue
		}
		if diag.Substitution != "empty-string" {
			t.Fatalf("%s substitution = %q, want empty-string", featureID, diag.Substitution)
		}
		return
	}
	t.Fatalf("missing %s diagnostic in %+v", featureID, result.ScriptCompatibility.Diagnostics)
}

func assertDiagnosticSource(t *testing.T, result compatibilityResult, featureID, source string) {
	t.Helper()
	for _, diag := range result.ScriptCompatibility.Diagnostics {
		if diag.FeatureID != featureID {
			continue
		}
		if diag.Source != source {
			t.Fatalf("%s source = %q, want %q", featureID, diag.Source, source)
		}
		return
	}
	t.Fatalf("missing %s diagnostic in %+v", featureID, result.ScriptCompatibility.Diagnostics)
}

func assertDiagnosticSourceLocation(t *testing.T, result compatibilityResult, featureID string) {
	t.Helper()
	for _, diag := range result.ScriptCompatibility.Diagnostics {
		if diag.FeatureID != featureID {
			continue
		}
		if diag.File == "" || diag.Line <= 0 {
			t.Fatalf("%s location = %s:%d:%d, want a source position", featureID, diag.File, diag.Line, diag.Column)
		}
		return
	}
	t.Fatalf("missing %s diagnostic in %+v", featureID, result.ScriptCompatibility.Diagnostics)
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func assertDiagnosticRuntimeReach(t *testing.T, result compatibilityResult, featureID string, first, last, hits int) {
	t.Helper()
	for _, diag := range result.ScriptCompatibility.Diagnostics {
		if diag.FeatureID != featureID {
			continue
		}
		if diag.FirstBar != first || diag.LastBar != last || diag.HitCount != hits {
			t.Fatalf("%s runtime reach = %d..%d hits=%d, want %d..%d hits=%d",
				featureID, diag.FirstBar, diag.LastBar, diag.HitCount, first, last, hits)
		}
		return
	}
	t.Fatalf("missing diagnostic feature=%s in %+v", featureID, result.ScriptCompatibility.Diagnostics)
}

// assertPlotRetained verifies a plot returns with data even when the strategy
// payload is withheld, so a suppressed backtest cannot be mistaken for a
// suppressed calculation.
func assertPlotRetained(t *testing.T, result compatibilityResult, title string) {
	t.Helper()
	series, ok := result.Indicators[title]
	if !ok {
		t.Fatalf("plot %q must return independently of the strategy payload; indicators = %v", title, indicatorTitles(result))
	}
	if len(series.Data) == 0 {
		t.Fatalf("plot %q must carry its data independently of the strategy payload", title)
	}
}

func indicatorTitles(result compatibilityResult) []string {
	titles := make([]string, 0, len(result.Indicators))
	for title := range result.Indicators {
		titles = append(titles, title)
	}
	sort.Strings(titles)
	return titles
}
