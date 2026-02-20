//go:build integration

package integration

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

type tradeAccessorOutput struct {
	Indicators map[string]struct {
		Data []struct {
			Time  int64   `json:"time"`
			Value float64 `json:"value"`
		} `json:"data"`
	} `json:"indicators"`
	Strategy struct {
		Trades []struct {
			EntryID    string  `json:"entryId"`
			EntryPrice float64 `json:"entryPrice"`
			ExitPrice  float64 `json:"exitPrice"`
			Size       float64 `json:"size"`
			Profit     float64 `json:"profit"`
			Direction  string  `json:"direction"`
		} `json:"trades"`
		OpenTrades []struct {
			EntryID    string  `json:"entryId"`
			EntryPrice float64 `json:"entryPrice"`
			Size       float64 `json:"size"`
			Direction  string  `json:"direction"`
		} `json:"openTrades"`
	} `json:"strategy"`
}

func parseTradeAccessorOutput(t *testing.T, raw []byte) tradeAccessorOutput {
	t.Helper()
	var out tradeAccessorOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	return out
}

/* lastPlotValue returns the last non-NaN value emitted for the named plot indicator */
func lastPlotValue(out tradeAccessorOutput, title string) (float64, bool) {
	series, ok := out.Indicators[title]
	if !ok {
		return 0, false
	}
	for i := len(series.Data) - 1; i >= 0; i-- {
		v := series.Data[i].Value
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			return v, true
		}
	}
	return 0, false
}

// Bar 0: entry signal. Bar 1: fills at open=100. Bar 2: close signal. Bar 3: exits at open=106. Bar 4: read plots.
// entryPrice=100, exitPrice=106, profit=6, size=1
func closedTradeSingleBars() []map[string]interface{} {
	return []map[string]interface{}{
		{"time": int64(1704067200), "open": 100.0, "high": 100.0, "low": 100.0, "close": 100.0, "volume": 1000.0},
		{"time": int64(1704070800), "open": 100.0, "high": 102.0, "low": 98.0, "close": 101.0, "volume": 1000.0},
		{"time": int64(1704074400), "open": 101.0, "high": 105.0, "low": 99.0, "close": 103.0, "volume": 1000.0},
		{"time": int64(1704078000), "open": 106.0, "high": 108.0, "low": 104.0, "close": 107.0, "volume": 1000.0},
		{"time": int64(1704081600), "open": 107.0, "high": 109.0, "low": 105.0, "close": 108.0, "volume": 1000.0},
	}
}

