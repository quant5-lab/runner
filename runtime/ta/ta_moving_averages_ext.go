package ta

import (
	"math"
)

// Wma computes Weighted Moving Average.
// Pine: ta.wma(source, length)
func Wma(source []float64, period int) []float64 {
	if period <= 0 || len(source) == 0 {
		return source
	}

	result := make([]float64, len(source))
	for i := range result {
		if i < period-1 {
			result[i] = math.NaN()
			continue
		}
		sum, weightSum := 0.0, 0.0
		for j := 0; j < period; j++ {
			weight := float64(period - j)
			sum += weight * source[i-j]
			weightSum += weight
		}
		result[i] = sum / weightSum
	}
	return result
}

// Alma computes the Arnaud Legoux Moving Average.
// offset controls the center of the Gaussian bell (default: 0.85).
// sigma controls the width of the bell (default: 6 — larger = closer to SMA, smaller = sharper).
// Pine: ta.alma(source, length, offset=0.85, sigma=6)
func Alma(source []float64, period int, offset, sigma float64) []float64 {
	if period <= 0 || len(source) == 0 {
		return source
	}

	m := offset * float64(period-1)
	s := float64(period) / sigma
	weights := make([]float64, period)
	wSum := 0.0
	for j := 0; j < period; j++ {
		d := float64(j) - m
		weights[j] = math.Exp(-(d * d) / (2 * s * s))
		wSum += weights[j]
	}

	result := make([]float64, len(source))
	for i := range result {
		if i < period-1 {
			result[i] = math.NaN()
			continue
		}
		val := 0.0
		for j := 0; j < period; j++ {
			val += weights[j] * source[i-period+1+j]
		}
		result[i] = val / wSum
	}
	return result
}

// Hma computes the Hull Moving Average: WMA(2*WMA(src, n/2) - WMA(src, n), round(sqrt(n))).
// Designed to reduce lag compared to WMA while preserving smoothness.
// Pine: ta.hma(source, length)
func Hma(source []float64, period int) []float64 {
	if period <= 1 || len(source) == 0 {
		return source
	}

	halfPeriod := period / 2
	sqrtPeriod := int(math.Round(math.Sqrt(float64(period))))

	wma1 := Wma(source, halfPeriod)
	wma2 := Wma(source, period)

	diff := make([]float64, len(source))
	for i := range diff {
		if math.IsNaN(wma1[i]) || math.IsNaN(wma2[i]) {
			diff[i] = math.NaN()
		} else {
			diff[i] = 2*wma1[i] - wma2[i]
		}
	}

	return Wma(diff, sqrtPeriod)
}

// Kcw computes the Keltner Channel Width: 2 * mult * atr(length) / ema(source, length).
// Pine: ta.kcw(source, length, mult=1.5)
func Kcw(source, high, low, closeVals []float64, period int, mult float64) []float64 {
	if period <= 0 || len(source) == 0 {
		return source
	}

	emaVals := Ema(source, period)
	atrVals := Atr(high, low, closeVals, period)

	result := make([]float64, len(source))
	for i := range result {
		if math.IsNaN(emaVals[i]) || math.IsNaN(atrVals[i]) {
			result[i] = math.NaN()
			continue
		}
		if emaVals[i] == 0 {
			result[i] = 0
			continue
		}
		result[i] = 2 * mult * atrVals[i] / emaVals[i]
	}
	return result
}

// Sar computes the Parabolic Stop-and-Reverse indicator.
// start is the initial acceleration factor, inc is the step increment, max is the maximum AF.
// Pine: ta.sar(start, inc, max)
func Sar(high, low []float64, start, inc, max float64) []float64 {
	n := len(high)
	if n == 0 {
		return nil
	}

	result := make([]float64, n)
	result[0] = math.NaN()
	if n == 1 {
		return result
	}

	// Initialize: detect initial trend from first two bars
	isUptrend := high[1] >= high[0]
	var sar, ep float64
	af := start

	if isUptrend {
		sar = low[0]
		ep = high[0]
	} else {
		sar = high[0]
		ep = low[0]
	}
	result[0] = sar

	for i := 1; i < n; i++ {
		projectedSAR := sar + af*(ep-sar)

		if isUptrend {
			if i >= 2 && projectedSAR > low[i-2] {
				projectedSAR = low[i-2]
			}
			if projectedSAR > low[i-1] {
				projectedSAR = low[i-1]
			}
			if low[i] < projectedSAR {
				isUptrend = false
				projectedSAR = ep
				ep = low[i]
				af = start
			} else {
				if high[i] > ep {
					ep = high[i]
					af = math.Min(af+inc, max)
				}
			}
		} else {
			if i >= 2 && projectedSAR < high[i-2] {
				projectedSAR = high[i-2]
			}
			if projectedSAR < high[i-1] {
				projectedSAR = high[i-1]
			}
			if high[i] > projectedSAR {
				isUptrend = true
				projectedSAR = ep
				ep = high[i]
				af = start
			} else {
				if low[i] < ep {
					ep = low[i]
					af = math.Min(af+inc, max)
				}
			}
		}

		sar = projectedSAR
		result[i] = sar
	}

	return result
}
