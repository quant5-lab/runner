package security

import (
	"fmt"
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

/* BenchmarkDirectContextAccess measures O(1) runtime pattern used by codegen */
func BenchmarkDirectContextAccess(b *testing.B) {
	sizes := []int{100, 500, 1000, 5000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("bars_%d", size), func(b *testing.B) {
			/* Setup context with N bars */
			secCtx := context.New("BTCUSDT", "1D", size)
			for i := 0; i < size; i++ {
				secCtx.AddBar(context.OHLCV{
					Time:   int64(i * 86400),
					Open:   100.0 + float64(i),
					High:   105.0 + float64(i),
					Low:    95.0 + float64(i),
					Close:  100.0 + float64(i),
					Volume: 1000.0,
				})
			}

			/* Simulate main strategy bar loop accessing security context */
			barIndex := size / 2 // midpoint access

			b.ResetTimer()
			b.ReportAllocs()

			/* Measure: Direct O(1) access pattern (what codegen generates) */
			for i := 0; i < b.N; i++ {
				/* This is what generated code does: secCtx.Data[secBarIdx].Close */
				_ = secCtx.Data[barIndex].Close
			}
		})
	}
}

/* BenchmarkDirectContextAccessLoop simulates per-bar security lookup in main loop */
func BenchmarkDirectContextAccessLoop(b *testing.B) {
	sizes := []int{100, 500, 1000, 5000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("bars_%d", size), func(b *testing.B) {
			/* Setup main context */
			mainCtx := context.New("BTCUSDT", "1h", size)
			for i := 0; i < size; i++ {
				mainCtx.AddBar(context.OHLCV{
					Time:  int64(i * 3600),
					Close: 100.0 + float64(i%10),
				})
			}

			/* Setup security context (daily) */
			secCtx := context.New("BTCUSDT", "1D", size/24)
			for i := 0; i < size/24; i++ {
				secCtx.AddBar(context.OHLCV{
					Time:  int64(i * 86400),
					Close: 100.0 + float64(i),
				})
			}

			b.ResetTimer()
			b.ReportAllocs()

			/* Measure: Full bar loop with security lookup (runtime pattern) */
			for n := 0; n < b.N; n++ {
				for i := 0; i < size; i++ {
					/* Find matching bar in security context */
					secBarIdx := context.FindBarIndexByTimestamp(secCtx, mainCtx.Data[i].Time)
					if secBarIdx >= 0 {
						/* Direct O(1) access */
						_ = secCtx.Data[secBarIdx].Close
					}
				}
			}
		})
	}
}
