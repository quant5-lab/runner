package arrayops

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/runtime/series"
)

func TestStatistics_Sum(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0, 4.0, 5.0})

	stats := NewStatistics()
	result := stats.Sum(s, 0)

	expected := 15.0
	if result != expected {
		t.Errorf("Sum: expected %f, got %f", expected, result)
	}
}

func TestStatistics_Avg(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{10.0, 20.0, 30.0})

	stats := NewStatistics()
	result := stats.Avg(s, 0)

	expected := 20.0
	if result != expected {
		t.Errorf("Avg: expected %f, got %f", expected, result)
	}
}

func TestStatistics_Min(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{5.0, 1.0, 9.0, 3.0})

	stats := NewStatistics()
	result := stats.Min(s, 0)

	expected := 1.0
	if result != expected {
		t.Errorf("Min: expected %f, got %f", expected, result)
	}
}

func TestStatistics_Max(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{5.0, 1.0, 9.0, 3.0})

	stats := NewStatistics()
	result := stats.Max(s, 0)

	expected := 9.0
	if result != expected {
		t.Errorf("Max: expected %f, got %f", expected, result)
	}
}

func TestStatistics_Median_OddLength(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 3.0, 5.0})

	stats := NewStatistics()
	result := stats.Median(s, 0)

	expected := 3.0
	if result != expected {
		t.Errorf("Median (odd): expected %f, got %f", expected, result)
	}
}

func TestStatistics_Median_EvenLength(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0, 4.0})

	stats := NewStatistics()
	result := stats.Median(s, 0)

	expected := 2.5
	if result != expected {
		t.Errorf("Median (even): expected %f, got %f", expected, result)
	}
}

func TestStatistics_Stdev(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0})

	stats := NewStatistics()
	result := stats.Stdev(s, 0, false)

	expected := 2.138090
	tolerance := 0.0001
	if math.Abs(result-expected) > tolerance {
		t.Errorf("Stdev (unbiased): expected %f, got %f", expected, result)
	}
}

func TestStatistics_Variance(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0})

	stats := NewStatistics()
	result := stats.Variance(s, 0, false)

	expected := 4.571429
	tolerance := 0.0001
	if math.Abs(result-expected) > tolerance {
		t.Errorf("Variance (unbiased): expected %f, got %f", expected, result)
	}
}

func TestStatistics_Range(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{3.0, 7.0, 1.0, 9.0, 5.0})

	stats := NewStatistics()
	result := stats.Range(s, 0)

	expected := 8.0
	if result != expected {
		t.Errorf("Range: expected %f, got %f", expected, result)
	}
}

func TestStatistics_PercentRank(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0, 4.0, 5.0})

	stats := NewStatistics()
	result := stats.PercentRank(s, 0, 3.0)

	expected := 40.0
	if result != expected {
		t.Errorf("PercentRank: expected %f, got %f", expected, result)
	}
}

func TestStatistics_Standardize(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{10.0, 20.0, 30.0})

	stats := NewStatistics()
	result := stats.Standardize(s, 0)

	if len(result) != 3 {
		t.Fatalf("Standardize: expected length 3, got %d", len(result))
	}

	sumStandardized := 0.0
	for _, v := range result {
		sumStandardized += v
	}
	tolerance := 0.0001
	if math.Abs(sumStandardized) > tolerance {
		t.Errorf("Standardize: expected sum ≈ 0, got %f", sumStandardized)
	}
}

func TestStatistics_Abs(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{-1.0, 2.0, -3.0, 4.0})

	stats := NewStatistics()
	result := stats.Abs(s, 0)

	expected := []float64{1.0, 2.0, 3.0, 4.0}
	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("Abs: expected result[%d] = %f, got %f", i, expected[i], result[i])
		}
	}
}

func TestStatistics_Sum_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	stats := NewStatistics()
	result := stats.Sum(s, 0)

	if result != 0.0 {
		t.Errorf("Sum empty: expected 0.0, got %f", result)
	}
}

func TestStatistics_Avg_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	stats := NewStatistics()
	result := stats.Avg(s, 0)

	if !math.IsNaN(result) {
		t.Errorf("Avg empty: expected NaN, got %f", result)
	}
}

func TestStatistics_Min_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	stats := NewStatistics()
	result := stats.Min(s, 0)

	if !math.IsNaN(result) {
		t.Errorf("Min empty: expected NaN, got %f", result)
	}
}

func TestStatistics_Max_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	stats := NewStatistics()
	result := stats.Max(s, 0)

	if !math.IsNaN(result) {
		t.Errorf("Max empty: expected NaN, got %f", result)
	}
}

func TestStatistics_Median_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	stats := NewStatistics()
	result := stats.Median(s, 0)

	if !math.IsNaN(result) {
		t.Errorf("Median empty: expected NaN, got %f", result)
	}
}

