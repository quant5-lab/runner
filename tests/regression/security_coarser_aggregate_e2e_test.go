package regression

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestSecurityCoarserAggregateE2E_AAPL_4h_NoFixture exercises the
// AggregateToCoarserPeriod fallback end-to-end: a compiled binary whose Pine
// strategy calls security(syminfo.tickerid, "240", close) must run to
// completion without aborting when no 4h fixture is present in the data
// directory, and the security() plot value must update only at
// session-anchored 4h boundaries rather than the UTC-floor grid.
//
// Under UTC-floor tiling the 16:00 UTC boundary falls at 12:00 EDT, causing
// a spurious value change at the 12:30 ET bar.  Under session-anchored tiling
// from NYSE open 09:30 ET, the first slot boundary falls at 13:30 ET; the
// 12:30 ET bar is always mid-slot and must not trigger a change.
func TestSecurityCoarserAggregateE2E_AAPL_4h_NoFixture(t *testing.T) {
	const pine = `//@version=5
strategy("coarser-agg-witness", overlay=true)
sec_close = security(syminfo.tickerid, "240", close)
plot(sec_close, "coarser_close")
`
	root := projectRootFromCwd()
	fixtureDir := filepath.Join(root, "tests", "golden", "fixtures", "data")
	aaplFixture := filepath.Join(fixtureDir, "AAPL-1h.json")

	tmpBuild := t.TempDir()

	// dataDir contains only the primary 1h fixture — no 4h fixture — so the
	// fallback aggregation path is the only option for the 240-minute request.
	noCoarserDir := t.TempDir()
	fixtureBytes, err := os.ReadFile(aaplFixture)
	if err != nil {
		t.Fatalf("read AAPL-1h fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(noCoarserDir, "AAPL-1h.json"), fixtureBytes, 0644); err != nil {
		t.Fatalf("copy fixture to isolated dataDir: %v", err)
	}

	built, ok := codegenAndBuild(t, tmpBuild, "coarser_agg_witness", pine, root)
	if !ok {
		t.Fatal("codegen/build failed — security(syminfo.tickerid, \"240\", close) strategy did not compile")
	}

	outputPath := filepath.Join(tmpBuild, "out.json")
	cmd := exec.Command(built.BinaryPath,
		"-symbol", "AAPL",
		"-timeframe", "1h",
		"-data", filepath.Join(noCoarserDir, "AAPL-1h.json"),
		"-datadir", noCoarserDir,
		"-output", outputPath,
	)
	runOut, runErr := cmd.CombinedOutput()
	if runErr != nil {
		t.Fatalf("binary exited non-zero — AggregateToCoarserPeriod fallback absent or broken:\n%s", runOut)
	}

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var parsed witnessOutput
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("parse output JSON: %v", err)
	}

	ind, ok := parsed.Indicators["coarser_close"]
	if !ok {
		t.Fatalf("coarser_close indicator absent from output; present: %v", witnessIndicatorNames(parsed))
	}
	if len(ind.Data) == 0 {
		t.Fatal("coarser_close has no data points — aggregation produced empty bars")
	}

	firings := boundaryFiringsFromPlot(ind.Data)
	if len(firings) == 0 {
		t.Fatal("security(\"240\", close) value never changed — aggregate mapper or evaluator is broken")
	}

	const tz = "America/New_York"
	// Session-anchored 4h slots from NYSE open 09:30 ET → slot boundary at 13:30 ET.
	assertAnyFiringAtLocalClock(t, firings, tz, 13, 30)
	// UTC-floor false boundary: 16:00 UTC = 12:00 EDT → next bar 12:30 ET must NOT fire.
	assertNoFiringAtLocalClock(t, firings, tz, 12, 30)
}
