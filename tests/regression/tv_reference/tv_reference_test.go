package tv_reference

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func parseUTC(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04", s)
	if err != nil {
		panic(err)
	}
	return t
}

func makeTVTrade(datetime string, price float64) TVTrade {
	return makeTVTradeWithDirection(datetime, price, "long")
}

func makeTVTradeWithDirection(datetime string, price float64, direction string) TVTrade {
	return TVTrade{EntryUTC: parseUTC(datetime), EntryPrice: price, Direction: direction}
}

func makeRunnerTrade(datetime string, price float64) RunnerTrade {
	return makeRunnerTradeWithDirection(datetime, price, "long")
}

func makeRunnerTradeWithDirection(datetime string, price float64, direction string) RunnerTrade {
	return RunnerTrade{EntryUTC: parseUTC(datetime), EntryPrice: price, Direction: direction}
}

func TestFilterByEntryWindow(t *testing.T) {
	base := []TVTrade{
		makeTVTrade("2025-01-01 10:00", 100),
		makeTVTrade("2025-01-02 12:00", 101),
		makeTVTrade("2025-01-03 14:00", 102),
		makeTVTrade("2025-01-04 16:00", 103),
		makeTVTrade("2025-01-05 18:00", 104),
	}

	cases := []struct {
		name       string
		start      string
		end        string
		wantPrices []float64
	}{
		{"all_in_window", "2025-01-01 10:00", "2025-01-05 18:00", []float64{100, 101, 102, 103, 104}},
		{"none_before_window", "2025-01-06 00:00", "2025-01-07 00:00", nil},
		{"none_after_window", "2024-12-31 00:00", "2024-12-31 23:59", nil},
		{"start_inclusive", "2025-01-02 12:00", "2025-01-05 18:00", []float64{101, 102, 103, 104}},
		{"end_inclusive", "2025-01-01 10:00", "2025-01-03 14:00", []float64{100, 101, 102}},
		{"single_bar_window", "2025-01-02 12:00", "2025-01-02 12:00", []float64{101}},
		{"partial_overlap", "2025-01-02 00:00", "2025-01-04 00:00", []float64{101, 102}},
		{"empty_input", "2025-01-01 00:00", "2025-01-06 00:00", nil},
		{"inverted_window_start_after_end_returns_empty", "2025-01-05 00:00", "2025-01-01 00:00", nil},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			input := base
			if tc.name == "empty_input" {
				input = nil
			}
			got := FilterByEntryWindow(input, parseUTC(tc.start), parseUTC(tc.end))
			if len(got) != len(tc.wantPrices) {
				t.Fatalf("got %d trades, want %d", len(got), len(tc.wantPrices))
			}
			for i, want := range tc.wantPrices {
				if got[i].EntryPrice != want {
					t.Errorf("[%d] price = %.0f, want %.0f", i, got[i].EntryPrice, want)
				}
			}
		})
	}
}

func TestMatchExact(t *testing.T) {
	timeTol := time.Hour
	priceTol := 1.0

	cases := []struct {
		name           string
		runner         []RunnerTrade
		tv             []TVTrade
		wantMatched    int
		wantRunnerOnly int
		wantTVOnly     int
	}{
		{
			"both_empty",
			nil,
			nil,
			0, 0, 0,
		},
		{
			"empty_runner",
			nil,
			[]TVTrade{makeTVTrade("2025-01-01 10:00", 100)},
			0, 0, 1,
		},
		{
			"empty_tv",
			[]RunnerTrade{makeRunnerTrade("2025-01-01 10:00", 100)},
			nil,
			0, 1, 0,
		},
		{
			"all_matched_exact",
			[]RunnerTrade{
				makeRunnerTrade("2025-01-01 10:00", 100),
				makeRunnerTrade("2025-01-02 12:00", 200),
			},
			[]TVTrade{
				makeTVTrade("2025-01-01 10:00", 100),
				makeTVTrade("2025-01-02 12:00", 200),
			},
			2, 0, 0,
		},
		{
			"none_matched_time_diff",
			[]RunnerTrade{makeRunnerTrade("2025-01-01 10:00", 100)},
			[]TVTrade{makeTVTrade("2025-01-02 10:00", 100)},
			0, 1, 1,
		},
		{
			"none_matched_price_diff",
			[]RunnerTrade{makeRunnerTrade("2025-01-01 10:00", 100)},
			[]TVTrade{makeTVTrade("2025-01-01 10:00", 200)},
			0, 1, 1,
		},
		{
			"time_at_tolerance_boundary_passes",
			[]RunnerTrade{makeRunnerTrade("2025-01-01 10:00", 100)},
			[]TVTrade{makeTVTrade("2025-01-01 11:00", 100)},
			1, 0, 0,
		},
		{
			"time_one_minute_over_tolerance_fails",
			[]RunnerTrade{makeRunnerTrade("2025-01-01 10:00", 100)},
			[]TVTrade{makeTVTrade("2025-01-01 11:01", 100)},
			0, 1, 1,
		},
		{
			"price_at_tolerance_boundary_passes",
			[]RunnerTrade{makeRunnerTrade("2025-01-01 10:00", 100)},
			[]TVTrade{makeTVTrade("2025-01-01 10:00", 101)},
			1, 0, 0,
		},
		{
			"price_just_over_tolerance_fails",
			[]RunnerTrade{makeRunnerTrade("2025-01-01 10:00", 100)},
			[]TVTrade{makeTVTrade("2025-01-01 10:00", 101.01)},
			0, 1, 1,
		},
		{
			"time_tolerance_symmetric_tv_before_runner",
			[]RunnerTrade{makeRunnerTrade("2025-01-01 11:00", 100)},
			[]TVTrade{makeTVTrade("2025-01-01 10:00", 100)},
			1, 0, 0,
		},
		{
			"price_tolerance_symmetric_runner_above_tv",
			[]RunnerTrade{makeRunnerTrade("2025-01-01 10:00", 101)},
			[]TVTrade{makeTVTrade("2025-01-01 10:00", 100)},
			1, 0, 0,
		},
		{
			"direction_mismatch_fails",
			[]RunnerTrade{makeRunnerTradeWithDirection("2025-01-01 10:00", 100, "long")},
			[]TVTrade{makeTVTradeWithDirection("2025-01-01 10:00", 100, "short")},
			0, 1, 1,
		},
		{
			"direction_match_passes",
			[]RunnerTrade{makeRunnerTradeWithDirection("2025-01-01 10:00", 100, "long")},
			[]TVTrade{makeTVTradeWithDirection("2025-01-01 10:00", 100, "long")},
			1, 0, 0,
		},
		{
			"partial_match_unmatched_on_both_sides",
			[]RunnerTrade{
				makeRunnerTrade("2025-01-01 10:00", 100),
				makeRunnerTrade("2025-01-02 12:00", 200),
				makeRunnerTrade("2025-01-05 00:00", 500),
			},
			[]TVTrade{
				makeTVTrade("2025-01-01 10:00", 100),
				makeTVTrade("2025-01-02 12:00", 200),
				makeTVTrade("2025-01-03 00:00", 300),
			},
			2, 1, 1,
		},
		{
			"full_scan_finds_match_after_many_non_matches",
			func() []RunnerTrade { return []RunnerTrade{makeRunnerTrade("2025-01-12 00:00", 100)} }(),
			func() []TVTrade {
				trades := make([]TVTrade, 11)
				for i := 0; i < 10; i++ {
					trades[i] = makeTVTrade("2025-01-12 00:00", 999)
				}
				trades[10] = makeTVTrade("2025-01-12 00:00", 100)
				return trades
			}(),
			1, 0, 10,
		},
		{
			"matched_tv_slot_not_consumed_twice",
			[]RunnerTrade{
				makeRunnerTrade("2025-01-01 10:00", 100),
				makeRunnerTrade("2025-01-02 10:00", 200),
			},
			[]TVTrade{
				makeTVTrade("2025-01-01 10:00", 100),
				makeTVTrade("2025-01-02 10:00", 200),
				makeTVTrade("2025-01-03 10:00", 300),
			},
			2, 0, 1,
		},
		{
			"input_order_invariant",
			[]RunnerTrade{
				makeRunnerTrade("2025-01-03 10:00", 300),
				makeRunnerTrade("2025-01-01 10:00", 100),
				makeRunnerTrade("2025-01-02 10:00", 200),
			},
			[]TVTrade{
				makeTVTrade("2025-01-02 10:00", 200),
				makeTVTrade("2025-01-03 10:00", 300),
				makeTVTrade("2025-01-01 10:00", 100),
			},
			3, 0, 0,
		},
		{
			"duplicate_candidates_are_consumed_once",
			[]RunnerTrade{
				makeRunnerTrade("2025-01-01 10:00", 100),
				makeRunnerTrade("2025-01-01 10:00", 100),
			},
			[]TVTrade{makeTVTrade("2025-01-01 10:00", 100)},
			1, 1, 0,
		},
		{
			"blank_runner_direction_fails",
			[]RunnerTrade{makeRunnerTradeWithDirection("2025-01-01 10:00", 100, "")},
			[]TVTrade{makeTVTradeWithDirection("2025-01-01 10:00", 100, "short")},
			0, 1, 1,
		},
		{
			"blank_tv_direction_fails",
			[]RunnerTrade{makeRunnerTradeWithDirection("2025-01-01 10:00", 100, "long")},
			[]TVTrade{makeTVTradeWithDirection("2025-01-01 10:00", 100, "")},
			0, 1, 1,
		},
		{
			"unknown_direction_fails",
			[]RunnerTrade{makeRunnerTradeWithDirection("2025-01-01 10:00", 100, "buy")},
			[]TVTrade{makeTVTradeWithDirection("2025-01-01 10:00", 100, "long")},
			0, 1, 1,
		},
		{
			"direction_compare_trims_case",
			[]RunnerTrade{makeRunnerTradeWithDirection("2025-01-01 10:00", 100, " Long ")},
			[]TVTrade{makeTVTradeWithDirection("2025-01-01 10:00", 100, "long")},
			1, 0, 0,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			matched, runnerOnly, tvOnly := MatchExact(tc.runner, tc.tv, timeTol, priceTol)
			if matched != tc.wantMatched {
				t.Errorf("matched = %d, want %d", matched, tc.wantMatched)
			}
			if runnerOnly != tc.wantRunnerOnly {
				t.Errorf("runnerOnly = %d, want %d", runnerOnly, tc.wantRunnerOnly)
			}
			if tvOnly != tc.wantTVOnly {
				t.Errorf("tvOnly = %d, want %d", tvOnly, tc.wantTVOnly)
			}
		})
	}
}

