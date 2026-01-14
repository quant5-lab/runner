package security

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

func TestSecurityCache_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "empty_symbol",
			test: func(t *testing.T) {
				cache := NewSecurityCache()
				ctx := context.New("", "1D", 1)
				entry := &CacheEntry{Context: ctx}
				cache.Set("", "1D", entry)

				retrieved, exists := cache.Get("", "1D")
				if !exists {
					t.Error("empty symbol should be valid key")
				}
				if retrieved.Context.Symbol != "" {
					t.Errorf("expected empty symbol, got %q", retrieved.Context.Symbol)
				}
			},
		},
		{
			name: "empty_timeframe",
			test: func(t *testing.T) {
				cache := NewSecurityCache()
				ctx := context.New("BTC", "", 1)
				entry := &CacheEntry{Context: ctx}
				cache.Set("BTC", "", entry)

				retrieved, exists := cache.Get("BTC", "")
				if !exists {
					t.Error("empty timeframe should be valid key")
				}
				if retrieved.Context.Timeframe != "" {
					t.Errorf("expected empty timeframe, got %q", retrieved.Context.Timeframe)
				}
			},
		},
		{
			name: "both_empty",
			test: func(t *testing.T) {
				cache := NewSecurityCache()
				ctx := context.New("", "", 1)
				entry := &CacheEntry{Context: ctx}
				cache.Set("", "", entry)

				_, exists := cache.Get("", "")
				if !exists {
					t.Error("both empty should be valid key")
				}
			},
		},
		{
			name: "special_characters_symbol",
			test: func(t *testing.T) {
				cache := NewSecurityCache()
				symbols := []string{"BTC:USD", "BTC/USDT", "BTC-PERP", "BTC.D"}
				for _, sym := range symbols {
					ctx := context.New(sym, "1D", 1)
					entry := &CacheEntry{Context: ctx}
					cache.Set(sym, "1D", entry)

					retrieved, exists := cache.Get(sym, "1D")
					if !exists {
						t.Errorf("symbol %q should be valid key", sym)
					}
					if retrieved.Context.Symbol != sym {
						t.Errorf("expected symbol %q, got %q", sym, retrieved.Context.Symbol)
					}
				}
			},
		},
		{
			name: "overwrite_entry",
			test: func(t *testing.T) {
				cache := NewSecurityCache()

				ctx1 := context.New("BTC", "1D", 10)
				entry1 := &CacheEntry{Context: ctx1}
				cache.Set("BTC", "1D", entry1)

				ctx2 := context.New("BTC", "1D", 20)
				ctx2.AddBar(context.OHLCV{Close: 100.0})
				ctx2.AddBar(context.OHLCV{Close: 101.0})
				entry2 := &CacheEntry{Context: ctx2}
				cache.Set("BTC", "1D", entry2)

				retrieved, _ := cache.Get("BTC", "1D")
				if len(retrieved.Context.Data) != 2 {
					t.Errorf("expected 2 bars (overwritten), got %d", len(retrieved.Context.Data))
				}

				if cache.Size() != 1 {
					t.Errorf("expected size 1 after overwrite, got %d", cache.Size())
				}
			},
		},
		{
			name: "nil_context",
			test: func(t *testing.T) {
				cache := NewSecurityCache()
				entry := &CacheEntry{Context: nil}
				cache.Set("TEST", "1D", entry)

				retrieved, exists := cache.Get("TEST", "1D")
				if !exists {
					t.Error("nil context entry should exist")
				}
				if retrieved.Context != nil {
					t.Error("expected nil context to remain nil")
				}
			},
		},
		{
			name: "unicode_symbols",
			test: func(t *testing.T) {
				cache := NewSecurityCache()
				symbols := []string{"币安", "ビットコイン", "비트코인", "₿TC"}
				for _, sym := range symbols {
					ctx := context.New(sym, "1D", 1)
					entry := &CacheEntry{Context: ctx}
					cache.Set(sym, "1D", entry)

					retrieved, exists := cache.Get(sym, "1D")
					if !exists {
						t.Errorf("unicode symbol %q should work", sym)
					}
					if retrieved.Context.Symbol != sym {
						t.Errorf("expected symbol %q, got %q", sym, retrieved.Context.Symbol)
					}
				}
			},
		},
		{
			name: "very_long_symbol",
			test: func(t *testing.T) {
				cache := NewSecurityCache()
				longSym := string(make([]byte, 1000))
				for i := range longSym {
					longSym = longSym[:i] + "A"
				}

				ctx := context.New(longSym, "1D", 1)
				entry := &CacheEntry{Context: ctx}
				cache.Set(longSym, "1D", entry)

				retrieved, exists := cache.Get(longSym, "1D")
				if !exists {
					t.Error("very long symbol should work")
				}
				if len(retrieved.Context.Symbol) != len(longSym) {
					t.Errorf("symbol length mismatch: expected %d, got %d", len(longSym), len(retrieved.Context.Symbol))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestSecurityCache_ConcurrentKeyGeneration(t *testing.T) {
	cache := NewSecurityCache()

	testCases := []struct {
		symbol        string
		timeframe     string
		shouldCollide bool
	}{
		{"BTC", "1D", false},
		{"BT", "C:1D", false}, // Could collide if key is naive concatenation
		{"B", "TC:1D", false},
		{"BTCUSDT", "1h", false},
		{"BTC:USDT", "1h", false}, // Different from above
		{"", "1D", false},
		{"BTC", "", false},
	}

	for _, tc := range testCases {
		ctx := context.New(tc.symbol, tc.timeframe, 1)
		entry := &CacheEntry{Context: ctx}
		cache.Set(tc.symbol, tc.timeframe, entry)
	}

	// All entries should be retrievable independently
	for _, tc := range testCases {
		retrieved, exists := cache.Get(tc.symbol, tc.timeframe)
		if !exists {
			t.Errorf("entry (%q, %q) should exist", tc.symbol, tc.timeframe)
		}
		if retrieved.Context.Symbol != tc.symbol {
			t.Errorf("symbol mismatch: expected %q, got %q", tc.symbol, retrieved.Context.Symbol)
		}
		if retrieved.Context.Timeframe != tc.timeframe {
			t.Errorf("timeframe mismatch: expected %q, got %q", tc.timeframe, retrieved.Context.Timeframe)
		}
	}

	expectedSize := len(testCases)
	if cache.Size() != expectedSize {
		t.Errorf("expected size %d, got %d - possible key collision", expectedSize, cache.Size())
	}
}

func TestSecurityCache_GetContextErrorMessages(t *testing.T) {
	tests := []struct {
		name      string
		symbol    string
		timeframe string
		setup     func(*SecurityCache)
		wantErr   bool
		contains  string
	}{
		{
			name:      "missing_entry",
			symbol:    "MISSING",
			timeframe: "1D",
			setup:     func(c *SecurityCache) {},
			wantErr:   true,
			contains:  "no cache entry for MISSING:1D",
		},
		{
			name:      "empty_symbol_missing",
			symbol:    "",
			timeframe: "1D",
			setup:     func(c *SecurityCache) {},
			wantErr:   true,
			contains:  "no cache entry for :1D",
		},
		{
			name:      "empty_timeframe_missing",
			symbol:    "BTC",
			timeframe: "",
			setup:     func(c *SecurityCache) {},
			wantErr:   true,
			contains:  "no cache entry for BTC:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewSecurityCache()
			tt.setup(cache)

			_, err := cache.GetContext(tt.symbol, tt.timeframe)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got error=%v", tt.wantErr, err)
			}
			if tt.wantErr && err != nil {
				if tt.contains != "" && err.Error() != tt.contains {
					t.Errorf("expected error containing %q, got %q", tt.contains, err.Error())
				}
			}
		})
	}
}

func TestSecurityCache_ClearIsolation(t *testing.T) {
	cache1 := NewSecurityCache()
	cache2 := NewSecurityCache()

	ctx1 := context.New("BTC", "1D", 1)
	entry1 := &CacheEntry{Context: ctx1}
	cache1.Set("BTC", "1D", entry1)

	ctx2 := context.New("ETH", "1h", 1)
	entry2 := &CacheEntry{Context: ctx2}
	cache2.Set("ETH", "1h", entry2)

	cache1.Clear()

	if cache1.Size() != 0 {
		t.Errorf("cache1 should be empty after clear, got size %d", cache1.Size())
	}

	if cache2.Size() != 1 {
		t.Errorf("cache2 should still have 1 entry, got size %d", cache2.Size())
	}

	_, exists := cache2.Get("ETH", "1h")
	if !exists {
		t.Error("cache2 entry should still exist after cache1 clear")
	}
}
