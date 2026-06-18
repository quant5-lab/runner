package market

import (
	"testing"
	"time"

	"github.com/quant5-lab/runner/runtime/context"
)

// barAtLocal builds an OHLCV bar whose timestamp falls at the given clock time
// in the named IANA timezone.  The date is fixed (2025-10-01) and is irrelevant
// to every test in this file; only the clock minute matters.
func barAtLocal(t *testing.T, tz string, hour, minute int) context.OHLCV {
	t.Helper()
	loc, err := time.LoadLocation(tz)
	if err != nil {
		t.Fatalf("load timezone %q: %v", tz, err)
	}
	ts := time.Date(2025, 10, 1, hour, minute, 0, 0, loc).Unix()
	return context.OHLCV{Time: ts, Open: 100, High: 101, Low: 99, Close: 100}
}

// barsForWeek builds a slice of OHLCV bars covering days [startDay, endDay) in
// October 2025, placing three bars per day at openH:openM, openH+2:openM, and
// openH+4:openM local time to provide a realistic multi-day dataset without
// requiring a full fixture file.
func barsForWeek(t *testing.T, tz string, startDay, endDay int, openH, openM int) []context.OHLCV {
	t.Helper()
	loc, err := time.LoadLocation(tz)
	if err != nil {
		t.Fatalf("load timezone %q: %v", tz, err)
	}
	var bars []context.OHLCV
	for day := startDay; day < endDay; day++ {
		for _, hm := range [][2]int{{openH, openM}, {openH + 2, openM}, {openH + 4, openM}} {
			ts := time.Date(2025, 10, day, hm[0], hm[1], 0, 0, loc).Unix()
			bars = append(bars, context.OHLCV{Time: ts, Open: 100, High: 101, Low: 99, Close: 100})
		}
	}
	return bars
}

// TestSessionAnchorFor_AllKnownExchangeProfiles verifies that SessionAnchorFor
// correctly maps each known exchange profile to its canonical IANA timezone and
// session-open clock minute.  The zero session-open minute is the correct value
// for always-open and UTC-midnight-anchored exchanges.
func TestSessionAnchorFor_AllKnownExchangeProfiles(t *testing.T) {
	cases := []struct {
		name     string
		profile  Profile
		wantTZ   string
		wantOpen int
	}{
		{
			name:     "UTC_always_open",
			profile:  Profile{Timezone: "UTC", ReferenceSession: ReferenceSessionAlwaysOpen, SessionOpenMinute: 0},
			wantTZ:   "UTC",
			wantOpen: 0,
		},
		{
			name:     "empty_timezone_normalized_to_UTC",
			profile:  Profile{Timezone: "", SessionOpenMinute: 0},
			wantTZ:   "UTC",
			wantOpen: 0,
		},
		{
			name:     "MOEX_regular_session",
			profile:  ResolveProfile("SBERP", ""),
			wantTZ:   "Europe/Moscow",
			wantOpen: 7 * 60,
		},
		{
			name:     "NYSE_explicit_exchange_field",
			profile:  ResolveProfileWithMetadata("AAPL", SourceMetadata{Exchange: "NYSE"}),
			wantTZ:   "America/New_York",
			wantOpen: 9*60 + 30,
		},
		{
			name:     "NASDAQ_explicit_exchange_field",
			profile:  ResolveProfileWithMetadata("NVDA", SourceMetadata{Exchange: "NASDAQ"}),
			wantTZ:   "America/New_York",
			wantOpen: 9*60 + 30,
		},
		{
			name:     "Binance_always_open",
			profile:  ResolveProfile("BINANCE:BTCUSDT", ""),
			wantTZ:   "UTC",
			wantOpen: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			anchor := SessionAnchorFor(tc.profile)
			if anchor.Timezone != tc.wantTZ {
				t.Errorf("Timezone: got %q, want %q", anchor.Timezone, tc.wantTZ)
			}
			if anchor.SessionOpenMinute != tc.wantOpen {
				t.Errorf("SessionOpenMinute: got %d, want %d", anchor.SessionOpenMinute, tc.wantOpen)
			}
		})
	}
}

