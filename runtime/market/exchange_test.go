package market

import (
	"testing"
	"time"

	"github.com/quant5-lab/runner/runtime/context"
)

func TestResolveExchange_KnownSymbolFamilies(t *testing.T) {
	tests := []struct {
		name     string
		symbol   string
		exchange Exchange
	}{
		{name: "moex base symbol", symbol: "SBERP", exchange: ExchangeMOEX},
		{name: "moex prefixed symbol", symbol: "MOEX:SBER", exchange: ExchangeMOEX},
		{name: "binance quote convention", symbol: "BTCUSDT", exchange: ExchangeBinance},
		{name: "binance prefixed symbol", symbol: "BINANCE:ETHUSDT", exchange: ExchangeBinance},
		{name: "unknown equity", symbol: "AAPL", exchange: ExchangeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveExchange(tt.symbol); got != tt.exchange {
				t.Fatalf("expected %q, got %q", tt.exchange, got)
			}
		})
	}
}

func TestResolveProfile_DefaultTimezone(t *testing.T) {
	tests := []struct {
		name     string
		symbol   string
		timezone string
		want     string
	}{
		{name: "moex default", symbol: "SBERP", want: "Europe/Moscow"},
		{name: "crypto default", symbol: "BTCUSDT", want: "UTC"},
		{name: "explicit timezone wins", symbol: "SBERP", timezone: "UTC", want: "UTC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := ResolveProfile(tt.symbol, tt.timezone)
			if profile.Timezone != tt.want {
				t.Fatalf("expected timezone %q, got %q", tt.want, profile.Timezone)
			}
		})
	}
}

func TestResolveProfileWithReferenceSession_CalendarSelection(t *testing.T) {
	tests := []struct {
		name         string
		symbol       string
		session      string
		wantSession  ReferenceSession
		wantCalendar any
	}{
		{name: "regular MOEX gets regular calendar", symbol: "SBERP", session: "regular", wantSession: ReferenceSessionRegular, wantCalendar: RegularSessionCalendar{}},
		{name: "regular crypto remains always open", symbol: "BTCUSDT", session: "regular", wantSession: ReferenceSessionRegular, wantCalendar: AlwaysOpenCalendar{}},
		{name: "blank MOEX applies exchange default (regular)", symbol: "SBERP", session: "", wantSession: ReferenceSessionRegular, wantCalendar: RegularSessionCalendar{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := ResolveProfileWithReferenceSession(tt.symbol, "", tt.session)
			if profile.ReferenceSession != tt.wantSession {
				t.Fatalf("reference session = %q, want %q", profile.ReferenceSession, tt.wantSession)
			}
			if profile.Calendar == nil {
				t.Fatal("calendar is nil")
			}
			switch tt.wantCalendar.(type) {
			case RegularSessionCalendar:
				if _, ok := profile.Calendar.(RegularSessionCalendar); !ok {
					t.Fatalf("calendar = %T, want RegularSessionCalendar", profile.Calendar)
				}
			case AlwaysOpenCalendar:
				if _, ok := profile.Calendar.(AlwaysOpenCalendar); !ok {
					t.Fatalf("calendar = %T, want AlwaysOpenCalendar", profile.Calendar)
				}
			}
		})
	}
}

func TestResolveProfileWithMetadata_UsesExplicitMetadataBeforeSymbolFallback(t *testing.T) {
	profile := ResolveProfileWithMetadata("UNKNOWN", SourceMetadata{
		Exchange:         "MOEX",
		Timezone:         "Asia/Tokyo",
		ReferenceSession: "regular",
		SessionSource:    "provider-feed",
		CalendarID:       "provider:regular",
	})

	if profile.Exchange != ExchangeMOEX {
		t.Fatalf("exchange = %q, want MOEX", profile.Exchange)
	}
	if profile.Timezone != "Asia/Tokyo" {
		t.Fatalf("timezone = %q, want Asia/Tokyo", profile.Timezone)
	}
	if profile.SessionSource != "provider-feed" || profile.CalendarID != "provider:regular" {
		t.Fatalf("metadata not preserved: source=%q calendarID=%q", profile.SessionSource, profile.CalendarID)
	}
}

