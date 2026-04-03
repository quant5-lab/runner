package ta

import (
	"math"
	"testing"
)

/* Max function tests */

func TestMax(t *testing.T) {
	source := []float64{10.0, 20.0, 15.0, 25.0, 30.0, 18.0}
	period := 3

	result := Max(source, period)

	expected := []float64{
		math.NaN(),
		math.NaN(),
		20.0,
		25.0,
		30.0,
		30.0,
	}

	for i := range expected {
		if math.IsNaN(expected[i]) {
			if !math.IsNaN(result[i]) {
				t.Errorf("Bar %d: expected NaN, got %f", i, result[i])
			}
		} else {
			if result[i] != expected[i] {
				t.Errorf("Bar %d: expected %f, got %f", i, expected[i], result[i])
			}
		}
	}
}

/* Min function tests */

func TestMin(t *testing.T) {
	source := []float64{10.0, 20.0, 15.0, 25.0, 30.0, 18.0}
	period := 3

	result := Min(source, period)

	expected := []float64{
		math.NaN(),
		math.NaN(),
		10.0,
		15.0,
		15.0,
		18.0,
	}

	for i := range expected {
		if math.IsNaN(expected[i]) {
			if !math.IsNaN(result[i]) {
				t.Errorf("Bar %d: expected NaN, got %f", i, result[i])
			}
		} else {
			if result[i] != expected[i] {
				t.Errorf("Bar %d: expected %f, got %f", i, expected[i], result[i])
			}
		}
	}
}

/* Median function tests */

func TestMedian(t *testing.T) {
	source := []float64{10.0, 20.0, 15.0, 25.0, 30.0}
	period := 3

	result := Median(source, period)

	expected := []float64{
		math.NaN(),
		math.NaN(),
		15.0,
		20.0,
		25.0,
	}

	for i := range expected {
		if math.IsNaN(expected[i]) {
			if !math.IsNaN(result[i]) {
				t.Errorf("Bar %d: expected NaN, got %f", i, result[i])
			}
		} else {
			if result[i] != expected[i] {
				t.Errorf("Bar %d: expected %f, got %f", i, expected[i], result[i])
			}
		}
	}
}

func TestVariance(t *testing.T) {
	source := []float64{10.0, 20.0, 30.0}
	period := 3

	result := Variance(source, period)

	if !math.IsNaN(result[0]) || !math.IsNaN(result[1]) {
		t.Error("First two bars should be NaN")
	}

	mean := 20.0
	expectedVariance := ((10.0-mean)*(10.0-mean) + (20.0-mean)*(20.0-mean) + (30.0-mean)*(30.0-mean)) / 3.0

	tolerance := 0.0001
	if math.Abs(result[2]-expectedVariance) > tolerance {
		t.Errorf("Variance: expected %f, got %f", expectedVariance, result[2])
	}
}

func TestRange(t *testing.T) {
	source := []float64{10.0, 20.0, 15.0, 25.0, 30.0}
	period := 3

	result := Range(source, period)

	expected := []float64{
		math.NaN(),
		math.NaN(),
		10.0,
		10.0,
		15.0,
	}

	for i := range expected {
		if math.IsNaN(expected[i]) {
			if !math.IsNaN(result[i]) {
				t.Errorf("Bar %d: expected NaN, got %f", i, result[i])
			}
		} else {
			if result[i] != expected[i] {
				t.Errorf("Bar %d: expected %f, got %f", i, expected[i], result[i])
			}
		}
	}
}

func TestMode(t *testing.T) {
	source := []float64{10.0, 10.0, 20.0, 10.0, 30.0}
	period := 3

	result := Mode(source, period)

	expected := []float64{
		math.NaN(),
		math.NaN(),
		10.0,
		10.0,
		30.0,
	}

	for i := range expected {
		if math.IsNaN(expected[i]) {
			if !math.IsNaN(result[i]) {
				t.Errorf("Bar %d: expected NaN, got %f", i, result[i])
			}
		} else {
			if result[i] != expected[i] {
				t.Errorf("Bar %d: expected %f, got %f", i, expected[i], result[i])
			}
		}
	}
}

