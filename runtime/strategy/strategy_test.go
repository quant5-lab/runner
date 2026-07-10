package strategy

import (
	"fmt"
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

/*
	TestOrderManager_ClearAll verifies ClearAll empties the queue for all initial sizes

and that the queue remains usable for new orders after clearing.
*/
func TestOrderManager_ClearAll(t *testing.T) {
	tests := []struct {
		name       string
		orderCount int
	}{
		{"empty queue is no-op", 0},
		{"single order removed", 1},
		{"multiple orders all removed", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			om := NewOrderManager()
			for i := 0; i < tt.orderCount; i++ {
				om.CreateOrder("o", Long, 10, 0, "")
			}
			om.ClearAll()

			if pending := om.GetPendingOrders(999); len(pending) != 0 {
				t.Errorf("expected empty queue after ClearAll, got %d orders", len(pending))
			}

			om.CreateOrder("post-clear", Long, 5, 0, "")
			if pending := om.GetPendingOrders(999); len(pending) != 1 {
				t.Errorf("expected 1 order after re-use, got %d", len(pending))
			}
		})
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
	closedTrade := th.CloseTrade("long1", "", 110, 10, 2000, "", 0.0)
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
	closedTrade := th.CloseTrade("long1", "", 110, 10, 2000, "Take profit", 0.0)
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
	closedTrade2 := th.CloseTrade("long2", "", 108, 3, 4000, "", 0.0)
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
	s.CallWithPyramiding("Test", 10000, 2) // pyramiding=2 allows 2 simultaneous same-direction entries

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

/*
TestExit_ClosesAllMatchingEntriesById verifies that Exit(exitID, fromEntry) exits every open
trade carrying the given fromEntry ID, mirroring the Close semantics for the strategy.exit path.
*/
func TestExit_ClosesAllMatchingEntriesById(t *testing.T) {
	type entry struct {
		id  string
		qty float64
	}
	tests := []struct {
		name            string
		pyramiding      int
		entries         []entry
		exitID          string
		fromEntry       string
		entryPrice      float64
		exitPrice       float64
		wantClosedCount int
		wantOpenCount   int
	}{
		{
			name:       "two_same_fromentry_both_closed",
			pyramiding: 2,
			entries:    []entry{{"e", 10}, {"e", 10}},
			exitID:     "exit", fromEntry: "e",
			entryPrice: 100, exitPrice: 110,
			wantClosedCount: 2, wantOpenCount: 0,
		},
		{
			name:       "mixed_fromentry_only_matching_closed",
			pyramiding: 3,
			entries:    []entry{{"e", 10}, {"e", 10}, {"other", 5}},
			exitID:     "exit", fromEntry: "e",
			entryPrice: 100, exitPrice: 110,
			wantClosedCount: 2, wantOpenCount: 1,
		},
		{
			name:       "nonexistent_fromentry_is_noop",
			pyramiding: 2,
			entries:    []entry{{"e", 10}, {"e", 10}},
			exitID:     "exit", fromEntry: "no_such_id",
			entryPrice: 100, exitPrice: 110,
			wantClosedCount: 0, wantOpenCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.CallWithPyramiding("Test", 10000, tt.pyramiding)

			bar, ts := 0, int64(1000)
			for _, e := range tt.entries {
				s.Entry(e.id, Long, e.qty, "")
				bar++
				ts += 1000
				s.OnBarUpdate(bar, tt.entryPrice, ts)
			}

			s.Exit(tt.exitID, tt.fromEntry, tt.exitPrice, ts, "")
			bar++
			ts += 1000
			s.OnBarUpdate(bar, tt.exitPrice, ts)

			closed := s.GetTradeHistory().GetClosedTrades()
			if len(closed) != tt.wantClosedCount {
				t.Errorf("closed count: got %d, want %d", len(closed), tt.wantClosedCount)
			}
			open := s.GetTradeHistory().GetOpenTrades()
			if len(open) != tt.wantOpenCount {
				t.Errorf("open count: got %d, want %d", len(open), tt.wantOpenCount)
			}
		})
	}
}

/* TestStrategyMixedComments verifies behavior with mixed comment/no-comment trades */
func TestStrategyMixedComments(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test Strategy", 10000, 2) // pyramiding=2 allows 2 simultaneous same-direction entries

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
	s.CallWithPyramiding("Test Strategy", 10000, 2)

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

/*
TestCloseAll_ClosesAllOpenTradesRegardlessOfEntryId verifies that CloseAll exits every open
trade irrespective of entry ID, covering N>2 positions, same-ID pyramiding, short direction,
variable lot sizes, and the no-op when no trades are open.
*/
func TestCloseAll_ClosesAllOpenTradesRegardlessOfEntryId(t *testing.T) {
	type entry struct {
		id        string
		direction string
		qty       float64
	}
	tests := []struct {
		name             string
		direction        string
		entries          []entry
		entryPrice       float64
		exitPrice        float64
		wantClosedCount  int
		wantPositionSize float64
		wantNetProfit    float64
	}{
		{
			name:       "three_unique_ids_long",
			direction:  Long,
			entries:    []entry{{"a", Long, 10}, {"b", Long, 10}, {"c", Long, 10}},
			entryPrice: 100, exitPrice: 110,
			wantClosedCount: 3, wantPositionSize: 0, wantNetProfit: 300, // (110-100)*10 * 3
		},
		{
			name:       "four_same_id_long",
			direction:  Long,
			entries:    []entry{{"e", Long, 10}, {"e", Long, 10}, {"e", Long, 10}, {"e", Long, 10}},
			entryPrice: 100, exitPrice: 110,
			wantClosedCount: 4, wantPositionSize: 0, wantNetProfit: 400,
		},
		{
			name:       "three_unique_ids_short",
			direction:  Short,
			entries:    []entry{{"a", Short, 5}, {"b", Short, 5}, {"c", Short, 5}},
			entryPrice: 100, exitPrice: 90,
			wantClosedCount: 3, wantPositionSize: 0, wantNetProfit: 150, // (100-90)*5 * 3
		},
		{
			name:       "variable_lot_sizes_profit_per_trade",
			direction:  Long,
			entries:    []entry{{"a", Long, 10}, {"b", Long, 20}, {"c", Long, 5}},
			entryPrice: 100, exitPrice: 110,
			wantClosedCount: 3, wantPositionSize: 0, wantNetProfit: 350, // (110-100)*(10+20+5)
		},
		{
			name:       "no_open_trades_is_noop",
			direction:  Long,
			entries:    nil,
			entryPrice: 100, exitPrice: 110,
			wantClosedCount: 0, wantPositionSize: 0, wantNetProfit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pyramiding := len(tt.entries)
			if pyramiding == 0 {
				pyramiding = 1
			}
			s := NewStrategy()
			s.CallWithPyramiding("Test", 10000, pyramiding)

			bar, ts := 0, int64(1000)
			for _, e := range tt.entries {
				s.Entry(e.id, e.direction, e.qty, "")
				bar++
				ts += 1000
				s.OnBarUpdate(bar, tt.entryPrice, ts)
			}

			s.CloseAll(tt.exitPrice, ts, "")
			bar++
			ts += 1000
			s.OnBarUpdate(bar, tt.exitPrice, ts)

			closed := s.GetTradeHistory().GetClosedTrades()
			if len(closed) != tt.wantClosedCount {
				t.Errorf("closed count: got %d, want %d", len(closed), tt.wantClosedCount)
			}
			if len(s.GetTradeHistory().GetOpenTrades()) != 0 {
				t.Errorf("open count: got %d, want 0", len(s.GetTradeHistory().GetOpenTrades()))
			}
			if s.GetPositionSize() != tt.wantPositionSize {
				t.Errorf("position size: got %.2f, want %.2f", s.GetPositionSize(), tt.wantPositionSize)
			}
			if s.GetNetProfit() != tt.wantNetProfit {
				t.Errorf("net profit: got %.2f, want %.2f", s.GetNetProfit(), tt.wantNetProfit)
			}
		})
	}
}

func TestStrategyGetInitialCapital(t *testing.T) {
	tests := []struct {
		name            string
		initialCapital  float64
		expectedCapital float64
	}{
		{
			name:            "standard_capital",
			initialCapital:  10000,
			expectedCapital: 10000,
		},
		{
			name:            "large_capital",
			initialCapital:  1000000,
			expectedCapital: 1000000,
		},
		{
			name:            "small_capital",
			initialCapital:  100,
			expectedCapital: 100,
		},
		{
			name:            "fractional_capital",
			initialCapital:  12345.67,
			expectedCapital: 12345.67,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.Call("Test", tt.initialCapital)

			result := s.GetInitialCapital()
			if result != tt.expectedCapital {
				t.Errorf("Expected %.2f, got %.2f", tt.expectedCapital, result)
			}
		})
	}
}

func TestStrategyGrossProfitLoss(t *testing.T) {
	tests := []struct {
		name                string
		trades              []struct{ dir, qty, entryPrice, exitPrice float64 }
		expectedGrossProfit float64
		expectedGrossLoss   float64
		expectedWinCount    int
		expectedLossCount   int
		expectedEvenCount   int
	}{
		{
			name: "mixed_long_short_trades",
			trades: []struct{ dir, qty, entryPrice, exitPrice float64 }{
				{1, 10, 100, 110}, // +100
				{1, 5, 105, 100},  // -25
				{-1, 10, 100, 95}, // +50
			},
			expectedGrossProfit: 150,
			expectedGrossLoss:   -25,
			expectedWinCount:    2,
			expectedLossCount:   1,
			expectedEvenCount:   0,
		},
		{
			name: "all_profitable_trades",
			trades: []struct{ dir, qty, entryPrice, exitPrice float64 }{
				{1, 10, 100, 110}, // +100
				{1, 5, 100, 120},  // +100
				{-1, 10, 100, 90}, // +100
			},
			expectedGrossProfit: 300,
			expectedGrossLoss:   0,
			expectedWinCount:    3,
			expectedLossCount:   0,
			expectedEvenCount:   0,
		},
		{
			name: "all_losing_trades",
			trades: []struct{ dir, qty, entryPrice, exitPrice float64 }{
				{1, 10, 100, 90},   // -100
				{1, 5, 100, 95},    // -25
				{-1, 10, 100, 110}, // -100
			},
			expectedGrossProfit: 0,
			expectedGrossLoss:   -225,
			expectedWinCount:    0,
			expectedLossCount:   3,
			expectedEvenCount:   0,
		},
		{
			name: "breakeven_trades",
			trades: []struct{ dir, qty, entryPrice, exitPrice float64 }{
				{1, 10, 100, 100}, // 0
				{-1, 5, 100, 100}, // 0
			},
			expectedGrossProfit: 0,
			expectedGrossLoss:   0,
			expectedWinCount:    0,
			expectedLossCount:   0,
			expectedEvenCount:   2,
		},
		{
			name: "large_position_sizes",
			trades: []struct{ dir, qty, entryPrice, exitPrice float64 }{
				{1, 1000, 100, 101}, // +1000
				{1, 500, 100, 98},   // -1000
			},
			expectedGrossProfit: 1000,
			expectedGrossLoss:   -1000,
			expectedWinCount:    1,
			expectedLossCount:   1,
			expectedEvenCount:   0,
		},
		{
			name: "fractional_quantities",
			trades: []struct{ dir, qty, entryPrice, exitPrice float64 }{
				{1, 0.5, 100, 110}, // +5
				{1, 0.25, 100, 90}, // -2.5
			},
			expectedGrossProfit: 5,
			expectedGrossLoss:   -2.5,
			expectedWinCount:    1,
			expectedLossCount:   1,
			expectedEvenCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.Call("Test", 10000)

			barIndex := 1
			timestamp := int64(1000)

			for i, trade := range tt.trades {
				entryID := "trade" + string(rune('0'+i))
				var direction string
				if trade.dir > 0 {
					direction = Long
				} else {
					direction = Short
				}

				s.Entry(entryID, direction, trade.qty, "")
				s.OnBarUpdate(barIndex, trade.entryPrice, timestamp)
				barIndex++
				timestamp += 1000

				s.Close(entryID, trade.exitPrice, timestamp, "")
				s.OnBarUpdate(barIndex, trade.exitPrice, timestamp)
				barIndex++
				timestamp += 1000
			}

			grossProfit := s.GetGrossProfit()
			if grossProfit != tt.expectedGrossProfit {
				t.Errorf("GrossProfit: expected %.2f, got %.2f", tt.expectedGrossProfit, grossProfit)
			}

			grossLoss := s.GetGrossLoss()
			if grossLoss != tt.expectedGrossLoss {
				t.Errorf("GrossLoss: expected %.2f, got %.2f", tt.expectedGrossLoss, grossLoss)
			}

			winCount := s.GetWinningTradesCount()
			if winCount != tt.expectedWinCount {
				t.Errorf("WinningTrades: expected %d, got %d", tt.expectedWinCount, winCount)
			}

			lossCount := s.GetLosingTradesCount()
			if lossCount != tt.expectedLossCount {
				t.Errorf("LosingTrades: expected %d, got %d", tt.expectedLossCount, lossCount)
			}

			evenCount := s.GetEvenTradesCount()
			if evenCount != tt.expectedEvenCount {
				t.Errorf("EvenTrades: expected %d, got %d", tt.expectedEvenCount, evenCount)
			}
		})
	}
}

func TestStrategyStatisticsBeforeAnyTrades(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	if s.GetGrossProfit() != 0 {
		t.Errorf("GrossProfit before trades: expected 0, got %.2f", s.GetGrossProfit())
	}

	if s.GetGrossLoss() != 0 {
		t.Errorf("GrossLoss before trades: expected 0, got %.2f", s.GetGrossLoss())
	}

	if s.GetWinningTradesCount() != 0 {
		t.Errorf("WinningTrades before trades: expected 0, got %d", s.GetWinningTradesCount())
	}

	if s.GetLosingTradesCount() != 0 {
		t.Errorf("LosingTrades before trades: expected 0, got %d", s.GetLosingTradesCount())
	}

	if s.GetEvenTradesCount() != 0 {
		t.Errorf("EvenTrades before trades: expected 0, got %d", s.GetEvenTradesCount())
	}
}