func TestMatchExact_ZeroTimeTolerance(t *testing.T) {
	priceTol := 1.0

	t.Run("identical_timestamps_match", func(t *testing.T) {
		ts := parseUTC("2025-06-01 09:00")
		runner := []RunnerTrade{{EntryUTC: ts, EntryPrice: 100, Direction: "long"}}
		tv := []TVTrade{{EntryUTC: ts, EntryPrice: 100, Direction: "long"}}
		matched, ro, to := MatchExact(runner, tv, 0, priceTol)
		if matched != 1 || ro != 0 || to != 0 {
			t.Errorf("matched=%d ro=%d to=%d, want 1/0/0", matched, ro, to)
		}
	})

	t.Run("one_second_ahead_no_match", func(t *testing.T) {
		ts := parseUTC("2025-06-01 09:00")
		runner := []RunnerTrade{{EntryUTC: ts, EntryPrice: 100, Direction: "long"}}
		tv := []TVTrade{{EntryUTC: ts.Add(time.Second), EntryPrice: 100, Direction: "long"}}
		matched, ro, to := MatchExact(runner, tv, 0, priceTol)
		if matched != 0 || ro != 1 || to != 1 {
			t.Errorf("matched=%d ro=%d to=%d, want 0/1/1", matched, ro, to)
		}
	})

	t.Run("one_second_behind_no_match", func(t *testing.T) {
		ts := parseUTC("2025-06-01 09:00")
		runner := []RunnerTrade{{EntryUTC: ts, EntryPrice: 100, Direction: "long"}}
		tv := []TVTrade{{EntryUTC: ts.Add(-time.Second), EntryPrice: 100, Direction: "long"}}
		matched, ro, to := MatchExact(runner, tv, 0, priceTol)
		if matched != 0 || ro != 1 || to != 1 {
			t.Errorf("matched=%d ro=%d to=%d, want 0/1/1", matched, ro, to)
		}
	})
}

func TestMatchExact_ZeroPriceTolerance(t *testing.T) {
	timeTol := time.Hour

	t.Run("exact_same_price_matches", func(t *testing.T) {
		runner := []RunnerTrade{makeRunnerTrade("2025-06-01 09:00", 100.00)}
		tv := []TVTrade{makeTVTrade("2025-06-01 09:00", 100.00)}
		matched, ro, to := MatchExact(runner, tv, timeTol, 0)
		if matched != 1 || ro != 0 || to != 0 {
			t.Errorf("matched=%d ro=%d to=%d, want 1/0/0", matched, ro, to)
		}
	})

	t.Run("price_above_by_smallest_representable_diff_no_match", func(t *testing.T) {
		runner := []RunnerTrade{makeRunnerTrade("2025-06-01 09:00", 100.000)}
		tv := []TVTrade{makeTVTrade("2025-06-01 09:00", 100.001)}
		matched, ro, to := MatchExact(runner, tv, timeTol, 0)
		if matched != 0 || ro != 1 || to != 1 {
			t.Errorf("matched=%d ro=%d to=%d, want 0/1/1", matched, ro, to)
		}
	})

	t.Run("price_below_by_smallest_representable_diff_no_match", func(t *testing.T) {
		runner := []RunnerTrade{makeRunnerTrade("2025-06-01 09:00", 99.999)}
		tv := []TVTrade{makeTVTrade("2025-06-01 09:00", 100.000)}
		matched, ro, to := MatchExact(runner, tv, timeTol, 0)
		if matched != 0 || ro != 1 || to != 1 {
			t.Errorf("matched=%d ro=%d to=%d, want 0/1/1", matched, ro, to)
		}
	})
}

func TestDirectionsMatch(t *testing.T) {
	cases := []struct {
		name   string
		runner string
		tv     string
		want   bool
	}{
		{"long", "long", "long", true},
		{"short", "short", "short", true},
		{"case_and_space_normalized", " Long ", "long", true},
		{"opposite_direction", "long", "short", false},
		{"blank_runner", "", "long", false},
		{"blank_tv", "long", "", false},
		{"blank_both", "", "", false},
		{"unknown_runner", "buy", "long", false},
		{"unknown_tv", "long", "sell", false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := directionsMatch(tc.runner, tc.tv); got != tc.want {
				t.Fatalf("directionsMatch(%q, %q) = %v, want %v", tc.runner, tc.tv, got, tc.want)
			}
		})
	}
}

func TestMatchExactCountsEveryUnmatchedTrade(t *testing.T) {
	runner := []RunnerTrade{
		makeRunnerTradeWithDirection("2025-01-01 10:00", 100, "long"),
		makeRunnerTradeWithDirection("2025-01-02 10:00", 200, "short"),
		makeRunnerTradeWithDirection("2025-01-03 10:00", 300, "long"),
	}
	tv := []TVTrade{
		makeTVTradeWithDirection("2025-01-01 10:00", 100, "long"),
		makeTVTradeWithDirection("2025-01-02 10:00", 200, "long"),
		makeTVTradeWithDirection("2025-01-04 10:00", 400, "short"),
	}

	matched, runnerOnly, tvOnly := MatchExact(runner, tv, time.Hour, 1)
	if matched != 1 || runnerOnly != 2 || tvOnly != 2 {
		t.Fatalf("matched=%d runnerOnly=%d tvOnly=%d, want 1/2/2", matched, runnerOnly, tvOnly)
	}
}

func TestMatchSize(t *testing.T) {
	runner := []RunnerTrade{
		{EntryUTC: parseUTC("2025-01-01 10:00"), EntryPrice: 100, Direction: "long", Size: 2.29560},
		{EntryUTC: parseUTC("2025-01-02 10:00"), EntryPrice: 200, Direction: "short", Size: 10.00},
	}
	tv := []TVTrade{
		{EntryUTC: parseUTC("2025-01-01 10:00"), EntryPrice: 100, Direction: "long", Size: 2.29560},
		{EntryUTC: parseUTC("2025-01-02 10:00"), EntryPrice: 200, Direction: "short", Size: 10.10},
	}

	matched, mismatched, maxResidual := MatchSize(runner, tv, time.Minute, 0.01, 0.005)
	if matched != 2 {
		t.Fatalf("matched = %d, want 2", matched)
	}
	if mismatched != 1 {
		t.Fatalf("mismatched = %d, want 1", mismatched)
	}
	if maxResidual <= 0.009 || maxResidual >= 0.010 {
		t.Fatalf("maxResidual = %.8f, want approximately 0.00990099", maxResidual)
	}
}

func TestSizeResidualUsesTVDenominatorFloor(t *testing.T) {
	tests := []struct {
		name       string
		runnerSize float64
		tvSize     float64
		want       float64
	}{
		{"exact", 1, 1, 0},
		{"normalizes_by_tv_size", 105, 100, 0.05},
		{"small_tv_size_uses_floor", 0.002, 0.001, 0.001},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := SizeResidual(tt.runnerSize, tt.tvSize)
			if math.Abs(got-tt.want) > 1e-12 {
				t.Fatalf("SizeResidual(%v, %v) = %.12f, want %.12f", tt.runnerSize, tt.tvSize, got, tt.want)
			}
		})
	}
}

func TestParseDateTime_UTC(t *testing.T) {
	cases := []struct {
		input   string
		wantUTC string
	}{
		{"2025-01-15 13:00", "2025-01-15 13:00"},
		{"2025-07-04 00:00", "2025-07-04 00:00"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			got, err := parseDateTime(tc.input, TVTimezoneUTC)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			want, _ := time.Parse("2006-01-02 15:04", tc.wantUTC)
			if !got.Equal(want) {
				t.Errorf("got %s, want %s", got.UTC().Format("2006-01-02 15:04"), want.Format("2006-01-02 15:04"))
			}
		})
	}
}

func TestParseDateTime_Moscow(t *testing.T) {
	// Moscow is UTC+3, no daylight saving — always subtract 3 h.
	cases := []struct {
		input    string
		wantUTCh int
	}{
		{"2025-01-15 13:00", 10},
		{"2025-07-15 03:00", 0},
		{"2025-12-31 23:00", 20},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			got, err := parseDateTime(tc.input, TVTimezoneMoscow)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.UTC().Hour() != tc.wantUTCh {
				t.Errorf("UTC hour = %d, want %d  (input %s)", got.UTC().Hour(), tc.wantUTCh, tc.input)
			}
		})
	}
}

