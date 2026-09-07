package strategy

import (
	"math"
	"testing"
)

// newStrategyWithLongTrade returns a strategy with one open long trade that
// filled at bar fillBar open price fillPrice.
func newStrategyWithLongTrade(fillBar int, fillPrice float64) *Strategy {
	strat := NewStrategy()
	strat.Call("test", 10000)
	strat.SetDefaultQty(1, "fixed")
	strat.OnBarUpdate(fillBar-1, fillPrice-1, int64(fillBar-1)*1000)
	strat.Entry("buy", Long, 1, "")
	strat.OnBarUpdate(fillBar, fillPrice, int64(fillBar)*1000)
	strat.OnBarMetrics(fillPrice, fillPrice+1, fillPrice-1, int64(fillBar)*1000)
	return strat
}

// newStrategyWithShortTrade returns a strategy with one open short trade.
func newStrategyWithShortTrade(fillBar int, fillPrice float64) *Strategy {
	strat := NewStrategy()
	strat.Call("test", 10000)
	strat.SetDefaultQty(1, "fixed")
	strat.OnBarUpdate(fillBar-1, fillPrice+1, int64(fillBar-1)*1000)
	strat.Entry("sell", Short, 1, "")
	strat.OnBarUpdate(fillBar, fillPrice, int64(fillBar)*1000)
	strat.OnBarMetrics(fillPrice, fillPrice+1, fillPrice-1, int64(fillBar)*1000)
	return strat
}

// TestExitWithLevels_EntrySignalPattern covers the common Pine idiom where
// strategy.entry() and strategy.exit() are called together in the same
// conditional block on bar N, the entry fills on bar N+1, and the exit order
// is already eligible to trigger on that same bar N+1.
func TestExitWithLevels_EntrySignalPattern(t *testing.T) {
	tests := []struct {
		name      string
		stop      float64
		limit     float64
		barHigh   float64
		barLow    float64
		wantExit  bool
		wantPrice float64
	}{
		{"limit_hit_on_fill_bar", math.NaN(), 115, 116, 100, true, 115},
		{"stop_hit_on_fill_bar", 90, math.NaN(), 102, 89, true, 90},
		{"neither_hit_on_fill_bar", 90, 115, 105, 95, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strat := NewStrategy()
			strat.Call("test", 10000)
			strat.SetDefaultQty(1, "fixed")

			// Bar 0: signal bar — entry + exit registered together
			strat.OnBarUpdate(0, 100, 1000)
			strat.Entry("buy", Long, 1, "")
			strat.ExitWithLevels("exit", "buy", tt.stop, tt.limit, 102, 99, 100, 1000, "")
			strat.OnBarMetrics(100, 102, 99, 1000)

			// Bar 1: entry fills at open; exit is now eligible
			strat.OnBarUpdate(1, 101, 2000)
			strat.ExitWithLevels("exit", "buy", tt.stop, tt.limit, tt.barHigh, tt.barLow, 101, 2000, "")
			strat.OnBarMetrics(101, tt.barHigh, tt.barLow, 2000)

			closed := strat.tradeHistory.GetClosedTrades()
			if tt.wantExit {
				if len(closed) != 1 {
					t.Fatalf("expected 1 closed trade, got %d", len(closed))
				}
				if closed[0].ExitPrice != tt.wantPrice {
					t.Errorf("exitPrice = %v, want %v", closed[0].ExitPrice, tt.wantPrice)
				}
			} else {
				if len(closed) != 0 {
					t.Fatalf("expected 0 closed trades, got %d", len(closed))
				}
			}
		})
	}
}

