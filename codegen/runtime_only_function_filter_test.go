package codegen

import "testing"

/* Validates exact string matching for registered runtime-only functions */
func TestRuntimeOnlyFunctionFilter_KnownFunctions(t *testing.T) {
	filter := NewRuntimeOnlyFunctionFilter()

	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"fixnan function", "fixnan", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := filter.IsRuntimeOnly(tt.funcName); result != tt.expected {
				t.Errorf("IsRuntimeOnly(%q) = %v, expected %v", tt.funcName, result, tt.expected)
			}
		})
	}
}

/* Validates rejection of regular TA and strategy functions */
func TestRuntimeOnlyFunctionFilter_RegularFunctions(t *testing.T) {
	filter := NewRuntimeOnlyFunctionFilter()

	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"ta.sma", "ta.sma", false},
		{"ta.ema", "ta.ema", false},
		{"ta.rsi", "ta.rsi", false},
		{"ta.macd", "ta.macd", false},
		{"ta.bb", "ta.bb", false},
		{"ta.pivothigh", "ta.pivothigh", false},
		{"ta.pivotlow", "ta.pivotlow", false},
		{"pivothigh", "pivothigh", false},
		{"pivotlow", "pivotlow", false},
		{"sma non-namespaced", "sma", false},
		{"ema non-namespaced", "ema", false},
		{"plot function", "plot", false},
		{"plotshape function", "plotshape", false},
		{"strategy.entry", "strategy.entry", false},
		{"strategy.exit", "strategy.exit", false},
		{"strategy.close", "strategy.close", false},
		{"math.abs", "math.abs", false},
		{"math.max", "math.max", false},
		{"user function", "myCustomFunction", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := filter.IsRuntimeOnly(tt.funcName); result != tt.expected {
				t.Errorf("IsRuntimeOnly(%q) = %v, expected %v", tt.funcName, result, tt.expected)
			}
		})
	}
}

/* Validates boundary conditions and edge cases */
func TestRuntimeOnlyFunctionFilter_BoundaryConditions(t *testing.T) {
	filter := NewRuntimeOnlyFunctionFilter()

	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"empty string", "", false},
		{"single character", "f", false},
		{"single dot", ".", false},
		{"ta namespace only", "ta.", false},
		{"very long name", "thisIsAVeryLongFunctionNameThatDoesNotExist", false},
		{"numeric only", "12345", false},
		{"special characters", "@#$%", false},
		{"unicode characters", "功能", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := filter.IsRuntimeOnly(tt.funcName); result != tt.expected {
				t.Errorf("IsRuntimeOnly(%q) = %v, expected %v", tt.funcName, result, tt.expected)
			}
		})
	}
}

/* Validates case sensitivity requirements */
func TestRuntimeOnlyFunctionFilter_CaseSensitivity(t *testing.T) {
	filter := NewRuntimeOnlyFunctionFilter()

	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"uppercase PIVOTHIGH", "PIVOTHIGH", false},
		{"uppercase PIVOTLOW", "PIVOTLOW", false},
		{"uppercase FIXNAN", "FIXNAN", false},
		{"mixed case PivotHigh", "PivotHigh", false},
		{"mixed case PivotLow", "PivotLow", false},
		{"mixed case FixNan", "FixNan", false},
		{"uppercase namespace TA.pivothigh", "TA.pivothigh", false},
		{"mixed namespace Ta.pivothigh", "Ta.pivothigh", false},
		{"correct lowercase pivothigh (codegen now)", "pivothigh", false},
		{"correct lowercase ta.pivothigh (codegen now)", "ta.pivothigh", false},
		{"correct lowercase fixnan", "fixnan", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := filter.IsRuntimeOnly(tt.funcName); result != tt.expected {
				t.Errorf("IsRuntimeOnly(%q) = %v, expected %v", tt.funcName, result, tt.expected)
			}
		})
	}
}

/* Validates rejection of near-miss partial matches */
func TestRuntimeOnlyFunctionFilter_PartialMatches(t *testing.T) {
	filter := NewRuntimeOnlyFunctionFilter()

	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"prefix pivot only", "pivot", false},
		{"prefix ta.pivot", "ta.pivot", false},
		{"suffix pivothighlow", "pivothighlow", false},
		{"suffix myfixnan", "myfixnan", false},
		{"prefix ta.pivothighest", "ta.pivothighest", false},
		{"suffix pivotlowest", "pivotlowest", false},
		{"substring mypivothigh", "mypivothigh", false},
		{"substring fixnan_custom", "fixnan_custom", false},
		{"suffix variation pivothigh2", "pivothigh2", false},
		{"prefix variation custom_pivothigh", "custom_pivothigh", false},
		{"similar fixnan_v2", "fixnan_v2", false},
		{"embedded ta.pivothigh.custom", "ta.pivothigh.custom", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := filter.IsRuntimeOnly(tt.funcName); result != tt.expected {
				t.Errorf("IsRuntimeOnly(%q) = %v, expected %v", tt.funcName, result, tt.expected)
			}
		})
	}
}

/* Validates whitespace handling */
func TestRuntimeOnlyFunctionFilter_WhitespaceHandling(t *testing.T) {
	filter := NewRuntimeOnlyFunctionFilter()

	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"leading space", " pivothigh", false},
		{"trailing space", "pivothigh ", false},
		{"both spaces", " pivothigh ", false},
		{"embedded space", "pivot high", false},
		{"tab character", "pivothigh\t", false},
		{"newline character", "pivothigh\n", false},
		{"multiple spaces", "  pivothigh  ", false},
		{"space in namespace", "ta. pivothigh", false},
		{"space before namespace", " ta.pivothigh", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := filter.IsRuntimeOnly(tt.funcName); result != tt.expected {
				t.Errorf("IsRuntimeOnly(%q) = %v, expected %v", tt.funcName, result, tt.expected)
			}
		})
	}
}

/* Validates idempotency across multiple invocations */
func TestRuntimeOnlyFunctionFilter_Idempotency(t *testing.T) {
	filter := NewRuntimeOnlyFunctionFilter()

	testCases := []string{
		"ta.pivothigh",
		"pivotlow",
		"fixnan",
		"ta.sma",
		"unknown",
		"",
	}

	for _, funcName := range testCases {
		t.Run(funcName, func(t *testing.T) {
			firstResult := filter.IsRuntimeOnly(funcName)
			secondResult := filter.IsRuntimeOnly(funcName)
			thirdResult := filter.IsRuntimeOnly(funcName)

			if firstResult != secondResult || secondResult != thirdResult {
				t.Errorf("IsRuntimeOnly(%q) not idempotent: got %v, %v, %v",
					funcName, firstResult, secondResult, thirdResult)
			}
		})
	}
}

/* Validates constructor creates valid filter instance */
func TestRuntimeOnlyFunctionFilter_Constructor(t *testing.T) {
	filter := NewRuntimeOnlyFunctionFilter()

	if filter == nil {
		t.Fatal("NewRuntimeOnlyFunctionFilter() returned nil")
	}

	if filter.runtimeOnlyFunctions == nil {
		t.Fatal("runtimeOnlyFunctions map is nil")
	}

	expectedCount := 1 // Only fixnan is runtime-only now (pivots have codegen)
	actualCount := len(filter.runtimeOnlyFunctions)
	if actualCount != expectedCount {
		t.Errorf("Expected %d runtime-only functions, got %d", expectedCount, actualCount)
	}
}
