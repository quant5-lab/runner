package ticker

import "testing"

func TestHeikinashi(t *testing.T) {
	tests := []struct {
		symbol   string
		expected string
	}{
		{"BTCUSDT", "HEIKINASHI:BTCUSDT"},
		{"BINANCE:BTCUSDT", "HEIKINASHI:BINANCE:BTCUSDT"},
		{"", "HEIKINASHI:"},
	}

	for _, tt := range tests {
		result := Heikinashi(tt.symbol)
		if result != tt.expected {
			t.Errorf("Heikinashi(%q) = %q, want %q", tt.symbol, result, tt.expected)
		}
	}
}

func TestRenko(t *testing.T) {
	result := Renko("BTCUSDT", "ATR", 14.0)
	expected := "RENKO:BTCUSDT:ATR:14.00"
	if result != expected {
		t.Errorf("Renko() = %q, want %q", result, expected)
	}
}

func TestKagi(t *testing.T) {
	result := Kagi("BTCUSDT", 3.5)
	expected := "KAGI:BTCUSDT:3.50"
	if result != expected {
		t.Errorf("Kagi() = %q, want %q", result, expected)
	}
}

func TestLineBreak(t *testing.T) {
	result := LineBreak("BTCUSDT", 3)
	expected := "LINEBREAK:BTCUSDT:3"
	if result != expected {
		t.Errorf("LineBreak() = %q, want %q", result, expected)
	}
}

/* TestParseModifiedSymbol validates modifier parsing across all supported types.
 * Tests correct extraction of base symbol and modifier type from formatted strings.
 */
func TestParseModifiedSymbol(t *testing.T) {
	tests := []struct {
		tickerID         string
		expectedBase     string
		expectedModifier ModifierType
		expectedHas      bool
	}{
		{"HEIKINASHI:BTCUSDT", "BTCUSDT", ModifierHeikinAshi, true},
		{"RENKO:BTCUSDT:ATR:14.00", "BTCUSDT", ModifierRenko, true},
		{"KAGI:ETHUSDT:3.50", "ETHUSDT", ModifierKagi, true},
		{"LINEBREAK:AAPL:3", "AAPL", ModifierLineBreak, true},
		{"POINTFIG:TSLA", "TSLA", ModifierPointFig, true},
		{"BTCUSDT", "BTCUSDT", "", false},
		{"BINANCE:BTCUSDT", "BINANCE:BTCUSDT", "", false},
		{"INVALID:SYMBOL", "INVALID:SYMBOL", "", false},
		{"", "", "", false},
	}

	for _, tt := range tests {
		base, modifier, has := ParseModifiedSymbol(tt.tickerID)
		if base != tt.expectedBase {
			t.Errorf("ParseModifiedSymbol(%q) base = %q, want %q", tt.tickerID, base, tt.expectedBase)
		}
		if modifier != tt.expectedModifier {
			t.Errorf("ParseModifiedSymbol(%q) modifier = %q, want %q", tt.tickerID, modifier, tt.expectedModifier)
		}
		if has != tt.expectedHas {
			t.Errorf("ParseModifiedSymbol(%q) has = %v, want %v", tt.tickerID, has, tt.expectedHas)
		}
	}
}

func TestIsModified(t *testing.T) {
	tests := []struct {
		tickerID string
		expected bool
	}{
		{"HEIKINASHI:BTCUSDT", true},
		{"RENKO:BTCUSDT:ATR:14.00", true},
		{"BTCUSDT", false},
		{"BINANCE:BTCUSDT", false},
	}

	for _, tt := range tests {
		result := IsModified(tt.tickerID)
		if result != tt.expected {
			t.Errorf("IsModified(%q) = %v, want %v", tt.tickerID, result, tt.expected)
		}
	}
}

func TestExtractBaseSymbol(t *testing.T) {
	tests := []struct {
		tickerID string
		expected string
	}{
		{"HEIKINASHI:BTCUSDT", "BTCUSDT"},
		{"RENKO:BTCUSDT:ATR:14.00", "BTCUSDT"},
		{"BTCUSDT", "BTCUSDT"},
		{"BINANCE:BTCUSDT", "BINANCE:BTCUSDT"},
	}

	for _, tt := range tests {
		result := ExtractBaseSymbol(tt.tickerID)
		if result != tt.expected {
			t.Errorf("ExtractBaseSymbol(%q) = %q, want %q", tt.tickerID, result, tt.expected)
		}
	}
}

func TestModifierParserBackwardCompatibility(t *testing.T) {
	parser := NewModifierParser()

	base, modifier, has := parser.Parse("HEIKINASHI:BTCUSDT")
	if base != "BTCUSDT" || modifier != ModifierHeikinAshi || !has {
		t.Error("ModifierParser.Parse() backward compatibility broken")
	}

	if !parser.IsModified("HEIKINASHI:BTCUSDT") {
		t.Error("ModifierParser.IsModified() backward compatibility broken")
	}

	if parser.ExtractBaseSymbol("HEIKINASHI:BTCUSDT") != "BTCUSDT" {
		t.Error("ModifierParser.ExtractBaseSymbol() backward compatibility broken")
	}
}
