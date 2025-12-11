package value_test

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/runtime/value"
)

func TestValuewhen_BasicOccurrences(t *testing.T) {
	tests := []struct {
		name       string
		condition  []bool
		source     []float64
		occurrence int
		want       []float64
	}{
		{
			name:       "occurrence 0 - most recent match",
			condition:  []bool{false, true, false, true, false, true},
			source:     []float64{10, 20, 30, 40, 50, 60},
			occurrence: 0,
			want:       []float64{math.NaN(), 20, 20, 40, 40, 60},
		},
		{
			name:       "occurrence 1 - second most recent",
			condition:  []bool{false, true, false, true, false, true},
			source:     []float64{10, 20, 30, 40, 50, 60},
			occurrence: 1,
			want:       []float64{math.NaN(), math.NaN(), math.NaN(), 20, 20, 40},
		},
		{
			name:       "occurrence 2 - third most recent",
			condition:  []bool{false, true, false, true, false, true},
			source:     []float64{10, 20, 30, 40, 50, 60},
			occurrence: 2,
			want:       []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), 20},
		},
		{
			name:       "high occurrence value",
			condition:  []bool{true, false, false, false, false, true, false, true},
			source:     []float64{100, 200, 300, 400, 500, 600, 700, 800},
			occurrence: 2,
			want:       []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := value.Valuewhen(tt.condition, tt.source, tt.occurrence)
			assertFloatSlicesEqual(t, got, tt.want)
		})
	}
}

func TestValuewhen_ConditionPatterns(t *testing.T) {
	tests := []struct {
		name       string
		condition  []bool
		source     []float64
		occurrence int
		want       []float64
	}{
		{
			name:       "no condition ever true",
			condition:  []bool{false, false, false, false},
			source:     []float64{10, 20, 30, 40},
			occurrence: 0,
			want:       []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN()},
		},
		{
			name:       "all conditions true",
			condition:  []bool{true, true, true, true},
			source:     []float64{10, 20, 30, 40},
			occurrence: 0,
			want:       []float64{10, 20, 30, 40},
		},
		{
			name:       "single condition true at start",
			condition:  []bool{true, false, false, false},
			source:     []float64{100, 200, 300, 400},
			occurrence: 0,
			want:       []float64{100, 100, 100, 100},
		},
		{
			name:       "single condition true at end",
			condition:  []bool{false, false, false, true},
			source:     []float64{10, 20, 30, 40},
			occurrence: 0,
			want:       []float64{math.NaN(), math.NaN(), math.NaN(), 40},
		},
		{
			name:       "sparse conditions",
			condition:  []bool{true, false, false, false, false, false, true, false, false, true},
			source:     []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			occurrence: 0,
			want:       []float64{1, 1, 1, 1, 1, 1, 7, 7, 7, 10},
		},
		{
			name:       "consecutive conditions",
			condition:  []bool{false, true, true, true, false, false},
			source:     []float64{10, 20, 30, 40, 50, 60},
			occurrence: 0,
			want:       []float64{math.NaN(), 20, 30, 40, 40, 40},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := value.Valuewhen(tt.condition, tt.source, tt.occurrence)
			assertFloatSlicesEqual(t, got, tt.want)
		})
	}
}

func TestValuewhen_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		condition  []bool
		source     []float64
		occurrence int
		want       []float64
	}{
		{
			name:       "empty arrays",
			condition:  []bool{},
			source:     []float64{},
			occurrence: 0,
			want:       []float64{},
		},
		{
			name:       "single bar - condition false",
			condition:  []bool{false},
			source:     []float64{42},
			occurrence: 0,
			want:       []float64{math.NaN()},
		},
		{
			name:       "single bar - condition true",
			condition:  []bool{true},
			source:     []float64{42},
			occurrence: 0,
			want:       []float64{42},
		},
		{
			name:       "occurrence exceeds available matches",
			condition:  []bool{true, false, true, false},
			source:     []float64{10, 20, 30, 40},
			occurrence: 5,
			want:       []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN()},
		},
		{
			name:       "occurrence exactly at match count boundary",
			condition:  []bool{true, false, true, false, true},
			source:     []float64{10, 20, 30, 40, 50},
			occurrence: 2,
			want:       []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), 10},
		},
		{
			name:       "negative source values",
			condition:  []bool{false, true, false, true},
			source:     []float64{-10, -20, -30, -40},
			occurrence: 0,
			want:       []float64{math.NaN(), -20, -20, -40},
		},
		{
			name:       "zero source values",
			condition:  []bool{true, false, true, false},
			source:     []float64{0, 1, 0, 3},
			occurrence: 0,
			want:       []float64{0, 0, 0, 0},
		},
		{
			name:       "floating point precision values",
			condition:  []bool{true, false, true},
			source:     []float64{1.23456789, 2.34567890, 3.45678901},
			occurrence: 0,
			want:       []float64{1.23456789, 1.23456789, 3.45678901},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := value.Valuewhen(tt.condition, tt.source, tt.occurrence)
			assertFloatSlicesEqual(t, got, tt.want)
		})
	}
}