// TestExitWithLevels_PostFillRegistration covers the scenario where strategy.exit()
// is first called on a bar after the trade is already open. The exit must not
// trigger on that same registration bar — only from the following bar onward.
func TestExitWithLevels_PostFillRegistration(t *testing.T) {
	tests := []struct {
		name            string
		stop            float64
		limit           float64
		registrationBar struct{ high, low float64 } // price would breach on registration bar
		nextBar         struct{ high, low float64 } // same price action, now eligible
		wantPrice       float64
		wantType        string
	}{
		{
			name: "stop_deferred_to_next_bar",
			stop: 95, limit: math.NaN(),
			registrationBar: struct{ high, low float64 }{100, 90},
			nextBar:         struct{ high, low float64 }{100, 90},
			wantPrice:       95, wantType: "stop",
		},
		{
			name: "limit_deferred_to_next_bar",
			stop: math.NaN(), limit: 115,
			registrationBar: struct{ high, low float64 }{120, 100},
			nextBar:         struct{ high, low float64 }{120, 100},
			wantPrice:       115, wantType: "limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strat := newStrategyWithLongTrade(1, 101)

			// Bar 2: exit registered; price already breaches level — must not trigger
			strat.OnBarUpdate(2, 102, 3000)
			strat.ExitWithLevels("exit", "buy", tt.stop, tt.limit,
				tt.registrationBar.high, tt.registrationBar.low, 95, 3000, "")
			strat.OnBarMetrics(102, tt.registrationBar.high, tt.registrationBar.low, 3000)

			if len(strat.tradeHistory.GetClosedTrades()) != 0 {
				t.Fatal("exit must not trigger on the bar it is first registered")
			}

			// Bar 3: now eligible
			strat.OnBarUpdate(3, 98, 4000)
			strat.ExitWithLevels("exit", "buy", tt.stop, tt.limit,
				tt.nextBar.high, tt.nextBar.low, 95, 4000, "")
			strat.OnBarMetrics(98, tt.nextBar.high, tt.nextBar.low, 4000)

			closed := strat.tradeHistory.GetClosedTrades()
			if len(closed) != 1 {
				t.Fatalf("expected 1 closed trade on bar after registration, got %d", len(closed))
			}
			if closed[0].ExitPrice != tt.wantPrice {
				t.Errorf("exitPrice = %v, want %v", closed[0].ExitPrice, tt.wantPrice)
			}
		})
	}
}

// TestExitWithLevels_LevelUpdateApplied verifies that when strategy.exit() is
// called again with updated stop/limit levels, the new levels are used when
// the exit eventually triggers, and the eligibility gate is not reset.
func TestExitWithLevels_LevelUpdateApplied(t *testing.T) {
	strat := newStrategyWithLongTrade(1, 101)

	// Bar 2: initial registration with stop=90
	strat.OnBarUpdate(2, 102, 3000)
	strat.ExitWithLevels("exit", "buy", 90, math.NaN(), 100, 95, 97, 3000, "")
	strat.OnBarMetrics(99, 100, 95, 3000) // not triggered: registration bar

	// Bar 3: stop updated to 97; old level would not have triggered
	strat.OnBarUpdate(3, 100, 4000)
	strat.ExitWithLevels("exit", "buy", 97, math.NaN(), 100, 96, 98, 4000, "")
	strat.OnBarMetrics(98, 100, 96, 4000)

	closed := strat.tradeHistory.GetClosedTrades()
	if len(closed) != 1 {
		t.Fatalf("expected 1 closed trade with updated stop, got %d", len(closed))
	}
	if closed[0].ExitPrice != 97 {
		t.Errorf("exitPrice = %v, want 97 (updated stop level)", closed[0].ExitPrice)
	}
}

// TestExitWithLevels_NoBreach verifies that no exit fires across many bars
// when price never reaches stop or limit, and the exit order persists correctly.
func TestExitWithLevels_NoBreach(t *testing.T) {
	strat := newStrategyWithLongTrade(1, 100)

	stop, limit := 80.0, 130.0
	for bar := 2; bar <= 20; bar++ {
		strat.OnBarUpdate(bar, 100, int64(bar)*1000)
		strat.ExitWithLevels("exit", "buy", stop, limit, 110, 90, 100, int64(bar)*1000, "")
		strat.OnBarMetrics(100, 110, 90, int64(bar)*1000)
	}

	if len(strat.tradeHistory.GetClosedTrades()) != 0 {
		t.Errorf("expected 0 closed trades when price never breaches stop/limit, got %d",
			len(strat.tradeHistory.GetClosedTrades()))
	}
	if len(strat.tradeHistory.GetOpenTrades()) != 1 {
		t.Errorf("expected 1 open trade, got %d", len(strat.tradeHistory.GetOpenTrades()))
	}
}