// TestProfile_RegularSessionHasNonZeroOpenMinute verifies that the
// SessionOpenMinute stored in a resolved Profile is non-zero for known
// regular-session exchanges and zero when the session is always-open.  This
// property drives the table-fallback path in DeriveSessionAnchor.
func TestProfile_RegularSessionHasNonZeroOpenMinute(t *testing.T) {
	cases := []struct {
		name       string
		regular    Profile
		alwaysOpen Profile
	}{
		{
			name:       "MOEX",
			regular:    ResolveProfileWithReferenceSession("SBERP", "Europe/Moscow", "regular"),
			alwaysOpen: ResolveProfileWithReferenceSession("SBERP", "Europe/Moscow", "always-open"),
		},
		{
			name:       "NYSE",
			regular:    ResolveProfileWithMetadata("AAPL", SourceMetadata{Exchange: "NYSE", ReferenceSession: "regular"}),
			alwaysOpen: ResolveProfileWithMetadata("AAPL", SourceMetadata{Exchange: "NYSE", ReferenceSession: "always-open"}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.regular.SessionOpenMinute == 0 {
				t.Errorf("regular %s profile must have non-zero SessionOpenMinute", tc.name)
			}
			if tc.alwaysOpen.SessionOpenMinute != 0 {
				t.Errorf("always-open %s profile must have SessionOpenMinute=0; got %d", tc.name, tc.alwaysOpen.SessionOpenMinute)
			}
		})
	}
}

// TestDeriveSessionAnchor_ObservationGating verifies which profile+bar
// combinations trigger per-bar observation vs. which return a fixed anchor:
//
//   - Known exchange + Regular session + bars → observed mode (table is ignored)
//   - Known exchange + AlwaysOpen session + bars → midnight (0), never observes
//   - Unknown exchange + any session type + bars → observed mode (unknown always observes)
//   - Any profile + no bars → table fallback, no observation
//
// The "observed prevails over table" cases use multi-day bar data so that the
// mode algorithm is exercised, not just a degenerate single-value frequency.
func TestDeriveSessionAnchor_ObservationGating(t *testing.T) {
	cases := []struct {
		name     string
		profile  Profile
		bars     []context.OHLCV
		wantOpen int
	}{
		{
			// Five MOEX trading days each opening at 09:30 MSK, not 07:00 MSK.
			// Intentional mismatch with the curated MOEX table entry (07:00=420)
			// to prove that the observed mode (09:30=570) beats the table.
			name:     "known_regular_MOEX_observed_mode_beats_table",
			profile:  ResolveProfile("SBERP", ""),
			bars:     barsForWeek(t, "Europe/Moscow", 1, 6, 9, 30),
			wantOpen: 9*60 + 30,
		},
		{
			// Five NYSE trading days each opening at 10:00 ET, not 09:30 ET.
			// Intentional mismatch with the curated NYSE table entry (09:30=570)
			// to prove that the observed mode (10:00=600) beats the table.
			name:     "known_regular_NYSE_observed_mode_beats_table",
			profile:  ResolveProfileWithMetadata("AAPL", SourceMetadata{Exchange: "NYSE"}),
			bars:     barsForWeek(t, "America/New_York", 1, 6, 10, 0),
			wantOpen: 10 * 60,
		},
		{
			// No bars supplied for a regular NYSE profile → table fallback.
			name:     "known_regular_no_bars_falls_back_to_table",
			profile:  ResolveProfileWithMetadata("AAPL", SourceMetadata{Exchange: "NYSE"}),
			bars:     nil,
			wantOpen: 9*60 + 30,
		},
		{
			// Binance is a known always-open exchange; observation must not run.
			name:     "known_always_open_Binance_bars_do_not_trigger_observation",
			profile:  ResolveProfile("BINANCE:BTCUSDT", ""),
			bars:     []context.OHLCV{barAtLocal(t, "UTC", 12, 0)},
			wantOpen: 0,
		},
		{
			// MOEX regular-session bars provided, but always-open override forces
			// midnight — observation must not run even when bars are present.
			name:     "known_always_open_override_MOEX_bars_do_not_trigger_observation",
			profile:  ResolveProfileWithReferenceSession("SBERP", "Europe/Moscow", "always-open"),
			bars:     barsForWeek(t, "Europe/Moscow", 1, 6, 7, 0),
			wantOpen: 0,
		},
		{
			// NYSE regular-session bars provided, but always-open override forces
			// midnight — observation must not run even when bars are present.
			name:     "known_always_open_override_NYSE_bars_do_not_trigger_observation",
			profile:  ResolveProfileWithMetadata("AAPL", SourceMetadata{Exchange: "NYSE", ReferenceSession: "always-open"}),
			bars:     barsForWeek(t, "America/New_York", 1, 6, 9, 30),
			wantOpen: 0,
		},
		{
			// Unknown exchange with the default ReferenceSession (AlwaysOpen) still
			// observes because unknown exchanges cannot be assumed to be 24/7.
			// This covers tickers like "AAPL" supplied without an exchange prefix,
			// which resolve to ExchangeUnknown with a default AlwaysOpen session.
			name:     "unknown_exchange_default_always_open_still_observes",
			profile:  ResolveProfileWithMetadata("AAPL", SourceMetadata{Timezone: "America/New_York"}),
			bars:     barsForWeek(t, "America/New_York", 1, 6, 9, 30),
			wantOpen: 9*60 + 30,
		},
		{
			// Same as above but with a UTC-anchored unknown ticker (e.g. an
			// unrecognised crypto asset).  Observation runs, mode = 0, which is
			// identical to the always-open midnight anchor — both are correct.
			name:     "unknown_exchange_UTC_24h_data_observes_to_midnight",
			profile:  ResolveProfileWithMetadata("MYTOKEN", SourceMetadata{Timezone: "UTC"}),
			bars:     barsForWeek(t, "UTC", 1, 6, 0, 0),
			wantOpen: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			anchor := DeriveSessionAnchor(tc.profile, tc.bars)
			if anchor.SessionOpenMinute != tc.wantOpen {
				t.Errorf("SessionOpenMinute: got %d, want %d", anchor.SessionOpenMinute, tc.wantOpen)
			}
		})
	}
}