func TestStatistics_Stdev_SingleElement(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{42.0})

	stats := NewStatistics()
	result := stats.Stdev(s, 0, false)

	if !math.IsNaN(result) {
		t.Errorf("Stdev single element (unbiased): expected NaN, got %f", result)
	}
}

func TestStatistics_Stdev_SingleElement_Biased(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{42.0})

	stats := NewStatistics()
	result := stats.Stdev(s, 0, true)

	if result != 0.0 {
		t.Errorf("Stdev single element (biased): expected 0.0, got %f", result)
	}
}

func TestStatistics_Variance_SingleElement(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{42.0})

	stats := NewStatistics()
	result := stats.Variance(s, 0, false)

	if !math.IsNaN(result) {
		t.Errorf("Variance single element (unbiased): expected NaN, got %f", result)
	}
}

func TestStatistics_Percentile_OutOfRange(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0})

	stats := NewStatistics()

	result := stats.Percentile(s, 0, -10.0, "linear")
	if !math.IsNaN(result) {
		t.Errorf("Percentile negative: expected NaN, got %f", result)
	}

	result = stats.Percentile(s, 0, 150.0, "linear")
	if !math.IsNaN(result) {
		t.Errorf("Percentile > 100: expected NaN, got %f", result)
	}
}

func TestStatistics_Percentile_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	stats := NewStatistics()
	result := stats.Percentile(s, 0, 50.0, "linear")

	if !math.IsNaN(result) {
		t.Errorf("Percentile empty: expected NaN, got %f", result)
	}
}

func TestStatistics_PercentRank_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	stats := NewStatistics()
	result := stats.PercentRank(s, 0, 5.0)

	if !math.IsNaN(result) {
		t.Errorf("PercentRank empty: expected NaN, got %f", result)
	}
}

func TestStatistics_Covariance_MismatchedLengths(t *testing.T) {
	s1 := series.NewArraySeries(10)
	s1.Set([]float64{1.0, 2.0, 3.0})

	s2 := series.NewArraySeries(10)
	s2.Set([]float64{1.0, 2.0})

	stats := NewStatistics()
	result := stats.Covariance(s1, 0, s2, 0, false)

	if !math.IsNaN(result) {
		t.Errorf("Covariance mismatched: expected NaN, got %f", result)
	}
}

func TestStatistics_Covariance_SingleElement(t *testing.T) {
	s1 := series.NewArraySeries(10)
	s1.Set([]float64{1.0})

	s2 := series.NewArraySeries(10)
	s2.Set([]float64{2.0})

	stats := NewStatistics()
	resultUnbiased := stats.Covariance(s1, 0, s2, 0, false)

	if !math.IsNaN(resultUnbiased) {
		t.Errorf("Covariance single (unbiased): expected NaN, got %f", resultUnbiased)
	}

	resultBiased := stats.Covariance(s1, 0, s2, 0, true)
	if resultBiased != 0.0 {
		t.Errorf("Covariance single (biased): expected 0.0, got %f", resultBiased)
	}
}

func TestStatistics_Covariance_EmptyArrays(t *testing.T) {
	s1 := series.NewArraySeries(10)
	s1.Set([]float64{})

	s2 := series.NewArraySeries(10)
	s2.Set([]float64{})

	stats := NewStatistics()
	result := stats.Covariance(s1, 0, s2, 0, false)

	if !math.IsNaN(result) {
		t.Errorf("Covariance empty: expected NaN, got %f", result)
	}
}

func TestStatistics_Standardize_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	stats := NewStatistics()
	result := stats.Standardize(s, 0)

	if result != nil {
		t.Errorf("Standardize empty: expected nil, got %v", result)
	}
}

func TestStatistics_Standardize_ZeroStdev(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{5.0, 5.0, 5.0})

	stats := NewStatistics()
	result := stats.Standardize(s, 0)

	for i, v := range result {
		if v != 0.0 {
			t.Errorf("Standardize zero stdev: expected all 0.0, got result[%d] = %f", i, v)
		}
	}
}

func TestStatistics_Abs_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	stats := NewStatistics()
	result := stats.Abs(s, 0)

	if len(result) != 0 {
		t.Errorf("Abs empty: expected empty, got %v", result)
	}
}

func TestStatistics_Abs_NilArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set(nil)

	stats := NewStatistics()
	result := stats.Abs(s, 0)

	if result != nil {
		t.Errorf("Abs nil: expected nil, got %v", result)
	}
}

func TestStatistics_HistoricalAccess(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0})
	s.Next()
	s.Set([]float64{10.0, 20.0, 30.0})

	stats := NewStatistics()

	currentSum := stats.Sum(s, 0)
	if currentSum != 60.0 {
		t.Errorf("Sum current bar: expected 60.0, got %f", currentSum)
	}

	previousSum := stats.Sum(s, 1)
	if previousSum != 6.0 {
		t.Errorf("Sum previous bar: expected 6.0, got %f", previousSum)
	}
}