// TestExitWithLevels_ShortDirection verifies stop and limit trigger semantics
// for short trades where directions are inverted relative to long.
func TestExitWithLevels_ShortDirection(t *testing.T) {
	tests := []struct {
		name      string
		stop      float64
		limit     float64
		barHigh   float64
		barLow    float64
		wantTrig  bool
		wantPrice float64
	}{
		{"short_stop_hit", 115, math.NaN(), 116, 100, true, 115},
		{"short_stop_not_hit", 115, math.NaN(), 114, 100, false, 0},
		{"short_limit_hit", math.NaN(), 85, 100, 84, true, 85},
		{"short_limit_not_hit", math.NaN(), 85, 100, 86, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strat := newStrategyWithShortTrade(1, 100)

			// Bar 2: register exit (registration bar — not eligible yet)
			strat.OnBarUpdate(2, 100, 3000)
			strat.ExitWithLevels("exit", "sell", tt.stop, tt.limit, 102, 98, 100, 3000, "")
			strat.OnBarMetrics(100, 102, 98, 3000)

			// Bar 3: eligible; price action applies
			strat.OnBarUpdate(3, 100, 4000)
			strat.ExitWithLevels("exit", "sell", tt.stop, tt.limit, tt.barHigh, tt.barLow, 100, 4000, "")
			strat.OnBarMetrics(100, tt.barHigh, tt.barLow, 4000)

			closed := strat.tradeHistory.GetClosedTrades()
			if tt.wantTrig {
				if len(closed) != 1 {
					t.Fatalf("expected 1 closed trade, got %d", len(closed))
				}
				if closed[0].ExitPrice != tt.wantPrice {
					t.Errorf("exitPrice = %v, want %v", closed[0].ExitPrice, tt.wantPrice)
				}
			} else {
				if len(closed) != 0 {
					t.Fatalf("expected 0 closed trades, got %d", len(closed))
				}
			}
		})
	}
}

// TestExecuteCloseAll_ClearsPendingExits verifies that a CloseAll removes all
// pending exit orders so no orphaned exits remain after a position is closed.
func TestExecuteCloseAll_ClearsPendingExits(t *testing.T) {
	strat := newStrategyWithLongTrade(1, 101)

	strat.OnBarUpdate(2, 102, 3000)
	strat.ExitWithLevels("exit", "buy", 90, 120, 102, 100, 101, 3000, "")

	if len(strat.pendingExitManager.exitOrders) != 1 {
		t.Fatal("expected 1 pending exit registered")
	}

	strat.CloseAll(102, 3000, "")
	strat.OnBarUpdate(3, 102, 4000)

	if len(strat.pendingExitManager.exitOrders) != 0 {
		t.Errorf("pending exits should be cleared after CloseAll fills, got %d",
			len(strat.pendingExitManager.exitOrders))
	}
}

