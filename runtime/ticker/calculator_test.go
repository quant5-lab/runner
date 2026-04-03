package ticker

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

func TestCalculateHeikinAshiBar_FirstBar(t *testing.T) {
	current := context.OHLCV{
		Open:  100.0,
		High:  110.0,
		Low:   95.0,
		Close: 105.0,
	}

	haBar := CalculateHeikinAshiBar(current, context.OHLCV{}, 0, 0)

	expectedHaClose := (100.0 + 110.0 + 95.0 + 105.0) / 4.0
	if haBar.Close != expectedHaClose {
		t.Errorf("haClose = %v, want %v", haBar.Close, expectedHaClose)
	}

	expectedHaOpen := (100.0 + 105.0) / 2.0
	if haBar.Open != expectedHaOpen {
		t.Errorf("haOpen = %v, want %v", haBar.Open, expectedHaOpen)
	}

	if haBar.High != 110.0 {
		t.Errorf("haHigh = %v, want 110.0", haBar.High)
	}

	if haBar.Low != 95.0 {
		t.Errorf("haLow = %v, want 95.0", haBar.Low)
	}
}

func TestCalculateHeikinAshiBar_SecondBar(t *testing.T) {
	current := context.OHLCV{
		Open:  105.0,
		High:  115.0,
		Low:   100.0,
		Close: 110.0,
	}

	prevHaOpen := 102.5
	prevHaClose := 102.5

	haBar := CalculateHeikinAshiBar(current, context.OHLCV{}, prevHaOpen, prevHaClose)

	expectedHaClose := (105.0 + 115.0 + 100.0 + 110.0) / 4.0
	if haBar.Close != expectedHaClose {
		t.Errorf("haClose = %v, want %v", haBar.Close, expectedHaClose)
	}

	expectedHaOpen := (prevHaOpen + prevHaClose) / 2.0
	if haBar.Open != expectedHaOpen {
		t.Errorf("haOpen = %v, want %v", haBar.Open, expectedHaOpen)
	}
}

func TestCalculateHeikinAshiBar_HighLowCalculation(t *testing.T) {
	tests := []struct {
		name           string
		current        context.OHLCV
		prevHaOpen     float64
		prevHaClose    float64
		expectedHaHigh float64
		expectedHaLow  float64
	}{
		{
			name: "high is max of all three",
			current: context.OHLCV{
				Open: 100.0, High: 120.0, Low: 95.0, Close: 105.0,
			},
			prevHaOpen:     102.0,
			prevHaClose:    103.0,
			expectedHaHigh: 120.0,
			expectedHaLow:  95.0,
		},
		{
			name: "haOpen is higher than high",
			current: context.OHLCV{
				Open: 100.0, High: 105.0, Low: 95.0, Close: 100.0,
			},
			prevHaOpen:     110.0,
			prevHaClose:    108.0,
			expectedHaHigh: 109.0,
			expectedHaLow:  95.0,
		},
		{
			name: "haClose is higher than high",
			current: context.OHLCV{
				Open: 100.0, High: 105.0, Low: 95.0, Close: 115.0,
			},
			prevHaOpen:     100.0,
			prevHaClose:    100.0,
			expectedHaHigh: 105.0,
			expectedHaLow:  95.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			haBar := CalculateHeikinAshiBar(tt.current, context.OHLCV{}, tt.prevHaOpen, tt.prevHaClose)

			if haBar.High != tt.expectedHaHigh {
				t.Errorf("haHigh = %v, want %v", haBar.High, tt.expectedHaHigh)
			}

			if haBar.Low != tt.expectedHaLow {
				t.Errorf("haLow = %v, want %v", haBar.Low, tt.expectedHaLow)
			}
		})
	}
}

func TestMax3(t *testing.T) {
	tests := []struct {
		a, b, c  float64
		expected float64
	}{
		{1.0, 2.0, 3.0, 3.0},
		{3.0, 2.0, 1.0, 3.0},
		{2.0, 3.0, 1.0, 3.0},
		{5.0, 5.0, 5.0, 5.0},
		{-1.0, -2.0, -3.0, -1.0},
	}

	for _, tt := range tests {
		result := max3(tt.a, tt.b, tt.c)
		if result != tt.expected {
			t.Errorf("max3(%v, %v, %v) = %v, want %v", tt.a, tt.b, tt.c, result, tt.expected)
		}
	}
}

