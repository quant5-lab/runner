package codegen

import (
	"math"
	"testing"
)

func TestLiteralFormatter_FormatFloat(t *testing.T) {
	formatter := NewLiteralFormatter()

	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{
			name:     "small_precision_0.001",
			input:    0.001,
			expected: "0.001",
		},
		{
			name:     "small_precision_0.00005",
			input:    0.00005,
			expected: "5e-05",
		},
		{
			name:     "decimal_711.6",
			input:    711.6,
			expected: "711.6",
		},
		{
			name:     "decimal_0.35",
			input:    0.35,
			expected: "0.35",
		},
		{
			name:     "large_2000000",
			input:    2000000,
			expected: "2e+06",
		},
		{
			name:     "large_600000",
			input:    600000,
			expected: "600000",
		},
		{
			name:     "integer_42",
			input:    42.0,
			expected: "42",
		},
		{
			name:     "zero",
			input:    0.0,
			expected: "0",
		},
		{
			name:     "negative_-2.5",
			input:    -2.5,
			expected: "-2.5",
		},
		{
			name:     "negative_small_-0.001",
			input:    -0.001,
			expected: "-0.001",
		},
		{
			name:     "bb_strategy_critical_0.001",
			input:    0.001,
			expected: "0.001",
		},
		{
			name:     "bb_strategy_stdev_0.35",
			input:    0.35,
			expected: "0.35",
		},
		{
			name:     "bb_strategy_price_635.4",
			input:    635.4,
			expected: "635.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatFloat(tt.input)
			if result != tt.expected {
				t.Errorf("FormatFloat(%f) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLiteralFormatter_FormatFloat_SpecialValues(t *testing.T) {
	formatter := NewLiteralFormatter()

	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{
			name:     "positive_infinity",
			input:    math.Inf(1),
			expected: "+Inf",
		},
		{
			name:     "negative_infinity",
			input:    math.Inf(-1),
			expected: "-Inf",
		},
		{
			name:     "NaN",
			input:    math.NaN(),
			expected: "NaN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatFloat(tt.input)
			if result != tt.expected {
				t.Errorf("FormatFloat(%f) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLiteralFormatter_FormatString(t *testing.T) {
	formatter := NewLiteralFormatter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple_string",
			input:    "hello",
			expected: `"hello"`,
		},
		{
			name:     "empty_string",
			input:    "",
			expected: `""`,
		},
		{
			name:     "string_with_quotes",
			input:    `say "hi"`,
			expected: `"say \"hi\""`,
		},
		{
			name:     "string_with_newline",
			input:    "line1\nline2",
			expected: `"line1\nline2"`,
		},
		{
			name:     "ticker_symbol",
			input:    "CNRU",
			expected: `"CNRU"`,
		},
		{
			name:     "timeframe_1h",
			input:    "1h",
			expected: `"1h"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatString(tt.input)
			if result != tt.expected {
				t.Errorf("FormatString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLiteralFormatter_FormatBool(t *testing.T) {
	formatter := NewLiteralFormatter()

	tests := []struct {
		name     string
		input    bool
		expected string
	}{
		{
			name:     "true",
			input:    true,
			expected: "true",
		},
		{
			name:     "false",
			input:    false,
			expected: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatBool(tt.input)
			if result != tt.expected {
				t.Errorf("FormatBool(%t) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLiteralFormatter_FormatGeneric(t *testing.T) {
	formatter := NewLiteralFormatter()

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "int",
			input:    42,
			expected: "42",
		},
		{
			name:     "float",
			input:    3.14,
			expected: "3.14",
		},
		{
			name:     "string",
			input:    "test",
			expected: `"test"`,
		},
		{
			name:     "bool_true",
			input:    true,
			expected: "true",
		},
		{
			name:     "bool_false",
			input:    false,
			expected: "false",
		},
		{
			name:     "null",
			input:    nil,
			expected: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := formatter.FormatGeneric(tt.input)
			if err != nil {
				t.Fatalf("FormatGeneric(%v) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("FormatGeneric(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLiteralFormatter_RegressionBB8Strategy(t *testing.T) {
	formatter := NewLiteralFormatter()

	// Critical regression test: BB8 strategy exit condition
	// Pine: bb_1d_low_range - 0.001
	// Bug: %.2f formatted 0.001 as 0.00, breaking exit logic
	criticalValue := 0.001
	result := formatter.FormatFloat(criticalValue)

	if result != "0.001" {
		t.Errorf("REGRESSION: FormatFloat(0.001) = %q, must be exactly \"0.001\" for BB8 exit logic", result)
	}

	// Verify subtraction scenario
	lowRange := 620.2
	offset := 0.001
	expectedExit := lowRange - offset // 620.199

	formattedOffset := formatter.FormatFloat(offset)
	if formattedOffset != "0.001" {
		t.Errorf("BB8 exit offset incorrectly formatted: %q != \"0.001\"", formattedOffset)
	}

	// Verify the actual exit value would be different
	formattedLowRange := formatter.FormatFloat(lowRange)
	formattedExpectedExit := formatter.FormatFloat(expectedExit)

	if formattedLowRange == formattedExpectedExit {
		t.Errorf("BB8 exit condition would fail: lowRange(%s) == expectedExit(%s)", formattedLowRange, formattedExpectedExit)
	}
}