func TestCalendarFor_MetadataUniversePrecedence(t *testing.T) {
	cases := []struct {
		name     string
		metadata SourceMetadata
		want     any
	}{
		{name: "exact timestamps override session calendar", metadata: SourceMetadata{ReferenceSession: "regular", IncludedTimestamps: []int64{1}}, want: ExactTimestampCalendar{}},
		{name: "included dates override session calendar", metadata: SourceMetadata{ReferenceSession: "regular", IncludedDates: []string{"2025-08-16"}}, want: DateWhitelistCalendar{}},
		{name: "regular exchange uses regular calendar", metadata: SourceMetadata{Exchange: "MOEX", ReferenceSession: "regular"}, want: RegularSessionCalendar{}},
		{name: "regular without calendar facts remains always-open", metadata: SourceMetadata{ReferenceSession: "regular"}, want: AlwaysOpenCalendar{}},
		{name: "always-open ignores exchange calendar", metadata: SourceMetadata{Exchange: "MOEX", ReferenceSession: "always-open"}, want: AlwaysOpenCalendar{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			profile := ResolveProfileWithMetadata("UNKNOWN", tc.metadata)
			switch tc.want.(type) {
			case ExactTimestampCalendar:
				if _, ok := profile.Calendar.(ExactTimestampCalendar); !ok {
					t.Fatalf("calendar = %T, want ExactTimestampCalendar", profile.Calendar)
				}
			case DateWhitelistCalendar:
				if _, ok := profile.Calendar.(DateWhitelistCalendar); !ok {
					t.Fatalf("calendar = %T, want DateWhitelistCalendar", profile.Calendar)
				}
			case RegularSessionCalendar:
				if _, ok := profile.Calendar.(RegularSessionCalendar); !ok {
					t.Fatalf("calendar = %T, want RegularSessionCalendar", profile.Calendar)
				}
			case AlwaysOpenCalendar:
				if _, ok := profile.Calendar.(AlwaysOpenCalendar); !ok {
					t.Fatalf("calendar = %T, want AlwaysOpenCalendar", profile.Calendar)
				}
			}
		})
	}
}

func TestRegularCalendarForExchange_WeekdayBehavior(t *testing.T) {
	// 2025-09-01 Mon through 2025-09-07 Sun at 12:00 UTC = 15:00 MSK (in-session for MOEX).
	type weekRow struct {
		weekday time.Weekday
		bar     context.OHLCV
	}
	base := time.Date(2025, time.September, 1, 12, 0, 0, 0, time.UTC) // Monday
	weekBars := make([]weekRow, 7)
	for i := range weekBars {
		instant := base.AddDate(0, 0, i)
		weekBars[i] = weekRow{weekday: instant.Weekday(), bar: context.OHLCV{Time: instant.Unix()}}
	}

	cases := []struct {
		name           string
		exchange       Exchange
		acceptWeekdays map[time.Weekday]bool
	}{
		{
			name:     "MOEX_accepts_all_seven_days_filtering_by_window_only",
			exchange: ExchangeMOEX,
			acceptWeekdays: map[time.Weekday]bool{
				time.Monday:    true,
				time.Tuesday:   true,
				time.Wednesday: true,
				time.Thursday:  true,
				time.Friday:    true,
				time.Saturday:  true,
				time.Sunday:    true,
			},
		},
		{
			name:     "Binance_accepts_all_seven_days",
			exchange: ExchangeBinance,
			acceptWeekdays: map[time.Weekday]bool{
				time.Monday:    true,
				time.Tuesday:   true,
				time.Wednesday: true,
				time.Thursday:  true,
				time.Friday:    true,
				time.Saturday:  true,
				time.Sunday:    true,
			},
		},
		{
			name:     "Unknown_accepts_all_seven_days",
			exchange: ExchangeUnknown,
			acceptWeekdays: map[time.Weekday]bool{
				time.Monday:    true,
				time.Tuesday:   true,
				time.Wednesday: true,
				time.Thursday:  true,
				time.Friday:    true,
				time.Saturday:  true,
				time.Sunday:    true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cal := RegularCalendarForExchange(tc.exchange, SourceMetadata{})
			for _, row := range weekBars {
				want := tc.acceptWeekdays[row.weekday]
				if got := cal.Accepts(row.bar, "1h", "UTC"); got != want {
					t.Errorf("%s: Accepts(%s) = %v, want %v", tc.name, row.weekday, got, want)
				}
			}
		})
	}

	// Window rejection (not weekday mask) — 03:00 UTC = 06:00 MSK, before 07:00 start.
	t.Run("MOEX_rejects_presession_hour_on_all_weekdays", func(t *testing.T) {
		loc, _ := time.LoadLocation("Europe/Moscow")
		cal := RegularCalendarForExchange(ExchangeMOEX, SourceMetadata{})
		preSessionBase := time.Date(2025, time.September, 1, 3, 0, 0, 0, time.UTC) // Mon 06:00 MSK
		for i := 0; i < 7; i++ {
			instant := preSessionBase.AddDate(0, 0, i)
			bar := context.OHLCV{Time: instant.Unix()}
			if cal.Accepts(bar, "1h", "Europe/Moscow") {
				t.Errorf("MOEX accepted pre-session bar on %s (%s) — session window must reject all weekdays before 07:00 MSK",
					instant.In(loc).Weekday(), instant.In(loc).Format("15:04"))
			}
		}
	})
}

