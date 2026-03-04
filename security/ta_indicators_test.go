package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

/*
TestTrueRangeCalculator_FirstBarPolicy verifies the two first-bar policies.

When isFirstBar=true there is no prevClose available.  The handleNA parameter
governs what is returned in that situation:

  - handleNA=true  → High-Low  (ATR seed: preserve a usable numeric value)
  - handleNA=false → NaN       (ta.tr variable: signal "not enough data")
*/
func TestTrueRangeCalculator_FirstBarPolicy(t *testing.T) {
	calc := NewTrueRangeCalculator()
	bars := []context.OHLCV{{High: 110, Low: 90, Close: 100}}

	tests := []struct {
		name      string
		handleNA  bool
		expectNaN bool
		expected  float64
	}{
		{
			name:     "handleNA_true_yields_HL_range",
			handleNA: true, expectNaN: false, expected: 20.0,
		},
		{
			name:     "handleNA_false_yields_NaN",
			handleNA: false, expectNaN: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.CalculateAtBar(bars, 0, 0, true, tt.handleNA)
			if tt.expectNaN {
				if !math.IsNaN(got) {
					t.Errorf("handleNA=false: expected NaN on first bar, got %.4f", got)
				}
			} else if math.Abs(got-tt.expected) > 0.001 {
				t.Errorf("handleNA=true: expected %.4f (High-Low), got %.4f", tt.expected, got)
			}
		})
	}
}

/*
TestTrueRangeCalculator_Algorithm verifies the three-component true-range
formula: max(High-Low, |High-prevClose|, |Low-prevClose|).

These cases are independent of the first-bar / handleNA policy because they
all supply isFirstBar=false with an explicit prevClose.  handleNA is provided
in both variants to prove the parameter has no effect on non-first bars.
*/
func TestTrueRangeCalculator_Algorithm(t *testing.T) {
	calc := NewTrueRangeCalculator()

	tests := []struct {
		name      string
		bar       context.OHLCV
		prevClose float64
		expected  float64
		dominant  string
	}{
		{
			name: "HL_dominates",
			bar:  context.OHLCV{High: 110, Low: 100}, prevClose: 102,
			expected: 10.0, dominant: "High-Low=10 > |High-pC|=8 > |Low-pC|=2",
		},
		{
			name: "HC_dominates_gap_up",
			bar:  context.OHLCV{High: 108, Low: 100}, prevClose: 90,
			expected: 18.0, dominant: "|High-pC|=18 > High-Low=8 > |Low-pC|=10",
		},
		{
			name: "LC_dominates_gap_down",
			bar:  context.OHLCV{High: 108, Low: 98}, prevClose: 110,
			expected: 12.0, dominant: "|Low-pC|=12 > High-Low=10 > |High-pC|=2",
		},
		{
			name: "large_gap_up",
			bar:  context.OHLCV{High: 125, Low: 120}, prevClose: 100,
			expected: 25.0, dominant: "|High-pC|=25 captures large gap up",
		},
		{
			name: "large_gap_down",
			bar:  context.OHLCV{High: 85, Low: 80}, prevClose: 100,
			expected: 20.0, dominant: "|Low-pC|=20 captures large gap down",
		},
		{
			name: "equal_high_low_doji",
			bar:  context.OHLCV{High: 100, Low: 100}, prevClose: 100,
			expected: 0.0, dominant: "all components zero — perfect doji",
		},
	}

	for _, handleNA := range []bool{false, true} {
		for _, tt := range tests {
			tt := tt
			handleNA := handleNA
			name := tt.name
			if handleNA {
				name += "/handleNA_true"
			} else {
				name += "/handleNA_false"
			}
			t.Run(name, func(t *testing.T) {
				bars := []context.OHLCV{tt.bar}
				got := calc.CalculateAtBar(bars, 0, tt.prevClose, false, handleNA)
				if math.Abs(got-tt.expected) > 0.001 {
					t.Errorf("%s: expected %.4f, got %.4f", tt.dominant, tt.expected, got)
				}
			})
		}
	}
}

/*
TestTrueRangeCalculator_BoundsGuard verifies that out-of-range bar indices
always return NaN regardless of isFirstBar or handleNA values.
*/
func TestTrueRangeCalculator_BoundsGuard(t *testing.T) {
	calc := NewTrueRangeCalculator()
	bars := []context.OHLCV{{High: 105, Low: 95, Close: 100}}

	tests := []struct {
		name   string
		barIdx int
	}{
		{"negative_index", -1},
		{"one_past_end", 1},
		{"far_out_of_bounds", 999},
	}

	for _, tt := range tests {
		for _, isFirst := range []bool{false, true} {
			for _, handleNA := range []bool{false, true} {
				isFirst, handleNA := isFirst, handleNA
				name := tt.name
				if isFirst {
					name += "/isFirst"
				}
				if handleNA {
					name += "/handleNA"
				}
				t.Run(name, func(t *testing.T) {
					got := calc.CalculateAtBar(bars, tt.barIdx, 100, isFirst, handleNA)
					if !math.IsNaN(got) {
						t.Errorf("OOB index %d: expected NaN, got %.4f", tt.barIdx, got)
					}
				})
			}
		}
	}
}

/*
TestTrueRangeCalculator_ZeroRange verifies that a bar with High==Low produces
0.0 when prevClose is also equal (no gap, no intra-bar range).
*/
func TestTrueRangeCalculator_ZeroRange(t *testing.T) {
	calc := NewTrueRangeCalculator()
	bars := []context.OHLCV{{High: 100, Low: 100}}

	got := calc.CalculateAtBar(bars, 0, 100.0, false, false)
	if math.IsNaN(got) {
		t.Fatal("zero-range bar with prevClose equal: expected 0.0, got NaN")
	}
	if got != 0.0 {
		t.Errorf("zero-range bar: expected 0.0, got %.4f", got)
	}
}
