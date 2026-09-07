package strategy

import (
	"testing"
)

func TestPositionReversal_Direction(t *testing.T) {
	tests := []struct {
		name          string
		firstDir      string
		firstQty      float64
		firstPrice    float64
		secondDir     string
		secondQty     float64
		secondPrice   float64
		wantPositon   float64
		wantClosed    int
		wantNetProfit float64
	}{
		{
			name:          "long_to_short_equal_quantity",
			firstDir:      Long,
			firstQty:      1.0,
			firstPrice:    100.0,
			secondDir:     Short,
			secondQty:     1.0,
			secondPrice:   110.0,
			wantPositon:   -1.0,
			wantClosed:    1,
			wantNetProfit: 10.0,
		},
		{
			name:          "short_to_long_equal_quantity",
			firstDir:      Short,
			firstQty:      1.0,
			firstPrice:    100.0,
			secondDir:     Long,
			secondQty:     1.0,
			secondPrice:   95.0,
			wantPositon:   1.0,
			wantClosed:    1,
			wantNetProfit: 5.0,
		},
		{
			name:          "long_to_short_larger_quantity",
			firstDir:      Long,
			firstQty:      2.0,
			firstPrice:    100.0,
			secondDir:     Short,
			secondQty:     5.0,
			secondPrice:   105.0,
			wantPositon:   -5.0,
			wantClosed:    1,
			wantNetProfit: 10.0,
		},
		{
			name:          "short_to_long_smaller_quantity",
			firstDir:      Short,
			firstQty:      10.0,
			firstPrice:    100.0,
			secondDir:     Long,
			secondQty:     1.0,
			secondPrice:   90.0,
			wantPositon:   1.0,
			wantClosed:    1,
			wantNetProfit: 100.0,
		},
		{
			name:          "long_to_short_with_loss",
			firstDir:      Long,
			firstQty:      1.0,
			firstPrice:    100.0,
			secondDir:     Short,
			secondQty:     1.0,
			secondPrice:   90.0,
			wantPositon:   -1.0,
			wantClosed:    1,
			wantNetProfit: -10.0,
		},
		{
			name:          "short_to_long_with_loss",
			firstDir:      Short,
			firstQty:      1.0,
			firstPrice:    100.0,
			secondDir:     Long,
			secondQty:     1.0,
			secondPrice:   110.0,
			wantPositon:   1.0,
			wantClosed:    1,
			wantNetProfit: -10.0,
		},
		{
			name:          "long_to_short_zero_profit",
			firstDir:      Long,
			firstQty:      1.0,
			firstPrice:    100.0,
			secondDir:     Short,
			secondQty:     1.0,
			secondPrice:   100.0,
			wantPositon:   -1.0,
			wantClosed:    1,
			wantNetProfit: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.Call("Position Reversal Test", 10000)

			s.Entry("First", tt.firstDir, tt.firstQty, "")
			s.OnBarUpdate(1, tt.firstPrice, 1001)

			s.Entry("Second", tt.secondDir, tt.secondQty, "")
			s.OnBarUpdate(2, tt.secondPrice, 1002)

			if s.GetPositionSize() != tt.wantPositon {
				t.Errorf("Position size: expected %.1f, got %.1f", tt.wantPositon, s.GetPositionSize())
			}

			closedTrades := s.tradeHistory.GetClosedTrades()
			if len(closedTrades) != tt.wantClosed {
				t.Errorf("Closed trades: expected %d, got %d", tt.wantClosed, len(closedTrades))
			}

			netProfit := s.GetNetProfit()
			if netProfit != tt.wantNetProfit {
				t.Errorf("Net profit: expected %.1f, got %.1f", tt.wantNetProfit, netProfit)
			}

			openTrades := s.tradeHistory.GetOpenTrades()
			if len(openTrades) != 1 {
				t.Errorf("Open trades: expected 1, got %d", len(openTrades))
			}
			if len(openTrades) > 0 && openTrades[0].Direction != tt.secondDir {
				t.Errorf("Open trade direction: expected %s, got %s", tt.secondDir, openTrades[0].Direction)
			}
		})
	}
}

