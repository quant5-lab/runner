package strategy

import (
	"math"
	"testing"
)

func newStrategyWithStep(initialCap, step float64) *Strategy {
	s := NewStrategy()
	s.Call("Test", initialCap)
	s.SetDefaultQty(100, QtyTypePercentOfEquity)
	s.SetQtyStep(step)
	return s
}

func openTrade0Size(t *testing.T, s *Strategy, entryPrice float64) float64 {
	t.Helper()
	if err := s.EntryWithDefaultQty("Buy", Long, ""); err != nil {
		t.Fatalf("EntryWithDefaultQty: %v", err)
	}
	s.OnBarUpdate(1, entryPrice, 1000)
	open := s.tradeHistory.GetOpenTrades()
	if len(open) != 1 {
		t.Fatalf("expected 1 open trade, got %d", len(open))
	}
	return open[0].Size
}

func TestStrategy_QtyStep_DefaultZeroStepAllowsFractionalSizes(t *testing.T) {
	// Without SetQtyStep the default zero step must not modify any qty — fractional
	// sizes are fully preserved. This is the correct behaviour for exchanges where
	// the lot size is unknown (e.g. Binance).
	s := NewStrategy()
	s.Call("Test", 10000)
	s.SetDefaultQty(100, QtyTypePercentOfEquity)

	// 10000 * 1.0 / 300 = 33.333…
	size := openTrade0Size(t, s, 300)
	if size == math.Floor(size) {
		t.Errorf("size = %v: expected fractional when qtyStep = 0", size)
	}
}

func TestStrategy_QtyStep_WholeStepFloorsPercentOfEquityEntry(t *testing.T) {
	// 10000 * 1.0 / 300 = 33.333… → floored to 33
	s := newStrategyWithStep(10000, 1)
	size := openTrade0Size(t, s, 300)
	if size != 33 {
		t.Errorf("size = %v, want 33 (floor of 33.333…)", size)
	}
}

func TestStrategy_QtyStep_WholeStepFloorsCashEntry(t *testing.T) {
	// 7000 / 300 = 23.333… → floored to 23
	s := NewStrategy()
	s.Call("Test", 10000)
	s.SetDefaultQty(7000, QtyTypeCash)
	s.SetQtyStep(1)
	size := openTrade0Size(t, s, 300)
	if size != 23 {
		t.Errorf("size = %v, want 23 (floor of 23.333…)", size)
	}
}

func TestStrategy_QtyStep_WholeStepFloorsExplicitFractionalEntry(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)
	s.SetQtyStep(1)
	if err := s.Entry("Buy", Long, 4.9, ""); err != nil {
		t.Fatalf("Entry: %v", err)
	}
	s.OnBarUpdate(1, 100, 1000)
	open := s.tradeHistory.GetOpenTrades()
	if len(open) != 1 {
		t.Fatalf("expected 1 open trade, got %d", len(open))
	}
	if open[0].Size != 4 {
		t.Errorf("size = %v, want 4 (floor of 4.9)", open[0].Size)
	}
}

func TestStrategy_QtyStep_WholeStepFloorsNetOrder(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)
	s.SetQtyStep(1)
	if err := s.Order("NetBuy", Long, 7.8, ""); err != nil {
		t.Fatalf("Order: %v", err)
	}
	s.OnBarUpdate(1, 100, 1000)
	open := s.tradeHistory.GetOpenTrades()
	if len(open) != 1 {
		t.Fatalf("expected 1 open trade, got %d", len(open))
	}
	if open[0].Size != 7 {
		t.Errorf("size = %v, want 7 (floor of 7.8)", open[0].Size)
	}
}

func TestStrategy_QtyStep_ExactMultipleIsStoredUnchanged(t *testing.T) {
	// 10000 / 200 = 50.0 — boundary: exactly representable, no drift risk
	s := NewStrategy()
	s.Call("Test", 10000)
	s.SetDefaultQty(10000, QtyTypeCash)
	s.SetQtyStep(1)
	size := openTrade0Size(t, s, 200)
	if size != 50 {
		t.Errorf("size = %v, want 50", size)
	}
}

