package market

import (
	"testing"
	"time"

	"github.com/quant5-lab/runner/runtime/context"
)

func TestNormalizeBars_AppliesExchangeProfileWithoutChangingOrder(t *testing.T) {
	bars := moscowWeekdayWeekendBars(t)

	cases := []struct {
		name       string
		symbol     string
		timezone   string
		wantCloses []float64
		wantZone   string
	}{
		{name: "moex profile applies exchange default (regular) preserving 7-day midday bars", symbol: "SBERP", wantCloses: []float64{1, 2, 3, 4}, wantZone: "Europe/Moscow"},
		{name: "binance profile keeps all days", symbol: "BTCUSDT", timezone: "UTC", wantCloses: []float64{1, 2, 3, 4}, wantZone: "UTC"},
		{name: "unknown profile preserves input", symbol: "AAPL", timezone: "America/New_York", wantCloses: []float64{1, 2, 3, 4}, wantZone: "America/New_York"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, gotZone := NormalizeBars(tc.symbol, "1h", tc.timezone, bars)
			if gotZone != tc.wantZone {
				t.Fatalf("timezone = %q, want %q", gotZone, tc.wantZone)
			}
			assertCloseSequence(t, got, tc.wantCloses)
		})
	}
}

func TestNormalizeBarsWithReferenceSession_SessionResolution(t *testing.T) {
	bars := moscowWeekdayWeekendBars(t)

	cases := []struct {
		name        string
		symbol      string
		timezone    string
		session     string
		wantCloses  []float64
		wantZone    string
		wantSession ReferenceSession
	}{
		{name: "explicit regular MOEX preserves 7-day midday bars (window only)", symbol: "SBERP", session: "regular", wantCloses: []float64{1, 2, 3, 4}, wantZone: "Europe/Moscow", wantSession: ReferenceSessionRegular},
		{name: "explicit always-open MOEX preserves all bars", symbol: "SBERP", session: "always-open", wantCloses: []float64{1, 2, 3, 4}, wantZone: "Europe/Moscow", wantSession: ReferenceSessionAlwaysOpen},
		{name: "blank session MOEX applies exchange default (regular) preserving midday bars", symbol: "SBERP", session: "", wantCloses: []float64{1, 2, 3, 4}, wantZone: "Europe/Moscow", wantSession: ReferenceSessionRegular},
		{name: "explicit regular unknown exchange preserves all bars", symbol: "AAPL", timezone: "America/New_York", session: "regular", wantCloses: []float64{1, 2, 3, 4}, wantZone: "America/New_York", wantSession: ReferenceSessionRegular},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, gotZone, gotSession := NormalizeBarsWithReferenceSession(tc.symbol, "1h", tc.timezone, tc.session, bars)
			if gotZone != tc.wantZone {
				t.Fatalf("timezone = %q, want %q", gotZone, tc.wantZone)
			}
			if gotSession != tc.wantSession {
				t.Fatalf("reference session = %q, want %q", gotSession, tc.wantSession)
			}
			assertCloseSequence(t, got, tc.wantCloses)
		})
	}
}

func TestNormalizeBarsWithReferenceSession_RegularCryptoSessionStaysAlwaysOpen(t *testing.T) {
	bars := []context.OHLCV{
		{Time: time.Date(2025, 8, 15, 13, 0, 0, 0, time.UTC).Unix(), Close: 1},
		{Time: time.Date(2025, 8, 16, 13, 0, 0, 0, time.UTC).Unix(), Close: 2},
		{Time: time.Date(2025, 8, 17, 13, 0, 0, 0, time.UTC).Unix(), Close: 3},
		{Time: time.Date(2025, 8, 18, 13, 0, 0, 0, time.UTC).Unix(), Close: 4},
	}

	got, gotZone, gotSession := NormalizeBarsWithReferenceSession("BTCUSDT", "1h", "", "regular", bars)

	if gotZone != "UTC" {
		t.Fatalf("timezone = %q, want UTC", gotZone)
	}
	if gotSession != ReferenceSessionRegular {
		t.Fatalf("reference session = %q, want %q", gotSession, ReferenceSessionRegular)
	}
	if len(got) != len(bars) {
		t.Fatalf("len = %d, want %d", len(got), len(bars))
	}
}