func TestPositionReversal_MultiplePositions(t *testing.T) {
	tests := []struct {
		name         string
		initialDir   string
		entries      []float64
		entryPrice   float64
		reverseDir   string
		reverseQty   float64
		reversePrice float64
		wantPosition float64
		wantClosed   int
	}{
		{
			name:         "two_positions_reversed",
			initialDir:   Long,
			entries:      []float64{1.0, 2.0},
			entryPrice:   100.0,
			reverseDir:   Short,
			reverseQty:   1.0,
			reversePrice: 105.0,
			wantPosition: -1.0,
			wantClosed:   2,
		},
		{
			name:         "three_positions_reversed",
			initialDir:   Short,
			entries:      []float64{1.0, 1.0, 1.0},
			entryPrice:   100.0,
			reverseDir:   Long,
			reverseQty:   5.0,
			reversePrice: 95.0,
			wantPosition: 5.0,
			wantClosed:   3,
		},
		{
			name:         "five_positions_reversed",
			initialDir:   Long,
			entries:      []float64{1.0, 1.0, 1.0, 1.0, 1.0},
			entryPrice:   100.0,
			reverseDir:   Short,
			reverseQty:   2.0,
			reversePrice: 110.0,
			wantPosition: -2.0,
			wantClosed:   5,
		},
		{
			name:         "varying_quantities_reversed",
			initialDir:   Long,
			entries:      []float64{1.0, 2.0, 3.0},
			entryPrice:   100.0,
			reverseDir:   Short,
			reverseQty:   10.0,
			reversePrice: 105.0,
			wantPosition: -10.0,
			wantClosed:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.Call("Multiple Positions Test", 10000)

			for i, qty := range tt.entries {
				s.Entry("Entry"+string(rune(i+'A')), tt.initialDir, qty, "")
			}
			s.OnBarUpdate(1, tt.entryPrice, 1001)

			openBefore := len(s.tradeHistory.GetOpenTrades())
			if openBefore != len(tt.entries) {
				t.Fatalf("Open trades before reversal: expected %d, got %d", len(tt.entries), openBefore)
			}

			s.Entry("Reverse", tt.reverseDir, tt.reverseQty, "")
			s.OnBarUpdate(2, tt.reversePrice, 1002)

			if s.GetPositionSize() != tt.wantPosition {
				t.Errorf("Position size: expected %.1f, got %.1f", tt.wantPosition, s.GetPositionSize())
			}

			closedTrades := s.tradeHistory.GetClosedTrades()
			if len(closedTrades) != tt.wantClosed {
				t.Errorf("Closed trades: expected %d, got %d", tt.wantClosed, len(closedTrades))
			}

			if len(closedTrades) != openBefore {
				t.Errorf("All positions should close: expected %d closed, got %d", openBefore, len(closedTrades))
			}

			openTrades := s.tradeHistory.GetOpenTrades()
			if len(openTrades) != 1 {
				t.Errorf("Open trades after reversal: expected 1, got %d", len(openTrades))
			}
		})
	}
}

