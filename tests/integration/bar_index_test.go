//go:build integration

package integration

import (
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestBarIndexBasic(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("bar_index Basic", overlay=false)
barIdx = bar_index
plot(barIdx, "Bar Index")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "bar-index-basic", pineScript)

	barIndexVals := exec.ExtractPlotValues(t, output, "Bar Index")

	if len(barIndexVals) < 10 {
		t.Fatal("Expected at least 10 bars")
	}

	if barIndexVals[0] != 0 {
		t.Errorf("bar_index[0] = %f, want 0", barIndexVals[0])
	}

	if barIndexVals[1] != 1 {
		t.Errorf("bar_index[1] = %f, want 1", barIndexVals[1])
	}

	for i := 0; i < minInt(len(barIndexVals), 100); i++ {
		if barIndexVals[i] != float64(i) {
			t.Errorf("bar_index[%d] = %f, want %d", i, barIndexVals[i], i)
		}
	}

}

func TestBarIndexModulo(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("bar_index Modulo", overlay=false)
mod5    = bar_index % 5
mod7    = bar_index % 7
every5  = (bar_index % 5) == 0 ? 1 : 0
plot(mod5,   "Mod 5")
plot(mod7,   "Mod 7")
plot(every5, "Every 5")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "bar-index-modulo", pineScript)

	mod5 := exec.ExtractPlotValues(t, output, "Mod 5")
	mod7 := exec.ExtractPlotValues(t, output, "Mod 7")
	every5 := exec.ExtractPlotValues(t, output, "Every 5")

	if len(mod5) < 30 {
		t.Fatal("Expected at least 30 bars for modulo validation")
	}

	for i := 0; i < minInt(len(mod5), 50); i++ {
		if mod5[i] != float64(i%5) {
			t.Errorf("bar %d: bar_index %% 5 = %f, want %f", i, mod5[i], float64(i%5))
		}
	}

	for i := 0; i < minInt(len(mod7), 50); i++ {
		if mod7[i] != float64(i%7) {
			t.Errorf("bar %d: bar_index %% 7 = %f, want %f", i, mod7[i], float64(i%7))
		}
	}

	for i := 0; i < minInt(len(every5), 50); i++ {
		want := float64(0)
		if i%5 == 0 {
			want = 1.0
		}
		if every5[i] != want {
			t.Errorf("bar %d: (bar_index %% 5) == 0 = %f, want %f", i, every5[i], want)
		}
	}

}

func TestBarIndexSecurity(t *testing.T) {
	t.Parallel()
	t.Skip("Security function not implemented - see e2e/fixtures/strategies/test-bar-index-security.pine.skip")
}

func TestBarIndexConditional(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("bar_index Conditional", overlay=false)
firstBar = bar_index == 0 ? 1 : 0
block = 0.0
if bar_index < 10
    block := 1.0
else if bar_index < 20
    block := 2.0
else
    block := 3.0
plot(firstBar, "First Bar")
plot(block,    "Block")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "bar-index-conditional", pineScript)

	firstBar := exec.ExtractPlotValues(t, output, "First Bar")
	block := exec.ExtractPlotValues(t, output, "Block")

	if len(block) < 25 {
		t.Fatal("Expected at least 25 bars for conditional validation")
	}

	if firstBar[0] != 1 {
		t.Errorf("bar 0: firstBar = %f, want 1 (bar_index == 0)", firstBar[0])
	}
	for i := 1; i < minInt(len(firstBar), 20); i++ {
		if firstBar[i] != 0 {
			t.Errorf("bar %d: firstBar = %f, want 0 (bar_index != 0)", i, firstBar[i])
		}
	}

	for i := 0; i < minInt(len(block), 50); i++ {
		var want float64
		switch {
		case i < 10:
			want = 1.0
		case i < 20:
			want = 2.0
		default:
			want = 3.0
		}
		if block[i] != want {
			t.Errorf("bar %d: block = %f, want %f", i, block[i], want)
		}
	}

}