// TestCloseAll_SuppressesIntrabarExit verifies the close_all priority rule: when
// CloseAll is called on bar N (before OnBarMetrics), any pending stop or limit that
// would breach during that bar's price action must not fire. The position closes at
// bar N+1 open price via the queued close-all order, not at the stop/limit level.
// Parameterised over stop-only, limit-only, and both-breached cases for both long
// and short directions.
func TestCloseAll_SuppressesIntrabarExit(t *testing.T) {
	tests := []struct {
		name        string
		direction   string
		stop        float64
		limit       float64
		barHigh     float64 // would breach stop or limit when eligible
		barLow      float64
		nextBarOpen float64 // expected fill price (close-all fills at next bar open)
	}{
		// Long: stop below entry, limit above entry
		{"long_stop_suppressed", Long, 90, math.NaN(), 100, 85, 102},
		{"long_limit_suppressed", Long, math.NaN(), 120, 125, 100, 98},
		{"long_stop_and_limit_both_breached", Long, 90, 120, 125, 85, 102},
		// Short: stop above entry, limit below entry
		{"short_stop_suppressed", Short, 115, math.NaN(), 116, 100, 98},
		{"short_limit_suppressed", Short, math.NaN(), 85, 100, 84, 102},
		{"short_stop_and_limit_both_breached", Short, 115, 85, 116, 84, 102},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var strat *Strategy
			var entryID string
			if tt.direction == Long {
				strat = newStrategyWithLongTrade(1, 100)
				entryID = "buy"
			} else {
				strat = newStrategyWithShortTrade(1, 100)
				entryID = "sell"
			}

			// Bar 2: register exit — not yet eligible (registration bar)
			strat.OnBarUpdate(2, 100, 3000)
			strat.ExitWithLevels("exit", entryID, tt.stop, tt.limit, 101, 99, 100, 3000, "")
			strat.OnBarMetrics(100, 101, 99, 3000)

			if len(strat.tradeHistory.GetClosedTrades()) != 0 {
				t.Fatal("setup: exit must not trigger on registration bar")
			}

			// Bar 3: exit is eligible; price breaches stop/limit; but CloseAll fires before OnBarMetrics
			strat.OnBarUpdate(3, 100, 4000)
			strat.ExitWithLevels("exit", entryID, tt.stop, tt.limit, tt.barHigh, tt.barLow, 100, 4000, "")
			strat.CloseAll(100, 4000, "")
			strat.OnBarMetrics(100, tt.barHigh, tt.barLow, 4000)

			if len(strat.tradeHistory.GetClosedTrades()) != 0 {
				t.Fatalf("stop/limit must not fire on bar where CloseAll was called: got %d closed trades",
					len(strat.tradeHistory.GetClosedTrades()))
			}
			if len(strat.tradeHistory.GetOpenTrades()) != 1 {
				t.Fatalf("trade must remain open until next bar open: got %d open trades",
					len(strat.tradeHistory.GetOpenTrades()))
			}

			// Bar 4: close-all order fills at this bar's open price
			strat.OnBarUpdate(4, tt.nextBarOpen, 5000)
			strat.OnBarMetrics(tt.nextBarOpen, tt.nextBarOpen+5, tt.nextBarOpen-5, 5000)

			closed := strat.tradeHistory.GetClosedTrades()
			if len(closed) != 1 {
				t.Fatalf("expected 1 closed trade after close-all fill bar, got %d", len(closed))
			}
			if closed[0].ExitPrice != tt.nextBarOpen {
				t.Errorf("exit price: got %.2f, want %.2f (next bar open, not stop/limit level)",
					closed[0].ExitPrice, tt.nextBarOpen)
			}
			if closed[0].ExitBar != 4 {
				t.Errorf("exit bar: got %d, want 4", closed[0].ExitBar)
			}
			if len(strat.pendingExitManager.exitOrders) != 0 {
				t.Errorf("pending exits must be empty after fill, got %d", len(strat.pendingExitManager.exitOrders))
			}
		})
	}
}

