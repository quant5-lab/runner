package security

import (
	"fmt"

	"github.com/quant5-lab/runner/runtime/context"
)

/* CacheEntry stores fetched context only (O(1) per-bar access pattern) */
type CacheEntry struct {
	Context *context.Context /* Security context (OHLCV data) */
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

/* GetContext retrieves context for symbol and timeframe */
func (c *SecurityCache) GetContext(symbol, timeframe string) (*context.Context, error) {
	entry, exists := c.Get(symbol, timeframe)
	if !exists {
		return nil, fmt.Errorf("no cache entry for %s:%s", symbol, timeframe)
	}

	return entry.Context, nil
}

/* Clear removes all cache entries */
func (c *SecurityCache) Clear() {
	c.entries = make(map[string]*CacheEntry)
}

/* Size returns number of cached entries */
func (c *SecurityCache) Size() int {
	return len(c.entries)
}
