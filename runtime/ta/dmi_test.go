package ta

import (
	"math"
	"testing"
)

func TestDmi_OutputLength(t *testing.T) {
	tests := []struct {
		name      string
		inputSize int
		diLength  int
		adxSmooth int
	}{
		{"standard_period", 50, 14, 14},
		{"short_period", 20, 5, 3},
		{"long_period", 100, 21, 21},
		{"minimum_data", 3, 2, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			high := make([]float64, tt.inputSize)
			low := make([]float64, tt.inputSize)
			close := make([]float64, tt.inputSize)

			for i := range high {
				high[i] = float64(10 + i)
				low[i] = float64(9 + i)
				close[i] = float64(9.5 + float64(i))
			}

			plusDI, minusDI, adx := Dmi(high, low, close, tt.diLength, tt.adxSmooth)

			if len(plusDI) != tt.inputSize {
				t.Errorf("plusDI length = %d, want %d", len(plusDI), tt.inputSize)
			}
			if len(minusDI) != tt.inputSize {
				t.Errorf("minusDI length = %d, want %d", len(minusDI), tt.inputSize)
			}
			if len(adx) != tt.inputSize {
				t.Errorf("adx length = %d, want %d", len(adx), tt.inputSize)
			}
		})
	}
}

func TestDmi_WarmupPeriod(t *testing.T) {
	high := make([]float64, 50)
	low := make([]float64, 50)
	close := make([]float64, 50)

	for i := range high {
		high[i] = float64(10 + i)
		low[i] = float64(9 + i)
		close[i] = float64(9.5 + float64(i))
	}

	plusDI, minusDI, adx := Dmi(high, low, close, 14, 14)

	for i := 0; i < 13; i++ {
		if !math.IsNaN(plusDI[i]) {
			t.Errorf("plusDI[%d] should be NaN during warmup", i)
		}
		if !math.IsNaN(minusDI[i]) {
			t.Errorf("minusDI[%d] should be NaN during warmup", i)
		}
	}

	adxWarmup := 13 + 14 - 1
	for i := 0; i < adxWarmup && i < len(adx); i++ {
		if !math.IsNaN(adx[i]) {
			t.Errorf("adx[%d] should be NaN during warmup", i)
		}
	}
}

func TestDmi_ValueRange(t *testing.T) {
	high := make([]float64, 50)
	low := make([]float64, 50)
	close := make([]float64, 50)

	for i := range high {
		high[i] = float64(100 + i*2)
		low[i] = float64(98 + i*2)
		close[i] = float64(99 + float64(i)*2)
	}

	plusDI, minusDI, adx := Dmi(high, low, close, 14, 14)

	for i := 13; i < len(plusDI); i++ {
		if !math.IsNaN(plusDI[i]) && (plusDI[i] < 0 || plusDI[i] > 100) {
			t.Errorf("plusDI[%d] = %f, should be between 0 and 100", i, plusDI[i])
		}
		if !math.IsNaN(minusDI[i]) && (minusDI[i] < 0 || minusDI[i] > 100) {
			t.Errorf("minusDI[%d] = %f, should be between 0 and 100", i, minusDI[i])
		}
	}

	for i := 26; i < len(adx); i++ {
		if !math.IsNaN(adx[i]) && (adx[i] < 0 || adx[i] > 100) {
			t.Errorf("adx[%d] = %f, should be between 0 and 100", i, adx[i])
		}
	}
}

func TestDmi_UpwardTrend(t *testing.T) {
	n := 50
	high := make([]float64, n)
	low := make([]float64, n)
	close := make([]float64, n)

	for i := range high {
		high[i] = float64(100 + i*3)
		low[i] = float64(98 + i*3)
		close[i] = float64(99 + i*3)
	}

	plusDI, minusDI, _ := Dmi(high, low, close, 14, 14)

	upwardCount := 0
	for i := 20; i < len(plusDI); i++ {
		if !math.IsNaN(plusDI[i]) && !math.IsNaN(minusDI[i]) {
			if plusDI[i] > minusDI[i] {
				upwardCount++
			}
		}
	}

	if upwardCount == 0 {
		t.Error("Expected +DI > -DI in strong upward trend")
	}
}

func TestDmi_DownwardTrend(t *testing.T) {
	n := 50
	high := make([]float64, n)
	low := make([]float64, n)
	close := make([]float64, n)

	for i := range high {
		high[i] = float64(200 - i*3)
		low[i] = float64(198 - i*3)
		close[i] = float64(199 - i*3)
	}

	plusDI, minusDI, _ := Dmi(high, low, close, 14, 14)

	downwardCount := 0
	for i := 20; i < len(plusDI); i++ {
		if !math.IsNaN(plusDI[i]) && !math.IsNaN(minusDI[i]) {
			if minusDI[i] > plusDI[i] {
				downwardCount++
			}
		}
	}

	if downwardCount == 0 {
		t.Error("Expected -DI > +DI in strong downward trend")
	}
}

func TestDmi_EdgeCases(t *testing.T) {
	t.Run("empty_input", func(t *testing.T) {
		plusDI, minusDI, adx := Dmi([]float64{}, []float64{}, []float64{}, 14, 14)
		if len(plusDI) != 0 || len(minusDI) != 0 || len(adx) != 0 {
			t.Error("empty input should return empty arrays")
		}
	})

	t.Run("zero_diLength", func(t *testing.T) {
		high := []float64{10, 11, 12}
		low := []float64{9, 10, 11}
		close := []float64{9.5, 10.5, 11.5}

		plusDI, minusDI, _ := Dmi(high, low, close, 0, 5)

		if len(plusDI) != 3 || len(minusDI) != 3 {
			t.Error("zero diLength should return arrays of input length")
		}
	})

	t.Run("zero_adxSmoothing", func(t *testing.T) {
		high := []float64{10, 11, 12, 13, 14}
		low := []float64{9, 10, 11, 12, 13}
		close := []float64{9.5, 10.5, 11.5, 12.5, 13.5}

		plusDI, minusDI, adx := Dmi(high, low, close, 2, 0)

		if len(plusDI) != 5 || len(minusDI) != 5 || len(adx) != 5 {
			t.Error("zero adxSmoothing should return arrays of input length")
		}
	})

	t.Run("constant_prices", func(t *testing.T) {
		n := 20
		high := make([]float64, n)
		low := make([]float64, n)
		close := make([]float64, n)

		for i := range high {
			high[i] = 100.0
			low[i] = 100.0
			close[i] = 100.0
		}

		plusDI, minusDI, _ := Dmi(high, low, close, 5, 5)

		for i := 10; i < len(plusDI); i++ {
			if !math.IsNaN(plusDI[i]) && plusDI[i] > 1.0 {
				t.Errorf("plusDI[%d] = %f, expected near-zero for constant prices", i, plusDI[i])
			}
			if !math.IsNaN(minusDI[i]) && minusDI[i] > 1.0 {
				t.Errorf("minusDI[%d] = %f, expected near-zero for constant prices", i, minusDI[i])
			}
		}
	})
}