func TestStrategyStatisticsWithOpenTrades(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test", 10000, 2)

	s.Entry("long1", Long, 10, "")
	s.OnBarUpdate(1, 100, 1000)
	s.Close("long1", 110, 2000, "")
	s.OnBarUpdate(2, 110, 2000)

	s.Entry("long2", Long, 10, "")
	s.OnBarUpdate(3, 100, 3000)

	grossProfit := s.GetGrossProfit()
	if grossProfit != 100 {
		t.Errorf("GrossProfit with open trade: expected 100, got %.2f", grossProfit)
	}

	winCount := s.GetWinningTradesCount()
	if winCount != 1 {
		t.Errorf("WinningTrades with open trade: expected 1, got %d", winCount)
	}

	s.Close("long2", 90, 4000, "")
	s.OnBarUpdate(4, 90, 4000)

	grossLoss := s.GetGrossLoss()
	if grossLoss != -100 {
		t.Errorf("GrossLoss after closing: expected -100, got %.2f", grossLoss)
	}

	lossCount := s.GetLosingTradesCount()
	if lossCount != 1 {
		t.Errorf("LosingTrades after closing: expected 1, got %d", lossCount)
	}
}

func TestStrategyStatisticsStability(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	s.Entry("long1", Long, 10, "")
	s.OnBarUpdate(1, 100, 1000)
	s.Close("long1", 110, 2000, "")
	s.OnBarUpdate(2, 110, 2000)

	profit1 := s.GetGrossProfit()
	win1 := s.GetWinningTradesCount()

	for i := 3; i < 100; i++ {
		s.OnBarUpdate(i, 110+float64(i), int64(1000*i))
	}

	profit2 := s.GetGrossProfit()
	win2 := s.GetWinningTradesCount()

	if profit1 != profit2 {
		t.Errorf("GrossProfit changed: expected %.2f, got %.2f", profit1, profit2)
	}

	if win1 != win2 {
		t.Errorf("WinCount changed: expected %d, got %d", win1, win2)
	}
}

func TestStrategyEvenTrades(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	s.Entry("long1", Long, 10, "")
	s.OnBarUpdate(1, 100, 1000)
	s.Close("long1", 100, 2000, "")
	s.OnBarUpdate(2, 100, 2000)

	evenCount := s.GetEvenTradesCount()
	if evenCount != 1 {
		t.Errorf("EvenTrades: expected 1, got %d", evenCount)
	}
}

/*
TestClose_ClosesAllMatchingEntriesById verifies that Close("id") exits every open trade
carrying that entry ID, while leaving unrelated entries untouched and computing profit
independently for each closed trade. Covers homogeneous and mixed pyramiding, both
directions, variable lot sizes, and the no-op for a non-existent ID.
*/
func TestClose_ClosesAllMatchingEntriesById(t *testing.T) {
	type entry struct {
		id  string
		qty float64
	}
	tests := []struct {
		name             string
		pyramiding       int
		direction        string
		entries          []entry
		closeID          string
		entryPrice       float64
		exitPrice        float64
		wantClosedCount  int
		wantOpenCount    int
		wantPositionSize float64
		wantNetProfit    float64
	}{
		{
			name:       "two_same_id_long",
			pyramiding: 2, direction: Long,
			entries: []entry{{"e", 10}, {"e", 10}},
			closeID: "e", entryPrice: 100, exitPrice: 110,
			wantClosedCount: 2, wantOpenCount: 0,
			wantPositionSize: 0, wantNetProfit: 200, // (110-100)*10 * 2
		},
		{
			name:       "three_same_id_long",
			pyramiding: 3, direction: Long,
			entries: []entry{{"e", 10}, {"e", 10}, {"e", 10}},
			closeID: "e", entryPrice: 100, exitPrice: 110,
			wantClosedCount: 3, wantOpenCount: 0,
			wantPositionSize: 0, wantNetProfit: 300,
		},
		{
			name:       "two_same_id_short",
			pyramiding: 2, direction: Short,
			entries: []entry{{"s", 5}, {"s", 5}},
			closeID: "s", entryPrice: 100, exitPrice: 90,
			wantClosedCount: 2, wantOpenCount: 0,
			wantPositionSize: 0, wantNetProfit: 100, // (100-90)*5 * 2
		},
		{
			name:       "mixed_ids_only_target_closed",
			pyramiding: 3, direction: Long,
			entries: []entry{{"target", 10}, {"target", 10}, {"other", 5}},
			closeID: "target", entryPrice: 100, exitPrice: 110,
			wantClosedCount: 2, wantOpenCount: 1,
			wantPositionSize: 5, wantNetProfit: 200,
		},
		{
			name:       "variable_lot_sizes_profit_per_trade",
			pyramiding: 3, direction: Long,
			entries: []entry{{"e", 10}, {"e", 20}, {"e", 5}},
			closeID: "e", entryPrice: 100, exitPrice: 110,
			wantClosedCount: 3, wantOpenCount: 0,
			wantPositionSize: 0, wantNetProfit: 350, // (110-100)*(10+20+5)
		},
		{
			name:       "nonexistent_id_is_noop",
			pyramiding: 2, direction: Long,
			entries: []entry{{"e", 10}, {"e", 10}},
			closeID: "no_such_id", entryPrice: 100, exitPrice: 110,
			wantClosedCount: 0, wantOpenCount: 2,
			wantPositionSize: 20, wantNetProfit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.CallWithPyramiding("Test", 10000, tt.pyramiding)

			bar, ts := 0, int64(1000)
			for _, e := range tt.entries {
				s.Entry(e.id, tt.direction, e.qty, "")
				bar++
				ts += 1000
				s.OnBarUpdate(bar, tt.entryPrice, ts)
			}

			s.Close(tt.closeID, tt.exitPrice, ts, "")
			bar++
			ts += 1000
			s.OnBarUpdate(bar, tt.exitPrice, ts)

			closed := s.GetTradeHistory().GetClosedTrades()
			if len(closed) != tt.wantClosedCount {
				t.Errorf("closed count: got %d, want %d", len(closed), tt.wantClosedCount)
			}
			open := s.GetTradeHistory().GetOpenTrades()
			if len(open) != tt.wantOpenCount {
				t.Errorf("open count: got %d, want %d", len(open), tt.wantOpenCount)
			}
			if s.GetPositionSize() != tt.wantPositionSize {
				t.Errorf("position size: got %.2f, want %.2f", s.GetPositionSize(), tt.wantPositionSize)
			}
			if s.GetNetProfit() != tt.wantNetProfit {
				t.Errorf("net profit: got %.2f, want %.2f", s.GetNetProfit(), tt.wantNetProfit)
			}
		})
	}
}

/*
	TestCommission_FullClose verifies entry+exit commission is correct for all commission types

via both the Entry()/Close() and Order() execution paths.
*/
func TestCommission_FullClose(t *testing.T) {
	tests := []struct {
		commType       string
		commRate       float64
		qty            float64
		entryPrice     float64
		exitPrice      float64
		wantCommission float64
	}{
		{CommissionPercent, 1.0, 10, 100, 110, 10*100*1.0/100 + 10*110*1.0/100},
		{CommissionCashPerOrder, 5.0, 10, 100, 110, 5.0 + 5.0},
		{CommissionCashPerContract, 2.0, 5, 100, 110, 5*2.0 + 5*2.0},
	}

	for _, tt := range tests {
		t.Run(tt.commType, func(t *testing.T) {
			t.Run("entry_close_path", func(t *testing.T) {
				s := NewStrategy()
				s.CallWithPyramiding("Test", 10000, 10)
				s.SetCommission(tt.commRate, tt.commType)
				s.Entry("e", Long, tt.qty, "")
				s.OnBarUpdate(1, tt.entryPrice, 1000)
				s.Close("e", tt.entryPrice, 1000, "")
				s.OnBarUpdate(2, tt.exitPrice, 2000)
				closed := s.GetTradeHistory().GetClosedTrades()
				if len(closed) != 1 {
					t.Fatalf("expected 1 closed trade, got %d", len(closed))
				}
				if closed[0].Commission != tt.wantCommission {
					t.Errorf("commission: got %.4f, want %.4f", closed[0].Commission, tt.wantCommission)
				}
			})
			t.Run("order_path", func(t *testing.T) {
				s := NewStrategy()
				s.Call("Test", 10000)
				s.SetCommission(tt.commRate, tt.commType)
				s.Order("buy", Long, tt.qty, "")
				s.OnBarUpdate(1, tt.entryPrice, 1000)
				s.Order("sell", Short, tt.qty, "")
				s.OnBarUpdate(2, tt.exitPrice, 2000)
				closed := s.GetTradeHistory().GetClosedTrades()
				if len(closed) != 1 {
					t.Fatalf("expected 1 closed trade, got %d", len(closed))
				}
				if closed[0].Commission != tt.wantCommission {
					t.Errorf("commission: got %.4f, want %.4f", closed[0].Commission, tt.wantCommission)
				}
			})
			t.Run("reversal_path", func(t *testing.T) {
				s := NewStrategy()
				s.CallWithPyramiding("Test", 10000, 10)
				s.SetCommission(tt.commRate, tt.commType)
				s.Entry("e", Long, tt.qty, "")
				s.OnBarUpdate(1, tt.entryPrice, 1000)
				s.Entry("rev", Short, tt.qty, "")
				s.OnBarUpdate(2, tt.exitPrice, 2000)
				closed := s.GetTradeHistory().GetClosedTrades()
				if len(closed) != 1 {
					t.Fatalf("expected 1 closed trade, got %d", len(closed))
				}
				if closed[0].Commission != tt.wantCommission {
					t.Errorf("commission: got %.4f, want %.4f", closed[0].Commission, tt.wantCommission)
				}
			})
		})
	}
}

func TestStrategyAllowedDirection_EntryFilter(t *testing.T) {
	tests := []struct {
		name             string
		allowedDirection string
		entryDirection   string
		wantTradeCreated bool
	}{
		{"long-only allows long", DirectionLong, Long, true},
		{"long-only blocks short", DirectionLong, Short, false},
		{"short-only allows short", DirectionShort, Short, true},
		{"short-only blocks long", DirectionShort, Long, false},
		{"all allows long", DirectionAll, Long, true},
		{"all allows short", DirectionAll, Short, true},
		{"empty direction allows long", "", Long, true},
		{"empty direction allows short", "", Short, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.CallWithPyramiding("Test", 10000, 10)
			if tt.allowedDirection != "" {
				s.SetAllowedDirection(tt.allowedDirection)
			}

			s.Entry("e", tt.entryDirection, 10, "")
			s.OnBarUpdate(1, 100, 1000)

			open := s.GetTradeHistory().GetOpenTrades()
			got := len(open) > 0
			if got != tt.wantTradeCreated {
				t.Errorf("wantTradeCreated=%v, got %v", tt.wantTradeCreated, got)
			}
		})
	}
}

func TestStrategyGetPositionEntryName_NoPosition(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	name := s.GetPositionEntryName()
	if name != "" {
		t.Errorf("Expected empty string with no open trades, got %q", name)
	}
}

func TestStrategyGetPositionEntryName_WithOpenTrade(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test", 10000, 10)

	s.Entry("myEntry", Long, 10, "")
	s.OnBarUpdate(1, 100, 1000)

	name := s.GetPositionEntryName()
	if name != "myEntry" {
		t.Errorf("Expected 'myEntry', got %q", name)
	}
}

func TestStrategyGetPositionEntryName_MultipleOpenTrades(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test", 10000, 3)

	s.Entry("first", Long, 5, "")
	s.OnBarUpdate(1, 100, 1000)
	s.Entry("second", Long, 5, "")
	s.OnBarUpdate(2, 105, 2000)

	name := s.GetPositionEntryName()
	if name != "second" {
		t.Errorf("Expected last entry 'second', got %q", name)
	}
}

func TestStrategyGetPositionEntryName_AfterAllClosed(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test", 10000, 10)

	s.Entry("e", Long, 10, "")
	s.OnBarUpdate(1, 100, 1000)
	s.CloseAll(100, 1000, "")
	s.OnBarUpdate(2, 110, 2000)

	name := s.GetPositionEntryName()
	if name != "" {
		t.Errorf("Expected empty string after all closed, got %q", name)
	}
}

func TestStrategyCancel_NonexistentOrder(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test", 10000, 10)

	s.Cancel("ghost")
	s.OnBarUpdate(1, 100, 1000)

	if len(s.GetTradeHistory().GetOpenTrades()) != 0 {
		t.Error("Cancel of non-existent order should not create trades")
	}
}

func TestStrategyCancel_RemovesPendingOrder(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test", 10000, 10)

	s.Entry("myOrder", Long, 10, "")
	s.Cancel("myOrder")
	s.OnBarUpdate(1, 100, 1000)

	if len(s.GetTradeHistory().GetOpenTrades()) != 0 {
		t.Error("Cancelled order should not produce an open trade")
	}
}

func TestStrategyCancelAll_EmptyQueue(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test", 10000, 10)

	s.CancelAll()
	s.OnBarUpdate(1, 100, 1000)

	if len(s.GetTradeHistory().GetOpenTrades()) != 0 {
		t.Error("CancelAll on empty queue should not panic or create trades")
	}
}

func TestStrategyCancelAll_ClearsAllPendingOrders(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test", 10000, 10)

	s.Entry("order1", Long, 10, "")
	s.Entry("order2", Long, 5, "")
	s.CancelAll()
	s.OnBarUpdate(1, 100, 1000)

	if len(s.GetTradeHistory().GetOpenTrades()) != 0 {
		t.Error("CancelAll should clear all pending orders")
	}
}