func TestPositionReversal_SameDirection(t *testing.T) {
	tests := []struct {
		name       string
		direction  string
		entries    []float64
		wantSize   float64
		wantClosed int
		wantOpen   int
	}{
		{
			name:       "multiple_long_entries",
			direction:  Long,
			entries:    []float64{1.0, 2.0},
			wantSize:   3.0,
			wantClosed: 0,
			wantOpen:   2,
		},
		{
			name:       "multiple_short_entries",
			direction:  Short,
			entries:    []float64{1.0, 1.0, 1.0},
			wantSize:   -3.0,
			wantClosed: 0,
			wantOpen:   3,
		},
		{
			name:       "varying_long_quantities",
			direction:  Long,
			entries:    []float64{5.0, 10.0, 2.0},
			wantSize:   17.0,
			wantClosed: 0,
			wantOpen:   3,
		},
		{
			name:       "ten_short_entries",
			direction:  Short,
			entries:    []float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0},
			wantSize:   -10.0,
			wantClosed: 0,
			wantOpen:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.Call("Same Direction Test", 10000)

			for i, qty := range tt.entries {
				s.Entry("Entry"+string(rune(i+'0')), tt.direction, qty, "")
				s.OnBarUpdate(i+1, 100.0+float64(i), int64(1001+i))
			}

			if s.GetPositionSize() != tt.wantSize {
				t.Errorf("Position size: expected %.1f, got %.1f", tt.wantSize, s.GetPositionSize())
			}

			closedTrades := s.tradeHistory.GetClosedTrades()
			if len(closedTrades) != tt.wantClosed {
				t.Errorf("Closed trades: expected %d, got %d", tt.wantClosed, len(closedTrades))
			}

			openTrades := s.tradeHistory.GetOpenTrades()
			if len(openTrades) != tt.wantOpen {
				t.Errorf("Open trades: expected %d, got %d", tt.wantOpen, len(openTrades))
			}
		})
	}
}

func TestPositionReversal_EquityTracking(t *testing.T) {
	tests := []struct {
		name         string
		initialCap   float64
		firstEntry   float64
		firstPrice   float64
		reversePrice float64
		wantClosedPL float64
		wantEquity   float64
	}{
		{
			name:         "profitable_reversal",
			initialCap:   10000,
			firstEntry:   1.0,
			firstPrice:   100.0,
			reversePrice: 150.0,
			wantClosedPL: 50.0,
			wantEquity:   10050.0,
		},
		{
			name:         "losing_reversal",
			initialCap:   10000,
			firstEntry:   1.0,
			firstPrice:   100.0,
			reversePrice: 80.0,
			wantClosedPL: -20.0,
			wantEquity:   9980.0,
		},
		{
			name:         "breakeven_reversal",
			initialCap:   10000,
			firstEntry:   1.0,
			firstPrice:   100.0,
			reversePrice: 100.0,
			wantClosedPL: 0.0,
			wantEquity:   10000.0,
		},
		{
			name:         "large_position_reversal",
			initialCap:   10000,
			firstEntry:   10.0,
			firstPrice:   100.0,
			reversePrice: 120.0,
			wantClosedPL: 200.0,
			wantEquity:   10200.0,
		},
		{
			name:         "small_capital_reversal",
			initialCap:   1000,
			firstEntry:   0.5,
			firstPrice:   50.0,
			reversePrice: 60.0,
			wantClosedPL: 5.0,
			wantEquity:   1005.0,
		},
		{
			name:         "large_capital_reversal",
			initialCap:   100000,
			firstEntry:   100.0,
			firstPrice:   200.0,
			reversePrice: 210.0,
			wantClosedPL: 1000.0,
			wantEquity:   101000.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.Call("Equity Tracking Test", tt.initialCap)

			s.Entry("Long", Long, tt.firstEntry, "")
			s.OnBarUpdate(1, tt.firstPrice, 1001)

			s.Entry("Short", Short, tt.firstEntry, "")
			s.OnBarUpdate(2, tt.reversePrice, 1002)

			closedTrades := s.tradeHistory.GetClosedTrades()
			if len(closedTrades) != 1 {
				t.Fatalf("Closed trades: expected 1, got %d", len(closedTrades))
			}

			if closedTrades[0].Profit != tt.wantClosedPL {
				t.Errorf("Trade P&L: expected %.1f, got %.1f", tt.wantClosedPL, closedTrades[0].Profit)
			}

			netProfit := s.GetNetProfit()
			if netProfit != tt.wantClosedPL {
				t.Errorf("Net profit: expected %.1f, got %.1f", tt.wantClosedPL, netProfit)
			}

			equity := s.GetEquity(tt.reversePrice)
			if equity != tt.wantEquity {
				t.Errorf("Equity: expected %.1f, got %.1f", tt.wantEquity, equity)
			}
		})
	}
}

