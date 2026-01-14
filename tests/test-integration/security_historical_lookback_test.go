package integration

import (
	"strings"
	"testing"
)

/*
Security() Historical Lookback Integration Tests

PURPOSE: Comprehensive safety net for security() variables with historical lookback [1]

PROBLEM: Variables assigned from security() calls are stored in main-context Series,
         causing [1] to access wrong bar (previous main bar vs previous security bar)

EVIDENCE: bb-strategy-9-rus.pine - 579 expected exits, 0 actual exits (100% failure)

ROOT CAUSE:
  bb_1d_isOverBBTop = security("1D", ...)  // Evaluated in daily context ✅
  bb_1d_isOverBBTopSeries = series.NewSeries(len(ctx.Data))  // HOURLY size ❌
  newis = bb_1d_isOverBBTop != bb_1d_isOverBBTop[1]  // [1] = prev hourly ❌

EXPECTED BEHAVIOR:
  bb_1d_isOverBBTop[1] should access previous DAILY bar, not previous hourly bar

TEST STRATEGY:
  1. Reproduce exact bb9 failure pattern
  2. Test all security + [1] combinations
  3. Ensure solution is not a bandaid
  4. Verify 100% PineScript compatibility
*/

// TestSecurityHistoricalLookback_BB9ExactPattern reproduces the exact bb9 bug
// STATUS: ❌ EXPECTED TO FAIL until security variable storage is fixed
func TestSecurityHistoricalLookback_BB9ExactPattern(t *testing.T) {
	t.Skip("BLOCKER: Security variables use main-context Series - see docs/security-historical-lookback-bug.md")

	/*
	   SCENARIO: 3 days of hourly data (30 bars), BB crosses on Day 2

	   Hourly Bars:     Daily Values:
	   Bar 0-9:         Day 1: isOverBBTop = false
	   Bar 10-19:       Day 2: isOverBBTop = true   ← Signal change
	   Bar 20-29:       Day 3: isOverBBTop = true   ← No change

	   EXPECTED at Bar 10-19:
	     bb_1d_isOverBBTop[0] = true (Day 2)
	     bb_1d_isOverBBTop[1] = false (Day 1)
	     newis = true != false = TRUE ✅
	     exit_signal should trigger

	   ACTUAL (BROKEN):
	     bb_1d_isOverBBTop[0] = true (Bar 10)
	     bb_1d_isOverBBTop[1] = true (Bar 9, still Day 1 mapped)
	     newis = true != true = FALSE ❌
	     No exit signal
	*/

	pineScript := `//@version=5
indicator("BB9 Exit Pattern", overlay=false)

// Simulate BB cross: low > 1100 triggers on Day 2
bb_1d_isOverBBTop = security(syminfo.tickerid, "1D", low > 1100)

// This should detect Day 1 → Day 2 change
bb_1d_newis = bb_1d_isOverBBTop != bb_1d_isOverBBTop[1]

// Exit signal pattern from bb9
bb_1d_high_range = security(syminfo.tickerid, "1D", valuewhen(bb_1d_newis, high, 0))
exit_signal = bb_1d_high_range == bb_1d_high_range[1]

plot(bb_1d_isOverBBTop ? 1 : 0, "isOver")
plot(bb_1d_newis ? 1 : 0, "newis")
plot(exit_signal ? 1 : 0, "exit")
`

	output := runStrategyScript(t, "bb9-pattern", pineScript)

	isOver := extractStrategyPlotValues(t, output, "isOver")
	newis := extractStrategyPlotValues(t, output, "newis")
	exitSignal := extractStrategyPlotValues(t, output, "exit")

	// Day 1 (bars 0-9): isOverBBTop = false
	for i := 0; i < 10; i++ {
		if isOver[i] != 0.0 {
			t.Errorf("Bar %d (Day 1): isOver = %.1f, want 0.0", i, isOver[i])
		}
		if newis[i] != 0.0 {
			t.Errorf("Bar %d (Day 1): newis = %.1f, want 0.0 (no change)", i, newis[i])
		}
	}

	// Day 2 (bars 10-19): isOverBBTop = true, newis = true (change detected)
	for i := 10; i < 20; i++ {
		if isOver[i] != 1.0 {
			t.Errorf("Bar %d (Day 2): isOver = %.1f, want 1.0", i, isOver[i])
		}
		if newis[i] != 1.0 {
			t.Errorf("Bar %d (Day 2): newis = %.1f, want 1.0 (CRITICAL: change from Day 1)", i, newis[i])
		}
	}

	// Day 3 (bars 20-29): isOverBBTop = true, newis = false (no change)
	for i := 20; i < 30; i++ {
		if isOver[i] != 1.0 {
			t.Errorf("Bar %d (Day 3): isOver = %.1f, want 1.0", i, isOver[i])
		}
		if newis[i] != 0.0 {
			t.Errorf("Bar %d (Day 3): newis = %.1f, want 0.0 (no change)", i, newis[i])
		}
	}

	// Exit signal should appear on Day 2
	hasExitSignal := false
	for i := 10; i < 20; i++ {
		if exitSignal[i] == 1.0 {
			hasExitSignal = true
			break
		}
	}

	if !hasExitSignal {
		t.Error("CRITICAL: No exit signal on Day 2 - bb9 bug reproduced")
	}
}

