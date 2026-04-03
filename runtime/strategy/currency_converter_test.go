package strategy

import (
	"math"
	"testing"
)

func TestCurrencyConverter_ToAccount(t *testing.T) {
	converter := NewCurrencyConverter()

	tests := []struct {
		name  string
		input float64
	}{
		{"positive value", 100.5},
		{"negative value", -50.25},
		{"zero", 0},
		{"large value", 1e10},
		{"small value", 1e-10},
		{"NaN", math.NaN()},
		{"positive infinity", math.Inf(1)},
		{"negative infinity", math.Inf(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ToAccount(tt.input)

			if math.IsNaN(tt.input) {
				if !math.IsNaN(result) {
					t.Errorf("ToAccount(%v) = %v, expected NaN", tt.input, result)
				}
			} else if result != tt.input {
				t.Errorf("ToAccount(%v) = %v, expected %v", tt.input, result, tt.input)
			}
		})
	}
}

func TestCurrencyConverter_ToSymbol(t *testing.T) {
	converter := NewCurrencyConverter()

	tests := []struct {
		name  string
		input float64
	}{
		{"positive value", 200.75},
		{"negative value", -100.5},
		{"zero", 0},
		{"large value", 1e12},
		{"small value", 1e-12},
		{"NaN", math.NaN()},
		{"positive infinity", math.Inf(1)},
		{"negative infinity", math.Inf(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ToSymbol(tt.input)

			if math.IsNaN(tt.input) {
				if !math.IsNaN(result) {
					t.Errorf("ToSymbol(%v) = %v, expected NaN", tt.input, result)
				}
			} else if result != tt.input {
				t.Errorf("ToSymbol(%v) = %v, expected %v", tt.input, result, tt.input)
			}
		})
	}
}