func TestTradeAccessorExecution_ClosedTradeProperties(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
strategy("ClosedTradeProps", overlay=true, default_qty_type=strategy.fixed, default_qty_value=1)

if bar_index == 0
    strategy.entry("L1", strategy.long)

if bar_index == 2
    strategy.close("L1")

closedProfit = strategy.closedtrades.profit(0)
closedEntryPrice = strategy.closedtrades.entry_price(0)
closedExitPrice = strategy.closedtrades.exit_price(0)
closedSize = strategy.closedtrades.size(0)

plot(closedProfit, "ClosedProfit")
plot(closedEntryPrice, "ClosedEntryPrice")
plot(closedExitPrice, "ClosedExitPrice")
plot(closedSize, "ClosedSize")
`

	exec := util.NewPineExecutor(t)
	raw := exec.ExecuteScriptWithCustomDataRaw(t, "trade-accessor-closed-props", pineScript, closedTradeSingleBars())
	out := parseTradeAccessorOutput(t, raw)

	checks := map[string]float64{
		"ClosedProfit":     6.0,
		"ClosedEntryPrice": 100.0,
		"ClosedExitPrice":  106.0,
		"ClosedSize":       1.0,
	}
	for title, want := range checks {
		got, ok := lastPlotValue(out, title)
		if !ok {
			t.Errorf("%s: no non-NaN plot value", title)
			continue
		}
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("%s: got %.6f, want %.6f", title, got, want)
		}
	}

	if len(out.Strategy.Trades) != 1 {
		t.Fatalf("trades: got %d, want 1", len(out.Strategy.Trades))
	}
	trade := out.Strategy.Trades[0]
	if trade.EntryID != "L1" {
		t.Errorf("entryId: got %q, want \"L1\"", trade.EntryID)
	}
	if trade.Direction != "long" {
		t.Errorf("direction: got %q, want \"long\"", trade.Direction)
	}
	if math.Abs(trade.Profit-6.0) > 1e-9 {
		t.Errorf("trade.profit: got %.6f, want 6.0", trade.Profit)
	}
}

func TestTradeAccessorExecution_OpenTradeProperties(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
strategy("OpenTradeProps", overlay=true, default_qty_type=strategy.fixed, default_qty_value=1)

if bar_index == 0
    strategy.entry("L1", strategy.long)

openEntryPrice = strategy.opentrades.entry_price(0)
openSize = strategy.opentrades.size(0)

plot(openEntryPrice, "OpenEntryPrice")
plot(openSize, "OpenSize")
`

	bars := []map[string]interface{}{
		{"time": int64(1704067200), "open": 100.0, "high": 100.0, "low": 100.0, "close": 100.0, "volume": 1000.0},
		{"time": int64(1704070800), "open": 100.0, "high": 103.0, "low": 97.0, "close": 102.0, "volume": 1000.0},
		{"time": int64(1704074400), "open": 102.0, "high": 105.0, "low": 100.0, "close": 104.0, "volume": 1000.0},
		{"time": int64(1704078000), "open": 104.0, "high": 107.0, "low": 102.0, "close": 106.0, "volume": 1000.0},
	}

	exec := util.NewPineExecutor(t)
	raw := exec.ExecuteScriptWithCustomDataRaw(t, "trade-accessor-open-props", pineScript, bars)
	out := parseTradeAccessorOutput(t, raw)

	entryPrice, ok := lastPlotValue(out, "OpenEntryPrice")
	if !ok {
		t.Fatal("OpenEntryPrice: no non-NaN plot value")
	}
	if math.Abs(entryPrice-100.0) > 1e-9 {
		t.Errorf("OpenEntryPrice: got %.6f, want 100.0", entryPrice)
	}

	size, ok := lastPlotValue(out, "OpenSize")
	if !ok {
		t.Fatal("OpenSize: no non-NaN plot value")
	}
	if math.Abs(size-1.0) > 1e-9 {
		t.Errorf("OpenSize: got %.6f, want 1.0", size)
	}

	if len(out.Strategy.OpenTrades) != 1 {
		t.Fatalf("openTrades: got %d, want 1", len(out.Strategy.OpenTrades))
	}
	ot := out.Strategy.OpenTrades[0]
	if ot.EntryID != "L1" {
		t.Errorf("openTrade.entryId: got %q, want \"L1\"", ot.EntryID)
	}
	if ot.Direction != "long" {
		t.Errorf("openTrade.direction: got %q, want \"long\"", ot.Direction)
	}
	if math.Abs(ot.EntryPrice-100.0) > 1e-9 {
		t.Errorf("openTrade.entryPrice: got %.6f, want 100.0", ot.EntryPrice)
	}
	if math.Abs(ot.Size-1.0) > 1e-9 {
		t.Errorf("openTrade.size: got %.6f, want 1.0", ot.Size)
	}
}