func TestParseDateTime_NewYork_EDT(t *testing.T) {
	// July: America/New_York is EDT = UTC-4.
	// "2025-07-15 13:00" local → 17:00 UTC.
	got, err := parseDateTime("2025-07-15 13:00", TVTimezoneNewYork)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2025, 7, 15, 17, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("EDT got %s, want %s", got.UTC().Format(time.RFC3339), want.Format(time.RFC3339))
	}
}

func TestParseDateTime_NewYork_EST(t *testing.T) {
	// January: America/New_York is EST = UTC-5.
	// "2025-01-15 13:00" local → 18:00 UTC.
	got, err := parseDateTime("2025-01-15 13:00", TVTimezoneNewYork)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2025, 1, 15, 18, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("EST got %s, want %s", got.UTC().Format(time.RFC3339), want.Format(time.RFC3339))
	}
}

func TestParseDateTime_InvalidFormat(t *testing.T) {
	datetimeCases := []string{
		"not-a-date",
		"2025/01/15 13:00",
		"",
		"2025-01-15 bad",
	}
	dateOnlyCases := []string{
		"2025-13-01",
		"2025-00-15",
		"2025-02-30",
	}
	tzs := []TVTimezone{TVTimezoneUTC, TVTimezoneMoscow, TVTimezoneNewYork}
	for _, tz := range tzs {
		for _, s := range append(datetimeCases, dateOnlyCases...) {
			if _, err := parseDateTime(s, tz); err == nil {
				t.Errorf("tz=%d input=%q: expected error, got nil", tz, s)
			}
		}
	}
}

func TestParseDateTime_DateOnly(t *testing.T) {
	type dateCase struct{ input, wantUTC string }
	run := func(tz TVTimezone, cases []dateCase) {
		t.Helper()
		for _, tc := range cases {
			tc := tc
			t.Run(fmt.Sprintf("tz=%d/%s", tz, tc.input), func(t *testing.T) {
				got, err := parseDateTime(tc.input, tz)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				want, _ := time.Parse("2006-01-02 15:04", tc.wantUTC)
				if !got.UTC().Equal(want) {
					t.Errorf("got %s, want %s", got.UTC().Format("2006-01-02 15:04"), want.Format("2006-01-02 15:04"))
				}
			})
		}
	}

	// UTC: midnight is midnight regardless of timezone parameter.
	run(TVTimezoneUTC, []dateCase{
		{"2025-01-15", "2025-01-15 00:00"},
		{"2018-02-01", "2018-02-01 00:00"},
		{"2024-02-29", "2024-02-29 00:00"}, // leap year
		{"1970-01-01", "1970-01-01 00:00"}, // Unix epoch
	})

	// Moscow: UTC+3 with no DST since 2014 — always subtract exactly 3h.
	// Midnight MSK falls on the preceding UTC calendar day.
	run(TVTimezoneMoscow, []dateCase{
		{"2025-01-15", "2025-01-14 21:00"}, // winter
		{"2025-07-04", "2025-07-03 21:00"}, // summer — offset identical to winter (no DST)
		{"2026-01-01", "2025-12-31 21:00"}, // year boundary: midnight Jan 1 MSK → Dec 31 21:00 UTC
		{"2024-02-29", "2024-02-28 21:00"}, // leap-year date
		{"1970-01-01", "1969-12-31 21:00"}, // Unix epoch boundary
	})

	// New York: DST-aware (EST = UTC-5 in winter, EDT = UTC-4 in summer).
	// Midnight is always before any intra-day DST switch, so the offset at midnight
	// is determined by which regime the date falls in.
	run(TVTimezoneNewYork, []dateCase{
		{"2025-01-15", "2025-01-15 05:00"}, // EST (UTC-5)
		{"2025-07-04", "2025-07-04 04:00"}, // EDT (UTC-4)
		// DST transitions: spring-forward is 2025-03-09 02:00 EST→EDT;
		// fall-back is 2025-11-02 02:00 EDT→EST. Midnight on each day
		// precedes the switch, so the offset matches the regime before the switch.
		{"2025-03-09", "2025-03-09 05:00"}, // spring-forward day — midnight still EST
		{"2025-11-02", "2025-11-02 04:00"}, // fall-back day   — midnight still EDT
		{"2024-02-29", "2024-02-29 05:00"}, // leap-year date (EST)
		{"2026-01-01", "2026-01-01 05:00"}, // year boundary (EST)
	})
}

// TestParseDateTime_DateOnly_EquivalentToMidnightDatetime verifies the contractual
// invariant of parseDateOnly: for every supported timezone, parsing a bare date is
// exactly equivalent to parsing that date suffixed with " 00:00" in datetime format.
// This invariant is the reason the date-only branch must honour the tz parameter —
// without it the two forms produce different UTC timestamps.
func TestParseDateTime_DateOnly_EquivalentToMidnightDatetime(t *testing.T) {
	dates := []string{
		"2025-01-15", // EST, MSK winter
		"2025-07-04", // EDT, MSK summer
		"2025-03-09", // NY DST spring-forward day
		"2025-11-02", // NY DST fall-back day
		"2024-02-29", // leap year
		"2026-01-01", // year boundary
		"1970-01-01", // Unix epoch
	}
	tzs := []TVTimezone{TVTimezoneUTC, TVTimezoneMoscow, TVTimezoneNewYork}
	for _, tz := range tzs {
		for _, date := range dates {
			tz, date := tz, date
			t.Run(fmt.Sprintf("tz=%d/%s", tz, date), func(t *testing.T) {
				dateOnly, err := parseDateTime(date, tz)
				if err != nil {
					t.Fatalf("date-only parse(%q, tz=%d): %v", date, tz, err)
				}
				midnight, err := parseDateTime(date+" 00:00", tz)
				if err != nil {
					t.Fatalf("midnight parse(%q, tz=%d): %v", date+" 00:00", tz, err)
				}
				if !dateOnly.UTC().Equal(midnight.UTC()) {
					t.Errorf(
						"date-only %q → %s, midnight %q → %s; must be equal",
						date, dateOnly.UTC().Format("2006-01-02 15:04:05"),
						date+" 00:00", midnight.UTC().Format("2006-01-02 15:04:05"),
					)
				}
			})
		}
	}
}

func TestAssembleTrades(t *testing.T) {
	d := func(s string) time.Time { return parseUTC(s) }

	t.Run("entry_exit_pair_forms_closed_trade", func(t *testing.T) {
		rows := []csvRow{
			{num: 1, rowType: "Entry Long", dt: d("2025-01-01 10:00"), price: 100, size: 12.345, netPnL: 42.5},
			{num: 1, rowType: "Exit Long", dt: d("2025-01-01 14:00"), price: 110, netPnL: 42.5},
		}
		trades := assembleTrades(rows)
		if len(trades) != 1 {
			t.Fatalf("got %d trades, want 1", len(trades))
		}
		tr := trades[0]
		if tr.EntryPrice != 100 || tr.ExitPrice != 110 {
			t.Errorf("prices: entry=%.0f exit=%.0f, want 100/110", tr.EntryPrice, tr.ExitPrice)
		}
		if tr.Direction != "long" {
			t.Errorf("direction = %q, want long", tr.Direction)
		}
		if tr.ExitUTC.IsZero() {
			t.Error("exit time must not be zero for a closed trade")
		}
		if tr.NetPnL != 42.5 {
			t.Errorf("NetPnL = %.2f, want 42.5 (from entry row)", tr.NetPnL)
		}
		if tr.Size != 12.345 {
			t.Errorf("Size = %.3f, want 12.345 (from entry row)", tr.Size)
		}
	})

	t.Run("net_pnl_absent_defaults_to_zero", func(t *testing.T) {
		rows := []csvRow{
			{num: 1, rowType: "Entry Long", dt: d("2025-01-01 10:00"), price: 100},
			{num: 1, rowType: "Exit Long", dt: d("2025-01-01 14:00"), price: 110},
		}
		trades := assembleTrades(rows)
		if trades[0].NetPnL != 0 {
			t.Errorf("NetPnL = %.2f, want 0 when column absent", trades[0].NetPnL)
		}
	})

	t.Run("entry_without_exit_forms_open_trade", func(t *testing.T) {
		rows := []csvRow{
			{num: 2, rowType: "Entry Short", dt: d("2025-01-02 10:00"), price: 200},
		}
		trades := assembleTrades(rows)
		if len(trades) != 1 {
			t.Fatalf("got %d trades, want 1", len(trades))
		}
		if trades[0].Direction != "short" {
			t.Errorf("direction = %q, want short", trades[0].Direction)
		}
		if !trades[0].ExitUTC.IsZero() {
			t.Error("exit time must be zero for an open trade")
		}
	})

	t.Run("exit_without_entry_is_skipped", func(t *testing.T) {
		rows := []csvRow{
			{num: 3, rowType: "Exit Short", dt: d("2025-01-03 10:00"), price: 300},
		}
		trades := assembleTrades(rows)
		if len(trades) != 0 {
			t.Fatalf("got %d trades, want 0 (orphan exit skipped)", len(trades))
		}
	})

	t.Run("insertion_order_preserved", func(t *testing.T) {
		rows := []csvRow{
			{num: 10, rowType: "Entry Long", dt: d("2025-01-10 10:00"), price: 10},
			{num: 10, rowType: "Exit Long", dt: d("2025-01-10 14:00"), price: 11},
			{num: 20, rowType: "Entry Short", dt: d("2025-01-20 10:00"), price: 20},
			{num: 20, rowType: "Exit Short", dt: d("2025-01-20 14:00"), price: 19},
		}
		trades := assembleTrades(rows)
		if len(trades) != 2 {
			t.Fatalf("got %d trades, want 2", len(trades))
		}
		if trades[0].Number != 10 || trades[1].Number != 20 {
			t.Errorf("order: got %d,%d, want 10,20", trades[0].Number, trades[1].Number)
		}
	})

	t.Run("multiple_mixed_entry_exit_rows_same_number", func(t *testing.T) {
		// Last Entry/Exit row wins for each trade number.
		rows := []csvRow{
			{num: 5, rowType: "Entry Long", dt: d("2025-01-05 09:00"), price: 50},
			{num: 5, rowType: "Exit Long", dt: d("2025-01-05 11:00"), price: 55},
			{num: 5, rowType: "Exit Long", dt: d("2025-01-05 13:00"), price: 60},
		}
		trades := assembleTrades(rows)
		if len(trades) != 1 {
			t.Fatalf("got %d trades, want 1", len(trades))
		}
		if trades[0].ExitPrice != 60 {
			t.Errorf("exit price = %.0f, want 60 (last exit row wins)", trades[0].ExitPrice)
		}
	})
}