// TestCloseAll_CancelsExitsAcrossAllOpenTrades verifies that CloseAll removes
// pending exits for every open trade simultaneously (not just one), and that
// every trade fills at the next bar's open via the single queued close-all order.
func TestCloseAll_CancelsExitsAcrossAllOpenTrades(t *testing.T) {
	tests := []struct {
		name        string
		entryIDs    []string
		stop        float64
		limit       float64
		barHigh     float64 // price action when CloseAll fires; would breach exit level
		barLow      float64
		nextBarOpen float64
	}{
		{
			name:     "two_trades_stop_would_breach",
			entryIDs: []string{"a", "b"},
			stop:     90, limit: math.NaN(),
			barHigh: 100, barLow: 85, nextBarOpen: 102,
		},
		{
			name:     "three_trades_limit_would_breach",
			entryIDs: []string{"a", "b", "c"},
			stop:     math.NaN(), limit: 120,
			barHigh: 125, barLow: 100, nextBarOpen: 98,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pyramiding := len(tt.entryIDs)
			strat := NewStrategy()
			strat.CallWithPyramiding("test", 10000, pyramiding)

			// Signal each entry on its own bar; the next bar's OnBarUpdate fills it
			bar, ts := 0, int64(0)
			for _, id := range tt.entryIDs {
				strat.OnBarUpdate(bar, 100, ts)
				strat.Entry(id, Long, 1, "")
				bar++
				ts += 1000
			}
			strat.OnBarUpdate(bar, 100, ts) // fills the last entry
			strat.OnBarMetrics(100, 101, 99, ts)

			if got := len(strat.tradeHistory.GetOpenTrades()); got != pyramiding {
				t.Fatalf("setup: expected %d open trades, got %d", pyramiding, got)
			}

			// Register one exit per trade — registration bar, not yet eligible
			bar++
			ts += 1000
			strat.OnBarUpdate(bar, 100, ts)
			for _, id := range tt.entryIDs {
				strat.ExitWithLevels("exit_"+id, id, tt.stop, tt.limit, 101, 99, 100, ts, "")
			}
			strat.OnBarMetrics(100, 101, 99, ts)

			if len(strat.tradeHistory.GetClosedTrades()) != 0 {
				t.Fatal("setup: no trade should close on registration bar")
			}

			// CloseAll bar: exits are eligible, price breaches, but CloseAll fires before OnBarMetrics
			bar++
			ts += 1000
			strat.OnBarUpdate(bar, 100, ts)
			for _, id := range tt.entryIDs {
				strat.ExitWithLevels("exit_"+id, id, tt.stop, tt.limit, tt.barHigh, tt.barLow, 100, ts, "")
			}
			strat.CloseAll(100, ts, "")
			strat.OnBarMetrics(100, tt.barHigh, tt.barLow, ts)

			if got := len(strat.tradeHistory.GetClosedTrades()); got != 0 {
				t.Fatalf("no trade must close on CloseAll bar: got %d", got)
			}
			if got := len(strat.pendingExitManager.exitOrders); got != 0 {
				t.Fatalf("all pending exits must be cancelled by CloseAll: got %d", got)
			}

			// Next bar: all trades fill at open
			bar++
			ts += 1000
			strat.OnBarUpdate(bar, tt.nextBarOpen, ts)
			strat.OnBarMetrics(tt.nextBarOpen, tt.nextBarOpen+5, tt.nextBarOpen-5, ts)

			closed := strat.tradeHistory.GetClosedTrades()
			if len(closed) != pyramiding {
				t.Fatalf("expected %d closed trades after close-all fill, got %d", pyramiding, len(closed))
			}
			for i, c := range closed {
				if c.ExitPrice != tt.nextBarOpen {
					t.Errorf("trade[%d] exit price: got %.2f, want %.2f (next bar open)",
						i, c.ExitPrice, tt.nextBarOpen)
				}
			}
		})
	}
}

