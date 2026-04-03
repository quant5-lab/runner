package ticker

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

// assertMappingLength verifies MainToSynthetic has one entry per source bar.
func assertMappingLength(t *testing.T, result TransformResult, inputLen int) {
	t.Helper()
	if len(result.MainToSynthetic) != inputLen {
		t.Errorf("MainToSynthetic length = %d, want %d", len(result.MainToSynthetic), inputLen)
	}
}

// assertMappingNonDecreasing verifies synthetic indices never decrease.
// Sentinel value -1 (no synthetic bar yet) satisfies this because -1 < any non-negative index.
func assertMappingNonDecreasing(t *testing.T, result TransformResult) {
	t.Helper()
	for i := 1; i < len(result.MainToSynthetic); i++ {
		if result.MainToSynthetic[i] < result.MainToSynthetic[i-1] {
			t.Errorf("MainToSynthetic[%d]=%d < MainToSynthetic[%d]=%d (must be non-decreasing)",
				i, result.MainToSynthetic[i], i-1, result.MainToSynthetic[i-1])
		}
	}
}

// assertMappingValidIndices verifies every mapping entry is a valid index into Bars.
// Do not call this for transformers that leave sentinel -1 values in the mapping (e.g. Renko).
func assertMappingValidIndices(t *testing.T, result TransformResult) {
	t.Helper()
	for i, idx := range result.MainToSynthetic {
		if idx < 0 || idx >= len(result.Bars) {
			t.Errorf("MainToSynthetic[%d]=%d out of range [0, %d)", i, idx, len(result.Bars))
		}
	}
}

// assertOHLCInvariants verifies H >= max(O,C) and L <= min(O,C) for every bar.
func assertOHLCInvariants(t *testing.T, bars []context.OHLCV) {
	t.Helper()
	for i, b := range bars {
		if b.High < b.Open || b.High < b.Close {
			t.Errorf("bar %d: High=%.4f violates H>=max(O=%.4f,C=%.4f)", i, b.High, b.Open, b.Close)
		}
		if b.Low > b.Open || b.Low > b.Close {
			t.Errorf("bar %d: Low=%.4f violates L<=min(O=%.4f,C=%.4f)", i, b.Low, b.Open, b.Close)
		}
	}
}