func TestPositionReversal_ExplicitClose(t *testing.T) {
	s := NewStrategy()
	s.Call("Explicit Close Test", 10000)

	s.Entry("Long", Long, 1.0, "")
	s.OnBarUpdate(1, 100.0, 1001)

	s.Close("Long", 110.0, 1002, "Manual exit")
	s.OnBarUpdate(2, 110.0, 1002)

	if s.GetPositionSize() != 0.0 {
		t.Errorf("Position after close: expected 0.0, got %.1f", s.GetPositionSize())
	}

	closedBefore := len(s.tradeHistory.GetClosedTrades())

	s.Entry("Short", Short, 1.0, "")
	s.OnBarUpdate(3, 110.0, 1003)

	if s.GetPositionSize() != -1.0 {
		t.Errorf("Position after short entry: expected -1.0, got %.1f", s.GetPositionSize())
	}

	closedAfter := len(s.tradeHistory.GetClosedTrades())
	if closedAfter != closedBefore {
		t.Errorf("No additional closes expected: got %d closed trades", closedAfter)
	}
}

func TestPositionReversal_Sequential(t *testing.T) {
	tests := []struct {
		name       string
		sequence   []string
		prices     []float64
		wantClosed int
		wantPL     float64
	}{
		{
			name:       "alternating_directions",
			sequence:   []string{Long, Short, Long, Short, Long},
			prices:     []float64{100, 110, 105, 115, 110},
			wantClosed: 4,
			wantPL:     30.0, // (110-100) + (110-105) + (115-105) + (115-110)
		},
		{
			name:       "profit_series",
			sequence:   []string{Long, Short, Long, Short},
			prices:     []float64{100, 105, 100, 95},
			wantClosed: 3,
			wantPL:     5.0, // (105-100) + (105-100) + (95-100)
		},
		{
			name:       "loss_series",
			sequence:   []string{Long, Short, Long, Short},
			prices:     []float64{100, 95, 100, 105},
			wantClosed: 3,
			wantPL:     -5.0, // (95-100) + (95-100) + (105-100)
		},
		{
			name:       "mixed_profit_loss",
			sequence:   []string{Long, Short, Long},
			prices:     []float64{100, 110, 95},
			wantClosed: 2,
			wantPL:     25.0, // (110-100) + (110-95)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.Call("Sequential Reversal Test", 10000)

			for i, dir := range tt.sequence {
				s.Entry("Entry", dir, 1.0, "")
				s.OnBarUpdate(i+1, tt.prices[i], int64(1001+i))

				expectedSize := 1.0
				if dir == Short {
					expectedSize = -1.0
				}
				if s.GetPositionSize() != expectedSize {
					t.Errorf("Bar %d position: expected %.1f, got %.1f", i+1, expectedSize, s.GetPositionSize())
				}
			}

			closedTrades := s.tradeHistory.GetClosedTrades()
			if len(closedTrades) != tt.wantClosed {
				t.Errorf("Closed trades: expected %d, got %d", tt.wantClosed, len(closedTrades))
			}

			netProfit := s.GetNetProfit()
			if netProfit != tt.wantPL {
				t.Errorf("Cumulative P&L: expected %.1f, got %.1f", tt.wantPL, netProfit)
			}
		})
	}
}

func TestPositionReversal_EdgeCases(t *testing.T) {
	t.Run("zero_quantity_entry", func(t *testing.T) {
		s := NewStrategy()
		s.Call("Edge Case Test", 10000)

		s.Entry("Long", Long, 1.0, "")
		s.OnBarUpdate(1, 100.0, 1001)

		s.Entry("Short", Short, 0.0, "")
		s.OnBarUpdate(2, 105.0, 1002)

		if len(s.tradeHistory.GetClosedTrades()) != 1 {
			t.Error("Zero quantity should still trigger reversal closure")
		}
	})

	t.Run("immediate_reversal", func(t *testing.T) {
		s := NewStrategy()
		s.Call("Immediate Reversal Test", 10000)

		s.Entry("Long", Long, 1.0, "")
		s.Entry("Short", Short, 1.0, "")
		s.OnBarUpdate(1, 100.0, 1001)

		if s.GetPositionSize() != -1.0 {
			t.Errorf("Immediate reversal position: expected -1.0, got %.1f", s.GetPositionSize())
		}
	})

	t.Run("fractional_quantities", func(t *testing.T) {
		s := NewStrategy()
		s.Call("Fractional Quantity Test", 10000)

		s.Entry("Long", Long, 0.5, "")
		s.OnBarUpdate(1, 100.0, 1001)

		s.Entry("Short", Short, 0.25, "")
		s.OnBarUpdate(2, 110.0, 1002)

		if s.GetPositionSize() != -0.25 {
			t.Errorf("Fractional position: expected -0.25, got %.2f", s.GetPositionSize())
		}
	})
}