// TestCloseAll_WithNoPendingExits still queues the close-all order and closes the
// position at the next bar open. Pending exits are not a precondition for CloseAll.
func TestCloseAll_WithNoPendingExits(t *testing.T) {
	tests := []struct {
		name        string
		direction   string
		nextBarOpen float64
	}{
		{"long_no_pending_exits", Long, 105},
		{"short_no_pending_exits", Short, 95},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var strat *Strategy
			if tt.direction == Long {
				strat = newStrategyWithLongTrade(1, 100)
			} else {
				strat = newStrategyWithShortTrade(1, 100)
			}

			// No ExitWithLevels registered at any point
			strat.OnBarUpdate(2, 100, 3000)
			strat.CloseAll(100, 3000, "")
			strat.OnBarMetrics(100, 105, 95, 3000)

			if len(strat.tradeHistory.GetClosedTrades()) != 0 {
				t.Fatal("close must not fill on the CloseAll bar itself")
			}

			strat.OnBarUpdate(3, tt.nextBarOpen, 4000)
			strat.OnBarMetrics(tt.nextBarOpen, tt.nextBarOpen+5, tt.nextBarOpen-5, 4000)

			closed := strat.tradeHistory.GetClosedTrades()
			if len(closed) != 1 {
				t.Fatalf("expected 1 closed trade at next bar, got %d", len(closed))
			}
			if closed[0].ExitPrice != tt.nextBarOpen {
				t.Errorf("exit price: got %.2f, want %.2f (next bar open)", closed[0].ExitPrice, tt.nextBarOpen)
			}
		})
	}
}

// TestExitWithLevels_BothNaN_NeverFires verifies that registering an exit with
// stop=NaN and limit=NaN on every bar never triggers a close, regardless of how
// large the price moves are. This is the base case for strategies that compute
// their TP/SL from secondary-timeframe data (e.g. daily pivots): when the data
// feed returns all-NaN, no exit must silently fire.
func TestExitWithLevels_BothNaN_NeverFires(t *testing.T) {
	for _, dir := range []struct {
		name  string
		setup func() (*Strategy, string)
	}{
		{"long", func() (*Strategy, string) { return newStrategyWithLongTrade(1, 100), "buy" }},
		{"short", func() (*Strategy, string) { return newStrategyWithShortTrade(1, 100), "sell" }},
	} {
		t.Run(dir.name, func(t *testing.T) {
			strat, entryID := dir.setup()

			for bar := 2; bar <= 30; bar++ {
				strat.OnBarUpdate(bar, 100, int64(bar)*1000)
				strat.ExitWithLevels("exit", entryID, math.NaN(), math.NaN(),
					200, 50, 100, int64(bar)*1000, "")
				// Large bar range (50..200) that would breach any non-NaN stop or limit
				strat.OnBarMetrics(100, 200, 50, int64(bar)*1000)
			}

			if got := len(strat.tradeHistory.GetClosedTrades()); got != 0 {
				t.Errorf("expected 0 closed trades when stop=NaN and limit=NaN, got %d", got)
			}
			if got := len(strat.tradeHistory.GetOpenTrades()); got != 1 {
				t.Errorf("expected 1 open trade, got %d", got)
			}
		})
	}
}

