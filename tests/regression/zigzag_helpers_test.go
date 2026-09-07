package regression

import (
	"math"
	"testing"
)

// Shared by runtime-evidence and pivot-evidence tests; a regression that re-zeroes
// the trade series surfaces immediately in either canary.
// Calibrated against the 484-trade count; occurrenceCount==4→==3 drops it to ~322.
const minZigzagClosedTrades = 484

// Field names match Pine variables verbatim (zigzag.pine lines 32-35).
type harmonicRatios struct {
	xab, xad, abc, bcd float64
}

// Window: w[0]=x (oldest) … w[4]=d (most recent), mirroring valuewhen(sz,sz,4..0).
// Raw Go division: nonzero/0 = +Inf (matches generated runtime lines 958-961); 0/0 = NaN.
func harmonicRatiosAt(w [5]float64) harmonicRatios {
	return harmonicRatios{
		xab: math.Abs(w[2]-w[1]) / math.Abs(w[0]-w[1]),
		xad: math.Abs(w[1]-w[4]) / math.Abs(w[0]-w[1]),
		abc: math.Abs(w[2]-w[3]) / math.Abs(w[1]-w[2]),
		bcd: math.Abs(w[3]-w[4]) / math.Abs(w[2]-w[3]),
	}
}

// Indices mirror valuewhen(sz,sz,4..0): result[0]=x (oldest), result[4]=d (most recent).
func pivotWindowAt(series []float64, bar int) ([5]float64, bool) {
	if bar >= len(series) {
		return [5]float64{}, false
	}
	var window [5]float64
	found := 0
	for i := bar; i >= 0 && found < 5; i-- {
		if !math.IsNaN(series[i]) {
			window[4-found] = series[i]
			found++
		}
	}
	return window, found == 5
}

func fifthPivotBar(series []float64) int {
	count := 0
	for i, v := range series {
		if !math.IsNaN(v) {
			count++
			if count == 5 {
				return i
			}
		}
	}
	return -1
}

func countPivots(series []float64, upToBar int) int {
	count := 0
	end := upToBar + 1
	if end > len(series) {
		end = len(series)
	}
	for _, v := range series[:end] {
		if !math.IsNaN(v) {
			count++
		}
	}
	return count
}

// Ratio formulas from zigzag.pine lines 32-35; window is [x,a,b,c,d] oldest-first.
func TestHarmonicRatiosAt(t *testing.T) {
	nan := math.NaN()
	inf := math.Inf(1)
	tol := 1e-9

	tests := []struct {
		name string
		w    [5]float64
		want harmonicRatios
	}{
		{
			name: "standard_window",
			w:    [5]float64{100, 110, 105, 108, 103},
			want: harmonicRatios{xab: 0.5, xad: 0.7, abc: 0.6, bcd: 5.0 / 3.0},
		},
		{
			name: "bcd_near_upper_band_edge",
			w:    [5]float64{119.06, 117.51, 119.21, 118.53, 120.30},
			want: harmonicRatios{
				xab: math.Abs(119.21-117.51) / math.Abs(119.06-117.51),
				xad: math.Abs(117.51-120.30) / math.Abs(119.06-117.51),
				abc: math.Abs(119.21-118.53) / math.Abs(117.51-119.21),
				bcd: math.Abs(118.53-120.30) / math.Abs(119.21-118.53),
			},
		},
		{
			name: "all_equal_pivots_yields_all_nan",
			w:    [5]float64{100, 100, 100, 100, 100},
			want: harmonicRatios{xab: nan, xad: nan, abc: nan, bcd: nan},
		},
		{
			name: "equal_x_and_a_yields_inf_for_xab_and_xad",
			w:    [5]float64{110, 110, 105, 108, 103},
			want: harmonicRatios{xab: inf, xad: inf, abc: 0.6, bcd: 5.0 / 3.0},
		},
		{
			name: "equal_a_and_b_yields_inf_for_abc",
			w:    [5]float64{100, 110, 110, 105, 103},
			want: harmonicRatios{xab: 0.0, xad: 0.7, abc: inf, bcd: math.Abs(105-103) / math.Abs(110-105)},
		},
		{
			name: "equal_b_and_c_yields_inf_for_bcd",
			w:    [5]float64{100, 110, 105, 105, 103},
			want: harmonicRatios{xab: 0.5, xad: 0.7, abc: 0.0, bcd: inf},
		},
		{
			name: "symmetric_window",
			w:    [5]float64{100, 110, 100, 110, 100},
			want: harmonicRatios{xab: 1.0, xad: 1.0, abc: 1.0, bcd: 1.0},
		},
		{
			name: "d_equals_a_yields_xad_zero",
			w:    [5]float64{100, 110, 105, 108, 110},
			// xad = |a-d|/|x-a| = |110-110|/|100-110| = 0/10 = 0.0 (zero numerator, non-zero denominator)
			want: harmonicRatios{
				xab: math.Abs(105-110) / math.Abs(100-110),
				xad: 0.0,
				abc: math.Abs(105-108) / math.Abs(110-105),
				bcd: math.Abs(108-110) / math.Abs(105-108),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := harmonicRatiosAt(tt.w)
			assertRatioField(t, "xab", got.xab, tt.want.xab, tol)
			assertRatioField(t, "xad", got.xad, tt.want.xad, tol)
			assertRatioField(t, "abc", got.abc, tt.want.abc, tol)
			assertRatioField(t, "bcd", got.bcd, tt.want.bcd, tol)
		})
	}
}