/* TestStrategyOrder_NotInitialized verifies Order() is a no-op when strategy not initialized */
func TestStrategyOrder_NotInitialized(t *testing.T) {
	s := NewStrategy()

	err := s.Order("buy", Long, 10, "")
	if err == nil {
		t.Error("Order() on uninitialized strategy should return error")
	}
	s.OnBarUpdate(1, 100, 1000)
	if len(s.GetTradeHistory().GetOpenTrades()) != 0 {
		t.Error("Uninitialized Order() should not create trades")
	}
}

/* TestStrategyOrder_FlatToLong verifies opening a long from flat position */
func TestStrategyOrder_FlatToLong(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	s.Order("buy", Long, 10, "")
	s.OnBarUpdate(1, 100, 1000)

	if s.GetPositionSize() != 10 {
		t.Errorf("Position should be +10, got %.2f", s.GetPositionSize())
	}
	open := s.GetTradeHistory().GetOpenTrades()
	if len(open) != 1 || open[0].Direction != Long || open[0].Size != 10 {
		t.Errorf("Expected 1 long trade of size 10, got %v", open)
	}
}

/* TestStrategyOrder_FlatToShort verifies opening a short from flat position */
func TestStrategyOrder_FlatToShort(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	s.Order("sell", Short, 5, "")
	s.OnBarUpdate(1, 100, 1000)

	if s.GetPositionSize() != -5 {
		t.Errorf("Position should be -5, got %.2f", s.GetPositionSize())
	}
	open := s.GetTradeHistory().GetOpenTrades()
	if len(open) != 1 || open[0].Direction != Short || open[0].Size != 5 {
		t.Errorf("Expected 1 short trade of size 5, got %v", open)
	}
}

/* TestStrategyOrder_IgnoresPyramiding verifies no pyramiding enforcement */
func TestStrategyOrder_IgnoresPyramiding(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test", 10000, 0) // pyramiding=0 blocks Entry()

	s.Order("buy1", Long, 5, "")
	s.OnBarUpdate(1, 100, 1000)

	s.Order("buy2", Long, 5, "")
	s.OnBarUpdate(2, 105, 2000)

	if s.GetPositionSize() != 10 {
		t.Errorf("Order() must ignore pyramiding: expected position +10, got %.2f", s.GetPositionSize())
	}
	open := s.GetTradeHistory().GetOpenTrades()
	if len(open) != 2 {
		t.Errorf("Expected 2 open trades (pyramiding ignored), got %d", len(open))
	}
}

/* TestStrategyOrder_EntryPyramidingBlocksButOrderDoesNot verifies behavioral difference from Entry */
func TestStrategyOrder_EntryPyramidingBlocksButOrderDoesNot(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test", 10000, 0)

	s.Entry("e1", Long, 5, "")
	s.OnBarUpdate(1, 100, 1000)
	s.Entry("e2", Long, 5, "") // blocked by pyramiding
	s.OnBarUpdate(2, 105, 2000)

	if s.GetPositionSize() != 5 {
		t.Errorf("Entry should be blocked at pyramiding=0: expected +5, got %.2f", s.GetPositionSize())
	}

	// Order() with same setup: must NOT be blocked
	s2 := NewStrategy()
	s2.CallWithPyramiding("Test", 10000, 0)

	s2.Order("o1", Long, 5, "")
	s2.OnBarUpdate(1, 100, 1000)
	s2.Order("o2", Long, 5, "") // must NOT be blocked
	s2.OnBarUpdate(2, 105, 2000)

	if s2.GetPositionSize() != 10 {
		t.Errorf("Order() must bypass pyramiding: expected +10, got %.2f", s2.GetPositionSize())
	}
}

/* TestStrategyOrder_ReducesLongPosition verifies partial close of long via short order */
func TestStrategyOrder_ReducesLongPosition(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	s.Order("buy", Long, 15, "")
	s.OnBarUpdate(1, 100, 1000)

	s.Order("sell1", Short, 5, "")
	s.OnBarUpdate(2, 110, 2000)

	if s.GetPositionSize() != 10 {
		t.Errorf("Expected position +10 after partial short, got %.2f", s.GetPositionSize())
	}
	if len(s.GetTradeHistory().GetOpenTrades()) != 1 {
		t.Errorf("Expected 1 open trade, got %d", len(s.GetTradeHistory().GetOpenTrades()))
	}
	if len(s.GetTradeHistory().GetClosedTrades()) != 1 {
		t.Errorf("Expected 1 closed trade, got %d", len(s.GetTradeHistory().GetClosedTrades()))
	}
	closed := s.GetTradeHistory().GetClosedTrades()[0]
	if closed.Size != 5 {
		t.Errorf("Expected closed trade size 5, got %.2f", closed.Size)
	}
	if closed.Profit != 50 { // (110-100)*5
		t.Errorf("Expected closed profit 50, got %.2f", closed.Profit)
	}
}

/* TestStrategyOrder_ThreeReducingOrders verifies Pine docs example: 15-5-5-5=0 */
func TestStrategyOrder_ThreeReducingOrders(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	s.Order("buy", Long, 15, "")
	s.OnBarUpdate(1, 100, 1000)

	s.Order("sell1", Short, 5, "")
	s.OnBarUpdate(2, 110, 2000)
	s.Order("sell2", Short, 5, "")
	s.OnBarUpdate(3, 115, 3000)
	s.Order("sell3", Short, 5, "")
	s.OnBarUpdate(4, 120, 4000)

	if s.GetPositionSize() != 0 {
		t.Errorf("Expected position 0 after 3x5 reduces of 15, got %.2f", s.GetPositionSize())
	}
	if len(s.GetTradeHistory().GetOpenTrades()) != 0 {
		t.Errorf("Expected 0 open trades, got %d", len(s.GetTradeHistory().GetOpenTrades()))
	}
	if len(s.GetTradeHistory().GetClosedTrades()) != 3 {
		t.Errorf("Expected 3 closed trades, got %d", len(s.GetTradeHistory().GetClosedTrades()))
	}
}

/* TestStrategyOrder_ExactClose verifies full close of long position */
func TestStrategyOrder_ExactClose(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	s.Order("buy", Long, 10, "")
	s.OnBarUpdate(1, 100, 1000)
	s.Order("sell", Short, 10, "")
	s.OnBarUpdate(2, 120, 2000)

	if s.GetPositionSize() != 0 {
		t.Errorf("Expected flat position, got %.2f", s.GetPositionSize())
	}
	closed := s.GetTradeHistory().GetClosedTrades()
	if len(closed) != 1 {
		t.Fatalf("Expected 1 closed trade, got %d", len(closed))
	}
	if closed[0].Profit != 200 { // (120-100)*10
		t.Errorf("Expected profit 200, got %.2f", closed[0].Profit)
	}
}

/* TestStrategyOrder_CrossZeroLongToShort verifies crossing zero from long to short */
func TestStrategyOrder_CrossZeroLongToShort(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	s.Order("buy", Long, 10, "")
	s.OnBarUpdate(1, 100, 1000)

	// Short 15 against long 10 → closes 10 long, opens 5 short
	s.Order("sell", Short, 15, "")
	s.OnBarUpdate(2, 120, 2000)

	if s.GetPositionSize() != -5 {
		t.Errorf("Expected position -5 after cross-zero, got %.2f", s.GetPositionSize())
	}
	open := s.GetTradeHistory().GetOpenTrades()
	if len(open) != 1 || open[0].Direction != Short || open[0].Size != 5 {
		t.Errorf("Expected 1 short trade of size 5, got %v", open)
	}
	closed := s.GetTradeHistory().GetClosedTrades()
	if len(closed) != 1 || closed[0].Size != 10 {
		t.Errorf("Expected 1 closed long trade of size 10, got %v", closed)
	}
}

/* TestStrategyOrder_CrossZeroShortToLong verifies crossing zero from short to long */
func TestStrategyOrder_CrossZeroShortToLong(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	s.Order("sell", Short, 8, "")
	s.OnBarUpdate(1, 100, 1000)

	// Buy 12 against short 8 → closes 8 short, opens 4 long
	s.Order("buy", Long, 12, "")
	s.OnBarUpdate(2, 90, 2000)

	if s.GetPositionSize() != 4 {
		t.Errorf("Expected position +4 after cross-zero, got %.2f", s.GetPositionSize())
	}
	open := s.GetTradeHistory().GetOpenTrades()
	if len(open) != 1 || open[0].Direction != Long || open[0].Size != 4 {
		t.Errorf("Expected 1 long trade of size 4, got %v", open)
	}
	closed := s.GetTradeHistory().GetClosedTrades()
	if len(closed) != 1 || closed[0].Size != 8 {
		t.Errorf("Expected 1 closed short trade of size 8, got %v", closed)
	}
	if closed[0].Profit != 80 { // (100-90)*8 short profit
		t.Errorf("Expected short profit 80, got %.2f", closed[0].Profit)
	}
}

/* TestStrategyOrder_FIFOPartialClose verifies FIFO ordering across multiple open trades */
func TestStrategyOrder_FIFOPartialClose(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	// Open two separate long trades (requires pyramiding bypass via Order)
	s.Order("buy1", Long, 10, "first")
	s.OnBarUpdate(1, 100, 1000)
	s.Order("buy2", Long, 10, "second")
	s.OnBarUpdate(2, 105, 2000)

	s.Order("sell", Short, 12, "")
	s.OnBarUpdate(3, 110, 3000)

	if s.GetPositionSize() != 8 {
		t.Errorf("Expected position +8, got %.2f", s.GetPositionSize())
	}

	open := s.GetTradeHistory().GetOpenTrades()
	if len(open) != 1 {
		t.Fatalf("Expected 1 open trade, got %d", len(open))
	}
	if open[0].EntryComment != "second" {
		t.Errorf("Expected remaining open trade to be 'second' (FIFO), got %q", open[0].EntryComment)
	}
	if open[0].Size != 8 {
		t.Errorf("Expected remaining open size 8, got %.2f", open[0].Size)
	}

	closed := s.GetTradeHistory().GetClosedTrades()
	if len(closed) != 2 {
		t.Fatalf("Expected 2 closed trades (buy1 full + buy2 partial), got %d", len(closed))
	}
	// FIFO: buy1 closed first (size 10), then buy2 partial (size 2)
	if closed[0].EntryComment != "first" || closed[0].Size != 10 {
		t.Errorf("Expected first closed trade 'first' size 10, got %q size %.2f", closed[0].EntryComment, closed[0].Size)
	}
	if closed[1].EntryComment != "second" || closed[1].Size != 2 {
		t.Errorf("Expected second closed trade 'second' size 2, got %q size %.2f", closed[1].EntryComment, closed[1].Size)
	}
}

/* TestStrategyOrder_NoAutoReversal verifies strategy.order does NOT auto-reverse like strategy.entry */
func TestStrategyOrder_NoAutoReversal(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	// Enter long with Entry() — establishes position
	s.Order("buy", Long, 10, "")
	s.OnBarUpdate(1, 100, 1000)

	// Order short 5 — should REDUCE long (not reverse like Entry would)
	s.Order("sell", Short, 5, "")
	s.OnBarUpdate(2, 110, 2000)

	// Entry() would have reversed: closed all 10 long, opened 5 short → position -5
	// Order() should net: 10 - 5 = +5 long
	if s.GetPositionSize() != 5 {
		t.Errorf("Order() must NOT auto-reverse: expected +5, got %.2f", s.GetPositionSize())
	}
}

/* TestStrategyOrder_AddsToShortPosition verifies adding to short position */
func TestStrategyOrder_AddsToShortPosition(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	s.Order("sell1", Short, 5, "")
	s.OnBarUpdate(1, 100, 1000)
	s.Order("sell2", Short, 3, "")
	s.OnBarUpdate(2, 105, 2000)

	if s.GetPositionSize() != -8 {
		t.Errorf("Expected position -8 after two short orders, got %.2f", s.GetPositionSize())
	}
	if len(s.GetTradeHistory().GetOpenTrades()) != 2 {
		t.Errorf("Expected 2 open trades, got %d", len(s.GetTradeHistory().GetOpenTrades()))
	}
}

/* TestStrategyOrder_CommentPropagates verifies comment reaches Trade.EntryComment */
func TestStrategyOrder_CommentPropagates(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	s.Order("buy", Long, 10, "my signal")
	s.OnBarUpdate(1, 100, 1000)

	open := s.GetTradeHistory().GetOpenTrades()
	if len(open) != 1 {
		t.Fatal("Expected 1 open trade")
	}
	if open[0].EntryComment != "my signal" {
		t.Errorf("Expected comment 'my signal', got %q", open[0].EntryComment)
	}
}

/* TestStrategyOrder_CancelRemovesOrder verifies Cancel() works on Order()-created pending orders */
func TestStrategyOrder_CancelRemovesOrder(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	s.Order("buy", Long, 10, "")
	s.Cancel("buy")
	s.OnBarUpdate(1, 100, 1000)

	if len(s.GetTradeHistory().GetOpenTrades()) != 0 {
		t.Error("Cancelled Order() should not produce a trade")
	}
}

/* TestStrategyOrder_AllowedDirectionRespected verifies direction restriction applies to Order() */
func TestStrategyOrder_AllowedDirectionRespected(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test", 10000, 10)
	s.SetAllowedDirection(DirectionLong)

	s.Order("sell", Short, 10, "")
	s.OnBarUpdate(1, 100, 1000)

	if len(s.GetTradeHistory().GetOpenTrades()) != 0 {
		t.Error("Order() should respect allowedDirection: short blocked when long-only")
	}

	s.Order("buy", Long, 10, "")
	s.OnBarUpdate(2, 100, 2000)

	if len(s.GetTradeHistory().GetOpenTrades()) != 1 {
		t.Error("Order() should respect allowedDirection: long allowed when long-only")
	}
}

