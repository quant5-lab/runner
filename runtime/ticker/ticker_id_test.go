package ticker

import (
	"strings"
	"testing"
)

func TestTickerID_Encode(t *testing.T) {
	tests := []struct {
		name     string
		id       TickerID
		expected string
	}{
		{
			name:     "prefix and symbol only",
			id:       TickerID{Prefix: "BINANCE", Symbol: "BTCUSDT"},
			expected: "BINANCE:BTCUSDT",
		},
		{
			name:     "symbol only",
			id:       TickerID{Symbol: "BTCUSDT"},
			expected: "BTCUSDT",
		},
		{
			name:     "with session",
			id:       TickerID{Prefix: "BINANCE", Symbol: "BTCUSDT", Session: SessionRegular},
			expected: "BINANCE:BTCUSDT|s=regular",
		},
		{
			name:     "with all modifiers",
			id:       TickerID{Prefix: "BINANCE", Symbol: "BTCUSDT", Session: SessionExtended, Adjustment: AdjustmentSplits, BackAdjustment: BackAdjustmentOn, Settlement: SettlementOff},
			expected: "BINANCE:BTCUSDT|s=extended|a=splits|ba=on|sc=off",
		},
		{
			name:     "symbol with partial modifiers",
			id:       TickerID{Symbol: "ETHUSDT", Adjustment: AdjustmentDividends},
			expected: "ETHUSDT|a=dividends",
		},
		{
			name:     "backadjustment and settlement only",
			id:       TickerID{Prefix: "CME", Symbol: "ES1!", BackAdjustment: BackAdjustmentOff, Settlement: SettlementOn},
			expected: "CME:ES1!|ba=off|sc=on",
		},
		{
			name:     "empty ticker",
			id:       TickerID{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.id.Encode()
			if result != tt.expected {
				t.Errorf("Encode() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTickerID_Encode_ModifierOrdering(t *testing.T) {
	id := TickerID{
		Symbol:         "BTCUSDT",
		Settlement:     SettlementOn,
		Session:        SessionRegular,
		BackAdjustment: BackAdjustmentOff,
		Adjustment:     AdjustmentSplits,
	}
	result := id.Encode()

	sIdx := strings.Index(result, "s=")
	aIdx := strings.Index(result, "a=")
	baIdx := strings.Index(result, "ba=")
	scIdx := strings.Index(result, "sc=")

	if sIdx >= aIdx || aIdx >= baIdx || baIdx >= scIdx {
		t.Errorf("Modifier ordering violated: s=%d a=%d ba=%d sc=%d in %q", sIdx, aIdx, baIdx, scIdx, result)
	}
}

func TestDecodeTickerID(t *testing.T) {
	tests := []struct {
		name     string
		encoded  string
		expected TickerID
	}{
		{
			name:     "prefix and symbol",
			encoded:  "BINANCE:BTCUSDT",
			expected: TickerID{Prefix: "BINANCE", Symbol: "BTCUSDT"},
		},
		{
			name:     "symbol only",
			encoded:  "BTCUSDT",
			expected: TickerID{Symbol: "BTCUSDT"},
		},
		{
			name:     "with session modifier",
			encoded:  "BINANCE:BTCUSDT|s=regular",
			expected: TickerID{Prefix: "BINANCE", Symbol: "BTCUSDT", Session: SessionRegular},
		},
		{
			name:     "with all modifiers",
			encoded:  "BINANCE:BTCUSDT|s=extended|a=splits|ba=on|sc=off",
			expected: TickerID{Prefix: "BINANCE", Symbol: "BTCUSDT", Session: SessionExtended, Adjustment: AdjustmentSplits, BackAdjustment: BackAdjustmentOn, Settlement: SettlementOff},
		},
		{
			name:     "chart modifier prefix preserved as prefix",
			encoded:  "HEIKINASHI:BTCUSDT|s=regular",
			expected: TickerID{Prefix: "HEIKINASHI", Symbol: "BTCUSDT", Session: SessionRegular},
		},
		{
			name:     "unknown modifier keys ignored",
			encoded:  "BTCUSDT|x=unknown|s=regular|z=ignored",
			expected: TickerID{Symbol: "BTCUSDT", Session: SessionRegular},
		},
		{
			name:     "malformed modifier without equals",
			encoded:  "BTCUSDT|noequals|s=regular",
			expected: TickerID{Symbol: "BTCUSDT", Session: SessionRegular},
		},
		{
			name:     "duplicate modifier keys last wins",
			encoded:  "BTCUSDT|s=regular|s=extended",
			expected: TickerID{Symbol: "BTCUSDT", Session: SessionExtended},
		},
		{
			name:     "empty string",
			encoded:  "",
			expected: TickerID{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DecodeTickerID(tt.encoded)
			if result != tt.expected {
				t.Errorf("DecodeTickerID(%q) = %+v, want %+v", tt.encoded, result, tt.expected)
			}
		})
	}
}

func TestTickerID_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		id   TickerID
	}{
		{
			name: "prefix and symbol",
			id:   TickerID{Prefix: "BINANCE", Symbol: "BTCUSDT"},
		},
		{
			name: "session and adjustment",
			id:   TickerID{Symbol: "ETHUSDT", Session: SessionRegular, Adjustment: AdjustmentSplits},
		},
		{
			name: "all fields populated",
			id:   TickerID{Prefix: "NYSE", Symbol: "AAPL", Session: SessionExtended, Adjustment: AdjustmentDividends, BackAdjustment: BackAdjustmentInherit, Settlement: SettlementOn},
		},
		{
			name: "symbol only",
			id:   TickerID{Symbol: "TSLA"},
		},
		{
			name: "empty",
			id:   TickerID{},
		},
		{
			name: "settlement only",
			id:   TickerID{Prefix: "CME", Symbol: "ES1!", Settlement: SettlementOff},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := tt.id.Encode()
			decoded := DecodeTickerID(encoded)
			reEncoded := decoded.Encode()
			if decoded != tt.id {
				t.Errorf("Round-trip struct mismatch: %+v → %q → %+v", tt.id, encoded, decoded)
			}
			if reEncoded != encoded {
				t.Errorf("Round-trip string mismatch: %q → %+v → %q", encoded, decoded, reEncoded)
			}
		})
	}
}

func TestStripModifiers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "all modifiers stripped",
			input:    "BINANCE:BTCUSDT|s=regular|a=splits",
			expected: "BINANCE:BTCUSDT",
		},
		{
			name:     "chart modifier with metadata",
			input:    "HEIKINASHI:BTCUSDT|s=extended",
			expected: "HEIKINASHI:BTCUSDT",
		},
		{
			name:     "no modifiers",
			input:    "BTCUSDT",
			expected: "BTCUSDT",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single modifier",
			input:    "ETHUSDT|a=dividends",
			expected: "ETHUSDT",
		},
		{
			name:     "prefix with single modifier",
			input:    "CME:ES1!|sc=on",
			expected: "CME:ES1!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripModifiers(tt.input)
			if result != tt.expected {
				t.Errorf("StripModifiers(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDecodeTickerID_AllModifierValues(t *testing.T) {
	sessions := []SessionType{SessionRegular, SessionExtended}
	adjustments := []AdjustmentType{AdjustmentNone, AdjustmentSplits, AdjustmentDividends}
	backAdj := []BackAdjustmentType{BackAdjustmentInherit, BackAdjustmentOn, BackAdjustmentOff}
	settlements := []SettlementType{SettlementInherit, SettlementOn, SettlementOff}

	for _, s := range sessions {
		for _, a := range adjustments {
			for _, ba := range backAdj {
				for _, sc := range settlements {
					id := TickerID{
						Prefix:         "TEST",
						Symbol:         "SYM",
						Session:        s,
						Adjustment:     a,
						BackAdjustment: ba,
						Settlement:     sc,
					}
					encoded := id.Encode()
					decoded := DecodeTickerID(encoded)
					if decoded != id {
						t.Errorf("Failed for s=%s a=%s ba=%s sc=%s: encoded=%q decoded=%+v",
							s, a, ba, sc, encoded, decoded)
					}
				}
			}
		}
	}
}
