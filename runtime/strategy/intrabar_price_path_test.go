package strategy

import "testing"

func TestIntrabarPath(t *testing.T) {
	tests := []struct {
		name     string
		open     float64
		high     float64
		low      float64
		wantPath IntrabarPricePath
	}{
		// Open closer to high
		{"open_near_high", 108, 110, 100, PathHighBeforeLow},
		{"open_at_high", 110, 110, 100, PathHighBeforeLow},

		// Open closer to low
		{"open_near_low", 102, 110, 100, PathLowBeforeHigh},
		{"open_at_low", 100, 110, 100, PathLowBeforeHigh},

		// Equidistant resolves to PathHighBeforeLow (TV tie-break)
		{"open_equidistant", 105, 110, 100, PathHighBeforeLow},

		// Doji (open == high == low) resolves to PathHighBeforeLow
		{"doji", 100, 100, 100, PathHighBeforeLow},

		// Exact midpoint with asymmetric range
		{"midpoint_14_vs_14", 114, 128, 100, PathHighBeforeLow},
		{"midpoint_open_far_from_high", 113, 130, 100, PathLowBeforeHigh},
		{"midpoint_open_near_high_strict", 116, 130, 100, PathHighBeforeLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IntrabarPath(tt.open, tt.high, tt.low)
			if got != tt.wantPath {
				t.Errorf("IntrabarPath(open=%v, high=%v, low=%v) = %v, want %v",
					tt.open, tt.high, tt.low, got, tt.wantPath)
			}
		})
	}
}