/* TestStrategyOrder_EquityUpdatedOnClose verifies equity is correctly updated after net-order closes position */
func TestStrategyOrder_EquityUpdatedOnClose(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	s.Order("buy", Long, 10, "")
	s.OnBarUpdate(1, 100, 1000)
	s.Order("sell", Short, 10, "")
	s.OnBarUpdate(2, 120, 2000)

	expectedEquity := 10000.0 + 200.0 // (120-100)*10
	if s.GetEquity(120) != expectedEquity {
		t.Errorf("Expected equity %.2f, got %.2f", expectedEquity, s.GetEquity(120))
	}
}

/* TestStrategyOrder_PartialCloseEquityUpdate verifies equity updates correctly on partial close */
func TestStrategyOrder_PartialCloseEquityUpdate(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	s.Order("buy", Long, 10, "")
	s.OnBarUpdate(1, 100, 1000)
	s.Order("sell", Short, 6, "")
	s.OnBarUpdate(2, 110, 2000)

	realizedEquity := s.GetEquity(110)
	if realizedEquity != 10100 {
		t.Errorf("Expected equity 10100, got %.2f", realizedEquity)
	}
}

/* TestStrategyOrder_MultiplePartialCloses verifies correct running position across sequential partial closes */
func TestStrategyOrder_MultiplePartialCloses(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	s.Order("buy", Long, 9, "")
	s.OnBarUpdate(1, 100, 1000)

	s.Order("sell1", Short, 3, "")
	s.OnBarUpdate(2, 110, 2000)
	s.Order("sell2", Short, 3, "")
	s.OnBarUpdate(3, 115, 3000)
	s.Order("sell3", Short, 3, "")
	s.OnBarUpdate(4, 120, 4000)

	if s.GetPositionSize() != 0 {
		t.Errorf("Expected flat position, got %.2f", s.GetPositionSize())
	}
	if len(s.GetTradeHistory().GetOpenTrades()) != 0 {
		t.Errorf("Expected 0 open trades, got %d", len(s.GetTradeHistory().GetOpenTrades()))
	}
	if len(s.GetTradeHistory().GetClosedTrades()) != 3 {
		t.Errorf("Expected 3 partial-closed trades, got %d", len(s.GetTradeHistory().GetClosedTrades()))
	}
}

/* TestCalcCommission verifies commission calculation for all types across qty and price variations. */
func TestCalcCommission(t *testing.T) {
	tests := []struct {
		name     string
		commType string
		commRate float64
		qty      float64
		price    float64
		want     float64
	}{
		{"percent_normal", CommissionPercent, 1.0, 10, 100, 10.0},
		{"percent_half_rate", CommissionPercent, 0.5, 10, 100, 5.0},
		{"percent_scales_with_price", CommissionPercent, 1.0, 1, 500, 5.0},
		{"percent_scales_with_qty", CommissionPercent, 1.0, 5, 100, 5.0},
		{"percent_zero_rate", CommissionPercent, 0.0, 10, 100, 0.0},
		{"per_order_flat_small_qty", CommissionCashPerOrder, 5.0, 1, 100, 5.0},
		{"per_order_flat_large_qty", CommissionCashPerOrder, 5.0, 100, 100, 5.0},
		{"per_order_ignores_price", CommissionCashPerOrder, 3.0, 10, 1000, 3.0},
		{"per_contract_scales_qty", CommissionCashPerContract, 2.0, 5, 100, 10.0},
		{"per_contract_single_unit", CommissionCashPerContract, 2.0, 1, 100, 2.0},
		{"per_contract_ignores_price", CommissionCashPerContract, 2.0, 5, 1000, 10.0},
		{"per_contract_zero_qty", CommissionCashPerContract, 2.0, 0, 100, 0.0},
		{"unknown_type_returns_zero", "invalid", 5.0, 10, 100, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.SetCommission(tt.commRate, tt.commType)
			got := s.calcCommission(tt.qty, tt.price)
			if got != tt.want {
				t.Errorf("calcCommission(qty=%.2f, price=%.2f) = %.4f, want %.4f", tt.qty, tt.price, got, tt.want)
			}
		})
	}
}

/*
	TestDirectionMultiplier verifies the profit sign convention: long profits rise with price,

short profits rise as price falls. Any non-Short string defaults to the long multiplier.
*/
func TestDirectionMultiplier(t *testing.T) {
	tests := []struct {
		direction string
		want      float64
	}{
		{Long, 1.0},
		{Short, -1.0},
		{"", 1.0},
		{"unknown", 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.direction, func(t *testing.T) {
			if got := directionMultiplier(tt.direction); got != tt.want {
				t.Errorf("directionMultiplier(%q) = %.1f, want %.1f", tt.direction, got, tt.want)
			}
		})
	}
}

/*
	TestBuildClosedTrade verifies that all entry fields propagate from the open trade,

all exit fields are applied correctly, and metricsScale proportions MaxDrawdown/MaxRunup.
*/
func TestBuildClosedTrade(t *testing.T) {
	open := Trade{
		EntryID:      "entry1",
		Direction:    Long,
		Size:         10,
		EntryPrice:   100,
		EntryBar:     1,
		EntryTime:    1000,
		EntryComment: "entry comment",
		Commission:   5.0,
		MaxDrawdown:  20.0,
		MaxRunup:     30.0,
	}
	const exitID = "exit1"
	const exitComment = "exit comment"
	const exitPrice = 110.0
	const exitBar = 2
	const exitTime = int64(2000)

	tests := []struct {
		name         string
		closeSize    float64
		grossProfit  float64
		commission   float64
		metricsScale float64
		wantMaxDD    float64
		wantMaxRU    float64
	}{
		{"full close preserves metrics at scale 1.0", 10, 100, 15, 1.0, 20.0, 30.0},
		{"half close scales metrics by 0.5", 5, 50, 7.5, 0.5, 10.0, 15.0},
		{"30 pct close scales metrics by 0.3", 3, 30, 4.5, 0.3, 6.0, 9.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			closed := buildClosedTrade(open, tt.closeSize, tt.grossProfit, tt.commission, tt.metricsScale, exitID, exitPrice, exitBar, exitTime, exitComment)

			if closed.EntryID != open.EntryID {
				t.Errorf("EntryID: got %q, want %q", closed.EntryID, open.EntryID)
			}
			if closed.Direction != open.Direction {
				t.Errorf("Direction: got %q, want %q", closed.Direction, open.Direction)
			}
			if closed.EntryPrice != open.EntryPrice {
				t.Errorf("EntryPrice: got %.2f, want %.2f", closed.EntryPrice, open.EntryPrice)
			}
			if closed.EntryBar != open.EntryBar {
				t.Errorf("EntryBar: got %d, want %d", closed.EntryBar, open.EntryBar)
			}
			if closed.EntryTime != open.EntryTime {
				t.Errorf("EntryTime: got %d, want %d", closed.EntryTime, open.EntryTime)
			}
			if closed.EntryComment != open.EntryComment {
				t.Errorf("EntryComment: got %q, want %q", closed.EntryComment, open.EntryComment)
			}
			if closed.ExitID != exitID {
				t.Errorf("ExitID: got %q, want %q", closed.ExitID, exitID)
			}
			if closed.ExitPrice != exitPrice {
				t.Errorf("ExitPrice: got %.2f, want %.2f", closed.ExitPrice, exitPrice)
			}
			if closed.ExitBar != exitBar {
				t.Errorf("ExitBar: got %d, want %d", closed.ExitBar, exitBar)
			}
			if closed.ExitTime != exitTime {
				t.Errorf("ExitTime: got %d, want %d", closed.ExitTime, exitTime)
			}
			if closed.ExitComment != exitComment {
				t.Errorf("ExitComment: got %q, want %q", closed.ExitComment, exitComment)
			}
			if closed.Size != tt.closeSize {
				t.Errorf("Size: got %.2f, want %.2f", closed.Size, tt.closeSize)
			}
			wantProfit := tt.grossProfit - tt.commission
			if closed.Profit != wantProfit {
				t.Errorf("Profit: got %.4f, want %.4f (gross=%.4f - commission=%.4f)",
					closed.Profit, wantProfit, tt.grossProfit, tt.commission)
			}
			if closed.Commission != tt.commission {
				t.Errorf("Commission: got %.4f, want %.4f", closed.Commission, tt.commission)
			}
			if closed.MaxDrawdown != tt.wantMaxDD {
				t.Errorf("MaxDrawdown: got %.4f, want %.4f", closed.MaxDrawdown, tt.wantMaxDD)
			}
			if closed.MaxRunup != tt.wantMaxRU {
				t.Errorf("MaxRunup: got %.4f, want %.4f", closed.MaxRunup, tt.wantMaxRU)
			}
		})
	}
}

/*
TestTradeHistory_MatchingOpenTrades verifies the entry-ID scoped snapshot: empty history,
no match, single match, multiple same-ID match, mixed IDs, and result isolation from
subsequent mutations to the source collection.
*/
func TestTradeHistory_MatchingOpenTrades(t *testing.T) {
	tests := []struct {
		name    string
		setup   []Trade
		entryID string
		wantLen int
		wantIDs []string
	}{
		{
			name:    "empty_history_returns_empty",
			setup:   nil,
			entryID: "x",
			wantLen: 0,
		},
		{
			name:    "no_matching_id_returns_empty",
			setup:   []Trade{{EntryID: "a"}, {EntryID: "b"}},
			entryID: "x",
			wantLen: 0,
		},
		{
			name:    "single_exact_match",
			setup:   []Trade{{EntryID: "a", Size: 10}},
			entryID: "a",
			wantLen: 1, wantIDs: []string{"a"},
		},
		{
			name:    "multiple_same_id_all_returned",
			setup:   []Trade{{EntryID: "a"}, {EntryID: "a"}, {EntryID: "a"}},
			entryID: "a",
			wantLen: 3, wantIDs: []string{"a", "a", "a"},
		},
		{
			name:    "mixed_ids_returns_matching_subset_preserving_order",
			setup:   []Trade{{EntryID: "a"}, {EntryID: "b"}, {EntryID: "a"}},
			entryID: "a",
			wantLen: 2, wantIDs: []string{"a", "a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTradeHistory()
			for _, tr := range tt.setup {
				th.AddOpenTrade(tr)
			}

			result := th.MatchingOpenTrades(tt.entryID)

			if len(result) != tt.wantLen {
				t.Fatalf("len: got %d, want %d", len(result), tt.wantLen)
			}
			for i, wantID := range tt.wantIDs {
				if result[i].EntryID != wantID {
					t.Errorf("result[%d].EntryID: got %q, want %q", i, result[i].EntryID, wantID)
				}
			}
		})
	}
}

/*
TestTradeHistory_MatchingOpenTrades_ResultIsIndependentCopy verifies that mutating the
returned slice does not affect the underlying open trade collection.
*/
func TestTradeHistory_MatchingOpenTrades_ResultIsIndependentCopy(t *testing.T) {
	th := NewTradeHistory()
	th.AddOpenTrade(Trade{EntryID: "a", Size: 10})
	th.AddOpenTrade(Trade{EntryID: "a", Size: 20})

	result := th.MatchingOpenTrades("a")
	result[0].EntryID = "mutated"

	open := th.GetOpenTrades()
	if open[0].EntryID != "a" {
		t.Errorf("mutating result affected source: got %q, want %q", open[0].EntryID, "a")
	}
}

/*
TestTradeHistory_AllOpenTradesSnapshot verifies the unfiltered safe snapshot: empty history,
single trade, multiple trades, and result isolation from subsequent mutations.
*/
func TestTradeHistory_AllOpenTradesSnapshot(t *testing.T) {
	tests := []struct {
		name    string
		setup   []Trade
		wantLen int
		wantIDs []string
	}{
		{
			name:    "empty_history_returns_empty",
			setup:   nil,
			wantLen: 0,
		},
		{
			name:    "single_trade_returned",
			setup:   []Trade{{EntryID: "a", Size: 10}},
			wantLen: 1, wantIDs: []string{"a"},
		},
		{
			name:    "multiple_trades_all_returned_preserving_insertion_order",
			setup:   []Trade{{EntryID: "a"}, {EntryID: "b"}, {EntryID: "c"}},
			wantLen: 3, wantIDs: []string{"a", "b", "c"},
		},
		{
			name:    "duplicate_ids_all_returned_preserving_insertion_order",
			setup:   []Trade{{EntryID: "x"}, {EntryID: "x"}, {EntryID: "y"}, {EntryID: "x"}},
			wantLen: 4, wantIDs: []string{"x", "x", "y", "x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTradeHistory()
			for _, tr := range tt.setup {
				th.AddOpenTrade(tr)
			}

			result := th.AllOpenTradesSnapshot()

			if len(result) != tt.wantLen {
				t.Fatalf("len: got %d, want %d", len(result), tt.wantLen)
			}
			for i, wantID := range tt.wantIDs {
				if result[i].EntryID != wantID {
					t.Errorf("result[%d].EntryID: got %q, want %q", i, result[i].EntryID, wantID)
				}
			}
		})
	}
}

/*
TestTradeHistory_AllOpenTradesSnapshot_ResultIsIndependentCopy verifies that mutating the
returned slice does not affect the underlying open trade collection.
*/
func TestTradeHistory_AllOpenTradesSnapshot_ResultIsIndependentCopy(t *testing.T) {
	th := NewTradeHistory()
	th.AddOpenTrade(Trade{EntryID: "a", Size: 10})
	th.AddOpenTrade(Trade{EntryID: "b", Size: 20})

	result := th.AllOpenTradesSnapshot()
	result[0].EntryID = "mutated"

	open := th.GetOpenTrades()
	if open[0].EntryID != "a" {
		t.Errorf("mutating result affected source: got %q, want %q", open[0].EntryID, "a")
	}
}

