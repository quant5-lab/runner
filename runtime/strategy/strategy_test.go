package strategy

import (
	"testing"
)

func TestOrderManager(t *testing.T) {
	om := NewOrderManager()

	// Create order
	order := om.CreateOrder("long1", Long, 1.0, 0, "")
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

/* TestOrderManagerWithComment verifies comment field propagation */
func TestOrderManagerWithComment(t *testing.T) {
	om := NewOrderManager()

	/* Create order with entry comment */
	order := om.CreateOrder("long1", Long, 1.0, 0, "Buy signal")
	if order.EntryComment != "Buy signal" {
		t.Errorf("Expected comment 'Buy signal', got %q", order.EntryComment)
	}

	/* Verify comment persists through retrieval */
	pending := om.GetPendingOrders(1)
	if len(pending) != 1 {
		t.Fatal("Should have 1 pending order")
	}
	if pending[0].EntryComment != "Buy signal" {
		t.Errorf("Expected comment 'Buy signal', got %q", pending[0].EntryComment)
	}

	/* Create order without comment (empty string default) */
	order2 := om.CreateOrder("long2", Long, 2.0, 0, "")
	if order2.EntryComment != "" {
		t.Errorf("Expected empty comment, got %q", order2.EntryComment)
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
	closedTrade := th.CloseTrade("long1", 110, 10, 2000, "")
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

/* TestTradeHistoryWithComment verifies comment flow through trade lifecycle */
func TestTradeHistoryWithComment(t *testing.T) {
	th := NewTradeHistory()

	/* Add open trade with entry comment */
	th.AddOpenTrade(Trade{
		EntryID:      "long1",
		Direction:    Long,
		Size:         10,
		EntryPrice:   100,
		EntryBar:     0,
		EntryTime:    1000,
		EntryComment: "Breakout entry",
	})

	openTrades := th.GetOpenTrades()
	if len(openTrades) != 1 {
		t.Fatal("Should have 1 open trade")
	}
	if openTrades[0].EntryComment != "Breakout entry" {
		t.Errorf("Expected entry comment 'Breakout entry', got %q", openTrades[0].EntryComment)
	}

	/* Close trade with exit comment */
	closedTrade := th.CloseTrade("long1", 110, 10, 2000, "Take profit")
	if closedTrade == nil {
		t.Fatal("Trade should be closed")
	}
	if closedTrade.EntryComment != "Breakout entry" {
		t.Errorf("Expected entry comment preserved, got %q", closedTrade.EntryComment)
	}
	if closedTrade.ExitComment != "Take profit" {
		t.Errorf("Expected exit comment 'Take profit', got %q", closedTrade.ExitComment)
	}

	/* Close trade without exit comment */
	th.AddOpenTrade(Trade{
		EntryID:      "long2",
		Direction:    Long,
		Size:         5,
		EntryPrice:   105,
		EntryBar:     2,
		EntryTime:    3000,
		EntryComment: "Second entry",
	})
	closedTrade2 := th.CloseTrade("long2", 108, 3, 4000, "")
	if closedTrade2.ExitComment != "" {
		t.Errorf("Expected empty exit comment, got %q", closedTrade2.ExitComment)
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
	s.CallWithPyramiding("Test Strategy", 10000, 0)

	// Place entry order
	err := s.Entry("long1", Long, 10, "")
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
	s.Close("long1", 110, 2000, "")
	s.OnBarUpdate(2, 110, 2000)

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

/* TestEquityCurrentPriceTracking verifies price tracking across bar updates */
func TestEquityCurrentPriceTracking(t *testing.T) {
	tests := []struct {
		name          string
		initialCap    float64
		barUpdates    []float64
		expectedPrice float64
		desc          string
	}{
		{"single_bar", 10000, []float64{100}, 100, "first bar sets price"},
		{"multiple_bars", 10000, []float64{100, 105, 110}, 110, "last bar price retained"},
		{"price_decline", 10000, []float64{100, 95, 90}, 90, "declining price tracked"},
		{"price_volatility", 10000, []float64{100, 110, 95, 105}, 105, "volatile price tracked"},
		{"zero_price", 10000, []float64{0}, 0, "zero price handled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.CallWithPyramiding("Test", tt.initialCap, 0)

			for i, price := range tt.barUpdates {
				s.OnBarUpdate(i, price, int64(1000+i))
			}

			/* Verify Equity() uses the last updated price */
			expectedEquity := s.GetEquity(tt.expectedPrice)
			if s.Equity() != expectedEquity {
				t.Errorf("%s: Equity() should use price %.2f, got equity %.2f vs expected %.2f",
					tt.desc, tt.expectedPrice, s.Equity(), expectedEquity)
			}
		})
	}
}

/* TestEquityUnrealizedProfitCalculation verifies unrealized P&L in equity */
func TestEquityUnrealizedProfitCalculation(t *testing.T) {
	tests := []struct {
		name         string
		initialCap   float64
		direction    string
		entryPrice   float64
		currentPrice float64
		qty          float64
		wantEquity   float64
		desc         string
	}{
		{"long_profit", 10000, Long, 100, 110, 10, 10100, "long with 100 profit"},
		{"long_loss", 10000, Long, 100, 90, 10, 9900, "long with 100 loss"},
		{"short_profit", 10000, Short, 100, 90, 10, 10100, "short with 100 profit"},
		{"short_loss", 10000, Short, 100, 110, 10, 9900, "short with 100 loss"},
		{"long_breakeven", 10000, Long, 100, 100, 10, 10000, "long at entry price"},
		{"short_breakeven", 10000, Short, 100, 100, 10, 10000, "short at entry price"},
		{"large_position", 10000, Long, 50, 60, 100, 11000, "large qty position"},
		{"small_position", 10000, Long, 100, 101, 1, 10001, "fractional profit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.CallWithPyramiding("Test", tt.initialCap, 0)

			err := s.Entry("trade1", tt.direction, tt.qty, "")
			if err != nil {
				t.Fatal("Entry failed:", err)
			}

			s.OnBarUpdate(1, tt.entryPrice, 1001)
			s.OnBarUpdate(2, tt.currentPrice, 1002)

			equity := s.Equity()
			if equity != tt.wantEquity {
				t.Errorf("%s: expected equity=%.2f, got %.2f",
					tt.desc, tt.wantEquity, equity)
			}
		})
	}
}

/* TestEquityMultiplePositions verifies equity with multiple open trades */
func TestEquityMultiplePositions(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test", 10000, 1) // pyramiding=1 allows 2 positions

	/* Open first long position */
	s.Entry("long1", Long, 10, "")
	s.OnBarUpdate(1, 100, 1001)

	/* Open second long position */
	s.Entry("long2", Long, 5, "")
	s.OnBarUpdate(2, 105, 1002)

	/* Update price: long1 +150 profit, long2 +25 profit */
	s.OnBarUpdate(3, 120, 1003)

	expectedEquity := 10000.0 + (120-100)*10 + (120-105)*5
	if s.Equity() != expectedEquity {
		t.Errorf("Expected equity %.2f with multiple positions, got %.2f",
			expectedEquity, s.Equity())
	}
}

/* TestEquityAfterClosedTrade verifies realized profit in equity */
func TestEquityAfterClosedTrade(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test", 10000, 0)

	/* Open and close first trade with profit */
	s.Entry("long1", Long, 10, "")
	s.OnBarUpdate(1, 100, 1001)
	s.Close("long1", 110, 1002, "")
	s.OnBarUpdate(2, 110, 1002)

	/* Equity should reflect realized profit */
	expectedEquity := 10000.0 + 100.0 /* (110-100)*10 */

	if s.Equity() != expectedEquity {
		t.Errorf("Expected equity %.2f after closed trade, got %.2f",
			expectedEquity, s.Equity())
	}

	/* Open new trade - equity should include both realized + unrealized */
	s.Entry("long2", Long, 5, "")
	s.OnBarUpdate(3, 110, 1003)
	s.OnBarUpdate(4, 120, 1004)

	expectedEquity = 10000.0 + 100.0 + (120-110)*5
	if s.Equity() != expectedEquity {
		t.Errorf("Expected equity %.2f with realized+unrealized, got %.2f",
			expectedEquity, s.Equity())
	}
}

/* TestEquityConsistencyWithGetEquity verifies wrapper delegates correctly */
func TestEquityConsistencyWithGetEquity(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test", 10000, 0)

	priceSequence := []float64{100, 105, 110, 95, 100, 120}

	for i, price := range priceSequence {
		s.OnBarUpdate(i, price, int64(1000+i))

		equityWrapper := s.Equity()
		equityDirect := s.GetEquity(price)

		if equityWrapper != equityDirect {
			t.Errorf("Bar %d (price=%.2f): Equity()=%.2f != GetEquity()=%.2f",
				i, price, equityWrapper, equityDirect)
		}
	}
}

/* TestEquityBeforeInitialization verifies behavior before Call() */
func TestEquityBeforeInitialization(t *testing.T) {
	s := NewStrategy()

	/* Equity before Call() uses default capital from NewEquityCalculator */
	equity := s.Equity()
	if equity != 10000 {
		t.Errorf("Expected equity=10000 (default capital), got %.2f", equity)
	}
}

/* TestEquityWithNoBarUpdates verifies behavior without price updates */
func TestEquityWithNoBarUpdates(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test", 10000, 0)

	/* Without OnBarUpdate, currentPrice is 0 */
	equity := s.Equity()
	if equity != 10000 {
		t.Errorf("Expected equity=10000 (initial capital, no positions), got %.2f", equity)
	}
}

/* TestStrategyEntryComment verifies entry comment propagation through full cycle */
func TestStrategyEntryComment(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test Strategy", 10000, 0)

	/* Place entry with comment */
	err := s.Entry("long1", Long, 10, "Buy on MA cross")
	if err != nil {
		t.Fatal("Entry failed:", err)
	}

	/* Process order - comment should propagate to Trade */
	s.OnBarUpdate(1, 100, 1000)

	openTrades := s.tradeHistory.GetOpenTrades()
	if len(openTrades) != 1 {
		t.Fatal("Should have 1 open trade")
	}
	if openTrades[0].EntryComment != "Buy on MA cross" {
		t.Errorf("Expected entry comment 'Buy on MA cross', got %q", openTrades[0].EntryComment)
	}

	/* Close and verify comment persists */
	s.Close("long1", 110, 2000, "Target reached")
	s.OnBarUpdate(2, 110, 2000)

	closedTrades := s.tradeHistory.GetClosedTrades()
	if len(closedTrades) != 1 {
		t.Fatal("Should have 1 closed trade")
	}
	if closedTrades[0].EntryComment != "Buy on MA cross" {
		t.Errorf("Entry comment should persist, got %q", closedTrades[0].EntryComment)
	}
	if closedTrades[0].ExitComment != "Target reached" {
		t.Errorf("Expected exit comment 'Target reached', got %q", closedTrades[0].ExitComment)
	}
}

func TestStrategyExitComment(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test Strategy", 10000, 2)

	s.Entry("long1", Long, 10, "Entry 1")
	s.OnBarUpdate(1, 100, 1000)
	s.Close("long1", 105, 2000, "Manual close")
	s.OnBarUpdate(2, 105, 2000)

	closedTrades := s.tradeHistory.GetClosedTrades()
	if len(closedTrades) != 1 {
		t.Fatal("Should have 1 closed trade")
	}
	if closedTrades[0].ExitComment != "Manual close" {
		t.Errorf("Expected 'Manual close', got %q", closedTrades[0].ExitComment)
	}

	/* Test CloseAll with comment */
	s.Entry("long2", Long, 5, "Entry 2")
	s.OnBarUpdate(2, 110, 3000)
	s.Entry("long3", Long, 3, "Entry 3")
	s.OnBarUpdate(3, 115, 4000)
	s.CloseAll(120, 5000, "Close all positions")
	s.OnBarUpdate(4, 120, 5000)

	closedTrades = s.tradeHistory.GetClosedTrades()
	if len(closedTrades) != 3 {
		t.Fatalf("Should have 3 closed trades, got %d", len(closedTrades))
	}
	/* Both new trades should have CloseAll comment */
	if closedTrades[1].ExitComment != "Close all positions" {
		t.Errorf("Expected 'Close all positions', got %q", closedTrades[1].ExitComment)
	}
	if closedTrades[2].ExitComment != "Close all positions" {
		t.Errorf("Expected 'Close all positions', got %q", closedTrades[2].ExitComment)
	}

	/* Test Exit with comment */
	s.Entry("long4", Long, 8, "Entry 4")
	s.OnBarUpdate(4, 125, 6000)
	s.OnBarUpdate(5, 130, 7000)
	s.Exit("exit1", "long4", 130, 7000, "Stop loss hit")
	s.OnBarUpdate(6, 130, 7000)

	closedTrades = s.tradeHistory.GetClosedTrades()
	if len(closedTrades) != 4 {
		t.Fatalf("Should have 4 closed trades, got %d", len(closedTrades))
	}
	if closedTrades[3].ExitComment != "Stop loss hit" {
		t.Errorf("Expected 'Stop loss hit', got %q", closedTrades[3].ExitComment)
	}
}

/* TestStrategyMixedComments verifies behavior with mixed comment/no-comment trades */
func TestStrategyMixedComments(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test Strategy", 10000, 1) // pyramiding=1 allows 2 positions

	/* Entry with comment */
	s.Entry("long1", Long, 10, "Signal A")
	s.OnBarUpdate(1, 100, 1000)

	/* Entry without comment */
	s.Entry("long2", Long, 5, "")
	s.OnBarUpdate(2, 105, 2000)

	/* Close with comment */
	s.Close("long1", 110, 3000, "Exit A")
	s.OnBarUpdate(3, 110, 3000)

	/* Close without comment */
	s.Close("long2", 108, 4000, "")
	s.OnBarUpdate(4, 108, 4000)

	closedTrades := s.tradeHistory.GetClosedTrades()
	if len(closedTrades) != 2 {
		t.Fatalf("Should have 2 closed trades, got %d", len(closedTrades))
	}

	/* Verify first trade has comments */
	if closedTrades[0].EntryComment != "Signal A" {
		t.Errorf("Expected 'Signal A', got %q", closedTrades[0].EntryComment)
	}
	if closedTrades[0].ExitComment != "Exit A" {
		t.Errorf("Expected 'Exit A', got %q", closedTrades[0].ExitComment)
	}

	/* Verify second trade has empty comments */
	if closedTrades[1].EntryComment != "" {
		t.Errorf("Expected empty entry comment, got %q", closedTrades[1].EntryComment)
	}
	if closedTrades[1].ExitComment != "" {
		t.Errorf("Expected empty exit comment, got %q", closedTrades[1].ExitComment)
	}
}

func TestStrategyShort(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test Strategy", 10000, 0)

	// Place short entry
	s.Entry("short1", Short, 5, "")
	s.OnBarUpdate(1, 100, 1000)

	// Check position (negative for short)
	if s.GetPositionSize() != -5 {
		t.Errorf("Position size should be -5, got %.2f", s.GetPositionSize())
	}

	// Close position with profit (price dropped)
	s.Close("short1", 90, 2000, "")
	s.OnBarUpdate(2, 90, 2000)

	// Check profit: (100-90)*5 = 50
	if s.GetNetProfit() != 50 {
		t.Errorf("Net profit should be 50, got %.2f", s.GetNetProfit())
	}
}

func TestStrategyCloseAll(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test Strategy", 10000, 1)

	s.Entry("long1", Long, 10, "")
	s.Entry("long2", Long, 5, "")
	s.OnBarUpdate(1, 100, 1000)
	s.OnBarUpdate(2, 105, 2000)

	// Check open trades
	openTrades := s.tradeHistory.GetOpenTrades()
	if len(openTrades) != 2 {
		t.Errorf("Should have 2 open trades, got %d", len(openTrades))
	}

	// Close all
	s.CloseAll(110, 3000, "")
	s.OnBarUpdate(3, 110, 3000)

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
