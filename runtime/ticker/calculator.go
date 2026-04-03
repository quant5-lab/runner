package ticker

import "github.com/quant5-lab/runner/runtime/context"

/* CalculateHeikinAshiBar computes HA OHLC from current and previous bars
 *
 * Formula:
 *   haClose = (open + high + low + close) / 4
 *   haOpen = (prevHaOpen + prevHaClose) / 2
 *   haHigh = max(high, haOpen, haClose)
 *   haLow = min(low, haOpen, haClose)
 */
func CalculateHeikinAshiBar(current, previous context.OHLCV, prevHaOpen, prevHaClose float64) context.OHLCV {
	haClose := (current.Open + current.High + current.Low + current.Close) / 4.0

	haOpen := prevHaOpen
	if prevHaOpen != 0 || prevHaClose != 0 {
		haOpen = (prevHaOpen + prevHaClose) / 2.0
	} else {
		haOpen = (current.Open + current.Close) / 2.0
	}

	haHigh := max3(current.High, haOpen, haClose)
	haLow := min3(current.Low, haOpen, haClose)

	return context.OHLCV{
		Time:   current.Time,
		Open:   haOpen,
		High:   haHigh,
		Low:    haLow,
		Close:  haClose,
		Volume: current.Volume,
	}
}

func max3(a, b, c float64) float64 {
	result := a
	if b > result {
		result = b
	}
	if c > result {
		result = c
	}
	return result
}

func min3(a, b, c float64) float64 {
	result := a
	if b < result {
		result = b
	}
	if c < result {
		result = c
	}
	return result
}
