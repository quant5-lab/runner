package regression

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/market"
)

func fixtureTimezone(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", filepath.Base(path), err)
	}
	var envelope struct {
		Timezone string `json:"timezone"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("parse timezone from %s: %v", filepath.Base(path), err)
	}
	return envelope.Timezone
}

// boundaryFiring records the bar timestamp at the start of each new period window.
type boundaryFiring struct {
	BarTimeSec int64
}

func collectBoundaryFirings(bars []context.OHLCV, timeframe string, anchor context.PeriodAnchor) []boundaryFiring {
	var firings []boundaryFiring
	var prev int64 = -1
	for _, bar := range bars {
		boundary := context.AlignTimestampToPeriodWithAnchor(bar.Time, timeframe, anchor)
		if boundary != prev {
			firings = append(firings, boundaryFiring{BarTimeSec: bar.Time})
			prev = boundary
		}
	}
	return firings
}

// assertNoFiringAtLocalClock fails if any firing falls at the given local clock
// time.  Used as a differential guard: a UTC-floor implementation fires a
// spurious boundary at a predictable clock time that session-anchored must not.
func assertNoFiringAtLocalClock(t *testing.T, firings []boundaryFiring, tz string, wantH, wantM int) {
	t.Helper()
	loc := mustLocation(t, tz)
	for _, f := range firings {
		local := time.Unix(f.BarTimeSec, 0).In(loc)
		if local.Hour() == wantH && local.Minute() == wantM {
			t.Errorf("boundary fired at %s — must not fire at %02d:%02d %s",
				local, wantH, wantM, tz)
		}
	}
}

func assertAnyFiringAtLocalClock(t *testing.T, firings []boundaryFiring, tz string, wantH, wantM int) {
	t.Helper()
	loc := mustLocation(t, tz)
	for _, f := range firings {
		local := time.Unix(f.BarTimeSec, 0).In(loc)
		if local.Hour() == wantH && local.Minute() == wantM {
			return
		}
	}
	t.Errorf("no boundary firing found at %02d:%02d %s — expected session-anchored boundary",
		wantH, wantM, tz)
}

func firingsBetweenDates(firings []boundaryFiring, from, to time.Time) []boundaryFiring {
	var out []boundaryFiring
	for _, f := range firings {
		ts := time.Unix(f.BarTimeSec, 0).UTC()
		if !ts.Before(from) && ts.Before(to) {
			out = append(out, f)
		}
	}
	return out
}

func firingsEqual(a, b []boundaryFiring) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// fixtureAnchor runs the exact pipeline the generated strategy template executes
// at startup: load bars, resolve profile, derive PeriodAnchor.
func fixtureAnchor(t *testing.T, symbol, path string) (context.PeriodAnchor, []context.OHLCV) {
	t.Helper()
	tz := fixtureTimezone(t, path)
	rawBars := fixtureBarsToRuntimeBars(loadOHLCVBars(t, path))
	bars, profile, err := market.NormalizeBarsWithMetadataE(symbol, "1h", market.SourceMetadata{Timezone: tz}, rawBars)
	if err != nil {
		t.Fatalf("normalize bars for %s: %v", symbol, err)
	}
	return market.DeriveSessionAnchor(profile, bars), bars
}

// TestSessionAnchoredBoundary_AAPL_4h_FiresAtSessionOpenNotUTCGrid uses the
// in-tree AAPL-1h fixture (America/New_York, 500 bars) to prove that 4h
// period-boundary firings align to the NYSE session open (09:30/13:30 ET),
// not the UTC-epoch grid.
//
// UTC-floor tiling fires a false boundary at 12:30 ET (EDT = UTC-4: 16:30 UTC
// crosses the 16:00 UTC 4h grid mark).  Session-anchored tiling from 09:30 ET
// never fires at 12:30 ET because that time falls inside the 09:30-13:29 ET
// slot.  This test cannot pass under a UTC-floor implementation.
func TestSessionAnchoredBoundary_AAPL_4h_FiresAtSessionOpenNotUTCGrid(t *testing.T) {
	root := projectRootFromCwd()
	anchor, bars := fixtureAnchor(t, "AAPL",
		filepath.Join(root, "tests", "golden", "fixtures", "data", "AAPL-1h.json"))

	if anchor.Timezone != "America/New_York" {
		t.Fatalf("anchor.Timezone = %q, want America/New_York", anchor.Timezone)
	}
	const wantSessionOpen = 9*60 + 30
	if anchor.SessionOpenMinute != wantSessionOpen {
		t.Fatalf("anchor.SessionOpenMinute = %d, want %d (09:30 ET)", anchor.SessionOpenMinute, wantSessionOpen)
	}

	firings := collectBoundaryFirings(bars, "240", anchor)
	if len(firings) == 0 {
		t.Fatal("no boundary firings found -- dataset or boundary computation is broken")
	}

	assertNoFiringAtLocalClock(t, firings, "America/New_York", 12, 30)
	assertAnyFiringAtLocalClock(t, firings, "America/New_York", 13, 30)

	// Session-anchored produces fewer firings than UTC-floor because it does not
	// generate spurious mid-session boundaries.
	utcFirings := collectBoundaryFirings(bars, "240", context.PeriodAnchor{})
	if len(utcFirings) <= len(firings) {
		t.Errorf("session-anchored count (%d) must be less than UTC-floor count (%d)",
			len(firings), len(utcFirings))
	}
}

// TestSessionAnchoredBoundary_AAPL_4h_DSTTransitionFiresAtLocalSessionTime
// selects bars from the first post-DST trading day in the AAPL-1h fixture
// (2025-11-03, after the US fall-back on 2025-11-02) and verifies that 4h
// boundaries still fire at 09:30 EST and 13:30 EST rather than at UTC-relative
// times shifted by the one-hour DST offset.  This proves boundary alignment
// uses DST-aware time.LoadLocation rather than a fixed UTC offset.
func TestSessionAnchoredBoundary_AAPL_4h_DSTTransitionFiresAtLocalSessionTime(t *testing.T) {
	root := projectRootFromCwd()
	anchor, bars := fixtureAnchor(t, "AAPL",
		filepath.Join(root, "tests", "golden", "fixtures", "data", "AAPL-1h.json"))

	firings := collectBoundaryFirings(bars, "240", anchor)

	et := mustLocation(t, "America/New_York")
	dstDay := time.Date(2025, 11, 3, 0, 0, 0, 0, et)
	postDSTFirings := firingsBetweenDates(firings, dstDay, dstDay.Add(24*time.Hour))

	if len(postDSTFirings) == 0 {
		t.Fatal("no firings on 2025-11-03 (post-DST Monday) -- AAPL-1h fixture must cover this date")
	}

	assertAnyFiringAtLocalClock(t, postDSTFirings, "America/New_York", 9, 30)
	assertAnyFiringAtLocalClock(t, postDSTFirings, "America/New_York", 13, 30)

	// UTC-floor (EST = UTC-5) fires at the bar entering the 16:00 UTC slot, which
	// is 11:00 EST.  Session-anchored must not fire there.
	assertNoFiringAtLocalClock(t, postDSTFirings, "America/New_York", 11, 0)
}

// TestSessionAnchoredBoundary_SBERP_3h_FiresAtSessionOpenNotUTCGrid verifies
// that for a MOEX (Europe/Moscow, UTC+3) symbol, 3h period boundaries tile from
// the 07:00 MSK session open rather than the UTC-epoch grid.
//
// UTC-floor places a 3h grid mark at 09:00 UTC = 12:00 MSK inside the MOEX
// session, so the bar at 12:00 MSK triggers a spurious boundary change under
// UTC-floor.  Session-anchored tiling from 07:00 MSK (420 min since midnight)
// places slot boundaries at 07:00, 10:00, 13:00, …; 12:00 MSK is always
// mid-slot and must never trigger a change.  The test cannot pass under a
// UTC-floor implementation.
func TestSessionAnchoredBoundary_SBERP_3h_FiresAtSessionOpenNotUTCGrid(t *testing.T) {
	root := projectRootFromCwd()
	anchor, bars := fixtureAnchor(t, "SBERP",
		filepath.Join(root, "tests", "golden", "fixtures", "data", "SBERP-1h.json"))

	if anchor.Timezone != "Europe/Moscow" {
		t.Fatalf("anchor.Timezone = %q, want Europe/Moscow", anchor.Timezone)
	}
	const wantSessionOpen = 7 * 60 // 07:00 MSK
	if anchor.SessionOpenMinute != wantSessionOpen {
		t.Fatalf("anchor.SessionOpenMinute = %d, want %d (07:00 MSK)", anchor.SessionOpenMinute, wantSessionOpen)
	}

	firings := collectBoundaryFirings(bars, "180", anchor)
	if len(firings) == 0 {
		t.Fatal("no boundary firings found — dataset or boundary computation is broken")
	}

	assertNoFiringAtLocalClock(t, firings, "Europe/Moscow", 12, 0)
	assertAnyFiringAtLocalClock(t, firings, "Europe/Moscow", 10, 0)
}

// TestSessionAnchoredBoundary_UTC_AlwaysOpen_MidnightOriginPreserved verifies
// that a UTC/always-open anchor (SessionOpenMinute=0) keeps boundaries at
// UTC midnight multiples, identical to the legacy UTC-floor algorithm.  This
// is the degenerate case: for 24/7 markets (crypto) the session-anchored
// function must not shift boundaries away from UTC midnight, because UTC
// midnight IS the correct origin.
func TestSessionAnchoredBoundary_UTC_AlwaysOpen_MidnightOriginPreserved(t *testing.T) {
	root := projectRootFromCwd()
	anchor, bars := fixtureAnchor(t, "BTCUSDT",
		filepath.Join(root, "tests", "golden", "fixtures", "data", "BTCUSDT-1h.json"))

	if anchor.Timezone != "UTC" {
		t.Fatalf("anchor.Timezone = %q, want UTC for BTCUSDT", anchor.Timezone)
	}
	if anchor.SessionOpenMinute != 0 {
		t.Fatalf("anchor.SessionOpenMinute = %d, want 0 for always-open UTC market", anchor.SessionOpenMinute)
	}

	sessionFirings := collectBoundaryFirings(bars, "240", anchor)
	if len(sessionFirings) == 0 {
		t.Fatal("no boundary firings — dataset or boundary computation is broken")
	}

	utcFirings := collectBoundaryFirings(bars, "240", context.PeriodAnchor{})
	if !firingsEqual(sessionFirings, utcFirings) {
		t.Errorf("UTC always-open anchor produces different firings than legacy UTC-floor: "+
			"session=%d, legacy=%d — zero-offset anchor must degenerate to UTC-floor arithmetic",
			len(sessionFirings), len(utcFirings))
	}

	assertAnyFiringAtLocalClock(t, sessionFirings, "UTC", 0, 0)
	assertAnyFiringAtLocalClock(t, sessionFirings, "UTC", 4, 0)
	assertAnyFiringAtLocalClock(t, sessionFirings, "UTC", 8, 0)
	assertAnyFiringAtLocalClock(t, sessionFirings, "UTC", 12, 0)
}

// TestSessionAnchoredBoundary_CrossExchange_SecondaryUsesOwnAnchor verifies
// that a SBERP (MOEX, Europe/Moscow) primary context and an AAPL
// (America/New_York) secondary context each derive an independent PeriodAnchor,
// and that applying the wrong anchor to AAPL data produces a materially
// different boundary set.
//
// This proves the independence requirement is load-bearing: swapping anchors
// changes outcomes, so any regression that makes a secondary context inherit
// the primary anchor would be detectable.
func TestSessionAnchoredBoundary_CrossExchange_SecondaryUsesOwnAnchor(t *testing.T) {
	root := projectRootFromCwd()
	dataDir := filepath.Join(root, "tests", "golden", "fixtures", "data")

	sberpAnchor, _ := fixtureAnchor(t, "SBERP", filepath.Join(dataDir, "SBERP-1h.json"))
	aaplAnchor, aaplBars := fixtureAnchor(t, "AAPL", filepath.Join(dataDir, "AAPL-1h.json"))

	if sberpAnchor.Timezone == aaplAnchor.Timezone {
		t.Errorf("anchors share timezone %q -- cross-exchange secondary incorrectly inherits primary anchor",
			sberpAnchor.Timezone)
	}
	if sberpAnchor.SessionOpenMinute == aaplAnchor.SessionOpenMinute {
		t.Errorf("anchors share SessionOpenMinute %d -- cross-exchange secondary incorrectly inherits primary anchor",
			sberpAnchor.SessionOpenMinute)
	}

	aaplFirings := collectBoundaryFirings(aaplBars, "240", aaplAnchor)
	assertAnyFiringAtLocalClock(t, aaplFirings, "America/New_York", 13, 30)
	assertNoFiringAtLocalClock(t, aaplFirings, "America/New_York", 12, 30)

	// Applying the SBERP anchor to AAPL data must produce a different boundary set --
	// proving anchors are not interchangeable and the separation is load-bearing.
	if firingsEqual(aaplFirings, collectBoundaryFirings(aaplBars, "240", sberpAnchor)) {
		t.Error("AAPL boundary firings are identical under own anchor and SBERP anchor -- " +
			"anchors must not be interchangeable for cross-exchange correctness to be testable")
	}
}
