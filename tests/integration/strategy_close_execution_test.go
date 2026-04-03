//go:build integration

package integration

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

// pyramidBars produces 9 bars with opens that step up in increments of 2 every two bars,
// then jump to exitOpen at bar 7.
// Bar 1 open=100, bar 3 open=102, bar 5 open=104, bar 7 open=exitOpen.
func pyramidBars(exitOpen float64) []map[string]interface{} {
	return []map[string]interface{}{
		{"time": int64(1704067200), "open": 100.0, "high": 101.0, "low": 99.0, "close": 100.0, "volume": 1000.0},
		{"time": int64(1704070800), "open": 100.0, "high": 102.0, "low": 98.0, "close": 101.0, "volume": 1000.0},
		{"time": int64(1704074400), "open": 102.0, "high": 103.0, "low": 101.0, "close": 102.0, "volume": 1000.0},
		{"time": int64(1704078000), "open": 102.0, "high": 104.0, "low": 100.0, "close": 103.0, "volume": 1000.0},
		{"time": int64(1704081600), "open": 104.0, "high": 105.0, "low": 103.0, "close": 104.0, "volume": 1000.0},
		{"time": int64(1704085200), "open": 104.0, "high": 106.0, "low": 102.0, "close": 105.0, "volume": 1000.0},
		{"time": int64(1704088800), "open": 106.0, "high": 107.0, "low": 105.0, "close": 106.0, "volume": 1000.0},
		{"time": int64(1704092400), "open": exitOpen, "high": exitOpen + 1, "low": exitOpen - 1, "close": exitOpen, "volume": 1000.0},
		{"time": int64(1704096000), "open": exitOpen, "high": exitOpen + 1, "low": exitOpen - 1, "close": exitOpen, "volume": 1000.0},
	}
}

/*
TestStrategyClose_ClosesAllMatchingEntriesByID verifies that strategy.close("id") exits every open
trade carrying that entry ID in a single fill at the next bar's open, per TradingView Pine semantics.
*/
func TestStrategyClose_ClosesAllMatchingEntriesByID(t *testing.T) {
	t.Parallel()
	// Bar 0→fills bar 1 (open=100): entry L1. Bar 2→fills bar 3 (open=102): entry L2.
	// Bar 4→fills bar 5 (open=104): entry L3. Bar 6→fills bar 7 (open=110): close("L") all three.
	// Expected profits: (110-100)*1=10, (110-102)*1=8, (110-104)*1=6.
	pineScript := `//@version=5
strategy("CloseAllMatchingByID", overlay=true, default_qty_type=strategy.fixed, default_qty_value=1, pyramiding=5)

if bar_index == 0
    strategy.entry("L", strategy.long)
if bar_index == 2
    strategy.entry("L", strategy.long)
if bar_index == 4
    strategy.entry("L", strategy.long)
if bar_index == 6
    strategy.close("L")
`
	exec := util.NewPineExecutor(t)
	raw := exec.ExecuteScriptWithCustomDataRaw(t, "strategy-close-all-matching", pineScript, pyramidBars(110.0))
	out := parseTradeAccessorOutput(t, raw)

	if len(out.Strategy.Trades) != 3 {
		t.Fatalf("closed trades: got %d, want 3", len(out.Strategy.Trades))
	}
	if len(out.Strategy.OpenTrades) != 0 {
		t.Fatalf("open trades: got %d, want 0", len(out.Strategy.OpenTrades))
	}

	wantProfits := []float64{10.0, 8.0, 6.0}
	for i, want := range wantProfits {
		got := out.Strategy.Trades[i].Profit
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("trade[%d].profit: got %.6f, want %.6f", i, got, want)
		}
		if out.Strategy.Trades[i].ExitPrice != 110.0 {
			t.Errorf("trade[%d].exitPrice: got %.6f, want 110.0", i, out.Strategy.Trades[i].ExitPrice)
		}
	}
}

