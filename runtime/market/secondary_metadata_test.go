package market

import "testing"

func TestCompleteSecondaryMetadata_AllResolutionPaths(t *testing.T) {
	moexDefaults := SecondaryContextDefaults{
		Timezone:         "Europe/Moscow",
		ReferenceSession: "regular",
	}

	cases := []struct {
		name           string
		symbol         string
		raw            SourceMetadata
		defaults       SecondaryContextDefaults
		wantTZ         string
		wantRefSession string
	}{
		{
			name:           "both_fields_declared_honoured_unchanged",
			symbol:         "AAPL",
			raw:            SourceMetadata{Timezone: "America/Chicago", ReferenceSession: "always-open"},
			defaults:       moexDefaults,
			wantTZ:         "America/Chicago",
			wantRefSession: "always-open",
		},
		{
			name:           "NYSE_from_metadata_exchange_field_fills_both",
			symbol:         "UNKNOWN_TICKER",
			raw:            SourceMetadata{Exchange: "NYSE"},
			defaults:       moexDefaults,
			wantTZ:         "America/New_York",
			wantRefSession: "regular",
		},
		{
			name:           "NASDAQ_maps_to_NYSE_timezone_and_session",
			symbol:         "NVDA",
			raw:            SourceMetadata{Exchange: "NASDAQ"},
			defaults:       moexDefaults,
			wantTZ:         "America/New_York",
			wantRefSession: "regular",
		},
		{
			name:           "MOEX_from_symbol_name_fills_both",
			symbol:         "SBERP",
			raw:            SourceMetadata{},
			defaults:       SecondaryContextDefaults{Timezone: "UTC", ReferenceSession: "always-open"},
			wantTZ:         "Europe/Moscow",
			wantRefSession: "regular",
		},
		{
			name:           "Binance_always_open_UTC",
			symbol:         "BINANCE:BTCUSDT",
			raw:            SourceMetadata{},
			defaults:       moexDefaults,
			wantTZ:         "UTC",
			wantRefSession: "always-open",
		},
		{
			// AAPL has no curated exchange entry; exchange is Unknown → primary defaults apply.
			// This is the correct behaviour for a same-symbol or opaque-fixture case.
			name:           "unknown_exchange_symbol_inherits_primary_defaults",
			symbol:         "AAPL",
			raw:            SourceMetadata{},
			defaults:       moexDefaults,
			wantTZ:         "Europe/Moscow",
			wantRefSession: "regular",
		},
		{
			name:           "unknown_exchange_empty_defaults_returns_empty_strings",
			symbol:         "EURUSD",
			raw:            SourceMetadata{},
			defaults:       SecondaryContextDefaults{},
			wantTZ:         "",
			wantRefSession: "",
		},
		{
			name:           "timezone_declared_refsession_from_exchange",
			symbol:         "UNKNOWN_TICKER",
			raw:            SourceMetadata{Exchange: "NYSE", Timezone: "America/New_York"},
			defaults:       moexDefaults,
			wantTZ:         "America/New_York",
			wantRefSession: "regular",
		},
		{
			name:           "refsession_declared_timezone_from_exchange",
			symbol:         "UNKNOWN_TICKER",
			raw:            SourceMetadata{Exchange: "NYSE", ReferenceSession: "always-open"},
			defaults:       moexDefaults,
			wantTZ:         "America/New_York",
			wantRefSession: "always-open",
		},
		{
			// metadata.Exchange field takes precedence over exchange inferred from symbol name.
			name:           "metadata_exchange_beats_symbol_inference",
			symbol:         "SBERP",
			raw:            SourceMetadata{Exchange: "NYSE"},
			defaults:       moexDefaults,
			wantTZ:         "America/New_York",
			wantRefSession: "regular",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CompleteSecondaryMetadata(tc.symbol, tc.raw, tc.defaults)
			if got.Timezone != tc.wantTZ {
				t.Errorf("Timezone: got %q, want %q", got.Timezone, tc.wantTZ)
			}
			if got.ReferenceSession != tc.wantRefSession {
				t.Errorf("ReferenceSession: got %q, want %q", got.ReferenceSession, tc.wantRefSession)
			}
		})
	}
}

func TestCompleteSecondaryMetadata_PreservesNonSessionFields(t *testing.T) {
	raw := SourceMetadata{
		Exchange:      "NYSE",
		SessionSource: "custom",
		CalendarID:    "cal-1",
		QtyStep:       0.01,
	}
	got := CompleteSecondaryMetadata("AAPL", raw, SecondaryContextDefaults{})
	if got.SessionSource != "custom" {
		t.Errorf("SessionSource mutated: got %q", got.SessionSource)
	}
	if got.CalendarID != "cal-1" {
		t.Errorf("CalendarID mutated: got %q", got.CalendarID)
	}
	if got.QtyStep != 0.01 {
		t.Errorf("QtyStep mutated: got %v", got.QtyStep)
	}
}
