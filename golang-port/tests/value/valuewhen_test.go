package value_test

import (
	"math"
	"testing"

	"github.com/borisquantlab/pinescript-go/runtime/value"
)

func TestValuewhen(t *testing.T) {
	tests := []struct {
		name       string
		condition  []bool
		source     []float64
		occurrence int
		want       []float64
	}{
		{
			name:       "basic valuewhen occurrence 0",
			condition:  []bool{false, true, false, true, false, true},
			source:     []float64{10, 20, 30, 40, 50, 60},
			occurrence: 0,
			want:       []float64{math.NaN(), 20, 20, 40, 40, 60},
		},
		{
			name:       "valuewhen occurrence 1",
			condition:  []bool{false, true, false, true, false, true},
			source:     []float64{10, 20, 30, 40, 50, 60},
			occurrence: 1,
			want:       []float64{math.NaN(), math.NaN(), math.NaN(), 20, 20, 40},
		},
		{
			name:       "valuewhen occurrence 2",
			condition:  []bool{false, true, false, true, false, true},
			source:     []float64{10, 20, 30, 40, 50, 60},
			occurrence: 2,
			want:       []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), 20},
		},
		{
			name:       "no condition true",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := value.Valuewhen(tt.condition, tt.source, tt.occurrence)

			if len(got) != len(tt.want) {
				t.Errorf("Valuewhen() length = %v, want %v", len(got), len(tt.want))
				return
			}

			for i := range got {
				if math.IsNaN(tt.want[i]) {
					if !math.IsNaN(got[i]) {
						t.Errorf("Valuewhen()[%d] = %v, want NaN", i, got[i])
					}
				} else {
					if got[i] != tt.want[i] {
						t.Errorf("Valuewhen()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}
