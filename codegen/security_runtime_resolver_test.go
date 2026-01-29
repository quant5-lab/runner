package codegen

import "testing"

func TestIsRuntimeSymbol(t *testing.T) {
	tests := []struct {
		name     string
		symbol   string
		expected bool
	}{
		{
			name:     "syminfo.tickerid",
			symbol:   "syminfo.tickerid",
			expected: true,
		},
		{
			name:     "syminfo.ticker",
			symbol:   "syminfo.ticker",
			expected: true,
		},
		{
			name:     "literal BTCUSDT",
			symbol:   "BTCUSDT",
			expected: false,
		},
		{
			name:     "literal ETHUSDT",
			symbol:   "ETHUSDT",
			expected: false,
		},
		{
			name:     "empty string",
			symbol:   "",
			expected: false,
		},
		{
			name:     "symbol variable name",
			symbol:   "symbol",
			expected: false,
		},
		{
			name:     "ticker variable name",
			symbol:   "ticker",
			expected: true, /* v4 shorthand for syminfo.ticker */
		},
		{
			name:     "partial match syminfo",
			symbol:   "syminfo",
			expected: false,
		},
		{
			name:     "tickerid shorthand",
			symbol:   "tickerid",
			expected: true, /* v4 shorthand for syminfo.tickerid */
		},
		{
			name:     "case mismatch SYMINFO.TICKERID",
			symbol:   "SYMINFO.TICKERID",
			expected: false,
		},
		{
			name:     "syminfo with space",
			symbol:   "syminfo .tickerid",
			expected: false,
		},
		{
			name:     "syminfo.tickerid with trailing space",
			symbol:   "syminfo.tickerid ",
			expected: false,
		},
		{
			name:     "unicode symbol",
			symbol:   "比特币USDT",
			expected: false,
		},
		{
			name:     "heikinashi wrapping syminfo.tickerid",
			symbol:   "heikinashi(syminfo.tickerid)",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRuntimeSymbol(tt.symbol)
			if result != tt.expected {
				t.Errorf("isRuntimeSymbol(%q) = %v, want %v", tt.symbol, result, tt.expected)
			}
		})
	}
}

func TestIsRuntimeTimeframe(t *testing.T) {
	tests := []struct {
		name      string
		timeframe string
		expected  bool
	}{
		{
			name:      "timeframe.period",
			timeframe: "timeframe.period",
			expected:  true,
		},
		{
			name:      "literal 1D",
			timeframe: "1D",
			expected:  false,
		},
		{
			name:      "literal 1h",
			timeframe: "1h",
			expected:  false,
		},
		{
			name:      "literal 5m",
			timeframe: "5m",
			expected:  false,
		},
		{
			name:      "literal 1W",
			timeframe: "1W",
			expected:  false,
		},
		{
			name:      "literal 1M",
			timeframe: "1M",
			expected:  false,
		},
		{
			name:      "empty string",
			timeframe: "",
			expected:  false,
		},
		{
			name:      "timeframe variable name",
			timeframe: "timeframe",
			expected:  false,
		},
		{
			name:      "period variable name",
			timeframe: "period",
			expected:  false,
		},
		{
			name:      "partial match timeframe",
			timeframe: "timeframe",
			expected:  false,
		},
		{
			name:      "partial match period",
			timeframe: "period",
			expected:  false,
		},
		{
			name:      "case mismatch TIMEFRAME.PERIOD",
			timeframe: "TIMEFRAME.PERIOD",
			expected:  false,
		},
		{
			name:      "timeframe with space",
			timeframe: "timeframe .period",
			expected:  false,
		},
		{
			name:      "timeframe.period with trailing space",
			timeframe: "timeframe.period ",
			expected:  false,
		},
		{
			name:      "timeframe.period with leading space",
			timeframe: " timeframe.period",
			expected:  false,
		},
		{
			name:      "numeric string",
			timeframe: "60",
			expected:  false,
		},
		{
			name:      "quoted literal",
			timeframe: `"1D"`,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRuntimeTimeframe(tt.timeframe)
			if result != tt.expected {
				t.Errorf("isRuntimeTimeframe(%q) = %v, want %v", tt.timeframe, result, tt.expected)
			}
		})
	}
}

func TestRuntimePlaceholder(t *testing.T) {
	result := runtimePlaceholder()
	expected := "%s"

	if result != expected {
		t.Errorf("runtimePlaceholder() = %q, want %q", result, expected)
	}
}

func TestRuntimePlaceholder_Consistency(t *testing.T) {
	/* Placeholder should be consistent across multiple calls */
	first := runtimePlaceholder()
	second := runtimePlaceholder()

	if first != second {
		t.Errorf("runtimePlaceholder() returned inconsistent values: %q != %q", first, second)
	}
}

func TestRuntimeResolution_CombinationScenarios(t *testing.T) {
	tests := []struct {
		name              string
		symbol            string
		timeframe         string
		expectedSymbol    bool
		expectedTimeframe bool
	}{
		{
			name:              "both runtime",
			symbol:            "syminfo.tickerid",
			timeframe:         "timeframe.period",
			expectedSymbol:    true,
			expectedTimeframe: true,
		},
		{
			name:              "runtime symbol, literal timeframe",
			symbol:            "syminfo.tickerid",
			timeframe:         "1D",
			expectedSymbol:    true,
			expectedTimeframe: false,
		},
		{
			name:              "literal symbol, runtime timeframe",
			symbol:            "BTCUSDT",
			timeframe:         "timeframe.period",
			expectedSymbol:    false,
			expectedTimeframe: true,
		},
		{
			name:              "both literal",
			symbol:            "BTCUSDT",
			timeframe:         "1h",
			expectedSymbol:    false,
			expectedTimeframe: false,
		},
		{
			name:              "syminfo.ticker with runtime timeframe",
			symbol:            "syminfo.ticker",
			timeframe:         "timeframe.period",
			expectedSymbol:    true,
			expectedTimeframe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symbolResult := isRuntimeSymbol(tt.symbol)
			timeframeResult := isRuntimeTimeframe(tt.timeframe)

			if symbolResult != tt.expectedSymbol {
				t.Errorf("isRuntimeSymbol(%q) = %v, want %v", tt.symbol, symbolResult, tt.expectedSymbol)
			}

			if timeframeResult != tt.expectedTimeframe {
				t.Errorf("isRuntimeTimeframe(%q) = %v, want %v", tt.timeframe, timeframeResult, tt.expectedTimeframe)
			}
		})
	}
}

func TestRuntimeResolution_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		testFunc func(string) bool
		expected bool
	}{
		{
			name:     "syminfo.tickerid with special chars prefix",
			input:    "$syminfo.tickerid",
			testFunc: isRuntimeSymbol,
			expected: false,
		},
		{
			name:     "timeframe.period with special chars suffix",
			input:    "timeframe.period!",
			testFunc: isRuntimeTimeframe,
			expected: false,
		},
		{
			name:     "syminfo.tickerid substring in longer string",
			input:    "my_syminfo.tickerid_var",
			testFunc: isRuntimeSymbol,
			expected: false,
		},
		{
			name:     "timeframe.period substring in longer string",
			input:    "current_timeframe.period_value",
			testFunc: isRuntimeTimeframe,
			expected: false,
		},
		{
			name:     "empty string symbol test",
			input:    "",
			testFunc: isRuntimeSymbol,
			expected: false,
		},
		{
			name:     "empty string timeframe test",
			input:    "",
			testFunc: isRuntimeTimeframe,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.testFunc(tt.input)
			if result != tt.expected {
				t.Errorf("test function(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
