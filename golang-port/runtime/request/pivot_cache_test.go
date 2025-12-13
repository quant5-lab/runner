package request

import (
	"math"
	"testing"
)

func TestPivotResultCache_ComputeOrRetrieve_CachesResult(t *testing.T) {
	cache := NewPivotResultCache()

	source := []float64{1, 5, 3, 2, 8, 4, 1}
	leftBars := 2
	rightBars := 2

	// First call computes
	result1 := cache.ComputeOrRetrieve(PivotTypeHigh, source, leftBars, rightBars)

	// Second call retrieves from cache
	result2 := cache.ComputeOrRetrieve(PivotTypeHigh, source, leftBars, rightBars)

	if len(result1) != len(result2) {
		t.Fatal("Cache returned different length array")
	}

	for i := range result1 {
		if math.IsNaN(result1[i]) && math.IsNaN(result2[i]) {
			continue
		}
		if result1[i] != result2[i] {
			t.Errorf("Cache returned different value at index %d: %f vs %f", i, result1[i], result2[i])
		}
	}
}

func TestPivotResultCache_BuildCacheKey_Uniqueness(t *testing.T) {
	cache := NewPivotResultCache()

	key1 := cache.buildCacheKey(PivotTypeHigh, 5, 5, 100)
	key2 := cache.buildCacheKey(PivotTypeHigh, 5, 5, 100)
	key3 := cache.buildCacheKey(PivotTypeLow, 5, 5, 100)
	key4 := cache.buildCacheKey(PivotTypeHigh, 10, 5, 100)

	if key1 != key2 {
		t.Error("Same parameters should produce same key")
	}

	if key1 == key3 {
		t.Error("Different pivot types should produce different keys")
	}

	if key1 == key4 {
		t.Error("Different leftBars should produce different keys")
	}
}

func TestPivotResultCache_Clear_RemovesCache(t *testing.T) {
	cache := NewPivotResultCache()

	source := []float64{1, 2, 3, 4, 5}
	cache.ComputeOrRetrieve(PivotTypeHigh, source, 1, 1)

	if len(cache.cache) == 0 {
		t.Fatal("Cache should contain entry after compute")
	}

	cache.Clear()

	if len(cache.cache) != 0 {
		t.Error("Cache should be empty after Clear()")
	}
}