func TestPositionReversal_Commission(t *testing.T) {
	// Verifies that a position reversal via strategy.entry correctly charges:
	//   - the exit commission on the trade being closed (same as explicit close)
	//   - the entry commission on the new opposing trade
	// This invariant holds for all three commission models.
	tests := []struct {
		name           string
		commType       string
		commRate       float64
		initDir        string
		reversalDir    string
		openQty        float64
		entryPrice     float64
		reversalQty    float64
		reversalPrice  float64
		wantClosedComm float64 // entry_comm(open) + exit_comm(reversal)
		wantOpenComm   float64 // entry_comm(reversal)
	}{
		{
			name:           "percent_long_to_short",
			commType:       CommissionPercent,
			commRate:       1.0,
			initDir:        Long,
			reversalDir:    Short,
			openQty:        10,
			entryPrice:     100,
			reversalQty:    10,
			reversalPrice:  110,
			wantClosedComm: 10*100*0.01 + 10*110*0.01,
			wantOpenComm:   10 * 110 * 0.01,
		},
		{
			name:           "cash_per_order_long_to_short",
			commType:       CommissionCashPerOrder,
			commRate:       5.0,
			initDir:        Long,
			reversalDir:    Short,
			openQty:        10,
			entryPrice:     100,
			reversalQty:    10,
			reversalPrice:  110,
			wantClosedComm: 5.0 + 5.0,
			wantOpenComm:   5.0,
		},
		{
			name:           "cash_per_contract_long_to_short",
			commType:       CommissionCashPerContract,
			commRate:       2.0,
			initDir:        Long,
			reversalDir:    Short,
			openQty:        5,
			entryPrice:     100,
			reversalQty:    5,
			reversalPrice:  110,
			wantClosedComm: 5*2.0 + 5*2.0,
			wantOpenComm:   5 * 2.0,
		},
		{
			name:           "percent_short_to_long",
			commType:       CommissionPercent,
			commRate:       0.5,
			initDir:        Short,
			reversalDir:    Long,
			openQty:        20,
			entryPrice:     200,
			reversalQty:    20,
			reversalPrice:  190,
			wantClosedComm: 20*200*0.005 + 20*190*0.005,
			wantOpenComm:   20 * 190 * 0.005,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStrategy()
			s.CallWithPyramiding("Test", 1000000, 10)
			s.SetCommission(tt.commRate, tt.commType)

			s.Entry("Init", tt.initDir, tt.openQty, "")
			s.OnBarUpdate(1, tt.entryPrice, 1000)

			s.Entry("Rev", tt.reversalDir, tt.reversalQty, "")
			s.OnBarUpdate(2, tt.reversalPrice, 2000)

			closed := s.GetTradeHistory().GetClosedTrades()
			if len(closed) != 1 {
				t.Fatalf("closed trades: want 1, got %d", len(closed))
			}
			if closed[0].Commission != tt.wantClosedComm {
				t.Errorf("closed commission: got %.4f, want %.4f", closed[0].Commission, tt.wantClosedComm)
			}

			open := s.GetTradeHistory().GetOpenTrades()
			if len(open) != 1 {
				t.Fatalf("open trades: want 1, got %d", len(open))
			}
			if open[0].Commission != tt.wantOpenComm {
				t.Errorf("new entry commission: got %.4f, want %.4f", open[0].Commission, tt.wantOpenComm)
			}
		})
	}
}
