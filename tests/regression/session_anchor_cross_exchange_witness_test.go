package regression

import (
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/quant5-lab/runner/runtime/context"
)

// crossExchangeWitnessPine plots the 4h period-boundary timestamp of a
// cross-exchange NYSE secondary (AAPL) from a MOEX primary (SBERP).
// The value of time("240") for each AAPL 4h secondary bar is the 4h slot
// start, which differs under NYSE anchor (09:30 ET) vs MOEX anchor (08:00 ET).
const crossExchangeWitnessPine = `//@version=5
strategy("cross-exchange-anchor witness", overlay=true)
aapl_slot_start = security("AAPL", "240", time("240"))
plot(aapl_slot_start, "aapl_slot_start")
`

// aaplFixture4hNoTimezone builds a 4h AAPL fixture from the hourly AAPL
// source, using the NYSE session anchor (09:30 America/New_York) for
// aggregation.  The resulting fixture declares only exchange:"NYSE" — no
// timezone field — so the binary under test must derive America/New_York
// from the exchange rather than inheriting the primary's Europe/Moscow.
func aaplFixture4hNoTimezone(t *testing.T, aaplHourlyPath string) []byte {
	t.Helper()

	rawBars := fixtureBarsToRuntimeBars(loadOHLCVBars(t, aaplHourlyPath))

	nyseAnchor := context.PeriodAnchor{
		Timezone:          "America/New_York",
		SessionOpenMinute: 9*60 + 30,
	}
	coarseBars := context.AggregateToCoarserPeriod(rawBars, "4h", nyseAnchor)
	if len(coarseBars) == 0 {
		t.Fatal("AggregateToCoarserPeriod produced no 4h AAPL bars")
	}

	type fixtureEnvelope struct {
		Exchange string          `json:"exchange"`
		Bars     []context.OHLCV `json:"bars"`
	}
	data, err := json.Marshal(fixtureEnvelope{Exchange: "NYSE", Bars: coarseBars})
	if err != nil {
		t.Fatalf("marshal 4h fixture: %v", err)
	}
	return data
}

// collectDistinctSlotStartsSec extracts all unique non-NaN values from the
// aapl_slot_start plot, converting from Pine milliseconds to UNIX seconds.
// Each distinct value is the 4h slot start that the AAPL secondary bar maps
// to under the anchor currently in use.
func collectDistinctSlotStartsSec(pts []witnessPoint) []int64 {
	seen := make(map[float64]bool)
	var result []int64
	for _, p := range pts {
		if p.BoundaryMs == nil || math.IsNaN(*p.BoundaryMs) {
			continue
		}
		v := *p.BoundaryMs
		if !seen[v] {
			seen[v] = true
			result = append(result, int64(v)/1000)
		}
	}
	return result
}

func assertAnySlotStartAtLocalClock(t *testing.T, slotStartsSec []int64, tz string, wantH, wantM int) {
	t.Helper()
	loc := mustLocation(t, tz)
	for _, ts := range slotStartsSec {
		local := time.Unix(ts, 0).In(loc)
		if local.Hour() == wantH && local.Minute() == wantM {
			return
		}
	}
	t.Errorf("no secondary slot-start value corresponds to %02d:%02d %s — session-anchor boundary missing", wantH, wantM, tz)
}

func assertNoSlotStartAtLocalClock(t *testing.T, slotStartsSec []int64, tz string, badH, badM int) {
	t.Helper()
	loc := mustLocation(t, tz)
	for _, ts := range slotStartsSec {
		local := time.Unix(ts, 0).In(loc)
		if local.Hour() == badH && local.Minute() == badM {
			t.Errorf("secondary slot-start value at %s corresponds to %02d:%02d %s — wrong anchor in use",
				local, badH, badM, tz)
		}
	}
}

// TestSessionAnchor_CrossExchange_SBERP_Primary_AAPL_Secondary_UsesNYSEAnchor
// proves, through the full Pine codegen → Go build → binary execution pipeline,
// that a MOEX-primary (SBERP, Europe/Moscow) strategy reading a 4h AAPL
// secondary fixture that carries only exchange:"NYSE" and no timezone field
// derives the NYSE session anchor (America/New_York, 09:30 ET) for that
// secondary instead of silently inheriting the primary's MOEX anchor.
//
// The discriminant is the VALUE of time("240") for each AAPL 4h secondary bar:
//
//	NYSE anchor (correct):  slot starts at 09:30 ET and 13:30 ET.
//	MOEX anchor (buggy):    slot starts at 08:00 ET and 12:00 ET.
//
// The test collects the set of distinct time("240") values emitted by the
// strategy and asserts that 09:30 ET appears (NYSE) while 08:00 ET does not
// (MOEX false boundary).  The two sets are mutually exclusive, so the test
// cannot pass under both anchors simultaneously.
func TestSessionAnchor_CrossExchange_SBERP_Primary_AAPL_Secondary_UsesNYSEAnchor(t *testing.T) {
	root := projectRootFromCwd()
	fixtureDir := filepath.Join(root, "tests", "golden", "fixtures", "data")

	// Secondary fixture: 4h AAPL bars with exchange:"NYSE", no timezone field.
	// candidateFixturePaths("AAPL","240") → AAPL_4h.json — named accordingly.
	secondaryDataDir := t.TempDir()
	fixture4h := aaplFixture4hNoTimezone(t, filepath.Join(fixtureDir, "AAPL-1h.json"))
	if err := os.WriteFile(filepath.Join(secondaryDataDir, "AAPL_4h.json"), fixture4h, 0644); err != nil {
		t.Fatalf("write 4h secondary fixture: %v", err)
	}

	tmpBuild := t.TempDir()
	built, ok := codegenAndBuild(t, tmpBuild, "cross_exchange_anchor_witness", crossExchangeWitnessPine, root)
	if !ok {
		t.Fatal("codegen/build failed for the cross-exchange anchor witness strategy")
	}

	outputPath := filepath.Join(tmpBuild, "witness_out.json")
	cmd := exec.Command(built.BinaryPath,
		"-symbol", "SBERP",
		"-timeframe", "1h",
		"-data", filepath.Join(fixtureDir, "SBERP-1h.json"),
		"-datadir", secondaryDataDir,
		"-output", outputPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("witness binary exited non-zero:\n%s", out)
	}

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read witness output: %v", err)
	}
	var parsed witnessOutput
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("parse witness output JSON: %v", err)
	}

	ind, ok := parsed.Indicators["aapl_slot_start"]
	if !ok {
		t.Fatalf("aapl_slot_start indicator absent from output; present: %v", witnessIndicatorNames(parsed))
	}

	slotStartsSec := collectDistinctSlotStartsSec(ind.Data)
	if len(slotStartsSec) == 0 {
		t.Fatal("aapl_slot_start has no non-null values — secondary fetch or time() evaluation broken")
	}

	const tz = "America/New_York"
	// NYSE anchor (09:30 ET origin) produces slot starts at 09:30 ET and 13:30 ET.
	assertAnySlotStartAtLocalClock(t, slotStartsSec, tz, 9, 30)
	// MOEX anchor (07:00 MSK = 04:00 ET origin) maps the 09:30 ET AAPL bar to
	// the 08:00–12:00 ET slot → slot start at 08:00 ET.  Must not appear.
	assertNoSlotStartAtLocalClock(t, slotStartsSec, tz, 8, 0)
	// MOEX anchor maps the 13:30 ET AAPL bar to the 12:00–16:00 ET slot → 12:00 ET.
	assertNoSlotStartAtLocalClock(t, slotStartsSec, tz, 12, 0)
}