func TestRegularCalendarForExchange_SessionWindowContract(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Moscow")
	mosUnix := func(datetime string) int64 {
		ts, _ := time.ParseInLocation("2006-01-02 15:04", datetime, loc)
		return ts.Unix()
	}

	cases := []struct {
		name     string
		exchange Exchange
		metadata SourceMetadata
		datetime string
		timezone string
		want     bool
	}{
		{name: "MOEX/before window start rejected", exchange: ExchangeMOEX, datetime: "2025-08-15 06:00", timezone: "Europe/Moscow", want: false},
		{name: "MOEX/at window start accepted", exchange: ExchangeMOEX, datetime: "2025-08-15 07:00", timezone: "Europe/Moscow", want: true},
		{name: "MOEX/midday accepted", exchange: ExchangeMOEX, datetime: "2025-08-15 13:00", timezone: "Europe/Moscow", want: true},
		{name: "MOEX/evening accepted", exchange: ExchangeMOEX, datetime: "2025-08-15 23:00", timezone: "Europe/Moscow", want: true},
		{name: "MOEX/one before window end accepted", exchange: ExchangeMOEX, datetime: "2025-08-15 23:49", timezone: "Europe/Moscow", want: true},
		{name: "MOEX/at window end rejected", exchange: ExchangeMOEX, datetime: "2025-08-15 23:50", timezone: "Europe/Moscow", want: false},
		{name: "MOEX/after window end rejected", exchange: ExchangeMOEX, datetime: "2025-08-16 00:00", timezone: "Europe/Moscow", want: false},

		{name: "Binance/midnight accepted", exchange: ExchangeBinance, datetime: "2025-08-15 00:00", timezone: "Europe/Moscow", want: true},
		{name: "Binance/early hour accepted", exchange: ExchangeBinance, datetime: "2025-08-15 06:00", timezone: "Europe/Moscow", want: true},
		{name: "Unknown/early hour accepted", exchange: ExchangeUnknown, datetime: "2025-08-15 06:00", timezone: "Europe/Moscow", want: true},

		{name: "metadata-override/0600 start accepts pre-default bar", exchange: ExchangeMOEX, metadata: SourceMetadata{SessionWindow: "0600-2350"}, datetime: "2025-08-15 06:00", timezone: "Europe/Moscow", want: true},
		{name: "metadata-override/0800 start rejects 07:00 bar", exchange: ExchangeMOEX, metadata: SourceMetadata{SessionWindow: "0800-2350"}, datetime: "2025-08-15 07:00", timezone: "Europe/Moscow", want: false},

		{name: "invalid-metadata/falls back to MOEX default", exchange: ExchangeMOEX, metadata: SourceMetadata{SessionWindow: "bad"}, datetime: "2025-08-15 06:00", timezone: "Europe/Moscow", want: false},
		{name: "invalid-metadata/exchange default window still open at 07:00", exchange: ExchangeMOEX, metadata: SourceMetadata{SessionWindow: "not-a-window"}, datetime: "2025-08-15 07:00", timezone: "Europe/Moscow", want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bar := context.OHLCV{Time: mosUnix(tc.datetime)}
			cal := RegularCalendarForExchange(tc.exchange, tc.metadata)
			if got := cal.Accepts(bar, "1h", tc.timezone); got != tc.want {
				t.Fatalf("Accepts(%s) = %v, want %v", tc.datetime, got, tc.want)
			}
		})
	}
}

