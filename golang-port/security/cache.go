package security

import (
	"fmt"

	"github.com/borisquantlab/pinescript-go/runtime/context"
)

/* CacheEntry stores fetched context and evaluated expression values */
type CacheEntry struct {
	Context     *context.Context      /* Security context (OHLCV data) */
	Expressions map[string][]float64  /* expressionName -> evaluated values */
}

/* SecurityCache stores multi-timeframe data and expressions */
type SecurityCache struct {
	entries map[string]*CacheEntry /* "symbol:timeframe" -> entry */
}

/* NewSecurityCache creates empty cache */
func NewSecurityCache() *SecurityCache {
	return &SecurityCache{
		entries: make(map[string]*CacheEntry),
	}
}

/* Get retrieves cache entry for symbol and timeframe */
func (c *SecurityCache) Get(symbol, timeframe string) (*CacheEntry, bool) {
	key := fmt.Sprintf("%s:%s", symbol, timeframe)
	entry, exists := c.entries[key]
	return entry, exists
}

/* Set stores cache entry for symbol and timeframe */
func (c *SecurityCache) Set(symbol, timeframe string, entry *CacheEntry) {
	key := fmt.Sprintf("%s:%s", symbol, timeframe)
	c.entries[key] = entry
}

/* GetExpression retrieves specific expression values */
func (c *SecurityCache) GetExpression(symbol, timeframe, exprName string) ([]float64, error) {
	entry, exists := c.Get(symbol, timeframe)
	if !exists {
		return nil, fmt.Errorf("no cache entry for %s:%s", symbol, timeframe)
	}

	values, exists := entry.Expressions[exprName]
	if !exists {
		return nil, fmt.Errorf("expression %s not found in %s:%s", exprName, symbol, timeframe)
	}

	return values, nil
}

/* SetExpression stores expression values in existing entry */
func (c *SecurityCache) SetExpression(symbol, timeframe, exprName string, values []float64) error {
	entry, exists := c.Get(symbol, timeframe)
	if !exists {
		return fmt.Errorf("no cache entry for %s:%s", symbol, timeframe)
	}

	if entry.Expressions == nil {
		entry.Expressions = make(map[string][]float64)
	}

	entry.Expressions[exprName] = values
	return nil
}

/* Clear removes all cache entries */
func (c *SecurityCache) Clear() {
	c.entries = make(map[string]*CacheEntry)
}

/* Size returns number of cached entries */
func (c *SecurityCache) Size() int {
	return len(c.entries)
}