func TestMaxNaNHandling(t *testing.T) {
	source := []float64{10.0, math.NaN(), 20.0, 15.0}
	period := 2

	result := Max(source, period)

	if !math.IsNaN(result[0]) {
		t.Error("Bar 0 should be NaN (warmup)")
	}

	if !math.IsNaN(result[1]) {
		t.Error("Bar 1 should be NaN (contains NaN in window)")
	}

	if !math.IsNaN(result[2]) {
		t.Error("Bar 2 should be NaN (window contains NaN at bar 1)")
	}

	if result[3] != 20.0 {
		t.Errorf("Bar 3: expected 20.0, got %f", result[3])
	}
}

func TestMinNaNHandling(t *testing.T) {
	source := []float64{10.0, math.NaN(), 20.0, 15.0}
	period := 2

	result := Min(source, period)

	if !math.IsNaN(result[0]) {
		t.Error("Bar 0 should be NaN (warmup)")
	}

	if !math.IsNaN(result[1]) {
		t.Error("Bar 1 should be NaN (contains NaN in window)")
	}

	if !math.IsNaN(result[2]) {
		t.Error("Bar 2 should be NaN (window contains NaN at bar 1)")
	}

	if result[3] != 15.0 {
		t.Errorf("Bar 3: expected 15.0, got %f", result[3])
	}
}

/* Edge case tests for aggregation functions */

func TestMax_EdgeCases(t *testing.T) {
	t.Run("empty_source", func(t *testing.T) {
		result := Max([]float64{}, 5)
		if len(result) != 0 {
			t.Errorf("empty source should return empty, got length %d", len(result))
		}
	})

	t.Run("period_one", func(t *testing.T) {
		source := []float64{5, 7, 3, 9}
		result := Max(source, 1)

		for i := range result {
			if result[i] != source[i] {
				t.Errorf("period=1 at index %d: got %f, want %f", i, result[i], source[i])
			}
		}
	})

	t.Run("period_exceeds_source", func(t *testing.T) {
		source := []float64{1, 2, 3}
		result := Max(source, 10)

		for i, v := range result {
			if !math.IsNaN(v) {
				t.Errorf("index %d: expected NaN for insufficient data, got %f", i, v)
			}
		}
	})

	t.Run("negative_values", func(t *testing.T) {
		source := []float64{-10, -5, -20, -3}
		result := Max(source, 3)

		if !math.IsNaN(result[0]) || !math.IsNaN(result[1]) {
			t.Error("First two bars should be NaN")
		}

		if result[2] != -5.0 {
			t.Errorf("Bar 2: expected -5.0, got %f", result[2])
		}

		if result[3] != -3.0 {
			t.Errorf("Bar 3: expected -3.0, got %f", result[3])
		}
	})

	t.Run("constant_values", func(t *testing.T) {
		source := []float64{42, 42, 42, 42, 42}
		result := Max(source, 3)

		for i := 2; i < len(result); i++ {
			if result[i] != 42.0 {
				t.Errorf("constant series at index %d: got %f, want 42.0", i, result[i])
			}
		}
	})

	t.Run("large_dataset", func(t *testing.T) {
		source := make([]float64, 1000)
		for i := range source {
			source[i] = float64(i % 100)
		}

		result := Max(source, 20)

		if len(result) != len(source) {
			t.Fatalf("Max length = %d, want %d", len(result), len(source))
		}

		for i := 19; i < len(result); i++ {
			if math.IsNaN(result[i]) {
				t.Errorf("index %d should not be NaN after warmup", i)
				break
			}
		}
	})
}

