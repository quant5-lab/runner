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
		{name: "moex profile keeps all days without explicit reference session", symbol: "SBERP", wantCloses: []float64{1, 2, 3, 4}, wantZone: "Europe/Moscow"},
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

func TestNormalizeBarsWithReferenceSession_FiltersOnlyWhenSourceRequestsRegularSession(t *testing.T) {
	bars := moscowWeekdayWeekendBars(t)

	cases := []struct {
		name       string
		symbol     string
		timezone   string
		session    string
		wantCloses []float64
		wantZone   string
	}{
		{name: "explicit regular MOEX filters closed weekdays", symbol: "SBERP", session: "regular", wantCloses: []float64{1, 4}, wantZone: "Europe/Moscow"},
		{name: "explicit always-open MOEX preserves raw bars", symbol: "SBERP", session: "always-open", wantCloses: []float64{1, 2, 3, 4}, wantZone: "Europe/Moscow"},
		{name: "blank session MOEX preserves raw bars", symbol: "SBERP", session: "", wantCloses: []float64{1, 2, 3, 4}, wantZone: "Europe/Moscow"},
		{name: "explicit regular unknown exchange preserves raw bars", symbol: "AAPL", timezone: "America/New_York", session: "regular", wantCloses: []float64{1, 2, 3, 4}, wantZone: "America/New_York"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, gotZone, gotSession := NormalizeBarsWithReferenceSession(tc.symbol, "1h", tc.timezone, tc.session, bars)
			if gotZone != tc.wantZone {
				t.Fatalf("timezone = %q, want %q", gotZone, tc.wantZone)
			}
			if gotSession != ParseReferenceSession(tc.session) {
				t.Fatalf("reference session = %q, want %q", gotSession, ParseReferenceSession(tc.session))
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
