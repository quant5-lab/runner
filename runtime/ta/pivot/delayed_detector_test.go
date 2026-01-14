package pivot

import (
	"math"
	"testing"
)

func TestDelayedDetectorHigh_NoFuturePeek(t *testing.T) {
	source := []float64{1, 2, 5, 3, 2, 1, 2, 4, 3, 2}
	leftBars := 2
	rightBars := 2

	detector := NewDelayedHigh(leftBars, rightBars)

	extractor := func(index int) float64 {
		if index < 0 || index >= len(source) {
			return math.NaN()
		}
		return source[index]
	}

	tests := []struct {
		currentBar    int
		expectedValue float64
		description   string
	}{
		{0, math.NaN(), "bar 0: insufficient history"},
		{1, math.NaN(), "bar 1: insufficient history"},
		{2, math.NaN(), "bar 2: insufficient history"},
		{3, math.NaN(), "bar 3: insufficient history"},
		{4, 5.0, "bar 4: detects pivot at bar 2 (value 5)"},
		{5, math.NaN(), "bar 5: no pivot detected"},
		{6, math.NaN(), "bar 6: no pivot detected"},
		{7, math.NaN(), "bar 7: no pivot detected"},
		{8, math.NaN(), "bar 8: no pivot detected"},
		{9, 4.0, "bar 9: detects pivot at bar 7 (value 4)"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := detector.DetectAtCurrentBar(tt.currentBar, extractor)

			if math.IsNaN(tt.expectedValue) {
				if !math.IsNaN(result) {
					t.Errorf("expected NaN, got %v", result)
				}
			} else {
				if math.IsNaN(result) {
					t.Errorf("expected %v, got NaN", tt.expectedValue)
				} else if result != tt.expectedValue {
					t.Errorf("expected %v, got %v", tt.expectedValue, result)
				}
			}
		})
	}
}

func TestDelayedDetectorLow_NoFuturePeek(t *testing.T) {
	source := []float64{5, 4, 1, 3, 4, 5, 4, 2, 3, 4}
	leftBars := 2
	rightBars := 2

	detector := NewDelayedLow(leftBars, rightBars)

	extractor := func(index int) float64 {
		if index < 0 || index >= len(source) {
			return math.NaN()
		}
		return source[index]
	}

	tests := []struct {
		currentBar    int
		expectedValue float64
		description   string
	}{
		{4, 1.0, "bar 4: detects pivot low at bar 2 (value 1)"},
		{9, 2.0, "bar 9: detects pivot low at bar 7 (value 2)"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := detector.DetectAtCurrentBar(tt.currentBar, extractor)

			if math.IsNaN(tt.expectedValue) {
				if !math.IsNaN(result) {
					t.Errorf("expected NaN, got %v", result)
				}
			} else {
				if math.IsNaN(result) {
					t.Errorf("expected %v, got NaN", tt.expectedValue)
				} else if result != tt.expectedValue {
					t.Errorf("expected %v, got %v", tt.expectedValue, result)
				}
			}
		})
	}
}

func TestDelayedDetectorHigh_OnlyUsesHistoricalData(t *testing.T) {
	source := []float64{1, 2, 5, 3, 2}
	leftBars := 2
	rightBars := 2

	detector := NewDelayedHigh(leftBars, rightBars)

	accessLog := make(map[int]bool)

	extractor := func(index int) float64 {
		accessLog[index] = true
		if index < 0 || index >= len(source) {
			return math.NaN()
		}
		return source[index]
	}

	currentBar := 4
	result := detector.DetectAtCurrentBar(currentBar, extractor)

	if math.IsNaN(result) {
		t.Errorf("expected pivot value 5, got NaN")
	}

	for accessedIndex := range accessLog {
		if accessedIndex > currentBar {
			t.Errorf("FUTURE PEEK DETECTED: accessed index %d when current bar is %d", accessedIndex, currentBar)
		}
	}

	expectedAccesses := []int{0, 1, 2, 3, 4}
	for _, expected := range expectedAccesses {
		if !accessLog[expected] {
			t.Errorf("expected to access index %d but didn't", expected)
		}
	}
}

func TestDelayedDetector_CanDetectAtCurrentBar(t *testing.T) {
	detector := NewDelayedHigh(2, 2)

	tests := []struct {
		currentBar int
		canDetect  bool
	}{
		{0, false},
		{1, false},
		{2, false},
		{3, false},
		{4, true},
		{5, true},
		{100, true},
	}

	for _, tt := range tests {
		result := detector.CanDetectAtCurrentBar(tt.currentBar)
		if result != tt.canDetect {
			t.Errorf("at bar %d: expected %v, got %v", tt.currentBar, tt.canDetect, result)
		}
	}
}
