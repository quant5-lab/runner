package ta

import (
	"math"
	"testing"
)

func TestLinreg(t *testing.T) {
	source := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result := Linreg(source, 5, 0)

	if len(result) != len(source) {
		t.Fatalf("Linreg length = %d, want %d", len(result), len(source))
	}

	for i := 0; i < 4; i++ {
		if !math.IsNaN(result[i]) {
			t.Errorf("Linreg[%d] should be NaN during warmup", i)
		}
	}

	if math.Abs(result[4]-5.0) > 0.0001 {
		t.Errorf("Linreg[4] = %f, want 5.0", result[4])
	}

	if math.Abs(result[9]-10.0) > 0.0001 {
		t.Errorf("Linreg[9] = %f, want 10.0", result[9])
	}
}

func TestLinregOffset(t *testing.T) {
	source := []float64{1, 2, 3, 4, 5, 6}

	offset0 := Linreg(source, 3, 0)
	offset1 := Linreg(source, 3, 1)
	offset2 := Linreg(source, 3, 2)

	for i := 0; i < 2; i++ {
		if !math.IsNaN(offset0[i]) || !math.IsNaN(offset1[i]) || !math.IsNaN(offset2[i]) {
			t.Errorf("Index %d should be NaN for all offsets", i)
		}
	}

	if math.Abs(offset0[2]-3.0) > 0.0001 {
		t.Errorf("offset=0 at index 2: got %f, want 3.0", offset0[2])
	}

	if math.Abs(offset1[2]-2.0) > 0.0001 {
		t.Errorf("offset=1 at index 2: got %f, want 2.0", offset1[2])
	}

	if math.Abs(offset2[2]-1.0) > 0.0001 {
		t.Errorf("offset=2 at index 2: got %f, want 1.0", offset2[2])
	}
}

func TestLinregEdgeCases(t *testing.T) {
	t.Run("empty_source", func(t *testing.T) {
		result := Linreg([]float64{}, 5, 0)
		if len(result) != 0 {
			t.Errorf("empty source should return empty, got length %d", len(result))
		}
	})

	t.Run("zero_length", func(t *testing.T) {
		source := []float64{1, 2, 3}
		result := Linreg(source, 0, 0)
		if len(result) != len(source) {
			t.Errorf("zero length should return source length, got %d", len(result))
		}
	})

	t.Run("length_exceeds_source", func(t *testing.T) {
		source := []float64{1, 2, 3}
		result := Linreg(source, 10, 0)
		for i, v := range result {
			if !math.IsNaN(v) {
				t.Errorf("index %d: expected NaN for insufficient data, got %f", i, v)
			}
		}
	})

	t.Run("length_one", func(t *testing.T) {
		source := []float64{5, 7, 9, 11}
		result := Linreg(source, 1, 0)
		for i := range result {
			if math.Abs(result[i]-source[i]) > 0.0001 {
				t.Errorf("length=1 at index %d: got %f, want %f", i, result[i], source[i])
			}
		}
	})

	t.Run("constant_values", func(t *testing.T) {
		source := []float64{42, 42, 42, 42, 42}
		result := Linreg(source, 3, 0)
		for i := 2; i < len(result); i++ {
			if math.Abs(result[i]-42.0) > 0.0001 {
				t.Errorf("constant series at index %d: got %f, want 42.0", i, result[i])
			}
		}
	})
}

func TestLinregNaN(t *testing.T) {
	source := []float64{1, 2, math.NaN(), 4, 5, 6}
	result := Linreg(source, 3, 0)

	if !math.IsNaN(result[2]) {
		t.Error("Result at index 2 should be NaN (window contains NaN)")
	}
	if !math.IsNaN(result[3]) {
		t.Error("Result at index 3 should be NaN (window contains NaN)")
	}

	if math.IsNaN(result[5]) {
		t.Error("Result at index 5 should be valid (NaN outside window)")
	}
}

func TestCalculateLeastSquares(t *testing.T) {
	t.Run("perfect_upward_trend", func(t *testing.T) {
		window := []float64{1, 2, 3, 4, 5}
		slope, intercept := calculateLeastSquares(window)

		if math.Abs(slope-1.0) > 0.0001 {
			t.Errorf("slope = %f, want 1.0", slope)
		}
		if math.Abs(intercept-1.0) > 0.0001 {
			t.Errorf("intercept = %f, want 1.0", intercept)
		}
	})

	t.Run("perfect_downward_trend", func(t *testing.T) {
		window := []float64{5, 4, 3, 2, 1}
		slope, intercept := calculateLeastSquares(window)

		if math.Abs(slope-(-1.0)) > 0.0001 {
			t.Errorf("slope = %f, want -1.0", slope)
		}
		if math.Abs(intercept-5.0) > 0.0001 {
			t.Errorf("intercept = %f, want 5.0", intercept)
		}
	})

	t.Run("constant_values", func(t *testing.T) {
		window := []float64{10, 10, 10}
		slope, intercept := calculateLeastSquares(window)

		if math.Abs(slope) > 0.0001 {
			t.Errorf("slope = %f, want 0.0 for constant", slope)
		}
		if math.Abs(intercept-10.0) > 0.0001 {
			t.Errorf("intercept = %f, want 10.0", intercept)
		}
	})

	t.Run("empty_window", func(t *testing.T) {
		slope, intercept := calculateLeastSquares([]float64{})
		if slope != 0 || intercept != 0 {
			t.Errorf("empty window should return 0,0, got %f,%f", slope, intercept)
		}
	})
}