func TestIndexColumns(t *testing.T) {
	cases := []struct {
		name     string
		header   []string
		tradeNum int
		rowType  int
		datetime int
		price    int
		size     int
		netPnL   int
	}{
		{
			"standard_no_net_pnl",
			[]string{"Trade number", "Type", "Date and time", "Price", "Contracts"},
			0, 1, 2, 3, 4, -1,
		},
		{
			"bom_prefix_on_trade_number",
			[]string{"\xef\xbb\xbfTrade number", "Type", "Date and time", "Price USDT"},
			0, 1, 2, 3, -1, -1,
		},
		{
			"missing_type_returns_negative_one",
			[]string{"Trade number", "Date and time", "Price"},
			0, -1, 1, 2, -1, -1,
		},
		{
			"price_takes_first_matching_column",
			[]string{"Trade number", "Type", "Date and time", "Price USD", "Price USDT"},
			0, 1, 2, 3, -1, -1,
		},
		{
			"net_pnl_usdt_detected",
			[]string{"Trade number", "Type", "Date and time", "Price USDT", "Size (qty)", "Net PnL USDT"},
			0, 1, 2, 3, 4, 5,
		},
		{
			"net_pnl_usd_detected",
			[]string{"Trade number", "Type", "Date and time", "Price USD", "Net PnL USD"},
			0, 1, 2, 3, -1, 4,
		},
		{
			"net_pnl_rub_detected",
			[]string{"Trade number", "Type", "Date and time", "Price RUB", "Size (qty)", "Net PnL RUB"},
			0, 1, 2, 3, 4, 5,
		},
		{
			"net_pnl_takes_first_matching_column",
			[]string{"Trade number", "Type", "Date and time", "Price", "Net PnL USD", "Net PnL %"},
			0, 1, 2, 3, -1, 4,
		},
		{
			"quantity_alias_detected",
			[]string{"Trade number", "Type", "Date and time", "Price", "Quantity"},
			0, 1, 2, 3, 4, -1,
		},
		{
			"size_alias_detected",
			[]string{"Trade number", "Type", "Date and time", "Price", "Size"},
			0, 1, 2, 3, 4, -1,
		},
		{
			"size_underscore_qty_normalized",
			[]string{"Trade number", "Type", "Date and time", "Price", "Size_(qty)"},
			0, 1, 2, 3, 4, -1,
		},
		{
			"net_pnl_no_currency_suffix_detected",
			[]string{"Trade number", "Type", "Date and time", "Price", "Net PnL"},
			0, 1, 2, 3, -1, 4,
		},
		{
			"all_columns_missing",
			[]string{"Foo", "Bar", "Baz"},
			-1, -1, -1, -1, -1, -1,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			idx := indexColumns(tc.header)
			if idx.tradeNum != tc.tradeNum {
				t.Errorf("tradeNum = %d, want %d", idx.tradeNum, tc.tradeNum)
			}
			if idx.rowType != tc.rowType {
				t.Errorf("rowType = %d, want %d", idx.rowType, tc.rowType)
			}
			if idx.datetime != tc.datetime {
				t.Errorf("datetime = %d, want %d", idx.datetime, tc.datetime)
			}
			if idx.price != tc.price {
				t.Errorf("price = %d, want %d", idx.price, tc.price)
			}
			if idx.size != tc.size {
				t.Errorf("size = %d, want %d", idx.size, tc.size)
			}
			if idx.netPnL != tc.netPnL {
				t.Errorf("netPnL = %d, want %d", idx.netPnL, tc.netPnL)
			}
		})
	}
}

func TestParseTrades_FullCSVRoundTrip(t *testing.T) {
	t.Run("without_net_pnl_column", func(t *testing.T) {
		rawCSV := `Trade number,Type,Date and time,Price,Contracts
1,Entry Long,2025-01-01 10:00,100,1
1,Exit Long,2025-01-01 14:00,110,1
2,Entry Short,2025-01-02 10:00,120,1
2,Exit Short,2025-01-02 15:00,115,1
`
		tmpPath := filepath.Join(t.TempDir(), "trades.csv")
		if err := os.WriteFile(tmpPath, []byte(rawCSV), 0644); err != nil {
			t.Fatalf("write temp CSV: %v", err)
		}

		trades, err := LoadTrades(tmpPath, TVTimezoneUTC)
		if err != nil {
			t.Fatalf("LoadTrades: %v", err)
		}
		if len(trades) != 2 {
			t.Fatalf("got %d trades, want 2", len(trades))
		}
		if trades[0].Direction != "long" || trades[1].Direction != "short" {
			t.Errorf("directions: %q/%q, want long/short", trades[0].Direction, trades[1].Direction)
		}
		if trades[0].EntryPrice != 100 || trades[0].ExitPrice != 110 {
			t.Errorf("trade[0] prices: entry=%.0f exit=%.0f, want 100/110", trades[0].EntryPrice, trades[0].ExitPrice)
		}
		if trades[0].NetPnL != 0 {
			t.Errorf("trade[0] NetPnL = %.2f, want 0 (column absent)", trades[0].NetPnL)
		}
		if trades[0].Size != 1 {
			t.Errorf("trade[0] Size = %.2f, want 1", trades[0].Size)
		}
	})

	t.Run("with_net_pnl_column", func(t *testing.T) {
		rawCSV := `Trade number,Type,Date and time,Price USD,Size (qty),Net PnL USD
1,Entry Long,2025-01-01 10:00,100,2.29560,9.80
1,Exit Long,2025-01-01 14:00,110,2.29560,9.80
2,Entry Short,2025-01-02 10:00,120,4,-4.85
2,Exit Short,2025-01-02 15:00,115,4,-4.85
`
		tmpPath := filepath.Join(t.TempDir(), "trades_pnl.csv")
		if err := os.WriteFile(tmpPath, []byte(rawCSV), 0644); err != nil {
			t.Fatalf("write temp CSV: %v", err)
		}

		trades, err := LoadTrades(tmpPath, TVTimezoneUTC)
		if err != nil {
			t.Fatalf("LoadTrades: %v", err)
		}
		if len(trades) != 2 {
			t.Fatalf("got %d trades, want 2", len(trades))
		}
		if trades[0].NetPnL != 9.80 {
			t.Errorf("trade[0] NetPnL = %.2f, want 9.80", trades[0].NetPnL)
		}
		if trades[1].NetPnL != -4.85 {
			t.Errorf("trade[1] NetPnL = %.2f, want -4.85", trades[1].NetPnL)
		}
		if trades[0].Size != 2.29560 || trades[1].Size != 4 {
			t.Errorf("sizes = %.5f/%.5f, want 2.29560/4", trades[0].Size, trades[1].Size)
		}
	})
}

func TestParseTrades_SizeColumnAliases(t *testing.T) {
	aliases := []string{"Size", "Quantity", "Contracts", "Size (qty)", "Size_(qty)", "qty"}
	for _, alias := range aliases {
		alias := alias
		t.Run(alias, func(t *testing.T) {
			rawCSV := "Trade number,Type,Date and time,Price," + alias + "\n" +
				"1,Entry Long,2025-01-01 10:00,100,3.25\n" +
				"1,Exit Long,2025-01-01 14:00,110,3.25\n"
			tmpPath := filepath.Join(t.TempDir(), "trades.csv")
			if err := os.WriteFile(tmpPath, []byte(rawCSV), 0644); err != nil {
				t.Fatalf("write temp CSV: %v", err)
			}

			trades, err := LoadTrades(tmpPath, TVTimezoneUTC)
			if err != nil {
				t.Fatalf("LoadTrades: %v", err)
			}
			if len(trades) != 1 {
				t.Fatalf("got %d trades, want 1", len(trades))
			}
			if trades[0].Size != 3.25 {
				t.Fatalf("Size = %.2f, want 3.25", trades[0].Size)
			}
		})
	}
}

