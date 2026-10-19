package regression

import (
	"path/filepath"
	"testing"

	tvref "github.com/quant5-lab/runner/tests/regression/tv_reference"
)

// fixtureWindowDisjointFromReference reports whether the fixture's own time window
// and the reference trade history share no overlap: a non-empty reference CSV with
// zero in-window trades is the temporal-displacement signature of a synthetic
// placeholder fixture whose bar timestamps were never part of the real instrument's
// price history.
//
// Returns false when the reference CSV is empty (an empty reference cannot be
// disjoint from anything) or when at least one reference trade falls inside the
// fixture window. Returns an error only when the fixture or reference CSV cannot
// be read.
func fixtureWindowDisjointFromReference(fixturePath, csvPath string, tz tvref.TVTimezone) (bool, error) {
	total, inWindow, err := referenceTradeOverlap(fixturePath, csvPath, tz)
	if err != nil {
		return false, err
	}
	return total > 0 && inWindow == 0, nil
}

// TestFixtureWindowDisjointFromReference verifies temporal-displacement detection
// across all structural configurations, boundary conditions, and error paths.
//
// All test fixtures use TVTimezoneUTC. Fixture window (unless overridden):
// 2024-01-05 00:00 UTC → 2024-01-15 00:00 UTC.
func TestFixtureWindowDisjointFromReference(t *testing.T) {
	const (
		windowStart int64 = 1704412800 // 2024-01-05 00:00 UTC
		windowEnd   int64 = 1705276800 // 2024-01-15 00:00 UTC
	)
	defaultWindow := []int64{windowStart, windowEnd}

	tradeInside := barcountTestTrade{Num: 1, Entry: "2024-01-07 10:00", Exit: "2024-01-10 10:00"}
	tradeBefore := barcountTestTrade{Num: 2, Entry: "2023-12-25 10:00", Exit: "2023-12-28 10:00"}
	tradeAfter := barcountTestTrade{Num: 3, Entry: "2024-01-20 10:00", Exit: "2024-01-22 10:00"}
	tradeOnStart := barcountTestTrade{Num: 4, Entry: "2024-01-05 00:00", Exit: "2024-01-06 10:00"}
	tradeOnEnd := barcountTestTrade{Num: 5, Entry: "2024-01-15 00:00", Exit: "2024-01-16 00:00"}

	cases := []struct {
		name         string
		timestamps   []int64 // nil → defaultWindow; []int64{} → zero-bar fixture
		trades       []barcountTestTrade
		wantDisjoint bool
		wantError    bool
	}{
		{
			name:         "empty_reference_not_disjoint",
			trades:       []barcountTestTrade{},
			wantDisjoint: false,
		},
		{
			name:         "trade_inside_window_not_disjoint",
			trades:       []barcountTestTrade{tradeInside},
			wantDisjoint: false,
		},
		{
			name:         "trade_on_window_start_boundary_not_disjoint",
			trades:       []barcountTestTrade{tradeOnStart},
			wantDisjoint: false,
		},
		{
			name:         "trade_on_window_end_boundary_not_disjoint",
			trades:       []barcountTestTrade{tradeOnEnd},
			wantDisjoint: false,
		},
		{
			name:         "all_trades_before_window_disjoint",
			trades:       []barcountTestTrade{tradeBefore},
			wantDisjoint: true,
		},
		{
			name:         "all_trades_after_window_disjoint",
			trades:       []barcountTestTrade{tradeAfter},
			wantDisjoint: true,
		},
		{
			name:         "trades_on_both_sides_outside_disjoint",
			trades:       []barcountTestTrade{tradeBefore, tradeAfter},
			wantDisjoint: true,
		},
		{
			name:         "mixed_inside_and_outside_not_disjoint",
			trades:       []barcountTestTrade{tradeBefore, tradeInside},
			wantDisjoint: false,
		},
		{
			name:         "single_bar_fixture_trade_at_its_timestamp_not_disjoint",
			timestamps:   []int64{windowStart},
			trades:       []barcountTestTrade{{Num: 6, Entry: "2024-01-05 00:00", Exit: "2024-01-10 10:00"}},
			wantDisjoint: false,
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
			got, err := fixtureWindowDisjointFromReference(fix, csv, tvref.TVTimezoneUTC)
			if tc.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantDisjoint {
				t.Errorf("fixtureWindowDisjointFromReference = %v, want %v", got, tc.wantDisjoint)
			}
		})
	}

	t.Run("missing_fixture_returns_error", func(t *testing.T) {
		dir := t.TempDir()
		csv := writeTempTradeCSV(t, dir, "trades.csv", []barcountTestTrade{tradeInside})
		_, err := fixtureWindowDisjointFromReference(filepath.Join(dir, "nonexistent.json"), csv, tvref.TVTimezoneUTC)
		if err == nil {
			t.Error("expected error for missing fixture, got nil")
		}
	})

	t.Run("missing_csv_returns_error", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempBarFixture(t, dir, "fixture.json", defaultWindow)
		_, err := fixtureWindowDisjointFromReference(fix, filepath.Join(dir, "nonexistent.csv"), tvref.TVTimezoneUTC)
		if err == nil {
			t.Error("expected error for missing CSV, got nil")
		}
	})
}
