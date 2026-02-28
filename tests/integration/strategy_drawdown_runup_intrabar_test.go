//go:build integration

package integration

import (
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestStrategyDrawdownRunup_IntrabarLongCapture(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
strategy("DD/RU Intrabar Long", overlay=true, pyramiding=1, initial_capital=10000)

var bool entered = false

if bar_index == 0 and not entered
    strategy.entry("Long", strategy.long, qty=10.0)
    entered := true

plot(strategy.max_drawdown, "DD")
plot(strategy.max_runup, "RU")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "dd-ru-intrabar-long", pineScript)

	dd := exec.ExtractPlotValues(t, output, "DD")
	ru := exec.ExtractPlotValues(t, output, "RU")

	if len(dd) < 5 {
		t.Fatal("Expected at least 5 bars")
	}

	finalDD := dd[len(dd)-1]
	finalRU := ru[len(ru)-1]

	if finalDD <= 0 {
		t.Errorf("max_drawdown must be positive after volatile bars, got %.2f", finalDD)
	}

	if finalRU <= 0 {
		t.Errorf("max_runup must be positive after volatile bars, got %.2f", finalRU)
	}

	for i := 1; i < len(dd); i++ {
		if dd[i] < dd[i-1] {
			t.Errorf("Bar %d: max_drawdown decreased from %.2f to %.2f (monotonicity violated)", i, dd[i-1], dd[i])
		}
	}

	for i := 1; i < len(ru); i++ {
		if ru[i] < ru[i-1] {
			t.Errorf("Bar %d: max_runup decreased from %.2f to %.2f (monotonicity violated)", i, ru[i-1], ru[i])
		}
	}

	t.Logf("✅ Intrabar long: DD=%.2f RU=%.2f (monotonic non-decreasing)", finalDD, finalRU)
}

func TestStrategyDrawdownRunup_IntrabarShortCapture(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
strategy("DD/RU Intrabar Short", overlay=true, pyramiding=1, initial_capital=10000)

var bool entered = false

if bar_index == 0 and not entered
    strategy.entry("Short", strategy.short, qty=10.0)
    entered := true

plot(strategy.max_drawdown, "DD")
plot(strategy.max_runup, "RU")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "dd-ru-intrabar-short", pineScript)

	dd := exec.ExtractPlotValues(t, output, "DD")
	ru := exec.ExtractPlotValues(t, output, "RU")

	if len(dd) < 5 {
		t.Fatal("Expected at least 5 bars")
	}

	finalDD := dd[len(dd)-1]
	finalRU := ru[len(ru)-1]

	if finalDD <= 0 {
		t.Errorf("max_drawdown must be positive for short after volatile bars, got %.2f", finalDD)
	}

	if finalRU <= 0 {
		t.Errorf("max_runup must be positive for short after volatile bars, got %.2f", finalRU)
	}

	for i := 1; i < len(dd); i++ {
		if dd[i] < dd[i-1] {
			t.Errorf("Bar %d: max_drawdown decreased (monotonicity violated)", i)
		}
	}

	for i := 1; i < len(ru); i++ {
		if ru[i] < ru[i-1] {
			t.Errorf("Bar %d: max_runup decreased (monotonicity violated)", i)
		}
	}

	t.Logf("✅ Intrabar short: DD=%.2f RU=%.2f (monotonic non-decreasing)", finalDD, finalRU)
}

func TestStrategyDrawdownRunup_PercentageConsistency(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
strategy("DD/RU Percentage", overlay=true, pyramiding=1, initial_capital=10000)

var bool entered = false

if bar_index == 0 and not entered
    strategy.entry("Long", strategy.long, qty=10.0)
    entered := true

plot(strategy.max_drawdown, "DD")
plot(strategy.max_drawdown_percent, "DD_PCT")
plot(strategy.max_runup, "RU")
plot(strategy.max_runup_percent, "RU_PCT")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "dd-ru-percentage", pineScript)

	dd := exec.ExtractPlotValues(t, output, "DD")
	ddPct := exec.ExtractPlotValues(t, output, "DD_PCT")
	ru := exec.ExtractPlotValues(t, output, "RU")
	ruPct := exec.ExtractPlotValues(t, output, "RU_PCT")

	if len(dd) < 5 {
		t.Fatal("Expected at least 5 bars")
	}

	finalDD := dd[len(dd)-1]
	finalDDPct := ddPct[len(ddPct)-1]

	if finalDD > 0 && finalDDPct <= 0 {
		t.Errorf("Drawdown=%.2f but percent=%.4f (should be positive)", finalDD, finalDDPct)
	}

	finalRU := ru[len(ru)-1]
	finalRUPct := ruPct[len(ruPct)-1]

	if finalRU > 0 && finalRUPct <= 0 {
		t.Errorf("Runup=%.2f but percent=%.4f (should be positive)", finalRU, finalRUPct)
	}

	t.Logf("✅ Percentage consistency: DD=%.2f (%.2f%%), RU=%.2f (%.2f%%)",
		finalDD, finalDDPct, finalRU, finalRUPct)
}

func TestStrategyDrawdownRunup_FlatPositionZeroMetrics(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
strategy("DD/RU Flat", overlay=true, pyramiding=1, initial_capital=10000)

plot(strategy.max_drawdown, "DD")
plot(strategy.max_runup, "RU")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "dd-ru-flat", pineScript)

	dd := exec.ExtractPlotValues(t, output, "DD")
	ru := exec.ExtractPlotValues(t, output, "RU")

	if len(dd) < 5 {
		t.Fatal("Expected at least 5 bars")
	}

	for i, val := range dd {
		if val != 0 {
			t.Errorf("Bar %d: flat position should have zero drawdown, got %.2f", i, val)
		}
	}

	for i, val := range ru {
		if val != 0 {
			t.Errorf("Bar %d: flat position should have zero runup, got %.2f", i, val)
		}
	}

	t.Logf("✅ Flat position: all DD and RU values are zero")
}
