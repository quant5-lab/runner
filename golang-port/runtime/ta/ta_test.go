package ta

import (
	"math"
	"testing"
)

func floatSliceEqual(a, b []float64, tolerance float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.IsNaN(a[i]) && math.IsNaN(b[i]) {
			continue
		}
		if math.Abs(a[i]-b[i]) > tolerance {
			return false
		}
	}
	return true
}

func TestSma(t *testing.T) {
	source := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result := Sma(source, 3)

	if len(result) != len(source) {
		t.Fatalf("Sma length = %d, want %d", len(result), len(source))
	}

	// First 2 values should be NaN (period-1)
	if !math.IsNaN(result[0]) || !math.IsNaN(result[1]) {
		t.Error("First 2 values should be NaN")
	}

	// SMA(3) at index 2: (1+2+3)/3 = 2
	if math.Abs(result[2]-2.0) > 0.0001 {
		t.Errorf("Sma[2] = %f, want 2.0", result[2])
	}

	// SMA(3) at index 9: (8+9+10)/3 = 9
	if math.Abs(result[9]-9.0) > 0.0001 {
		t.Errorf("Sma[9] = %f, want 9.0", result[9])
	}
}

func TestEma(t *testing.T) {
	source := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result := Ema(source, 3)

	if len(result) != len(source) {
		t.Fatalf("Ema length = %d, want %d", len(result), len(source))
	}

	// First 2 values should be NaN (period-1)
	if !math.IsNaN(result[0]) || !math.IsNaN(result[1]) {
		t.Error("First 2 values should be NaN")
	}

	// EMA should exist from index 2 onwards
	if math.IsNaN(result[2]) {
		t.Error("Ema[2] should have value")
	}
}

func TestRsi(t *testing.T) {
	source := []float64{44, 44.34, 44.09, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28}
	result := Rsi(source, 14)

	if len(result) != len(source) {
		t.Fatalf("Rsi length = %d, want %d", len(result), len(source))
	}

	// First 13 values should be NaN (period-1)
	for i := 0; i < 13; i++ {
		if !math.IsNaN(result[i]) {
			t.Errorf("Rsi[%d] should be NaN", i)
		}
	}

	// RSI should be between 0 and 100
	if !math.IsNaN(result[13]) && (result[13] < 0 || result[13] > 100) {
		t.Errorf("Rsi[13] = %f, should be between 0 and 100", result[13])
	}
}

func TestAtr(t *testing.T) {
	high := []float64{48.70, 48.72, 48.90, 48.87, 48.82}
	low := []float64{47.79, 48.14, 48.39, 48.37, 48.24}
	close := []float64{48.16, 48.61, 48.75, 48.63, 48.74}

	result := Atr(high, low, close, 3)

	if len(result) != len(high) {
		t.Fatalf("Atr length = %d, want %d", len(result), len(high))
	}

	// First 2 values should be NaN
	if !math.IsNaN(result[0]) || !math.IsNaN(result[1]) {
		t.Error("First 2 values should be NaN")
	}

	// ATR should be positive
	if result[4] <= 0 {
		t.Errorf("Atr[4] = %f, should be positive", result[4])
	}
}

func TestBBands(t *testing.T) {
	source := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	upper, middle, lower := BBands(source, 3, 2.0)

	if len(upper) != len(source) || len(middle) != len(source) || len(lower) != len(source) {
		t.Fatal("BBands output length mismatch")
	}

	// First 2 values should be NaN
	if !math.IsNaN(upper[0]) || !math.IsNaN(middle[0]) || !math.IsNaN(lower[0]) {
		t.Error("First values should be NaN")
	}

	// Middle band should equal SMA
	smaResult := Sma(source, 3)
	for i := 2; i < len(source); i++ {
		if math.Abs(middle[i]-smaResult[i]) > 0.0001 {
			t.Errorf("Middle[%d] = %f, want %f (SMA)", i, middle[i], smaResult[i])
		}
	}

	// Upper > Middle > Lower
	for i := 2; i < len(source); i++ {
		if upper[i] <= middle[i] || middle[i] <= lower[i] {
			t.Errorf("At index %d: upper=%f, middle=%f, lower=%f (wrong order)", i, upper[i], middle[i], lower[i])
		}
	}
}

func TestMacd(t *testing.T) {
	source := make([]float64, 50)
	for i := range source {
		source[i] = float64(i + 1)
	}

	macd, signal, histogram := Macd(source, 12, 26, 9)

	if len(macd) != len(source) || len(signal) != len(source) || len(histogram) != len(source) {
		t.Fatal("Macd output length mismatch")
	}

	// First 25 values should be NaN (slowPeriod-1)
	for i := 0; i < 25; i++ {
		if !math.IsNaN(macd[i]) {
			t.Errorf("Macd[%d] should be NaN", i)
		}
	}

	// Check last value has all components
	lastIdx := len(source) - 1
	if math.IsNaN(macd[lastIdx]) || math.IsNaN(signal[lastIdx]) || math.IsNaN(histogram[lastIdx]) {
		t.Error("Last MACD values should not be NaN")
	}
}

func TestStoch(t *testing.T) {
	high := make([]float64, 20)
	low := make([]float64, 20)
	close := make([]float64, 20)

	for i := range high {
		high[i] = float64(i + 10)
		low[i] = float64(i)
		close[i] = float64(i + 5)
	}

	k, d := Stoch(high, low, close, 14, 3)

	if len(k) != len(high) || len(d) != len(high) {
		t.Fatal("Stoch output length mismatch")
	}

	// First values should be NaN
	if !math.IsNaN(k[0]) || !math.IsNaN(d[0]) {
		t.Error("First Stoch values should be NaN")
	}

	// Stochastic should be between 0 and 100
	for i := 14; i < len(k); i++ {
		if !math.IsNaN(k[i]) && (k[i] < 0 || k[i] > 100) {
			t.Errorf("Stoch K[%d] = %f, should be between 0 and 100", i, k[i])
		}
		if !math.IsNaN(d[i]) && (d[i] < 0 || d[i] > 100) {
			t.Errorf("Stoch D[%d] = %f, should be between 0 and 100", i, d[i])
		}
	}
}
