package regression

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func sessionAnchorWitnessPine(tf string) string {
	return fmt.Sprintf(`//@version=5
strategy("session-anchor witness %s", overlay=true)
b = time("%s")
plot(b, "tf_boundary")
`, tf, tf)
}

type witnessOutput struct {
	Indicators map[string]witnessIndicator `json:"indicators"`
}

type witnessIndicator struct {
	Data []witnessPoint `json:"data"`
}

// BoundaryMs is nil when Pine emits na (the binary marshals NaN as JSON null).
type witnessPoint struct {
	BarTimeSec int64    `json:"time"`
	BoundaryMs *float64 `json:"value"`
}

func boundaryFiringsFromPlot(points []witnessPoint) []boundaryFiring {
	var firings []boundaryFiring
	var prev float64
	hasPrev := false
	for _, p := range points {
		if p.BoundaryMs == nil || math.IsNaN(*p.BoundaryMs) {
			continue
		}
		cur := *p.BoundaryMs
		if !hasPrev || cur != prev {
			firings = append(firings, boundaryFiring{BarTimeSec: p.BarTimeSec})
			prev = cur
			hasPrev = true
		}
	}
	return firings
}

// pineSource must emit a "tf_boundary" plot (see sessionAnchorWitnessPine).
// timeframe is the primary bar resolution of the fixture (e.g. "1h").
func buildAndRunPeriodBoundaryWitness(t *testing.T, root, pineSource, symbol, timeframe, fixtureFile string) []witnessPoint {
	t.Helper()

	tmpDir := t.TempDir()
	built, ok := codegenAndBuild(t, tmpDir, "session_anchor_witness", pineSource, root)
	if !ok {
		t.Fatal("codegen/build failed for the session-anchor witness strategy")
	}

	outputPath := filepath.Join(tmpDir, "witness_out.json")
	cmd := exec.Command(built.BinaryPath,
		"-symbol", symbol,
		"-timeframe", timeframe,
		"-data", fixtureFile,
		"-datadir", filepath.Dir(fixtureFile),
		"-output", outputPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("witness binary run failed: %v\n%s", err, out)
	}

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read witness output: %v", err)
	}

	var parsed witnessOutput
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("parse witness output JSON: %v", err)
	}

	ind, ok := parsed.Indicators["tf_boundary"]
	if !ok {
		t.Fatalf("tf_boundary indicator absent from output; present: %v", witnessIndicatorNames(parsed))
	}
	return ind.Data
}

func witnessIndicatorNames(o witnessOutput) []string {
	names := make([]string, 0, len(o.Indicators))
	for k := range o.Indicators {
		names = append(names, k)
	}
	return names
}

// TestSessionAnchor_AAPL_4h_CompiledBinaryWitness proves, through the full
// Pine codegen → Go build → binary execution pipeline on the real AAPL-1h
// fixture (America/New_York, NYSE session 09:30 ET), that 4h period-boundary
// changes produced by time("240") tile from the session open, not the UTC-epoch
// grid.
//
// Under UTC-floor tiling the 16:00 UTC grid mark falls at 12:00 ET (EDT = UTC-4),
// causing a spurious boundary change at the 12:30 ET hourly bar — the first bar
// after that mark.  Session-anchored tiling from 09:30 ET places the next slot
// boundary at 13:30 ET, so 12:30 ET is always mid-slot and never triggers a
// change.  The test cannot pass under a UTC-floor implementation.
func TestSessionAnchor_AAPL_4h_CompiledBinaryWitness(t *testing.T) {
	root := projectRootFromCwd()
	fixtureDir := filepath.Join(root, "tests", "golden", "fixtures", "data")

	points := buildAndRunPeriodBoundaryWitness(t, root,
		sessionAnchorWitnessPine("240"),
		"AAPL", "1h",
		filepath.Join(fixtureDir, "AAPL-1h.json"),
	)

	if len(points) == 0 {
		t.Fatal("no plot points in witness output — binary produced an empty series")
	}

	firings := boundaryFiringsFromPlot(points)
	if len(firings) == 0 {
		t.Fatal(`time("240") never changed across the AAPL-1h dataset — boundary detection is broken`)
	}

	const tz = "America/New_York"
	assertNoFiringAtLocalClock(t, firings, tz, 12, 30)
	assertAnyFiringAtLocalClock(t, firings, tz, 13, 30)
}

// TestSessionAnchor_SBERP_3h_CompiledBinaryWitness proves, through the full
// Pine codegen → Go build → binary execution pipeline on the real SBERP-1h
// fixture (Europe/Moscow, MOEX session 07:00 MSK), that 3h period-boundary
// changes produced by time("180") tile from the session open, not the UTC-epoch
// grid.
//
// Under UTC-floor tiling the 09:00 UTC grid mark falls at 12:00 MSK (UTC+3),
// causing a spurious boundary change at the 12:00 MSK hourly bar.
// Session-anchored tiling from 07:00 MSK places the next slot boundary at
// 10:00 MSK, so 12:00 MSK is always mid-slot and never triggers a change.
// The test cannot pass under a UTC-floor implementation.
func TestSessionAnchor_SBERP_3h_CompiledBinaryWitness(t *testing.T) {
	root := projectRootFromCwd()
	fixtureDir := filepath.Join(root, "tests", "golden", "fixtures", "data")

	points := buildAndRunPeriodBoundaryWitness(t, root,
		sessionAnchorWitnessPine("180"),
		"SBERP", "1h",
		filepath.Join(fixtureDir, "SBERP-1h.json"),
	)

	if len(points) == 0 {
		t.Fatal("no plot points in witness output — binary produced an empty series")
	}

	firings := boundaryFiringsFromPlot(points)
	if len(firings) == 0 {
		t.Fatal(`time("180") never changed across the SBERP-1h dataset — boundary detection is broken`)
	}

	const tz = "Europe/Moscow"
	// Session-anchored must fire at 12:00 MSK far fewer times than UTC-floor.
	// UTC-floor fires at 12:00 MSK on every normal trading day; session-anchored
	// fires only on exceptional late-open days.  Allow ≤ 2 in the fixture window.
	sess12Count := countFiresAtLocalClock(firings, tz, 12, 0)
	const maxLateOpenFires = 2
	if sess12Count > maxLateOpenFires {
		t.Errorf("expected ≤ %d late-open 12:00 MSK firings in compiled binary output, got %d — "+
			"session-anchored is firing spuriously at 12:00 MSK", maxLateOpenFires, sess12Count)
	}
	assertAnyFiringAtLocalClock(t, firings, tz, 10, 0)
}