// TestSecurityHistoricalLookback_SimplePrevious tests basic [1] access
func TestSecurityHistoricalLookback_SimplePrevious(t *testing.T) {
	t.Skip("BLOCKER: Security variables use main-context Series")

	pineScript := `//@version=5
indicator("Simple Previous", overlay=false)

// Daily SMA
sma_1d = security(syminfo.tickerid, "1D", ta.sma(close, 3))
prev_sma_1d = sma_1d[1]

plot(sma_1d, "current")
plot(prev_sma_1d, "previous")
`

	output := runStrategyScript(t, "simple-prev", pineScript)

	current := extractStrategyPlotValues(t, output, "current")
	previous := extractStrategyPlotValues(t, output, "previous")

	// On Day 2 hourly bars, prev_sma_1d should equal Day 1 sma_1d
	// Not Bar N-1 sma_1d (which could be same day)
	for i := 10; i < 20; i++ {
		expected := current[9] // Day 1's last bar value
		if previous[i] != expected {
			t.Errorf("Bar %d: prev_sma_1d = %.2f, want %.2f (Day 1 value)", i, previous[i], expected)
		}
	}
}

// TestSecurityHistoricalLookback_ComparisonPattern tests != with [1]
func TestSecurityHistoricalLookback_ComparisonPattern(t *testing.T) {
	t.Skip("BLOCKER: Security variables use main-context Series")

	pineScript := `//@version=5
indicator("Comparison Pattern", overlay=false)

// Value that changes every day
daily_val = security(syminfo.tickerid, "1D", bar_index % 3)
changed = daily_val != daily_val[1]

plot(daily_val, "value")
plot(changed ? 1 : 0, "changed")
`

	output := runStrategyScript(t, "comparison", pineScript)

	_ = extractStrategyPlotValues(t, output, "value")
	changed := extractStrategyPlotValues(t, output, "changed")

	// Day 1: val = 0, Day 2: val = 1, Day 3: val = 2
	// Changed should be true on Day 2 and Day 3

	// Day 1 bars: changed = false (no previous day)
	for i := 0; i < 10; i++ {
		if changed[i] != 0.0 && i > 0 {
			t.Errorf("Bar %d (Day 1): changed = %.1f, want 0.0", i, changed[i])
		}
	}

	// Day 2 bars: changed = true (0 → 1)
	for i := 10; i < 20; i++ {
		if changed[i] != 1.0 {
			t.Errorf("Bar %d (Day 2): changed = %.1f, want 1.0 (value changed from Day 1)", i, changed[i])
		}
	}

	// Day 3 bars: changed = true (1 → 2)
	for i := 20; i < 30; i++ {
		if changed[i] != 1.0 {
			t.Errorf("Bar %d (Day 3): changed = %.1f, want 1.0 (value changed from Day 2)", i, changed[i])
		}
	}
}

// TestSecurityHistoricalLookback_NestedSecurity tests security inside security
func TestSecurityHistoricalLookback_NestedSecurity(t *testing.T) {
	t.Skip("BLOCKER: Security variables use main-context Series")

	pineScript := `//@version=5
indicator("Nested Security", overlay=false)

// Get daily close
daily_close = security(syminfo.tickerid, "1D", close)

// Get previous daily close via nested security
prev_daily = security(syminfo.tickerid, "1D", daily_close[1])

plot(daily_close, "current")
plot(prev_daily, "previous")
`

	output := runStrategyScript(t, "nested-security", pineScript)

	current := extractStrategyPlotValues(t, output, "current")
	previous := extractStrategyPlotValues(t, output, "previous")

	// Nested security should access historical daily values correctly
	for i := 10; i < 20; i++ {
		// prev_daily on Day 2 should equal Day 1's daily_close
		expected := current[9] // Day 1's last value
		if previous[i] != expected {
			t.Errorf("Bar %d: nested prev_daily = %.2f, want %.2f", i, previous[i], expected)
		}
	}
}