func TestNormalizeBarsForProfile_UsesCalendarAndPreservesSessionIdentity(t *testing.T) {
	bars := moscowWeekdayWeekendBars(t)
	profile := Profile{
		Exchange:         ExchangeMOEX,
		Timezone:         "Europe/Moscow",
		Calendar:         NewRegularWeekdayCalendar(NewWeekdaySet(time.Saturday, time.Sunday), NewDateSet("2025-08-17"), NewDateSet("2025-08-18")),
		ReferenceSession: ReferenceSessionRegular,
	}

	got, gotSession := NormalizeBarsForProfile(profile, "1h", bars)

	if gotSession != ReferenceSessionRegular {
		t.Fatalf("reference session = %q, want regular", gotSession)
	}
	assertCloseSequence(t, got, []float64{1, 3})
}

func TestNormalizeBars_HandlesEmptyInputsWithoutInventingBars(t *testing.T) {
	cases := []struct {
		name     string
		symbol   string
		timezone string
		bars     []context.OHLCV
		wantZone string
	}{
		{name: "nil regular-session input", symbol: "SBERP", bars: nil, wantZone: "Europe/Moscow"},
		{name: "empty always-open input", symbol: "BTCUSDT", bars: []context.OHLCV{}, wantZone: "UTC"},
		{name: "unknown symbol keeps explicit timezone", symbol: "UNKNOWN", timezone: "Asia/Tokyo", bars: nil, wantZone: "Asia/Tokyo"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, gotZone := NormalizeBars(tc.symbol, "1h", tc.timezone, tc.bars)
			if gotZone != tc.wantZone {
				t.Fatalf("timezone = %q, want %q", gotZone, tc.wantZone)
			}
			if len(got) != 0 {
				t.Fatalf("expected no bars, got %d", len(got))
			}
		})
	}
}

func moscowWeekdayWeekendBars(t *testing.T) []context.OHLCV {
	t.Helper()
	return []context.OHLCV{
		{Time: unixInMoscow(t, "2025-08-15 13:00"), Close: 1},
		{Time: unixInMoscow(t, "2025-08-16 13:00"), Close: 2},
		{Time: unixInMoscow(t, "2025-08-17 13:00"), Close: 3},
		{Time: unixInMoscow(t, "2025-08-18 13:00"), Close: 4},
	}
}

func assertCloseSequence(t *testing.T, got []context.OHLCV, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].Close != want[i] {
			t.Fatalf("bar %d close = %.2f, want %.2f", i, got[i].Close, want[i])
		}
	}
}

func TestNormalizeBars_DoesNotMutateAcceptedBars(t *testing.T) {
	bars := []context.OHLCV{
		{Time: unixInMoscow(t, "2025-08-18 13:00"), Open: 10, High: 12, Low: 9, Close: 11, Volume: 100},
	}

	got, _ := NormalizeBars("SBERP", "1h", "", bars)

	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0] != bars[0] {
		t.Fatalf("bar mutated: got %+v want %+v", got[0], bars[0])
	}
}

func TestNormalizeBarsWithMetadata_ExactTimestampUniverseWins(t *testing.T) {
	bars := moscowWeekdayWeekendBars(t)
	metadata := SourceMetadata{
		Timezone:           "Europe/Moscow",
		ReferenceSession:   "regular",
		IncludedTimestamps: []int64{bars[1].Time, bars[3].Time * 1000},
	}

	got, profile := NormalizeBarsWithMetadata("SBERP", "1h", metadata, bars)

	if profile.ReferenceSession != ReferenceSessionRegular {
		t.Fatalf("reference session = %q, want regular", profile.ReferenceSession)
	}
	assertCloseSequence(t, got, []float64{2, 4})
}

func TestNormalizeBarsWithMetadata_DateUniverseWins(t *testing.T) {
	bars := moscowWeekdayWeekendBars(t)
	metadata := SourceMetadata{
		Timezone:         "Europe/Moscow",
		ReferenceSession: "regular",
		IncludedDates:    []string{"2025-08-16", "2025-08-18"},
	}

	got, _ := NormalizeBarsWithMetadata("SBERP", "1h", metadata, bars)

	assertCloseSequence(t, got, []float64{2, 4})
}

