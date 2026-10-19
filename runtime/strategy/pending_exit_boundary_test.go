package strategy

import (
	"math"
	"testing"
)

// TestCloseAll_FillsAtNextBarOpen_RegardlessOfExitLevels verifies that a queued
// close-all order always fills at the next bar open, regardless of where that open
// falls relative to any registered stop or limit. This covers adverse gaps through
// the stop level and favorable gaps through the limit level — both must fill at
// bar open, never at the registered exit price.
func TestCloseAll_FillsAtNextBarOpen_RegardlessOfExitLevels(t *testing.T) {
	tests := []struct {
		name      string
		direction string
		stop      float64
		limit     float64
		gapOpen   float64
	}{
		// Adverse gaps: bar opens beyond stop level (worse than stop for the trader)
		{"long_adverse_gap_through_stop", Long, 95, math.NaN(), 80},
		{"short_adverse_gap_through_stop", Short, 105, math.NaN(), 120},
		// Favorable gaps: bar opens beyond limit level (better than limit for the trader)
		{"long_favorable_gap_through_limit", Long, math.NaN(), 115, 130},
		{"short_favorable_gap_through_limit", Short, math.NaN(), 85, 70},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const entryPrice = 100.0
			var (
				strat   *Strategy
				entryID string
			)
			if tt.direction == Long {
				strat = newStrategyWithLongTrade(1, entryPrice)
				entryID = "buy"
			} else {
				strat = newStrategyWithShortTrade(1, entryPrice)
				entryID = "sell"
			}

			// Bar 2: register exit — not eligible (registration bar).
			strat.OnBarUpdate(2, entryPrice, 3000)
			strat.ExitWithLevels("exit", entryID, tt.stop, tt.limit,
				entryPrice+1, entryPrice-1, entryPrice, 3000, "")
			strat.OnBarMetrics(entryPrice, entryPrice+1, entryPrice-1, 3000)

			// Bar 3: exit eligible, no breach; CloseAll queued.
			strat.OnBarUpdate(3, entryPrice, 4000)
			strat.ExitWithLevels("exit", entryID, tt.stop, tt.limit,
				entryPrice+1, entryPrice-1, entryPrice, 4000, "")
			strat.CloseAll(entryPrice, 4000, "")
			strat.OnBarMetrics(entryPrice, entryPrice+1, entryPrice-1, 4000)

			if got := len(strat.tradeHistory.GetClosedTrades()); got != 0 {
				t.Fatalf("no close must happen on CloseAll bar, got %d", got)
			}

			// Bar 4: opens through a registered exit level — CloseAll fills at open.
			strat.OnBarUpdate(4, tt.gapOpen, 5000)
			strat.OnBarMetrics(tt.gapOpen, tt.gapOpen+5, tt.gapOpen-5, 5000)

			closed := strat.tradeHistory.GetClosedTrades()
			if len(closed) != 1 {
				t.Fatalf("expected 1 closed trade after fill bar, got %d", len(closed))
			}
			if closed[0].ExitPrice != tt.gapOpen {
				t.Errorf("exit price: got %.2f, want %.2f (gap open, not exit level)",
					closed[0].ExitPrice, tt.gapOpen)
			}
			if closed[0].ExitBar != 4 {
				t.Errorf("exit bar: got %d, want 4", closed[0].ExitBar)
			}
			if len(strat.pendingExitManager.exitOrders) != 0 {
				t.Errorf("pending exits must be empty after close-all fill, got %d",
					len(strat.pendingExitManager.exitOrders))
			}
		})
	}
}

// TestExitWithLevels_LevelSelection_BothRegistered verifies that when both a stop
// and limit are registered simultaneously, only the breached level fires and the
// correct fill price is used. Tests all three outcomes (stop only, limit only,
// neither) for both long and short trade directions.
func TestExitWithLevels_LevelSelection_BothRegistered(t *testing.T) {
	tests := []struct {
		name      string
		direction string
		stop      float64
		limit     float64
		barHigh   float64
		barLow    float64
		wantTrig  bool
		wantPrice float64
	}{
		// Long: stop below entry, limit above entry
		{"long_only_stop_breached", Long, 90, 115, 109, 85, true, 90},
		{"long_only_limit_breached", Long, 90, 115, 120, 95, true, 115},
		{"long_neither_breached", Long, 90, 115, 109, 92, false, 0},
		// Long: both breached in same bar — stop always wins (conservative-fill rule)
		{"long_both_breached_stop_wins", Long, 90, 115, 120, 85, true, 90},
		// Short: stop above entry, limit below entry
		{"short_only_stop_breached", Short, 110, 88, 112, 92, true, 110},
		{"short_only_limit_breached", Short, 110, 88, 108, 85, true, 88},
		{"short_neither_breached", Short, 110, 88, 108, 92, false, 0},
		// Short: both breached in same bar — stop always wins (conservative-fill rule)
		{"short_both_breached_stop_wins", Short, 110, 88, 115, 83, true, 110},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				strat   *Strategy
				entryID string
			)
			if tt.direction == Long {
				strat = newStrategyWithLongTrade(1, 100)
				entryID = "buy"
			} else {
				strat = newStrategyWithShortTrade(1, 100)
				entryID = "sell"
			}

			// Bar 2: register both levels — not eligible (registration bar).
			strat.OnBarUpdate(2, 100, 3000)
			strat.ExitWithLevels("exit", entryID, tt.stop, tt.limit, 101, 99, 100, 3000, "")
			strat.OnBarMetrics(100, 101, 99, 3000)

			// Bar 3: eligible; price action selects which level fires.
			strat.OnBarUpdate(3, 100, 4000)
			strat.ExitWithLevels("exit", entryID, tt.stop, tt.limit, tt.barHigh, tt.barLow, 100, 4000, "")
			strat.OnBarMetrics(100, tt.barHigh, tt.barLow, 4000)

			closed := strat.tradeHistory.GetClosedTrades()
			if tt.wantTrig {
				if len(closed) != 1 {
					t.Fatalf("expected 1 closed trade, got %d", len(closed))
				}
				if closed[0].ExitPrice != tt.wantPrice {
					t.Errorf("exit price: got %.2f, want %.2f", closed[0].ExitPrice, tt.wantPrice)
				}
			} else {
				if len(closed) != 0 {
					t.Fatalf("expected 0 closed trades (neither level breached), got %d", len(closed))
				}
			}
		})
	}
}
