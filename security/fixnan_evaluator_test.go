package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

// ── AST builders ──────────────────────────────────────────────────────────────

func makeFixnanCall(inner ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "fixnan"},
		Arguments: []ast.Expression{inner},
	}
}

func makeTAFixnanCall(inner ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "fixnan"},
		},
		Arguments: []ast.Expression{inner},
	}
}

func makePivotHighExpr(source ast.Expression, left, right int) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "pivothigh"},
		},
		Arguments: []ast.Expression{
			source,
			&ast.Literal{Value: float64(left)},
			&ast.Literal{Value: float64(right)},
		},
	}
}

func makePivotLowExpr(source ast.Expression, left, right int) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "pivotlow"},
		},
		Arguments: []ast.Expression{
			source,
			&ast.Literal{Value: float64(left)},
			&ast.Literal{Value: float64(right)},
		},
	}
}

func subscriptExpr(expr ast.Expression, offset int) *ast.MemberExpression {
	return &ast.MemberExpression{
		Object:   expr,
		Property: &ast.Literal{Value: float64(offset)},
	}
}

// ── context builders ──────────────────────────────────────────────────────────

func ohlcvFromHighs(highs []float64) *context.Context {
	data := make([]context.OHLCV, len(highs))
	for i, h := range highs {
		data[i].High = h
	}
	return &context.Context{Data: data}
}

func ohlcvFromHighsLows(highs, lows []float64) *context.Context {
	data := make([]context.OHLCV, len(highs))
	for i := range highs {
		data[i].High = highs[i]
		data[i].Low = lows[i]
	}
	return &context.Context{Data: data}
}

// ── evaluation helper ─────────────────────────────────────────────────────────

func evalFixnanAt(t *testing.T, ev *StreamingBarEvaluator, call *ast.CallExpression, ctx *context.Context, barIdx int) float64 {
	t.Helper()
	result, err := ev.EvaluateAtBar(call, ctx, barIdx)
	if err != nil {
		t.Fatalf("EvaluateAtBar(bar=%d): %v", barIdx, err)
	}
	return result
}

// ── output sequence ───────────────────────────────────────────────────────────

// TestFixnan_ForwardFillsLastValidValue verifies the full per-bar output
// sequence: NaN before any valid value appears, the pivot value on its first
// detection bar, the carried value on every subsequent NaN bar, and the
// updated value when a new pivot appears.
func TestFixnan_ForwardFillsLastValidValue(t *testing.T) {
	// high: [100,105,110,108,103,102,104,107,106,101]
	// pivothigh(2,2): NaN×4, 110 at bar4, NaN×4, 107 at bar9
	ctx := ohlcvFromHighs([]float64{100, 105, 110, 108, 103, 102, 104, 107, 106, 101})
	ev := NewStreamingBarEvaluator()
	call := makeFixnanCall(makePivotHighExpr(&ast.Identifier{Name: "high"}, 2, 2))

	tests := []struct {
		name string
		bar  int
		want float64
	}{
		{"bar0_no_pivot_yet", 0, math.NaN()},
		{"bar1_no_pivot_yet", 1, math.NaN()},
		{"bar2_no_pivot_yet", 2, math.NaN()},
		{"bar3_no_pivot_yet", 3, math.NaN()},
		{"bar4_first_pivot", 4, 110},
		{"bar5_carry_110", 5, 110},
		{"bar6_carry_110", 6, 110},
		{"bar7_carry_110", 7, 110},
		{"bar8_carry_110", 8, 110},
		{"bar9_second_pivot", 9, 107},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertFloat64(t, tt.name, evalFixnanAt(t, ev, call, ctx, tt.bar), tt.want, 0)
		})
	}
}

// ── correctness under arbitrary query order ───────────────────────────────────