func TestParseTrades_MissingRequiredColumns_ReturnsError(t *testing.T) {
	csv := `Foo,Bar,Baz
1,x,y
`
	tmpPath := filepath.Join(t.TempDir(), "bad.csv")
	os.WriteFile(tmpPath, []byte(csv), 0644)

	_, err := LoadTrades(tmpPath, TVTimezoneUTC)
	if err == nil {
		t.Fatal("expected error for CSV missing required columns, got nil")
	}
	if !strings.Contains(err.Error(), "missing required columns") {
		t.Errorf("error should mention missing columns, got: %v", err)
	}
}

func TestParseTrades_MoscowTimezone_ConvertsToUTC(t *testing.T) {
	// Entry at 13:00 MSK = 10:00 UTC (Moscow is UTC+3, no DST).
	csv := `Trade number,Type,Date and time,Price
1,Entry Long,2025-06-15 13:00,100
1,Exit Long,2025-06-15 17:00,110
`
	tmpPath := filepath.Join(t.TempDir(), "mos.csv")
	os.WriteFile(tmpPath, []byte(csv), 0644)

	trades, err := LoadTrades(tmpPath, TVTimezoneMoscow)
	if err != nil {
		t.Fatalf("LoadTrades: %v", err)
	}
	if len(trades) != 1 {
		t.Fatalf("got %d trades, want 1", len(trades))
	}
	if trades[0].EntryUTC.UTC().Hour() != 10 {
		t.Errorf("entry UTC hour = %d, want 10 (13:00 MSK = 10:00 UTC)", trades[0].EntryUTC.UTC().Hour())
	}
}

func TestParseTrades_NewYorkTimezone_HonoursDST(t *testing.T) {
	// Summer (EDT = UTC-4): 13:00 EDT → 17:00 UTC.
	// Winter (EST = UTC-5): 13:00 EST → 18:00 UTC.
	csv := `Trade number,Type,Date and time,Price
1,Entry Long,2025-07-15 13:00,100
1,Exit Long,2025-07-15 15:00,110
2,Entry Long,2025-01-15 13:00,200
2,Exit Long,2025-01-15 15:00,210
`
	tmpPath := filepath.Join(t.TempDir(), "nyc.csv")
	os.WriteFile(tmpPath, []byte(csv), 0644)

	trades, err := LoadTrades(tmpPath, TVTimezoneNewYork)
	if err != nil {
		t.Fatalf("LoadTrades: %v", err)
	}
	if len(trades) != 2 {
		t.Fatalf("got %d trades, want 2", len(trades))
	}
	if trades[0].EntryUTC.UTC().Hour() != 17 {
		t.Errorf("summer entry UTC hour = %d, want 17 (13:00 EDT = 17:00 UTC)", trades[0].EntryUTC.UTC().Hour())
	}
	if trades[1].EntryUTC.UTC().Hour() != 18 {
		t.Errorf("winter entry UTC hour = %d, want 18 (13:00 EST = 18:00 UTC)", trades[1].EntryUTC.UTC().Hour())
	}
}

// TestParseTrades_DateOnlyFormat_MoscowTimezone mirrors TestParseTrades_MoscowTimezone_ConvertsToUTC
// but exercises the date-only branch (YYYY-MM-DD without a time component), as exported
// by TV's Strategy Tester for monthly-bar strategies.
func TestParseTrades_DateOnlyFormat_MoscowTimezone(t *testing.T) {
	rawCSV := `Trade number,Type,Date and time,Price
1,Entry Long,2025-07-01,100
1,Exit Long,2025-08-01,110
`
	tmpPath := filepath.Join(t.TempDir(), "mos_dateonly.csv")
	if err := os.WriteFile(tmpPath, []byte(rawCSV), 0644); err != nil {
		t.Fatalf("write temp CSV: %v", err)
	}

	trades, err := LoadTrades(tmpPath, TVTimezoneMoscow)
	if err != nil {
		t.Fatalf("LoadTrades: %v", err)
	}
	if len(trades) != 1 {
		t.Fatalf("got %d trades, want 1", len(trades))
	}

	wantEntry := time.Date(2025, 6, 30, 21, 0, 0, 0, time.UTC) // midnight Jul 1 MSK → Jun 30 21:00 UTC
	wantExit := time.Date(2025, 7, 31, 21, 0, 0, 0, time.UTC)  // midnight Aug 1 MSK → Jul 31 21:00 UTC
	if !trades[0].EntryUTC.Equal(wantEntry) {
		t.Errorf("entry = %s, want %s (date-only 2025-07-01 MSK)", trades[0].EntryUTC.UTC().Format(time.RFC3339), wantEntry.Format(time.RFC3339))
	}
	if !trades[0].ExitUTC.Equal(wantExit) {
		t.Errorf("exit = %s, want %s (date-only 2025-08-01 MSK)", trades[0].ExitUTC.UTC().Format(time.RFC3339), wantExit.Format(time.RFC3339))
	}
}

// TestParseTrades_DateOnlyFormat_NewYorkTimezone mirrors TestParseTrades_NewYorkTimezone_HonoursDST
// but exercises the date-only branch with DST-sensitive dates: a summer date (EDT) and a winter
// date (EST) both round-tripped to their correct UTC midnight.
func TestParseTrades_DateOnlyFormat_NewYorkTimezone(t *testing.T) {
	rawCSV := `Trade number,Type,Date and time,Price
1,Entry Long,2025-07-01,100
1,Exit Long,2025-08-01,110
2,Entry Long,2025-01-15,200
2,Exit Long,2025-02-15,210
`
	tmpPath := filepath.Join(t.TempDir(), "ny_dateonly.csv")
	if err := os.WriteFile(tmpPath, []byte(rawCSV), 0644); err != nil {
		t.Fatalf("write temp CSV: %v", err)
	}

	trades, err := LoadTrades(tmpPath, TVTimezoneNewYork)
	if err != nil {
		t.Fatalf("LoadTrades: %v", err)
	}
	if len(trades) != 2 {
		t.Fatalf("got %d trades, want 2", len(trades))
	}

	wantSummerEntry := time.Date(2025, 7, 1, 4, 0, 0, 0, time.UTC)  // midnight EDT → 04:00 UTC
	wantWinterEntry := time.Date(2025, 1, 15, 5, 0, 0, 0, time.UTC) // midnight EST → 05:00 UTC
	if !trades[0].EntryUTC.Equal(wantSummerEntry) {
		t.Errorf("summer entry = %s, want %s (date-only 2025-07-01 EDT)", trades[0].EntryUTC.UTC().Format(time.RFC3339), wantSummerEntry.Format(time.RFC3339))
	}
	if !trades[1].EntryUTC.Equal(wantWinterEntry) {
		t.Errorf("winter entry = %s, want %s (date-only 2025-01-15 EST)", trades[1].EntryUTC.UTC().Format(time.RFC3339), wantWinterEntry.Format(time.RFC3339))
	}
}

func TestLoadRunnerTrades(t *testing.T) {
	t.Run("closed_and_open_trades_concatenated", func(t *testing.T) {
		data := `{
			"result": {
				"trades": [
					{"entryTime": 1700000000, "entryPrice": 100.0, "direction": "long"},
					{"entryTime": 1700003600, "entryPrice": 200.0, "direction": "short"}
				],
				"openTrades": [
					{"entryTime": 1700007200, "entryPrice": 300.0, "direction": "long"}
				]
			}
		}`
		tmpPath := filepath.Join(t.TempDir(), "golden.json")
		os.WriteFile(tmpPath, []byte(data), 0644)

		trades, err := LoadRunnerTrades(tmpPath)
		if err != nil {
			t.Fatalf("LoadRunnerTrades: %v", err)
		}
		if len(trades) != 3 {
			t.Fatalf("got %d trades, want 3 (2 closed + 1 open)", len(trades))
		}
		wantTimes := []int64{1700000000, 1700003600, 1700007200}
		wantDirs := []string{"long", "short", "long"}
		for i, wantT := range wantTimes {
			if trades[i].EntryUTC.Unix() != wantT {
				t.Errorf("[%d] EntryUTC.Unix() = %d, want %d", i, trades[i].EntryUTC.Unix(), wantT)
			}
			if trades[i].Direction != wantDirs[i] {
				t.Errorf("[%d] direction = %q, want %q", i, trades[i].Direction, wantDirs[i])
			}
		}
	})

	t.Run("closed_only", func(t *testing.T) {
		data := `{"result": {"trades": [{"entryTime": 1700000000, "entryPrice": 50.0, "direction": "long"}], "openTrades": []}}`
		tmpPath := filepath.Join(t.TempDir(), "closed_only.json")
		os.WriteFile(tmpPath, []byte(data), 0644)

		trades, err := LoadRunnerTrades(tmpPath)
		if err != nil {
			t.Fatalf("LoadRunnerTrades: %v", err)
		}
		if len(trades) != 1 {
			t.Fatalf("got %d trades, want 1", len(trades))
		}
	})

	t.Run("empty_result_returns_empty_slice", func(t *testing.T) {
		data := `{"result": {"trades": [], "openTrades": []}}`
		tmpPath := filepath.Join(t.TempDir(), "empty.json")
		os.WriteFile(tmpPath, []byte(data), 0644)

		trades, err := LoadRunnerTrades(tmpPath)
		if err != nil {
			t.Fatalf("LoadRunnerTrades: %v", err)
		}
		if len(trades) != 0 {
			t.Fatalf("got %d trades, want 0", len(trades))
		}
	})

	t.Run("entry_time_converted_to_utc", func(t *testing.T) {
		data := `{"result": {"trades": [{"entryTime": 0, "entryPrice": 1.0, "direction": "long"}], "openTrades": []}}`
		tmpPath := filepath.Join(t.TempDir(), "epoch.json")
		os.WriteFile(tmpPath, []byte(data), 0644)

		trades, err := LoadRunnerTrades(tmpPath)
		if err != nil {
			t.Fatalf("LoadRunnerTrades: %v", err)
		}
		if trades[0].EntryUTC.Location() != time.UTC {
			t.Errorf("EntryUTC location = %v, want UTC", trades[0].EntryUTC.Location())
		}
	})

	t.Run("missing_file_returns_error", func(t *testing.T) {
		_, err := LoadRunnerTrades(filepath.Join(t.TempDir(), "nonexistent.json"))
		if err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
	})

	t.Run("invalid_json_returns_error", func(t *testing.T) {
		tmpPath := filepath.Join(t.TempDir(), "bad.json")
		os.WriteFile(tmpPath, []byte("{not json"), 0644)
		_, err := LoadRunnerTrades(tmpPath)
		if err == nil {
			t.Fatal("expected error for invalid JSON, got nil")
		}
	})
}

