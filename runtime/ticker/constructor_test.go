package ticker

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		prefix         string
		symbol         string
		session        SessionType
		adjustment     AdjustmentType
		backAdjustment BackAdjustmentType
		settlement     SettlementType
		expected       string
	}{
		{
			name:     "basic prefix and symbol",
			prefix:   "BINANCE",
			symbol:   "BTCUSDT",
			expected: "BINANCE:BTCUSDT",
		},
		{
			name:     "with session",
			prefix:   "BINANCE",
			symbol:   "BTCUSDT",
			session:  SessionRegular,
			expected: "BINANCE:BTCUSDT|s=regular",
		},
		{
			name:       "with session and adjustment",
			prefix:     "NYSE",
			symbol:     "AAPL",
			session:    SessionExtended,
			adjustment: AdjustmentSplits,
			expected:   "NYSE:AAPL|s=extended|a=splits",
		},
		{
			name:           "all modifiers",
			prefix:         "CME",
			symbol:         "ES1!",
			session:        SessionRegular,
			adjustment:     AdjustmentDividends,
			backAdjustment: BackAdjustmentOn,
			settlement:     SettlementOff,
			expected:       "CME:ES1!|s=regular|a=dividends|ba=on|sc=off",
		},
		{
			name:           "backadjustment and settlement only",
			prefix:         "CME",
			symbol:         "NQ1!",
			backAdjustment: BackAdjustmentOff,
			settlement:     SettlementOn,
			expected:       "CME:NQ1!|ba=off|sc=on",
		},
		{
			name:     "empty prefix",
			symbol:   "BTCUSDT",
			expected: "BTCUSDT",
		},
		{
			name:     "empty prefix with session",
			symbol:   "ETHUSDT",
			session:  SessionExtended,
			expected: "ETHUSDT|s=extended",
		},
		{
			name:     "empty inputs",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := New(tt.prefix, tt.symbol, tt.session, tt.adjustment, tt.backAdjustment, tt.settlement)
			if result != tt.expected {
				t.Errorf("New() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestModify(t *testing.T) {
	tests := []struct {
		name           string
		tickerid       string
		session        SessionType
		adjustment     AdjustmentType
		backAdjustment BackAdjustmentType
		settlement     SettlementType
		expected       string
	}{
		{
			name:     "add session to plain ticker",
			tickerid: "BINANCE:BTCUSDT",
			session:  SessionExtended,
			expected: "BINANCE:BTCUSDT|s=extended",
		},
		{
			name:     "override existing session",
			tickerid: "BINANCE:BTCUSDT|s=regular",
			session:  SessionExtended,
			expected: "BINANCE:BTCUSDT|s=extended",
		},
		{
			name:     "no changes when empty modifiers",
			tickerid: "BINANCE:BTCUSDT",
			expected: "BINANCE:BTCUSDT",
		},
		{
			name:     "preserve existing modifiers when adding new",
			tickerid: "BINANCE:BTCUSDT|a=splits",
			session:  SessionRegular,
			expected: "BINANCE:BTCUSDT|s=regular|a=splits",
		},
		{
			name:       "add adjustment",
			tickerid:   "NYSE:AAPL",
			adjustment: AdjustmentDividends,
			expected:   "NYSE:AAPL|a=dividends",
		},
		{
			name:           "add backadjustment",
			tickerid:       "CME:ES1!",
			backAdjustment: BackAdjustmentOn,
			expected:       "CME:ES1!|ba=on",
		},
		{
			name:       "add settlement",
			tickerid:   "CME:NQ1!",
			settlement: SettlementOff,
			expected:   "CME:NQ1!|sc=off",
		},
		{
			name:           "override all modifiers at once",
			tickerid:       "BINANCE:BTCUSDT|s=regular|a=none|ba=inherit|sc=inherit",
			session:        SessionExtended,
			adjustment:     AdjustmentSplits,
			backAdjustment: BackAdjustmentOff,
			settlement:     SettlementOn,
			expected:       "BINANCE:BTCUSDT|s=extended|a=splits|ba=off|sc=on",
		},
		{
			name:     "idempotent with same session",
			tickerid: "BTCUSDT|s=regular",
			session:  SessionRegular,
			expected: "BTCUSDT|s=regular",
		},
		{
			name:     "symbol only input",
			tickerid: "BTCUSDT",
			session:  SessionRegular,
			expected: "BTCUSDT|s=regular",
		},
		{
			name:     "empty tickerid",
			tickerid: "",
			session:  SessionRegular,
			expected: "|s=regular",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Modify(tt.tickerid, tt.session, tt.adjustment, tt.backAdjustment, tt.settlement)
			if result != tt.expected {
				t.Errorf("Modify() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestInherit(t *testing.T) {
	tests := []struct {
		name     string
		from     string
		symbol   string
		expected string
	}{
		{
			name:     "inherit session and adjustment",
			from:     "BINANCE:BTCUSDT|s=regular|a=splits",
			symbol:   "ETHUSDT",
			expected: "ETHUSDT|s=regular|a=splits",
		},
		{
			name:     "inherit from plain ticker",
			from:     "BINANCE:BTCUSDT",
			symbol:   "ETHUSDT",
			expected: "ETHUSDT",
		},
		{
			name:     "inherit all modifiers",
			from:     "NYSE:AAPL|s=extended|a=dividends|ba=on|sc=off",
			symbol:   "MSFT",
			expected: "MSFT|s=extended|a=dividends|ba=on|sc=off",
		},
		{
			name:     "inherit strips source prefix",
			from:     "BINANCE:BTCUSDT|s=regular",
			symbol:   "BINANCE:ETHUSDT",
			expected: "BINANCE:ETHUSDT|s=regular",
		},
		{
			name:     "empty source",
			from:     "",
			symbol:   "BTCUSDT",
			expected: "BTCUSDT",
		},
		{
			name:     "empty symbol",
			from:     "BINANCE:BTCUSDT|s=regular",
			symbol:   "",
			expected: "|s=regular",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Inherit(tt.from, tt.symbol)
			if result != tt.expected {
				t.Errorf("Inherit() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPointFigure(t *testing.T) {
	tests := []struct {
		name     string
		symbol   string
		source   string
		style    string
		param    float64
		reversal float64
		validate func(t *testing.T, result string)
	}{
		{
			name:     "standard parameters",
			symbol:   "BTCUSDT",
			source:   "close",
			style:    "ATR",
			param:    14.0,
			reversal: 3.0,
			validate: func(t *testing.T, result string) {
				if !strings.HasPrefix(result, "POINTFIG:BTCUSDT:") {
					t.Errorf("missing POINTFIG prefix: %q", result)
				}
				if !strings.Contains(result, "close") {
					t.Errorf("missing source: %q", result)
				}
				if !strings.Contains(result, "ATR") {
					t.Errorf("missing style: %q", result)
				}
				if !strings.Contains(result, "14.00") {
					t.Errorf("missing param: %q", result)
				}
				if !strings.Contains(result, "3.00") {
					t.Errorf("missing reversal: %q", result)
				}
			},
		},
		{
			name:     "traditional style",
			symbol:   "AAPL",
			source:   "hl2",
			style:    "Traditional",
			param:    1.0,
			reversal: 1.0,
			validate: func(t *testing.T, result string) {
				if !strings.HasPrefix(result, "POINTFIG:AAPL:") {
					t.Errorf("missing POINTFIG prefix: %q", result)
				}
				if !strings.Contains(result, "Traditional") {
					t.Errorf("missing style: %q", result)
				}
			},
		},
		{
			name:     "zero parameters",
			symbol:   "ETHUSDT",
			source:   "close",
			style:    "ATR",
			param:    0.0,
			reversal: 0.0,
			validate: func(t *testing.T, result string) {
				if !strings.HasPrefix(result, "POINTFIG:") {
					t.Errorf("missing POINTFIG prefix: %q", result)
				}
				if !strings.Contains(result, "0.00") {
					t.Errorf("missing zero param: %q", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PointFigure(tt.symbol, tt.source, tt.style, tt.param, tt.reversal)
			tt.validate(t, result)
		})
	}
}

func TestStandard(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "heikinashi chart modifier",
			input:    "HEIKINASHI:BTCUSDT",
			expected: "BTCUSDT",
		},
		{
			name:     "renko chart modifier",
			input:    "RENKO:BTCUSDT:ATR:14.00",
			expected: "BTCUSDT",
		},
		{
			name:     "plain symbol",
			input:    "BTCUSDT",
			expected: "BTCUSDT",
		},
		{
			name:     "exchange prefixed symbol",
			input:    "BINANCE:BTCUSDT",
			expected: "BINANCE:BTCUSDT",
		},
		{
			name:     "chart modifier with metadata",
			input:    "HEIKINASHI:BTCUSDT|s=regular",
			expected: "BTCUSDT",
		},
		{
			name:     "metadata only no chart modifier",
			input:    "BINANCE:BTCUSDT|s=extended|a=splits",
			expected: "BINANCE:BTCUSDT",
		},
		{
			name:     "plain symbol with metadata",
			input:    "BTCUSDT|s=regular",
			expected: "BTCUSDT",
		},
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Standard(tt.input)
			if result != tt.expected {
				t.Errorf("Standard(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConstructor_NewModifyRoundTrip(t *testing.T) {
	created := New("NYSE", "AAPL", SessionRegular, AdjustmentSplits, "", "")
	modified := Modify(created, SessionExtended, "", "", "")

	decoded := DecodeTickerID(modified)
	if decoded.Session != SessionExtended {
		t.Errorf("session after Modify = %q, want %q", decoded.Session, SessionExtended)
	}
	if decoded.Adjustment != AdjustmentSplits {
		t.Errorf("adjustment not preserved after Modify = %q, want %q", decoded.Adjustment, AdjustmentSplits)
	}

	standard := Standard(modified)
	if standard != "NYSE:AAPL" {
		t.Errorf("Standard() after New+Modify = %q, want %q", standard, "NYSE:AAPL")
	}
}

func TestConstructor_InheritModifyComposition(t *testing.T) {
	source := New("NYSE", "AAPL", SessionExtended, AdjustmentDividends, BackAdjustmentOn, SettlementOff)
	inherited := Inherit(source, "MSFT")
	modified := Modify(inherited, "", AdjustmentSplits, "", "")

	decoded := DecodeTickerID(modified)
	if decoded.Symbol != "MSFT" {
		t.Errorf("symbol = %q, want MSFT", decoded.Symbol)
	}
	if decoded.Session != SessionExtended {
		t.Errorf("session = %q, want %q (inherited)", decoded.Session, SessionExtended)
	}
	if decoded.Adjustment != AdjustmentSplits {
		t.Errorf("adjustment = %q, want %q (modified)", decoded.Adjustment, AdjustmentSplits)
	}
	if decoded.BackAdjustment != BackAdjustmentOn {
		t.Errorf("backAdjustment = %q, want %q (inherited)", decoded.BackAdjustment, BackAdjustmentOn)
	}
}
