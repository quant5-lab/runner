package request

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

func floatSliceEqual(a, b []float64, tolerance float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.IsNaN(a[i]) && math.IsNaN(b[i]) {
			continue
		}
		if math.Abs(a[i]-b[i]) > tolerance {
			return false
		}
	}
	return true
}

func TestPivotResultCache_EmptySource(t *testing.T) {
	cache := NewPivotResultCache()

	source := []float64{}
	result := cache.ComputeOrRetrieve(PivotTypeHigh, source, 5, 5)

	if len(result) != 0 {
		t.Errorf("Expected empty result for empty source, got length %d", len(result))
	}
}

func TestPivotResultCache_SingleElement(t *testing.T) {
	cache := NewPivotResultCache()

	source := []float64{42.0}
	result := cache.ComputeOrRetrieve(PivotTypeHigh, source, 1, 1)

	if len(result) != 1 {
		t.Fatalf("Expected result length 1, got %d", len(result))
	}

	if !math.IsNaN(result[0]) {
		t.Error("Expected NaN for single element with bars=1")
	}
}

func TestPivotResultCache_AllNaNSource(t *testing.T) {
	cache := NewPivotResultCache()

	source := []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}
	result := cache.ComputeOrRetrieve(PivotTypeHigh, source, 1, 1)

	if len(result) != len(source) {
		t.Fatalf("Expected result length %d, got %d", len(source), len(result))
	}

	for i, val := range result {
		if !math.IsNaN(val) {
			t.Errorf("Expected NaN at index %d for all-NaN source, got %f", i, val)
		}
	}
}

func TestPivotResultCache_ZeroBars(t *testing.T) {
	cache := NewPivotResultCache()

	source := []float64{1, 5, 3, 8, 2, 6, 4}

	tests := []struct {
		name      string
		leftBars  int
		rightBars int
	}{
		{"zero_both", 0, 0},
		{"zero_left", 0, 2},
		{"zero_right", 2, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cache.ComputeOrRetrieve(PivotTypeHigh, source, tt.leftBars, tt.rightBars)

			if len(result) != len(source) {
				t.Errorf("Expected result length %d, got %d", len(source), len(result))
			}
		})
	}
}

func TestPivotResultCache_LargeBars(t *testing.T) {
	cache := NewPivotResultCache()

	source := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	leftBars := 100
	rightBars := 100

	result := cache.ComputeOrRetrieve(PivotTypeHigh, source, leftBars, rightBars)

	if len(result) != len(source) {
		t.Errorf("Expected result length %d, got %d", len(source), len(result))
	}

	for i, val := range result {
		if !math.IsNaN(val) {
			t.Errorf("Expected NaN at index %d when bars exceed array length, got %f", i, val)
		}
	}
}

func TestPivotResultCache_LargeArray(t *testing.T) {
	cache := NewPivotResultCache()

	size := 10000
	source := make([]float64, size)
	for i := 0; i < size; i++ {
		source[i] = float64(i % 100)
	}

	result := cache.ComputeOrRetrieve(PivotTypeHigh, source, 5, 5)

	if len(result) != size {
		t.Errorf("Expected result length %d, got %d", size, len(result))
	}
}

func TestPivotResultCache_CacheKeyUniqueness_Extended(t *testing.T) {
	cache := NewPivotResultCache()

	tests := []struct {
		name         string
		type1        PivotFunctionType
		left1        int
		right1       int
		len1         int
		type2        PivotFunctionType
		left2        int
		right2       int
		len2         int
		shouldDiffer bool
	}{
		{"same_all", PivotTypeHigh, 5, 5, 100, PivotTypeHigh, 5, 5, 100, false},
		{"diff_type", PivotTypeHigh, 5, 5, 100, PivotTypeLow, 5, 5, 100, true},
		{"diff_left", PivotTypeHigh, 5, 5, 100, PivotTypeHigh, 10, 5, 100, true},
		{"diff_right", PivotTypeHigh, 5, 5, 100, PivotTypeHigh, 5, 10, 100, true},
		{"diff_length", PivotTypeHigh, 5, 5, 100, PivotTypeHigh, 5, 5, 200, true},
		{"diff_multiple", PivotTypeHigh, 5, 5, 100, PivotTypeLow, 10, 10, 200, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key1 := cache.buildCacheKey(tt.type1, tt.left1, tt.right1, tt.len1)
			key2 := cache.buildCacheKey(tt.type2, tt.left2, tt.right2, tt.len2)

			if tt.shouldDiffer {
				if key1 == key2 {
					t.Errorf("Keys should differ but are equal: %s", key1)
				}
			} else {
				if key1 != key2 {
					t.Errorf("Keys should be equal but differ: %s vs %s", key1, key2)
				}
			}
		})
	}
}

func TestPivotResultCache_MultipleCalls_Independence(t *testing.T) {
	cache := NewPivotResultCache()

	source1 := []float64{10, 50, 30, 20, 80, 40, 10}
	source2 := []float64{100, 10, 120, 5, 140, 8, 160}

	resultHigh := cache.ComputeOrRetrieve(PivotTypeHigh, source1, 2, 2)
	resultLow := cache.ComputeOrRetrieve(PivotTypeLow, source2, 2, 2)

	if len(cache.cache) != 2 {
		t.Errorf("Expected 2 cache entries, got %d", len(cache.cache))
	}

	if len(resultHigh) != len(resultLow) {
		return
	}

	foundDifference := false
	for i := range resultHigh {
		bothNaN := math.IsNaN(resultHigh[i]) && math.IsNaN(resultLow[i])
		if !bothNaN && resultHigh[i] != resultLow[i] {
			foundDifference = true
			break
		}
	}

	if !foundDifference {
		t.Error("Expected at least some different values for different sources and types")
	}
}