func TestPivotWindowAt(t *testing.T) {
	nan := math.NaN()

	tests := []struct {
		name   string
		series []float64
		bar    int
		wantW  [5]float64
		wantOK bool
	}{
		{name: "empty_series", series: []float64{}, bar: 0, wantOK: false},
		{name: "bar_beyond_series_length", series: []float64{1, 2, 3, 4, 5}, bar: 10, wantOK: false},
		{
			name:   "exactly_five_non_nan_consecutive",
			series: []float64{1, 2, 3, 4, 5},
			bar:    4,
			wantW:  [5]float64{1, 2, 3, 4, 5},
			wantOK: true,
		},
		{
			name:   "fewer_than_five_non_nan",
			series: []float64{1, nan, 2, nan, 3, nan},
			bar:    5,
			wantOK: false,
		},
		{
			name:   "five_non_nan_with_nan_interleaved",
			series: []float64{1, nan, 2, nan, 3, nan, 4, nan, 5},
			bar:    8,
			wantW:  [5]float64{1, 2, 3, 4, 5},
			wantOK: true,
		},
		{
			name:   "window_stops_at_bar_not_end_of_series",
			series: []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			bar:    4,
			wantW:  [5]float64{1, 2, 3, 4, 5},
			wantOK: true,
		},
		{
			name:   "six_non_nan_returns_most_recent_five",
			series: []float64{1, 2, 3, 4, 5, 6},
			bar:    5,
			wantW:  [5]float64{2, 3, 4, 5, 6},
			wantOK: true,
		},
		{
			name:   "all_nan_except_five",
			series: []float64{nan, nan, 10, nan, 20, nan, 30, nan, 40, nan, 50},
			bar:    10,
			wantW:  [5]float64{10, 20, 30, 40, 50},
			wantOK: true,
		},
		{name: "bar_zero_with_one_value", series: []float64{42.0}, bar: 0, wantOK: false},
		{
			name:   "bar_before_fifth_pivot_yields_false",
			series: []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			bar:    3,
			wantOK: false,
		},
		{
			name:   "bar_at_nan_tail_with_four_non_nan_below_yields_false",
			series: []float64{1, 2, 3, 4, math.NaN()},
			bar:    4,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := pivotWindowAt(tt.series, tt.bar)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			for i, v := range got {
				if math.Abs(v-tt.wantW[i]) > 1e-9 {
					t.Errorf("w[%d] = %.9f, want %.9f", i, v, tt.wantW[i])
				}
			}
		})
	}
}

func TestFifthPivotBar(t *testing.T) {
	nan := math.NaN()

	tests := []struct {
		name   string
		series []float64
		want   int
	}{
		{name: "empty_series", series: []float64{}, want: -1},
		{name: "all_nan", series: []float64{nan, nan, nan}, want: -1},
		{name: "exactly_four_non_nan", series: []float64{1, 2, 3, 4, nan}, want: -1},
		{name: "exactly_five_non_nan_consecutive", series: []float64{1, 2, 3, 4, 5}, want: 4},
		{name: "fifth_non_nan_with_gaps", series: []float64{1, nan, 2, nan, 3, nan, 4, nan, 5, nan}, want: 8},
		{name: "sixth_non_nan_returns_fifth_position", series: []float64{1, 2, 3, 4, 5, 6}, want: 4},
		{name: "leading_nan_shifts_fifth_right", series: []float64{nan, nan, 1, 2, 3, 4, 5}, want: 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fifthPivotBar(tt.series)
			if got != tt.want {
				t.Errorf("fifthPivotBar = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCountPivots(t *testing.T) {
	nan := math.NaN()

	tests := []struct {
		name    string
		series  []float64
		upToBar int
		want    int
	}{
		{name: "empty_series", series: []float64{}, upToBar: 0, want: 0},
		{name: "upToBar_negative_yields_zero", series: []float64{1, 2, 3}, upToBar: -1, want: 0},
		{name: "all_nan", series: []float64{nan, nan, nan}, upToBar: 2, want: 0},
		{name: "upToBar_clamps_beyond_series", series: []float64{1, 2, 3}, upToBar: 100, want: 3},
		{name: "upToBar_zero_includes_first_bar", series: []float64{1, 2, 3}, upToBar: 0, want: 1},
		{name: "upToBar_mid_series", series: []float64{1, nan, 2, nan, 3}, upToBar: 2, want: 2},
		{name: "all_non_nan", series: []float64{1, 2, 3, 4, 5}, upToBar: 4, want: 5},
		{
			name:    "interleaved_nan_partial_bar",
			series:  []float64{1, nan, 2, nan, 3, nan, 4},
			upToBar: 4,
			want:    3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countPivots(tt.series, tt.upToBar)
			if got != tt.want {
				t.Errorf("countPivots(upToBar=%d) = %d, want %d", tt.upToBar, got, tt.want)
			}
		})
	}
}

func assertRatioField(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.IsNaN(want) {
		if !math.IsNaN(got) {
			t.Errorf("%s = %.9f, want NaN", name, got)
		}
		return
	}
	if math.IsInf(want, 0) {
		if got != want {
			t.Errorf("%s = %v, want %v", name, got, want)
		}
		return
	}
	if math.IsNaN(got) {
		t.Errorf("%s = NaN, want %.9f", name, want)
		return
	}
	if math.Abs(got-want) > tol {
		t.Errorf("%s = %.9f, want %.9f (diff %.2e)", name, got, want, math.Abs(got-want))
	}
}