// TestExitWithLevels_IndependentPerEntry_OnlyMatchingBreachFires verifies that
// when two concurrent open trades have separate exit registrations, breaching
// one trade's stop level closes only that trade and leaves the other open. This
// covers the per-entry isolation invariant: a pending exit bound to entryID A
// must never close a trade for entryID B.
func TestExitWithLevels_IndependentPerEntry_OnlyMatchingBreachFires(t *testing.T) {
	strat := NewStrategy()
	strat.Call("test", 100000)
	strat.SetDefaultQty(1, "fixed")

	// Bar 0: schedule two long entries with different IDs via pyramiding.
	strat.OnBarUpdate(0, 99, 1000)
	strat.Entry("buy1", Long, 1, "")
	strat.Entry("buy2", Long, 1, "")
	strat.OnBarMetrics(99, 100, 98, 1000)

	// Bar 1: both entries fill at open 100.
	strat.OnBarUpdate(1, 100, 2000)
	strat.OnBarMetrics(100, 101, 99, 2000)

	if got := len(strat.tradeHistory.GetOpenTrades()); got != 2 {
		t.Fatalf("expected 2 open trades after dual entry, got %d", got)
	}

	// Bar 2: register stop exits — buy1 stop=90, buy2 stop=80 (not eligible yet).
	strat.OnBarUpdate(2, 100, 3000)
	strat.ExitWithLevels("exit1", "buy1", 90, math.NaN(), 102, 98, 100, 3000, "")
	strat.ExitWithLevels("exit2", "buy2", 80, math.NaN(), 102, 98, 100, 3000, "")
	strat.OnBarMetrics(100, 102, 98, 3000)

	// Bar 3: eligible; low=88 breaches buy1 stop (88 ≤ 90) but not buy2 stop (88 > 80).
	strat.OnBarUpdate(3, 100, 4000)
	strat.ExitWithLevels("exit1", "buy1", 90, math.NaN(), 95, 88, 100, 4000, "")
	strat.ExitWithLevels("exit2", "buy2", 80, math.NaN(), 95, 88, 100, 4000, "")
	strat.OnBarMetrics(100, 95, 88, 4000)

	closed := strat.tradeHistory.GetClosedTrades()
	if len(closed) != 1 {
		t.Fatalf("expected 1 closed trade (buy1), got %d", len(closed))
	}
	if closed[0].EntryID != "buy1" {
		t.Errorf("wrong trade closed: got entryID=%q, want buy1", closed[0].EntryID)
	}
	if closed[0].ExitPrice != 90 {
		t.Errorf("exit price: got %.2f, want 90 (buy1 stop level)", closed[0].ExitPrice)
	}

	open := strat.tradeHistory.GetOpenTrades()
	if len(open) != 1 {
		t.Fatalf("expected 1 open trade (buy2), got %d", len(open))
	}
	if open[0].EntryID != "buy2" {
		t.Errorf("wrong trade remains open: got entryID=%q, want buy2", open[0].EntryID)
	}
}

// TestExitWithLevels_UnknownEntryId_SafeNoOp verifies that calling ExitWithLevels
// with an entryId that has no corresponding open trade is a safe no-op: it must
// not crash, must not close the unrelated open trade, and must not corrupt pending
// exit state. This covers the class of strategies that call strategy.exit() for an
// entry that was already closed or was never opened (e.g. conditional entry blocks
// where the entry did not fire on that bar).
func TestExitWithLevels_UnknownEntryId_SafeNoOp(t *testing.T) {
	strat := newStrategyWithLongTrade(1, 100)

	// Bar 2: register exit for "buy" (the open trade) and also for "ghost"
	// (an entryId with no open trade). ghost must be silently ignored.
	strat.OnBarUpdate(2, 100, 3000)
	strat.ExitWithLevels("exit", "buy", 90, 120, 102, 98, 100, 3000, "")
	strat.ExitWithLevels("ghost_exit", "ghost", 95, 115, 102, 98, 100, 3000, "")
	strat.OnBarMetrics(100, 102, 98, 3000)

	// Bar 3: eligible; price breaches ghost stop (if it were active) but not buy stop.
	// Only the active exit for "buy" should be evaluated.
	strat.OnBarUpdate(3, 100, 4000)
	strat.ExitWithLevels("exit", "buy", 90, 120, 108, 93, 100, 4000, "")
	strat.ExitWithLevels("ghost_exit", "ghost", 95, 115, 108, 93, 100, 4000, "")
	// low=93 < ghost_stop=95 → would fire if ghost were active
	// low=93 > buy_stop=90 → buy exit must NOT fire
	strat.OnBarMetrics(100, 108, 93, 4000)

	if got := len(strat.tradeHistory.GetClosedTrades()); got != 0 {
		t.Errorf("expected 0 closed trades (ghost exit must be ignored, buy not breached), got %d", got)
	}
	if got := len(strat.tradeHistory.GetOpenTrades()); got != 1 {
		t.Errorf("expected 1 open trade, got %d", got)
	}
	if got := strat.tradeHistory.GetOpenTrades()[0].EntryID; got != "buy" {
		t.Errorf("open trade entryID: got %q, want buy", got)
	}
}
