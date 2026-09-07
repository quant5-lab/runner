package regression

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Session/warmup trim of the 21927-bar SBERP-1h fixture yields 21436 emitted bars;
// pinned so any data-pipeline truncation is caught immediately.
const ultimaFullSeriesBars = 21436

// Ultima is inert by design on SBERP-1h: nz(close) is never ±1 on a ~100–300 RUB
// instrument, so both entry gates are permanently false. Three checks guard distinct
// silent-failure modes: a codegen regression dropping an entry site, a gate-literal
// mutation, or a broken engine producing a zero indistinguishable from the by-design
// zero without a positive control.
func TestUltima_SBERP_Hourly_RuntimeEvidence(t *testing.T) {
	root := projectRootFromCwd()

	source, err := os.ReadFile(filepath.Join(root, "strategies", "top10", "ultima.pine"))
	if err != nil {
		t.Fatalf("read ultima strategy: %v", err)
	}

	tmpDir := t.TempDir()
	built, ok := codegenAndBuild(t, tmpDir, "ultima_evidence", string(source), root)
	if !ok {
		t.Fatal("ultima codegen/build failed — ext_source input or signal-gate codegen regression suspected")
	}

	assertUltimaGeneratedEntryGates(t, built.GeneratedPath)

	fixtureDir := filepath.Join(root, "tests", "golden", "fixtures", "data")
	outputPath := filepath.Join(tmpDir, "out.json")
	cmd := exec.Command(built.BinaryPath,
		"-symbol", "SBERP",
		"-timeframe", "1h",
		"-data", filepath.Join(fixtureDir, "SBERP-1h.json"),
		"-output", outputPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ultima run failed: %v\n%s", err, out)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var result struct {
		Candlestick []struct{} `json:"candlestick"`
		Strategy    struct {
			Trades     []struct{} `json:"trades"`
			OpenTrades []struct{} `json:"openTrades"`
		} `json:"strategy"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("parse output: %v", err)
	}

	if got := len(result.Candlestick); got != ultimaFullSeriesBars {
		t.Errorf("candlestick length: got %d, want %d — full SBERP-1h series required for end-to-end coverage claim", got, ultimaFullSeriesBars)
	}
	if got := len(result.Strategy.Trades); got != 0 {
		t.Errorf("closed trades: got %d, want 0 — ext_source==±1 gates are permanently false on SBERP price (~247 RUB); any non-zero count means entry logic fired unexpectedly", got)
	}
	if got := len(result.Strategy.OpenTrades); got != 0 {
		t.Errorf("open trades: got %d, want 0 — no entry signal can fire, so no position should be open at end of series", got)
	}

	assertUltimaPositiveControl(t, string(source), root)
}

// Entry-deletion and gate-literal mutations collapse to byte-identical run output;
// only a source-level check proves both entry sites and both ±1 literals survived codegen.
func assertUltimaGeneratedEntryGates(t *testing.T, generatedPath string) {
	t.Helper()
	src, err := os.ReadFile(generatedPath)
	if err != nil {
		t.Fatalf("read generated source %q: %v", generatedPath, err)
	}
	gen := string(src)

	for _, check := range []struct {
		needle, label string
	}{
		{`strat.EntryWithDefaultQty("Long", strategy.Long`, "Long entry site"},
		{`strat.EntryWithDefaultQty("Short", strategy.Short`, "Short entry site"},
		{`ext_sourceSeries.GetCurrent() == 1)`, "bull gate literal (ext_source == 1)"},
		{`ext_sourceSeries.GetCurrent() == -1)`, "bear gate literal (ext_source == -1)"},
	} {
		if !strings.Contains(gen, check.needle) {
			t.Errorf("generated code missing %s: needle %q not found in %s", check.label, check.needle, generatedPath)
		}
	}
}

// A broken engine produces zero entries that look identical to the ±1-gate zero;
// relaxing the Long gate to > 0 (always true on SBERP) must fire at least one entry
// or the ±1 causation claim is unsupported.
func assertUltimaPositiveControl(t *testing.T, source, root string) {
	t.Helper()

	ctlSource := strings.ReplaceAll(source, "ext_source == 1", "ext_source > 0")
	ctlDir := t.TempDir()
	built, ok := codegenAndBuild(t, ctlDir, "ultima_control", ctlSource, root)
	if !ok {
		t.Fatal("positive control codegen/build failed")
	}

	fixtureDir := filepath.Join(root, "tests", "golden", "fixtures", "data")
	outputPath := filepath.Join(ctlDir, "ctl.json")
	cmd := exec.Command(built.BinaryPath,
		"-symbol", "SBERP",
		"-timeframe", "1h",
		"-data", filepath.Join(fixtureDir, "SBERP-1h.json"),
		"-output", outputPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("positive control run failed: %v\n%s", err, out)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read positive control output: %v", err)
	}

	var result struct {
		Strategy struct {
			Trades     []struct{} `json:"trades"`
			OpenTrades []struct{} `json:"openTrades"`
		} `json:"strategy"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("parse positive control output: %v", err)
	}

	if total := len(result.Strategy.Trades) + len(result.Strategy.OpenTrades); total == 0 {
		t.Error("positive control produced 0 entries with ext_source > 0 gate — the ±1 literal is not the cause of the zero-trade baseline; by-design rationale must be re-examined")
	}
}
