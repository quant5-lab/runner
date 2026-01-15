package integration

import (
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/*
Security() Historical Lookback Integration Tests

PURPOSE: Validate security() variables with historical [1] [2] [3] lookback access

COVERAGE:
  1. Basic [1] access on security() variables
  2. Comparison patterns with [1] (value != value[1])
  3. Multiple offsets [1] [2] [3] simultaneously
  4. valuewhen() with security-derived conditions
  5. Strategy logic with security [1] access

ALIGNMENT: Tests are generalized to validate algorithm behavior, not specific bug cases
           All tests use real SPY_1D.json data with 90% match rate thresholds
*/

// TestSecurityHistoricalLookback_SimplePrevious validates basic [1] access on security variables
func TestSecurityHistoricalLookback_SimplePrevious(t *testing.T) {
	pineScript := `//@version=5
indicator("Simple Previous", overlay=false)

// Daily SMA with [1] access
sma_1d = security(syminfo.tickerid, "1D", ta.sma(close, 3))
prev_sma_1d = sma_1d[1]

plot(sma_1d, "current")
plot(prev_sma_1d, "previous")
`

	executor := util.NewPineExecutor(t)
	output := executor.ExecuteScript(t, "simple-prev", pineScript)

	current := executor.ExtractPlotValues(t, output, "current")
	previous := executor.ExtractPlotValues(t, output, "previous")

	if len(current) < 10 || len(previous) < 10 {
		t.Fatalf("Insufficient data: current=%d, previous=%d bars", len(current), len(previous))
	}

	/* Validate [1] access: previous[i] should equal current[i-1] */
	mismatchCount := 0
	for i := 2; i < len(current) && i < len(previous); i++ {
		if previous[i] != current[i-1] {
			mismatchCount++
		}
	}

	matchRate := float64(len(current)-2-mismatchCount) / float64(len(current)-2)
	if matchRate < 0.9 {
		t.Errorf("[1] access broken: only %.0f%% of previous values match current[i-1]", matchRate*100)
	}
}

// TestSecurityHistoricalLookback_ComparisonPattern validates != comparison with [1]
func TestSecurityHistoricalLookback_ComparisonPattern(t *testing.T) {
	pineScript := `//@version=5
indicator("Comparison Pattern", overlay=false)

// Daily close change detection
daily_close = security(syminfo.tickerid, "1D", close)
changed = daily_close != daily_close[1]

plot(daily_close, "value")
plot(changed ? 1 : 0, "changed")
`

	executor := util.NewPineExecutor(t)
	output := executor.ExecuteScript(t, "comparison", pineScript)

	values := executor.ExtractPlotValues(t, output, "value")
	changed := executor.ExtractPlotValues(t, output, "changed")

	if len(values) < 5 || len(changed) < 5 {
		t.Fatalf("Insufficient data: values=%d, changed=%d bars", len(values), len(changed))
	}

	/* Validate: when value changes, changed should be 1 */
	correctCount := 0
	for i := 1; i < len(values) && i < len(changed); i++ {
		valueChanged := values[i] != values[i-1]
		changedFlag := changed[i] == 1.0
		if valueChanged == changedFlag {
			correctCount++
		}
	}

	matchRate := float64(correctCount) / float64(len(values)-1)
	if matchRate < 0.9 {
		t.Errorf("[1] comparison broken: only %.0f%% of changed flags correct", matchRate*100)
	}
}

// TestSecurityHistoricalLookback_ValuewhenChain validates valuewhen with security [1]
func TestSecurityHistoricalLookback_ValuewhenChain(t *testing.T) {
	pineScript := `//@version=5
indicator("Valuewhen Chain", overlay=false)

// Daily high with valuewhen on [1] condition
daily_high = security(syminfo.tickerid, "1D", high)
condition = daily_high > daily_high[1]
captured = valuewhen(condition, daily_high, 0)

plot(daily_high, "high")
plot(condition ? 1 : 0, "condition")
plot(captured, "captured")
`

	executor := util.NewPineExecutor(t)
	output := executor.ExecuteScript(t, "valuewhen-chain", pineScript)

	high := executor.ExtractPlotValues(t, output, "high")
	captured := executor.ExtractPlotValues(t, output, "captured")

	if len(high) < 5 || len(captured) < 5 {
		t.Fatalf("Insufficient data: high=%d, captured=%d bars", len(high), len(captured))
	}

	/* Valuewhen should capture values when condition is true */
	hasNonZeroCaptured := false
	for _, v := range captured {
		if v > 0 {
			hasNonZeroCaptured = true
			break
		}
	}

	if !hasNonZeroCaptured {
		t.Error("Valuewhen failed to capture values")
	}
}

// TestSecurityHistoricalLookback_MultipleOffsets validates [1] [2] [3] offsets simultaneously
func TestSecurityHistoricalLookback_MultipleOffsets(t *testing.T) {
	pineScript := `//@version=5
indicator("Multiple Offsets", overlay=false)

daily_close = security(syminfo.tickerid, "1D", close)
prev1 = daily_close[1]
prev2 = daily_close[2]
prev3 = daily_close[3]

plot(daily_close, "current")
plot(prev1, "prev1")
plot(prev2, "prev2")
plot(prev3, "prev3")
`

	executor := util.NewPineExecutor(t)
	output := executor.ExecuteScript(t, "multiple-offsets", pineScript)

	current := executor.ExtractPlotValues(t, output, "current")
	prev1 := executor.ExtractPlotValues(t, output, "prev1")
	prev2 := executor.ExtractPlotValues(t, output, "prev2")
	prev3 := executor.ExtractPlotValues(t, output, "prev3")

	if len(current) < 10 {
		t.Fatalf("Insufficient data: %d bars", len(current))
	}

	/* Validate offset chain: prev1[i] = current[i-1], prev2[i] = current[i-2], etc */
	correctPrev1, correctPrev2, correctPrev3 := 0, 0, 0
	total := 0
	for i := 4; i < len(current); i++ {
		total++
		if prev1[i] == current[i-1] {
			correctPrev1++
		}
		if prev2[i] == current[i-2] {
			correctPrev2++
		}
		if prev3[i] == current[i-3] {
			correctPrev3++
		}
	}

	if float64(correctPrev1)/float64(total) < 0.9 {
		t.Errorf("[1] offset broken: %.0f%% match", float64(correctPrev1)/float64(total)*100)
	}
	if float64(correctPrev2)/float64(total) < 0.9 {
		t.Errorf("[2] offset broken: %.0f%% match", float64(correctPrev2)/float64(total)*100)
	}
	if float64(correctPrev3)/float64(total) < 0.9 {
		t.Errorf("[3] offset broken: %.0f%% match", float64(correctPrev3)/float64(total)*100)
	}
}

// TestSecurityHistoricalLookback_WithStrategyLogic validates strategy with security [1]
func TestSecurityHistoricalLookback_WithStrategyLogic(t *testing.T) {
	pineScript := `//@version=5
strategy("Security Strategy", overlay=false)

// Daily trend detection with [1] comparison
daily_close = security(syminfo.tickerid, "1D", close)
daily_sma = security(syminfo.tickerid, "1D", ta.sma(close, 5))
daily_trend = daily_close > daily_sma ? 1 : 0
trend_changed = daily_trend != daily_trend[1]

// Entry/exit on trend changes
if trend_changed and daily_trend == 1
    strategy.entry("Long", strategy.long)

if trend_changed and daily_trend == 0
    strategy.close("Long")

plot(daily_trend, "trend")
plot(trend_changed ? 1 : 0, "changed")
`

	executor := util.NewPineExecutor(t)
	output := executor.ExecuteScript(t, "strategy-security", pineScript)

	trend := executor.ExtractPlotValues(t, output, "trend")
	changed := executor.ExtractPlotValues(t, output, "changed")

	if len(trend) < 5 || len(changed) < 5 {
		t.Fatalf("Insufficient data: trend=%d, changed=%d bars", len(trend), len(changed))
	}

	/* Validate trend_changed correctly detects transitions */
	correctChanges := 0
	totalChanges := 0
	for i := 1; i < len(trend) && i < len(changed); i++ {
		actualChange := trend[i] != trend[i-1]
		flaggedChange := changed[i] == 1.0
		if actualChange {
			totalChanges++
			if flaggedChange {
				correctChanges++
			}
		}
	}

	if totalChanges > 0 && float64(correctChanges)/float64(totalChanges) < 0.8 {
		t.Errorf("Strategy [1] comparison broken: only %d/%d trend changes detected", correctChanges, totalChanges)
	}
}
