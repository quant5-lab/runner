package strategy

import (
	"testing"
)

func TestOrderManager(t *testing.T) {
	om := NewOrderManager()

	// Create order
	order := om.CreateOrder("long1", Long, 1.0, 0)
	if order.ID != "long1" || order.Direction != Long || order.Qty != 1.0 {
		t.Error("Order creation failed")
	}

	// Get pending orders (should be empty - same bar)
	pending := om.GetPendingOrders(0)
	if len(pending) != 0 {
		t.Error("Should not have pending orders on same bar")
	}

	// Get pending orders (next bar)
	pending = om.GetPendingOrders(1)
	if len(pending) != 1 {
		t.Error("Should have 1 pending order on next bar")
	}

	// Remove order
	om.RemoveOrder("long1")
	pending = om.GetPendingOrders(1)
	if len(pending) != 0 {
		t.Error("Order should be removed")
	}
}

func TestPositionTracker(t *testing.T) {
	pt := NewPositionTracker()

	// Open long position
	pt.UpdatePosition(10, 100, Long)
	if pt.GetPositionSize() != 10 {
		t.Errorf("Position size should be 10, got %.2f", pt.GetPositionSize())
	}
	if pt.GetAvgPrice() != 100 {
		t.Errorf("Avg price should be 100, got %.2f", pt.GetAvgPrice())
	}

	// Add to position
	pt.UpdatePosition(5, 110, Long)
	if pt.GetPositionSize() != 15 {
		t.Errorf("Position size should be 15, got %.2f", pt.GetPositionSize())
	}
	expectedAvg := (10*100 + 5*110) / 15.0
	if pt.GetAvgPrice() != expectedAvg {
		t.Errorf("Avg price should be %.2f, got %.2f", expectedAvg, pt.GetAvgPrice())
	}

	// Close position
	pt.UpdatePosition(15, 120, Short)
	if pt.GetPositionSize() != 0 {
		t.Errorf("Position size should be 0, got %.2f", pt.GetPositionSize())
	}
}

func TestTradeHistory(t *testing.T) {
	th := NewTradeHistory()

	// Add open trade
	th.AddOpenTrade(Trade{
		EntryID:    "long1",
		Direction:  Long,
		Size:       10,
		EntryPrice: 100,
		EntryBar:   0,
		EntryTime:  1000,
	})

	openTrades := th.GetOpenTrades()
	if len(openTrades) != 1 {
		t.Error("Should have 1 open trade")
	}

	// Close trade
	closedTrade := th.CloseTrade("long1", 110, 10, 2000)
	if closedTrade == nil {
		t.Fatal("Trade should be closed")
	}
	if closedTrade.Profit != 100 { // (110-100)*10
		t.Errorf("Profit should be 100, got %.2f", closedTrade.Profit)
	}

	openTrades = th.GetOpenTrades()
	if len(openTrades) != 0 {
		t.Error("Should have 0 open trades")
	}

	closedTrades := th.GetClosedTrades()
	if len(closedTrades) != 1 {
		t.Error("Should have 1 closed trade")
	}
}

func TestEquityCalculator(t *testing.T) {
	ec := NewEquityCalculator(10000)

	// Initial equity
	if ec.GetEquity(0) != 10000 {
		t.Error("Initial equity should be 10000")
	}

	// Update with closed trade
	ec.UpdateFromClosedTrade(Trade{Profit: 500})
	if ec.GetEquity(0) != 10500 {
		t.Errorf("Equity should be 10500, got %.2f", ec.GetEquity(0))
	}
	if ec.GetNetProfit() != 500 {
		t.Errorf("Net profit should be 500, got %.2f", ec.GetNetProfit())
	}

	// Include unrealized profit
	if ec.GetEquity(200) != 10700 {
		t.Errorf("Equity with unrealized should be 10700, got %.2f", ec.GetEquity(200))
	}
}

func TestStrategy(t *testing.T) {
	s := NewStrategy()
	s.Call("Test Strategy", 10000)

	// Place entry order
	err := s.Entry("long1", Long, 10)
	if err != nil {
		t.Fatal("Entry failed:", err)
	}

	// Process order on next bar
	s.OnBarUpdate(1, 100, 1000)

	// Check position
	if s.GetPositionSize() != 10 {
		t.Errorf("Position size should be 10, got %.2f", s.GetPositionSize())
	}
	if s.GetPositionAvgPrice() != 100 {
		t.Errorf("Avg price should be 100, got %.2f", s.GetPositionAvgPrice())
	}

	// Check open trades
	openTrades := s.tradeHistory.GetOpenTrades()
	if len(openTrades) != 1 {
		t.Error("Should have 1 open trade")
	}

	// Close position
	s.Close("long1", 110, 2000)

	// Check position closed
	if s.GetPositionSize() != 0 {
		t.Errorf("Position should be closed, got %.2f", s.GetPositionSize())
	}

	// Check equity
	expectedEquity := 10000.0 + 100.0 // Initial + profit (110-100)*10
	if s.GetEquity(110) != expectedEquity {
		t.Errorf("Equity should be %.2f, got %.2f", expectedEquity, s.GetEquity(110))
	}
}

func TestStrategyShort(t *testing.T) {
	s := NewStrategy()
	s.Call("Test Strategy", 10000)

	// Place short entry
	s.Entry("short1", Short, 5)
	s.OnBarUpdate(1, 100, 1000)

	// Check position (negative for short)
	if s.GetPositionSize() != -5 {
		t.Errorf("Position size should be -5, got %.2f", s.GetPositionSize())
	}

	// Close position with profit (price dropped)
	s.Close("short1", 90, 2000)

	// Check profit: (100-90)*5 = 50
	if s.GetNetProfit() != 50 {
		t.Errorf("Net profit should be 50, got %.2f", s.GetNetProfit())
	}
}

func TestStrategyCloseAll(t *testing.T) {
	s := NewStrategy()
	s.Call("Test Strategy", 10000)

	// Open multiple positions
	s.Entry("long1", Long, 10)
	s.Entry("long2", Long, 5)
	s.OnBarUpdate(1, 100, 1000)
	s.OnBarUpdate(2, 105, 2000)

	// Check open trades
	openTrades := s.tradeHistory.GetOpenTrades()
	if len(openTrades) != 2 {
		t.Errorf("Should have 2 open trades, got %d", len(openTrades))
	}

	// Close all
	s.CloseAll(110, 3000)

	// Check all closed
	openTrades = s.tradeHistory.GetOpenTrades()
	if len(openTrades) != 0 {
		t.Error("Should have 0 open trades")
	}

	closedTrades := s.tradeHistory.GetClosedTrades()
	if len(closedTrades) != 2 {
		t.Errorf("Should have 2 closed trades, got %d", len(closedTrades))
	}
}
