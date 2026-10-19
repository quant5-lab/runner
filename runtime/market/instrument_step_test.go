package market

import "testing"

func TestInstrumentQtyStep_ExchangeDefaults(t *testing.T) {
	cases := []struct {
		name     string
		exchange Exchange
		want     float64
	}{
		{"MOEX/whole-share lots", ExchangeMOEX, 1},
		{"Binance/per-symbol step unavailable offline", ExchangeBinance, 0},
		{"unknown/defaults to whole-share via switch default", ExchangeUnknown, 1},
		{"empty/treated as unknown exchange", Exchange(""), 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := InstrumentQtyStep(tc.exchange); got != tc.want {
				t.Fatalf("InstrumentQtyStep(%q) = %v, want %v", tc.exchange, got, tc.want)
			}
		})
	}
}

func TestResolveQtyStep_ResolutionPriority(t *testing.T) {
	cases := []struct {
		name     string
		exchange Exchange
		metadata SourceMetadata
		want     float64
	}{
		{"Binance/explicit step overrides zero default", ExchangeBinance, SourceMetadata{QtyStep: 0.001}, 0.001},
		{"MOEX/explicit step overrides whole-share default", ExchangeMOEX, SourceMetadata{QtyStep: 0.5}, 0.5},
		{"unknown/explicit step overrides whole-share default", ExchangeUnknown, SourceMetadata{QtyStep: 10}, 10},
		{"any/tiny positive step is valid and used", ExchangeMOEX, SourceMetadata{QtyStep: 1e-8}, 1e-8},

		{"Binance/zero metadata falls through to Binance default (0)", ExchangeBinance, SourceMetadata{QtyStep: 0}, 0},
		{"MOEX/zero metadata falls through to MOEX default (1)", ExchangeMOEX, SourceMetadata{QtyStep: 0}, 1},
		{"MOEX/negative metadata falls through to MOEX default (1)", ExchangeMOEX, SourceMetadata{QtyStep: -1}, 1},
		{"unknown/empty metadata falls through to switch-default (1)", ExchangeUnknown, SourceMetadata{}, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveQtyStep(tc.exchange, tc.metadata); got != tc.want {
				t.Fatalf("ResolveQtyStep(%q, QtyStep=%v) = %v, want %v",
					tc.exchange, tc.metadata.QtyStep, got, tc.want)
			}
		})
	}
}

func TestResolveQtyStep_PropagatesViaProfile(t *testing.T) {
	cases := []struct {
		name     string
		symbol   string
		metadata SourceMetadata
		want     float64
	}{
		{"SBERP/no metadata QtyStep uses MOEX whole-share default", "SBERP", SourceMetadata{}, 1},
		{"BTCUSDT/no metadata QtyStep uses Binance offline-step default", "BTCUSDT", SourceMetadata{}, 0},
		{"AAPL/no metadata QtyStep uses unknown-exchange default", "AAPL", SourceMetadata{}, 1},

		{"BTCUSDT/declared step used verbatim", "BTCUSDT", SourceMetadata{QtyStep: 0.001}, 0.001},
		{"SBERP/declared step overrides MOEX whole-share", "SBERP", SourceMetadata{QtyStep: 0.5}, 0.5},
		{"AAPL/declared step overrides unknown default", "AAPL", SourceMetadata{QtyStep: 0.01}, 0.01},

		{"any/tiny positive step propagates without loss", "BTCUSDT", SourceMetadata{QtyStep: 1e-8}, 1e-8},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			profile := ResolveProfileWithMetadata(tc.symbol, tc.metadata)
			if profile.QtyStep != tc.want {
				t.Fatalf("ResolveProfileWithMetadata(%q, QtyStep=%v).QtyStep = %v, want %v",
					tc.symbol, tc.metadata.QtyStep, profile.QtyStep, tc.want)
			}
		})
	}
}

func TestResolveQtyStep_PropagatesViaProfileE(t *testing.T) {
	cases := []struct {
		name     string
		symbol   string
		metadata SourceMetadata
		want     float64
	}{
		{"SBERP/no metadata QtyStep uses MOEX default", "SBERP", SourceMetadata{}, 1},
		{"BTCUSDT/no metadata QtyStep uses Binance default", "BTCUSDT", SourceMetadata{}, 0},
		{"BTCUSDT/declared step propagates through error path", "BTCUSDT", SourceMetadata{QtyStep: 0.001}, 0.001},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			profile, err := ResolveProfileWithMetadataE(tc.symbol, tc.metadata)
			if err != nil {
				t.Fatalf("ResolveProfileWithMetadataE: unexpected error: %v", err)
			}
			if profile.QtyStep != tc.want {
				t.Fatalf("ResolveProfileWithMetadataE(%q, QtyStep=%v).QtyStep = %v, want %v",
					tc.symbol, tc.metadata.QtyStep, profile.QtyStep, tc.want)
			}
		})
	}
}