func TestMin3(t *testing.T) {
	tests := []struct {
		a, b, c  float64
		expected float64
	}{
		{1.0, 2.0, 3.0, 1.0},
		{3.0, 2.0, 1.0, 1.0},
		{2.0, 3.0, 1.0, 1.0},
		{5.0, 5.0, 5.0, 5.0},
		{-1.0, -2.0, -3.0, -3.0},
	}

	for _, tt := range tests {
		result := min3(tt.a, tt.b, tt.c)
		if result != tt.expected {
			t.Errorf("min3(%v, %v, %v) = %v, want %v", tt.a, tt.b, tt.c, result, tt.expected)
		}
	}
}

func TestCalculateHeikinAshiBar_EdgeCases(t *testing.T) {
	tests := []struct {
		name               string
		bar                context.OHLCV
		prevHa             context.OHLCV
		validateInvariants func(t *testing.T, result context.OHLCV)
	}{
		{
			name:   "zero_prices",
			bar:    context.OHLCV{Open: 0, High: 0, Low: 0, Close: 0, Volume: 100, Time: 1000},
			prevHa: context.OHLCV{Open: 0, High: 0, Low: 0, Close: 0},
			validateInvariants: func(t *testing.T, result context.OHLCV) {
				if result.Open != 0 || result.High != 0 || result.Low != 0 || result.Close != 0 {
					t.Errorf("zero prices: got OHLC=(%v,%v,%v,%v), want all zeros",
						result.Open, result.High, result.Low, result.Close)
				}
			},
		},
		{
			name:   "negative_prices",
			bar:    context.OHLCV{Open: -10, High: -5, Low: -15, Close: -8, Volume: 100, Time: 2000},
			prevHa: context.OHLCV{Open: -12, High: -5, Low: -15, Close: -10},
			validateInvariants: func(t *testing.T, result context.OHLCV) {
				if result.High < result.Open || result.High < result.Close {
					t.Errorf("negative prices: haHigh=%v must be >= max(haOpen=%v, haClose=%v)",
						result.High, result.Open, result.Close)
				}
				if result.Low > result.Open || result.Low > result.Close {
					t.Errorf("negative prices: haLow=%v must be <= min(haOpen=%v, haClose=%v)",
						result.Low, result.Open, result.Close)
				}
			},
		},
		{
			name:   "extreme_large_values",
			bar:    context.OHLCV{Open: 1e10, High: 1.1e10, Low: 0.9e10, Close: 1.05e10, Volume: 1000, Time: 3000},
			prevHa: context.OHLCV{Open: 1e10, High: 1.1e10, Low: 0.9e10, Close: 1e10},
			validateInvariants: func(t *testing.T, result context.OHLCV) {
				expectedClose := (1e10 + 1.1e10 + 0.9e10 + 1.05e10) / 4
				if math.Abs(result.Close-expectedClose) > 1e8 {
					t.Errorf("extreme values: haClose=%v, want %v", result.Close, expectedClose)
				}
			},
		},
		{
			name:   "extreme_small_values",
			bar:    context.OHLCV{Open: 1e-10, High: 1.1e-10, Low: 0.9e-10, Close: 1.05e-10, Volume: 100, Time: 4000},
			prevHa: context.OHLCV{Open: 1e-10, High: 1.1e-10, Low: 0.9e-10, Close: 1e-10},
			validateInvariants: func(t *testing.T, result context.OHLCV) {
				if result.High < max3(result.Open, result.Close, 0.9e-10) {
					t.Errorf("extreme small: haHigh=%v too small", result.High)
				}
			},
		},
		{
			name:   "all_same_price",
			bar:    context.OHLCV{Open: 100, High: 100, Low: 100, Close: 100, Volume: 500, Time: 5000},
			prevHa: context.OHLCV{Open: 100, High: 100, Low: 100, Close: 100},
			validateInvariants: func(t *testing.T, result context.OHLCV) {
				if result.Open != 100 || result.High != 100 || result.Low != 100 || result.Close != 100 {
					t.Errorf("all same price: got OHLC=(%v,%v,%v,%v), want all 100",
						result.Open, result.High, result.Low, result.Close)
				}
			},
		},
		{
			name:   "volume_preservation",
			bar:    context.OHLCV{Open: 100, High: 110, Low: 95, Close: 105, Volume: 12345, Time: 6000},
			prevHa: context.OHLCV{Open: 98, High: 108, Low: 93, Close: 102},
			validateInvariants: func(t *testing.T, result context.OHLCV) {
				if result.Volume != 12345 {
					t.Errorf("volume not preserved: got %v, want 12345", result.Volume)
				}
			},
		},
		{
			name:   "time_preservation",
			bar:    context.OHLCV{Open: 100, High: 110, Low: 95, Close: 105, Volume: 100, Time: 123456789},
			prevHa: context.OHLCV{Open: 98, High: 108, Low: 93, Close: 102},
			validateInvariants: func(t *testing.T, result context.OHLCV) {
				if result.Time != 123456789 {
					t.Errorf("time not preserved: got %v, want 123456789", result.Time)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateHeikinAshiBar(tt.bar, tt.prevHa, tt.prevHa.Open, tt.prevHa.Close)
			tt.validateInvariants(t, result)
		})
	}
}

func TestCalculateHeikinAshiBar_AlgorithmInvariants(t *testing.T) {
	tests := []struct {
		name   string
		bar    context.OHLCV
		prevHa context.OHLCV
	}{
		{
			name:   "uptrend_bar",
			bar:    context.OHLCV{Open: 100, High: 110, Low: 95, Close: 108, Volume: 100, Time: 1000},
			prevHa: context.OHLCV{Open: 98, High: 105, Low: 93, Close: 102},
		},
		{
			name:   "downtrend_bar",
			bar:    context.OHLCV{Open: 100, High: 105, Low: 90, Close: 92, Volume: 100, Time: 2000},
			prevHa: context.OHLCV{Open: 102, High: 108, Low: 98, Close: 104},
		},
		{
			name:   "doji_bar",
			bar:    context.OHLCV{Open: 100, High: 101, Low: 99, Close: 100, Volume: 100, Time: 3000},
			prevHa: context.OHLCV{Open: 100, High: 102, Low: 98, Close: 100},
		},
		{
			name:   "wide_range_bar",
			bar:    context.OHLCV{Open: 100, High: 150, Low: 50, Close: 120, Volume: 100, Time: 4000},
			prevHa: context.OHLCV{Open: 90, High: 110, Low: 80, Close: 95},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateHeikinAshiBar(tt.bar, tt.prevHa, tt.prevHa.Open, tt.prevHa.Close)

			maxOC := max3(result.Open, result.Close, result.Close)
			if result.High < maxOC {
				t.Errorf("invariant violation: haHigh=%v < max(haOpen=%v, haClose=%v)=%v",
					result.High, result.Open, result.Close, maxOC)
			}

			minOC := min3(result.Open, result.Close, result.Close)
			if result.Low > minOC {
				t.Errorf("invariant violation: haLow=%v > min(haOpen=%v, haClose=%v)=%v",
					result.Low, result.Open, result.Close, minOC)
			}

			if result.High < result.Low {
				t.Errorf("invariant violation: haHigh=%v < haLow=%v", result.High, result.Low)
			}

			expectedClose := (tt.bar.Open + tt.bar.High + tt.bar.Low + tt.bar.Close) / 4
			if math.Abs(result.Close-expectedClose) > 0.0001 {
				t.Errorf("invariant violation: haClose=%v, expected (O+H+L+C)/4=%v",
					result.Close, expectedClose)
			}

			if tt.prevHa.Open != 0 || tt.prevHa.Close != 0 {
				expectedOpen := (tt.prevHa.Open + tt.prevHa.Close) / 2
				if math.Abs(result.Open-expectedOpen) > 0.0001 {
					t.Errorf("invariant violation: haOpen=%v, expected (prevOpen+prevClose)/2=%v",
						result.Open, expectedOpen)
				}
			}
		})
	}
}

