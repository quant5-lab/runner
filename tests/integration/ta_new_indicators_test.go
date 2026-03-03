//go:build integration

package integration

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

// kcwWarmup is the number of leading NaN bars for ta.kcw with a constant period.
// Builder emits `ctx.BarIndex < period-1`, so bars [0, period-2] are NaN.
func kcwWarmup(period int) int { return period - 1 }

func TestKCWWarmupAndPositivity(t *testing.T) {
	t.Parallel()

	script := loadFixture(t, "integration/test-kcw-warmup.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "kcw-warmup", script)

	width := exec.ExtractPlotValues(t, output, "KCW Width")

	warmup := kcwWarmup(20)
	if len(width) <= warmup {
		t.Fatalf("need > %d bars, got %d", warmup, len(width))
	}

	for i := 0; i < warmup; i++ {
		if !math.IsNaN(width[i]) {
			t.Errorf("bar[%d] = %f, want NaN during warmup", i, width[i])
		}
	}

	for i := warmup; i < minIntBBW(len(width), warmup+50); i++ {
		if math.IsNaN(width[i]) {
			t.Errorf("bar[%d] is NaN after warmup", i)
			continue
		}
		if width[i] <= 0 {
			t.Errorf("bar[%d] = %f, KCW must be positive (ATR/EMA ratio)", i, width[i])
		}
	}
}

func TestKCWMultScaling(t *testing.T) {
	t.Parallel()

	script := loadFixture(t, "integration/test-kcw-mult-scaling.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "kcw-mult-scaling", script)

	unit := exec.ExtractPlotValues(t, output, "Unit 1.0x")
	double := exec.ExtractPlotValues(t, output, "Double 2.0x")

	if len(unit) != len(double) {
		t.Fatal("unit and double series must have same length")
	}

	warmup := kcwWarmup(20)
	if len(unit) <= warmup {
		t.Fatalf("need > %d bars, got %d", warmup, len(unit))
	}

	for i := warmup; i < minIntBBW(len(unit), warmup+100); i++ {
		if math.IsNaN(unit[i]) || math.IsNaN(double[i]) {
			continue
		}
		expected := 2.0 * unit[i]
		if math.Abs(double[i]-expected) > 1e-10 {
			t.Errorf("bar[%d]: double=%f, 2*unit=%f — mult must scale linearly", i, double[i], expected)
		}
	}
}

func TestKCWIdentifierMult(t *testing.T) {
	t.Parallel()

	script := loadFixture(t, "integration/test-kcw-identifier-mult.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "kcw-identifier-mult", script)

	withVar := exec.ExtractPlotValues(t, output, "Variable Mult")
	withLit := exec.ExtractPlotValues(t, output, "Literal Mult")
	diff := exec.ExtractPlotValues(t, output, "Difference")

	if len(withVar) != len(withLit) {
		t.Fatal("variable-mult and literal-mult series must have same length")
	}

	warmup := kcwWarmup(20)
	if len(withVar) <= warmup {
		t.Fatalf("need > %d bars, got %d", warmup, len(withVar))
	}

	for i := warmup; i < minIntBBW(len(diff), warmup+100); i++ {
		if math.IsNaN(withVar[i]) || math.IsNaN(withLit[i]) {
			continue
		}
		if math.Abs(diff[i]) > 1e-10 {
			t.Errorf("bar[%d]: variable=%f, literal=%f, diff=%f — identifier mult must render identically to literal",
				i, withVar[i], withLit[i], diff[i])
		}
	}
}

func TestKCWArrowEquivalence(t *testing.T) {
	t.Parallel()

	script := loadFixture(t, "integration/test-kcw-arrow-mixed.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "kcw-arrow-mixed", script)

	direct := exec.ExtractPlotValues(t, output, "Direct")
	arrow := exec.ExtractPlotValues(t, output, "Arrow")
	diff := exec.ExtractPlotValues(t, output, "Difference")

	if len(direct) != len(arrow) {
		t.Fatal("direct and arrow series must have same length")
	}

	warmup := kcwWarmup(20)
	if len(direct) <= warmup {
		t.Fatalf("need > %d bars, got %d", warmup, len(direct))
	}

	for i := warmup; i < minIntBBW(len(diff), warmup+100); i++ {
		if math.IsNaN(direct[i]) || math.IsNaN(arrow[i]) {
			continue
		}
		if math.Abs(diff[i]) > 1e-10 {
			t.Errorf("bar[%d]: direct=%f, arrow=%f, diff=%f — arrow must produce identical result to direct call",
				i, direct[i], arrow[i], diff[i])
		}
	}
}

func TestALMAWarmupAndFiniteness(t *testing.T) {
	t.Parallel()

	script := loadFixture(t, "integration/test-alma-basic.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "alma-basic", script)

	vals := exec.ExtractPlotValues(t, output, "ALMA")

	// warmup = period (handler: ctx.BarIndex < period)
	const warmup = 20
	if len(vals) <= warmup {
		t.Fatalf("need > %d bars, got %d", warmup, len(vals))
	}

	for i := 0; i < warmup; i++ {
		if !math.IsNaN(vals[i]) {
			t.Errorf("bar[%d] = %f, want NaN during warmup", i, vals[i])
		}
	}

	for i := warmup; i < minIntBBW(len(vals), warmup+50); i++ {
		if math.IsNaN(vals[i]) {
			t.Errorf("bar[%d] is NaN after warmup", i)
		}
	}
}

func TestHMAWarmupAndFiniteness(t *testing.T) {
	t.Parallel()

	script := loadFixture(t, "integration/test-hma-basic.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "hma-basic", script)

	vals := exec.ExtractPlotValues(t, output, "HMA")

	// period=16, sqrtPeriod=round(sqrt(16))=4 → totalWarmup = 16 + 4 - 1 = 19
	const period, sqrtPeriod = 16, 4
	warmup := period + sqrtPeriod - 1
	if len(vals) <= warmup {
		t.Fatalf("need > %d bars, got %d", warmup, len(vals))
	}

	for i := 0; i < warmup; i++ {
		if !math.IsNaN(vals[i]) {
			t.Errorf("bar[%d] = %f, want NaN during warmup", i, vals[i])
		}
	}

	for i := warmup; i < minIntBBW(len(vals), warmup+50); i++ {
		if math.IsNaN(vals[i]) {
			t.Errorf("bar[%d] is NaN after warmup", i)
		}
	}
}

func TestSARWarmupAndFiniteness(t *testing.T) {
	t.Parallel()

	script := loadFixture(t, "integration/test-sar-basic.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "sar-basic", script)

	vals := exec.ExtractPlotValues(t, output, "SAR")

	if len(vals) < 10 {
		t.Fatalf("need at least 10 bars, got %d", len(vals))
	}

	if !math.IsNaN(vals[0]) {
		t.Errorf("bar[0] = %f, want NaN (SAR undefined on first bar)", vals[0])
	}

	for i := 1; i < minIntBBW(len(vals), 50); i++ {
		if math.IsNaN(vals[i]) {
			t.Errorf("bar[%d] is NaN — SAR must be defined from bar 1 onward", i)
			continue
		}
		if vals[i] <= 0 {
			t.Errorf("bar[%d] = %f, SAR is a price level and must be positive", i, vals[i])
		}
	}
}

func TestPercentileOrdering(t *testing.T) {
	t.Parallel()

	script := loadFixture(t, "integration/test-percentile-basic.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "percentile-basic", script)

	pnr75 := exec.ExtractPlotValues(t, output, "NR 75th")
	pli75 := exec.ExtractPlotValues(t, output, "LI 75th")
	pnr25 := exec.ExtractPlotValues(t, output, "NR 25th")

	if len(pnr75) != len(pli75) || len(pli75) != len(pnr25) {
		t.Fatal("all percentile plots must have same length")
	}

	// warmup = period (handler: ctx.BarIndex < period)
	const warmup = 20
	if len(pnr75) <= warmup {
		t.Fatalf("need > %d bars, got %d", warmup, len(pnr75))
	}

	for i := 0; i < warmup; i++ {
		if !math.IsNaN(pnr75[i]) || !math.IsNaN(pli75[i]) || !math.IsNaN(pnr25[i]) {
			t.Errorf("bar[%d]: expected all NaN during warmup", i)
		}
	}

	for i := warmup; i < minIntBBW(len(pnr75), warmup+100); i++ {
		if math.IsNaN(pnr75[i]) || math.IsNaN(pli75[i]) || math.IsNaN(pnr25[i]) {
			t.Errorf("bar[%d]: unexpected NaN after warmup", i)
			continue
		}
		if pnr75[i] < pnr25[i] {
			t.Errorf("bar[%d]: 75th percentile=%f < 25th percentile=%f", i, pnr75[i], pnr25[i])
		}
		if pli75[i] < pnr25[i] {
			t.Errorf("bar[%d]: linear-interpolation 75th=%f < nearest-rank 25th=%f", i, pli75[i], pnr25[i])
		}
	}
}

func TestPercentrankRange(t *testing.T) {
	t.Parallel()

	script := loadFixture(t, "integration/test-percentrank-basic.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "percentrank-basic", script)

	rank := exec.ExtractPlotValues(t, output, "PercentRank")

	// warmup = period (handler: ctx.BarIndex < period)
	const warmup = 20
	if len(rank) <= warmup {
		t.Fatalf("need > %d bars, got %d", warmup, len(rank))
	}

	for i := 0; i < warmup; i++ {
		if !math.IsNaN(rank[i]) {
			t.Errorf("bar[%d] = %f, want NaN during warmup", i, rank[i])
		}
	}

	for i := warmup; i < minIntBBW(len(rank), warmup+100); i++ {
		if math.IsNaN(rank[i]) {
			t.Errorf("bar[%d] is NaN after warmup", i)
			continue
		}
		if rank[i] < 0 || rank[i] > 100 {
			t.Errorf("bar[%d] = %f, percentrank must be in [0, 100]", i, rank[i])
		}
	}
}

func TestCorrelationSelf(t *testing.T) {
	t.Parallel()

	script := loadFixture(t, "integration/test-correlation-self.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "correlation-self", script)

	selfCorr := exec.ExtractPlotValues(t, output, "Self Corr")
	closeVolCorr := exec.ExtractPlotValues(t, output, "Close-Vol Corr")

	if len(selfCorr) != len(closeVolCorr) {
		t.Fatal("both correlation series must have same length")
	}

	// warmup = period (handler: ctx.BarIndex < period)
	const warmup = 20
	if len(selfCorr) <= warmup {
		t.Fatalf("need > %d bars, got %d", warmup, len(selfCorr))
	}

	for i := 0; i < warmup; i++ {
		if !math.IsNaN(selfCorr[i]) || !math.IsNaN(closeVolCorr[i]) {
			t.Errorf("bar[%d]: expected NaN during warmup", i)
		}
	}

	for i := warmup; i < minIntBBW(len(selfCorr), warmup+100); i++ {
		if math.IsNaN(selfCorr[i]) {
			t.Errorf("bar[%d]: self-correlation is NaN after warmup", i)
			continue
		}
		if math.Abs(selfCorr[i]-1.0) > 1e-10 {
			t.Errorf("bar[%d]: self-correlation=%f, want 1.0 (Pearson r of series with itself)", i, selfCorr[i])
		}
		if !math.IsNaN(closeVolCorr[i]) && (closeVolCorr[i] < -1.0 || closeVolCorr[i] > 1.0) {
			t.Errorf("bar[%d]: close-volume correlation=%f outside [-1, 1]", i, closeVolCorr[i])
		}
	}
}
