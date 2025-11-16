package security

import (
	"testing"

	"github.com/borisquantlab/pinescript-go/runtime/context"
)

func TestSecurityCache_SetAndGet(t *testing.T) {
	cache := NewSecurityCache()

	/* Create test entry */
	ctx := context.New("BTC", "1D", 10)
	entry := &CacheEntry{
		Context:     ctx,
		Expressions: map[string][]float64{"sma20": {100, 101, 102}},
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

	if len(retrieved.Expressions["sma20"]) != 3 {
		t.Errorf("Expected 3 values, got %d", len(retrieved.Expressions["sma20"]))
	}
}

func TestSecurityCache_GetNonexistent(t *testing.T) {
	cache := NewSecurityCache()

	_, exists := cache.Get("ETH", "1h")
	if exists {
		t.Error("Expected nonexistent entry to return false")
	}
}

func TestSecurityCache_GetExpression(t *testing.T) {
	cache := NewSecurityCache()

	ctx := context.New("TEST", "1h", 5)
	entry := &CacheEntry{
		Context: ctx,
		Expressions: map[string][]float64{
			"close":  {100, 101, 102, 103, 104},
			"sma10":  {99, 100, 101, 102, 103},
		},
	}

	cache.Set("TEST", "1h", entry)

	/* Get existing expression */
	values, err := cache.GetExpression("TEST", "1h", "sma10")
	if err != nil {
		t.Fatalf("GetExpression failed: %v", err)
	}

	if len(values) != 5 {
		t.Errorf("Expected 5 values, got %d", len(values))
	}

	if values[0] != 99 {
		t.Errorf("Expected first value 99, got %.2f", values[0])
	}
}

func TestSecurityCache_GetExpressionNotFound(t *testing.T) {
	cache := NewSecurityCache()

	ctx := context.New("TEST", "1D", 1)
	entry := &CacheEntry{
		Context:     ctx,
		Expressions: map[string][]float64{},
	}

	cache.Set("TEST", "1D", entry)

	_, err := cache.GetExpression("TEST", "1D", "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent expression")
	}
}

func TestSecurityCache_SetExpression(t *testing.T) {
	cache := NewSecurityCache()

	/* Create entry without expressions */
	ctx := context.New("TEST", "1W", 3)
	entry := &CacheEntry{Context: ctx}

	cache.Set("TEST", "1W", entry)

	/* Add expression */
	values := []float64{10, 20, 30}
	err := cache.SetExpression("TEST", "1W", "ema9", values)
	if err != nil {
		t.Fatalf("SetExpression failed: %v", err)
	}

	/* Verify expression was stored */
	retrieved, _ := cache.GetExpression("TEST", "1W", "ema9")
	if len(retrieved) != 3 {
		t.Errorf("Expected 3 values, got %d", len(retrieved))
	}
}

func TestSecurityCache_SetExpressionNoEntry(t *testing.T) {
	cache := NewSecurityCache()

	err := cache.SetExpression("NONE", "1m", "test", []float64{1})
	if err == nil {
		t.Error("Expected error for nonexistent entry")
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

func TestSecurityCache_MultipleExpressions(t *testing.T) {
	cache := NewSecurityCache()

	ctx := context.New("MULTI", "1D", 2)
	entry := &CacheEntry{
		Context: ctx,
		Expressions: map[string][]float64{
			"sma":  {100, 101},
			"ema":  {102, 103},
			"rsi":  {50, 51},
		},
	}

	cache.Set("MULTI", "1D", entry)

	/* Verify all expressions */
	for name := range entry.Expressions {
		vals, err := cache.GetExpression("MULTI", "1D", name)
		if err != nil {
			t.Errorf("Failed to get expression %s: %v", name, err)
		}
		if len(vals) != 2 {
			t.Errorf("Expression %s: expected 2 values, got %d", name, len(vals))
		}
	}
}
