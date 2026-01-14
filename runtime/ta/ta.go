package ta

import (
	"math"
)

/* Sma calculates Simple Moving Average (PineTS compatible) */
func Sma(source []float64, period int) []float64 {
	if period <= 0 || len(source) == 0 {
		return source
	}

	result := make([]float64, len(source))
	for i := range result {
		if i < period-1 {
			result[i] = math.NaN()
			continue
		}
		sum := 0.0
		for j := 0; j < period; j++ {
			sum += source[i-j]
		}
		result[i] = sum / float64(period)
	}
	return result
}

/* Ema calculates Exponential Moving Average (PineTS compatible) */
func Ema(source []float64, period int) []float64 {
	if period <= 0 || len(source) == 0 {
		return source
	}

	result := make([]float64, len(source))
	multiplier := 2.0 / float64(period+1)

	// Find first non-NaN values and calculate initial SMA
	validCount := 0
	sum := 0.0
	startIdx := -1

	for i := 0; i < len(source); i++ {
		result[i] = math.NaN()

		if !math.IsNaN(source[i]) {
			if startIdx == -1 {
				startIdx = i
			}
			sum += source[i]
			validCount++

			if validCount == period {
				result[i] = sum / float64(period)
				startIdx = i
				break
			}
		}
	}

	// EMA calculation for remaining values
	if startIdx >= 0 && startIdx < len(source)-1 {
		for i := startIdx + 1; i < len(source); i++ {
			if !math.IsNaN(source[i]) {
				result[i] = (source[i]-result[i-1])*multiplier + result[i-1]
			} else {
				result[i] = math.NaN()
			}
		}
	}

	return result
}

/* Rma calculates Relative Moving Average (PineTS compatible) */
func Rma(source []float64, period int) []float64 {
	if period <= 0 || len(source) == 0 {
		return source
	}

	result := make([]float64, len(source))
	alpha := 1.0 / float64(period)

	// First value is SMA
	sum := 0.0
	for i := 0; i < period; i++ {
		if i >= len(source) {
			result[i] = math.NaN()
			continue
		}
		result[i] = math.NaN()
		sum += source[i]
	}

	if period <= len(source) {
		result[period-1] = sum / float64(period)
	}

	// RMA calculation
	for i := period; i < len(source); i++ {
		result[i] = alpha*source[i] + (1-alpha)*result[i-1]
	}

	return result
}

/* Rsi calculates Relative Strength Index (PineTS compatible) */
func Rsi(source []float64, period int) []float64 {
	if period <= 0 || len(source) < 2 {
		result := make([]float64, len(source))
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}

	// Calculate price changes
	changes := make([]float64, len(source))
	changes[0] = math.NaN()
	for i := 1; i < len(source); i++ {
		changes[i] = source[i] - source[i-1]
	}

	// Separate gains and losses
	gains := make([]float64, len(changes))
	losses := make([]float64, len(changes))
	for i := range changes {
		if math.IsNaN(changes[i]) {
			gains[i] = 0
			losses[i] = 0
		} else if changes[i] > 0 {
			gains[i] = changes[i]
			losses[i] = 0
		} else {
			gains[i] = 0
			losses[i] = -changes[i]
		}
	}

	// Calculate RMA of gains and losses
	avgGain := Rma(gains, period)
	avgLoss := Rma(losses, period)

	// Calculate RSI
	result := make([]float64, len(source))
	for i := range result {
		if math.IsNaN(avgGain[i]) || math.IsNaN(avgLoss[i]) {
			result[i] = math.NaN()
		} else if avgLoss[i] == 0 {
			result[i] = 100.0
		} else {
			rs := avgGain[i] / avgLoss[i]
			result[i] = 100.0 - (100.0 / (1.0 + rs))
		}
	}

	return result
}

/* Tr calculates True Range (PineTS compatible) */
func Tr(high, low, close []float64) []float64 {
	if len(high) == 0 || len(low) == 0 || len(close) == 0 {
		return []float64{}
	}

	minLen := len(high)
	if len(low) < minLen {
		minLen = len(low)
	}
	if len(close) < minLen {
		minLen = len(close)
	}

	result := make([]float64, minLen)

	// First bar: high - low
	result[0] = high[0] - low[0]

	// Subsequent bars: max(high-low, abs(high-prevClose), abs(low-prevClose))
	for i := 1; i < minLen; i++ {
		hl := high[i] - low[i]
		hc := math.Abs(high[i] - close[i-1])
		lc := math.Abs(low[i] - close[i-1])

		result[i] = math.Max(hl, math.Max(hc, lc))
	}

	return result
}

/* Atr calculates Average True Range (PineTS compatible) */
func Atr(high, low, close []float64, period int) []float64 {
	tr := Tr(high, low, close)
	return Rma(tr, period)
}

/* BBands calculates Bollinger Bands (upper, middle, lower) */
func BBands(source []float64, period int, stdDev float64) ([]float64, []float64, []float64) {
	middle := Sma(source, period)

	upper := make([]float64, len(source))
	lower := make([]float64, len(source))

	for i := range source {
		if i < period-1 {
			upper[i] = math.NaN()
			lower[i] = math.NaN()
			continue
		}

		// Calculate standard deviation
		sum := 0.0
		for j := 0; j < period; j++ {
			diff := source[i-j] - middle[i]
			sum += diff * diff
		}
		std := math.Sqrt(sum / float64(period))

		upper[i] = middle[i] + stdDev*std
		lower[i] = middle[i] - stdDev*std
	}

	return upper, middle, lower
}

