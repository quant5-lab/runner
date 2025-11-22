package ta_test

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/runtime/ta"
)

func TestPivothigh(t *testing.T) {
	tests := []struct {
		name      string
		source    []float64
		leftBars  int
		rightBars int
		want      []float64
	}{
		{
			name:      "basic pivot high",
			source:    []float64{1, 2, 5, 3, 2, 1, 2, 4, 3, 2},
			leftBars:  2,
			rightBars: 2,
			want:      []float64{math.NaN(), math.NaN(), 5, math.NaN(), math.NaN(), math.NaN(), math.NaN(), 4, math.NaN(), math.NaN()},
		},
		{
			name:      "no pivot high",
			source:    []float64{1, 2, 3, 4, 5},
			leftBars:  1,
			rightBars: 1,
			want:      []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()},
		},
		{
			name:      "single bar pivot",
			source:    []float64{1, 5, 2},
			leftBars:  1,
			rightBars: 1,
			want:      []float64{math.NaN(), 5, math.NaN()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ta.Pivothigh(tt.source, tt.leftBars, tt.rightBars)

			if len(got) != len(tt.want) {
				t.Errorf("Pivothigh() length = %v, want %v", len(got), len(tt.want))
				return
			}

			for i := range got {
				if math.IsNaN(tt.want[i]) {
					if !math.IsNaN(got[i]) {
						t.Errorf("Pivothigh()[%d] = %v, want NaN", i, got[i])
					}
				} else {
					if got[i] != tt.want[i] {
						t.Errorf("Pivothigh()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func TestPivotlow(t *testing.T) {
	tests := []struct {
		name      string
		source    []float64
		leftBars  int
		rightBars int
		want      []float64
	}{
		{
			name:      "basic pivot low",
			source:    []float64{5, 4, 1, 3, 4, 5, 4, 2, 3, 4},
			leftBars:  2,
			rightBars: 2,
			want:      []float64{math.NaN(), math.NaN(), 1, math.NaN(), math.NaN(), math.NaN(), math.NaN(), 2, math.NaN(), math.NaN()},
		},
		{
			name:      "no pivot low",
			source:    []float64{5, 4, 3, 2, 1},
			leftBars:  1,
			rightBars: 1,
			want:      []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()},
		},
		{
			name:      "single bar pivot",
			source:    []float64{5, 1, 4},
			leftBars:  1,
			rightBars: 1,
			want:      []float64{math.NaN(), 1, math.NaN()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ta.Pivotlow(tt.source, tt.leftBars, tt.rightBars)

			if len(got) != len(tt.want) {
				t.Errorf("Pivotlow() length = %v, want %v", len(got), len(tt.want))
				return
			}

			for i := range got {
				if math.IsNaN(tt.want[i]) {
					if !math.IsNaN(got[i]) {
						t.Errorf("Pivotlow()[%d] = %v, want NaN", i, got[i])
					}
				} else {
					if got[i] != tt.want[i] {
						t.Errorf("Pivotlow()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}
