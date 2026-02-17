package ta

import (
	"math"
	"sort"
)

/* Max returns maximum value over period (PineScript compatible) */
func Max(source []float64, period int) []float64 {
	if period <= 0 || len(source) == 0 {
		return source
	}

	result := make([]float64, len(source))
	for i := range result {
		if i < period-1 {
			result[i] = math.NaN()
			continue
		}

		maxVal := math.Inf(-1)
		hasNaN := false

		for j := 0; j < period; j++ {
			val := source[i-j]
			if math.IsNaN(val) {
				hasNaN = true
				break
			}
			if val > maxVal {
				maxVal = val
			}
		}

		if hasNaN {
			result[i] = math.NaN()
		} else {
			result[i] = maxVal
		}
	}

	return result
}

/* Min returns minimum value over period (PineScript compatible) */
func Min(source []float64, period int) []float64 {
	if period <= 0 || len(source) == 0 {
		return source
	}

	result := make([]float64, len(source))
	for i := range result {
		if i < period-1 {
			result[i] = math.NaN()
			continue
		}

		minVal := math.Inf(1)
		hasNaN := false

		for j := 0; j < period; j++ {
			val := source[i-j]
			if math.IsNaN(val) {
				hasNaN = true
				break
			}
			if val < minVal {
				minVal = val
			}
		}

		if hasNaN {
			result[i] = math.NaN()
		} else {
			result[i] = minVal
		}
	}

	return result
}

/* Median returns median value over period (PineScript compatible) */
func Median(source []float64, period int) []float64 {
	if period <= 0 || len(source) == 0 {
		return source
	}

	result := make([]float64, len(source))
	window := make([]float64, period)

	for i := range result {
		if i < period-1 {
			result[i] = math.NaN()
			continue
		}

		hasNaN := false
		for j := 0; j < period; j++ {
			val := source[i-j]
			if math.IsNaN(val) {
				hasNaN = true
				break
			}
			window[j] = val
		}

		if hasNaN {
			result[i] = math.NaN()
			continue
		}

		sorted := make([]float64, period)
		copy(sorted, window)
		sort.Float64s(sorted)

		if period%2 == 0 {
			result[i] = (sorted[period/2-1] + sorted[period/2]) / 2.0
		} else {
			result[i] = sorted[period/2]
		}
	}

	return result
}

/* Variance returns variance over period (PineScript compatible) */
func Variance(source []float64, period int) []float64 {
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
		hasNaN := false

		for j := 0; j < period; j++ {
			val := source[i-j]
			if math.IsNaN(val) {
				hasNaN = true
				break
			}
			sum += val
		}

		if hasNaN {
			result[i] = math.NaN()
			continue
		}

		mean := sum / float64(period)
		sumSquaredDiff := 0.0

		for j := 0; j < period; j++ {
			diff := source[i-j] - mean
			sumSquaredDiff += diff * diff
		}

		result[i] = sumSquaredDiff / float64(period)
	}

	return result
}

/* Range returns difference between max and min over period (PineScript compatible) */
func Range(source []float64, period int) []float64 {
	if period <= 0 || len(source) == 0 {
		return source
	}

	result := make([]float64, len(source))

	for i := range result {
		if i < period-1 {
			result[i] = math.NaN()
			continue
		}

		minVal := math.Inf(1)
		maxVal := math.Inf(-1)
		hasNaN := false

		for j := 0; j < period; j++ {
			val := source[i-j]
			if math.IsNaN(val) {
				hasNaN = true
				break
			}
			if val < minVal {
				minVal = val
			}
			if val > maxVal {
				maxVal = val
			}
		}

		if hasNaN {
			result[i] = math.NaN()
		} else {
			result[i] = maxVal - minVal
		}
	}

	return result
}

/* Mode returns most frequent value over period (PineScript compatible) */
func Mode(source []float64, period int) []float64 {
	if period <= 0 || len(source) == 0 {
		return source
	}

	result := make([]float64, len(source))

	for i := range result {
		if i < period-1 {
			result[i] = math.NaN()
			continue
		}

		hasNaN := false
		frequency := make(map[float64]int)

		for j := 0; j < period; j++ {
			val := source[i-j]
			if math.IsNaN(val) {
				hasNaN = true
				break
			}
			frequency[val]++
		}

		if hasNaN {
			result[i] = math.NaN()
			continue
		}

		maxFreq := 0
		modeVal := 0.0

		for val, freq := range frequency {
			if freq > maxFreq || (freq == maxFreq && val > modeVal) {
				maxFreq = freq
				modeVal = val
			}
		}

		result[i] = modeVal
	}

	return result
}
