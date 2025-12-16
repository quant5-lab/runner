package pivot

import "math"

type DelayedDetector struct {
	window  Window
	checker ExtremaChecker
}

func NewDelayedHigh(leftBars, rightBars int) *DelayedDetector {
	return &DelayedDetector{
		window:  NewWindow(leftBars, rightBars),
		checker: MaximumChecker{},
	}
}

func NewDelayedLow(leftBars, rightBars int) *DelayedDetector {
	return &DelayedDetector{
		window:  NewWindow(leftBars, rightBars),
		checker: MinimumChecker{},
	}
}

func (d *DelayedDetector) CanDetectAtCurrentBar(currentBarIndex int) bool {
	minimumRequiredBars := d.window.leftBars + d.window.rightBars
	return currentBarIndex >= minimumRequiredBars
}

func (d *DelayedDetector) DetectAtCurrentBar(currentBarIndex int, extractor ValueExtractor) float64 {
	if !d.CanDetectAtCurrentBar(currentBarIndex) {
		return math.NaN()
	}

	centerIndex := currentBarIndex - d.window.rightBars
	centerValue := extractor(centerIndex)

	if math.IsNaN(centerValue) {
		return math.NaN()
	}

	neighbors := d.collectNeighbors(centerIndex, extractor)
	if !d.checker.IsCenterExtremum(centerValue, neighbors) {
		return math.NaN()
	}

	return centerValue
}

func (d *DelayedDetector) collectNeighbors(centerIndex int, extractor ValueExtractor) []float64 {
	totalNeighbors := d.window.leftBars + d.window.rightBars
	neighbors := make([]float64, 0, totalNeighbors)

	for i := centerIndex - d.window.leftBars; i < centerIndex; i++ {
		neighbors = append(neighbors, extractor(i))
	}

	for i := centerIndex + 1; i <= centerIndex+d.window.rightBars; i++ {
		neighbors = append(neighbors, extractor(i))
	}

	return neighbors
}