func TestNormalizeBarsWithMetadata_ExplicitCalendarOverridesExchangeGuess(t *testing.T) {
	bars := moscowWeekdayWeekendBars(t)
	metadata := SourceMetadata{
		Exchange:         "MOEX",
		Timezone:         "Europe/Moscow",
		ReferenceSession: "regular",
		ClosedWeekdays:   []string{"monday"},
		ClosedDates:      []string{"2025-08-15"},
		OpenDates:        []string{"2025-08-18"},
	}

	got, profile := NormalizeBarsWithMetadata("UNKNOWN", "1h", metadata, bars)

	if profile.Exchange != ExchangeMOEX {
		t.Fatalf("exchange = %q, want MOEX", profile.Exchange)
	}
	assertCloseSequence(t, got, []float64{2, 3, 4})
}

func TestNormalizeBarsWithReferenceSession_MOEXRegularFiltersBeforeWindowStart(t *testing.T) {
	bars := []context.OHLCV{
		{Time: unixInMoscow(t, "2025-08-15 06:00"), Close: 1},
		{Time: unixInMoscow(t, "2025-08-15 07:00"), Close: 2},
		{Time: unixInMoscow(t, "2025-08-15 13:00"), Close: 3},
		{Time: unixInMoscow(t, "2025-08-16 06:00"), Close: 4},
		{Time: unixInMoscow(t, "2025-08-18 06:00"), Close: 5},
		{Time: unixInMoscow(t, "2025-08-18 07:00"), Close: 6},
	}

	got, _, _ := NormalizeBarsWithReferenceSession("SBERP", "1h", "", "regular", bars)

	assertCloseSequence(t, got, []float64{2, 3, 6})
}

func TestNormalizeBarsWithReferenceSession_MOEXAlwaysOpenPreservesAllHours(t *testing.T) {
	bars := []context.OHLCV{
		{Time: unixInMoscow(t, "2025-08-15 06:00"), Close: 1},
		{Time: unixInMoscow(t, "2025-08-15 07:00"), Close: 2},
	}

	got, _, _ := NormalizeBarsWithReferenceSession("SBERP", "1h", "", "always-open", bars)

	assertCloseSequence(t, got, []float64{1, 2})
}

func TestNormalizeBarsWithMetadata_SessionWindowOverridesExchangeDefault(t *testing.T) {
	bars := []context.OHLCV{
		{Time: unixInMoscow(t, "2025-08-15 06:00"), Close: 1},
		{Time: unixInMoscow(t, "2025-08-15 07:00"), Close: 2},
		{Time: unixInMoscow(t, "2025-08-15 13:00"), Close: 3},
	}
	metadata := SourceMetadata{
		Exchange:         "MOEX",
		Timezone:         "Europe/Moscow",
		ReferenceSession: "regular",
		SessionWindow:    "0600-2350",
	}

	got, _ := NormalizeBarsWithMetadata("SBERP", "1h", metadata, bars)

	assertCloseSequence(t, got, []float64{1, 2, 3})
}

func TestNormalizeBarsWithMetadata_SplitSessionWindows(t *testing.T) {
	bars := []context.OHLCV{
		{Time: unixInMoscow(t, "2025-08-15 05:59"), Close: 1},
		{Time: unixInMoscow(t, "2025-08-15 06:00"), Close: 2},
		{Time: unixInMoscow(t, "2025-08-15 22:59"), Close: 3},
		{Time: unixInMoscow(t, "2025-08-15 23:00"), Close: 4},
		{Time: unixInMoscow(t, "2025-08-16 10:59"), Close: 5},
		{Time: unixInMoscow(t, "2025-08-16 11:00"), Close: 6},
		{Time: unixInMoscow(t, "2025-08-16 16:59"), Close: 7},
		{Time: unixInMoscow(t, "2025-08-16 17:00"), Close: 8},
	}
	metadata := SourceMetadata{
		Exchange:             "MOEX",
		Timezone:             "Europe/Moscow",
		ReferenceSession:     "regular",
		WeekdaySessionWindow: "0600-2300",
		WeekendSessionWindow: "1100-1700",
	}

	got, _ := NormalizeBarsWithMetadata("SBERP", "1h", metadata, bars)

	assertCloseSequence(t, got, []float64{2, 3, 6, 7})
}