// TestFixnan_OutOfOrderQueryConsistency verifies that querying bars in any
// order — including querying a later bar before an earlier one — always returns
// the value that would be produced by a strictly sequential 0…N scan. It covers
// both sparse sources (with NaN pivot gaps) and dense sources (never NaN).
func TestFixnan_OutOfOrderQueryConsistency(t *testing.T) {
	t.Run("sparse_source_with_nan_gaps", func(t *testing.T) {
		// high: [100,105,110,108,103,102,104,107,106,101]
		// pivothigh(2,2): bar4=110, bar9=107, all others NaN
		// fixnan sequence: NaN×4, 110, 110, 110, 110, 110, 107
		ctx := ohlcvFromHighs([]float64{100, 105, 110, 108, 103, 102, 104, 107, 106, 101})
		ev := NewStreamingBarEvaluator()
		call := makeFixnanCall(makePivotHighExpr(&ast.Identifier{Name: "high"}, 2, 2))

		assertFloat64(t, "bar9_first", evalFixnanAt(t, ev, call, ctx, 9), 107, 0)
		assertFloat64(t, "bar6_after_bar9", evalFixnanAt(t, ev, call, ctx, 6), 110, 0)
		assertFloat64(t, "bar4_pivot_bar", evalFixnanAt(t, ev, call, ctx, 4), 110, 0)
		assertFloat64(t, "bar1_before_any_pivot", evalFixnanAt(t, ev, call, ctx, 1), math.NaN(), 0)
	})

	t.Run("dense_source_always_valid", func(t *testing.T) {
		// close is always valid: fixnan is a pass-through, each bar holds its own value.
		ctx := closesOnlyCtx(100, 110, 120, 115, 125, 130)
		ev := NewStreamingBarEvaluator()
		call := makeFixnanCall(&ast.Identifier{Name: "close"})

		assertFloat64(t, "bar5_first", evalFixnanAt(t, ev, call, ctx, 5), 130, 0)
		assertFloat64(t, "bar2_after_bar5", evalFixnanAt(t, ev, call, ctx, 2), 120, 0)
		assertFloat64(t, "bar4_after_bar2", evalFixnanAt(t, ev, call, ctx, 4), 125, 0)
	})
}

// ── NaN boundary conditions ───────────────────────────────────────────────────

// TestFixnan_AllNaNSourceRemainsNaN verifies fixnan never fabricates a value
// when no non-NaN input has appeared on any bar up to the queried point.
func TestFixnan_AllNaNSourceRemainsNaN(t *testing.T) {
	ctx := ohlcvFromHighs([]float64{100, 102, 103, 101})
	ev := NewStreamingBarEvaluator()
	call := makeFixnanCall(makePivotHighExpr(&ast.Identifier{Name: "high"}, 2, 2))

	assertFloat64(t, "bar1_no_pivot_exists", evalFixnanAt(t, ev, call, ctx, 1), math.NaN(), 0)
}

// TestFixnan_SingleValidThenAllNaN verifies the carried value persists
// indefinitely once set, without any decay or reset.
func TestFixnan_SingleValidThenAllNaN(t *testing.T) {
	ctx := ohlcvFromHighs([]float64{100, 105, 110, 108, 103, 102, 101, 100, 99, 98})
	ev := NewStreamingBarEvaluator()
	call := makeFixnanCall(makePivotHighExpr(&ast.Identifier{Name: "high"}, 2, 2))

	assertFloat64(t, "bar4_pivot", evalFixnanAt(t, ev, call, ctx, 4), 110, 0)
	for i := 5; i <= 9; i++ {
		got := evalFixnanAt(t, ev, call, ctx, i)
		if !floatEq(got, 110) {
			t.Errorf("bar %d: expected forward-fill 110, got %.6f", i, got)
		}
	}
}

// TestFixnan_ExtremeValues verifies that zero, negative, and magnitude-extreme
// values are not misidentified as NaN sentinels and are correctly forward-filled.
func TestFixnan_ExtremeValues(t *testing.T) {
	nan := math.NaN()
	tests := []struct {
		name     string
		closes   []float64
		expected []float64
	}{
		{
			name:     "negative_values",
			closes:   []float64{-100.0, nan, nan, -50.0, nan},
			expected: []float64{-100.0, -100.0, -100.0, -50.0, -50.0},
		},
		{
			name:     "zero_is_not_na",
			closes:   []float64{0.0, nan, 10.0, 0.0, nan},
			expected: []float64{0.0, 0.0, 10.0, 0.0, 0.0},
		},
		{
			name:     "large_positive",
			closes:   []float64{1e10, nan, 1e11, nan},
			expected: []float64{1e10, 1e10, 1e11, 1e11},
		},
		{
			name:     "very_small",
			closes:   []float64{1e-10, nan, 1e-11, nan},
			expected: []float64{1e-10, 1e-10, 1e-11, 1e-11},
		},
		{
			name:     "alternating_valid_nan",
			closes:   []float64{10.0, nan, 20.0, nan, 30.0, nan},
			expected: []float64{10.0, 10.0, 20.0, 20.0, 30.0, 30.0},
		},
		{
			name:     "leading_nan_then_valid",
			closes:   []float64{nan, nan, 5.0, nan},
			expected: []float64{nan, nan, 5.0, 5.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := closesOnlyCtx(tt.closes...)
			ev := NewStreamingBarEvaluator()
			call := makeFixnanCall(&ast.Identifier{Name: "close"})

			for i, want := range tt.expected {
				assertFloat64(t, tt.name, evalFixnanAt(t, ev, call, ctx, i), want, 1e-9)
			}
		})
	}
}

// ── cache isolation ───────────────────────────────────────────────────────────

