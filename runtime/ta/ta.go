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

	sum := 0.0
	validCount := 0
	firstValidIdx := -1

	for i := 0; i < len(source); i++ {
		result[i] = math.NaN()

		if !math.IsNaN(source[i]) {
			if firstValidIdx == -1 {
				firstValidIdx = i
			}
			sum += source[i]
			validCount++

			if validCount == period {
				result[i] = sum / float64(period)
				break
			}
		}
	}

	if validCount < period {
		return result
	}

	startIdx := firstValidIdx + validCount - 1

	for i := startIdx + 1; i < len(source); i++ {
		if !math.IsNaN(source[i]) {
			result[i] = alpha*source[i] + (1-alpha)*result[i-1]
		}
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

		// Calculate standard deviation with TV epsilon rounding (matches ta.stdev).
		sum := 0.0
		for j := 0; j < period; j++ {
			diff := tvStdevDiff(source[i-j], middle[i])
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

/* Stdev calculates standard deviation matching Pine ta.stdev (biased, population N).
 * Applies TV's epsilon-rounding compensation to each (source-mean) deviation:
 *   |diff| <= 1e-10 → 0   (cancels float drift from cancellation)
 *   |diff| <= 1e-4  → 1e-5 (clamps tiny non-zero diffs to a small constant)
 * This rule comes from Pine's reference implementation of ta.stdev and is what
 * lets marginal BB-band crossings agree with TradingView on the bar where a tiny
 * diff would otherwise flip a > vs < comparison. */
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

		// Calculate variance with TV epsilon rounding
		variance := 0.0
		for j := 0; j < period; j++ {
			diff := tvStdevDiff(source[i-j], mean)
			variance += diff * diff
		}
		variance /= float64(period)

		result[i] = math.Sqrt(variance)
	}
	return result
}

/* tvStdevDiff applies the TradingView ta.stdev per-deviation compensation rule:
 *   |diff| <= 1e-10 → 0     (treats sub-ULP differences as exact cancellations)
 *   |diff| <= 1e-4  → 1e-5  (clamps tiny non-zero magnitudes to a fixed floor)
 *   otherwise               → diff unchanged
 * This is published in the Pine reference implementation of ta.stdev. Applying
 * it makes the runner's stdev bit-identical to TV's where it matters for band
 * crossings. */
func tvStdevDiff(value, mean float64) float64 {
	diff := value - mean
	if math.Abs(diff) <= 1e-10 {
		return 0
	}
	if math.Abs(diff) <= 1e-4 {
		return 1e-5
	}
	return diff
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

/* Swma calculates Symmetrically Weighted Moving Average over fixed 4-bar window.
 * Weights: [1/6, 2/6, 2/6, 1/6] applied oldest-to-newest. */
func Swma(source []float64) []float64 {
	result := make([]float64, len(source))
	for i := range result {
		if i < 3 {
			result[i] = math.NaN()
			continue
		}
		result[i] = source[i-3]*(1.0/6.0) + source[i-2]*(2.0/6.0) + source[i-1]*(2.0/6.0) + source[i]*(1.0/6.0)
	}
	return result
}

/* Cci calculates Commodity Channel Index: (source - sma) / (0.015 * meanDev).
 * meanDev is the mean absolute deviation of source from its SMA over length bars. */
func Cci(source []float64, length int) []float64 {
	if length <= 0 || len(source) == 0 {
		result := make([]float64, len(source))
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}

	sma := Sma(source, length)
	result := make([]float64, len(source))

	for i := range result {
		if i < length-1 || math.IsNaN(sma[i]) {
			result[i] = math.NaN()
			continue
		}
		meanDev := 0.0
		for j := 0; j < length; j++ {
			meanDev += math.Abs(source[i-j] - sma[i])
		}
		meanDev /= float64(length)
		if meanDev == 0 {
			result[i] = 0
		} else {
			result[i] = (source[i] - sma[i]) / (0.015 * meanDev)
		}
	}
	return result
}

/* Bbw calculates Bollinger Bands Width: 2 * mult * stdev / sma over length bars. */
func Bbw(source []float64, length int, mult float64) []float64 {
	if length <= 0 || len(source) == 0 {
		result := make([]float64, len(source))
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}

	smaVals := Sma(source, length)
	stdevVals := Stdev(source, length)
	result := make([]float64, len(source))

	for i := range result {
		if math.IsNaN(smaVals[i]) || math.IsNaN(stdevVals[i]) || smaVals[i] == 0 {
			result[i] = math.NaN()
			continue
		}
		result[i] = 2.0 * mult * stdevVals[i] / smaVals[i]
	}
	return result
}

/* Cog calculates Center of Gravity oscillator: -sum(source[i]*(i+1)) / sum(source[i]).
 * Index 0 is the most recent bar, length-1 is the oldest within the window. */
func Cog(source []float64, length int) []float64 {
	if length <= 0 || len(source) == 0 {
		result := make([]float64, len(source))
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}

	result := make([]float64, len(source))

	for i := range result {
		if i < length-1 {
			result[i] = math.NaN()
			continue
		}
		num, denom := 0.0, 0.0
		for j := 0; j < length; j++ {
			num += source[i-j] * float64(j+1)
			denom += source[i-j]
		}
		if denom == 0 {
			result[i] = 0
		} else {
			result[i] = -num / denom
		}
	}
	return result
}

/* Tsi calculates True Strength Index: 100 * doubleEMA(momentum) / doubleEMA(|momentum|).
 * momentum = source - source[1]; short/long EMA periods control smoothing. */
func Tsi(source []float64, shortLength, longLength int) []float64 {
	if shortLength <= 0 || longLength <= 0 || len(source) < 2 {
		result := make([]float64, len(source))
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}

	momentum := make([]float64, len(source))
	absMomentum := make([]float64, len(source))
	momentum[0] = math.NaN()
	absMomentum[0] = math.NaN()
	for i := 1; i < len(source); i++ {
		m := source[i] - source[i-1]
		momentum[i] = m
		absMomentum[i] = math.Abs(m)
	}

	smoothed := Ema(Ema(momentum, longLength), shortLength)
	smoothedAbs := Ema(Ema(absMomentum, longLength), shortLength)

	result := make([]float64, len(source))
	for i := range result {
		if math.IsNaN(smoothed[i]) || math.IsNaN(smoothedAbs[i]) || smoothedAbs[i] == 0 {
			result[i] = math.NaN()
			continue
		}
		result[i] = 100.0 * smoothed[i] / smoothedAbs[i]
	}
	return result
}

/* Cum calculates cumulative sum of source (PineScript compatible) */
func Cum(source []float64) []float64 {
	if len(source) == 0 {
		return source
	}

	result := make([]float64, len(source))
	cumSum := 0.0

	for i := range source {
		if math.IsNaN(source[i]) {
			result[i] = math.NaN()
		} else {
			cumSum += source[i]
			result[i] = cumSum
		}
	}

	return result
}
