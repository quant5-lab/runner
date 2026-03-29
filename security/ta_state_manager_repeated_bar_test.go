package security

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestTAStateManager_RepeatedBarIdempotency verifies that calling ComputeAtBar
// for the same barIdx multiple times returns the identical value without
// mutating cursor state.
func TestTAStateManager_RepeatedBarIdempotency(t *testing.T) {
	for _, tt := range allTATypes {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createContextWithBars(30)
			m := createTAManager(tt.name, tt.cacheKey, tt.period, 30, NewStreamingBarEvaluator())
			src := &ast.Identifier{Name: "close"}

			targetBar := 20
			first, err := m.ComputeAtBar(ctx, src, targetBar)
			if err != nil {
				t.Fatalf("first call bar %d: %v", targetBar, err)
			}

			for rep := 1; rep <= 5; rep++ {
				v, err := m.ComputeAtBar(ctx, src, targetBar)
				if err != nil {
					t.Fatalf("rep %d bar %d: %v", rep, targetBar, err)
				}
				if !floatEq(v, first) {
					t.Errorf("rep %d: value mutated %.6f → %.6f", rep, first, v)
				}
			}
		})
	}
}

// TestTAStateManager_HistoricalAnchorStableAfterAdvance verifies that a value
// computed at an anchor bar remains identical after advancing the cursor past
// that bar and re-querying it multiple times.
func TestTAStateManager_HistoricalAnchorStableAfterAdvance(t *testing.T) {
	for _, tt := range allTATypes {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createContextWithBars(50)
			m := createTAManager(tt.name, tt.cacheKey, tt.period, 50, NewStreamingBarEvaluator())
			src := &ast.Identifier{Name: "close"}

			anchor := 15
			first, err := m.ComputeAtBar(ctx, src, anchor)
			if err != nil {
				t.Fatalf("anchor bar %d: %v", anchor, err)
			}

			_, _ = m.ComputeAtBar(ctx, src, 45)

			for rep := 1; rep <= 3; rep++ {
				v, err := m.ComputeAtBar(ctx, src, anchor)
				if err != nil {
					t.Fatalf("rep %d anchor bar %d: %v", rep, anchor, err)
				}
				if !floatEq(v, first) {
					t.Errorf("rep %d: anchor value changed after advance %.6f → %.6f", rep, first, v)
				}
			}
		})
	}
}

// TestTAStateManager_FullHistoricalConsistency computes all bars sequentially,
// saves the results, then re-queries every bar and verifies no value changed.
// This guards against any cursor-arithmetic regression across the full history.
func TestTAStateManager_FullHistoricalConsistency(t *testing.T) {
	dataSize := 30

	for _, tt := range allTATypes {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createContextWithBars(dataSize)
			m := createTAManager(tt.name, tt.cacheKey, tt.period, dataSize, NewStreamingBarEvaluator())
			src := &ast.Identifier{Name: "close"}

			saved := make([]float64, dataSize)
			for i := 0; i < dataSize; i++ {
				v, err := m.ComputeAtBar(ctx, src, i)
				if err != nil {
					t.Fatalf("bar %d: %v", i, err)
				}
				saved[i] = v
			}

			for i := 0; i < dataSize; i++ {
				v, err := m.ComputeAtBar(ctx, src, i)
				if err != nil {
					t.Fatalf("re-query bar %d: %v", i, err)
				}
				if !floatEq(v, saved[i]) {
					t.Errorf("bar %d: historical value changed %.6f → %.6f", i, saved[i], v)
				}
			}
		})
	}
}
