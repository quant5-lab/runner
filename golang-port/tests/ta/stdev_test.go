package ta_test

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/runtime/ta"
)

func TestStdev(t *testing.T) {
	tests := []struct {
		name   string
		source []float64
		period int
		want   []float64
	}{
		{
			name:   "basic stdev",
			source: []float64{10, 12, 14, 16, 18, 20},
			period: 3,
			want:   []float64{math.NaN(), math.NaN(), 1.632993, 1.632993, 1.632993, 1.632993},
		},
		{
			name:   "constant values",
			source: []float64{5, 5, 5, 5, 5},
			period: 3,
			want:   []float64{math.NaN(), math.NaN(), 0, 0, 0},
		},
		{
			name:   "period too large",
			source: []float64{1, 2, 3},
			period: 5,
			want:   []float64{math.NaN(), math.NaN(), math.NaN()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ta.Stdev(tt.source, tt.period)

			if len(got) != len(tt.want) {
				t.Errorf("Stdev() length = %v, want %v", len(got), len(tt.want))
				return
			}

			for i := range got {
				if math.IsNaN(tt.want[i]) {
					if !math.IsNaN(got[i]) {
						t.Errorf("Stdev()[%d] = %v, want NaN", i, got[i])
					}
				} else {
					if math.Abs(got[i]-tt.want[i]) > 0.01 {
						t.Errorf("Stdev()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}
