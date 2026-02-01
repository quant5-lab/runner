package ticker

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

func TestIdentityTransformer(t *testing.T) {
	transformer := &IdentityTransformer{}

	bars := []context.OHLCV{
		{Open: 100.0, High: 110.0, Low: 95.0, Close: 105.0},
		{Open: 105.0, High: 115.0, Low: 100.0, Close: 110.0},
	}

	result := transformer.Transform(bars)

	if len(result) != len(bars) {
		t.Fatalf("Expected %d bars, got %d", len(bars), len(result))
	}

	for i := range bars {
		if result[i] != bars[i] {
			t.Errorf("Bar %d changed: %+v != %+v", i, result[i], bars[i])
		}
	}

	if transformer.Type() != "" {
		t.Errorf("IdentityTransformer.Type() = %q, want empty", transformer.Type())
	}
}

func TestHeikinAshiTransformer_EmptyBars(t *testing.T) {
	transformer := &HeikinAshiTransformer{}
	result := transformer.Transform([]context.OHLCV{})

	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d bars", len(result))
	}
}

func TestHeikinAshiTransformer_SingleBar(t *testing.T) {
	transformer := &HeikinAshiTransformer{}

	bars := []context.OHLCV{
		{Open: 100.0, High: 110.0, Low: 95.0, Close: 105.0, Volume: 1000.0},
	}

	result := transformer.Transform(bars)

	if len(result) != 1 {
		t.Fatalf("Expected 1 bar, got %d", len(result))
	}

	haBar := result[0]
	expectedHaClose := (100.0 + 110.0 + 95.0 + 105.0) / 4.0
	if haBar.Close != expectedHaClose {
		t.Errorf("haClose = %v, want %v", haBar.Close, expectedHaClose)
	}

	if haBar.Volume != 1000.0 {
		t.Errorf("Volume changed: %v != 1000.0", haBar.Volume)
	}
}

func TestHeikinAshiTransformer_MultipleBars(t *testing.T) {
	transformer := &HeikinAshiTransformer{}

	bars := []context.OHLCV{
		{Open: 100.0, High: 110.0, Low: 95.0, Close: 105.0},
		{Open: 105.0, High: 115.0, Low: 100.0, Close: 110.0},
		{Open: 110.0, High: 120.0, Low: 105.0, Close: 115.0},
	}

	result := transformer.Transform(bars)

	if len(result) != 3 {
		t.Fatalf("Expected 3 bars, got %d", len(result))
	}

	for i := 1; i < len(result); i++ {
		prevBar := result[i-1]
		currBar := result[i]

		expectedHaOpen := (prevBar.Open + prevBar.Close) / 2.0
		if currBar.Open != expectedHaOpen {
			t.Errorf("Bar %d: haOpen = %v, want %v (from prev haOpen=%v, haClose=%v)",
				i, currBar.Open, expectedHaOpen, prevBar.Open, prevBar.Close)
		}
	}
}

func TestHeikinAshiTransformer_Type(t *testing.T) {
	transformer := &HeikinAshiTransformer{}

	if transformer.Type() != ModifierHeikinAshi {
		t.Errorf("Type() = %q, want %q", transformer.Type(), ModifierHeikinAshi)
	}
}

func TestNewTransformer(t *testing.T) {
	tests := []struct {
		modifierType ModifierType
		expectedType string
	}{
		{ModifierHeikinAshi, "*ticker.HeikinAshiTransformer"},
		{ModifierRenko, "*ticker.IdentityTransformer"},
		{ModifierKagi, "*ticker.IdentityTransformer"},
		{ModifierLineBreak, "*ticker.IdentityTransformer"},
		{"", "*ticker.IdentityTransformer"},
	}

	for _, tt := range tests {
		t.Run(string(tt.modifierType), func(t *testing.T) {
			transformer := NewTransformer(tt.modifierType)
			if transformer == nil {
				t.Fatal("NewTransformer returned nil")
			}

			actualType := getTypeName(transformer)
			if actualType != tt.expectedType {
				t.Errorf("NewTransformer(%q) returned %s, want %s",
					tt.modifierType, actualType, tt.expectedType)
			}
		})
	}
}

func getTypeName(v interface{}) string {
	switch v.(type) {
	case *HeikinAshiTransformer:
		return "*ticker.HeikinAshiTransformer"
	case *IdentityTransformer:
		return "*ticker.IdentityTransformer"
	default:
		return "unknown"
	}
}

