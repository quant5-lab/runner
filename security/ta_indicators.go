package security

import (
	"math"

	"github.com/quant5-lab/runner/runtime/context"
)

type TrueRangeCalculator struct{}

func NewTrueRangeCalculator() *TrueRangeCalculator {
	return &TrueRangeCalculator{}
}

func (c *TrueRangeCalculator) CalculateAtBar(bars []context.OHLCV, barIdx int, prevClose float64, isFirstBar bool) float64 {
	if barIdx < 0 || barIdx >= len(bars) {
		return math.NaN()
	}

	bar := bars[barIdx]

	if isFirstBar {
		return bar.High - bar.Low
	}

	highLowRange := bar.High - bar.Low
	highPrevCloseRange := math.Abs(bar.High - prevClose)
	lowPrevCloseRange := math.Abs(bar.Low - prevClose)

	return math.Max(highLowRange, math.Max(highPrevCloseRange, lowPrevCloseRange))
}