/* Macd calculates MACD (macd, signal, histogram) */
func Macd(source []float64, fastPeriod, slowPeriod, signalPeriod int) ([]float64, []float64, []float64) {
	fastEma := Ema(source, fastPeriod)
	slowEma := Ema(source, slowPeriod)

	macd := make([]float64, len(source))
	for i := range source {
		if math.IsNaN(fastEma[i]) || math.IsNaN(slowEma[i]) {
			macd[i] = math.NaN()
		} else {
			macd[i] = fastEma[i] - slowEma[i]
		}
	}

	signal := Ema(macd, signalPeriod)

	histogram := make([]float64, len(source))
	for i := range source {
		if math.IsNaN(macd[i]) || math.IsNaN(signal[i]) {
			histogram[i] = math.NaN()
		} else {
			histogram[i] = macd[i] - signal[i]
		}
	}

	return macd, signal, histogram
}

/* Stoch calculates Stochastic Oscillator (k, d) */
func Stoch(high, low, close []float64, kPeriod, dPeriod int) ([]float64, []float64) {
	minLen := len(high)
	if len(low) < minLen {
		minLen = len(low)
	}
	if len(close) < minLen {
		minLen = len(close)
	}

	k := make([]float64, minLen)

	for i := range k {
		if i < kPeriod-1 {
			k[i] = math.NaN()
			continue
		}

		// Find highest high and lowest low in period
		highestHigh := high[i]
		lowestLow := low[i]
		for j := 1; j < kPeriod; j++ {
			if high[i-j] > highestHigh {
				highestHigh = high[i-j]
			}
			if low[i-j] < lowestLow {
				lowestLow = low[i-j]
			}
		}

		if highestHigh == lowestLow {
			k[i] = 50.0
		} else {
			k[i] = 100.0 * (close[i] - lowestLow) / (highestHigh - lowestLow)
		}
	}

	// Calculate %D as SMA of %K
	d := Sma(k, dPeriod)

	return k, d
}

/* Stdev calculates standard deviation (PineTS compatible) */
func Stdev(source []float64, period int) []float64 {
	if period <= 0 || len(source) == 0 {
		return source
	}

	result := make([]float64, len(source))
	for i := range result {
		if i < period-1 {
			result[i] = math.NaN()
			continue
		}

		// Calculate mean
		sum := 0.0
		for j := 0; j < period; j++ {
			sum += source[i-j]
		}
		mean := sum / float64(period)

		// Calculate variance
		variance := 0.0
		for j := 0; j < period; j++ {
			diff := source[i-j] - mean
			variance += diff * diff
		}
		variance /= float64(period)

		result[i] = math.Sqrt(variance)
	}
	return result
}

/* Change calculates bar-to-bar difference (source - source[1]) (PineTS compatible) */
func Change(source []float64) []float64 {
	if len(source) == 0 {
		return source
	}

	result := make([]float64, len(source))
	result[0] = math.NaN()

	for i := 1; i < len(source); i++ {
		if math.IsNaN(source[i]) || math.IsNaN(source[i-1]) {
			result[i] = math.NaN()
		} else {
			result[i] = source[i] - source[i-1]
		}
	}
	return result
}

/* Pivothigh detects pivot high points (local maxima) (PineTS compatible) */
func Pivothigh(source []float64, leftBars, rightBars int) []float64 {
	if len(source) == 0 || leftBars < 0 || rightBars < 0 {
		return source
	}

	result := make([]float64, len(source))
	for i := range result {
		result[i] = math.NaN()
	}

	// Need leftBars before and rightBars after current bar
	for i := leftBars; i < len(source)-rightBars; i++ {
		isPivot := true
		center := source[i]

		if math.IsNaN(center) {
			continue
		}

		// Check left bars - all must be less than or equal to center
		for j := 1; j <= leftBars; j++ {
			if math.IsNaN(source[i-j]) || source[i-j] > center {
				isPivot = false
				break
			}
		}

		// Check right bars - all must be less than or equal to center
		if isPivot {
			for j := 1; j <= rightBars; j++ {
				if math.IsNaN(source[i+j]) || source[i+j] > center {
					isPivot = false
					break
				}
			}
		}

		if isPivot {
			result[i] = center
		}
	}

	return result
}

/* Pivotlow detects pivot low points (local minima) (PineTS compatible) */
func Pivotlow(source []float64, leftBars, rightBars int) []float64 {
	if len(source) == 0 || leftBars < 0 || rightBars < 0 {
		return source
	}

	result := make([]float64, len(source))
	for i := range result {
		result[i] = math.NaN()
	}

	// Need leftBars before and rightBars after current bar
	for i := leftBars; i < len(source)-rightBars; i++ {
		isPivot := true
		center := source[i]

		if math.IsNaN(center) {
			continue
		}

		// Check left bars - all must be greater than or equal to center
		for j := 1; j <= leftBars; j++ {
			if math.IsNaN(source[i-j]) || source[i-j] < center {
				isPivot = false
				break
			}
		}

		// Check right bars - all must be greater than or equal to center
		if isPivot {
			for j := 1; j <= rightBars; j++ {
				if math.IsNaN(source[i+j]) || source[i+j] < center {
					isPivot = false
					break
				}
			}
		}

		if isPivot {
			result[i] = center
		}
	}

	return result
}