/*
	TestTradeHistory_PartialCloseTrades verifies FIFO close algorithm: qty distribution, commission

proportioning, and profit calculation for all close patterns (full, partial, multi-trade, short).
*/
func TestTradeHistory_PartialCloseTrades(t *testing.T) {
	type closedWant struct{ size, commission, profit float64 }
	type openWant struct{ size, commission float64 }
	tests := []struct {
		name                string
		setup               []Trade
		direction           string
		qty                 float64
		exitPrice           float64
		totalExitCommission float64
		wantQtyClosed       float64
		wantNewlyClosedLen  int
		wantClosed          []closedWant
		wantOpen            []openWant
	}{
		{
			name:      "full_close_single_long",
			setup:     []Trade{{EntryID: "t1", Direction: Long, Size: 10, EntryPrice: 100, Commission: 20}},
			direction: Long, qty: 10, exitPrice: 110, totalExitCommission: 15,
			wantQtyClosed: 10, wantNewlyClosedLen: 1,
			wantClosed: []closedWant{{10, 35, 65}}, // net = gross(100) - commission(35)
		},
		{
			name:      "partial_close_single_long_1_of_10",
			setup:     []Trade{{EntryID: "t1", Direction: Long, Size: 10, EntryPrice: 100, Commission: 10}},
			direction: Long, qty: 1, exitPrice: 110, totalExitCommission: 1,
			wantQtyClosed: 1, wantNewlyClosedLen: 1,
			wantClosed: []closedWant{{1, 2, 8}}, // net = gross(10) - commission(2)
			wantOpen:   []openWant{{9, 9}},
		},
		{
			name:      "partial_close_single_long_4_of_10",
			setup:     []Trade{{EntryID: "t1", Direction: Long, Size: 10, EntryPrice: 100, Commission: 20}},
			direction: Long, qty: 4, exitPrice: 110, totalExitCommission: 8,
			wantQtyClosed: 4, wantNewlyClosedLen: 1,
			wantClosed: []closedWant{{4, 16, 24}}, // net = gross(40) - commission(16)
			wantOpen:   []openWant{{6, 12}},
		},
		{
			name:      "partial_close_single_long_9_of_10",
			setup:     []Trade{{EntryID: "t1", Direction: Long, Size: 10, EntryPrice: 100, Commission: 10}},
			direction: Long, qty: 9, exitPrice: 110, totalExitCommission: 9,
			wantQtyClosed: 9, wantNewlyClosedLen: 1,
			wantClosed: []closedWant{{9, 18, 72}}, // net = gross(90) - commission(18)
			wantOpen:   []openWant{{1, 1}},
		},
		{
			name:      "full_close_single_short",
			setup:     []Trade{{EntryID: "s1", Direction: Short, Size: 8, EntryPrice: 100, Commission: 16}},
			direction: Short, qty: 8, exitPrice: 90, totalExitCommission: 16,
			wantQtyClosed: 8, wantNewlyClosedLen: 1,
			wantClosed: []closedWant{{8, 32, 48}}, // net = gross(80) - commission(32)
		},
		{
			name:      "partial_close_single_short_6_of_10",
			setup:     []Trade{{EntryID: "s1", Direction: Short, Size: 10, EntryPrice: 100, Commission: 20}},
			direction: Short, qty: 6, exitPrice: 90, totalExitCommission: 12,
			wantQtyClosed: 6, wantNewlyClosedLen: 1,
			wantClosed: []closedWant{{6, 24, 36}}, // net = gross(60) - commission(24)
			wantOpen:   []openWant{{4, 8}},
		},
		{
			name: "fifo_two_trades_full_close",
			setup: []Trade{
				{EntryID: "t1", Direction: Long, Size: 5, EntryPrice: 100, Commission: 10},
				{EntryID: "t2", Direction: Long, Size: 5, EntryPrice: 105, Commission: 10},
			},
			direction: Long, qty: 10, exitPrice: 115, totalExitCommission: 20,
			wantQtyClosed: 10, wantNewlyClosedLen: 2,
			wantClosed: []closedWant{{5, 20, 55}, {5, 20, 30}}, // net = gross-commission per lot
		},
		{
			name: "fifo_first_full_second_partial",
			setup: []Trade{
				{EntryID: "t1", Direction: Long, Size: 6, EntryPrice: 100, Commission: 12},
				{EntryID: "t2", Direction: Long, Size: 4, EntryPrice: 105, Commission: 8},
			},
			direction: Long, qty: 8, exitPrice: 110, totalExitCommission: 8,
			wantQtyClosed: 8, wantNewlyClosedLen: 2,
			wantClosed: []closedWant{{6, 18, 42}, {2, 6, 4}}, // net = gross-commission per lot
			wantOpen:   []openWant{{2, 4}},
		},
		{
			name: "fifo_three_trades_spanning_all",
			setup: []Trade{
				{EntryID: "t1", Direction: Long, Size: 3, EntryPrice: 100, Commission: 6},
				{EntryID: "t2", Direction: Long, Size: 3, EntryPrice: 102, Commission: 6},
				{EntryID: "t3", Direction: Long, Size: 4, EntryPrice: 104, Commission: 8},
			},
			direction: Long, qty: 10, exitPrice: 110, totalExitCommission: 10,
			wantQtyClosed: 10, wantNewlyClosedLen: 3,
			wantClosed: []closedWant{{3, 9, 21}, {3, 9, 15}, {4, 12, 12}}, // net = gross-commission per lot
		},
		{
			name:      "no_matching_direction",
			setup:     []Trade{{EntryID: "t1", Direction: Long, Size: 10, EntryPrice: 100, Commission: 20}},
			direction: Short, qty: 5, exitPrice: 110, totalExitCommission: 10,
			wantQtyClosed: 0, wantNewlyClosedLen: 0,
			wantOpen: []openWant{{10, 20}},
		},
		{
			name:      "qty_exceeds_available_capped",
			setup:     []Trade{{EntryID: "t1", Direction: Long, Size: 5, EntryPrice: 100, Commission: 10}},
			direction: Long, qty: 10, exitPrice: 110, totalExitCommission: 10,
			wantQtyClosed: 5, wantNewlyClosedLen: 1,
			wantClosed: []closedWant{{5, 15, 35}}, // net = gross(50) - commission(15)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTradeHistory()
			for _, trade := range tt.setup {
				th.AddOpenTrade(trade)
			}

			qtyClosed, newlyClosed := th.PartialCloseTrades(tt.direction, "exit", tt.qty, tt.exitPrice, 1, 1000, "", tt.totalExitCommission)

			if qtyClosed != tt.wantQtyClosed {
				t.Errorf("qtyClosed: got %.2f, want %.2f", qtyClosed, tt.wantQtyClosed)
			}
			if len(newlyClosed) != tt.wantNewlyClosedLen {
				t.Errorf("newlyClosed count: got %d, want %d", len(newlyClosed), tt.wantNewlyClosedLen)
			}

			closedTrades := th.GetClosedTrades()
			if len(closedTrades) != len(tt.wantClosed) {
				t.Fatalf("closed count: got %d, want %d", len(closedTrades), len(tt.wantClosed))
			}
			for i, want := range tt.wantClosed {
				if closedTrades[i].Size != want.size {
					t.Errorf("closed[%d].Size: got %.2f, want %.2f", i, closedTrades[i].Size, want.size)
				}
				if closedTrades[i].Commission != want.commission {
					t.Errorf("closed[%d].Commission: got %.4f, want %.4f", i, closedTrades[i].Commission, want.commission)
				}
				if closedTrades[i].Profit != want.profit {
					t.Errorf("closed[%d].Profit: got %.4f, want %.4f", i, closedTrades[i].Profit, want.profit)
				}
			}

			openTrades := th.GetOpenTrades()
			if len(openTrades) != len(tt.wantOpen) {
				t.Fatalf("open count: got %d, want %d", len(openTrades), len(tt.wantOpen))
			}
			for i, want := range tt.wantOpen {
				if openTrades[i].Size != want.size {
					t.Errorf("open[%d].Size: got %.2f, want %.2f", i, openTrades[i].Size, want.size)
				}
				if openTrades[i].Commission != want.commission {
					t.Errorf("open[%d].Commission: got %.4f, want %.4f", i, openTrades[i].Commission, want.commission)
				}
			}
		})
	}
}

/*
	TestCommission_PartialClose verifies commission split for partial closes across all commission types

and multiple close ratios, including the conservation invariant:
closed.Commission + remaining.Commission = entryCommission + exitCommission.
*/
func TestCommission_PartialClose(t *testing.T) {
	tests := []struct {
		name              string
		commType          string
		commRate          float64
		openQty           float64
		entryPrice        float64
		closeQty          float64
		exitPrice         float64
		wantClosedComm    float64
		wantRemainingComm float64
	}{
		/* CashPerContract */
		{"cpc_1_of_10", CommissionCashPerContract, 2.0, 10, 100, 1, 110, 4.0, 18.0},
		{"cpc_4_of_10", CommissionCashPerContract, 2.0, 10, 100, 4, 110, 16.0, 12.0},
		{"cpc_half", CommissionCashPerContract, 2.0, 10, 100, 5, 110, 20.0, 10.0},
		{"cpc_9_of_10", CommissionCashPerContract, 2.0, 10, 100, 9, 110, 36.0, 2.0},
		/* CashPerOrder: flat fee split proportionally by closeQty/openQty */
		{"cpo_4_of_10", CommissionCashPerOrder, 5.0, 10, 100, 4, 110, 7.0, 3.0},
		{"cpo_half", CommissionCashPerOrder, 5.0, 10, 100, 5, 110, 7.5, 2.5},
		/* Percent */
		{"pct_4_of_10", CommissionPercent, 1.0, 10, 100, 4, 110, 8.4, 6.0},
		{"pct_half", CommissionPercent, 1.0, 10, 100, 5, 110, 10.5, 5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.Call("Test", 10000)
			s.SetCommission(tt.commRate, tt.commType)

			s.Order("buy", Long, tt.openQty, "")
			s.OnBarUpdate(1, tt.entryPrice, 1000)
			s.Order("sell", Short, tt.closeQty, "")
			s.OnBarUpdate(2, tt.exitPrice, 2000)

			closed := s.GetTradeHistory().GetClosedTrades()
			if len(closed) != 1 {
				t.Fatalf("expected 1 closed trade, got %d", len(closed))
			}
			if closed[0].Commission != tt.wantClosedComm {
				t.Errorf("closed commission: got %.4f, want %.4f", closed[0].Commission, tt.wantClosedComm)
			}

			open := s.GetTradeHistory().GetOpenTrades()
			if len(open) != 1 {
				t.Fatalf("expected 1 remaining open trade, got %d", len(open))
			}
			if open[0].Commission != tt.wantRemainingComm {
				t.Errorf("remaining commission: got %.4f, want %.4f", open[0].Commission, tt.wantRemainingComm)
			}

			/* Conservation: closed + remaining = total entry + exit commissions */
			totalEntry := s.calcCommission(tt.openQty, tt.entryPrice)
			totalExit := s.calcCommission(tt.closeQty, tt.exitPrice)
			if closed[0].Commission+open[0].Commission != totalEntry+totalExit {
				t.Errorf("conservation violated: closed+remaining=%.4f, entry+exit=%.4f",
					closed[0].Commission+open[0].Commission, totalEntry+totalExit)
			}
		})
	}
}

/*
	TestCommission_FIFOClose verifies commission distribution when a close order spans multiple open

trades in FIFO order, for all commission types.
*/
func TestCommission_FIFOClose(t *testing.T) {
	type tradeInput struct{ qty, entryPrice float64 }
	type closedWant struct{ size, commission float64 }
	type openWant struct{ size, commission float64 }
	tests := []struct {
		name       string
		commType   string
		commRate   float64
		trades     []tradeInput
		closeQty   float64
		exitPrice  float64
		wantClosed []closedWant
		wantOpen   []openWant
	}{
		{
			name:     "cpc_first_full_second_partial",
			commType: CommissionCashPerContract, commRate: 1.0,
			trades:   []tradeInput{{6, 100}, {4, 105}},
			closeQty: 8, exitPrice: 110,
			wantClosed: []closedWant{{6, 12}, {2, 4}},
			wantOpen:   []openWant{{2, 2}},
		},
		{
			name:     "cpc_both_trades_full_close",
			commType: CommissionCashPerContract, commRate: 1.0,
			trades:   []tradeInput{{5, 100}, {5, 105}},
			closeQty: 10, exitPrice: 115,
			wantClosed: []closedWant{{5, 10}, {5, 10}},
		},
		{
			name:     "cpo_first_full_second_partial",
			commType: CommissionCashPerOrder, commRate: 5.0,
			trades:   []tradeInput{{6, 100}, {4, 105}},
			closeQty: 8, exitPrice: 110,
			wantClosed: []closedWant{{6, 8.75}, {2, 3.75}},
			wantOpen:   []openWant{{2, 2.5}},
		},
		{
			name:     "pct_first_full_second_partial",
			commType: CommissionPercent, commRate: 2.0,
			trades:   []tradeInput{{6, 100}, {4, 100}},
			closeQty: 8, exitPrice: 100,
			/* uniform price — avoids float64 precision loss in percent commission proportioning */
			wantClosed: []closedWant{{6, 24}, {2, 8}},
			wantOpen:   []openWant{{2, 4}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.Call("Test", 10000)
			s.SetCommission(tt.commRate, tt.commType)

			for i, trade := range tt.trades {
				s.Order("buy"+string(rune('1'+i)), Long, trade.qty, "trade"+string(rune('1'+i)))
				s.OnBarUpdate(i+1, trade.entryPrice, int64(1000*(i+1)))
			}
			baseBar := len(tt.trades) + 1
			s.Order("sell", Short, tt.closeQty, "")
			s.OnBarUpdate(baseBar, tt.exitPrice, int64(1000*baseBar))

			closed := s.GetTradeHistory().GetClosedTrades()
			if len(closed) != len(tt.wantClosed) {
				t.Fatalf("closed count: got %d, want %d", len(closed), len(tt.wantClosed))
			}
			for i, want := range tt.wantClosed {
				if closed[i].Size != want.size {
					t.Errorf("closed[%d].Size: got %.2f, want %.2f", i, closed[i].Size, want.size)
				}
				if closed[i].Commission != want.commission {
					t.Errorf("closed[%d].Commission: got %.4f, want %.4f", i, closed[i].Commission, want.commission)
				}
			}

			open := s.GetTradeHistory().GetOpenTrades()
			if len(open) != len(tt.wantOpen) {
				t.Fatalf("open count: got %d, want %d", len(open), len(tt.wantOpen))
			}
			for i, want := range tt.wantOpen {
				if open[i].Size != want.size {
					t.Errorf("open[%d].Size: got %.2f, want %.2f", i, open[i].Size, want.size)
				}
				if open[i].Commission != want.commission {
					t.Errorf("open[%d].Commission: got %.4f, want %.4f", i, open[i].Commission, want.commission)
				}
			}
		})
	}
}