func TestFixtureWindow(t *testing.T) {
	write := func(t *testing.T, content string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "fixture.json")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
		return path
	}

	t.Run("millisecond_timestamps_divided_by_1000", func(t *testing.T) {
		// > 1e10 → treated as milliseconds
		path := write(t, `{"bars": [{"time": 1700000000000}, {"time": 1700003600000}]}`)
		start, end, err := FixtureWindow(path)
		if err != nil {
			t.Fatalf("FixtureWindow: %v", err)
		}
		if start.Unix() != 1700000000 {
			t.Errorf("start.Unix() = %d, want 1700000000", start.Unix())
		}
		if end.Unix() != 1700003600 {
			t.Errorf("end.Unix() = %d, want 1700003600", end.Unix())
		}
	})

	t.Run("second_timestamps_used_as_is", func(t *testing.T) {
		// ≤ 1e10 → treated as seconds
		path := write(t, `{"bars": [{"time": 1700000000}, {"time": 1700003600}]}`)
		start, end, err := FixtureWindow(path)
		if err != nil {
			t.Fatalf("FixtureWindow: %v", err)
		}
		if start.Unix() != 1700000000 {
			t.Errorf("start.Unix() = %d, want 1700000000", start.Unix())
		}
		if end.Unix() != 1700003600 {
			t.Errorf("end.Unix() = %d, want 1700003600", end.Unix())
		}
	})

	t.Run("single_bar_start_equals_end", func(t *testing.T) {
		path := write(t, `{"bars": [{"time": 1700000000}]}`)
		start, end, err := FixtureWindow(path)
		if err != nil {
			t.Fatalf("FixtureWindow: %v", err)
		}
		if !start.Equal(end) {
			t.Errorf("single bar: start %v != end %v", start, end)
		}
	})

	t.Run("empty_bars_returns_error", func(t *testing.T) {
		path := write(t, `{"bars": []}`)
		_, _, err := FixtureWindow(path)
		if err == nil {
			t.Fatal("expected error for empty bars array, got nil")
		}
	})

	t.Run("timestamps_returned_in_utc", func(t *testing.T) {
		path := write(t, `{"bars": [{"time": 0}, {"time": 3600}]}`)
		start, end, err := FixtureWindow(path)
		if err != nil {
			t.Fatalf("FixtureWindow: %v", err)
		}
		if start.Location() != time.UTC {
			t.Errorf("start location = %v, want UTC", start.Location())
		}
		if end.Location() != time.UTC {
			t.Errorf("end location = %v, want UTC", end.Location())
		}
	})

	t.Run("missing_file_returns_error", func(t *testing.T) {
		_, _, err := FixtureWindow(filepath.Join(t.TempDir(), "nonexistent.json"))
		if err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
	})
}

func TestIsEntry_IsExit(t *testing.T) {
	// TV Strategy Tester CSV uses capitalized "Entry"/"Exit"/"Close" — matching is case-sensitive.
	cases := []struct {
		s       string
		isEntry bool
		isExit  bool
	}{
		{"Entry Long", true, false},
		{"Entry Short", true, false},
		{"Buy Entry", true, false},
		{"Exit Long", false, true},
		{"Exit Short", false, true},
		{"Close Long", false, true},
		{"Close position", false, true},
		{"Signal", false, false},
		{"Note", false, false},
		{"", false, false},
	}

	for _, tc := range cases {
		if got := isEntry(tc.s); got != tc.isEntry {
			t.Errorf("isEntry(%q) = %v, want %v", tc.s, got, tc.isEntry)
		}
		if got := isExit(tc.s); got != tc.isExit {
			t.Errorf("isExit(%q) = %v, want %v", tc.s, got, tc.isExit)
		}
	}
}