// L1: entry=100, exit=105, profit=5. L2: entry=105, exit=112, profit=7.
// index 0 = oldest (L1), index 1 = newer (L2).
func TestTradeAccessorExecution_MultipleTradeIndexing(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
strategy("MultiTradeIndex", overlay=true, default_qty_type=strategy.fixed, default_qty_value=1)

if bar_index == 0
    strategy.entry("L1", strategy.long)
if bar_index == 2
    strategy.close("L1")
if bar_index == 4
    strategy.entry("L2", strategy.long)
if bar_index == 6
    strategy.close("L2")

trade0Profit = strategy.closedtrades.profit(0)
trade1Profit = strategy.closedtrades.profit(1)
trade0EntryPrice = strategy.closedtrades.entry_price(0)
trade1EntryPrice = strategy.closedtrades.entry_price(1)

plot(trade0Profit, "Trade0Profit")
plot(trade1Profit, "Trade1Profit")
plot(trade0EntryPrice, "Trade0EntryPrice")
plot(trade1EntryPrice, "Trade1EntryPrice")
`

	// Bar 0: L1 signal. Bar 1: L1 fills at 100. Bar 2: close L1 signal. Bar 3: L1 exits at 105 (profit=5).
	// Bar 4: L2 signal. Bar 5: L2 fills at 105. Bar 6: close L2 signal. Bar 7: L2 exits at 112 (profit=7).
	// Bar 8: verify both plots.
	bars := []map[string]interface{}{
		{"time": int64(1704067200), "open": 100.0, "high": 101.0, "low": 99.0, "close": 100.0, "volume": 1000.0},
		{"time": int64(1704070800), "open": 100.0, "high": 102.0, "low": 99.0, "close": 101.0, "volume": 1000.0},
		{"time": int64(1704074400), "open": 102.0, "high": 104.0, "low": 101.0, "close": 103.0, "volume": 1000.0},
		{"time": int64(1704078000), "open": 105.0, "high": 106.0, "low": 104.0, "close": 105.0, "volume": 1000.0},
		{"time": int64(1704081600), "open": 105.0, "high": 106.0, "low": 104.0, "close": 105.0, "volume": 1000.0},
		{"time": int64(1704085200), "open": 105.0, "high": 107.0, "low": 104.0, "close": 106.0, "volume": 1000.0},
		{"time": int64(1704088800), "open": 108.0, "high": 110.0, "low": 107.0, "close": 109.0, "volume": 1000.0},
		{"time": int64(1704092400), "open": 112.0, "high": 113.0, "low": 111.0, "close": 112.0, "volume": 1000.0},
		{"time": int64(1704096000), "open": 112.0, "high": 113.0, "low": 111.0, "close": 112.0, "volume": 1000.0},
	}

	exec := util.NewPineExecutor(t)
	raw := exec.ExecuteScriptWithCustomDataRaw(t, "trade-accessor-multi-index", pineScript, bars)
	out := parseTradeAccessorOutput(t, raw)

	t0Profit, ok := lastPlotValue(out, "Trade0Profit")
	if !ok {
		t.Fatal("Trade0Profit: no non-NaN value")
	}
	if math.Abs(t0Profit-5.0) > 1e-9 {
		t.Errorf("Trade0Profit (L1): got %.6f, want 5.0", t0Profit)
	}

	t1Profit, ok := lastPlotValue(out, "Trade1Profit")
	if !ok {
		t.Fatal("Trade1Profit: no non-NaN value")
	}
	if math.Abs(t1Profit-7.0) > 1e-9 {
		t.Errorf("Trade1Profit (L2): got %.6f, want 7.0", t1Profit)
	}

	t0Entry, ok := lastPlotValue(out, "Trade0EntryPrice")
	if !ok {
		t.Fatal("Trade0EntryPrice: no non-NaN value")
	}
	if math.Abs(t0Entry-100.0) > 1e-9 {
		t.Errorf("Trade0EntryPrice (L1): got %.6f, want 100.0", t0Entry)
	}

	t1Entry, ok := lastPlotValue(out, "Trade1EntryPrice")
	if !ok {
		t.Fatal("Trade1EntryPrice: no non-NaN value")
	}
	if math.Abs(t1Entry-105.0) > 1e-9 {
		t.Errorf("Trade1EntryPrice (L2): got %.6f, want 105.0", t1Entry)
	}

	if len(out.Strategy.Trades) != 2 {
		t.Fatalf("trades: got %d, want 2", len(out.Strategy.Trades))
	}
}

// Entry at open=100, held for bars with high reaching 114 and low reaching 96.
// MaxRunup = (114-100)*1 = 14.0, MaxDrawdown = (100-96)*1 = 4.0
func TestTradeAccessorExecution_DrawdownAndRunup(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
strategy("DrawdownRunup", overlay=true, default_qty_type=strategy.fixed, default_qty_value=1)

if bar_index == 0
    strategy.entry("L1", strategy.long)

if bar_index == 3
    strategy.close("L1")

maxDrawdown = strategy.closedtrades.max_drawdown(0)
maxRunup = strategy.closedtrades.max_runup(0)

plot(maxDrawdown, "MaxDrawdown")
plot(maxRunup, "MaxRunup")
`

	// Bar 0: signal. Bar 1: fills at open=100, OnBarMetrics(H=106,L=96) → runup=6, drawdown=4.
	// Bar 2: OnBarMetrics(H=112,L=99) → runup=12, drawdown stays 4.
	// Bar 3: close signal, OnBarMetrics(H=114,L=107) → runup=14, drawdown stays 4.
	// Bar 4: exits at open=110. Closed trade: MaxRunup=14, MaxDrawdown=4.
	bars := []map[string]interface{}{
		{"time": int64(1704067200), "open": 100.0, "high": 100.0, "low": 100.0, "close": 100.0, "volume": 1000.0},
		{"time": int64(1704070800), "open": 100.0, "high": 106.0, "low": 96.0, "close": 103.0, "volume": 1000.0},
		{"time": int64(1704074400), "open": 103.0, "high": 112.0, "low": 99.0, "close": 110.0, "volume": 1000.0},
		{"time": int64(1704078000), "open": 108.0, "high": 114.0, "low": 107.0, "close": 112.0, "volume": 1000.0},
		{"time": int64(1704081600), "open": 110.0, "high": 111.0, "low": 109.0, "close": 110.0, "volume": 1000.0},
	}

	exec := util.NewPineExecutor(t)
	raw := exec.ExecuteScriptWithCustomDataRaw(t, "trade-accessor-drawdown-runup", pineScript, bars)
	out := parseTradeAccessorOutput(t, raw)

	drawdown, ok := lastPlotValue(out, "MaxDrawdown")
	if !ok {
		t.Fatal("MaxDrawdown: no non-NaN plot value")
	}
	if math.Abs(drawdown-4.0) > 1e-9 {
		t.Errorf("MaxDrawdown: got %.6f, want 4.0", drawdown)
	}

	runup, ok := lastPlotValue(out, "MaxRunup")
	if !ok {
		t.Fatal("MaxRunup: no non-NaN plot value")
	}
	if math.Abs(runup-14.0) > 1e-9 {
		t.Errorf("MaxRunup: got %.6f, want 14.0", runup)
	}
}
