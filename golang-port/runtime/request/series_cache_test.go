package request

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/series"
)

func TestSeriesCache_GetSet(t *testing.T) {
	cache := NewSeriesCache()

	key := "test:key"
	seriesBuffer := series.NewSeries(10)

	_, found := cache.Get(key)
	if found {
		t.Error("Expected no entry in empty cache")
	}

	cache.Set(key, seriesBuffer)

	retrieved, found := cache.Get(key)
	if !found {
		t.Error("Expected to find cached series")
	}

	if retrieved != seriesBuffer {
		t.Error("Retrieved series does not match stored series")
	}
}

func TestSeriesCache_Clear(t *testing.T) {
	cache := NewSeriesCache()

	cache.Set("key1", series.NewSeries(10))
	cache.Set("key2", series.NewSeries(20))

	cache.Clear()

	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")

	if found1 || found2 {
		t.Error("Expected cache to be empty after Clear()")
	}
}

func TestBuildSeriesCacheKey(t *testing.T) {
	expr := &ast.Identifier{Name: "close"}

	key1 := BuildSeriesCacheKey("BTCUSD", "1D", expr)
	key2 := BuildSeriesCacheKey("BTCUSD", "1D", expr)
	key3 := BuildSeriesCacheKey("ETHUSD", "1D", expr)

	if key1 != key2 {
		t.Error("Expected same key for same inputs")
	}

	if key1 == key3 {
		t.Error("Expected different keys for different symbols")
	}
}
