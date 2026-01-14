package request

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/series"
)

type SeriesCache struct {
	cache map[string]*series.Series
}

func NewSeriesCache() *SeriesCache {
	return &SeriesCache{
		cache: make(map[string]*series.Series),
	}
}

func (c *SeriesCache) Get(key string) (*series.Series, bool) {
	seriesBuffer, found := c.cache[key]
	return seriesBuffer, found
}

func (c *SeriesCache) Set(key string, seriesBuffer *series.Series) {
	c.cache[key] = seriesBuffer
}

func (c *SeriesCache) Clear() {
	c.cache = make(map[string]*series.Series)
}

func BuildSeriesCacheKey(symbol, timeframe string, expr ast.Expression) string {
	return fmt.Sprintf("%s:%s:%p", symbol, timeframe, expr)
}
