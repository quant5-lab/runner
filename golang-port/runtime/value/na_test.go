package value

import (
	"math"
	"testing"
)

func TestIsNa(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  bool
	}{
		{"NaN is NA", math.NaN(), true},
		{"Zero is not NA", 0.0, false},
		{"Positive is not NA", 42.5, false},
		{"Negative is not NA", -10.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNa(tt.value)
			if got != tt.want {
				t.Errorf("IsNa(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestNz(t *testing.T) {
	tests := []struct {
		name        string
		value       float64
		replacement float64
		want        float64
	}{
		{"NaN replaced with 0", math.NaN(), 0.0, 0.0},
		{"NaN replaced with 100", math.NaN(), 100.0, 100.0},
		{"Valid value unchanged", 42.5, 0.0, 42.5},
		{"Zero unchanged", 0.0, 100.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Nz(tt.value, tt.replacement)
			if math.IsNaN(got) || got != tt.want {
				t.Errorf("Nz(%v, %v) = %v, want %v", tt.value, tt.replacement, got, tt.want)
			}
		})
	}
}

func TestFixnan(t *testing.T) {
	tests := []struct {
		name   string
		source []float64
		want   []float64
	}{
		{
			name:   "All NaN returns all NaN",
			source: []float64{math.NaN(), math.NaN(), math.NaN()},
			want:   []float64{math.NaN(), math.NaN(), math.NaN()},
		},
		{
			name:   "First value NaN filled from second",
			source: []float64{math.NaN(), 100.0, 110.0},
			want:   []float64{100.0, 100.0, 110.0},
		},
		{
			name:   "Middle NaN filled with last valid",
			source: []float64{100.0, math.NaN(), 110.0},
			want:   []float64{100.0, 110.0, 110.0},
		},
		{
			name:   "Last NaN keeps NaN",
			source: []float64{100.0, 110.0, math.NaN()},
			want:   []float64{100.0, 110.0, math.NaN()},
		},
		{
			name:   "No NaN returns unchanged",
			source: []float64{100.0, 105.0, 110.0},
			want:   []float64{100.0, 105.0, 110.0},
		},
		{
			name:   "Empty slice",
			source: []float64{},
			want:   []float64{},
		},
		{
			name:   "Alternating NaN pattern",
			source: []float64{100.0, math.NaN(), 110.0, math.NaN(), 120.0},
			want:   []float64{100.0, 110.0, 110.0, 120.0, 120.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Fixnan(tt.source)

			if len(got) != len(tt.want) {
				t.Fatalf("Fixnan() length = %d, want %d", len(got), len(tt.want))
			}

			for i := range got {
				bothNaN := math.IsNaN(got[i]) && math.IsNaN(tt.want[i])
				if !bothNaN && got[i] != tt.want[i] {
					t.Errorf("Fixnan()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