func TestRegularCalendarForExchange_DailyTimeframeBypassesSessionWindow(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Moscow")
	mosUnixAt := func(datetime string) context.OHLCV {
		ts, _ := time.ParseInLocation("2006-01-02 15:04", datetime, loc)
		return context.OHLCV{Time: ts.Unix()}
	}

	cases := []struct {
		name      string
		exchange  Exchange
		metadata  SourceMetadata
		bar       context.OHLCV
		timeframe string
		timezone  string
		want      bool
	}{
		{name: "MOEX/1D bar at 00:00 accepted", exchange: ExchangeMOEX, bar: mosUnixAt("2025-08-15 00:00"), timeframe: "1D", timezone: "Europe/Moscow", want: true},
		{name: "MOEX/1D bar at 23:00 accepted", exchange: ExchangeMOEX, bar: mosUnixAt("2025-08-15 23:00"), timeframe: "1D", timezone: "Europe/Moscow", want: true},
		{name: "MOEX/1D bar at 06:00 accepted", exchange: ExchangeMOEX, bar: mosUnixAt("2025-08-15 06:00"), timeframe: "1D", timezone: "Europe/Moscow", want: true},
		{name: "MOEX/1D bar at 03:00 accepted", exchange: ExchangeMOEX, bar: mosUnixAt("2025-08-15 03:00"), timeframe: "1D", timezone: "Europe/Moscow", want: true},
		{name: "MOEX/D alias accepted", exchange: ExchangeMOEX, bar: mosUnixAt("2025-08-15 00:00"), timeframe: "D", timezone: "Europe/Moscow", want: true},

		{name: "MOEX/1h bar at 06:00 rejected", exchange: ExchangeMOEX, bar: mosUnixAt("2025-08-15 06:00"), timeframe: "1h", timezone: "Europe/Moscow", want: false},
		{name: "MOEX/1h bar at 07:00 accepted", exchange: ExchangeMOEX, bar: mosUnixAt("2025-08-15 07:00"), timeframe: "1h", timezone: "Europe/Moscow", want: true},

		{name: "MOEX/1W bar at 00:00 accepted", exchange: ExchangeMOEX, bar: mosUnixAt("2025-08-15 00:00"), timeframe: "1W", timezone: "Europe/Moscow", want: true},
		{name: "MOEX/1M bar at 00:00 accepted", exchange: ExchangeMOEX, bar: mosUnixAt("2025-08-15 00:00"), timeframe: "1M", timezone: "Europe/Moscow", want: true},

		{name: "Binance/1D any hour accepted", exchange: ExchangeBinance, bar: mosUnixAt("2025-08-15 00:00"), timeframe: "1D", timezone: "Europe/Moscow", want: true},
		{name: "Unknown/1D any hour accepted", exchange: ExchangeUnknown, bar: mosUnixAt("2025-08-15 00:00"), timeframe: "1D", timezone: "Europe/Moscow", want: true},

		{name: "MOEX/metadata-window/1D at 05:00 accepted", exchange: ExchangeMOEX,
			metadata: SourceMetadata{SessionWindow: "0600-2350"}, bar: mosUnixAt("2025-08-15 05:00"), timeframe: "1D", timezone: "Europe/Moscow", want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cal := RegularCalendarForExchange(tc.exchange, tc.metadata)
			if got := cal.Accepts(tc.bar, tc.timeframe, tc.timezone); got != tc.want {
				t.Fatalf("Accepts(timeframe=%q) = %v, want %v", tc.timeframe, got, tc.want)
			}
		})
	}
}

func TestDefaultReferenceSession(t *testing.T) {
	cases := []struct {
		exchange Exchange
		want     ReferenceSession
	}{
		{ExchangeMOEX, ReferenceSessionRegular},
		{ExchangeBinance, ReferenceSessionAlwaysOpen},
		{ExchangeUnknown, ReferenceSessionAlwaysOpen},
	}

	for _, tc := range cases {
		if got := DefaultReferenceSession(tc.exchange); got != tc.want {
			t.Errorf("DefaultReferenceSession(%v) = %v, want %v", tc.exchange, got, tc.want)
		}
	}
}

func TestResolveReferenceSession(t *testing.T) {
	cases := []struct {
		name     string
		metadata SourceMetadata
		exchange Exchange
		want     ReferenceSession
	}{
		{name: "MOEX/unspecified uses exchange default (regular)", exchange: ExchangeMOEX, want: ReferenceSessionRegular},
		{name: "MOEX/explicit always-open overrides default", exchange: ExchangeMOEX, metadata: SourceMetadata{ReferenceSession: "always-open"}, want: ReferenceSessionAlwaysOpen},
		{name: "MOEX/explicit regular stays regular", exchange: ExchangeMOEX, metadata: SourceMetadata{ReferenceSession: "regular"}, want: ReferenceSessionRegular},
		{name: "Binance/unspecified uses exchange default (always-open)", exchange: ExchangeBinance, want: ReferenceSessionAlwaysOpen},
		{name: "Binance/explicit regular overrides default", exchange: ExchangeBinance, metadata: SourceMetadata{ReferenceSession: "regular"}, want: ReferenceSessionRegular},
		{name: "Unknown/unspecified falls back to always-open", exchange: ExchangeUnknown, want: ReferenceSessionAlwaysOpen},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveReferenceSession(tc.metadata, tc.exchange); got != tc.want {
				t.Fatalf("resolveReferenceSession(%v) = %v, want %v", tc.exchange, got, tc.want)
			}
		})
	}
}