func TestNormalizeBarsWithMetadata_LegacyInvalidSessionWindowFallsBackToExchangeDefault(t *testing.T) {
	bars := []context.OHLCV{
		{Time: unixInMoscow(t, "2025-08-15 06:00"), Close: 1},
		{Time: unixInMoscow(t, "2025-08-15 07:00"), Close: 2},
		{Time: unixInMoscow(t, "2025-08-15 13:00"), Close: 3},
	}
	metadata := SourceMetadata{
		Exchange:         "MOEX",
		Timezone:         "Europe/Moscow",
		ReferenceSession: "regular",
		SessionWindow:    "not-a-window",
	}

	got, _ := NormalizeBarsWithMetadata("SBERP", "1h", metadata, bars)

	assertCloseSequence(t, got, []float64{2, 3})
}

func TestNormalizeBarsWithMetadataE_InvalidSessionMetadataFails(t *testing.T) {
	bars := []context.OHLCV{{Time: unixInMoscow(t, "2025-08-15 07:00"), Close: 1}}

	cases := []struct {
		name     string
		metadata SourceMetadata
	}{
		{
			name: "uniform window",
			metadata: SourceMetadata{
				Exchange:         "MOEX",
				Timezone:         "Europe/Moscow",
				ReferenceSession: "regular",
				SessionWindow:    "not-a-window",
			},
		},
		{
			name: "weekday window",
			metadata: SourceMetadata{
				Exchange:             "MOEX",
				Timezone:             "Europe/Moscow",
				ReferenceSession:     "regular",
				WeekdaySessionWindow: "not-a-window",
			},
		},
		{
			name: "weekend window",
			metadata: SourceMetadata{
				Exchange:             "MOEX",
				Timezone:             "Europe/Moscow",
				ReferenceSession:     "regular",
				WeekendSessionWindow: "not-a-window",
			},
		},
		{
			name: "date window date",
			metadata: SourceMetadata{
				Exchange:           "MOEX",
				Timezone:           "Europe/Moscow",
				ReferenceSession:   "regular",
				DateSessionWindows: map[string]string{"2025/08/15": "0700-1900"},
			},
		},
		{
			name: "date window value",
			metadata: SourceMetadata{
				Exchange:           "MOEX",
				Timezone:           "Europe/Moscow",
				ReferenceSession:   "regular",
				DateSessionWindows: map[string]string{"2025-08-15": "not-a-window"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := NormalizeBarsWithMetadataE("SBERP", "1h", tc.metadata, bars)
			if err == nil {
				t.Fatal("expected metadata validation error")
			}
		})
	}
}

func TestNormalizeBarsWithMetadata_ExplicitUniversePrecedence(t *testing.T) {
	bars := []context.OHLCV{
		{Time: unixInMoscow(t, "2025-08-15 06:00"), Close: 1},
		{Time: unixInMoscow(t, "2025-08-15 07:00"), Close: 2},
		{Time: unixInMoscow(t, "2025-08-16 09:00"), Close: 3},
		{Time: unixInMoscow(t, "2025-08-16 10:00"), Close: 4},
	}

	cases := []struct {
		name       string
		metadata   SourceMetadata
		wantCloses []float64
	}{
		{
			name: "timestamps ignore invalid session metadata",
			metadata: SourceMetadata{
				Exchange:             "MOEX",
				Timezone:             "Europe/Moscow",
				ReferenceSession:     "regular",
				SessionWindow:        "not-a-window",
				WeekdaySessionWindow: "also-bad",
				DateSessionWindows:   map[string]string{"not-a-date": "bad"},
				IncludedTimestamps:   []int64{bars[0].Time, bars[2].Time * 1000},
			},
			wantCloses: []float64{1, 3},
		},
		{
			name: "dates ignore invalid session metadata",
			metadata: SourceMetadata{
				Exchange:             "MOEX",
				Timezone:             "Europe/Moscow",
				ReferenceSession:     "regular",
				WeekendSessionWindow: "not-a-window",
				DateSessionWindows:   map[string]string{"not-a-date": "bad"},
				IncludedDates:        []string{"2025-08-16"},
			},
			wantCloses: []float64{3, 4},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, _, err := NormalizeBarsWithMetadataE("SBERP", "1h", tc.metadata, bars)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertCloseSequence(t, got, tc.wantCloses)
		})
	}
}

