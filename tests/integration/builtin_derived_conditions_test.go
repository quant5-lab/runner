package integration

import (
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* Derived builtins in standalone ternary conditions — the exact path
 * through generateConditionExpression that previously failed for
 * hl2/hlc3/ohlc4/hlcc4/tr (generated non-existent xxxSeries variables).
 */
func TestBuiltinDerivedInTernaryConditions(t *testing.T) {
	pineScript := `//@version=5
indicator("Derived Builtins in Ternary Conditions", overlay=false)
hl2_cond   = hl2   > close ? 1.0 : 0.0
hlc3_cond  = hlc3  > close ? 1.0 : 0.0
ohlc4_cond = ohlc4 > close ? 1.0 : 0.0
hlcc4_cond = hlcc4 > close ? 1.0 : 0.0
tr_cond    = tr    > 0     ? 1.0 : 0.0
plot(hl2_cond,   "hl2_cond")
plot(hlc3_cond,  "hlc3_cond")
plot(ohlc4_cond, "ohlc4_cond")
plot(hlcc4_cond, "hlcc4_cond")
plot(tr_cond,    "tr_cond")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "builtin-derived-ternary-cond", pineScript)

	plots := []string{"hl2_cond", "hlc3_cond", "ohlc4_cond", "hlcc4_cond", "tr_cond"}
	for _, name := range plots {
		values := exec.ExtractPlotValues(t, output, name)
		if len(values) < 10 {
			t.Fatalf("%s: expected at least 10 bars, got %d", name, len(values))
		}
		for i := 1; i < minIntBuiltin(len(values), 20); i++ {
			v := values[i]
			if v != 0.0 && v != 1.0 {
				t.Errorf("%s[%d] = %f, want 0.0 or 1.0", name, i, v)
			}
		}
	}
}

/* Compound boolean conditions mixing multiple derived builtins
 * via `and`/`or` operators in a single ternary expression.
 */
func TestBuiltinDerivedInCompoundConditions(t *testing.T) {
	pineScript := `//@version=5
indicator("Derived Builtins in Compound Conditions", overlay=false)
compound = close > ohlc4 and hlc3 > hlcc4 ? 1.0 : 0.0
mixed    = hl2 > close or ohlc4 < close   ? 1.0 : 0.0
plot(compound, "compound")
plot(mixed,    "mixed")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "builtin-derived-compound-cond", pineScript)

	for _, name := range []string{"compound", "mixed"} {
		values := exec.ExtractPlotValues(t, output, name)
		if len(values) < 10 {
			t.Fatalf("%s: expected at least 10 bars, got %d", name, len(values))
		}
		for i := 0; i < minIntBuiltin(len(values), 20); i++ {
			v := values[i]
			if v != 0.0 && v != 1.0 {
				t.Errorf("%s[%d] = %f, want 0.0 or 1.0", name, i, v)
			}
		}
	}
}

/* Derived builtins as ternary branch values — both the condition operand
 * and the result branches exercise the identifier resolution path.
 */
func TestBuiltinDerivedAsTernaryBranchValues(t *testing.T) {
	pineScript := `//@version=5
indicator("Derived Builtins as Ternary Branch Values", overlay=false)
pick_hl2   = close > open ? hl2   : close
pick_hlc3  = close > open ? close : hlc3
pick_ohlc4 = ohlc4 > hl2  ? ohlc4 : hl2
plot(pick_hl2,   "pick_hl2")
plot(pick_hlc3,  "pick_hlc3")
plot(pick_ohlc4, "pick_ohlc4")
plot(high, "high")
plot(low,  "low")
plot(close, "close")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "builtin-derived-branch-values", pineScript)

	high := exec.ExtractPlotValues(t, output, "high")
	low := exec.ExtractPlotValues(t, output, "low")
	closePx := exec.ExtractPlotValues(t, output, "close")
	pickHl2 := exec.ExtractPlotValues(t, output, "pick_hl2")

	if len(pickHl2) < 10 {
		t.Fatal("Expected at least 10 bars")
	}

	/* pick_hl2 should equal hl2 when close > open, else close.
	 * Either way the value must sit within [low, high].
	 */
	for i := 0; i < minIntBuiltin(len(pickHl2), 20); i++ {
		v := pickHl2[i]
		if v < low[i]-0.001 || v > high[i]+0.001 {
			t.Errorf("pick_hl2[%d] = %f outside [%f, %f]", i, v, low[i], high[i])
		}
	}

	/* pick_hlc3 should be either close or hlc3 — both bounded by [low, high]. */
	pickHlc3 := exec.ExtractPlotValues(t, output, "pick_hlc3")
	for i := 0; i < minIntBuiltin(len(pickHlc3), 20); i++ {
		if pickHlc3[i] < low[i]-0.001 || pickHlc3[i] > high[i]+0.001 {
			t.Errorf("pick_hlc3[%d] = %f outside [%f, %f]", i, pickHlc3[i], low[i], high[i])
		}
	}

	/* pick_ohlc4 is max(ohlc4, hl2) — both averages of OHLC components,
	 * so result must be close to close (sanity: not NaN or wildly off).
	 */
	pickOhlc4 := exec.ExtractPlotValues(t, output, "pick_ohlc4")
	for i := 0; i < minIntBuiltin(len(pickOhlc4), 20); i++ {
		if isNaNBuiltin(pickOhlc4[i]) {
			t.Errorf("pick_ohlc4[%d] is NaN", i)
		}
		if absBuiltin(pickOhlc4[i]-closePx[i]) > closePx[i]*0.1 {
			t.Errorf("pick_ohlc4[%d] = %f deviates >10%% from close %f", i, pickOhlc4[i], closePx[i])
		}
	}
}
