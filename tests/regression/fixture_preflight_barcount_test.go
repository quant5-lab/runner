package regression

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tvref "github.com/quant5-lab/runner/tests/regression/tv_reference"
)

// referenceCSVPath resolves the TV reference CSV file for a wired alignment case.
func referenceCSVPath(root string, tc tvAlignmentCase) string {
	return filepath.Join(root, "tests", "regression", "tv_reference", "fixtures", tc.CSV)
}

// referenceTradeOverlap returns both the total trade count from csvPath and the
// count of those whose entry falls inside the fixture's own time window. Two
// callers need different verdicts from the same I/O: inWindowTradeCount derives
// the bar-count floor; fixtureWindowDisjointFromReference checks whether any
// overlap exists at all. Sharing the load avoids duplicate file reads.
func referenceTradeOverlap(fixturePath, csvPath string, tz tvref.TVTimezone) (total, inWindow int, err error) {
	start, end, err := tvref.FixtureWindow(fixturePath)
	if err != nil {
		return 0, 0, fmt.Errorf("fixture window %s: %w", fixturePath, err)
	}
	all, err := tvref.LoadTrades(csvPath, tz)
	if err != nil {
		return 0, 0, fmt.Errorf("load reference CSV %s: %w", csvPath, err)
	}
	return len(all), len(tvref.FilterByEntryWindow(all, start, end)), nil
}

// inWindowTradeCount returns the number of reference trades whose entry bar falls
// inside the fixture's own time window [firstBar, lastBar]. This is the minimum
// bar count the fixture must have: a fixture cannot host more closed trades than
// it has bars, so len(fixture.Bars) < inWindowTradeCount signals a structural
// impossibility.
//
// Returns zero when no reference trades fall inside the fixture window — a
// non-error result that always satisfies the bar-count check (e.g. a placeholder
// fixture whose synthetic timestamps lie after the real reference history).
func inWindowTradeCount(fixturePath, csvPath string, tz tvref.TVTimezone) (int, error) {
	_, inWindow, err := referenceTradeOverlap(fixturePath, csvPath, tz)
	return inWindow, err
}

// minimalCSVHeader is the smallest valid header accepted by tvref.LoadTrades.
const minimalCSVHeader = "Trade number,Type,Date and time,Price USD"

type barcountTestTrade struct {
	Num   int
	Entry string
	Exit  string
}

// writeTempTradeCSV writes a minimal TV-format CSV with one entry+exit row per
// trade and returns the file path.
func writeTempTradeCSV(t *testing.T, dir, name string, trades []barcountTestTrade) string {
	t.Helper()
	var sb strings.Builder
	sb.WriteString(minimalCSVHeader + "\n")
	for _, tr := range trades {
		fmt.Fprintf(&sb, "%d,Entry long,%s,100.00\n", tr.Num, tr.Entry)
		fmt.Fprintf(&sb, "%d,Exit long,%s,105.00\n", tr.Num, tr.Exit)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		t.Fatalf("write trade CSV %s: %v", name, err)
	}
	return path
}

