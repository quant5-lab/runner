package ticker

import "github.com/quant5-lab/runner/runtime/context"

// TransformResult pairs synthetic bars with a per-source-bar synthetic index lookup.
//
// MainToSynthetic[i] is the index of the last closed synthetic bar at the point when
// source bar i was processed.  -1 means no synthetic bar has closed yet at that point.
type TransformResult struct {
	Bars            []context.OHLCV
	MainToSynthetic []int
}

// identityMapping returns a slice where mapping[i] == i.
// Used by bar-count-preserving transforms (Identity, HeikinAshi).
func identityMapping(barCount int) []int {
	m := make([]int, barCount)
	for i := range m {
		m[i] = i
	}
	return m
}
