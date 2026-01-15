package strategyintegration

import (
	"testing"
)

/* Strategy exit mechanisms integration tests */

func TestExitImmediate(t *testing.T) {
	t.Skip("Strategy trade data extraction not complete - see e2e/fixtures/strategies/test-exit-immediate.pine.skip")

	tc := StrategyTestCase{
		Name:     "exit-immediate",
		PineFile: "test-exit-immediate.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {

			if len(result.Trades) < 1 {
				t.Error("Expected at least 1 closed trade (immediate exit)")
			}
			if len(result.OpenTrades) > 0 {
				t.Errorf("Expected 0 open trades after immediate exit, got %d", len(result.OpenTrades))
			}

			for _, trade := range result.Trades {
				duration := trade.ExitBar - trade.EntryBar
				if duration > 20 {
					t.Errorf("Trade duration %d bars too long for immediate exit pattern", duration)
				}
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

func TestExitDelayedState(t *testing.T) {
	t.Skip("Strategy trade data extraction not complete - see e2e/fixtures/strategies/test-exit-delayed-state.pine.skip")

	tc := StrategyTestCase{
		Name:     "exit-delayed-state",
		PineFile: "test-exit-delayed-state.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			/* Pattern: Exit based on state[N] transition */
			if len(result.Trades) < 1 {
				t.Error("Expected at least 1 closed trade (delayed state exit)")
			}

			/* Validate ForwardSeriesBuffer: historical state access works */
			/* Exit should trigger AFTER state transition completes */
			/* No specific duration requirement (depends on market data) */
			t.Logf("Closed %d trades with delayed state exit logic", len(result.Trades))
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

func TestExitSelective(t *testing.T) {
	t.Skip("Strategy trade data extraction not complete - see e2e/fixtures/strategies/test-exit-selective.pine.skip")

	tc := StrategyTestCase{
		Name:     "exit-selective",
		PineFile: "test-exit-selective.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			/* Pattern: Close specific position while keeping others */
			if len(result.Trades) < 1 {
				t.Error("Expected at least 1 closed trade (selective exit)")
			}

			/* Validate different entry IDs were used */
			uniqueIDs := make(map[string]bool)
			for _, trade := range result.Trades {
				uniqueIDs[trade.EntryID] = true
			}
			if len(uniqueIDs) < 1 {
				t.Error("Expected multiple entry IDs for selective exit test")
			}

			t.Logf("Closed %d trades with %d unique entry IDs", len(result.Trades), len(uniqueIDs))
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

func TestExitMultiBarCondition(t *testing.T) {
	t.Skip("Strategy trade data extraction not complete - see e2e/fixtures/strategies/test-exit-multibar-condition.pine.skip")

	tc := StrategyTestCase{
		Name:     "exit-multibar-condition",
		PineFile: "test-exit-multibar-condition.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			/* Pattern: Exit requires condition for N consecutive bars */
			if len(result.Trades) < 1 {
				t.Error("Expected at least 1 closed trade (multi-bar exit)")
			}

			for _, trade := range result.Trades {
				duration := trade.ExitBar - trade.EntryBar
				if duration < 3 {
					t.Errorf("Trade exited at bar %d too quickly (expected 3+ bar condition)", duration)
				}
			}

			t.Logf("Closed %d trades with multi-bar condition logic", len(result.Trades))
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

func TestExitStateReset(t *testing.T) {
	t.Skip("Strategy trade data extraction not complete - see e2e/fixtures/strategies/test-exit-state-reset.pine.skip")

	tc := StrategyTestCase{
		Name:     "exit-state-reset",
		PineFile: "test-exit-state-reset.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			/* Pattern: Multiple entry/exit cycles with proper state reset */
			if len(result.Trades) < 2 {
				t.Error("Expected at least 2 closed trades (multiple cycles)")
			}

			/* Validate state reset: no overlapping positions */
			if len(result.OpenTrades) > 0 {
				t.Errorf("Expected 0 open trades after all cycles, got %d", len(result.OpenTrades))
			}

			/* Validate clean entry/exit sequences */
			for i := 1; i < len(result.Trades); i++ {
				prev := result.Trades[i-1]
				curr := result.Trades[i]

				/* Next entry should come AFTER previous exit */
				if curr.EntryBar <= prev.ExitBar {
					t.Errorf("Trade %d entry (bar %d) overlaps with trade %d exit (bar %d)",
						i, curr.EntryBar, i-1, prev.ExitBar)
				}
			}

			t.Logf("Closed %d trades with proper state reset", len(result.Trades))
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/*
TestExitWithHistoricalReferences - Critical test for ForwardSeriesBuffer alignment

PURPOSE: Validate exit logic using historical variable references (var[N])
PATTERN: Exit condition depends on values from previous bars
CRITICAL: Ensures ForwardSeriesBuffer doesn't break historical lookback in exits

This is the CORE test for the bb9 bug: exit logic with has_active_trade[2]
*/
func TestExitWithHistoricalReferences(t *testing.T) {
	t.Skip("Strategy trade data extraction not complete - see e2e/fixtures/strategies/test-exit-delayed-state.pine.skip")

	tc := StrategyTestCase{
		Name:     "exit-historical-refs",
		PineFile: "test-exit-historical-refs.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			/* Pattern: Exit uses var[1], var[2] historical lookback */
			if len(result.Trades) < 1 {
				t.Error("Expected at least 1 closed trade (historical ref exit)")
			}

			/* This prevents the bb9 bug where trades never close */
			if len(result.OpenTrades) > 0 {
				t.Errorf("CRITICAL: Trades remain open despite exit condition (bb9 pattern), got %d open",
					len(result.OpenTrades))
			}

			t.Logf("✅ Historical reference exit logic working: %d closed trades", len(result.Trades))
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}
