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

// TestSessionOpenMinuteFor_RegularVsAlwaysOpenAcrossExchanges verifies that the
// session-open minute carried in the profile is non-zero for regular-session
// exchanges (MOEX, NYSE) and zero for always-open exchanges, independently of
// the exchange type.  This property is what determines whether period tiling
// is session-anchored or midnight-anchored.
func TestSessionOpenMinuteFor_RegularVsAlwaysOpenAcrossExchanges(t *testing.T) {
	cases := []struct {
		name        string
		regular     Profile
		alwaysOpen  Profile
		wantNonZero bool
	}{
		{
			name:        "MOEX",
			regular:     ResolveProfileWithReferenceSession("SBERP", "Europe/Moscow", "regular"),
			alwaysOpen:  ResolveProfileWithReferenceSession("SBERP", "Europe/Moscow", "always-open"),
			wantNonZero: true,
		},
		{
			name:        "NYSE",
			regular:     ResolveProfileWithMetadata("AAPL", SourceMetadata{Exchange: "NYSE", ReferenceSession: "regular"}),
			alwaysOpen:  ResolveProfileWithMetadata("AAPL", SourceMetadata{Exchange: "NYSE", ReferenceSession: "always-open"}),
			wantNonZero: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantNonZero && tc.regular.SessionOpenMinute == 0 {
				t.Errorf("regular %s profile must have non-zero SessionOpenMinute", tc.name)
			}
			if tc.alwaysOpen.SessionOpenMinute != 0 {
				t.Errorf("always-open %s profile must have SessionOpenMinute=0; got %d", tc.name, tc.alwaysOpen.SessionOpenMinute)
			}
		})
	}
}

