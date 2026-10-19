package strategy

import (
	"testing"
)

// TestCloseAll_ThenEntryOnSameBar_BothFillNextBarOpen verifies that when CloseAll
// and a new Entry are both signalled on bar N (default execution mode), the position
// closes and the re-entry both fill at bar N+1 open. This is the correct Pine
// semantics: orders queued on bar N are executed at the start of bar N+1.
func TestCloseAll_ThenEntryOnSameBar_BothFillNextBarOpen(t *testing.T) {
	tests := []struct {
		name       string
		initialDir string
		reentryDir string
	}{
		{"long_closes_and_reenters_long", Long, Long},
		{"long_closes_and_opens_short", Long, Short},
		{"short_closes_and_reopens_short", Short, Short},
		{"short_closes_and_opens_long", Short, Long},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strat := NewStrategy()
			strat.CallWithPyramiding("test", 10000, 2)
			strat.SetDefaultQty(1, "fixed")

			strat.OnBarUpdate(0, 100, 0)
			strat.Entry("initial", tt.initialDir, 1, "")

			strat.OnBarUpdate(1, 100, 1000)
			strat.OnBarMetrics(100, 101, 99, 1000)

			if got := len(strat.tradeHistory.GetOpenTrades()); got != 1 {
				t.Fatalf("bar 1: want 1 open trade, got %d", got)
			}

			strat.OnBarUpdate(2, 100, 2000)
			strat.CloseAll(100, 2000, "")
			strat.Entry("reentry", tt.reentryDir, 1, "")
			strat.OnBarMetrics(100, 101, 99, 2000)

			if got := len(strat.tradeHistory.GetClosedTrades()); got != 0 {
				t.Fatalf("bar 2: no trade must close during signal bar, got %d", got)
			}
			if got := len(strat.tradeHistory.GetOpenTrades()); got != 1 {
				t.Fatalf("bar 2: original trade must still be open, got %d open", got)
			}

			const nextOpen = 103.0
			strat.OnBarUpdate(3, nextOpen, 3000)
			strat.OnBarMetrics(nextOpen, nextOpen+5, nextOpen-5, 3000)

			closed := strat.tradeHistory.GetClosedTrades()
			if len(closed) != 1 {
				t.Fatalf("bar 3: want 1 closed trade, got %d", len(closed))
			}
			if closed[0].ExitPrice != nextOpen {
				t.Errorf("close fill price: got %.2f, want %.2f (bar 3 open)", closed[0].ExitPrice, nextOpen)
			}
			if closed[0].ExitBar != 3 {
				t.Errorf("close fill bar: got %d, want 3", closed[0].ExitBar)
			}

			opens := strat.tradeHistory.GetOpenTrades()
			if len(opens) != 1 {
				t.Fatalf("bar 3: want 1 open (re-entry) trade, got %d", len(opens))
			}
			if opens[0].EntryPrice != nextOpen {
				t.Errorf("re-entry fill price: got %.2f, want %.2f (bar 3 open)", opens[0].EntryPrice, nextOpen)
			}
			if opens[0].EntryBar != 3 {
				t.Errorf("re-entry fill bar: got %d, want 3", opens[0].EntryBar)
			}
			if opens[0].Direction != tt.reentryDir {
				t.Errorf("re-entry direction: got %q, want %q", opens[0].Direction, tt.reentryDir)
			}
		})
	}
}

// TestCloseAll_ThenMultipleEntriesSameBar verifies that when CloseAll and
// multiple new entries are all signalled on bar N, all entries fill at bar N+1
// open regardless of the number of positions being closed.
func TestCloseAll_ThenMultipleEntriesSameBar(t *testing.T) {
	strat := NewStrategy()
	strat.CallWithPyramiding("test", 100000, 4)
	strat.SetDefaultQty(1, "fixed")

	strat.OnBarUpdate(0, 100, 0)
	strat.Entry("a", Long, 1, "")
	strat.OnBarUpdate(1, 100, 1000)
	strat.Entry("b", Long, 1, "")
	strat.OnBarUpdate(2, 100, 2000)
	strat.OnBarMetrics(100, 101, 99, 2000)

	if got := len(strat.tradeHistory.GetOpenTrades()); got != 2 {
		t.Fatalf("setup: want 2 open trades, got %d", got)
	}

	strat.OnBarUpdate(3, 100, 3000)
	strat.CloseAll(100, 3000, "")
	strat.Entry("c", Short, 1, "")
	strat.Entry("d", Short, 1, "")
	strat.OnBarMetrics(100, 101, 99, 3000)

	if got := len(strat.tradeHistory.GetClosedTrades()); got != 0 {
		t.Fatalf("bar 3: original trades must not close during signal bar, got %d closed", got)
	}

	const fillOpen = 98.0
	strat.OnBarUpdate(4, fillOpen, 4000)
	strat.OnBarMetrics(fillOpen, fillOpen+2, fillOpen-2, 4000)

	closed := strat.tradeHistory.GetClosedTrades()
	if len(closed) != 2 {
		t.Fatalf("bar 4: want 2 closed (original longs), got %d", len(closed))
	}
	for i, c := range closed {
		if c.ExitPrice != fillOpen {
			t.Errorf("closed[%d] exit price: got %.2f, want %.2f", i, c.ExitPrice, fillOpen)
		}
	}

	opens := strat.tradeHistory.GetOpenTrades()
	if len(opens) != 2 {
		t.Fatalf("bar 4: want 2 open (new shorts), got %d", len(opens))
	}
	for i, o := range opens {
		if o.EntryPrice != fillOpen {
			t.Errorf("opens[%d] entry price: got %.2f, want %.2f", i, o.EntryPrice, fillOpen)
		}
		if o.Direction != Short {
			t.Errorf("opens[%d] direction: got %q, want Short", i, o.Direction)
		}
	}
}

// TestCloseAll_NoOpenTrades_IsNoOp verifies that CloseAll with no open positions
// produces no closed trades — the close side is a no-op when there is nothing to
// close. Entry orders queued on the same bar remain unaffected.
func TestCloseAll_NoOpenTrades_IsNoOp(t *testing.T) {
	strat := NewStrategy()
	strat.CallWithPyramiding("test", 10000, 2)
	strat.SetDefaultQty(1, "fixed")

	strat.OnBarUpdate(0, 100, 0)
	strat.CloseAll(100, 0, "")
	strat.Entry("e1", Long, 1, "")
	strat.OnBarMetrics(100, 101, 99, 0)

	if got := len(strat.tradeHistory.GetClosedTrades()); got != 0 {
		t.Fatalf("bar 0: CloseAll with no positions must produce 0 closed trades, got %d", got)
	}
	if got := len(strat.tradeHistory.GetOpenTrades()); got != 0 {
		t.Fatalf("bar 0: entry not yet filled, want 0 open, got %d", got)
	}

	const fillOpen = 102.0
	strat.OnBarUpdate(1, fillOpen, 1000)
	strat.OnBarMetrics(fillOpen, fillOpen+2, fillOpen-2, 1000)

	opens := strat.tradeHistory.GetOpenTrades()
	if len(opens) != 1 {
		t.Fatalf("bar 1: want 1 open trade, got %d", len(opens))
	}
	if opens[0].EntryPrice != fillOpen {
		t.Errorf("entry fill price: got %.2f, want %.2f", opens[0].EntryPrice, fillOpen)
	}
}
