package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

func TestTrueRangeCalculator_FirstBar(t *testing.T) {
	calc := NewTrueRangeCalculator()

	bars := []context.OHLCV{
		{Open: 100, High: 105, Low: 95, Close: 102},
	}

	tr := calc.CalculateAtBar(bars, 0, 0, true)
	expected := 10.0

	if tr != expected {
		t.Errorf("First bar TR: expected %.2f, got %.2f", expected, tr)
	}
}

func TestTrueRangeCalculator_TrueRangeComponents(t *testing.T) {
	calc := NewTrueRangeCalculator()

	tests := []struct {
		name      string
		bars      []context.OHLCV
		prevClose float64
		expected  float64
		desc      string
	}{
		{
			name: "HL_highest",
			bars: []context.OHLCV{
				{High: 110, Low: 100},
			},
			prevClose: 102,
			expected:  10.0,
			desc:      "high-low (10) > abs(high-prevClose) (8) > abs(low-prevClose) (2)",
		},
		{
			name: "HC_highest",
			bars: []context.OHLCV{
				{High: 108, Low: 100},
			},
			prevClose: 90,
			expected:  18.0,
			desc:      "abs(high-prevClose) (18) > high-low (8) > abs(low-prevClose) (10)",
		},
		{
			name: "LC_highest",
			bars: []context.OHLCV{
				{High: 108, Low: 98},
			},
			prevClose: 110,
			expected:  12.0,
			desc:      "abs(low-prevClose) (12) > high-low (10) > abs(high-prevClose) (2)",
		},
		{
			name: "gap_up",
			bars: []context.OHLCV{
				{High: 125, Low: 120},
			},
			prevClose: 100,
			expected:  25.0,
			desc:      "gap up: abs(high-prevClose) (25) captures gap",
		},
		{
			name: "gap_down",
			bars: []context.OHLCV{
				{High: 85, Low: 80},
			},
			prevClose: 100,
			expected:  20.0,
			desc:      "gap down: abs(low-prevClose) (20) captures gap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := calc.CalculateAtBar(tt.bars, 0, tt.prevClose, false)
			if math.Abs(tr-tt.expected) > 0.01 {
				t.Errorf("%s: expected %.2f, got %.2f", tt.desc, tt.expected, tr)
			}
		})
	}
}

func TestTrueRangeCalculator_BoundaryConditions(t *testing.T) {
	calc := NewTrueRangeCalculator()

	tests := []struct {
		name      string
		bars      []context.OHLCV
		barIdx    int
		prevClose float64
		isFirst   bool
		expectNaN bool
		desc      string
	}{
		{
			name:      "out_of_bounds_positive",
			bars:      []context.OHLCV{{High: 105, Low: 95}},
			barIdx:    10,
			prevClose: 100,
			isFirst:   false,
			expectNaN: true,
			desc:      "index beyond data length",
		},
		{
			name:      "negative_index",
			bars:      []context.OHLCV{{High: 105, Low: 95}},
			barIdx:    -1,
			prevClose: 100,
			isFirst:   false,
			expectNaN: true,
			desc:      "negative bar index",
		},
		{
			name:      "zero_range",
			bars:      []context.OHLCV{{High: 100, Low: 100}},
			barIdx:    0,
			prevClose: 100,
			isFirst:   true,
			expectNaN: false,
			desc:      "no price movement within bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := calc.CalculateAtBar(tt.bars, tt.barIdx, tt.prevClose, tt.isFirst)
			if tt.expectNaN {
				if !math.IsNaN(tr) {
					t.Errorf("%s: expected NaN, got %.2f", tt.desc, tr)
				}
			} else {
				if math.IsNaN(tr) {
					t.Errorf("%s: expected valid value, got NaN", tt.desc)
				}
			}
		})
	}
}