/*
	TestCommission_ShortFullClose verifies commission symmetry on short positions: the commission

formula is direction-agnostic and produces the same structure as long-side commission.
*/
func TestCommission_ShortFullClose(t *testing.T) {
	tests := []struct {
		commType       string
		commRate       float64
		qty            float64
		entryPrice     float64
		exitPrice      float64
		wantCommission float64
	}{
		{CommissionPercent, 1.0, 8, 100, 90, 8*100*1.0/100 + 8*90*1.0/100},
		{CommissionCashPerOrder, 5.0, 8, 100, 90, 5.0 + 5.0},
		{CommissionCashPerContract, 2.0, 8, 100, 90, 8*2.0 + 8*2.0},
	}

	for _, tt := range tests {
		t.Run(tt.commType, func(t *testing.T) {
			s := NewStrategy()
			s.Call("Test", 10000)
			s.SetCommission(tt.commRate, tt.commType)

			s.Order("sell", Short, tt.qty, "")
			s.OnBarUpdate(1, tt.entryPrice, 1000)
			s.Order("buy", Long, tt.qty, "")
			s.OnBarUpdate(2, tt.exitPrice, 2000)

			closed := s.GetTradeHistory().GetClosedTrades()
			if len(closed) != 1 {
				t.Fatalf("expected 1 closed trade, got %d", len(closed))
			}
			if closed[0].Commission != tt.wantCommission {
				t.Errorf("commission: got %.4f, want %.4f", closed[0].Commission, tt.wantCommission)
			}
		})
	}
}

/*
	TestStrategyAllowedDirection_DirectionChange verifies that SetAllowedDirection can be called

multiple times to update the restriction mid-strategy, and the new restriction takes effect
immediately on the next call. Each phase uses a flat position to avoid reversal interaction.
*/
func TestStrategyAllowedDirection_DirectionChange(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("Test", 10000, 10)

	s.SetAllowedDirection(DirectionLong)
	s.Entry("shortBlocked", Short, 5, "")
	s.OnBarUpdate(1, 100, 1000)
	if len(s.GetTradeHistory().GetOpenTrades()) != 0 {
		t.Fatal("phase 1: short entry must be blocked when direction=long")
	}
	s.Entry("longAccepted", Long, 5, "")
	s.OnBarUpdate(2, 100, 2000)
	if len(s.GetTradeHistory().GetOpenTrades()) != 1 {
		t.Fatal("phase 1: long entry must be accepted when direction=long")
	}
	s.CloseAll(100, 2000, "")
	s.OnBarUpdate(3, 100, 3000)

	s.SetAllowedDirection(DirectionShort)
	s.Entry("longBlocked", Long, 5, "")
	s.OnBarUpdate(4, 100, 4000)
	if len(s.GetTradeHistory().GetOpenTrades()) != 0 {
		t.Error("phase 2: long entry must be blocked when direction=short")
	}
	s.Entry("shortAccepted", Short, 5, "")
	s.OnBarUpdate(5, 100, 5000)
	if len(s.GetTradeHistory().GetOpenTrades()) != 1 {
		t.Error("phase 2: short entry must be accepted when direction=short")
	}
	s.CloseAll(100, 5000, "")
	s.OnBarUpdate(6, 100, 6000)

	s.SetAllowedDirection(DirectionAll)
	s.Entry("longAll", Long, 5, "")
	s.OnBarUpdate(7, 100, 7000)
	if len(s.GetTradeHistory().GetOpenTrades()) != 1 {
		t.Error("phase 3: long entry must be accepted when direction=all")
	}
}

/*
	TestStrategyAllowedDirection_OrderFilter verifies that Order() enforces the allowed direction

filter across all direction combinations — parallel coverage to TestStrategyAllowedDirection_EntryFilter.
*/
func TestStrategyAllowedDirection_OrderFilter(t *testing.T) {
	tests := []struct {
		name             string
		allowedDirection string
		orderDirection   string
		wantTradeCreated bool
	}{
		{"long-only allows long", DirectionLong, Long, true},
		{"long-only blocks short", DirectionLong, Short, false},
		{"short-only allows short", DirectionShort, Short, true},
		{"short-only blocks long", DirectionShort, Long, false},
		{"all allows long", DirectionAll, Long, true},
		{"all allows short", DirectionAll, Short, true},
		{"empty direction allows long", "", Long, true},
		{"empty direction allows short", "", Short, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.Call("Test", 10000)
			if tt.allowedDirection != "" {
				s.SetAllowedDirection(tt.allowedDirection)
			}

			s.Order("e", tt.orderDirection, 10, "")
			s.OnBarUpdate(1, 100, 1000)

			open := s.GetTradeHistory().GetOpenTrades()
			got := len(open) > 0
			if got != tt.wantTradeCreated {
				t.Errorf("wantTradeCreated=%v, got %v (direction=%s, allowed=%q)", tt.wantTradeCreated, got, tt.orderDirection, tt.allowedDirection)
			}
		})
	}
}

func TestStrategy_SetDefaultQty(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)

	s.SetDefaultQty(50, QtyTypePercentOfEquity)

	if s.defaultQtyValue != 50 {
		t.Errorf("defaultQtyValue = %v, expected 50", s.defaultQtyValue)
	}
	if s.defaultQtyType != QtyTypePercentOfEquity {
		t.Errorf("defaultQtyType = %q, expected %q", s.defaultQtyType, QtyTypePercentOfEquity)
	}
}

func TestStrategy_DefaultEntryQty(t *testing.T) {
	tests := []struct {
		name       string
		qtyType    string
		qtyValue   float64
		fillPrice  float64
		initialCap float64
		expected   float64
	}{
		{
			name:       "fixed type returns qty value directly",
			qtyType:    QtyTypeFixed,
			qtyValue:   10,
			fillPrice:  100,
			initialCap: 10000,
			expected:   10,
		},
		{
			name:       "cash type divides by fill price",
			qtyType:    QtyTypeCash,
			qtyValue:   1000,
			fillPrice:  50,
			initialCap: 10000,
			expected:   20,
		},
		{
			name:       "percent_of_equity uses current equity",
			qtyType:    QtyTypePercentOfEquity,
			qtyValue:   10,
			fillPrice:  50,
			initialCap: 10000,
			expected:   20,
		},
		{
			name:       "zero fill price returns zero",
			qtyType:    QtyTypeCash,
			qtyValue:   1000,
			fillPrice:  0,
			initialCap: 10000,
			expected:   0,
		},
		{
			name:       "negative fill price returns zero",
			qtyType:    QtyTypeCash,
			qtyValue:   1000,
			fillPrice:  -50,
			initialCap: 10000,
			expected:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.Call("Test", tt.initialCap)
			s.SetDefaultQty(tt.qtyValue, tt.qtyType)

			result := s.DefaultEntryQty(tt.fillPrice)

			if result != tt.expected {
				t.Errorf("DefaultEntryQty(%v) = %v, expected %v", tt.fillPrice, result, tt.expected)
			}
		})
	}
}

func TestStrategy_DefaultEntryQty_DynamicEquity(t *testing.T) {
	s := NewStrategy()
	s.Call("Test", 10000)
	s.SetDefaultQty(10, QtyTypePercentOfEquity)

	result1 := s.DefaultEntryQty(100)
	expected1 := 10.0
	if result1 != expected1 {
		t.Errorf("initial equity: got %v, want %v", result1, expected1)
	}

	s.Entry("long", Long, 5, "")
	s.OnBarUpdate(1, 100, 1000)
	s.OnBarUpdate(2, 120, 2000)

	result2 := s.DefaultEntryQty(120)
	if result2 == result1 {
		t.Errorf("qty should change with equity, both=%v", result1)
	}
}

func TestStrategy_DefaultEntryQty_CommissionExcludedFromSizing(t *testing.T) {
	// Commission applies to profit only — qty is sized on raw fill price regardless of commission type or rate.
	calc := NewDefaultQtyCalculator()

	cases := []struct {
		name       string
		qtyType    string
		qtyValue   float64
		fillPrice  float64
		initialCap float64
		commType   string
		commRate   float64
	}{
		{"cash / percent commission / typical rate", QtyTypeCash, 1000, 50, 10000, CommissionPercent, 0.03},
		{"cash / percent commission / high rate", QtyTypeCash, 1000, 50, 10000, CommissionPercent, 50.0},
		{"percent-of-equity / percent commission", QtyTypePercentOfEquity, 10, 50, 10000, CommissionPercent, 0.03},
		{"cash / cash-per-contract commission", QtyTypeCash, 1000, 50, 10000, CommissionCashPerContract, 2.0},
		{"percent-of-equity / cash-per-contract commission", QtyTypePercentOfEquity, 10, 50, 10000, CommissionCashPerContract, 2.0},
		{"cash / cash-per-order commission", QtyTypeCash, 1000, 50, 10000, CommissionCashPerOrder, 5.0},
		{"percent-of-equity / cash-per-order commission", QtyTypePercentOfEquity, 10, 50, 10000, CommissionCashPerOrder, 5.0},
		{"fixed / percent commission", QtyTypeFixed, 7, 50, 10000, CommissionPercent, 10.0},
		{"cash / zero commission", QtyTypeCash, 1000, 50, 10000, CommissionPercent, 0.0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewStrategy()
			s.Call("Test", tc.initialCap)
			s.SetDefaultQty(tc.qtyValue, tc.qtyType)
			s.SetCommission(tc.commRate, tc.commType)

			want := calc.CalculateQty(tc.qtyType, tc.qtyValue, tc.fillPrice, tc.initialCap)
			got := s.DefaultEntryQty(tc.fillPrice)

			if got != want {
				t.Errorf("got %v, want %v — commission (type=%s rate=%.4f) must not enter qty denominator",
					got, want, tc.commType, tc.commRate)
			}
		})
	}
}

func TestStrategy_ConvertToAccount(t *testing.T) {
	s := NewStrategy()

	tests := []struct {
		name  string
		input float64
	}{
		{"positive value", 100},
		{"negative value", -50},
		{"zero", 0},
		{"large value", 1e10},
		{"small value", 1e-10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.ConvertToAccount(tt.input)
			if result != tt.input {
				t.Errorf("ConvertToAccount(%v) = %v, want %v", tt.input, result, tt.input)
			}
		})
	}
}

func TestStrategy_ConvertToSymbol(t *testing.T) {
	s := NewStrategy()

	tests := []struct {
		name  string
		input float64
	}{
		{"positive value", 200},
		{"negative value", -100},
		{"zero", 0},
		{"large value", 1e12},
		{"small value", 1e-12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.ConvertToSymbol(tt.input)
			if result != tt.input {
				t.Errorf("ConvertToSymbol(%v) = %v, want %v", tt.input, result, tt.input)
			}
		})
	}
}

func TestOrderManager_CreateEntryOrderWithDefaultQty(t *testing.T) {
	om := NewOrderManager()
	order := om.CreateEntryOrderWithDefaultQty("buy", Long, 3, "signal comment")

	if order.ID != "buy" {
		t.Errorf("ID: got %q, want %q", order.ID, "buy")
	}
	if order.Action != OrderActionEntry {
		t.Errorf("Action: got %q, want %q", order.Action, OrderActionEntry)
	}
	if order.Direction != Long {
		t.Errorf("Direction: got %q, want %q", order.Direction, Long)
	}
	if !order.UseDefaultQty {
		t.Errorf("UseDefaultQty: expected true")
	}
	if order.Qty != 0 {
		t.Errorf("Qty: expected 0 for deferred sizing, got %v", order.Qty)
	}
	if order.EntryComment != "signal comment" {
		t.Errorf("EntryComment: got %q, want %q", order.EntryComment, "signal comment")
	}
	if order.CreatedBar != 3 {
		t.Errorf("CreatedBar: got %d, want 3", order.CreatedBar)
	}

	// Same-ID create replaces the existing pending order (idempotent scheduling)
	om.CreateEntryOrderWithDefaultQty("buy", Short, 7, "updated")
	count := 0
	for _, o := range om.GetPendingOrders(10) {
		if o.ID == "buy" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("deduplication: expected 1 pending order for 'buy', got %d", count)
	}
}