func TestDirectionFromType(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"Entry Long", "long"},
		{"Exit Long", "long"},
		{"entry long", "long"},
		{"Entry Short", "short"},
		{"Exit Short", "short"},
		{"unknown type uses fallback direction", "short"},
		{"", "short"},
	}
	for _, tc := range cases {
		if got := directionFromType(tc.input); got != tc.want {
			t.Errorf("directionFromType(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestHasSizeData(t *testing.T) {
	cases := []struct {
		name   string
		trades []TVTrade
		want   bool
	}{
		{"nil_slice", nil, false},
		{"empty_slice", []TVTrade{}, false},
		{"all_zero_sizes", []TVTrade{{Size: 0}, {Size: 0}, {Size: 0}}, false},
		{"single_nonzero", []TVTrade{{Size: 1}}, true},
		{"first_nonzero_rest_zero", []TVTrade{{Size: 5}, {Size: 0}, {Size: 0}}, true},
		{"last_nonzero", []TVTrade{{Size: 0}, {Size: 0}, {Size: 0.001}}, true},
		{"all_nonzero", []TVTrade{{Size: 10}, {Size: 20.5}, {Size: 387}}, true},
		{"fractional_crypto_size", []TVTrade{{Size: 0.00000001}}, true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := HasSizeData(tc.trades)
			if got != tc.want {
				t.Errorf("HasSizeData = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMatchSize_EdgeCases(t *testing.T) {
	timeTol := time.Hour
	priceTol := 1.0
	relTol := 0.01

	t0 := parseUTC("2025-01-01 10:00")
	t1 := parseUTC("2025-01-02 10:00")
	t2 := parseUTC("2025-01-03 10:00")

	cases := []struct {
		name            string
		runner          []RunnerTrade
		tv              []TVTrade
		wantMatched     int
		wantMismatched  int
		wantMaxResidual float64
	}{
		{
			name:        "nil_both",
			wantMatched: 0, wantMismatched: 0, wantMaxResidual: 0,
		},
		{
			name:        "nil_runner",
			tv:          []TVTrade{{EntryUTC: t0, EntryPrice: 100, Size: 10}},
			wantMatched: 0, wantMismatched: 0, wantMaxResidual: 0,
		},
		{
			name:        "nil_tv",
			runner:      []RunnerTrade{{EntryUTC: t0, EntryPrice: 100, Size: 10}},
			wantMatched: 0, wantMismatched: 0, wantMaxResidual: 0,
		},
		{
			name: "all_sizes_exact",
			runner: []RunnerTrade{
				{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 5.0},
				{EntryUTC: t1, EntryPrice: 200, Direction: "long", Size: 10.0},
			},
			tv: []TVTrade{
				{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 5.0},
				{EntryUTC: t1, EntryPrice: 200, Direction: "long", Size: 10.0},
			},
			wantMatched: 2, wantMismatched: 0, wantMaxResidual: 0,
		},
		{
			name:        "no_match_time_outside_tolerance",
			runner:      []RunnerTrade{{EntryUTC: t0, EntryPrice: 100, Size: 5}},
			tv:          []TVTrade{{EntryUTC: t2, EntryPrice: 100, Size: 5}},
			wantMatched: 0, wantMismatched: 0, wantMaxResidual: 0,
		},
		{
			name:        "no_match_price_outside_tolerance",
			runner:      []RunnerTrade{{EntryUTC: t0, EntryPrice: 100, Size: 5}},
			tv:          []TVTrade{{EntryUTC: t0, EntryPrice: 200, Size: 5}},
			wantMatched: 0, wantMismatched: 0, wantMaxResidual: 0,
		},
		{
			name: "multiple_mismatches_max_residual_is_largest",
			runner: []RunnerTrade{
				{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 110}, // residual 10/100=0.10
				{EntryUTC: t1, EntryPrice: 200, Direction: "long", Size: 105}, // residual  5/100=0.05
			},
			tv: []TVTrade{
				{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 100},
				{EntryUTC: t1, EntryPrice: 200, Direction: "long", Size: 100},
			},
			wantMatched: 2, wantMismatched: 2, wantMaxResidual: 0.10,
		},
		{
			name: "zero_runner_size_with_tv_size_is_mismatch",
			runner: []RunnerTrade{
				{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 0},
			},
			tv: []TVTrade{
				{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 5},
			},
			wantMatched: 1, wantMismatched: 1, wantMaxResidual: 1.0,
		},
		{
			name: "zero_tv_size_with_runner_size_uses_denominator_floor",
			runner: []RunnerTrade{
				{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 0.5},
			},
			tv: []TVTrade{
				{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 0},
			},
			wantMatched: 1, wantMismatched: 1, wantMaxResidual: 0.5,
		},
		{
			name: "tv_trade_not_reused_for_second_runner_match",
			runner: []RunnerTrade{
				{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 5},
				{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 5},
			},
			tv: []TVTrade{
				{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 5},
			},
			wantMatched: 1, wantMismatched: 0, wantMaxResidual: 0,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			matched, mismatched, maxResidual := MatchSize(tc.runner, tc.tv, timeTol, priceTol, relTol)
			if matched != tc.wantMatched {
				t.Errorf("matched = %d, want %d", matched, tc.wantMatched)
			}
			if mismatched != tc.wantMismatched {
				t.Errorf("mismatched = %d, want %d", mismatched, tc.wantMismatched)
			}
			if math.Abs(maxResidual-tc.wantMaxResidual) > 1e-10 {
				t.Errorf("maxResidual = %.10f, want %.10f", maxResidual, tc.wantMaxResidual)
			}
		})
	}
}

func TestZeroSizeRunnerMatches(t *testing.T) {
	timeTol := time.Hour
	priceTol := 1.0

	t0 := parseUTC("2025-01-01 10:00")
	t1 := parseUTC("2025-01-02 10:00")
	t2 := parseUTC("2025-01-03 10:00")
	tFar := parseUTC("2025-02-01 10:00")

	cases := []struct {
		name   string
		runner []RunnerTrade
		tv     []TVTrade
		want   int
	}{
		{
			name: "nil_both",
			want: 0,
		},
		{
			name: "nil_runner",
			tv:   []TVTrade{{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 5}},
			want: 0,
		},
		{
			name:   "nil_tv",
			runner: []RunnerTrade{{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 0}},
			want:   0,
		},
		{
			name:   "empty_both",
			runner: []RunnerTrade{},
			tv:     []TVTrade{},
			want:   0,
		},
		{
			name:   "all_positive_sizes_no_zero",
			runner: []RunnerTrade{{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 5}},
			tv:     []TVTrade{{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 5}},
			want:   0,
		},
		{
			name:   "single_zero_size_runner_matched",
			runner: []RunnerTrade{{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 0}},
			tv:     []TVTrade{{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 5}},
			want:   1,
		},
		{
			name:   "zero_size_runner_outside_time_tolerance_not_counted",
			runner: []RunnerTrade{{EntryUTC: tFar, EntryPrice: 100, Direction: "long", Size: 0}},
			tv:     []TVTrade{{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 5}},
			want:   0,
		},
		{
			name:   "zero_size_runner_outside_price_tolerance_not_counted",
			runner: []RunnerTrade{{EntryUTC: t0, EntryPrice: 999, Direction: "long", Size: 0}},
			tv:     []TVTrade{{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 5}},
			want:   0,
		},
		{
			name:   "zero_size_runner_direction_mismatch_not_counted",
			runner: []RunnerTrade{{EntryUTC: t0, EntryPrice: 100, Direction: "short", Size: 0}},
			tv:     []TVTrade{{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 5}},
			want:   0,
		},
		{
			name: "multiple_zero_size_runners_all_matched",
			runner: []RunnerTrade{
				{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 0},
				{EntryUTC: t1, EntryPrice: 200, Direction: "long", Size: 0},
			},
			tv: []TVTrade{
				{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 5},
				{EntryUTC: t1, EntryPrice: 200, Direction: "long", Size: 5},
			},
			want: 2,
		},
		{
			name: "mixed_zero_and_positive_runner_sizes_counts_only_zeros",
			runner: []RunnerTrade{
				{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 0},
				{EntryUTC: t1, EntryPrice: 200, Direction: "long", Size: 5},
				{EntryUTC: t2, EntryPrice: 300, Direction: "long", Size: 0},
			},
			tv: []TVTrade{
				{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 5},
				{EntryUTC: t1, EntryPrice: 200, Direction: "long", Size: 5},
				{EntryUTC: t2, EntryPrice: 300, Direction: "long", Size: 5},
			},
			want: 2,
		},
		{
			name:   "open_runner_with_zero_size_excluded",
			runner: []RunnerTrade{{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 0, IsOpen: true}},
			tv:     []TVTrade{{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 5}},
			want:   0,
		},
		{
			// Distinguishes from MatchSize: SizeResidual(0,0)==0 would not flag this,
			// but a zero runner size is a defect regardless of the TV size.
			name:   "tv_size_also_zero_runner_zero_still_detected",
			runner: []RunnerTrade{{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 0}},
			tv:     []TVTrade{{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 0}},
			want:   1,
		},
		{
			name: "single_tv_slot_not_reused_for_two_zero_size_runners",
			runner: []RunnerTrade{
				{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 0},
				{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 0},
			},
			tv:   []TVTrade{{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 5}},
			want: 1,
		},
		{
			name:   "fractional_positive_size_not_treated_as_zero",
			runner: []RunnerTrade{{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 0.00000001}},
			tv:     []TVTrade{{EntryUTC: t0, EntryPrice: 100, Direction: "long", Size: 5}},
			want:   0,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := ZeroSizeRunnerMatches(tc.runner, tc.tv, timeTol, priceTol)
			if got != tc.want {
				t.Errorf("ZeroSizeRunnerMatches = %d, want %d", got, tc.want)
			}
		})
	}
}

// TestZeroSizeRunnerMatches_InputOrderInvariant verifies that rearranging the
// runner or TV slices does not change the count, since the underlying matching
// uses maximum bipartite matching (not greedy, not position-dependent).
func TestZeroSizeRunnerMatches_InputOrderInvariant(t *testing.T) {
	timeTol := time.Hour
	priceTol := 1.0

	runner := []RunnerTrade{
		{EntryUTC: parseUTC("2025-01-03 10:00"), EntryPrice: 300, Direction: "long", Size: 0},
		{EntryUTC: parseUTC("2025-01-01 10:00"), EntryPrice: 100, Direction: "long", Size: 0},
		{EntryUTC: parseUTC("2025-01-02 10:00"), EntryPrice: 200, Direction: "long", Size: 5},
	}
	tv := []TVTrade{
		{EntryUTC: parseUTC("2025-01-02 10:00"), EntryPrice: 200, Direction: "long", Size: 5},
		{EntryUTC: parseUTC("2025-01-03 10:00"), EntryPrice: 300, Direction: "long", Size: 5},
		{EntryUTC: parseUTC("2025-01-01 10:00"), EntryPrice: 100, Direction: "long", Size: 5},
	}

	want := ZeroSizeRunnerMatches(runner, tv, timeTol, priceTol)

	revRunner := append([]RunnerTrade(nil), runner...)
	for i, j := 0, len(revRunner)-1; i < j; i, j = i+1, j-1 {
		revRunner[i], revRunner[j] = revRunner[j], revRunner[i]
	}
	revTV := append([]TVTrade(nil), tv...)
	for i, j := 0, len(revTV)-1; i < j; i, j = i+1, j-1 {
		revTV[i], revTV[j] = revTV[j], revTV[i]
	}

	for _, tc := range []struct {
		name   string
		runner []RunnerTrade
		tv     []TVTrade
	}{
		{"runner_reversed", revRunner, tv},
		{"tv_reversed", runner, revTV},
		{"both_reversed", revRunner, revTV},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := ZeroSizeRunnerMatches(tc.runner, tc.tv, timeTol, priceTol)
			if got != want {
				t.Errorf("ZeroSizeRunnerMatches = %d, want %d", got, want)
			}
		})
	}
}

func TestParseTrades_AbsentSizeColumn_TradesHaveZeroSize(t *testing.T) {
	rawCSV := "Trade number,Type,Date and time,Price\n" +
		"1,Entry Long,2025-01-01 10:00,100\n" +
		"1,Exit Long,2025-01-01 14:00,110\n"
	tmpPath := filepath.Join(t.TempDir(), "no_size.csv")
	if err := os.WriteFile(tmpPath, []byte(rawCSV), 0644); err != nil {
		t.Fatalf("write temp CSV: %v", err)
	}
	trades, err := LoadTrades(tmpPath, TVTimezoneUTC)
	if err != nil {
		t.Fatalf("LoadTrades: %v", err)
	}
	if len(trades) != 1 {
		t.Fatalf("got %d trades, want 1", len(trades))
	}
	if trades[0].Size != 0 {
		t.Errorf("Size = %v, want 0 when size column is absent", trades[0].Size)
	}
	if HasSizeData(trades) {
		t.Errorf("HasSizeData returned true when CSV had no size column — would produce a false size assertion")
	}
}

func TestTVEquityAtWindowStart(t *testing.T) {
	const capital = 10000.0
	windowStart := parseUTC("2025-01-05 00:00")

	makeTrade := func(datetime string, pnl float64) TVTrade {
		return TVTrade{EntryUTC: parseUTC(datetime), NetPnL: pnl}
	}

	cases := []struct {
		name      string
		allTrades []TVTrade
		want      float64
	}{
		{
			"nil_trades_returns_initial_capital",
			nil,
			capital,
		},
		{
			"empty_slice_returns_initial_capital",
			[]TVTrade{},
			capital,
		},
		{
			"trade_exactly_at_window_start_excluded",
			[]TVTrade{makeTrade("2025-01-05 00:00", 500)},
			capital,
		},
		{
			"all_trades_after_window_start_ignored",
			[]TVTrade{makeTrade("2025-01-06 00:00", 500), makeTrade("2025-01-07 00:00", 200)},
			capital,
		},
		{
			"single_pre_window_profitable_trade",
			[]TVTrade{makeTrade("2025-01-04 23:59", 500)},
			capital + 500,
		},
		{
			"single_pre_window_losing_trade",
			[]TVTrade{makeTrade("2025-01-04 23:59", -200)},
			capital - 200,
		},
		{
			"multiple_pre_window_trades_accumulated",
			[]TVTrade{makeTrade("2025-01-03 00:00", 100), makeTrade("2025-01-04 00:00", 200)},
			capital + 300,
		},
		{
			"mixed_pre_and_post_window_only_pre_counted",
			[]TVTrade{makeTrade("2025-01-04 00:00", 100), makeTrade("2025-01-06 00:00", 999)},
			capital + 100,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := TVEquityAtWindowStart(tc.allTrades, windowStart, capital)
			if got != tc.want {
				t.Errorf("TVEquityAtWindowStart = %.2f, want %.2f", got, tc.want)
			}
		})
	}
}

func TestMatchNetPnL(t *testing.T) {
	t0 := parseUTC("2025-01-01 10:00")
	t1 := parseUTC("2025-01-02 10:00")
	timeTol := time.Hour
	priceTol := 1.0
	relTol := 0.01

	makeRunner := func(dt time.Time, price, pnl float64) RunnerTrade {
		return RunnerTrade{EntryUTC: dt, EntryPrice: price, Direction: "long", NetPnL: pnl}
	}
	makeTV := func(dt time.Time, price, pnl float64) TVTrade {
		return TVTrade{EntryUTC: dt, EntryPrice: price, Direction: "long", NetPnL: pnl}
	}

	t.Run("empty_runner_and_tv", func(t *testing.T) {
		matched, mismatch := MatchNetPnL(nil, nil, 1.0, timeTol, priceTol, relTol)
		if matched != 0 || mismatch != 0 {
			t.Errorf("got (%d, %d), want (0, 0)", matched, mismatch)
		}
	})

	t.Run("empty_runner_with_tv_trades", func(t *testing.T) {
		matched, mismatch := MatchNetPnL(nil, []TVTrade{makeTV(t0, 100, 50)}, 1.0, timeTol, priceTol, relTol)
		if matched != 0 || mismatch != 0 {
			t.Errorf("got (%d, %d), want (0, 0)", matched, mismatch)
		}
	})

	t.Run("exact_pnl_equity_ratio_one", func(t *testing.T) {
		matched, mismatch := MatchNetPnL(
			[]RunnerTrade{makeRunner(t0, 100, 50)},
			[]TVTrade{makeTV(t0, 100, 50)},
			1.0, timeTol, priceTol, relTol,
		)
		if matched != 1 || mismatch != 0 {
			t.Errorf("got (%d, %d), want (1, 0)", matched, mismatch)
		}
	})

	t.Run("pnl_exceeds_relative_tolerance", func(t *testing.T) {
		matched, mismatch := MatchNetPnL(
			[]RunnerTrade{makeRunner(t0, 100, 60)},
			[]TVTrade{makeTV(t0, 100, 50)},
			1.0, timeTol, priceTol, relTol,
		)
		if matched != 1 || mismatch != 1 {
			t.Errorf("got (%d, %d), want (1, 1)", matched, mismatch)
		}
	})

	t.Run("equity_ratio_scales_runner_pnl_before_comparison", func(t *testing.T) {
		// runner made 100 profit; TV equity at window start was half of runner initial capital,
		// so runner position sizes are 2x TV — equityRatio=0.5 normalises to TV basis → scaled=50
		matched, mismatch := MatchNetPnL(
			[]RunnerTrade{makeRunner(t0, 100, 100)},
			[]TVTrade{makeTV(t0, 100, 50)},
			0.5, timeTol, priceTol, relTol,
		)
		if matched != 1 || mismatch != 0 {
			t.Errorf("got (%d, %d), want (1, 0) — scaled runner PnL should match TV PnL", matched, mismatch)
		}
	})

	t.Run("runner_with_no_tv_match_not_counted_as_matched", func(t *testing.T) {
		matched, mismatch := MatchNetPnL(
			[]RunnerTrade{makeRunner(t0, 100, 50)},
			[]TVTrade{makeTV(t1, 100, 50)},
			1.0, time.Minute, priceTol, relTol,
		)
		if matched != 0 || mismatch != 0 {
			t.Errorf("got (%d, %d), want (0, 0) — unmatched runner trade silently skipped", matched, mismatch)
		}
	})

	t.Run("tv_trade_not_reused_across_runner_matches", func(t *testing.T) {
		matched, mismatch := MatchNetPnL(
			[]RunnerTrade{makeRunner(t0, 100, 50), makeRunner(t0, 100, 50)},
			[]TVTrade{makeTV(t0, 100, 50)},
			1.0, timeTol, priceTol, relTol,
		)
		if matched != 1 || mismatch != 0 {
			t.Errorf("got (%d, %d), want (1, 0) — TV trade must not be consumed twice", matched, mismatch)
		}
	})

	t.Run("multiple_trades_mixed_pnl_match", func(t *testing.T) {
		matched, mismatch := MatchNetPnL(
			[]RunnerTrade{makeRunner(t0, 100, 50), makeRunner(t1, 200, 60)},
			[]TVTrade{makeTV(t0, 100, 50), makeTV(t1, 200, 40)},
			1.0, timeTol, priceTol, relTol,
		)
		if matched != 2 || mismatch != 1 {
			t.Errorf("got (%d, %d), want (2, 1)", matched, mismatch)
		}
	})
}

func TestPnlMatchRelative(t *testing.T) {
	cases := []struct {
		name   string
		a, b   float64
		relTol float64
		want   bool
	}{
		{"equal_values", 100, 100, 0.01, true},
		{"both_zero", 0, 0, 0.01, true},
		{"within_relative_tolerance", 100, 101, 0.02, true},
		{"exceeds_relative_tolerance", 100, 103, 0.02, false},
		{"at_exact_boundary_inclusive", 100, 120, 0.20, true},
		{"larger_magnitude_used_as_denominator", 50, 100, 0.5, true},
		{"negative_values_within_tolerance", -100, -99, 0.02, true},
		{"runner_zero_tv_nonzero_fails", 0, 10, 0.50, false},
		{"opposite_signs_large_deviation", 10, -10, 0.99, false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := pnlMatchRelative(tc.a, tc.b, tc.relTol)
			if got != tc.want {
				t.Errorf("pnlMatchRelative(%.0f, %.0f, %.2f) = %v, want %v", tc.a, tc.b, tc.relTol, got, tc.want)
			}
		})
	}
}