// TestDeriveSessionAnchor_ObservationReflectsFirstBarPerDay verifies that the
// observation algorithm extracts the session-open minute from the first bar of
// each calendar day in the symbol's own timezone, not the global minimum or a
// fixed-offset approximation.
func TestDeriveSessionAnchor_ObservationReflectsFirstBarPerDay(t *testing.T) {
	cases := []struct {
		name     string
		tz       string
		barTimes [][2]int // (hour, minute) for a single trading day
		wantOpen int
	}{
		{
			name:     "half_hour_precision_preserved",
			tz:       "America/New_York",
			barTimes: [][2]int{{9, 30}, {10, 0}, {11, 0}, {12, 0}},
			wantOpen: 9*60 + 30,
		},
		{
			name:     "hour_boundary_open",
			tz:       "Europe/Moscow",
			barTimes: [][2]int{{7, 0}, {8, 0}, {9, 0}, {10, 0}},
			wantOpen: 7 * 60,
		},
		{
			name:     "UTC_midnight_open",
			tz:       "UTC",
			barTimes: [][2]int{{0, 0}, {1, 0}, {12, 0}, {23, 0}},
			wantOpen: 0,
		},
		{
			name:     "single_bar_day",
			tz:       "America/New_York",
			barTimes: [][2]int{{9, 30}},
			wantOpen: 9*60 + 30,
		},
		{
			name:     "intraday_bars_only_first_counts",
			tz:       "America/New_York",
			barTimes: [][2]int{{9, 30}, {10, 0}, {11, 0}, {11, 30}, {14, 0}, {15, 30}},
			wantOpen: 9*60 + 30,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Use ExchangeUnknown so the table yields 0 and observation is the
			// sole source of the answer.
			profile := ResolveProfileWithMetadata("UNKNOWN", SourceMetadata{Timezone: tc.tz})

			bars := make([]context.OHLCV, len(tc.barTimes))
			for i, hm := range tc.barTimes {
				bars[i] = barAtLocal(t, tc.tz, hm[0], hm[1])
			}

			anchor := DeriveSessionAnchor(profile, bars)

			if anchor.SessionOpenMinute != tc.wantOpen {
				t.Errorf("SessionOpenMinute: got %d, want %d", anchor.SessionOpenMinute, tc.wantOpen)
			}
			if anchor.Timezone != tc.tz {
				t.Errorf("Timezone: got %q, want %q", anchor.Timezone, tc.tz)
			}
		})
	}
}

