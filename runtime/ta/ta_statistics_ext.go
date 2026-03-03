package ta

import (
	"math"
	"sort"
)

// Percentrank returns the percent rank of source[i] in the lookback window of size period.
// Pine: ta.percentrank(source, length)
func Percentrank(source []float64, period int) []float64 {
	if period <= 0 || len(source) == 0 {
		return source
	}

	result := make([]float64, len(source))
	for i := range result {
		if i < period-1 {
			result[i] = math.NaN()
			continue
		}
		current := source[i]
		count := 0
		for j := 0; j < period; j++ {
			if source[i-j] < current {
				count++
			}
		}
		result[i] = float64(count) / float64(period) * 100.0
	}
	return result
}

// PercentileNearestRank returns the value at pct percentile over the last period bars
// using the nearest-rank method.
// Pine: ta.percentile_nearest_rank(source, length, percentage)
func PercentileNearestRank(source []float64, period int, pct float64) []float64 {
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
		for j := 0; j < period; j++ {
			window[j] = source[i-j]
		}
		sort.Float64s(window)
		idx := int(math.Ceil(pct/100.0*float64(period))) - 1
		if idx < 0 {
			idx = 0
		}
		if idx >= period {
			idx = period - 1
		}
		result[i] = window[idx]
	}
	return result
}

// PercentileLinearInterpolation returns the value at pct percentile over the last period bars
// using linear interpolation.
// Pine: ta.percentile_linear_interpolation(source, length, percentage)
func PercentileLinearInterpolation(source []float64, period int, pct float64) []float64 {
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
		for j := 0; j < period; j++ {
			window[j] = source[i-j]
		}
		sort.Float64s(window)
		rank := pct / 100.0 * float64(period-1)
		lower := int(math.Floor(rank))
		upper := lower + 1
		frac := rank - float64(lower)
		if upper >= period {
			result[i] = window[period-1]
		} else {
			result[i] = window[lower] + frac*(window[upper]-window[lower])
		}
	}
	return result
}

// Correlation computes the Pearson correlation coefficient between source1 and source2
// over a rolling window of size period.
// Pine: ta.correlation(source1, source2, length)
func Correlation(source1, source2 []float64, period int) []float64 {
	n := len(source1)
	if len(source2) < n {
		n = len(source2)
	}
	if period <= 0 || n == 0 {
		return make([]float64, n)
	}

	result := make([]float64, n)
	for i := range result {
		if i < period-1 {
			result[i] = math.NaN()
			continue
		}
		result[i] = pearsonCorrelation(source1, source2, i, period)
	}
	return result
}

func pearsonCorrelation(x, y []float64, endIdx, period int) float64 {
	sum1, sum2 := 0.0, 0.0
	for j := 0; j < period; j++ {
		sum1 += x[endIdx-j]
		sum2 += y[endIdx-j]
	}
	mean1 := sum1 / float64(period)
	mean2 := sum2 / float64(period)

	cov, var1, var2 := 0.0, 0.0, 0.0
	for j := 0; j < period; j++ {
		d1 := x[endIdx-j] - mean1
		d2 := y[endIdx-j] - mean2
		cov += d1 * d2
		var1 += d1 * d1
		var2 += d2 * d2
	}

	if var1 == 0 || var2 == 0 {
		return 0
	}
	return cov / math.Sqrt(var1*var2)
}