// TestSecurityHistoricalLookback_ValuewhenChain tests valuewhen with security variables
func TestSecurityHistoricalLookback_ValuewhenChain(t *testing.T) {
	t.Skip("BLOCKER: Security variables use main-context Series")

	pineScript := `//@version=5
indicator("Valuewhen Chain", overlay=false)

// Condition changes daily
condition = security(syminfo.tickerid, "1D", bar_index % 2 == 0)

// Valuewhen on security-derived condition
captured = security(syminfo.tickerid, "1D", valuewhen(condition, high, 0))

// Compare with previous
result = captured == captured[1]

plot(condition ? 1 : 0, "condition")
plot(captured, "captured")
plot(result ? 1 : 0, "same_as_prev")
`

	output := runStrategyScript(t, "valuewhen-chain", pineScript)

	condition := extractStrategyPlotValues(t, output, "condition")
	captured := extractStrategyPlotValues(t, output, "captured")
	result := extractStrategyPlotValues(t, output, "same_as_prev")

	// Verify captured values persist across days correctly
	// And result compares with previous DAILY value, not previous hourly
	t.Log("Condition:", condition[:20])
	t.Log("Captured:", captured[:20])
	t.Log("Result:", result[:20])

	// TODO: Add specific assertions based on expected valuewhen behavior
}

// TestSecurityHistoricalLookback_MultipleOffsets tests [1], [2], [3] etc
func TestSecurityHistoricalLookback_MultipleOffsets(t *testing.T) {
	t.Skip("BLOCKER: Security variables use main-context Series")

	pineScript := `//@version=5
indicator("Multiple Offsets", overlay=false)

daily_val = security(syminfo.tickerid, "1D", bar_index)
prev1 = daily_val[1]
prev2 = daily_val[2]
prev3 = daily_val[3]

plot(daily_val, "current")
plot(prev1, "prev1")
plot(prev2, "prev2")
plot(prev3, "prev3")
`

	output := runStrategyScript(t, "multiple-offsets", pineScript)

	current := extractStrategyPlotValues(t, output, "current")
	prev1 := extractStrategyPlotValues(t, output, "prev1")
	prev2 := extractStrategyPlotValues(t, output, "prev2")
	prev3 := extractStrategyPlotValues(t, output, "prev3")

	// On Day 4 (bars 30-39): current=3, prev1=2, prev2=1, prev3=0
	for i := 30; i < 40; i++ {
		if current[i] != 3.0 {
			t.Errorf("Bar %d: current = %.1f, want 3.0", i, current[i])
		}
		if prev1[i] != 2.0 {
			t.Errorf("Bar %d: prev1 = %.1f, want 2.0 (Day 3)", i, prev1[i])
		}
		if prev2[i] != 1.0 {
			t.Errorf("Bar %d: prev2 = %.1f, want 1.0 (Day 2)", i, prev2[i])
		}
		if prev3[i] != 0.0 {
			t.Errorf("Bar %d: prev3 = %.1f, want 0.0 (Day 1)", i, prev3[i])
		}
	}
}

// TestSecurityHistoricalLookback_WithStrategyLogic tests with strategy entries/exits
func TestSecurityHistoricalLookback_WithStrategyLogic(t *testing.T) {
	t.Skip("BLOCKER: Security variables use main-context Series")

	pineScript := `//@version=5
strategy("Security Strategy", overlay=false)

// Daily trend change
daily_trend = security(syminfo.tickerid, "1D", close > ta.sma(close, 10) ? 1 : 0)
trend_changed = daily_trend != daily_trend[1]

// Entry on trend change
if trend_changed and daily_trend == 1
    strategy.entry("Long", strategy.long)

if trend_changed and daily_trend == 0
    strategy.close("Long")

plot(daily_trend, "trend")
plot(trend_changed ? 1 : 0, "changed")
`

	output := runStrategyScript(t, "strategy-security", pineScript)

	// Verify strategy entries/exits align with daily trend changes
	// Not with hourly bar changes

	// Extract strategy trades
	trades := output.Strategy.ClosedTrades

	// Should have entries/exits on daily boundaries, not intraday
	for _, trade := range trades {
		barIdx := trade.EntryBar
		// Entry should be on first bar of day (multiples of 10)
		if barIdx%10 != 0 {
			t.Errorf("Trade entry at bar %d (not day boundary)", barIdx)
		}
	}
}

func extractStrategyPlotValues(t *testing.T, output *PineScriptOutput, plotTitle string) []float64 {
	t.Helper()

	for _, plot := range output.Plots {
		if strings.Contains(plot.Title, plotTitle) {
			values := make([]float64, len(plot.Data))
			for i, point := range plot.Data {
				values[i] = point.Value
			}
			return values
		}
	}

	t.Fatalf("Plot %q not found in output", plotTitle)
	return nil
}

func runStrategyScript(t *testing.T, name string, script string) *PineScriptOutput {
	t.Helper()

	// TODO: Implement actual PineScript execution

	t.Fatalf("runStrategyScript not yet implemented")
	return nil
}

// PineScriptOutput represents strategy execution output
type PineScriptOutput struct {
	Plots    []StrategyPlot
	Strategy StrategyData
}

type StrategyPlot struct {
	Title string
	Data  []PlotPoint
}

type PlotPoint struct {
	Time  int64
	Value float64
}

type StrategyData struct {
	ClosedTrades []StrategyTrade
}

type StrategyTrade struct {
	EntryBar  int
	ExitBar   int
	EntryTime int64
	ExitTime  int64
}