// TestFixnan_TwoExpressionsAreIsolated verifies that fixnan(pivothigh) and
// fixnan(pivotlow) maintain independent series on the same StreamingBarEvaluator.
func TestFixnan_TwoExpressionsAreIsolated(t *testing.T) {
	ctx := ohlcvFromHighsLows(
		[]float64{100, 105, 110, 108, 103},
		[]float64{90, 85, 80, 82, 87},
	)
	ev := NewStreamingBarEvaluator()

	highCall := makeFixnanCall(makePivotHighExpr(&ast.Identifier{Name: "high"}, 2, 2))
	lowCall := makeFixnanCall(makePivotLowExpr(&ast.Identifier{Name: "low"}, 2, 2))

	assertFloat64(t, "fixnan(pivothigh) bar4", evalFixnanAt(t, ev, highCall, ctx, 4), 110, 0)
	assertFloat64(t, "fixnan(pivotlow) bar4", evalFixnanAt(t, ev, lowCall, ctx, 4), 80, 0)
}

// ── access pattern invariants ─────────────────────────────────────────────────

// TestFixnan_IdempotentRequery verifies that re-querying the same bar multiple
// times never mutates the buffered result.
func TestFixnan_IdempotentRequery(t *testing.T) {
	ctx := closesOnlyCtx(10, math.NaN(), 30, math.NaN(), 50)
	ev := NewStreamingBarEvaluator()
	call := makeFixnanCall(&ast.Identifier{Name: "close"})

	first := evalFixnanAt(t, ev, call, ctx, 4)
	for rep := 1; rep <= 5; rep++ {
		got := evalFixnanAt(t, ev, call, ctx, 4)
		if !floatEq(got, first) {
			t.Errorf("rep %d: value mutated %.6f → %.6f", rep, first, got)
		}
	}
}

// TestFixnan_FullHistoricalConsistency verifies that every buffered bar value
// remains stable after the full sequence has been computed and re-queried.
func TestFixnan_FullHistoricalConsistency(t *testing.T) {
	closes := []float64{math.NaN(), 10, math.NaN(), 20, math.NaN(), 30, math.NaN(), 40}
	ctx := closesOnlyCtx(closes...)
	ev := NewStreamingBarEvaluator()
	call := makeFixnanCall(&ast.Identifier{Name: "close"})

	n := len(closes)
	saved := make([]float64, n)
	for i := 0; i < n; i++ {
		saved[i] = evalFixnanAt(t, ev, call, ctx, i)
	}

	for i := 0; i < n; i++ {
		got := evalFixnanAt(t, ev, call, ctx, i)
		if !floatEq(got, saved[i]) {
			t.Errorf("bar %d: historical value changed %.6f → %.6f", i, saved[i], got)
		}
	}
}

// ── context growth ────────────────────────────────────────────────────────────

// TestFixnan_GrowsWhenContextExpands verifies that presenting a larger context
// after initial computation produces correct results for all bars, including
// those computed under the original smaller context.
func TestFixnan_GrowsWhenContextExpands(t *testing.T) {
	ev := NewStreamingBarEvaluator()
	call := makeFixnanCall(&ast.Identifier{Name: "close"})

	smallCtx := closesOnlyCtx(10, math.NaN(), 30, math.NaN(), 50)
	assertFloat64(t, "bar4_small_ctx", evalFixnanAt(t, ev, call, smallCtx, 4), 50, 0)

	largeCtx := closesOnlyCtx(10, math.NaN(), 30, math.NaN(), 50, math.NaN(), 70, math.NaN())
	assertFloat64(t, "bar7_large_ctx", evalFixnanAt(t, ev, call, largeCtx, 7), 70, 0)
	assertFloat64(t, "bar4_recheck_after_growth", evalFixnanAt(t, ev, call, largeCtx, 4), 50, 0)
}

// ── large gap ─────────────────────────────────────────────────────────────────

// TestFixnan_LargeGapForwardFill verifies that a value set early in a long series
// is correctly carried to a bar queried far later without accumulating error.
func TestFixnan_LargeGapForwardFill(t *testing.T) {
	data := make([]context.OHLCV, 1002)
	data[0].High = 100
	data[1].High = 105
	data[2].High = 110
	data[3].High = 108
	data[4].High = 103
	for i := 5; i < 1002; i++ {
		data[i].High = 102
	}
	ctx := &context.Context{Data: data}
	ev := NewStreamingBarEvaluator()
	call := makeFixnanCall(makePivotHighExpr(&ast.Identifier{Name: "high"}, 2, 2))

	result, err := ev.EvaluateAtBar(call, ctx, 1000)
	if err != nil {
		t.Fatalf("large gap forward-fill: %v", err)
	}
	assertFloat64(t, "bar1000", result, 110, 0)
}

// ── argument validation ───────────────────────────────────────────────────────

