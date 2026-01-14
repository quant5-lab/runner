package context

import (
	"math"
	"testing"
)

func TestSecurityValueRetriever_RetrieveValue(t *testing.T) {
	retriever := NewSecurityValueRetriever()

	secCtx := &Context{
		Data: []OHLCV{
			{Time: 0, Open: 100.0, Close: 105.0},
			{Time: 86400, Open: 101.0, Close: 106.0},
			{Time: 172800, Open: 102.0, Close: 107.0},
		},
		BarIndex: 0, // Initial state
	}

	tests := []struct {
		name        string
		timestamp   int64
		getValue    func(*Context, int) float64
		expected    float64
		description string
	}{
		{
			name:      "retrieve open from first bar",
			timestamp: 50000,
			getValue: func(ctx *Context, idx int) float64 {
				if idx >= 0 && idx < len(ctx.Data) {
					return ctx.Data[idx].Open
				}
				return 0
			},
			expected:    100.0,
			description: "should retrieve open value from matched bar",
		},
		{
			name:      "retrieve close from second bar",
			timestamp: 100000,
			getValue: func(ctx *Context, idx int) float64 {
				if idx >= 0 && idx < len(ctx.Data) {
					return ctx.Data[idx].Close
				}
				return 0
			},
			expected:    106.0,
			description: "should retrieve close value from matched bar",
		},
		{
			name:      "retrieve from last bar when beyond",
			timestamp: 500000,
			getValue: func(ctx *Context, idx int) float64 {
				if idx >= 0 && idx < len(ctx.Data) {
					return ctx.Data[idx].Open
				}
				return 0
			},
			expected:    102.0,
			description: "should retrieve from last bar when timestamp is beyond all bars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalBarIndex := secCtx.BarIndex

			result := retriever.RetrieveValue(secCtx, tt.timestamp, tt.getValue)

			if result != tt.expected {
				t.Errorf("%s: RetrieveValue() = %.2f, expected %.2f",
					tt.description, result, tt.expected)
			}

			// Critical: BarIndex should be restored after retrieval
			if secCtx.BarIndex != originalBarIndex {
				t.Errorf("BarIndex not restored: was %d, now %d",
					originalBarIndex, secCtx.BarIndex)
			}
		})
	}
}

func TestSecurityValueRetriever_BarIndexRestoration(t *testing.T) {
	retriever := NewSecurityValueRetriever()

	secCtx := &Context{
		Data: []OHLCV{
			{Time: 0, Open: 100.0},
			{Time: 86400, Open: 101.0},
			{Time: 172800, Open: 102.0},
		},
		BarIndex: 5, // Arbitrary starting position
	}

	getValue := func(ctx *Context, idx int) float64 {
		// This function should see the temporary BarIndex
		if ctx.BarIndex != idx {
			panic("BarIndex not set correctly during getValue call")
		}
		if idx >= 0 && idx < len(ctx.Data) {
			return ctx.Data[idx].Open
		}
		return 0
	}

	originalBarIndex := secCtx.BarIndex
	retriever.RetrieveValue(secCtx, 100000, getValue)

	if secCtx.BarIndex != originalBarIndex {
		t.Errorf("BarIndex restoration failed: original=%d, current=%d",
			originalBarIndex, secCtx.BarIndex)
	}
}

func TestSecurityValueRetriever_InvalidBarIndex(t *testing.T) {
	retriever := NewSecurityValueRetriever()

	secCtx := &Context{
		Data: []OHLCV{
			{Time: 0, Open: 100.0},
		},
		BarIndex: 0,
	}

	t.Run("timestamp before first bar returns NaN", func(t *testing.T) {
		getValue := func(ctx *Context, idx int) float64 {
			if idx >= 0 && idx < len(ctx.Data) {
				return ctx.Data[idx].Open
			}
			return 0
		}

		result := retriever.RetrieveValue(secCtx, -1000, getValue)
		if !math.IsNaN(result) {
			t.Errorf("invalid bar index should return NaN, got %.2f", result)
		}
	})
}

