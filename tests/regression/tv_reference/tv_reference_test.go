package tv_reference

import (
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
			"partial_greedy_ordered",
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
			"greedy_advances_j_past_consumed_tv_trades",
			// runner[0] consumes tv[0]; runner[1] must look from tv[1], not tv[0]
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
	cases := []string{"not-a-date", "2025/01/15 13:00", "", "2025-01-15"}
	for _, tz := range []TVTimezone{TVTimezoneUTC, TVTimezoneMoscow, TVTimezoneNewYork} {
		for _, s := range cases {
			if _, err := parseDateTime(s, tz); err == nil {
				t.Errorf("tz=%d input=%q: expected error, got nil", tz, s)
			}
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
	aliases := []string{"Size", "Quantity", "Contracts", "Size (qty)", "qty"}
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
