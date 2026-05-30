package market

import "testing"

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
		{name: "blank MOEX remains always open", symbol: "SBERP", session: "", wantSession: ReferenceSessionAlwaysOpen, wantCalendar: AlwaysOpenCalendar{}},
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