func TestMin_EdgeCases(t *testing.T) {
	t.Run("empty_source", func(t *testing.T) {
		result := Min([]float64{}, 5)
		if len(result) != 0 {
			t.Errorf("empty source should return empty, got length %d", len(result))
		}
	})

	t.Run("period_one", func(t *testing.T) {
		source := []float64{5, 7, 3, 9}
		result := Min(source, 1)

		for i := range result {
			if result[i] != source[i] {
				t.Errorf("period=1 at index %d: got %f, want %f", i, result[i], source[i])
			}
		}
	})

	t.Run("period_exceeds_source", func(t *testing.T) {
		source := []float64{1, 2, 3}
		result := Min(source, 10)

		for i, v := range result {
			if !math.IsNaN(v) {
				t.Errorf("index %d: expected NaN for insufficient data, got %f", i, v)
			}
		}
	})

	t.Run("negative_values", func(t *testing.T) {
		source := []float64{-10, -5, -20, -3}
		result := Min(source, 3)

		if !math.IsNaN(result[0]) || !math.IsNaN(result[1]) {
			t.Error("First two bars should be NaN")
		}

		if result[2] != -20.0 {
			t.Errorf("Bar 2: expected -20.0, got %f", result[2])
		}

		if result[3] != -20.0 {
			t.Errorf("Bar 3: expected -20.0, got %f", result[3])
		}
	})

	t.Run("constant_values", func(t *testing.T) {
		source := []float64{42, 42, 42, 42, 42}
		result := Min(source, 3)

		for i := 2; i < len(result); i++ {
			if result[i] != 42.0 {
				t.Errorf("constant series at index %d: got %f, want 42.0", i, result[i])
			}
		}
	})
}

func TestMedian_EdgeCases(t *testing.T) {
	t.Run("empty_source", func(t *testing.T) {
		result := Median([]float64{}, 5)
		if len(result) != 0 {
			t.Errorf("empty source should return empty, got length %d", len(result))
		}
	})

	t.Run("period_one", func(t *testing.T) {
		source := []float64{5, 7, 3, 9}
		result := Median(source, 1)

		for i := range result {
			if result[i] != source[i] {
				t.Errorf("period=1 at index %d: got %f, want %f", i, result[i], source[i])
			}
		}
	})

	t.Run("even_period", func(t *testing.T) {
		source := []float64{1, 2, 3, 4, 5, 6}
		result := Median(source, 4)

		if !math.IsNaN(result[0]) || !math.IsNaN(result[1]) || !math.IsNaN(result[2]) {
			t.Error("First three bars should be NaN")
		}

		expected := (2.0 + 3.0) / 2.0
		if result[3] != expected {
			t.Errorf("Bar 3 (even period): expected %f, got %f", expected, result[3])
		}
	})

	t.Run("negative_values", func(t *testing.T) {
		source := []float64{-10, -5, -20, -3, -15}
		result := Median(source, 3)

		if !math.IsNaN(result[0]) || !math.IsNaN(result[1]) {
			t.Error("First two bars should be NaN")
		}

		if result[2] != -10.0 {
			t.Errorf("Bar 2: expected -10.0, got %f", result[2])
		}
	})

	t.Run("constant_values", func(t *testing.T) {
		source := []float64{42, 42, 42, 42, 42}
		result := Median(source, 3)

		for i := 2; i < len(result); i++ {
			if result[i] != 42.0 {
				t.Errorf("constant series at index %d: got %f, want 42.0", i, result[i])
			}
		}
	})
}

