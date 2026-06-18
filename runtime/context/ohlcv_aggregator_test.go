package context

import (
	"testing"
)

func TestAggregateToCoarserPeriod_EmptyInput(t *testing.T) {
	result := AggregateToCoarserPeriod(nil, "240", PeriodAnchor{})
	if result != nil {
		t.Errorf("nil input: got %v, want nil", result)
	}
	result = AggregateToCoarserPeriod([]OHLCV{}, "240", PeriodAnchor{})
	if result != nil {
		t.Errorf("empty slice: got %v, want nil", result)
	}
}

// TestAggregateToCoarserPeriod_UTC verifies UTC-aligned aggregation: four
// consecutive hourly bars starting at the UTC epoch floor of a 4h slot produce
// exactly one coarser bar with the expected OHLCV values.
func TestAggregateToCoarserPeriod_UTC(t *testing.T) {
	// Four 1h bars: 08:00, 09:00, 10:00, 11:00 UTC
	// 4h UTC slots: [08:00–12:00), so all four belong to the same slot.
	base := int64(1_700_000_000)  // some UTC-aligned unix time divisible by 3600
	base = (base / 14400) * 14400 // snap to 4h UTC boundary

	bars := []OHLCV{
		{Time: base, Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		{Time: base + 3600, Open: 105, High: 115, Low: 100, Close: 108, Volume: 1200},
		{Time: base + 7200, Open: 108, High: 120, Low: 107, Close: 118, Volume: 800},
		{Time: base + 10800, Open: 118, High: 125, Low: 115, Close: 122, Volume: 900},
	}

	result := AggregateToCoarserPeriod(bars, "240", PeriodAnchor{})
	if len(result) != 1 {
		t.Fatalf("got %d coarser bars, want 1", len(result))
	}

	got := result[0]
	if got.Time != base {
		t.Errorf("Time: got %d, want %d", got.Time, base)
	}
	if got.Open != 100 {
		t.Errorf("Open: got %g, want 100", got.Open)
	}
	if got.High != 125 {
		t.Errorf("High: got %g, want 125", got.High)
	}
	if got.Low != 95 {
		t.Errorf("Low: got %g, want 95", got.Low)
	}
	if got.Close != 122 {
		t.Errorf("Close: got %g, want 122", got.Close)
	}
	if got.Volume != 3900 {
		t.Errorf("Volume: got %g, want 3900", got.Volume)
	}
}

// TestAggregateToCoarserPeriod_TwoSlots verifies that bars spanning two
// coarser slots produce two output bars.
func TestAggregateToCoarserPeriod_TwoSlots(t *testing.T) {
	// UTC 4h slot boundary at 4*3600 = 14400s from epoch.
	// Slot 0: 0–14400, Slot 1: 14400–28800
	bars := []OHLCV{
		{Time: 0, Open: 10, High: 12, Low: 9, Close: 11, Volume: 100},
		{Time: 3600, Open: 11, High: 13, Low: 10, Close: 12, Volume: 200},
		{Time: 14400, Open: 20, High: 22, Low: 19, Close: 21, Volume: 300},
		{Time: 18000, Open: 21, High: 25, Low: 20, Close: 24, Volume: 150},
	}

	result := AggregateToCoarserPeriod(bars, "240", PeriodAnchor{})
	if len(result) != 2 {
		t.Fatalf("got %d coarser bars, want 2", len(result))
	}

	if result[0].Time != 0 || result[0].Open != 10 || result[0].High != 13 || result[0].Low != 9 || result[0].Close != 12 {
		t.Errorf("slot 0: %+v", result[0])
	}
	if result[1].Time != 14400 || result[1].Open != 20 || result[1].High != 25 || result[1].Low != 19 || result[1].Close != 24 {
		t.Errorf("slot 1: %+v", result[1])
	}
}

// TestAggregateToCoarserPeriod_SessionAnchor_MOEX verifies that with a
// MOEX-like session anchor (07:00 MSK = UTC+3), 3h slot boundaries tile from
// 07:00 local rather than the UTC floor.
//
// A UTC-floor 3h boundary at 09:00 UTC = 12:00 MSK falls INSIDE the MOEX
// session.  Session-anchored tiling from 07:00 MSK places boundaries at 07:00,
// 10:00, 13:00 MSK; 12:00 MSK is therefore always mid-slot and must not start
// a new coarser bar.
func TestAggregateToCoarserPeriod_SessionAnchor_MOEX(t *testing.T) {
	const tz = "Europe/Moscow" // UTC+3

	// Four hourly bars: 07:00, 08:00, 09:00, 10:00 MSK.
	// Session anchor 07:00 MSK, 3h period → slots [07:00–10:00), [10:00–13:00).
	// UTC timestamps: 07:00 MSK = 04:00 UTC = 1696305600.
	bar0700 := int64(1_696_305_600) // 2023-10-03 04:00 UTC = 07:00 MSK
	bar0800 := bar0700 + 3600
	bar0900 := bar0700 + 7200
	bar1000 := bar0700 + 10800

	bars := []OHLCV{
		{Time: bar0700, Open: 100, High: 105, Low: 98, Close: 103, Volume: 500},
		{Time: bar0800, Open: 103, High: 108, Low: 101, Close: 106, Volume: 600},
		{Time: bar0900, Open: 106, High: 110, Low: 104, Close: 109, Volume: 700},
		{Time: bar1000, Open: 109, High: 115, Low: 107, Close: 113, Volume: 800},
	}

	anchor := PeriodAnchor{
		Timezone:          tz,
		SessionOpenMinute: 7 * 60, // 07:00
	}

	result := AggregateToCoarserPeriod(bars, "180", anchor)
	if len(result) != 2 {
		t.Fatalf("got %d coarser bars, want 2 (slots 07:00–10:00 and 10:00–13:00 MSK)", len(result))
	}

	// Slot 0 (07:00–10:00 MSK): bars at 07:00, 08:00, 09:00 MSK.
	s0 := result[0]
	if s0.Time != bar0700 {
		t.Errorf("slot 0 Time: got %d, want %d (07:00 MSK)", s0.Time, bar0700)
	}
	if s0.Open != 100 || s0.High != 110 || s0.Low != 98 || s0.Close != 109 {
		t.Errorf("slot 0 OHLC: O=%g H=%g L=%g C=%g", s0.Open, s0.High, s0.Low, s0.Close)
	}

	// Slot 1 (10:00–13:00 MSK): bar at 10:00 MSK only.
	s1 := result[1]
	if s1.Time != bar1000 {
		t.Errorf("slot 1 Time: got %d, want %d (10:00 MSK)", s1.Time, bar1000)
	}
	if s1.Open != 109 || s1.High != 115 || s1.Low != 107 || s1.Close != 113 {
		t.Errorf("slot 1 OHLC: O=%g H=%g L=%g C=%g", s1.Open, s1.High, s1.Low, s1.Close)
	}
}

// TestAggregateToCoarserPeriod_SessionAnchor_NYSE verifies NYSE-like session
// (09:30 ET, UTC-4 during EDT).  4h slots tile from 09:30: [09:30–13:30),
// [13:30–17:30).  The UTC-floor 4h boundary at 12:00 ET (16:00 UTC) must
// NOT split a slot; only 13:30 ET starts a new coarser bar.
func TestAggregateToCoarserPeriod_SessionAnchor_NYSE(t *testing.T) {
	const tz = "America/New_York"

	// Five hourly bars: 09:30, 10:30, 11:30, 12:30, 13:30 ET (EDT = UTC-4).
	// UTC timestamps: 09:30 ET = 13:30 UTC.
	bar0930 := int64(1_696_339_800) // 2023-10-03 13:30 UTC = 09:30 EDT
	bar1030 := bar0930 + 3600
	bar1130 := bar0930 + 7200
	bar1230 := bar0930 + 10800
	bar1330 := bar0930 + 14400

	bars := []OHLCV{
		{Time: bar0930, Open: 170, High: 172, Low: 169, Close: 171, Volume: 1000},
		{Time: bar1030, Open: 171, High: 175, Low: 170, Close: 174, Volume: 1200},
		{Time: bar1130, Open: 174, High: 176, Low: 173, Close: 175, Volume: 900},
		{Time: bar1230, Open: 175, High: 178, Low: 174, Close: 177, Volume: 1100},
		{Time: bar1330, Open: 177, High: 180, Low: 176, Close: 179, Volume: 800},
	}

	anchor := PeriodAnchor{
		Timezone:          tz,
		SessionOpenMinute: 9*60 + 30, // 09:30
	}

	result := AggregateToCoarserPeriod(bars, "240", anchor)
	if len(result) != 2 {
		t.Fatalf("got %d coarser bars, want 2 (slots 09:30–13:30 and 13:30–17:30 ET)", len(result))
	}

	// Slot 0 covers 09:30–13:30 ET (bars at 09:30, 10:30, 11:30, 12:30).
	s0 := result[0]
	if s0.Time != bar0930 {
		t.Errorf("slot 0 Time: got %d, want %d (09:30 ET)", s0.Time, bar0930)
	}
	if s0.High != 178 {
		t.Errorf("slot 0 High: got %g, want 178 (not 175 from UTC-floor split at 12:30)", s0.High)
	}
	if s0.Close != 177 {
		t.Errorf("slot 0 Close: got %g, want 177 (12:30 ET bar, not mid-session split)", s0.Close)
	}

	// Slot 1 covers 13:30–17:30 ET (bar at 13:30 only).
	s1 := result[1]
	if s1.Time != bar1330 {
		t.Errorf("slot 1 Time: got %d, want %d (13:30 ET)", s1.Time, bar1330)
	}
}

// TestAggregateToCoarserPeriod_SingleBar verifies a single-bar input produces
// one coarser bar with identical OHLCV.
func TestAggregateToCoarserPeriod_SingleBar(t *testing.T) {
	bar := OHLCV{Time: 3600, Open: 50, High: 55, Low: 48, Close: 52, Volume: 300}
	result := AggregateToCoarserPeriod([]OHLCV{bar}, "240", PeriodAnchor{})
	if len(result) != 1 {
		t.Fatalf("got %d bars, want 1", len(result))
	}
	got := result[0]
	if got.Open != 50 || got.High != 55 || got.Low != 48 || got.Close != 52 || got.Volume != 300 {
		t.Errorf("OHLCV mismatch: %+v", got)
	}
}

// TestAggregateToCoarserPeriod_GapProducesNoEmptySlot verifies that primary data
// gaps (missing bars in the middle of a session) do not produce empty coarser bars
// — only slots that contain at least one primary bar appear in the output.
func TestAggregateToCoarserPeriod_GapProducesNoEmptySlot(t *testing.T) {
	// UTC 4h slots: [0–14400), [14400–28800), [28800–43200).
	// Primary bars exist in slots 0 and 2 but not slot 1.
	bars := []OHLCV{
		{Time: 0, Open: 10, High: 12, Low: 9, Close: 11, Volume: 100},
		{Time: 3600, Open: 11, High: 13, Low: 10, Close: 12, Volume: 200},
		// slot 1 [14400–28800): no bars
		{Time: 28800, Open: 20, High: 22, Low: 19, Close: 21, Volume: 300},
		{Time: 32400, Open: 21, High: 25, Low: 20, Close: 24, Volume: 150},
	}

	result := AggregateToCoarserPeriod(bars, "240", PeriodAnchor{})
	if len(result) != 2 {
		t.Fatalf("got %d coarser bars, want 2 (no bar for empty slot 1)", len(result))
	}
	if result[0].Time != 0 {
		t.Errorf("slot 0 Time: got %d, want 0", result[0].Time)
	}
	if result[1].Time != 28800 {
		t.Errorf("slot 2 Time: got %d, want 28800", result[1].Time)
	}
}

// TestAggregateToCoarserPeriod_VolumeAccumulation verifies that volume is summed
// correctly across many bars in a single slot, with no truncation or rounding.
func TestAggregateToCoarserPeriod_VolumeAccumulation(t *testing.T) {
	const nBars = 60
	// All bars in UTC slot 0 (1h each for 60h, but we keep the period wide enough
	// that they all fit).  Use "D" (daily=1440min) to ensure a single slot.
	base := int64(0)
	bars := make([]OHLCV, nBars)
	var totalVol float64
	for i := 0; i < nBars; i++ {
		v := float64((i + 1) * 100)
		bars[i] = OHLCV{Time: base + int64(i)*3600, Open: 1, High: 2, Low: 0.5, Close: 1, Volume: v}
		totalVol += v
	}

	result := AggregateToCoarserPeriod(bars, "D", PeriodAnchor{})
	if len(result) == 0 {
		t.Fatal("no coarser bars produced")
	}
	// First (and likely only) slot must accumulate the correct total.
	var gotVol float64
	for _, r := range result {
		gotVol += r.Volume
	}
	if gotVol != totalVol {
		t.Errorf("total volume: got %g, want %g", gotVol, totalVol)
	}
}

// TestAggregateToCoarserPeriod_OHLCSemantics verifies the OHLC assignment rules
// across multiple bars in a slot: Open=first, High=max, Low=min, Close=last.
func TestAggregateToCoarserPeriod_OHLCSemantics(t *testing.T) {
	// All four bars in a single UTC 4h slot starting at 0.
	bars := []OHLCV{
		{Time: 0, Open: 100, High: 105, Low: 98, Close: 104, Volume: 10},
		{Time: 3600, Open: 104, High: 120, Low: 103, Close: 119, Volume: 20},
		{Time: 7200, Open: 119, High: 119, Low: 90, Close: 95, Volume: 30},
		{Time: 10800, Open: 95, High: 97, Low: 80, Close: 85, Volume: 40},
	}

	result := AggregateToCoarserPeriod(bars, "240", PeriodAnchor{})
	if len(result) != 1 {
		t.Fatalf("got %d coarser bars, want 1", len(result))
	}

	got := result[0]
	if got.Open != 100 {
		t.Errorf("Open: got %g, want 100 (first bar)", got.Open)
	}
	if got.High != 120 {
		t.Errorf("High: got %g, want 120 (max across all bars)", got.High)
	}
	if got.Low != 80 {
		t.Errorf("Low: got %g, want 80 (min across all bars)", got.Low)
	}
	if got.Close != 85 {
		t.Errorf("Close: got %g, want 85 (last bar)", got.Close)
	}
}

// TestAggregateToCoarserPeriod_SlotTimestampIsSlotOpen verifies that each output
// bar's Time equals the computed slot-open timestamp (not the first primary bar's
// time, which may differ when the primary bar does not start exactly on the slot
// boundary).
func TestAggregateToCoarserPeriod_SlotTimestampIsSlotOpen(t *testing.T) {
	// UTC 4h slot starting at 0.  Feed a bar at time=100 (not on the slot boundary).
	// The output bar's Time must be 0 (the slot open), not 100.
	bars := []OHLCV{
		{Time: 100, Open: 50, High: 55, Low: 48, Close: 52, Volume: 10},
	}

	result := AggregateToCoarserPeriod(bars, "240", PeriodAnchor{})
	if len(result) != 1 {
		t.Fatalf("got %d coarser bars, want 1", len(result))
	}
	if result[0].Time != 0 {
		t.Errorf("slot Time: got %d, want 0 (slot open, not primary bar time)", result[0].Time)
	}
}

// TestAggregateToCoarserPeriod_SessionAnchor_SlotTimestampIsSessionSlotOpen
// verifies that when a primary bar arrives mid-slot under a non-UTC session
// anchor, the output coarser bar's Time equals the session-anchored slot open,
// not the primary bar's timestamp.  Session-anchor analogue of
// TestAggregateToCoarserPeriod_SlotTimestampIsSlotOpen which covers the UTC
// case.
func TestAggregateToCoarserPeriod_SessionAnchor_SlotTimestampIsSessionSlotOpen(t *testing.T) {
	const tz = "America/New_York"

	// 2023-10-03 09:30 EDT = 2023-10-03 13:30 UTC — the slot open.
	slotOpenUTC := int64(1_696_339_800)
	// 2023-10-03 11:30 EDT = 2023-10-03 15:30 UTC — mid-slot primary bar.
	midSlotUTC := slotOpenUTC + 2*3600

	bars := []OHLCV{
		{Time: midSlotUTC, Open: 170, High: 175, Low: 168, Close: 173, Volume: 500},
	}
	anchor := PeriodAnchor{
		Timezone:          tz,
		SessionOpenMinute: 9*60 + 30,
	}

	result := AggregateToCoarserPeriod(bars, "240", anchor)
	if len(result) != 1 {
		t.Fatalf("got %d coarser bars, want 1", len(result))
	}
	if result[0].Time != slotOpenUTC {
		t.Errorf("slot Time: got %d, want %d (09:30 ET slot open, not mid-slot bar time %d)",
			result[0].Time, slotOpenUTC, midSlotUTC)
	}
}