// TestDeriveSessionAnchor_PerDayFirstBarModeConvergence verifies the end-to-end
// correctness of the mode-of-per-day-first-bar design: the derived anchor
// converges to the dominant session open even when some trading days contribute
// anomalous first-bar times (e.g. pre-market or extended-hours prints).
func TestDeriveSessionAnchor_PerDayFirstBarModeConvergence(t *testing.T) {
	const tz = "America/New_York"
	loc := func() *time.Location {
		l, err := time.LoadLocation(tz)
		if err != nil {
			panic(err)
		}
		return l
	}()

	makeDay := func(day, openH, openM int, extraHours ...int) []context.OHLCV {
		var bars []context.OHLCV
		ts := func(h, m int) int64 {
			return time.Date(2025, 10, day, h, m, 0, 0, loc).Unix()
		}
		bars = append(bars, context.OHLCV{Time: ts(openH, openM)})
		for _, h := range extraHours {
			bars = append(bars, context.OHLCV{Time: ts(h, openM)})
		}
		return bars
	}

	cases := []struct {
		name     string
		bars     []context.OHLCV
		wantOpen int
	}{
		{
			name: "consistent_five_day_dataset",
			bars: func() []context.OHLCV {
				var all []context.OHLCV
				for day := 1; day <= 5; day++ {
					all = append(all, makeDay(day, 9, 30, 10, 11, 13, 15)...)
				}
				return all
			}(),
			wantOpen: 9*60 + 30,
		},
		{
			name: "minority_pre_market_outlier_does_not_shift_anchor",
			bars: func() []context.OHLCV {
				// One anomaly day opens at 04:00; four regular days open at 09:30.
				// Mode must be 09:30, not 04:00.
				anomaly := makeDay(6, 4, 0, 10, 13)
				var regular []context.OHLCV
				for day := 7; day <= 10; day++ {
					regular = append(regular, makeDay(day, 9, 30, 10, 11, 13, 15)...)
				}
				return append(anomaly, regular...)
			}(),
			wantOpen: 9*60 + 30,
		},
		{
			name: "two_thirds_majority_beats_one_third_minority",
			bars: func() []context.OHLCV {
				// 6 days at 09:30, 3 days at 10:00 — 09:30 must win.
				var all []context.OHLCV
				for day := 1; day <= 6; day++ {
					all = append(all, makeDay(day, 9, 30, 10, 11, 13)...)
				}
				for day := 7; day <= 9; day++ {
					all = append(all, makeDay(day, 10, 0, 11, 12, 14)...)
				}
				return all
			}(),
			wantOpen: 9*60 + 30,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// ExchangeUnknown keeps the table at 0 so the answer comes purely from
			// observation, not from a coincidental table match.
			profile := ResolveProfileWithMetadata("AAPL", SourceMetadata{Timezone: tz})

			anchor := DeriveSessionAnchor(profile, tc.bars)
			if anchor.SessionOpenMinute != tc.wantOpen {
				t.Errorf("SessionOpenMinute = %d, want %d", anchor.SessionOpenMinute, tc.wantOpen)
			}
		})
	}
}

// TestDeriveSessionAnchor_NoBarsFallbackBehaviour verifies that nil and empty
// bar slices are treated identically as "no data available" and return the
// curated table entry for regular-session profiles or midnight for always-open.
func TestDeriveSessionAnchor_NoBarsFallbackBehaviour(t *testing.T) {
	cases := []struct {
		name     string
		profile  Profile
		wantOpen int
		wantTZ   string
	}{
		{
			name:     "regular_NYSE_returns_table_entry",
			profile:  ResolveProfileWithMetadata("AAPL", SourceMetadata{Exchange: "NYSE"}),
			wantOpen: 9*60 + 30,
			wantTZ:   "America/New_York",
		},
		{
			name:     "regular_MOEX_returns_table_entry",
			profile:  ResolveProfile("SBERP", ""),
			wantOpen: 7 * 60,
			wantTZ:   "Europe/Moscow",
		},
		{
			name:     "always_open_Binance_returns_midnight",
			profile:  ResolveProfile("BINANCE:BTCUSDT", ""),
			wantOpen: 0,
			wantTZ:   "UTC",
		},
		{
			name:     "unknown_exchange_returns_zero",
			profile:  ResolveProfileWithMetadata("UNKNOWN", SourceMetadata{Timezone: "America/New_York"}),
			wantOpen: 0,
			wantTZ:   "America/New_York",
		},
	}

	for _, tc := range cases {
		for _, bars := range [][]context.OHLCV{nil, {}} {
			label := "nil_bars"
			if bars != nil {
				label = "empty_bars"
			}
			t.Run(tc.name+"/"+label, func(t *testing.T) {
				anchor := DeriveSessionAnchor(tc.profile, bars)
				if anchor.SessionOpenMinute != tc.wantOpen {
					t.Errorf("SessionOpenMinute: got %d, want %d", anchor.SessionOpenMinute, tc.wantOpen)
				}
				if anchor.Timezone != tc.wantTZ {
					t.Errorf("Timezone: got %q, want %q", anchor.Timezone, tc.wantTZ)
				}
			})
		}
	}
}