func TestStrategy_QtyStep_SubUnitStepFloors(t *testing.T) {
	// step=0.5: 10000/300 = 33.333… → nearest 0.5 downward = 33.0
	s := NewStrategy()
	s.Call("Test", 10000)
	s.SetDefaultQty(100, QtyTypePercentOfEquity)
	s.SetQtyStep(0.5)
	size := openTrade0Size(t, s, 300)
	if size != 33.0 {
		t.Errorf("size = %v, want 33.0 (floor of 33.333… to 0.5 step)", size)
	}
}

func TestStrategy_QtyStep_FlooredSizePropagatesIntoProfit(t *testing.T) {
	s := newStrategyWithStep(10000, 1)
	openTrade0Size(t, s, 300)
	s.Close("Buy", 330, 2000, "")
	s.OnBarUpdate(2, 330, 2000)
	closed := s.tradeHistory.GetClosedTrades()
	if len(closed) != 1 {
		t.Fatalf("expected 1 closed trade, got %d", len(closed))
	}
	const wantProfit = 33 * 30.0 // 33 floored shares × 30 point move
	if math.Abs(closed[0].Profit-wantProfit) > 1e-9 {
		t.Errorf("profit = %.6f, want %.6f", closed[0].Profit, wantProfit)
	}
}

func TestStrategy_QtyStep_DeferredEntryAtReversalFillAlsoFloors(t *testing.T) {
	// EntryWithDefaultQty after a reversal: qty is computed at the fill bar open
	// (post-reversal equity). The deferred qty must also be floored.
	//
	// Setup: long 1 at 100, signal short at bar 2, fill at bar 3 open=130.
	// Post-reversal equity = 10000 + 1*(130-100) = 10030.
	// Unrounded qty = 10030/130 = 77.153… → floored to 77.
	s := NewStrategy()
	s.Call("Test", 10000)
	s.SetDefaultQty(100, QtyTypePercentOfEquity)
	s.SetQtyStep(1)

	s.Entry("Long", Long, 1, "")
	s.OnBarUpdate(1, 100, 1000)
	s.OnBarUpdate(2, 120, 2000)

	if err := s.EntryWithDefaultQty("Short", Short, ""); err != nil {
		t.Fatalf("EntryWithDefaultQty: %v", err)
	}
	s.OnBarUpdate(3, 130, 3000)

	open := s.tradeHistory.GetOpenTrades()
	if len(open) != 1 {
		t.Fatalf("expected 1 open trade after reversal, got %d", len(open))
	}
	if open[0].Direction != Short {
		t.Errorf("direction = %s, want Short", open[0].Direction)
	}
	postReversalEquity := 10000.0 + 1*(130-100)
	wantSize := math.Floor(postReversalEquity / 130)
	if open[0].Size != wantSize {
		t.Errorf("size = %v, want %v (floor of post-reversal %.6f/130)", open[0].Size, wantSize, postReversalEquity)
	}
}

func TestStrategy_QtyStep_SetterReplacesPreviousStep(t *testing.T) {
	// 10000/300 = 33.333… → floor with step=1 = 33
	s := NewStrategy()
	s.Call("Test", 10000)
	s.SetDefaultQty(100, QtyTypePercentOfEquity)
	s.SetQtyStep(0.5)
	s.SetQtyStep(1)
	size := openTrade0Size(t, s, 300)
	if size != 33 {
		t.Errorf("size = %v, want 33 after step was updated from 0.5 to 1", size)
	}
}

func TestStrategy_QtyStep_QtyZeroRemainsZero(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)
	s.SetQtyStep(1)
	if err := s.Entry("Buy", Long, 0.3, ""); err != nil {
		t.Fatalf("Entry: %v", err)
	}
	s.OnBarUpdate(1, 100, 1000)
	open := s.tradeHistory.GetOpenTrades()
	if len(open) != 1 {
		t.Fatalf("expected 1 open trade, got %d", len(open))
	}
	if open[0].Size != 0 {
		t.Errorf("size = %v, want 0 (fractional qty below one step)", open[0].Size)
	}
}