func TestEntryWithDefaultQty_GuardBehavior(t *testing.T) {
	t.Run("not_initialized_returns_error", func(t *testing.T) {
		s := NewStrategy()
		if err := s.EntryWithDefaultQty("Buy", Long, ""); err == nil {
			t.Errorf("expected error when not initialized")
		}
	})

	t.Run("blocked_direction_skips_entry", func(t *testing.T) {
		s := NewStrategy()
		s.Call("Test", 10000)
		s.SetAllowedDirection(DirectionLong)

		if err := s.EntryWithDefaultQty("Short", Short, ""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, o := range s.orderManager.GetPendingOrders(s.currentBar + 1) {
			if o.ID == "Short" && o.Action == OrderActionEntry {
				t.Errorf("blocked direction should not create an entry order")
			}
		}
	})

	t.Run("allowed_direction_creates_deferred_order", func(t *testing.T) {
		s := NewStrategy()
		s.Call("Test", 10000)
		s.SetAllowedDirection(DirectionLong)

		if err := s.EntryWithDefaultQty("Buy", Long, ""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var found *Order
		for _, o := range s.orderManager.GetPendingOrders(s.currentBar + 1) {
			if o.ID == "Buy" && o.Action == OrderActionEntry {
				o := o
				found = &o
			}
		}
		if found == nil {
			t.Fatalf("expected pending entry order for 'Buy'")
		}
		if !found.UseDefaultQty {
			t.Errorf("UseDefaultQty: expected true")
		}
	})

	t.Run("pyramiding_cap_blocks_excess_same_direction", func(t *testing.T) {
		s := NewStrategy()
		s.CallWithPyramiding("Test", 10000, 1)
		s.SetDefaultQty(1, QtyTypeFixed)
		s.Entry("Buy1", Long, 1, "")
		s.OnBarUpdate(1, 100, 1000)

		if err := s.EntryWithDefaultQty("Buy2", Long, ""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, o := range s.orderManager.GetPendingOrders(s.currentBar + 1) {
			if o.ID == "Buy2" {
				t.Errorf("pyramiding cap=1 should block a second same-direction entry")
			}
		}
	})

	t.Run("pyramiding_cap_allows_within_limit", func(t *testing.T) {
		s := NewStrategy()
		s.CallWithPyramiding("Test", 10000, 2)
		s.SetDefaultQty(1, QtyTypeFixed)
		s.Entry("Buy1", Long, 1, "")
		s.OnBarUpdate(1, 100, 1000)

		if err := s.EntryWithDefaultQty("Buy2", Long, ""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var found bool
		for _, o := range s.orderManager.GetPendingOrders(s.currentBar + 1) {
			if o.ID == "Buy2" {
				found = true
			}
		}
		if !found {
			t.Errorf("pyramiding cap=2 should allow a second same-direction entry")
		}
	})

	t.Run("blocked_direction_still_schedules_opposite_close", func(t *testing.T) {
		// TV semantics: even a blocked entry must close any open opposite position
		s := NewStrategy()
		s.CallWithPyramiding("Test", 10000, 5)
		s.SetAllowedDirection(DirectionLong)
		s.Entry("Long", Long, 1, "")
		s.OnBarUpdate(1, 100, 1000)

		if err := s.EntryWithDefaultQty("Short", Short, ""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s.OnBarUpdate(2, 110, 2000)

		if len(s.GetTradeHistory().GetClosedTrades()) != 1 {
			t.Errorf("blocked entry should still close the opposite open trade")
		}
		if len(s.GetTradeHistory().GetOpenTrades()) != 0 {
			t.Errorf("no new short should have opened (direction blocked)")
		}
	})
}

func TestEntryWithDefaultQty_DeferredQtyAtFillPrice(t *testing.T) {
	tests := []struct {
		name       string
		qtyType    string
		qtyValue   float64
		initialCap float64
		openPrice  float64
		wantQty    float64
	}{
		{
			name:       "percent_of_equity_100pct",
			qtyType:    QtyTypePercentOfEquity,
			qtyValue:   100,
			initialCap: 10000,
			openPrice:  200,
			wantQty:    50,
		},
		{
			name:       "percent_of_equity_50pct",
			qtyType:    QtyTypePercentOfEquity,
			qtyValue:   50,
			initialCap: 10000,
			openPrice:  100,
			wantQty:    50,
		},
		{
			name:       "cash_type",
			qtyType:    QtyTypeCash,
			qtyValue:   5000,
			initialCap: 10000,
			openPrice:  250,
			wantQty:    20,
		},
		{
			name:       "higher_open_price_yields_smaller_qty",
			qtyType:    QtyTypePercentOfEquity,
			qtyValue:   100,
			initialCap: 10000,
			openPrice:  500,
			wantQty:    20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.Call("Test", tt.initialCap)
			s.SetDefaultQty(tt.qtyValue, tt.qtyType)

			if err := s.EntryWithDefaultQty("Buy", Long, ""); err != nil {
				t.Fatalf("EntryWithDefaultQty: %v", err)
			}
			s.OnBarUpdate(1, tt.openPrice, 1000)

			open := s.GetTradeHistory().GetOpenTrades()
			if len(open) != 1 {
				t.Fatalf("open trades: want 1, got %d", len(open))
			}
			if open[0].Size != tt.wantQty {
				t.Errorf("size: got %v, want %v", open[0].Size, tt.wantQty)
			}
			if open[0].EntryPrice != tt.openPrice {
				t.Errorf("entry price: got %v, want %v", open[0].EntryPrice, tt.openPrice)
			}
		})
	}
}

func TestEntryWithDefaultQty_PostReversalEquity(t *testing.T) {
	// Deferred qty for percent_of_equity uses the equity AFTER the reversal close
	// executes — not the pre-reversal equity that includes unrealized PnL of the
	// position being reversed. The fill price is deliberately different from the
	// signal-bar price so pre- and post-reversal equity values diverge.
	tests := []struct {
		name        string
		initialCap  float64
		initDir     string
		reversalDir string
		openQty     float64
		entryPrice  float64
		signalPrice float64 // currentPrice at signal bar (drives unrealized PnL)
		fillPrice   float64 // bar open at fill; reversal executes at this price
		wantQty     float64 // postReversalEquity / fillPrice
	}{
		{
			name:        "profitable_long_reversal",
			initialCap:  10000,
			initDir:     Long,
			reversalDir: Short,
			openQty:     1.0,
			entryPrice:  100.0,
			signalPrice: 150.0,
			fillPrice:   160.0,
			wantQty:     62.875,
		},
		{
			name:        "losing_long_reversal",
			initialCap:  10000,
			initDir:     Long,
			reversalDir: Short,
			openQty:     1.0,
			entryPrice:  100.0,
			signalPrice: 80.0,
			fillPrice:   70.0,
			wantQty:     9970.0 / 70.0,
		},
		{
			name:        "profitable_short_reversal",
			initialCap:  10000,
			initDir:     Short,
			reversalDir: Long,
			openQty:     1.0,
			entryPrice:  200.0,
			signalPrice: 180.0,
			fillPrice:   170.0,
			wantQty:     10030.0 / 170.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.Call("Test", tt.initialCap)
			s.SetDefaultQty(100, QtyTypePercentOfEquity)

			s.Entry("Init", tt.initDir, tt.openQty, "")
			s.OnBarUpdate(1, tt.entryPrice, 1000)

			// Signal bar: advance currentPrice to reflect unrealized PnL
			s.OnBarUpdate(2, tt.signalPrice, 2000)
			if err := s.EntryWithDefaultQty("Rev", tt.reversalDir, ""); err != nil {
				t.Fatalf("EntryWithDefaultQty: %v", err)
			}

			// Fill bar: HandleReversal closes Init position, then opens Rev with deferred qty
			s.OnBarUpdate(3, tt.fillPrice, 3000)

			closed := s.GetTradeHistory().GetClosedTrades()
			if len(closed) != 1 {
				t.Fatalf("closed trades: want 1, got %d", len(closed))
			}
			open := s.GetTradeHistory().GetOpenTrades()
			if len(open) != 1 {
				t.Fatalf("open trades: want 1, got %d", len(open))
			}
			if open[0].Direction != tt.reversalDir {
				t.Errorf("direction: got %s, want %s", open[0].Direction, tt.reversalDir)
			}
			if open[0].Size != tt.wantQty {
				t.Errorf("size: got %.6f, want %.6f", open[0].Size, tt.wantQty)
			}
		})
	}
}

// TestOnBarClose_NoOpWhenDisabled verifies that when process_orders_on_close is
// disabled, OnBarClose is a no-op and pending orders defer to bar N+1 open.
func TestOnBarClose_NoOpWhenDisabled(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("test", 10000, 0)

	const (
		barOpen    = 100.0
		closePrice = 90.0
		nextOpen   = 107.0
	)
	s.OnBarUpdate(0, barOpen, 1000)
	s.Entry("e1", Long, 1, "")
	s.OnBarMetrics(barOpen, barOpen+5, barOpen-5, 1000)
	s.OnBarClose(closePrice, 1000)

	if len(s.GetTradeHistory().GetOpenTrades()) != 0 {
		t.Error("OnBarClose must not fill orders when process_orders_on_close is false")
	}

	s.OnBarUpdate(1, nextOpen, 2000)
	s.OnBarMetrics(nextOpen, nextOpen+5, nextOpen-5, 2000)

	opens := s.GetTradeHistory().GetOpenTrades()
	if len(opens) != 1 {
		t.Fatalf("deferred entry: want 1 open trade after next bar, got %d", len(opens))
	}
	if opens[0].EntryPrice != nextOpen {
		t.Errorf("deferred entry price: got %.2f, want %.2f (next bar open)", opens[0].EntryPrice, nextOpen)
	}
	if opens[0].EntryBar != 1 {
		t.Errorf("deferred entry bar: got %d, want 1", opens[0].EntryBar)
	}
}

// TestOnBarClose_FillsEntryAtClosePrice verifies that an Entry order placed during
// bar N fills at bar N's close price and time when process_orders_on_close is
// enabled, for both long and short directions.
func TestOnBarClose_FillsEntryAtClosePrice(t *testing.T) {
	tests := []struct {
		name      string
		direction string
	}{
		{"long_entry_fills_at_close", Long},
		{"short_entry_fills_at_close", Short},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.CallWithPyramiding("test", 10000, 0)
			s.SetProcessOrdersOnClose(true)

			const closePrice = 105.0
			const barTime = int64(1000)

			s.OnBarUpdate(0, 100, barTime)
			s.Entry("e1", tt.direction, 2, "")
			s.OnBarMetrics(100, 110, 95, barTime)
			s.OnBarClose(closePrice, barTime)

			opens := s.GetTradeHistory().GetOpenTrades()
			if len(opens) != 1 {
				t.Fatalf("want 1 open trade, got %d", len(opens))
			}
			if opens[0].EntryPrice != closePrice {
				t.Errorf("entry price: got %.2f, want %.2f", opens[0].EntryPrice, closePrice)
			}
			if opens[0].EntryTime != barTime {
				t.Errorf("entry time: got %d, want %d (bar open time)", opens[0].EntryTime, barTime)
			}
			if opens[0].EntryBar != 0 {
				t.Errorf("entry bar: got %d, want 0 (same bar as signal)", opens[0].EntryBar)
			}
			if opens[0].Size != 2 {
				t.Errorf("size: got %.2f, want 2", opens[0].Size)
			}
			if opens[0].Direction != tt.direction {
				t.Errorf("direction: got %q, want %q", opens[0].Direction, tt.direction)
			}
		})
	}
}

// TestOnBarClose_CloseAllAtClosePrice verifies that CloseAll filled via OnBarClose
// uses the bar close price, for both long and short directions.
func TestOnBarClose_CloseAllAtClosePrice(t *testing.T) {
	tests := []struct {
		name      string
		direction string
	}{
		{"long_closeall_fills_at_close", Long},
		{"short_closeall_fills_at_close", Short},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s *Strategy
			if tt.direction == Long {
				s = newStrategyWithLongTrade(1, 100)
			} else {
				s = newStrategyWithShortTrade(1, 100)
			}
			s.SetProcessOrdersOnClose(true)

			const closePrice = 120.0
			const closeTime = int64(3000)
			s.OnBarUpdate(2, 110, closeTime)
			s.CloseAll(closePrice, closeTime, "")
			s.OnBarMetrics(110, 120, 110, closeTime)
			s.OnBarClose(closePrice, closeTime)

			closed := s.GetTradeHistory().GetClosedTrades()
			if len(closed) != 1 {
				t.Fatalf("want 1 closed trade, got %d", len(closed))
			}
			if closed[0].ExitPrice != closePrice {
				t.Errorf("exit price: got %.2f, want %.2f", closed[0].ExitPrice, closePrice)
			}
			if closed[0].ExitBar != 2 {
				t.Errorf("exit bar: got %d, want 2 (same bar as signal)", closed[0].ExitBar)
			}
		})
	}
}

/* TestOnBarClose_ReversalSameBarTimestamp verifies that when a strategy reverses direction
 * under process_orders_on_close, the exit and new entry share the same bar timestamp. */
func TestOnBarClose_ReversalSameBarTimestamp(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("test", 10000, 0)
	s.SetProcessOrdersOnClose(true)

	s.OnBarUpdate(0, 100, 1000)
	s.Entry("long1", Long, 1, "")
	s.OnBarMetrics(100, 100, 100, 1000)
	s.OnBarClose(100, 1000)

	const reverseTime = int64(2000)
	s.OnBarUpdate(1, 110, reverseTime)
	s.CloseAll(110, reverseTime, "")
	s.Entry("short1", Short, 1, "")
	s.OnBarMetrics(110, 110, 110, reverseTime)
	s.OnBarClose(110, reverseTime)

	closed := s.GetTradeHistory().GetClosedTrades()
	opens := s.GetTradeHistory().GetOpenTrades()
	if len(closed) != 1 || len(opens) != 1 {
		t.Fatalf("want 1 closed + 1 open, got %d closed %d open", len(closed), len(opens))
	}
	if closed[0].ExitTime != reverseTime {
		t.Errorf("exit time: got %d, want %d", closed[0].ExitTime, reverseTime)
	}
	if opens[0].EntryTime != reverseTime {
		t.Errorf("entry time: got %d, want %d", opens[0].EntryTime, reverseTime)
	}
}

