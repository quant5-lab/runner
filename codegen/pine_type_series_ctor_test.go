package codegen

import "testing"

func TestSeriesCtorForType(t *testing.T) {
	tests := []struct {
		varType  string
		expected string
	}{
		{"bool", "series.NewBoolSeries"},
		{"float64", "series.NewSeries"},
		{"string", "series.NewSeries"},
		{"array_series", "series.NewSeries"},
		{"", "series.NewSeries"},
		{"unknown", "series.NewSeries"},
	}

	for _, tt := range tests {
		t.Run(tt.varType, func(t *testing.T) {
			got := SeriesCtorForType(tt.varType)
			if got != tt.expected {
				t.Errorf("SeriesCtorForType(%q) = %q, want %q", tt.varType, got, tt.expected)
			}
		})
	}
}