func TestHeikinAshiTransformer_RealWorldScenario(t *testing.T) {
	transformer := &HeikinAshiTransformer{}

	bars := []context.OHLCV{
		{Open: 50000.0, High: 51000.0, Low: 49500.0, Close: 50500.0},
		{Open: 50500.0, High: 52000.0, Low: 50000.0, Close: 51500.0},
		{Open: 51500.0, High: 51800.0, Low: 50800.0, Close: 51000.0},
	}

	result := transformer.Transform(bars)

	if len(result) != 3 {
		t.Fatalf("Expected 3 bars, got %d", len(result))
	}

	for i, haBar := range result {
		if haBar.High < haBar.Open {
			t.Errorf("Bar %d: High (%v) < Open (%v)", i, haBar.High, haBar.Open)
		}
		if haBar.High < haBar.Close {
			t.Errorf("Bar %d: High (%v) < Close (%v)", i, haBar.High, haBar.Close)
		}
		if haBar.Low > haBar.Open {
			t.Errorf("Bar %d: Low (%v) > Open (%v)", i, haBar.Low, haBar.Open)
		}
		if haBar.Low > haBar.Close {
			t.Errorf("Bar %d: Low (%v) > Close (%v)", i, haBar.Low, haBar.Close)
		}
	}
}