func TestPivotResultCache_CachePersistence(t *testing.T) {
	cache := NewPivotResultCache()

	source := []float64{1, 5, 3, 8, 2, 6, 4}

	result1 := cache.ComputeOrRetrieve(PivotTypeHigh, source, 2, 2)

	source[0] = 999.0

	result2 := cache.ComputeOrRetrieve(PivotTypeHigh, source, 2, 2)

	if !floatSliceEqual(result1, result2, 0.0001) {
		t.Error("Cache should return same result regardless of source mutation")
	}
}

func TestExtractSourceSeries_High_Integration(t *testing.T) {
	secCtx := &context.Context{
		Data: []context.OHLCV{
			{High: 100.5, Low: 99.0},
			{High: 101.2, Low: 99.5},
			{High: 102.0, Low: 100.0},
			{High: 101.8, Low: 100.5},
			{High: 103.0, Low: 101.0},
		},
	}

	result := extractHighSeries(secCtx)

	expected := []float64{100.5, 101.2, 102.0, 101.8, 103.0}

	if !floatSliceEqual(result, expected, 0.0001) {
		t.Errorf("extractHighSeries = %v, want %v", result, expected)
	}
}

func TestExtractSourceSeries_Low_Integration(t *testing.T) {
	secCtx := &context.Context{
		Data: []context.OHLCV{
			{High: 100.5, Low: 99.0},
			{High: 101.2, Low: 99.5},
			{High: 102.0, Low: 100.0},
			{High: 101.8, Low: 100.5},
			{High: 103.0, Low: 101.0},
		},
	}

	result := extractLowSeries(secCtx)

	expected := []float64{99.0, 99.5, 100.0, 100.5, 101.0}

	if !floatSliceEqual(result, expected, 0.0001) {
		t.Errorf("extractLowSeries = %v, want %v", result, expected)
	}
}

func TestExtractSourceSeries_CustomSource_Priority(t *testing.T) {
	secCtx := &context.Context{
		Data: []context.OHLCV{
			{High: 100.0, Low: 90.0},
			{High: 110.0, Low: 95.0},
		},
	}

	customSource := []float64{42.0, 43.0}

	resultHigh := ExtractSourceSeries(PivotTypeHigh, secCtx, customSource)
	resultLow := ExtractSourceSeries(PivotTypeLow, secCtx, customSource)

	if !floatSliceEqual(resultHigh, customSource, 0.0001) {
		t.Error("Custom source should take priority for PivotTypeHigh")
	}

	if !floatSliceEqual(resultLow, customSource, 0.0001) {
		t.Error("Custom source should take priority for PivotTypeLow")
	}
}

func TestExtractSourceSeries_NilCustomSource(t *testing.T) {
	secCtx := &context.Context{
		Data: []context.OHLCV{
			{High: 100.0, Low: 90.0},
			{High: 110.0, Low: 95.0},
		},
	}

	resultHigh := ExtractSourceSeries(PivotTypeHigh, secCtx, nil)
	resultLow := ExtractSourceSeries(PivotTypeLow, secCtx, nil)

	expectedHigh := []float64{100.0, 110.0}
	expectedLow := []float64{90.0, 95.0}

	if !floatSliceEqual(resultHigh, expectedHigh, 0.0001) {
		t.Errorf("High series = %v, want %v", resultHigh, expectedHigh)
	}

	if !floatSliceEqual(resultLow, expectedLow, 0.0001) {
		t.Errorf("Low series = %v, want %v", resultLow, expectedLow)
	}
}

func TestExtractSourceSeries_EmptyCustomSource(t *testing.T) {
	secCtx := &context.Context{
		Data: []context.OHLCV{
			{High: 100.0, Low: 90.0},
		},
	}

	emptyCustom := []float64{}

	resultHigh := ExtractSourceSeries(PivotTypeHigh, secCtx, emptyCustom)

	expected := []float64{100.0}

	if !floatSliceEqual(resultHigh, expected, 0.0001) {
		t.Error("Empty custom source should fallback to default series")
	}
}

func TestExtractSourceSeries_EmptyContext(t *testing.T) {
	secCtx := &context.Context{
		Data: []context.OHLCV{},
	}

	resultHigh := extractHighSeries(secCtx)
	resultLow := extractLowSeries(secCtx)

	if len(resultHigh) != 0 {
		t.Errorf("Expected empty high series, got length %d", len(resultHigh))
	}

	if len(resultLow) != 0 {
		t.Errorf("Expected empty low series, got length %d", len(resultLow))
	}
}

func TestPivotResultCache_TypeSpecificity(t *testing.T) {
	cache := NewPivotResultCache()

	source := []float64{50, 100, 80, 60, 150, 70, 40, 120, 50}

	resultHigh := cache.ComputeOrRetrieve(PivotTypeHigh, source, 2, 2)
	resultLow := cache.ComputeOrRetrieve(PivotTypeLow, source, 2, 2)

	if len(cache.cache) != 2 {
		t.Errorf("Expected 2 separate cache entries, got %d", len(cache.cache))
	}

	foundDifference := false
	for i := range resultHigh {
		highNaN := math.IsNaN(resultHigh[i])
		lowNaN := math.IsNaN(resultLow[i])

		if highNaN != lowNaN {
			foundDifference = true
			break
		}

		if !highNaN && !lowNaN && resultHigh[i] != resultLow[i] {
			foundDifference = true
			break
		}
	}

	if !foundDifference {
		t.Error("PivotTypeHigh and PivotTypeLow should produce different results for peak/valley data")
	}
}