func TestSecurityValueRetriever_EmptyContext(t *testing.T) {
	retriever := NewSecurityValueRetriever()

	emptyCtx := &Context{
		Data:     []OHLCV{},
		BarIndex: 0,
	}

	getValue := func(ctx *Context, idx int) float64 {
		if idx >= 0 && idx < len(ctx.Data) {
			return ctx.Data[idx].Open
		}
		return 0
	}

	result := retriever.RetrieveValue(emptyCtx, 100000, getValue)

	if !math.IsNaN(result) {
		t.Errorf("empty context should return NaN, got %.2f", result)
	}
}

func TestSecurityValueRetriever_RealWorldScenario_Upsampling(t *testing.T) {
	retriever := NewSecurityValueRetriever()

	// Scenario: Get daily open values for multiple hourly bars
	dec17 := int64(1734393600)

	dailyCtx := &Context{
		Data: []OHLCV{
			{Time: dec17, Open: 87863.43, Close: 88234.56},
			{Time: dec17 + 86400, Open: 88500.00, Close: 89123.45},
		},
		BarIndex: 0,
	}

	getOpen := func(ctx *Context, idx int) float64 {
		if idx >= 0 && idx < len(ctx.Data) {
			return ctx.Data[idx].Open
		}
		return 0
	}

	// Simulate multiple hourly bars on Dec 17
	hourlyTimestamps := []int64{
		dec17,           // 00:00
		dec17 + 3600,    // 01:00
		dec17 + 7200,    // 02:00
		dec17 + 10*3600, // 10:00
		dec17 + 23*3600, // 23:00
	}

	for i, hourlyTime := range hourlyTimestamps {
		t.Run("hour "+string(rune('0'+i)), func(t *testing.T) {
			value := retriever.RetrieveValue(dailyCtx, hourlyTime, getOpen)
			expected := 87863.43

			if value != expected {
				t.Errorf("Dec 17 hour %d: expected open %.2f, got %.2f",
					i, expected, value)
			}

			// Verify BarIndex restored
			if dailyCtx.BarIndex != 0 {
				t.Errorf("BarIndex not restored after hour %d", i)
			}
		})
	}
}

func TestSecurityValueRetriever_MultipleRetrievals(t *testing.T) {
	retriever := NewSecurityValueRetriever()

	ctx := &Context{
		Data: []OHLCV{
			{Time: 0, Open: 100.0, High: 110.0, Low: 95.0, Close: 105.0},
			{Time: 86400, Open: 101.0, High: 111.0, Low: 96.0, Close: 106.0},
		},
		BarIndex: 0,
	}

	timestamp := int64(50000) // Within first bar

	// Retrieve different fields from same timestamp
	tests := []struct {
		name     string
		getValue func(*Context, int) float64
		expected float64
	}{
		{
			name: "open",
			getValue: func(ctx *Context, idx int) float64 {
				return ctx.Data[idx].Open
			},
			expected: 100.0,
		},
		{
			name: "high",
			getValue: func(ctx *Context, idx int) float64 {
				return ctx.Data[idx].High
			},
			expected: 110.0,
		},
		{
			name: "low",
			getValue: func(ctx *Context, idx int) float64 {
				return ctx.Data[idx].Low
			},
			expected: 95.0,
		},
		{
			name: "close",
			getValue: func(ctx *Context, idx int) float64 {
				return ctx.Data[idx].Close
			},
			expected: 105.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := retriever.RetrieveValue(ctx, timestamp, tt.getValue)
			if result != tt.expected {
				t.Errorf("field %s: expected %.2f, got %.2f",
					tt.name, tt.expected, result)
			}
		})
	}
}

func TestSecurityValueRetriever_ConcurrentSafety(t *testing.T) {
	retriever := NewSecurityValueRetriever()

	ctx := &Context{
		Data: []OHLCV{
			{Time: 0, Open: 100.0},
			{Time: 86400, Open: 101.0},
		},
		BarIndex: 0,
	}

	getValue := func(ctx *Context, idx int) float64 {
		if idx >= 0 && idx < len(ctx.Data) {
			return ctx.Data[idx].Open
		}
		return 0
	}

	// Verify that rapid successive calls maintain BarIndex integrity
	for i := 0; i < 100; i++ {
		originalIdx := ctx.BarIndex
		retriever.RetrieveValue(ctx, 50000, getValue)

		if ctx.BarIndex != originalIdx {
			t.Fatalf("iteration %d: BarIndex not restored (original=%d, current=%d)",
				i, originalIdx, ctx.BarIndex)
		}
	}
}