func TestVariance_EdgeCases(t *testing.T) {
	t.Run("empty_source", func(t *testing.T) {
		result := Variance([]float64{}, 5)
		if len(result) != 0 {
			t.Errorf("empty source should return empty, got length %d", len(result))
		}
	})

	t.Run("period_one", func(t *testing.T) {
		source := []float64{5, 7, 3, 9}
		result := Variance(source, 1)

		for i := range result {
			if result[i] != 0.0 {
				t.Errorf("period=1 variance at index %d: got %f, want 0.0", i, result[i])
			}
		}
	})

	t.Run("constant_values", func(t *testing.T) {
		source := []float64{42, 42, 42, 42, 42}
		result := Variance(source, 3)

		for i := 2; i < len(result); i++ {
			if result[i] != 0.0 {
				t.Errorf("constant series variance at index %d: got %f, want 0.0", i, result[i])
			}
		}
	})

	t.Run("negative_values", func(t *testing.T) {
		source := []float64{-10, -20, -30}
		result := Variance(source, 3)

		if !math.IsNaN(result[0]) || !math.IsNaN(result[1]) {
			t.Error("First two bars should be NaN")
		}

		mean := -20.0
		expectedVariance := (((-10.0 - mean) * (-10.0 - mean)) + ((-20.0 - mean) * (-20.0 - mean)) + ((-30.0 - mean) * (-30.0 - mean))) / 3.0

		tolerance := 0.0001
		if math.Abs(result[2]-expectedVariance) > tolerance {
			t.Errorf("Variance of negative values: expected %f, got %f", expectedVariance, result[2])
		}
	})

	t.Run("non_negative_result", func(t *testing.T) {
		source := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		result := Variance(source, 5)

		for i := 4; i < len(result); i++ {
			if result[i] < 0 {
				t.Errorf("Variance at index %d is negative: %f", i, result[i])
			}
		}
	})
}

func TestRange_EdgeCases(t *testing.T) {
	t.Run("empty_source", func(t *testing.T) {
		result := Range([]float64{}, 5)
		if len(result) != 0 {
			t.Errorf("empty source should return empty, got length %d", len(result))
		}
	})

	t.Run("period_one", func(t *testing.T) {
		source := []float64{5, 7, 3, 9}
		result := Range(source, 1)

		for i := range result {
			if result[i] != 0.0 {
				t.Errorf("period=1 range at index %d: got %f, want 0.0", i, result[i])
			}
		}
	})

	t.Run("constant_values", func(t *testing.T) {
		source := []float64{42, 42, 42, 42, 42}
		result := Range(source, 3)

		for i := 2; i < len(result); i++ {
			if result[i] != 0.0 {
				t.Errorf("constant series range at index %d: got %f, want 0.0", i, result[i])
			}
		}
	})

	t.Run("negative_values", func(t *testing.T) {
		source := []float64{-10, -5, -20, -3}
		result := Range(source, 3)

		if !math.IsNaN(result[0]) || !math.IsNaN(result[1]) {
			t.Error("First two bars should be NaN")
		}

		expected := -5.0 - (-20.0)
		if result[2] != expected {
			t.Errorf("Bar 2: expected %f, got %f", expected, result[2])
		}
	})

	t.Run("non_negative_result", func(t *testing.T) {
		source := []float64{1, 5, 2, 8, 3, 9, 1, 7}
		result := Range(source, 4)

		for i := 3; i < len(result); i++ {
			if result[i] < 0 {
				t.Errorf("Range at index %d is negative: %f", i, result[i])
			}
		}
	})
}