func TestHeikinAshiTransformer_ComprehensiveEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		bars         []context.OHLCV
		validateBars func(t *testing.T, result []context.OHLCV)
	}{
		{
			name: "zero_value_bars",
			bars: []context.OHLCV{
				{Open: 0, High: 0, Low: 0, Close: 0, Volume: 100, Time: 1000},
				{Open: 0, High: 0, Low: 0, Close: 0, Volume: 200, Time: 2000},
			},
			validateBars: func(t *testing.T, result []context.OHLCV) {
				for i, bar := range result {
					if bar.Open != 0 || bar.High != 0 || bar.Low != 0 || bar.Close != 0 {
						t.Errorf("bar %d: expected all zeros, got OHLC=(%v,%v,%v,%v)",
							i, bar.Open, bar.High, bar.Low, bar.Close)
					}
				}
			},
		},
		{
			name: "negative_price_sequence",
			bars: []context.OHLCV{
				{Open: -10, High: -5, Low: -15, Close: -8, Volume: 100, Time: 1000},
				{Open: -8, High: -3, Low: -12, Close: -6, Volume: 200, Time: 2000},
			},
			validateBars: func(t *testing.T, result []context.OHLCV) {
				for i, bar := range result {
					if bar.High < bar.Low {
						t.Errorf("bar %d: haHigh=%v < haLow=%v", i, bar.High, bar.Low)
					}
				}
			},
		},
		{
			name: "extreme_large_sequence",
			bars: []context.OHLCV{
				{Open: 1e10, High: 1.1e10, Low: 0.9e10, Close: 1.05e10, Volume: 1000, Time: 1000},
				{Open: 1.05e10, High: 1.2e10, Low: 1.0e10, Close: 1.15e10, Volume: 2000, Time: 2000},
			},
			validateBars: func(t *testing.T, result []context.OHLCV) {
				if len(result) != 2 {
					t.Fatalf("expected 2 bars, got %d", len(result))
				}
			},
		},
		{
			name: "all_same_price_sequence",
			bars: []context.OHLCV{
				{Open: 100, High: 100, Low: 100, Close: 100, Volume: 100, Time: 1000},
				{Open: 100, High: 100, Low: 100, Close: 100, Volume: 200, Time: 2000},
				{Open: 100, High: 100, Low: 100, Close: 100, Volume: 300, Time: 3000},
			},
			validateBars: func(t *testing.T, result []context.OHLCV) {
				for i, bar := range result {
					if bar.Open != 100 || bar.High != 100 || bar.Low != 100 || bar.Close != 100 {
						t.Errorf("bar %d: expected all 100, got OHLC=(%v,%v,%v,%v)",
							i, bar.Open, bar.High, bar.Low, bar.Close)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := &HeikinAshiTransformer{}
			result := transformer.Transform(tt.bars)
			tt.validateBars(t, result)
		})
	}
}

func TestHeikinAshiTransformer_StateContinuity(t *testing.T) {
	barCount := 10
	bars := make([]context.OHLCV, barCount)
	basePrice := 100.0

	for i := 0; i < barCount; i++ {
		price := basePrice + float64(i)*5
		bars[i] = context.OHLCV{
			Open:   price,
			High:   price + 10,
			Low:    price - 5,
			Close:  price + 3,
			Volume: float64(100 + i*10),
			Time:   int64(1000 * (i + 1)),
		}
	}

	transformer := &HeikinAshiTransformer{}
	result := transformer.Transform(bars)

	if len(result) != barCount {
		t.Fatalf("expected %d bars, got %d", barCount, len(result))
	}

	for i := 1; i < len(result); i++ {
		expectedOpen := (result[i-1].Open + result[i-1].Close) / 2
		if math.Abs(result[i].Open-expectedOpen) > 0.0001 {
			t.Errorf("bar %d: state continuity broken, haOpen=%v, expected %v",
				i, result[i].Open, expectedOpen)
		}
	}

	for i := 0; i < len(result); i++ {
		if result[i].Volume != bars[i].Volume {
			t.Errorf("bar %d: volume changed from %v to %v", i, bars[i].Volume, result[i].Volume)
		}
		if result[i].Time != bars[i].Time {
			t.Errorf("bar %d: time changed from %v to %v", i, bars[i].Time, result[i].Time)
		}
	}
}

func TestHeikinAshiTransformer_StressTest(t *testing.T) {
	barCount := 1000
	bars := make([]context.OHLCV, barCount)

	for i := 0; i < barCount; i++ {
		bars[i] = context.OHLCV{
			Open:   100 + float64(i%100),
			High:   110 + float64(i%100),
			Low:    95 + float64(i%100),
			Close:  105 + float64(i%100),
			Volume: float64(1000 + i),
			Time:   int64(i * 1000),
		}
	}

	transformer := &HeikinAshiTransformer{}
	result := transformer.Transform(bars)

	if len(result) != barCount {
		t.Fatalf("expected %d bars, got %d", barCount, len(result))
	}

	checkIndices := []int{0, 100, 500, 999}
	for _, i := range checkIndices {
		bar := result[i]
		if bar.High < bar.Low {
			t.Errorf("bar %d: haHigh=%v < haLow=%v", i, bar.High, bar.Low)
		}
		if bar.Volume != bars[i].Volume {
			t.Errorf("bar %d: volume mismatch", i)
		}
	}
}

func TestHeikinAshiTransformer_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name string
		bars []context.OHLCV
	}{
		{
			name: "single_bar_various_volumes",
			bars: []context.OHLCV{
				{Open: 100, High: 110, Low: 95, Close: 105, Volume: 0, Time: 1000},
			},
		},
		{
			name: "two_bars_identical",
			bars: []context.OHLCV{
				{Open: 100, High: 110, Low: 95, Close: 105, Volume: 100, Time: 1000},
				{Open: 100, High: 110, Low: 95, Close: 105, Volume: 100, Time: 2000},
			},
		},
		{
			name: "three_bars_reverse_trend",
			bars: []context.OHLCV{
				{Open: 100, High: 110, Low: 95, Close: 108, Volume: 100, Time: 1000},
				{Open: 108, High: 115, Low: 100, Close: 102, Volume: 200, Time: 2000},
				{Open: 102, High: 105, Low: 90, Close: 92, Volume: 300, Time: 3000},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := &HeikinAshiTransformer{}
			result := transformer.Transform(tt.bars)

			if len(result) != len(tt.bars) {
				t.Errorf("length mismatch: got %d, want %d", len(result), len(tt.bars))
			}

			for i, bar := range result {
				if bar.High < bar.Open || bar.High < bar.Close {
					t.Errorf("bar %d: haHigh=%v violates invariant", i, bar.High)
				}
				if bar.Low > bar.Open || bar.Low > bar.Close {
					t.Errorf("bar %d: haLow=%v violates invariant", i, bar.Low)
				}
			}
		})
	}
}

func TestHeikinAshiTransformer_DataIntegrity(t *testing.T) {
	bars := []context.OHLCV{
		{Open: 100, High: 110, Low: 95, Close: 105, Volume: 12345, Time: 1609459200},
		{Open: 105, High: 115, Low: 100, Close: 112, Volume: 67890, Time: 1609545600},
		{Open: 112, High: 120, Low: 108, Close: 118, Volume: 11111, Time: 1609632000},
	}

	transformer := &HeikinAshiTransformer{}
	result := transformer.Transform(bars)

	for i := 0; i < len(bars); i++ {
		if result[i].Volume != bars[i].Volume {
			t.Errorf("bar %d: volume not preserved, got %v, want %v",
				i, result[i].Volume, bars[i].Volume)
		}
		if result[i].Time != bars[i].Time {
			t.Errorf("bar %d: time not preserved, got %v, want %v",
				i, result[i].Time, bars[i].Time)
		}
	}
}
