package regression

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/strategy"
)

// TestStrategyAllowEntryIn_BlocksShortEntryButClosesOpenLongOnReversalSignal
// locks the TV-aligned semantics of strategy.risk.allow_entry_in(long):
//
//   - calls to strategy.entry("...", short) must NOT create a short order, AND
//   - any open long position must close as if the blocked entry had performed
//     the "close opposite" half of TV's reversal flow.
//
// Removing the close-on-blocked-entry behaviour from strategy.Entry surfaces the
// original Hull-strategy alignment failure (5x trade inflation + open positions
// that never close because the counter-signal cannot reverse them).
func TestStrategyAllowEntryIn_BlocksShortEntryButClosesOpenLongOnReversalSignal(t *testing.T) {
	s := strategy.NewStrategy()
	s.CallWithPyramiding("hull-like", 10000, 1)
	s.SetAllowedDirection(strategy.DirectionLong)

	// Bar 0: open long (allowed direction).
	if err := s.Entry("buy", strategy.Long, 10, ""); err != nil {
		t.Fatalf("long entry: %v", err)
	}
	s.OnBarUpdate(1, 100, 1000)
	if got := len(s.GetTradeHistory().GetOpenTrades()); got != 1 {
		t.Fatalf("after long entry want 1 open trade, got %d", got)
	}

	// Bar 1: counter-signal triggers strategy.entry("sell", short) — TV semantics
	// require the long to close even though the short open is blocked.
	if err := s.Entry("sell", strategy.Short, 10, ""); err != nil {
		t.Fatalf("blocked short entry returned error: %v", err)
	}
	s.OnBarUpdate(2, 110, 2000)

	open := s.GetTradeHistory().GetOpenTrades()
	closed := s.GetTradeHistory().GetClosedTrades()
	if len(open) != 0 {
		t.Errorf("blocked short must still close the open long, got %d open trades", len(open))
	}
	if len(closed) != 1 {
		t.Errorf("expected exactly 1 closed long after blocked-short reversal, got %d", len(closed))
	} else if closed[0].Direction != strategy.Long {
		t.Errorf("closed trade direction = %q, want %q", closed[0].Direction, strategy.Long)
	}

	// And no short position must have been opened.
	for _, tr := range open {
		if tr.Direction == strategy.Short {
			t.Errorf("blocked short direction created open trade: %+v", tr)
		}
	}
	for _, tr := range closed {
		if tr.Direction == strategy.Short {
			t.Errorf("blocked short direction created closed trade: %+v", tr)
		}
	}
}

// TestStrategyAllowEntryIn_BlockedEntryFromFlatIsNoop verifies that a blocked
// entry with no open opposite-direction position is a pure no-op — no order
// scheduled, no spurious close trade created. Locks against over-zealous
// implementations of the close-on-block behaviour that would emit phantom
// close orders.
func TestStrategyAllowEntryIn_BlockedEntryFromFlatIsNoop(t *testing.T) {
	s := strategy.NewStrategy()
	s.Call("flat-test", 10000)
	s.SetAllowedDirection(strategy.DirectionLong)

	if err := s.Entry("sell", strategy.Short, 5, ""); err != nil {
		t.Fatalf("blocked short entry: %v", err)
	}
	s.OnBarUpdate(1, 100, 1000)

	if got := len(s.GetTradeHistory().GetOpenTrades()); got != 0 {
		t.Errorf("blocked short from flat must not open any trade, got %d", got)
	}
	if got := len(s.GetTradeHistory().GetClosedTrades()); got != 0 {
		t.Errorf("blocked short from flat must not close any trade, got %d", got)
	}
}

// TestStrategyAllowEntryIn_DirectionAllAllowsBoth checks that the explicit
// DirectionAll constant (matching strategy.direction.all in Pine) does not
// block entries in either direction — the codegen for strategy.direction.all
// must resolve to a value the runtime gate treats as unrestricted.
//
// Trades alternate via TV's standard reversal semantics (each new opposite-
// direction entry closes the existing position), so the invariant checked here
// is that BOTH directions were able to OPEN at least one trade across the run.
func TestStrategyAllowEntryIn_DirectionAllAllowsBoth(t *testing.T) {
	s := strategy.NewStrategy()
	s.CallWithPyramiding("all-direction", 10000, 1)
	s.SetAllowedDirection(strategy.DirectionAll)

	if err := s.Entry("buy", strategy.Long, 5, ""); err != nil {
		t.Fatalf("long entry under DirectionAll: %v", err)
	}
	s.OnBarUpdate(1, 100, 1000)

	if err := s.Entry("sell", strategy.Short, 5, ""); err != nil {
		t.Fatalf("short entry under DirectionAll: %v", err)
	}
	s.OnBarUpdate(2, 100, 2000)

	closed := s.GetTradeHistory().GetClosedTrades()
	open := s.GetTradeHistory().GetOpenTrades()

	sawLong, sawShort := false, false
	for _, tr := range closed {
		if tr.Direction == strategy.Long {
			sawLong = true
		}
		if tr.Direction == strategy.Short {
			sawShort = true
		}
	}
	for _, tr := range open {
		if tr.Direction == strategy.Long {
			sawLong = true
		}
		if tr.Direction == strategy.Short {
			sawShort = true
		}
	}
	if !sawLong {
		t.Errorf("DirectionAll must permit long entries; open=%+v closed=%+v", open, closed)
	}
	if !sawShort {
		t.Errorf("DirectionAll must permit short entries; open=%+v closed=%+v", open, closed)
	}
}