/*
TestStrategyClose_OnlyClosesMatchingID verifies that strategy.close("id") leaves trades with a
different entry ID open, closing only the exact match.
*/
func TestStrategyClose_OnlyClosesMatchingID(t *testing.T) {
	t.Parallel()
	// Bar 0→fills bar 1 (open=100): entry A. Bar 2→fills bar 3 (open=102): entry B.
	// Bar 4→fills bar 5 (open=108): close("A"). B stays open.
	// Expected: 1 closed trade (A, profit=8), 1 open trade (B).
	pineScript := `//@version=5
strategy("CloseOnlyMatchingID", overlay=true, default_qty_type=strategy.fixed, default_qty_value=1, pyramiding=5)

if bar_index == 0
    strategy.entry("A", strategy.long)
if bar_index == 2
    strategy.entry("B", strategy.long)
if bar_index == 4
    strategy.close("A")
`
	bars := []map[string]interface{}{
		{"time": int64(1704067200), "open": 100.0, "high": 101.0, "low": 99.0, "close": 100.0, "volume": 1000.0},
		{"time": int64(1704070800), "open": 100.0, "high": 102.0, "low": 98.0, "close": 101.0, "volume": 1000.0},
		{"time": int64(1704074400), "open": 102.0, "high": 103.0, "low": 101.0, "close": 102.0, "volume": 1000.0},
		{"time": int64(1704078000), "open": 102.0, "high": 104.0, "low": 100.0, "close": 103.0, "volume": 1000.0},
		{"time": int64(1704081600), "open": 106.0, "high": 107.0, "low": 105.0, "close": 106.0, "volume": 1000.0},
		{"time": int64(1704085200), "open": 108.0, "high": 109.0, "low": 107.0, "close": 108.0, "volume": 1000.0},
		{"time": int64(1704088800), "open": 108.0, "high": 109.0, "low": 107.0, "close": 108.0, "volume": 1000.0},
	}

	exec := util.NewPineExecutor(t)
	raw := exec.ExecuteScriptWithCustomDataRaw(t, "strategy-close-only-matching", pineScript, bars)
	out := parseTradeAccessorOutput(t, raw)

	if len(out.Strategy.Trades) != 1 {
		t.Fatalf("closed trades: got %d, want 1", len(out.Strategy.Trades))
	}
	if len(out.Strategy.OpenTrades) != 1 {
		t.Fatalf("open trades: got %d, want 1", len(out.Strategy.OpenTrades))
	}

	closed := out.Strategy.Trades[0]
	if closed.EntryID != "A" {
		t.Errorf("closed.entryId: got %q, want \"A\"", closed.EntryID)
	}
	if math.Abs(closed.Profit-8.0) > 1e-9 {
		t.Errorf("closed.profit: got %.6f, want 8.0", closed.Profit)
	}
	if closed.ExitPrice != 108.0 {
		t.Errorf("closed.exitPrice: got %.6f, want 108.0", closed.ExitPrice)
	}

	open := out.Strategy.OpenTrades[0]
	if open.EntryID != "B" {
		t.Errorf("open.entryId: got %q, want \"B\"", open.EntryID)
	}
}

/*
TestStrategyCloseAll_ClosesAllOpenTradesRegardlessOfEntryID verifies that strategy.close_all()
exits every open trade in a single fill regardless of entry ID, per TradingView Pine semantics.
*/
func TestStrategyCloseAll_ClosesAllOpenTradesRegardlessOfEntryID(t *testing.T) {
	t.Parallel()
	// Bar 0→fills bar 1 (open=100): entry A. Bar 2→fills bar 3 (open=102): entry B.
	// Bar 4→fills bar 5 (open=104): entry C. Bar 6→fills bar 7 (open=110): close_all.
	// Expected profits: A=(110-100)*1=10, B=(110-102)*1=8, C=(110-104)*1=6.
	pineScript := `//@version=5
strategy("CloseAllRegardlessOfID", overlay=true, default_qty_type=strategy.fixed, default_qty_value=1, pyramiding=5)

if bar_index == 0
    strategy.entry("A", strategy.long)
if bar_index == 2
    strategy.entry("B", strategy.long)
if bar_index == 4
    strategy.entry("C", strategy.long)
if bar_index == 6
    strategy.close_all()
`
	exec := util.NewPineExecutor(t)
	raw := exec.ExecuteScriptWithCustomDataRaw(t, "strategy-close-all-ids", pineScript, pyramidBars(110.0))
	out := parseTradeAccessorOutput(t, raw)

	if len(out.Strategy.Trades) != 3 {
		t.Fatalf("closed trades: got %d, want 3", len(out.Strategy.Trades))
	}
	if len(out.Strategy.OpenTrades) != 0 {
		t.Fatalf("open trades: got %d, want 0", len(out.Strategy.OpenTrades))
	}

	wantEntries := []struct {
		id     string
		profit float64
	}{
		{"A", 10.0},
		{"B", 8.0},
		{"C", 6.0},
	}
	for i, want := range wantEntries {
		trade := out.Strategy.Trades[i]
		if trade.EntryID != want.id {
			t.Errorf("trade[%d].entryId: got %q, want %q", i, trade.EntryID, want.id)
		}
		if math.Abs(trade.Profit-want.profit) > 1e-9 {
			t.Errorf("trade[%d].profit: got %.6f, want %.6f", i, trade.Profit, want.profit)
		}
		if trade.ExitPrice != 110.0 {
			t.Errorf("trade[%d].exitPrice: got %.6f, want 110.0", i, trade.ExitPrice)
		}
	}
}