/* TestGetCurrentBarOrders verifies that GetCurrentBarOrders filters by bar index. */
func TestGetCurrentBarOrders(t *testing.T) {
	om := NewOrderManager()
	om.CreateEntryOrder("a", Long, 1, 0, "")
	om.CreateEntryOrder("b", Long, 1, 1, "")
	om.CreateCloseAllOrder(1, "") // bar 1, auto-ID
	om.CreateEntryOrder("d", Long, 1, 2, "")

	bar1Orders := om.GetCurrentBarOrders(1)
	if len(bar1Orders) != 2 {
		t.Fatalf("want 2 orders for bar 1, got %d", len(bar1Orders))
	}
	// bar 0 and bar 2 must not appear
	for _, o := range bar1Orders {
		if o.CreatedBar != 1 {
			t.Errorf("unexpected order from bar %d in bar-1 result: %+v", o.CreatedBar, o)
		}
	}
	// Bar 0 and bar 2 orders must still be retrievable
	if len(om.GetCurrentBarOrders(0)) != 1 {
		t.Errorf("expected 1 order for bar 0")
	}
	if len(om.GetCurrentBarOrders(2)) != 1 {
		t.Errorf("expected 1 order for bar 2")
	}
}

/* TestOnBarClose_ShortEntryFillsAtClosePrice verifies that a Short entry order
 * placed during bar N is filled at bar N's close price when process_orders_on_close
 * is enabled, mirroring the Long direction test. */
func TestOnBarClose_ShortEntryFillsAtClosePrice(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("test", 10000, 0)
	s.SetProcessOrdersOnClose(true)

	const closePrice = 95.0
	const barTime = int64(1000)

	s.OnBarUpdate(0, 100, barTime)
	s.Entry("e1", Short, 2, "")
	s.OnBarMetrics(100, 100, 90, barTime)
	s.OnBarClose(closePrice, barTime)

	opens := s.GetTradeHistory().GetOpenTrades()
	if len(opens) != 1 {
		t.Fatalf("want 1 open short trade, got %d", len(opens))
	}
	if opens[0].Direction != Short {
		t.Errorf("direction: got %q, want %q", opens[0].Direction, Short)
	}
	if opens[0].EntryPrice != closePrice {
		t.Errorf("entry price: got %.2f, want %.2f", opens[0].EntryPrice, closePrice)
	}
	if opens[0].Size != 2 {
		t.Errorf("size: got %.2f, want 2", opens[0].Size)
	}
}

/* TestOnBarClose_CloseSpecificEntry verifies that strategy.close("id") queued during bar N
 * is filled at bar N's close price when process_orders_on_close is enabled, while leaving
 * other open trades untouched. */
func TestOnBarClose_CloseSpecificEntry(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("test", 10000, 2) // pyramiding=2 allows two same-direction concurrent open trades
	s.SetProcessOrdersOnClose(true)

	// Bar 0: open two long positions.
	s.OnBarUpdate(0, 100, 1000)
	s.Entry("e1", Long, 1, "")
	s.Entry("e2", Long, 1, "")
	s.OnBarMetrics(100, 100, 100, 1000)
	s.OnBarClose(100, 1000)

	if got := len(s.GetTradeHistory().GetOpenTrades()); got != 2 {
		t.Fatalf("setup: want 2 open trades after bar 0, got %d", got)
	}

	// Bar 1: close only e1; e2 must remain open.
	const exitPrice = 120.0
	const exitTime = int64(2000)
	s.OnBarUpdate(1, 110, exitTime)
	s.Close("e1", 0, 0, "")
	s.OnBarMetrics(110, 125, 110, exitTime)
	s.OnBarClose(exitPrice, exitTime)

	closed := s.GetTradeHistory().GetClosedTrades()
	opens := s.GetTradeHistory().GetOpenTrades()
	if len(closed) != 1 {
		t.Fatalf("want 1 closed trade, got %d", len(closed))
	}
	if len(opens) != 1 {
		t.Fatalf("want 1 remaining open trade (e2), got %d", len(opens))
	}
	if closed[0].ExitPrice != exitPrice {
		t.Errorf("exit price: got %.2f, want %.2f", closed[0].ExitPrice, exitPrice)
	}
	if closed[0].EntryID != "e1" {
		t.Errorf("closed trade entry ID: got %q, want %q", closed[0].EntryID, "e1")
	}
	if opens[0].EntryID != "e2" {
		t.Errorf("remaining open trade ID: got %q, want %q", opens[0].EntryID, "e2")
	}
}

/* TestOnBarClose_NetOrderFillsAtClosePrice verifies that strategy.order() (net order)
 * queued during bar N is filled at bar N's close price when process_orders_on_close is
 * enabled, exercising the OrderActionOrder path in OnBarClose. */
func TestOnBarClose_NetOrderFillsAtClosePrice(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("test", 10000, 0)
	s.SetProcessOrdersOnClose(true)

	const closePrice = 110.0
	const barTime = int64(1000)

	s.OnBarUpdate(0, 100, barTime)
	s.Order("net1", Long, 3, "")
	s.OnBarMetrics(100, 115, 100, barTime)
	s.OnBarClose(closePrice, barTime)

	opens := s.GetTradeHistory().GetOpenTrades()
	if len(opens) != 1 {
		t.Fatalf("want 1 open trade from net order, got %d", len(opens))
	}
	if opens[0].EntryPrice != closePrice {
		t.Errorf("net order fill price: got %.2f, want %.2f", opens[0].EntryPrice, closePrice)
	}
	if opens[0].Size != 3 {
		t.Errorf("net order size: got %.2f, want 3", opens[0].Size)
	}
}

/* TestOnBarClose_MultipleOrdersSameBar verifies that when two entry orders are placed
 * on the same bar, both are processed by OnBarClose and filled at the bar close price. */
func TestOnBarClose_MultipleOrdersSameBar(t *testing.T) {
	s := NewStrategy()
	s.CallWithPyramiding("test", 10000, 2) // pyramiding=2: two simultaneous same-direction entries allowed
	s.SetProcessOrdersOnClose(true)

	const closePrice = 105.0
	const barTime = int64(1000)

	s.OnBarUpdate(0, 100, barTime)
	s.Entry("e1", Long, 1, "")
	s.Entry("e2", Long, 1, "")
	s.OnBarMetrics(100, 110, 100, barTime)
	s.OnBarClose(closePrice, barTime)

	opens := s.GetTradeHistory().GetOpenTrades()
	if len(opens) != 2 {
		t.Fatalf("want 2 open trades after two same-bar entries, got %d", len(opens))
	}
	for _, tr := range opens {
		if tr.EntryPrice != closePrice {
			t.Errorf("trade %q: entry price got %.2f, want %.2f", tr.EntryID, tr.EntryPrice, closePrice)
		}
		if tr.EntryTime != barTime {
			t.Errorf("trade %q: entry time got %d, want %d", tr.EntryID, tr.EntryTime, barTime)
		}
	}
}

/* TestOnBarClose_NotInitializedIsNoOp verifies that OnBarClose returns without
 * panicking or modifying strategy state when the strategy has not yet been
 * initialized via CallWithPyramiding, even if process_orders_on_close is set. */
func TestOnBarClose_NotInitializedIsNoOp(t *testing.T) {
	s := NewStrategy()
	s.SetProcessOrdersOnClose(true)

	// Must not panic.
	s.OnBarClose(100, 1000)

	if s.GetPositionSize() != 0 {
		t.Errorf("position size after pre-init OnBarClose: got %.2f, want 0", s.GetPositionSize())
	}
	if len(s.GetTradeHistory().GetOpenTrades()) != 0 {
		t.Errorf("open trades after pre-init OnBarClose: got %d, want 0", len(s.GetTradeHistory().GetOpenTrades()))
	}
}

/* TestGetCurrentBarOrders_EmptyManager verifies that an order manager with no orders
 * returns an empty (not nil) slice for any bar index. */
func TestGetCurrentBarOrders_EmptyManager(t *testing.T) {
	om := NewOrderManager()
	orders := om.GetCurrentBarOrders(0)
	if len(orders) != 0 {
		t.Errorf("want 0 orders from empty manager, got %d", len(orders))
	}
}

/* TestGetCurrentBarOrders_NoOrdersForBar verifies that GetCurrentBarOrders returns
 * no orders when the requested bar index has no orders, even if orders exist for
 * other bars. */
func TestGetCurrentBarOrders_NoOrdersForBar(t *testing.T) {
	om := NewOrderManager()
	om.CreateEntryOrder("a", Long, 1, 0, "")
	om.CreateEntryOrder("b", Long, 1, 2, "")

	// Bar 1 has no orders; bars 0 and 2 do.
	orders := om.GetCurrentBarOrders(1)
	if len(orders) != 0 {
		t.Errorf("want 0 orders for bar 1, got %d", len(orders))
	}
}

func TestEntryFill_PriceIsNextBarOpen(t *testing.T) {
	tests := []struct {
		name          string
		signalBarOpen float64
		fillBarOpen   float64
		direction     string
	}{
		{"long_no_gap", 100, 100, Long},
		{"long_gap_down", 100, 90, Long},
		{"long_gap_up", 100, 115, Long},
		{"short_no_gap", 200, 200, Short},
		{"short_gap_down", 200, 185, Short},
		{"short_gap_up", 200, 210, Short},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.CallWithPyramiding("test", 10000, 0)

			s.OnBarUpdate(0, tt.signalBarOpen, 1000)
			if err := s.Entry("e1", tt.direction, 1, ""); err != nil {
				t.Fatalf("Entry: %v", err)
			}

			s.OnBarUpdate(1, tt.fillBarOpen, 2000)

			open := s.GetTradeHistory().GetOpenTrades()
			if len(open) != 1 {
				t.Fatalf("open trades: want 1, got %d", len(open))
			}
			if open[0].EntryPrice != tt.fillBarOpen {
				t.Errorf("entry price: got %.4f, want %.4f (bar N+1 open, not signal bar price or prior close)",
					open[0].EntryPrice, tt.fillBarOpen)
			}
		})
	}
}

func TestEntryFill_TimeIsNextBarOpenTime(t *testing.T) {
	const signalBarTime = int64(1000)
	const fillBarTime = int64(3600)

	s := NewStrategy()
	s.CallWithPyramiding("test", 10000, 0)

	s.OnBarUpdate(0, 100, signalBarTime)
	if err := s.Entry("e1", Long, 1, ""); err != nil {
		t.Fatalf("Entry: %v", err)
	}
	s.OnBarUpdate(1, 105, fillBarTime)

	open := s.GetTradeHistory().GetOpenTrades()
	if len(open) != 1 {
		t.Fatalf("open trades: want 1, got %d", len(open))
	}
	if open[0].EntryTime != fillBarTime {
		t.Errorf("entry time: got %d, want %d (fill bar open time, not signal bar time %d)",
			open[0].EntryTime, fillBarTime, signalBarTime)
	}
}

func TestExitFill_PriceIsNextBarOpen(t *testing.T) {
	tests := []struct {
		name        string
		fillBarOpen float64
		direction   string
	}{
		{"long_exit_no_gap", 110, Long},
		{"long_exit_gap_down", 95, Long},
		{"long_exit_gap_up", 130, Long},
		{"short_exit_no_gap", 90, Short},
		{"short_exit_gap_down", 75, Short},
		{"short_exit_gap_up", 110, Short},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.CallWithPyramiding("test", 10000, 0)

			s.OnBarUpdate(0, 100, 1000)
			if err := s.Entry("e1", tt.direction, 1, ""); err != nil {
				t.Fatalf("Entry: %v", err)
			}
			s.OnBarUpdate(1, 100, 2000)
			s.Close("e1", 0, 3000, "")
			s.OnBarUpdate(2, tt.fillBarOpen, 3000)

			closed := s.GetTradeHistory().GetClosedTrades()
			if len(closed) != 1 {
				t.Fatalf("closed trades: want 1, got %d", len(closed))
			}
			if closed[0].ExitPrice != tt.fillBarOpen {
				t.Errorf("exit price: got %.4f, want %.4f (fill bar open, not signal price or prior close)",
					closed[0].ExitPrice, tt.fillBarOpen)
			}
		})
	}
}

func TestCloseAllFill_PriceIsNextBarOpen(t *testing.T) {
	tests := []struct {
		name        string
		fillBarOpen float64
	}{
		{"no_gap", 100},
		{"gap_down", 85},
		{"gap_up", 120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.CallWithPyramiding("test", 10000, 5)

			s.OnBarUpdate(0, 100, 1000)
			if err := s.Entry("e1", Long, 1, ""); err != nil {
				t.Fatalf("Entry e1: %v", err)
			}
			s.OnBarUpdate(1, 100, 2000)
			if err := s.Entry("e2", Long, 1, ""); err != nil {
				t.Fatalf("Entry e2: %v", err)
			}
			s.OnBarUpdate(2, 100, 3000)
			s.CloseAll(0, 4000, "")
			s.OnBarUpdate(3, tt.fillBarOpen, 4000)

			closed := s.GetTradeHistory().GetClosedTrades()
			if len(closed) != 2 {
				t.Fatalf("closed trades: want 2, got %d", len(closed))
			}
			for _, c := range closed {
				if c.ExitPrice != tt.fillBarOpen {
					t.Errorf("exit price for %s: got %.4f, want %.4f (fill bar open)",
						c.EntryID, c.ExitPrice, tt.fillBarOpen)
				}
			}
		})
	}
}

func TestEntryFill_ConsecutiveGapBars(t *testing.T) {
	barOpens := []float64{100, 90, 115, 85, 120}
	wantFills := barOpens[1:]

	s := NewStrategy()
	s.CallWithPyramiding("test", 1000000, 10)

	for i, openPrice := range barOpens {
		s.OnBarUpdate(i, openPrice, int64(i*3600))
		if i < len(barOpens)-1 {
			id := fmt.Sprintf("e%d", i)
			if err := s.Entry(id, Long, 1, ""); err != nil {
				t.Fatalf("bar %d: Entry %s: %v", i, id, err)
			}
		}
	}

	open := s.GetTradeHistory().GetOpenTrades()
	if len(open) != len(wantFills) {
		t.Fatalf("open trades: want %d, got %d", len(wantFills), len(open))
	}
	for idx, trade := range open {
		if trade.EntryPrice != wantFills[idx] {
			t.Errorf("trade[%d] entry price: got %.4f, want %.4f (bar %d open)",
				idx, trade.EntryPrice, wantFills[idx], idx+1)
		}
	}
}
