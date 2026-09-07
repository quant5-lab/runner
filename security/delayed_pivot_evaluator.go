package security

import (
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/ta/pivot"
)

type DelayedPivotEvaluator struct {
	detector *pivot.DelayedDetector
}

func NewDelayedPivotHighEvaluator(leftBars, rightBars int) *DelayedPivotEvaluator {
	return &DelayedPivotEvaluator{
		detector: pivot.NewDelayedHigh(leftBars, rightBars),
	}
}

func NewDelayedPivotLowEvaluator(leftBars, rightBars int) *DelayedPivotEvaluator {
	return &DelayedPivotEvaluator{
		detector: pivot.NewDelayedLow(leftBars, rightBars),
	}
}

func (e *DelayedPivotEvaluator) EvaluateAtBar(data []context.OHLCV, sourceField string, currentBarIndex int) float64 {
	extractor := createFieldExtractor(data, sourceField)
	return e.detector.DetectAtCurrentBar(currentBarIndex, extractor)
}

func createFieldExtractor(data []context.OHLCV, sourceField string) pivot.ValueExtractor {
	return func(index int) float64 {
		if index < 0 || index >= len(data) {
			return 0.0
		}
		return extractFieldValue(data[index], sourceField)
	}
}

func extractFieldValue(bar context.OHLCV, field string) float64 {
	switch field {
	case "open":
		return bar.Open
	case "high":
		return bar.High
	case "low":
		return bar.Low
	case "close":
		return bar.Close
	case "volume":
		return bar.Volume
	default:
		return bar.Close
	}
}
