package ticker

import "testing"

func TestRange(t *testing.T) {
	tests := []struct {
		symbol   string
		expected string
	}{
		{"BTCUSDT", "RANGE:BTCUSDT"},
		{"BINANCE:BTCUSDT", "RANGE:BINANCE:BTCUSDT"},
		{"", "RANGE:"},
	}

	for _, tt := range tests {
		result := Range(tt.symbol)
		if result != tt.expected {
			t.Errorf("Range(%q) = %q, want %q", tt.symbol, result, tt.expected)
		}
	}
}

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
	tests := []struct {
		name     string
		symbol   string
		style    string
		param    float64
		expected string
	}{
		{"basic", "BTCUSDT", "ATR", 14.0, "RENKO:BTCUSDT:ATR:14.00"},
		{"exchange prefixed", "BINANCE:BTCUSDT", "ATR", 14.0, "RENKO:BINANCE:BTCUSDT:ATR:14.00"},
		{"traditional style", "AAPL", "Traditional", 1.0, "RENKO:AAPL:Traditional:1.00"},
		{"empty symbol", "", "ATR", 14.0, "RENKO::ATR:14.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Renko(tt.symbol, tt.style, tt.param)
			if result != tt.expected {
				t.Errorf("Renko() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestKagi(t *testing.T) {
	tests := []struct {
		name     string
		symbol   string
		reversal float64
		expected string
	}{
		{"basic", "BTCUSDT", 3.5, "KAGI:BTCUSDT:3.50"},
		{"exchange prefixed", "BINANCE:BTCUSDT", 3.5, "KAGI:BINANCE:BTCUSDT:3.50"},
		{"empty symbol", "", 3.5, "KAGI::3.50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Kagi(tt.symbol, tt.reversal)
			if result != tt.expected {
				t.Errorf("Kagi() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestLineBreak(t *testing.T) {
	tests := []struct {
		name          string
		symbol        string
		numberOfLines int
		expected      string
	}{
		{"basic", "BTCUSDT", 3, "LINEBREAK:BTCUSDT:3"},
		{"exchange prefixed", "BINANCE:BTCUSDT", 3, "LINEBREAK:BINANCE:BTCUSDT:3"},
		{"single line", "AAPL", 1, "LINEBREAK:AAPL:1"},
		{"empty symbol", "", 3, "LINEBREAK::3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LineBreak(tt.symbol, tt.numberOfLines)
			if result != tt.expected {
				t.Errorf("LineBreak() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseModifiedSymbol(t *testing.T) {
	tests := []struct {
		name             string
		tickerID         string
		expectedBase     string
		expectedModifier ModifierType
		expectedHas      bool
	}{
		{"heikinashi", "HEIKINASHI:BTCUSDT", "BTCUSDT", ModifierHeikinAshi, true},
		{"renko", "RENKO:BTCUSDT:ATR:14.00", "BTCUSDT", ModifierRenko, true},
		{"kagi", "KAGI:ETHUSDT:3.50", "ETHUSDT", ModifierKagi, true},
		{"linebreak", "LINEBREAK:AAPL:3", "AAPL", ModifierLineBreak, true},
		{"pointfigure", "POINTFIG:TSLA", "TSLA", ModifierPointFig, true},
		{"range", "RANGE:BTCUSDT", "BTCUSDT", ModifierRange, true},
		{"plain symbol", "BTCUSDT", "BTCUSDT", "", false},
		{"exchange prefixed", "BINANCE:BTCUSDT", "BINANCE:BTCUSDT", "", false},
		{"unknown prefix", "INVALID:SYMBOL", "INVALID:SYMBOL", "", false},
		{"empty", "", "", "", false},
		{"heikinashi with session metadata", "HEIKINASHI:BTCUSDT|s=regular", "BTCUSDT", ModifierHeikinAshi, true},
		{"renko with adjustment metadata", "RENKO:BTCUSDT:ATR:14.00|a=splits", "BTCUSDT", ModifierRenko, true},
		{"exchange prefix with metadata", "BINANCE:BTCUSDT|s=extended", "BINANCE:BTCUSDT", "", false},
		{"plain symbol with metadata", "BTCUSDT|s=regular", "BTCUSDT", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		})
	}
}

func TestIsModified(t *testing.T) {
	tests := []struct {
		name     string
		tickerID string
		expected bool
	}{
		{"heikinashi", "HEIKINASHI:BTCUSDT", true},
		{"renko", "RENKO:BTCUSDT:ATR:14.00", true},
		{"kagi", "KAGI:ETHUSDT:3.50", true},
		{"linebreak", "LINEBREAK:AAPL:3", true},
		{"pointfigure", "POINTFIG:TSLA", true},
		{"range", "RANGE:BTCUSDT", true},
		{"plain symbol", "BTCUSDT", false},
		{"exchange prefixed", "BINANCE:BTCUSDT", false},
		{"unknown prefix", "INVALID:SYMBOL", false},
		{"empty", "", false},
		{"heikinashi with metadata", "HEIKINASHI:BTCUSDT|s=regular", true},
		{"exchange prefix with metadata", "BINANCE:BTCUSDT|s=extended", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsModified(tt.tickerID)
			if result != tt.expected {
				t.Errorf("IsModified(%q) = %v, want %v", tt.tickerID, result, tt.expected)
			}
		})
	}
}

func TestExtractBaseSymbol(t *testing.T) {
	tests := []struct {
		name     string
		tickerID string
		expected string
	}{
		{"heikinashi", "HEIKINASHI:BTCUSDT", "BTCUSDT"},
		{"renko", "RENKO:BTCUSDT:ATR:14.00", "BTCUSDT"},
		{"kagi", "KAGI:ETHUSDT:3.50", "ETHUSDT"},
		{"linebreak", "LINEBREAK:AAPL:3", "AAPL"},
		{"pointfigure", "POINTFIG:TSLA", "TSLA"},
		{"range", "RANGE:BTCUSDT", "BTCUSDT"},
		{"plain symbol", "BTCUSDT", "BTCUSDT"},
		{"exchange prefixed", "BINANCE:BTCUSDT", "BINANCE:BTCUSDT"},
		{"empty", "", ""},
		{"heikinashi with metadata", "HEIKINASHI:BTCUSDT|s=regular", "BTCUSDT"},
		{"exchange prefix with metadata", "BINANCE:BTCUSDT|s=extended|a=splits", "BINANCE:BTCUSDT"},
		{"plain symbol with metadata", "BTCUSDT|a=dividends", "BTCUSDT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractBaseSymbol(tt.tickerID)
			if result != tt.expected {
				t.Errorf("ExtractBaseSymbol(%q) = %q, want %q", tt.tickerID, result, tt.expected)
			}
		})
	}
}
