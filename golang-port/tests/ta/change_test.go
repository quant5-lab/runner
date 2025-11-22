package ta_test

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/runtime/ta"
)

func TestChange(t *testing.T) {
	tests := []struct {
		name   string
		source []float64
		want   []float64
	}{
		{
			name:   "basic change",
			source: []float64{10, 12, 11, 15, 14},
			want:   []float64{math.NaN(), 2, -1, 4, -1},
		},
		{
			name:   "constant values",
			source: []float64{5, 5, 5, 5},
			want:   []float64{math.NaN(), 0, 0, 0},
		},
		{
			name:   "single value",
			source: []float64{10},
			want:   []float64{math.NaN()},
		},
		{
			name:   "with NaN",
			source: []float64{10, math.NaN(), 15},
			want:   []float64{math.NaN(), math.NaN(), math.NaN()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ta.Change(tt.source)

			if len(got) != len(tt.want) {
				t.Errorf("Change() length = %v, want %v", len(got), len(tt.want))
				return
			}

			for i := range got {
				if math.IsNaN(tt.want[i]) {
					if !math.IsNaN(got[i]) {
						t.Errorf("Change()[%d] = %v, want NaN", i, got[i])
					}
				} else {
					if got[i] != tt.want[i] {
						t.Errorf("Change()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}
