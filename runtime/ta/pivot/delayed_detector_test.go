package pivot

import (
	"math"
	"testing"
)

func sliceExtractor(source []float64) ValueExtractor {
	return func(index int) float64 {
		if index < 0 || index >= len(source) {
			return math.NaN()
		}
		return source[index]
	}
}

func assertDetection(t *testing.T, got float64, wantNaN bool, want float64) {
	t.Helper()
	if wantNaN {
		if !math.IsNaN(got) {
			t.Errorf("expected NaN, got %.4f", got)
		}
		return
	}
	if math.IsNaN(got) {
		t.Errorf("expected %.4f, got NaN", want)
	} else if got != want {
		t.Errorf("expected %.4f, got %.4f", want, got)
	}
}

// TestDelayedDetector_NoFuturePeek verifies the full per-bar output sequence for
// both high and low detectors: NaN during warmup, extremum value at each detection
// bar, and NaN at non-extremum bars — confirming that detection is delayed by
// rightBars and no future data is implicitly consumed.
func TestDelayedDetector_NoFuturePeek(t *testing.T) {
	type barCase struct {
		name      string
		bar       int
		wantNaN   bool
		wantValue float64
	}
	tests := []struct {
		name      string
		newDetect func() *DelayedDetector
		source    []float64
		cases     []barCase
	}{
		{
			name:      "high",
			newDetect: func() *DelayedDetector { return NewDelayedHigh(2, 2) },
			source:    []float64{1, 2, 5, 3, 2, 1, 2, 4, 3, 2},
			cases: []barCase{
				{"pre_warmup_bar0", 0, true, 0},
				{"pre_warmup_bar1", 1, true, 0},
				{"pre_warmup_bar2", 2, true, 0},
				{"pre_warmup_bar3", 3, true, 0},
				{"detection_bar4", 4, false, 5.0},
				{"non_detection_bar5", 5, true, 0},
				{"non_detection_bar6", 6, true, 0},
				{"non_detection_bar7", 7, true, 0},
				{"non_detection_bar8", 8, true, 0},
				{"detection_bar9", 9, false, 4.0},
			},
		},
		{
			name:      "low",
			newDetect: func() *DelayedDetector { return NewDelayedLow(2, 2) },
			source:    []float64{5, 4, 1, 3, 4, 5, 4, 2, 3, 4},
			cases: []barCase{
				{"pre_warmup_bar0", 0, true, 0},
				{"pre_warmup_bar1", 1, true, 0},
				{"pre_warmup_bar2", 2, true, 0},
				{"pre_warmup_bar3", 3, true, 0},
				{"detection_bar4", 4, false, 1.0},
				{"non_detection_bar5", 5, true, 0},
				{"non_detection_bar6", 6, true, 0},
				{"non_detection_bar7", 7, true, 0},
				{"non_detection_bar8", 8, true, 0},
				{"detection_bar9", 9, false, 2.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := tt.newDetect()
			extractor := sliceExtractor(tt.source)
			for _, bc := range tt.cases {
				t.Run(bc.name, func(t *testing.T) {
					result := detector.DetectAtCurrentBar(bc.bar, extractor)
					assertDetection(t, result, bc.wantNaN, bc.wantValue)
				})
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

// TestDelayedDetector_NaNNeighborBlocking verifies that a NaN neighbor value
// blocks pivot detection, matching PineScript semantics where any comparison
// involving na evaluates to false — so "center > na_neighbor" is false and
// the pivot is not confirmed.
func TestDelayedDetector_NaNNeighborBlocking(t *testing.T) {
	nan := math.NaN()

	tests := []struct {
		name       string
		newDetect  func() *DelayedDetector
		source     []float64
		currentBar int
		wantNaN    bool
		wantValue  float64
	}{
		// NaN neighbor → pivot blocked regardless of other neighbors.
		{"high_left_nan", func() *DelayedDetector { return NewDelayedHigh(2, 2) },
			[]float64{nan, nan, 5, 3, 2}, 4, true, 0},
		{"high_right_nan", func() *DelayedDetector { return NewDelayedHigh(2, 2) },
			[]float64{3, 2, 5, nan, nan}, 4, true, 0},
		{"high_all_nan", func() *DelayedDetector { return NewDelayedHigh(2, 2) },
			[]float64{nan, nan, 5, nan, nan}, 4, true, 0},
		{"high_mixed_nan_one_valid_left_blocks", func() *DelayedDetector { return NewDelayedHigh(2, 2) },
			[]float64{nan, 3, 5, 2, nan}, 4, true, 0},
		{"low_left_nan", func() *DelayedDetector { return NewDelayedLow(2, 2) },
			[]float64{nan, nan, 1, 3, 4}, 4, true, 0},
		{"low_right_nan", func() *DelayedDetector { return NewDelayedLow(2, 2) },
			[]float64{3, 4, 1, nan, nan}, 4, true, 0},
		// NaN center → always NaN (independent of neighbor NaN blocking).
		{"high_nan_center", func() *DelayedDetector { return NewDelayedHigh(2, 2) },
			[]float64{1, 2, nan, 2, 1}, 4, true, 0},
		{"low_nan_center", func() *DelayedDetector { return NewDelayedLow(2, 2) },
			[]float64{3, 4, nan, 4, 5}, 4, true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.newDetect().DetectAtCurrentBar(tt.currentBar, sliceExtractor(tt.source))
			assertDetection(t, result, tt.wantNaN, tt.wantValue)
		})
	}
}

func TestDelayedDetector_StrictNeighborInequality(t *testing.T) {
	tests := []struct {
		name      string
		newDetect func() *DelayedDetector
		source    []float64
		wantNaN   bool
		wantValue float64
	}{
		{"high_equal_right_neighbor",
			func() *DelayedDetector { return NewDelayedHigh(2, 2) },
			[]float64{2, 2, 5, 5, 2}, true, 0},
		{"high_equal_left_neighbor",
			func() *DelayedDetector { return NewDelayedHigh(2, 2) },
			[]float64{2, 5, 5, 2, 2}, true, 0},
		{"high_all_equal",
			func() *DelayedDetector { return NewDelayedHigh(2, 2) },
			[]float64{5, 5, 5, 5, 5}, true, 0},
		{"high_strictly_greater_contrast",
			func() *DelayedDetector { return NewDelayedHigh(2, 2) },
			[]float64{3, 4, 5, 4, 3}, false, 5.0},
		{"low_equal_right_neighbor",
			func() *DelayedDetector { return NewDelayedLow(2, 2) },
			[]float64{4, 4, 1, 1, 4}, true, 0},
		{"low_equal_left_neighbor",
			func() *DelayedDetector { return NewDelayedLow(2, 2) },
			[]float64{4, 1, 1, 4, 4}, true, 0},
		{"low_all_equal",
			func() *DelayedDetector { return NewDelayedLow(2, 2) },
			[]float64{1, 1, 1, 1, 1}, true, 0},
		{"low_strictly_less_contrast",
			func() *DelayedDetector { return NewDelayedLow(2, 2) },
			[]float64{3, 4, 1, 4, 3}, false, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.newDetect().DetectAtCurrentBar(4, sliceExtractor(tt.source))
			assertDetection(t, result, tt.wantNaN, tt.wantValue)
		})
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