// TestFixnan_NoArgumentsReturnsError verifies that calling fixnan with no
// arguments produces an error rather than a silent NaN or panic.
func TestFixnan_NoArgumentsReturnsError(t *testing.T) {
	ctx := &context.Context{Data: []context.OHLCV{{High: 100}}}
	ev := NewStreamingBarEvaluator()
	call := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "fixnan"},
		Arguments: []ast.Expression{},
	}

	_, err := ev.EvaluateAtBar(call, ctx, 0)
	if err == nil {
		t.Error("expected error for zero arguments")
	}
}

// ── subscripted inner expression ──────────────────────────────────────────────

// TestFixnan_SubscriptedInnerExpression verifies fixnan(expr[N]) where the inner
// expression carries a historical subscript: the subscript shifts which bar's
// value is visible at each position, and out-of-range subscripts at early bars
// are treated as na (forward-fill) without error.
func TestFixnan_SubscriptedInnerExpression(t *testing.T) {
	ctx := ohlcvFromHighs([]float64{100, 105, 110, 108, 103, 102, 104, 107, 106, 101})
	pivotBase := makePivotHighExpr(&ast.Identifier{Name: "high"}, 2, 2)

	t.Run("offset_1_shifts_pivot_into_view", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		call := makeFixnanCall(subscriptExpr(pivotBase, 1))
		// pivot fires at bar 4; with [1] shift, it is visible at bar 5
		assertFloat64(t, "bar5", evalFixnanAt(t, ev, call, ctx, 5), 110, 0)
		assertFloat64(t, "bar6_carry", evalFixnanAt(t, ev, call, ctx, 6), 110, 0)
	})

	t.Run("offset_2_no_error", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		call := makeFixnanCall(subscriptExpr(pivotBase, 2))
		// out-of-range subscripts at early bars must not produce an error
		_, err := ev.EvaluateAtBar(call, ctx, 9)
		if err != nil {
			t.Errorf("offset_2 at bar 9 should not error: %v", err)
		}
	})
}

// ── arbitrary source expressions ──────────────────────────────────────────────

// TestFixnan_WithTAFunctions verifies fixnan handles TA function inner
// expressions, forward-filling through their warmup NaN period, and that two
// distinct TA expressions backed by the same evaluator maintain independent
// state.
func TestFixnan_WithTAFunctions(t *testing.T) {
	ctx := closesOnlyCtx(100, 102, 104, 106, 108, 110, 112, 114)
	closeIdent := &ast.Identifier{Name: "close"}

	t.Run("sma_warmup_nans_filled", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		call := makeFixnanCall(makeCallWithExprSource("sma", closeIdent, 3))
		want := (100.0 + 102.0 + 104.0) / 3.0
		assertFloat64(t, "bar2_sma3", evalFixnanAt(t, ev, call, ctx, 2), want, 0.01)
	})

	t.Run("ema_produces_valid_result", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		call := makeFixnanCall(makeCallWithExprSource("ema", closeIdent, 3))
		result := evalFixnanAt(t, ev, call, ctx, 5)
		if math.IsNaN(result) || result <= 0 {
			t.Errorf("expected valid EMA at bar 5, got %.6f", result)
		}
	})

	t.Run("two_distinct_expressions_use_separate_state", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		smaCall := makeFixnanCall(makeCallWithExprSource("sma", closeIdent, 3))
		emaCall := makeFixnanCall(makeCallWithExprSource("ema", closeIdent, 3))

		// Compute ema first; then verify sma at the same bar reflects only sma
		// state — proving the two expressions use independent cache entries.
		evalFixnanAt(t, ev, emaCall, ctx, 5)
		// sma(close,3) at bar 5: (close[3]+close[4]+close[5])/3 = (106+108+110)/3 = 108
		assertFloat64(t, "sma3_bar5_after_ema", evalFixnanAt(t, ev, smaCall, ctx, 5), 108, 0.01)
	})
}

// ── namespace alias ───────────────────────────────────────────────────────────

// TestFixnan_TANamespaceAliasIsEquivalent verifies that ta.fixnan and fixnan are
// registered to the same handler and produce identical results.
func TestFixnan_TANamespaceAliasIsEquivalent(t *testing.T) {
	ctx := ohlcvFromHighs([]float64{100, 105, 110, 108, 103, 102, 104, 107, 106, 101})
	pivotExpr := makePivotHighExpr(&ast.Identifier{Name: "high"}, 2, 2)

	ev1 := NewStreamingBarEvaluator()
	result1 := evalFixnanAt(t, ev1, makeFixnanCall(pivotExpr), ctx, 4)

	ev2 := NewStreamingBarEvaluator()
	result2 := evalFixnanAt(t, ev2, makeTAFixnanCall(pivotExpr), ctx, 4)

	if !floatEq(result1, result2) {
		t.Errorf("fixnan and ta.fixnan diverge: %.6f vs %.6f", result1, result2)
	}
}