func TestBarIndexComparisons(t *testing.T) {
	t.Parallel()
	// All 6 comparison operators tested at the same boundary (10) so a single
	// property-based loop can verify correct semantics for every bar.
	pineScript := `//@version=5
indicator("bar_index Comparisons", overlay=false)
gt  = bar_index > 10  ? 1 : 0
gte = bar_index >= 10 ? 1 : 0
lt  = bar_index < 10  ? 1 : 0
lte = bar_index <= 10 ? 1 : 0
eq  = bar_index == 10 ? 1 : 0
ne  = bar_index != 10 ? 1 : 0
plot(gt,  "GT")
plot(gte, "GTE")
plot(lt,  "LT")
plot(lte, "LTE")
plot(eq,  "EQ")
plot(ne,  "NE")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "bar-index-comparisons", pineScript)

	gt := exec.ExtractPlotValues(t, output, "GT")
	gte := exec.ExtractPlotValues(t, output, "GTE")
	lt := exec.ExtractPlotValues(t, output, "LT")
	lte := exec.ExtractPlotValues(t, output, "LTE")
	eq := exec.ExtractPlotValues(t, output, "EQ")
	ne := exec.ExtractPlotValues(t, output, "NE")

	const boundary = 10
	n := minInt(len(gt), 30)
	if n < boundary+5 {
		t.Fatal("Expected at least boundary+5 bars for comparison validation")
	}

	for i := 0; i < n; i++ {
		wantGT := boolToFloat(i > boundary)
		wantGTE := boolToFloat(i >= boundary)
		wantLT := boolToFloat(i < boundary)
		wantLTE := boolToFloat(i <= boundary)
		wantEQ := boolToFloat(i == boundary)
		wantNE := boolToFloat(i != boundary)

		if gt[i] != wantGT {
			t.Errorf("bar %d: GT = %f, want %f", i, gt[i], wantGT)
		}
		if gte[i] != wantGTE {
			t.Errorf("bar %d: GTE = %f, want %f", i, gte[i], wantGTE)
		}
		if lt[i] != wantLT {
			t.Errorf("bar %d: LT = %f, want %f", i, lt[i], wantLT)
		}
		if lte[i] != wantLTE {
			t.Errorf("bar %d: LTE = %f, want %f", i, lte[i], wantLTE)
		}
		if eq[i] != wantEQ {
			t.Errorf("bar %d: EQ = %f, want %f", i, eq[i], wantEQ)
		}
		if ne[i] != wantNE {
			t.Errorf("bar %d: NE = %f, want %f", i, ne[i], wantNE)
		}
	}

}

func TestBarIndexHistorical(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("bar_index Historical", overlay=false)
prev1 = bar_index[1]
prev2 = bar_index[2]
diff  = bar_index - nz(bar_index[1])
plot(prev1, "Prev 1")
plot(prev2, "Prev 2")
plot(diff,  "Diff")
`

	exec := util.NewPineExecutor(t)
	out := exec.ExecuteScript(t, "bar-index-historical", pineScript)

	prev1 := exec.ExtractPlotValues(t, out, "Prev 1")
	prev2 := exec.ExtractPlotValues(t, out, "Prev 2")
	diff := exec.ExtractPlotValues(t, out, "Diff")

	if len(prev1) < 10 {
		t.Fatal("Expected at least 10 bars")
	}

	for i := 1; i < minInt(len(prev1), 20); i++ {
		if prev1[i] != float64(i-1) {
			t.Errorf("bar %d: bar_index[1] = %f, want %f", i, prev1[i], float64(i-1))
		}
	}

	for i := 2; i < minInt(len(prev2), 20); i++ {
		if prev2[i] != float64(i-2) {
			t.Errorf("bar %d: bar_index[2] = %f, want %f", i, prev2[i], float64(i-2))
		}
	}

	if diff[0] != 0 {
		t.Errorf("bar 0: diff = %f, want 0 (nz absorbs pre-history NaN)", diff[0])
	}
	for i := 1; i < minInt(len(diff), 20); i++ {
		if diff[i] != 1 {
			t.Errorf("bar %d: diff = %f, want 1", i, diff[i])
		}
	}

}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