func TestNormalizeBarsWithMetadata_DateSessionWindowOverridesSingleDate(t *testing.T) {
	bars := []context.OHLCV{
		{Time: unixInMoscow(t, "2025-08-16 11:59"), Close: 1},
		{Time: unixInMoscow(t, "2025-08-16 12:00"), Close: 2},
		{Time: unixInMoscow(t, "2025-08-16 14:59"), Close: 3},
		{Time: unixInMoscow(t, "2025-08-16 15:00"), Close: 4},
		{Time: unixInMoscow(t, "2025-08-17 12:00"), Close: 5},
	}
	metadata := SourceMetadata{
		Exchange:           "MOEX",
		Timezone:           "Europe/Moscow",
		ReferenceSession:   "regular",
		DateSessionWindows: map[string]string{"2025-08-16": "1200-1500"},
	}

	got, _, err := NormalizeBarsWithMetadataE("SBERP", "1h", metadata, bars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertCloseSequence(t, got, []float64{2, 3, 5})
}

func TestNormalizeBarsForProfile_NilCalendarPassesThroughBars(t *testing.T) {
	bars := moscowWeekdayWeekendBars(t)
	profile := Profile{
		Exchange:         ExchangeMOEX,
		Timezone:         "Europe/Moscow",
		Calendar:         nil,
		ReferenceSession: ReferenceSessionRegular,
	}

	got, gotSession := NormalizeBarsForProfile(profile, "1h", bars)

	if gotSession != ReferenceSessionRegular {
		t.Fatalf("reference session = %q, want regular", gotSession)
	}
	assertCloseSequence(t, got, []float64{1, 2, 3, 4})
}

func TestNormalizeBarsWithReferenceSession_MOEXRegularFiltersMillisecondTimestampBars(t *testing.T) {
	bars := []context.OHLCV{
		{Time: unixInMoscow(t, "2025-08-15 06:00") * 1000, Close: 1},
		{Time: unixInMoscow(t, "2025-08-15 07:00") * 1000, Close: 2},
		{Time: unixInMoscow(t, "2025-08-15 13:00") * 1000, Close: 3},
	}

	got, _, _ := NormalizeBarsWithReferenceSession("SBERP", "1h", "", "regular", bars)

	assertCloseSequence(t, got, []float64{2, 3})
}

func TestNormalizeBars_DailyTimeframePreservesAllBars(t *testing.T) {
	dailyBars := []context.OHLCV{
		{Time: unixInMoscow(t, "2025-08-15 00:00"), Close: 1},
		{Time: unixInMoscow(t, "2025-08-16 23:00"), Close: 2},
		{Time: unixInMoscow(t, "2025-08-17 00:00"), Close: 3},
		{Time: unixInMoscow(t, "2025-08-18 05:00"), Close: 4},
	}

	cases := []struct {
		name      string
		symbol    string
		timeframe string
		timezone  string
	}{
		{name: "MOEX/1D", symbol: "SBERP", timeframe: "1D", timezone: "Europe/Moscow"},
		{name: "MOEX/D alias", symbol: "SBERP", timeframe: "D", timezone: "Europe/Moscow"},
		{name: "MOEX/1W", symbol: "SBERP", timeframe: "1W", timezone: "Europe/Moscow"},
		{name: "MOEX/1M", symbol: "SBERP", timeframe: "1M", timezone: "Europe/Moscow"},
		{name: "Binance/1D", symbol: "BTCUSDT", timeframe: "1D", timezone: "UTC"},
		{name: "Unknown/1D", symbol: "AAPL", timeframe: "1D", timezone: "America/New_York"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := NormalizeBars(tc.symbol, tc.timeframe, tc.timezone, dailyBars)
			if len(got) != len(dailyBars) {
				t.Fatalf("NormalizeBars(%q) = %d bars, want %d — daily bars must not be filtered by the session window",
					tc.timeframe, len(got), len(dailyBars))
			}
		})
	}
}