func TestCalculateHeikinAshiBar_SequentialStatePropagation(t *testing.T) {
	bars := []context.OHLCV{
		{Open: 100, High: 110, Low: 95, Close: 105, Volume: 100, Time: 1000},
		{Open: 105, High: 115, Low: 100, Close: 112, Volume: 200, Time: 2000},
		{Open: 112, High: 120, Low: 108, Close: 118, Volume: 150, Time: 3000},
		{Open: 118, High: 125, Low: 115, Close: 122, Volume: 300, Time: 4000},
		{Open: 122, High: 130, Low: 120, Close: 128, Volume: 250, Time: 5000},
	}

	prevHa := context.OHLCV{}
	for i, bar := range bars {
		result := CalculateHeikinAshiBar(bar, prevHa, prevHa.Open, prevHa.Close)

		if i > 0 {
			expectedOpen := (prevHa.Open + prevHa.Close) / 2
			if math.Abs(result.Open-expectedOpen) > 0.0001 {
				t.Errorf("bar %d: state propagation failed, haOpen=%v, want %v",
					i, result.Open, expectedOpen)
			}
		}

		if result.Volume != bar.Volume {
			t.Errorf("bar %d: volume changed from %v to %v", i, bar.Volume, result.Volume)
		}
		if result.Time != bar.Time {
			t.Errorf("bar %d: time changed from %v to %v", i, bar.Time, result.Time)
		}

		prevHa = result
	}
}