func TestMode_EdgeCases(t *testing.T) {
	t.Run("empty_source", func(t *testing.T) {
		result := Mode([]float64{}, 5)
		if len(result) != 0 {
			t.Errorf("empty source should return empty, got length %d", len(result))
		}
	})

	t.Run("period_one", func(t *testing.T) {
		source := []float64{5, 7, 3, 9}
		result := Mode(source, 1)

		for i := range result {
			if result[i] != source[i] {
				t.Errorf("period=1 at index %d: got %f, want %f", i, result[i], source[i])
			}
		}
	})

	t.Run("all_unique_values", func(t *testing.T) {
		source := []float64{1, 2, 3, 4, 5}
		result := Mode(source, 3)

		if !math.IsNaN(result[0]) || !math.IsNaN(result[1]) {
			t.Error("First two bars should be NaN")
		}

		if result[2] != 3.0 {
			t.Errorf("Bar 2 (all unique, should choose max): expected 3.0, got %f", result[2])
		}
	})

	t.Run("clear_mode", func(t *testing.T) {
		source := []float64{5, 5, 7, 5, 7}
		result := Mode(source, 3)

		if result[2] != 5.0 {
			t.Errorf("Bar 2: expected 5.0 (appears twice), got %f", result[2])
		}
	})

	t.Run("tie_chooses_max", func(t *testing.T) {
		source := []float64{1, 2, 1, 2, 3}
		result := Mode(source, 4)

		if !math.IsNaN(result[0]) || !math.IsNaN(result[1]) || !math.IsNaN(result[2]) {
			t.Error("First three bars should be NaN")
		}

		if result[3] != 2.0 {
			t.Errorf("Bar 3 (tie between 1 and 2): expected 2.0 (max), got %f", result[3])
		}
	})

	t.Run("negative_values", func(t *testing.T) {
		source := []float64{-10, -10, -5, -10, -3}
		result := Mode(source, 3)

		if result[2] != -10.0 {
			t.Errorf("Bar 2: expected -10.0 (appears twice), got %f", result[2])
		}
	})
}

/* NaN propagation tests */

func TestMedian_NaNHandling(t *testing.T) {
	source := []float64{10.0, math.NaN(), 20.0, 15.0, 25.0}
	period := 3

	result := Median(source, period)

	if !math.IsNaN(result[0]) || !math.IsNaN(result[1]) {
		t.Error("First two bars should be NaN (warmup)")
	}

	if !math.IsNaN(result[2]) {
		t.Error("Bar 2 should be NaN (window contains NaN)")
	}

	if !math.IsNaN(result[3]) {
		t.Error("Bar 3 should be NaN (window contains NaN)")
	}

	if math.IsNaN(result[4]) {
		t.Error("Bar 4 should be valid (NaN outside window)")
	}
}

func TestVariance_NaNHandling(t *testing.T) {
	source := []float64{10.0, math.NaN(), 20.0, 15.0}
	period := 2

	result := Variance(source, period)

	if !math.IsNaN(result[0]) {
		t.Error("Bar 0 should be NaN (warmup)")
	}

	if !math.IsNaN(result[1]) {
		t.Error("Bar 1 should be NaN (contains NaN in window)")
	}

	if !math.IsNaN(result[2]) {
		t.Error("Bar 2 should be NaN (window contains NaN at bar 1)")
	}

	if math.IsNaN(result[3]) {
		t.Error("Bar 3 should be valid (NaN outside window)")
	}
}

func TestRange_NaNHandling(t *testing.T) {
	source := []float64{10.0, math.NaN(), 20.0, 15.0}
	period := 2

	result := Range(source, period)

	if !math.IsNaN(result[0]) {
		t.Error("Bar 0 should be NaN (warmup)")
	}

	if !math.IsNaN(result[1]) {
		t.Error("Bar 1 should be NaN (contains NaN in window)")
	}

	if !math.IsNaN(result[2]) {
		t.Error("Bar 2 should be NaN (window contains NaN at bar 1)")
	}

	if math.IsNaN(result[3]) {
		t.Error("Bar 3 should be valid (NaN outside window)")
	}
}

func TestMode_NaNHandling(t *testing.T) {
	source := []float64{10.0, math.NaN(), 10.0, 20.0}
	period := 2

	result := Mode(source, period)

	if !math.IsNaN(result[0]) {
		t.Error("Bar 0 should be NaN (warmup)")
	}

	if !math.IsNaN(result[1]) {
		t.Error("Bar 1 should be NaN (contains NaN in window)")
	}

	if !math.IsNaN(result[2]) {
		t.Error("Bar 2 should be NaN (window contains NaN at bar 1)")
	}

	if math.IsNaN(result[3]) {
		t.Error("Bar 3 should be valid (NaN outside window)")
	}
}