// writeTempBarFixture writes a minimal fixture JSON with the given Unix-second
// timestamps and fixed OHLCV values, and returns the file path.
func writeTempBarFixture(t *testing.T, dir, name string, timestamps []int64) string {
	t.Helper()
	type bar struct {
		Time   int64   `json:"time"`
		Open   float64 `json:"open"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Close  float64 `json:"close"`
		Volume float64 `json:"volume"`
	}
	bars := make([]bar, len(timestamps))
	for i, ts := range timestamps {
		bars[i] = bar{Time: ts, Open: 100, High: 101, Low: 99, Close: 100, Volume: 1000}
	}
	data, err := json.Marshal(struct {
		Timezone string `json:"timezone"`
		Bars     []bar  `json:"bars"`
	}{Timezone: "UTC", Bars: bars})
	if err != nil {
		t.Fatalf("marshal bar fixture: %v", err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write bar fixture %s: %v", name, err)
	}
	return path
}

// TestReferenceTradeOverlap is the comprehensive behavioural test for the window-
// overlap logic shared by inWindowTradeCount and fixtureWindowDisjointFromReference.
//
// Fixture window used throughout (unless overridden per case):
// 2024-01-05 00:00 UTC → 2024-01-15 00:00 UTC (Unix 1704412800 → 1705276800).
func TestReferenceTradeOverlap(t *testing.T) {
	const (
		windowStart int64 = 1704412800 // 2024-01-05 00:00 UTC
		windowEnd   int64 = 1705276800 // 2024-01-15 00:00 UTC
	)
	defaultWindow := []int64{windowStart, windowEnd}

	tradeInside := barcountTestTrade{Num: 1, Entry: "2024-01-07 10:00", Exit: "2024-01-10 10:00"}
	tradeInsideB := barcountTestTrade{Num: 2, Entry: "2024-01-09 08:00", Exit: "2024-01-11 12:00"}
	tradeInsideC := barcountTestTrade{Num: 3, Entry: "2024-01-13 14:00", Exit: "2024-01-14 16:00"}
	tradeBefore := barcountTestTrade{Num: 4, Entry: "2023-12-25 10:00", Exit: "2023-12-28 10:00"}
	tradeAfter := barcountTestTrade{Num: 5, Entry: "2024-01-20 10:00", Exit: "2024-01-22 10:00"}
	tradeOnStart := barcountTestTrade{Num: 6, Entry: "2024-01-05 00:00", Exit: "2024-01-06 10:00"}
	tradeOnEnd := barcountTestTrade{Num: 7, Entry: "2024-01-15 00:00", Exit: "2024-01-16 00:00"}

	cases := []struct {
		name       string
		timestamps []int64 // nil → defaultWindow; []int64{} → zero-bar fixture
		trades     []barcountTestTrade
		wantTotal  int
		wantIn     int
		wantError  bool
	}{
		{
			name:      "empty_csv_both_zero",
			trades:    []barcountTestTrade{},
			wantTotal: 0, wantIn: 0,
		},
		{
			name:      "single_in_window_both_counts_one",
			trades:    []barcountTestTrade{tradeInside},
			wantTotal: 1, wantIn: 1,
		},
		{
			name:      "multiple_in_window_all_counted",
			trades:    []barcountTestTrade{tradeInside, tradeInsideB, tradeInsideC},
			wantTotal: 3, wantIn: 3,
		},
		{
			name:      "all_trades_before_window_inwindow_zero",
			trades:    []barcountTestTrade{tradeBefore},
			wantTotal: 1, wantIn: 0,
		},
		{
			name:      "all_trades_after_window_inwindow_zero",
			trades:    []barcountTestTrade{tradeAfter},
			wantTotal: 1, wantIn: 0,
		},
		{
			name:      "mixed_inside_and_outside_total_exceeds_inwindow",
			trades:    []barcountTestTrade{tradeBefore, tradeInside, tradeAfter},
			wantTotal: 3, wantIn: 1,
		},
		{
			name:      "trades_on_both_sides_of_window_inwindow_zero",
			trades:    []barcountTestTrade{tradeBefore, tradeAfter},
			wantTotal: 2, wantIn: 0,
		},
		{
			name:      "start_boundary_inclusive",
			trades:    []barcountTestTrade{tradeOnStart},
			wantTotal: 1, wantIn: 1,
		},
		{
			name:      "end_boundary_inclusive",
			trades:    []barcountTestTrade{tradeOnEnd},
			wantTotal: 1, wantIn: 1,
		},
		{
			name:       "single_bar_fixture_trade_at_its_timestamp",
			timestamps: []int64{windowStart},
			trades:     []barcountTestTrade{{Num: 8, Entry: "2024-01-05 00:00", Exit: "2024-01-10 10:00"}},
			wantTotal:  1, wantIn: 1,
		},
		{
			name:       "zero_bar_fixture_returns_error",
			timestamps: []int64{},
			trades:     []barcountTestTrade{tradeInside},
			wantError:  true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			ts := tc.timestamps
			if ts == nil {
				ts = defaultWindow
			}
			fix := writeTempBarFixture(t, dir, "fixture.json", ts)
			csv := writeTempTradeCSV(t, dir, "trades.csv", tc.trades)
			gotTotal, gotIn, err := referenceTradeOverlap(fix, csv, tvref.TVTimezoneUTC)
			if tc.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotTotal != tc.wantTotal {
				t.Errorf("total = %d, want %d", gotTotal, tc.wantTotal)
			}
			if gotIn != tc.wantIn {
				t.Errorf("inWindow = %d, want %d", gotIn, tc.wantIn)
			}
		})
	}

	t.Run("missing_fixture_returns_error", func(t *testing.T) {
		dir := t.TempDir()
		csv := writeTempTradeCSV(t, dir, "trades.csv", []barcountTestTrade{tradeInside})
		_, _, err := referenceTradeOverlap(filepath.Join(dir, "nonexistent.json"), csv, tvref.TVTimezoneUTC)
		if err == nil {
			t.Error("expected error for missing fixture, got nil")
		}
	})

	t.Run("missing_csv_returns_error", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempBarFixture(t, dir, "fixture.json", defaultWindow)
		_, _, err := referenceTradeOverlap(fix, filepath.Join(dir, "nonexistent.csv"), tvref.TVTimezoneUTC)
		if err == nil {
			t.Error("expected error for missing CSV, got nil")
		}
	})
}

// TestInWindowTradeCount verifies the wrapper contract of inWindowTradeCount:
// boundary semantics, the degenerate single-bar fixture, and error propagation.
// Exhaustive window-filtering coverage lives in TestReferenceTradeOverlap.
//
// Fixture window: 2024-01-05 00:00 UTC → 2024-01-15 00:00 UTC.
func TestInWindowTradeCount(t *testing.T) {
	const (
		windowStart int64 = 1704412800 // 2024-01-05 00:00 UTC
		windowEnd   int64 = 1705276800 // 2024-01-15 00:00 UTC
	)
	defaultWindow := []int64{windowStart, windowEnd}

	tradeBefore := barcountTestTrade{Num: 9, Entry: "2023-12-25 10:00", Exit: "2023-12-28 10:00"}
	tradeInside := barcountTestTrade{Num: 10, Entry: "2024-01-07 10:00", Exit: "2024-01-10 10:00"}
	tradeOnStart := barcountTestTrade{Num: 11, Entry: "2024-01-05 00:00", Exit: "2024-01-06 10:00"}
	tradeOnEnd := barcountTestTrade{Num: 12, Entry: "2024-01-15 00:00", Exit: "2024-01-16 00:00"}

	cases := []struct {
		name       string
		timestamps []int64
		trades     []barcountTestTrade
		wantCount  int
		wantError  bool
	}{
		{
			name:      "returns_inwindow_count_not_total",
			trades:    []barcountTestTrade{tradeBefore, tradeInside},
			wantCount: 1,
		},
		{
			name:      "entry_on_start_boundary_is_counted",
			trades:    []barcountTestTrade{tradeOnStart},
			wantCount: 1,
		},
		{
			name:      "entry_on_end_boundary_is_counted",
			trades:    []barcountTestTrade{tradeOnEnd},
			wantCount: 1,
		},
		{
			name:       "single_bar_fixture_counts_trade_on_exact_timestamp",
			timestamps: []int64{windowStart},
			trades:     []barcountTestTrade{{Num: 13, Entry: "2024-01-05 00:00", Exit: "2024-01-10 10:00"}},
			wantCount:  1,
		},
		{
			name:       "zero_bar_fixture_returns_error",
			timestamps: []int64{},
			wantError:  true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			ts := tc.timestamps
			if ts == nil {
				ts = defaultWindow
			}
			fix := writeTempBarFixture(t, dir, "fixture.json", ts)
			csv := writeTempTradeCSV(t, dir, "trades.csv", tc.trades)
			got, err := inWindowTradeCount(fix, csv, tvref.TVTimezoneUTC)
			if tc.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantCount {
				t.Errorf("inWindowTradeCount = %d, want %d", got, tc.wantCount)
			}
		})
	}

	t.Run("missing_fixture_returns_error", func(t *testing.T) {
		dir := t.TempDir()
		csv := writeTempTradeCSV(t, dir, "trades.csv", []barcountTestTrade{tradeInside})
		_, err := inWindowTradeCount(filepath.Join(dir, "nonexistent.json"), csv, tvref.TVTimezoneUTC)
		if err == nil {
			t.Error("expected error for missing fixture, got nil")
		}
	})

	t.Run("missing_csv_returns_error", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempBarFixture(t, dir, "fixture.json", defaultWindow)
		_, err := inWindowTradeCount(fix, filepath.Join(dir, "nonexistent.csv"), tvref.TVTimezoneUTC)
		if err == nil {
			t.Error("expected error for missing CSV, got nil")
		}
	})
}