// TestLoadRunnerTrades_IsOpenFlag verifies that LoadRunnerTrades marks closed
// and open trades correctly via the IsOpen field introduced to RunnerTrade.
// This is separate from TestLoadRunnerTrades, which checks count and field values
// but not the IsOpen flag.
func TestLoadRunnerTrades_IsOpenFlag(t *testing.T) {
	data := `{
		"result": {
			"trades": [
				{"entryTime": 1700000000, "entryPrice": 100.0, "direction": "long",  "size": 1.0},
				{"entryTime": 1700003600, "entryPrice": 200.0, "direction": "short", "size": 2.0}
			],
			"openTrades": [
				{"entryTime": 1700007200, "entryPrice": 300.0, "direction": "long",  "size": 3.0}
			]
		}
	}`
	tmpPath := filepath.Join(t.TempDir(), "golden.json")
	if err := os.WriteFile(tmpPath, []byte(data), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	trades, err := LoadRunnerTrades(tmpPath)
	if err != nil {
		t.Fatalf("LoadRunnerTrades: %v", err)
	}
	if len(trades) != 3 {
		t.Fatalf("got %d trades, want 3", len(trades))
	}

	t.Run("first_closed_trade_is_not_open", func(t *testing.T) {
		if trades[0].IsOpen {
			t.Errorf("trades[0].IsOpen = true, want false (closed trade)")
		}
	})
	t.Run("second_closed_trade_is_not_open", func(t *testing.T) {
		if trades[1].IsOpen {
			t.Errorf("trades[1].IsOpen = true, want false (closed trade)")
		}
	})
	t.Run("open_trade_is_marked_open", func(t *testing.T) {
		if !trades[2].IsOpen {
			t.Errorf("trades[2].IsOpen = false, want true (open trade)")
		}
	})

	t.Run("all_closed_when_no_open_trades_key", func(t *testing.T) {
		d := `{"result": {"trades": [{"entryTime": 1, "entryPrice": 1.0, "direction": "long"}], "openTrades": []}}`
		p := filepath.Join(t.TempDir(), "closed_only.json")
		os.WriteFile(p, []byte(d), 0644)
		got, _ := LoadRunnerTrades(p)
		for i, tr := range got {
			if tr.IsOpen {
				t.Errorf("trades[%d].IsOpen = true, want false (no open trades in result)", i)
			}
		}
	})

	t.Run("open_trade_field_values_preserved", func(t *testing.T) {
		if trades[2].EntryPrice != 300.0 {
			t.Errorf("open trade EntryPrice = %.1f, want 300.0", trades[2].EntryPrice)
		}
		if trades[2].Size != 3.0 {
			t.Errorf("open trade Size = %.1f, want 3.0", trades[2].Size)
		}
		if trades[2].Direction != "long" {
			t.Errorf("open trade Direction = %q, want long", trades[2].Direction)
		}
	})
}
