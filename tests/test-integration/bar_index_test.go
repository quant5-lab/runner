package integration

import (
	"testing"

	"github.com/quant5-lab/runner/tests/testutil"
)

func TestBarIndexBasic(t *testing.T) {
	pineScript := `//@version=5
indicator("bar_index Basic", overlay=false)
barIdx = bar_index
plot(barIdx, "Bar Index")
`

	exec := testutil.NewPineExecutor(t)
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

	t.Logf("✅ bar_index sequence validated: 0 to %d", len(barIndexVals)-1)
}

func TestBarIndexModulo(t *testing.T) {
	pineScript := `//@version=5
indicator("bar_index Modulo", overlay=false)
mod5 = bar_index % 5
mod20 = bar_index % 20
plot(mod5, "Mod 5")
plot(mod20, "Mod 20")
`

	exec := testutil.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "bar-index-modulo", pineScript)

	mod5 := exec.ExtractPlotValues(t, output, "Mod 5")
	mod20 := exec.ExtractPlotValues(t, output, "Mod 20")

	if len(mod5) < 4 {
		t.Fatal("Not enough data points for mod 5 validation")
	}

	if mod5[0] != 0 {
		t.Error("Mod 5 pattern incorrect at bar 0")
	}

	if len(mod5) > 5 && mod5[5] != 0 {
		t.Error("Mod 5 pattern incorrect at bar 5")
	}

	if len(mod5) > 10 && mod5[10] != 0 {
		t.Error("Mod 5 pattern incorrect at bar 10")
	}

	if mod5[3] != 3 {
		t.Error("Mod 5 pattern incorrect at offset 3")
	}

	if len(mod5) > 8 && mod5[8] != 3 {
		t.Error("Mod 5 pattern incorrect at bar 8")
	}

	if len(mod20) < 21 && mod20[0] != 0 {
		t.Error("Mod 20 should be 0 at bar 0")
	}

	if len(mod20) > 20 && mod20[20] != 0 {
		t.Error("Mod 20 pattern incorrect at bar 20")
	}

	if len(mod20) > 40 && mod20[40] != 0 {
		t.Error("Mod 20 pattern incorrect at bar 40")
	}

	t.Log("✅ bar_index modulo operations validated")
}

func TestBarIndexSecurity(t *testing.T) {
	t.Skip("Security function not implemented - see e2e/fixtures/strategies/test-bar-index-security.pine.skip")
}

func TestBarIndexConditional(t *testing.T) {
	pineScript := `//@version=5
indicator("bar_index Conditional", overlay=false)
firstBar = bar_index == 0 ? 1 : 0
every10th = (bar_index % 10) == 0 ? 1 : 0
plot(firstBar, "First Bar")
plot(every10th, "Every 10th")
`

	exec := testutil.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "bar-index-conditional", pineScript)

	firstBar := exec.ExtractPlotValues(t, output, "First Bar")
	every10th := exec.ExtractPlotValues(t, output, "Every 10th")

	if firstBar[0] != 1 {
		t.Error("First bar flag should be 1 at bar 0")
	}
	if len(firstBar) > 1 && firstBar[1] != 0 {
		t.Error("First bar flag should be 0 after bar 0")
	}

	if len(every10th) > 20 {
		if every10th[0] != 1 || every10th[10] != 1 || every10th[20] != 1 {
			t.Error("Every 10th bar flag incorrect")
		}
	}

	t.Log("✅ bar_index conditional logic validated")
}

func TestBarIndexComparisons(t *testing.T) {
	pineScript := `//@version=5
indicator("bar_index Comparisons", overlay=false)
gtTen = bar_index > 10 ? 1 : 0
eqTwenty = bar_index == 20 ? 1 : 0
plot(gtTen, "Greater Than 10")
plot(eqTwenty, "Equals 20")
`

	exec := testutil.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "bar-index-comparisons", pineScript)

	gtTen := exec.ExtractPlotValues(t, output, "Greater Than 10")
	eqTwenty := exec.ExtractPlotValues(t, output, "Equals 20")

	if len(gtTen) > 11 {
		if gtTen[10] != 0 || gtTen[11] != 1 {
			t.Error("Greater than 10 comparison incorrect")
		}
	}

	/* == 20 should be true only at bar 20 */
	if len(eqTwenty) > 21 {
		if eqTwenty[19] != 0 || eqTwenty[20] != 1 || eqTwenty[21] != 0 {
			t.Error("Equals 20 comparison incorrect")
		}
	}

	t.Log("✅ bar_index comparisons validated")
}

func TestBarIndexHistorical(t *testing.T) {
	t.Skip("Requires bar_index historical access codegen - see e2e/fixtures/strategies/test-bar-index-historical.pine.skip")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
