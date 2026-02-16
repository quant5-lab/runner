package ta

import "math"

/* Dmi calculates Directional Movement Index
 * Returns [+DI, -DI, ADX] tuple following PineScript ta.dmi() specification
 * diLength: DI calculation period
 * adxSmoothing: ADX smoothing period */
func Dmi(high, low, close []float64, diLength, adxSmoothing int) ([]float64, []float64, []float64) {
	n := len(high)
	if n == 0 || diLength <= 0 || adxSmoothing <= 0 {
		return make([]float64, n), make([]float64, n), make([]float64, n)
	}

	plusDM := make([]float64, n)
	minusDM := make([]float64, n)
	tr := Tr(high, low, close)

	// Initialize first element as 0 (no movement on first bar)
	plusDM[0] = 0
	minusDM[0] = 0

	for i := 1; i < n; i++ {
		upMove := high[i] - high[i-1]
		downMove := low[i-1] - low[i]

		if upMove > downMove && upMove > 0 {
			plusDM[i] = upMove
		} else {
			plusDM[i] = 0
		}

		if downMove > upMove && downMove > 0 {
			minusDM[i] = downMove
		} else {
			minusDM[i] = 0
		}
	}

	smoothedPlusDM := Rma(plusDM, diLength)
	smoothedMinusDM := Rma(minusDM, diLength)
	smoothedTR := Rma(tr, diLength)

	plusDI := make([]float64, n)
	minusDI := make([]float64, n)
	for i := 0; i < n; i++ {
		plusDI[i] = math.NaN()
		minusDI[i] = math.NaN()

		if !math.IsNaN(smoothedTR[i]) && smoothedTR[i] != 0 {
			plusDI[i] = (smoothedPlusDM[i] / smoothedTR[i]) * 100
			minusDI[i] = (smoothedMinusDM[i] / smoothedTR[i]) * 100
		}
	}

	dx := make([]float64, n)
	for i := 0; i < n; i++ {
		dx[i] = math.NaN()
		if !math.IsNaN(plusDI[i]) && !math.IsNaN(minusDI[i]) {
			sum := plusDI[i] + minusDI[i]
			if sum != 0 {
				dx[i] = math.Abs(plusDI[i]-minusDI[i]) / sum * 100
			}
		}
	}

	adx := Rma(dx, adxSmoothing)

	return plusDI, minusDI, adx
}
