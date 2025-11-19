package security

import (
	"testing"

	"github.com/borisquantlab/pinescript-go/runtime/context"
)

func TestSecurityCache_SetAndGet(t *testing.T) {
	cache := NewSecurityCache()

	/* Create test entry with context only */
	ctx := context.New("BTC", "1D", 10)
	entry := &CacheEntry{
		Context: ctx,
	}

	/* Store entry */
	cache.Set("BTC", "1D", entry)

	/* Retrieve entry */
	retrieved, exists := cache.Get("BTC", "1D")
	if !exists {
		t.Fatal("Expected entry to exist")
	}

	if retrieved.Context.Symbol != "BTC" {
		t.Errorf("Expected symbol BTC, got %s", retrieved.Context.Symbol)
	}

	if retrieved.Context.Timeframe != "1D" {
		t.Errorf("Expected timeframe 1D, got %s", retrieved.Context.Timeframe)
	}
}

func TestSecurityCache_GetNonexistent(t *testing.T) {
	cache := NewSecurityCache()

	_, exists := cache.Get("ETH", "1h")
	if exists {
		t.Error("Expected nonexistent entry to return false")
	}
}

func TestSecurityCache_GetContext(t *testing.T) {
	cache := NewSecurityCache()

	ctx := context.New("TEST", "1h", 5)
	entry := &CacheEntry{
		Context: ctx,
	}

	cache.Set("TEST", "1h", entry)

	/* Get context */
	retrieved, err := cache.GetContext("TEST", "1h")
	if err != nil {
		t.Fatalf("GetContext failed: %v", err)
	}

	if retrieved.Symbol != "TEST" {
		t.Errorf("Expected symbol TEST, got %s", retrieved.Symbol)
	}

	if retrieved.Timeframe != "1h" {
		t.Errorf("Expected timeframe 1h, got %s", retrieved.Timeframe)
	}
}

func TestSecurityCache_GetContextNotFound(t *testing.T) {
	cache := NewSecurityCache()

	_, err := cache.GetContext("NONE", "1D")
	if err == nil {
		t.Error("Expected error for nonexistent context")
	}
}

func TestSecurityCache_Clear(t *testing.T) {
	cache := NewSecurityCache()

	/* Add entries */
	cache.Set("BTC", "1h", &CacheEntry{Context: context.New("BTC", "1h", 1)})
	cache.Set("ETH", "1D", &CacheEntry{Context: context.New("ETH", "1D", 1)})

	if cache.Size() != 2 {
		t.Errorf("Expected size 2, got %d", cache.Size())
	}

	/* Clear cache */
	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}

	_, exists := cache.Get("BTC", "1h")
	if exists {
		t.Error("Expected entry to not exist after clear")
	}
}

func TestSecurityCache_MultipleContexts(t *testing.T) {
	cache := NewSecurityCache()

	/* Add multiple contexts */
	cache.Set("BTC", "1h", &CacheEntry{Context: context.New("BTC", "1h", 100)})
	cache.Set("ETH", "1D", &CacheEntry{Context: context.New("ETH", "1D", 50)})
	cache.Set("SOL", "1W", &CacheEntry{Context: context.New("SOL", "1W", 10)})

	if cache.Size() != 3 {
		t.Errorf("Expected size 3, got %d", cache.Size())
	}

	/* Verify all contexts */
	btcCtx, err := cache.GetContext("BTC", "1h")
	if err != nil || btcCtx.Symbol != "BTC" {
		t.Error("Failed to retrieve BTC context")
	}

	ethCtx, err := cache.GetContext("ETH", "1D")
	if err != nil || ethCtx.Symbol != "ETH" {
		t.Error("Failed to retrieve ETH context")
	}

	solCtx, err := cache.GetContext("SOL", "1W")
	if err != nil || solCtx.Symbol != "SOL" {
		t.Error("Failed to retrieve SOL context")
	}
}
