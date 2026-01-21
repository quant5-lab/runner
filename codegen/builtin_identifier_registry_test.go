package codegen

import "testing"

func TestBuiltinIdentifierRegistry_IsBuiltinSeriesIdentifier(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"close is builtin", "close", true},
		{"open is builtin", "open", true},
		{"high is builtin", "high", true},
		{"low is builtin", "low", true},
		{"volume is builtin", "volume", true},
		{"tr is builtin", "tr", true},
		{"bar_index is builtin", "bar_index", true},
		{"hl2 is builtin", "hl2", true},
		{"hlc3 is builtin", "hlc3", true},
		{"ohlc4 is builtin", "ohlc4", true},
		{"hlcc4 is builtin", "hlcc4", true},
		{"user_var not builtin", "user_var", false},
		{"CLOSE uppercase not builtin", "CLOSE", false},
		{"HL2 uppercase not builtin", "HL2", false},
		{"empty string not builtin", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.IsBuiltinSeriesIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("IsBuiltinSeriesIdentifier(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuiltinIdentifierRegistry_IsDerivedPrice(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"hl2 is derived", "hl2", true},
		{"hlc3 is derived", "hlc3", true},
		{"ohlc4 is derived", "ohlc4", true},
		{"hlcc4 is derived", "hlcc4", true},
		{"close not derived", "close", false},
		{"open not derived", "open", false},
		{"high not derived", "high", false},
		{"low not derived", "low", false},
		{"volume not derived", "volume", false},
		{"tr not derived", "tr", false},
		{"user_var not derived", "user_var", false},
		{"HL2 uppercase not derived", "HL2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.IsDerivedPrice(tt.input)
			if result != tt.expected {
				t.Errorf("IsDerivedPrice(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuiltinIdentifierRegistry_IsOHLCVField(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"close is OHLCV", "close", true},
		{"open is OHLCV", "open", true},
		{"high is OHLCV", "high", true},
		{"low is OHLCV", "low", true},
		{"volume is OHLCV", "volume", true},
		{"tr is OHLCV", "tr", true},
		{"bar_index is OHLCV", "bar_index", true},
		{"hl2 not OHLCV", "hl2", false},
		{"hlc3 not OHLCV", "hlc3", false},
		{"ohlc4 not OHLCV", "ohlc4", false},
		{"hlcc4 not OHLCV", "hlcc4", false},
		{"user_var not OHLCV", "user_var", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.IsOHLCVField(tt.input)
			if result != tt.expected {
				t.Errorf("IsOHLCVField(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuiltinIdentifierRegistry_MutualExclusivity(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()

	allBuiltins := []string{"close", "open", "high", "low", "volume", "tr", "bar_index", "hl2", "hlc3", "ohlc4", "hlcc4"}

	for _, builtin := range allBuiltins {
		t.Run(builtin, func(t *testing.T) {
			isBuiltin := registry.IsBuiltinSeriesIdentifier(builtin)
			isDerived := registry.IsDerivedPrice(builtin)
			isOHLCV := registry.IsOHLCVField(builtin)

			if !isBuiltin {
				t.Errorf("%s should be recognized as builtin", builtin)
			}

			if isDerived && isOHLCV {
				t.Errorf("%s cannot be both derived price and OHLCV field", builtin)
			}

			if !isDerived && !isOHLCV {
				t.Errorf("%s should be either derived price or OHLCV field", builtin)
			}
		})
	}
}