// TestDeriveSessionAnchor_TableExchangePrevailsOverBars verifies that when the
// exchange table carries a non-zero session-open minute (MOEX, NYSE), the bar
// data is ignored -- the table answer is authoritative and stable regardless of
// what bars are supplied.
func TestDeriveSessionAnchor_TableExchangePrevailsOverBars(t *testing.T) {
	cases := []struct {
		name     string
		profile  Profile
		bars     []context.OHLCV
		wantOpen int
	}{
		{
			name:    "MOEX_ignores_bars_at_wrong_time",
			profile: ResolveProfile("SBERP", ""),
			// Bars at 09:30 MSK differ from MOEX 07:00, yet the table must win.
			bars:     []context.OHLCV{barAtLocal(t, "Europe/Moscow", 9, 30)},
			wantOpen: 7 * 60,
		},
		{
			name:    "NYSE_explicit_ignores_bars_at_wrong_time",
			profile: ResolveProfileWithMetadata("AAPL", SourceMetadata{Exchange: "NYSE"}),
			// Bars at 10:00 ET differ from NYSE 09:30, yet the table must win.
			bars:     []context.OHLCV{barAtLocal(t, "America/New_York", 10, 0)},
			wantOpen: 9*60 + 30,
		},
		{
			name:    "Binance_known_exchange_returns_zero_regardless_of_bars",
			profile: ResolveProfile("BINANCE:BTCUSDT", ""),
			// 0 is the correct midnight-origin anchor for 24/7 always-open markets.
			bars:     []context.OHLCV{barAtLocal(t, "UTC", 12, 0)},
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

// TestDeriveSessionAnchor_UnknownExchangeObservesSessionOpenFromBars verifies
// that when the exchange table has no session-open offset (ExchangeUnknown),
// DeriveSessionAnchor derives the session-open minute from the supplied bars.
// It recognises the dominant session open across varying timezones and session
// configurations, covering any exchange whose metadata lacks an explicit exchange
// code but whose fixture data consistently begins at the session open.
func TestDeriveSessionAnchor_UnknownExchangeObservesSessionOpenFromBars(t *testing.T) {
	cases := []struct {
		name     string
		tz       string
		barTimes [][2]int // (hour, minute) pairs
		wantOpen int
	}{
		{
			name:     "NYSE_session_open_from_ET_bars",
			tz:       "America/New_York",
			barTimes: [][2]int{{9, 30}, {10, 30}, {11, 30}, {14, 0}, {15, 30}},
			wantOpen: 9*60 + 30,
		},
		{
			name:     "MOEX_session_open_from_MSK_bars_without_exchange_code",
			tz:       "Europe/Moscow",
			barTimes: [][2]int{{7, 0}, {8, 0}, {11, 0}, {23, 0}},
			wantOpen: 7 * 60,
		},
		{
			name:     "UTC_midnight_session_open_stays_zero",
			tz:       "UTC",
			barTimes: [][2]int{{0, 0}, {1, 0}, {12, 0}, {23, 0}},
			wantOpen: 0,
		},
		{
			name:     "non_round_hour_session_open_preserved",
			tz:       "America/New_York",
			barTimes: [][2]int{{9, 30}, {10, 0}},
			wantOpen: 9*60 + 30,
		},
		{
			name:     "single_bar_dataset",
			tz:       "America/New_York",
			barTimes: [][2]int{{9, 30}},
			wantOpen: 9*60 + 30,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			profile := ResolveProfileWithMetadata("UNKNOWN", SourceMetadata{Timezone: tc.tz})

			// Precondition: table must yield 0 so the data-derived path is exercised.
			if profile.SessionOpenMinute != 0 {
				t.Fatalf("precondition violated: unknown-exchange profile must have SessionOpenMinute=0; got %d", profile.SessionOpenMinute)
			}

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

// TestDeriveSessionAnchor_MultiDayDataSetConvergesToSessionOpen verifies that a
// dataset spanning multiple trading days, where every day begins at the same
// session-open minute, produces an anchor whose SessionOpenMinute matches that
// consistent session open.
func TestDeriveSessionAnchor_MultiDayDataSetConvergesToSessionOpen(t *testing.T) {
	tz := "America/New_York"
	profile := ResolveProfileWithMetadata("AAPL", SourceMetadata{Timezone: tz})

	loc, _ := time.LoadLocation(tz)
	var bars []context.OHLCV
	for day := 1; day <= 5; day++ {
		for hour := 9; hour <= 15; hour++ {
			minute := 30
			if hour > 9 {
				minute = 0
			}
			ts := time.Date(2025, 10, day, hour, minute, 0, 0, loc).Unix()
			bars = append(bars, context.OHLCV{Time: ts})
		}
	}

	anchor := DeriveSessionAnchor(profile, bars)

	const wantOpen = 9*60 + 30
	if anchor.SessionOpenMinute != wantOpen {
		t.Errorf("five-day dataset: SessionOpenMinute = %d, want %d (09:30 ET)", anchor.SessionOpenMinute, wantOpen)
	}
}

// TestDeriveSessionAnchor_NilAndEmptyBarsYieldTableFallback verifies that when
// no bar data is available the function returns the table-based answer (zero for
// unknown exchanges) and never panics.
func TestDeriveSessionAnchor_NilAndEmptyBarsYieldTableFallback(t *testing.T) {
	profile := ResolveProfileWithMetadata("AAPL", SourceMetadata{Timezone: "America/New_York"})

	for _, bars := range [][]context.OHLCV{nil, {}} {
		label := "nil"
		if bars != nil {
			label = "empty"
		}
		t.Run(label, func(t *testing.T) {
			anchor := DeriveSessionAnchor(profile, bars)
			if anchor.SessionOpenMinute != 0 {
				t.Errorf("%s bars: got %d, want 0 (table fallback for unknown exchange)", label, anchor.SessionOpenMinute)
			}
			if anchor.Timezone != "America/New_York" {
				t.Errorf("%s bars: timezone got %q, want America/New_York", label, anchor.Timezone)
			}
		})
	}
}

// TestDeriveSessionAnchor_TimezoneSourceIsAlwaysProfile verifies that the
// timezone in the returned anchor always comes from the profile, not inferred
// from bar timestamps.  Even if bars are supplied in a different timezone than
// the profile, the anchor carries the profile timezone.
func TestDeriveSessionAnchor_TimezoneSourceIsAlwaysProfile(t *testing.T) {
	// Profile carries Europe/Moscow but bars are at UTC timestamps.
	profile := ResolveProfileWithMetadata("UNKNOWN", SourceMetadata{Timezone: "Europe/Moscow"})
	bars := []context.OHLCV{
		{Time: time.Date(2025, 10, 1, 10, 0, 0, 0, time.UTC).Unix()},
	}

	anchor := DeriveSessionAnchor(profile, bars)

	if anchor.Timezone != "Europe/Moscow" {
		t.Errorf("Timezone: got %q, want Europe/Moscow (must come from profile, not bars)", anchor.Timezone)
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
			// Minority of two does not override majority of eight.
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

// TestDeriveSessionAnchor_OutlierDaysDoNotShiftAnchorFromDominantOpen verifies
// that when a minority of trading days in the bar set have an unusually early
// first bar (e.g. a pre-market or extended-hours print), the derived anchor
// still reflects the dominant session open across the remaining days.  This is
// the end-to-end correctness guarantee of the mode-of-per-day-first-bar design:
// isolated anomalies do not distort the anchor as long as they are in the
// minority.
func TestDeriveSessionAnchor_OutlierDaysDoNotShiftAnchorFromDominantOpen(t *testing.T) {
	tz := "America/New_York"
	profile := ResolveProfileWithMetadata("AAPL", SourceMetadata{Timezone: tz})

	// Precondition: ExchangeUnknown (no explicit exchange code) must have
	// SessionOpenMinute=0 from the table so the data-derived path is exercised.
	if profile.SessionOpenMinute != 0 {
		t.Fatalf("precondition violated: AAPL without exchange code must have SessionOpenMinute=0; got %d",
			profile.SessionOpenMinute)
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		t.Fatalf("load timezone: %v", err)
	}

	regularHours := [][2]int{{9, 30}, {10, 30}, {11, 30}, {13, 30}, {15, 30}}
	anomalyBar := context.OHLCV{Time: time.Date(2025, 10, 6, 4, 0, 0, 0, loc).Unix()}

	var bars []context.OHLCV
	bars = append(bars, anomalyBar)
	for day := 6; day <= 9; day++ {
		for _, hm := range regularHours {
			ts := time.Date(2025, 10, day, hm[0], hm[1], 0, 0, loc).Unix()
			bars = append(bars, context.OHLCV{Time: ts})
		}
	}

	anchor := DeriveSessionAnchor(profile, bars)

	const wantOpen = 9*60 + 30
	if anchor.SessionOpenMinute != wantOpen {
		t.Errorf("SessionOpenMinute = %d, want %d (09:30 ET); anomaly day must not shift anchor",
			anchor.SessionOpenMinute, wantOpen)
	}
}
