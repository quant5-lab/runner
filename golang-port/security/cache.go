package security

import (
	"fmt"

	"github.com/quant5-lab/runner/runtime/context"
)

type CacheEntry struct {
	Context *context.Context
}

type SecurityCache struct {
	entries map[string]*CacheEntry
}

func NewSecurityCache() *SecurityCache {
	return &SecurityCache{
		entries: make(map[string]*CacheEntry),
	}
}

func (c *SecurityCache) Get(symbol, timeframe string) (*CacheEntry, bool) {
	key := fmt.Sprintf("%s:%s", symbol, timeframe)
	entry, exists := c.entries[key]
	return entry, exists
}

func (c *SecurityCache) Set(symbol, timeframe string, entry *CacheEntry) {
	key := fmt.Sprintf("%s:%s", symbol, timeframe)
	c.entries[key] = entry
}

func (c *SecurityCache) GetContext(symbol, timeframe string) (*context.Context, error) {
	entry, exists := c.Get(symbol, timeframe)
	if !exists {
		return nil, fmt.Errorf("no cache entry for %s:%s", symbol, timeframe)
	}

	return entry.Context, nil
}

func (c *SecurityCache) Clear() {
	c.entries = make(map[string]*CacheEntry)
}

func (c *SecurityCache) Size() int {
	return len(c.entries)
}