func TestValuewhen_WarmupBehavior(t *testing.T) {
	tests := []struct {
		name       string
		condition  []bool
		source     []float64
		occurrence int
		want       []float64
	}{
		{
			name:       "warmup period - no historical data",
			condition:  []bool{false, false, true, false},
			source:     []float64{10, 20, 30, 40},
			occurrence: 0,
			want:       []float64{math.NaN(), math.NaN(), 30, 30},
		},
		{
			name:       "occurrence 1 warmup - needs two matches",
			condition:  []bool{false, true, false, false, true},
			source:     []float64{10, 20, 30, 40, 50},
			occurrence: 1,
			want:       []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), 20},
		},
		{
			name:       "progressive warmup with occurrence 0",
			condition:  []bool{true, false, false, true, false, true},
			source:     []float64{1, 2, 3, 4, 5, 6},
			occurrence: 0,
			want:       []float64{1, 1, 1, 4, 4, 6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := value.Valuewhen(tt.condition, tt.source, tt.occurrence)
			assertFloatSlicesEqual(t, got, tt.want)
		})
	}
}

func TestValuewhen_SourceValueTracking(t *testing.T) {
	tests := []struct {
		name       string
		condition  []bool
		source     []float64
		occurrence int
		want       []float64
	}{
		{
			name:       "tracks correct source value at condition match",
			condition:  []bool{false, true, false, false, true, false},
			source:     []float64{100, 200, 300, 400, 500, 600},
			occurrence: 0,
			want:       []float64{math.NaN(), 200, 200, 200, 500, 500},
		},
		{
			name:       "source changes between condition matches",
			condition:  []bool{true, false, false, true, false, false},
			source:     []float64{10, 20, 30, 40, 50, 60},
			occurrence: 0,
			want:       []float64{10, 10, 10, 40, 40, 40},
		},
		{
			name:       "occurrence 1 tracks second-to-last match",
			condition:  []bool{true, true, false, true, false, false},
			source:     []float64{11, 22, 33, 44, 55, 66},
			occurrence: 1,
			want:       []float64{math.NaN(), 11, 11, 22, 22, 22},
		},
		{
			name:       "different source values at each match",
			condition:  []bool{true, false, true, false, true, false, true},
			source:     []float64{1, 2, 3, 4, 5, 6, 7},
			occurrence: 0,
			want:       []float64{1, 1, 3, 3, 5, 5, 7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := value.Valuewhen(tt.condition, tt.source, tt.occurrence)
			assertFloatSlicesEqual(t, got, tt.want)
		})
	}
}

func TestValuewhen_ArraySizeMismatch(t *testing.T) {
	tests := []struct {
		name       string
		condition  []bool
		source     []float64
		occurrence int
	}{
		{
			name:       "condition longer than source",
			condition:  []bool{true, false, true, false, true},
			source:     []float64{10, 20, 30},
			occurrence: 0,
		},
		{
			name:       "source longer than condition",
			condition:  []bool{true, false},
			source:     []float64{10, 20, 30, 40},
			occurrence: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := value.Valuewhen(tt.condition, tt.source, tt.occurrence)
			if len(got) != len(tt.source) {
				t.Errorf("expected result length = %d (source length), got %d", len(tt.source), len(got))
			}
			for i := range got {
				if !math.IsNaN(got[i]) && got[i] != 0.0 {
					t.Errorf("expected NaN or 0.0 for mismatched arrays, got %v at index %d", got[i], i)
				}
			}
		})
	}
}

func assertFloatSlicesEqual(t *testing.T, got, want []float64) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("length mismatch: got %d, want %d", len(got), len(want))
	}

	for i := range got {
		if math.IsNaN(want[i]) {
			if !math.IsNaN(got[i]) {
				t.Errorf("[%d] = %v, want NaN", i, got[i])
			}
		} else {
			if got[i] != want[i] {
				t.Errorf("[%d] = %v, want %v", i, got[i], want[i])
			}
		}
	}
}