// TestDeriveSessionAnchor_TimezoneAlwaysFromProfile verifies that the timezone
// in the returned anchor always comes from the profile regardless of observation
// path, bars presence, or exchange type.  This holds for: the observation path
// on known exchanges, the observation path on unknown exchanges, and the
// non-observation (always-open) path.
func TestDeriveSessionAnchor_TimezoneAlwaysFromProfile(t *testing.T) {
	// Precompute bars before the cases table so barAtLocal receives a valid *t.
	utcTimestampBar := []context.OHLCV{{Time: time.Date(2025, 10, 1, 10, 0, 0, 0, time.UTC).Unix()}}
	nyBars := []context.OHLCV{barAtLocal(t, "America/New_York", 9, 30)}
	utcBars := []context.OHLCV{barAtLocal(t, "UTC", 12, 0)}

	cases := []struct {
		name    string
		profile Profile
		bars    []context.OHLCV
		wantTZ  string
	}{
		{
			// Bars are timestamped in UTC while the profile carries Europe/Moscow;
			// timezone must come from the profile, never inferred from bar times.
			name:    "unknown_exchange_observation_path",
			profile: ResolveProfileWithMetadata("UNKNOWN", SourceMetadata{Timezone: "Europe/Moscow"}),
			bars:    utcTimestampBar,
			wantTZ:  "Europe/Moscow",
		},
		{
			name:    "known_regular_NYSE_observation_path",
			profile: ResolveProfileWithMetadata("AAPL", SourceMetadata{Exchange: "NYSE"}),
			bars:    nyBars,
			wantTZ:  "America/New_York",
		},
		{
			name:    "known_always_open_Binance_no_observation_path",
			profile: ResolveProfile("BINANCE:BTCUSDT", ""),
			bars:    utcBars,
			wantTZ:  "UTC",
		},
		{
			name:    "no_bars_table_fallback_path",
			profile: ResolveProfile("SBERP", ""),
			bars:    nil,
			wantTZ:  "Europe/Moscow",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			anchor := DeriveSessionAnchor(tc.profile, tc.bars)
			if anchor.Timezone != tc.wantTZ {
				t.Errorf("Timezone: got %q, want %q", anchor.Timezone, tc.wantTZ)
			}
		})
	}
}

// TestModeMinute_AlgorithmProperties covers the full behavioural contract of
// modeMinute as a frequency-selection function: it returns the most common
// element, resolves ties toward the smaller value, and handles edge-case input
// sizes.  Cases are expressed in terms of algorithm properties so the tests
// remain valid if modeMinute is reused in non-market contexts.
func TestModeMinute_AlgorithmProperties(t *testing.T) {
	cases := []struct {
		name     string
		input    []int
		wantMode int
	}{
		{
			name:     "empty_input_returns_zero",
			input:    nil,
			wantMode: 0,
		},
		{
			name:     "single_element_returns_that_element",
			input:    []int{570},
			wantMode: 570,
		},
		{
			name:     "all_same_returns_that_value",
			input:    []int{420, 420, 420, 420},
			wantMode: 420,
		},
		{
			name:     "clear_majority_wins",
			input:    []int{570, 570, 570, 570, 240},
			wantMode: 570,
		},
		{
			name:     "large_minority_does_not_override_majority",
			input:    []int{240, 240, 570, 570, 570, 570, 570, 570, 570, 570},
			wantMode: 570,
		},
		{
			name:     "tie_resolved_by_smaller_value",
			input:    []int{600, 570},
			wantMode: 570,
		},
		{
			name:     "tie_with_zero_returns_zero",
			input:    []int{0, 60, 120},
			wantMode: 0,
		},
		{
			name:     "two_element_tie_smaller_wins",
			input:    []int{420, 570},
			wantMode: 420,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := modeMinute(tc.input)
			if got != tc.wantMode {
				t.Errorf("modeMinute(%v) = %d, want %d", tc.input, got, tc.wantMode)
			}
		})
	}
}
