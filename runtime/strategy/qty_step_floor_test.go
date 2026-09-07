package strategy

import (
	"math"
	"testing"
)

func TestFloorToStep(t *testing.T) {
	cases := []struct {
		name string
		qty  float64
		step float64
		want float64
		// tol is the absolute tolerance; non-zero only for sub-unit step sizes
		// where float64 arithmetic introduces representation error in the multiple.
		tol float64
	}{
		{"fractional qty floors down", 44.29482636, 1, 44, 0},
		{"already exact multiple is unchanged", 44.0, 1, 44, 0},
		{"qty just below next multiple floors down", 44.9999, 1, 44, 0},
		{"zero qty stays zero", 0, 1, 0, 0},
		{"large fractional qty", 393.4933747390324, 1, 393, 0},
		{"step larger than qty floors to zero", 0.3, 1, 0, 0},
		{"step equals qty is exact multiple", 5.0, 5.0, 5.0, 0},
		{"sub-unit step floors to nearest multiple", 2.75, 0.25, 2.75, 0},
		{"sub-unit step with rounding residual", 2.73, 0.5, 2.5, 1e-12},
		{"four-decimal step", 0.93992044, 0.0001, 0.9399, 1e-12},

		{"zero step pass-through fractional", 0.939920, 0, 0.939920, 0},
		{"zero step pass-through integer", 44.0, 0, 44.0, 0},
		{"negative step pass-through", 1.5, -1, 1.5, 0},

		{"NaN qty with positive step", math.NaN(), 1, math.NaN(), 0},
		{"NaN qty with zero step", math.NaN(), 0, math.NaN(), 0},
		{"positive infinity with zero step", math.Inf(1), 0, math.Inf(1), 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := floorToStep(tc.qty, tc.step)
			if math.IsNaN(tc.want) {
				if !math.IsNaN(got) {
					t.Errorf("floorToStep(%v, %v) = %v, want NaN", tc.qty, tc.step, got)
				}
				return
			}
			if tc.tol == 0 {
				if got != tc.want {
					t.Errorf("floorToStep(%v, %v) = %.15g, want %.15g", tc.qty, tc.step, got, tc.want)
				}
			} else {
				if math.Abs(got-tc.want) > tc.tol {
					t.Errorf("floorToStep(%v, %v) = %.15g, want %.15g (±%g)", tc.qty, tc.step, got, tc.want, tc.tol)
				}
			}
		})
	}
}
