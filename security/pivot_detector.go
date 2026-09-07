package security

import (
	"math"

	"github.com/quant5-lab/runner/runtime/context"
)

type PivotDetector struct {
	leftBars  int
	rightBars int
}

func NewPivotDetector(leftBars, rightBars int) *PivotDetector {
	return &PivotDetector{
		leftBars:  leftBars,
		rightBars: rightBars,
	}
}

func (p *PivotDetector) DetectHighAtBar(data []context.OHLCV, sourceField string, barIdx int) float64 {
	if !p.canDetectPivotAt(barIdx, len(data)) {
		return math.NaN()
	}

	pivotValue := p.extractFieldValue(data[barIdx], sourceField)
	if math.IsNaN(pivotValue) {
		return math.NaN()
	}

	if !p.isLocalMaximum(data, sourceField, barIdx, pivotValue) {
		return math.NaN()
	}

	return pivotValue
}

func (p *PivotDetector) DetectLowAtBar(data []context.OHLCV, sourceField string, barIdx int) float64 {
	if !p.canDetectPivotAt(barIdx, len(data)) {
		return math.NaN()
	}

	pivotValue := p.extractFieldValue(data[barIdx], sourceField)
	if math.IsNaN(pivotValue) {
		return math.NaN()
	}

	if !p.isLocalMinimum(data, sourceField, barIdx, pivotValue) {
		return math.NaN()
	}

	return pivotValue
}

func (p *PivotDetector) canDetectPivotAt(barIdx, dataLen int) bool {
	return barIdx >= p.leftBars && barIdx+p.rightBars < dataLen
}

func (p *PivotDetector) isLocalMaximum(data []context.OHLCV, sourceField string, centerIdx int, centerValue float64) bool {
	for i := centerIdx - p.leftBars; i < centerIdx; i++ {
		if p.extractFieldValue(data[i], sourceField) >= centerValue {
			return false
		}
	}

	for i := centerIdx + 1; i <= centerIdx+p.rightBars; i++ {
		if p.extractFieldValue(data[i], sourceField) >= centerValue {
			return false
		}
	}

	return true
}

func (p *PivotDetector) isLocalMinimum(data []context.OHLCV, sourceField string, centerIdx int, centerValue float64) bool {
	for i := centerIdx - p.leftBars; i < centerIdx; i++ {
		if p.extractFieldValue(data[i], sourceField) <= centerValue {
			return false
		}
	}

	for i := centerIdx + 1; i <= centerIdx+p.rightBars; i++ {
		if p.extractFieldValue(data[i], sourceField) <= centerValue {
			return false
		}
	}

	return true
}

func (p *PivotDetector) extractFieldValue(bar context.OHLCV, fieldName string) float64 {
	switch fieldName {
	case "high":
		return bar.High
	case "low":
		return bar.Low
	case "close":
		return bar.Close
	case "open":
		return bar.Open
	default:
		return math.NaN()
	}
}
