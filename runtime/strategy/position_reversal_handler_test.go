package strategy

import "testing"

func TestPositionReversalHandler_NoOpenTrades(t *testing.T) {
	th := NewTradeHistory()
	pt := NewPositionTracker()
	ec := NewEquityCalculator(10000)
	handler := NewPositionReversalHandler(th, pt, ec)

	handler.HandleReversal(Long, 100.0, 1, 1001)

	if len(th.GetClosedTrades()) != 0 {
		t.Errorf("Closed trades: expected 0, got %d", len(th.GetClosedTrades()))
	}
	if pt.GetPositionSize() != 0 {
		t.Errorf("Position size: expected 0, got %.1f", pt.GetPositionSize())
	}
}

func TestPositionReversalHandler_NoOppositeDirection(t *testing.T) {
	tests := []struct {
		name             string
		existingDir      string
		entryDir         string
		wantClosed       int
		wantPositionSize float64
	}{
		{
			name:             "long_entry_with_long_position",
			existingDir:      Long,
			entryDir:         Long,
			wantClosed:       0,
			wantPositionSize: 1.0,
		},
		{
			name:             "short_entry_with_short_position",
			existingDir:      Short,
			entryDir:         Short,
			wantClosed:       0,
			wantPositionSize: -1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTradeHistory()
			pt := NewPositionTracker()
			ec := NewEquityCalculator(10000)
			handler := NewPositionReversalHandler(th, pt, ec)

			th.AddOpenTrade(Trade{
				EntryID:    "existing",
				Direction:  tt.existingDir,
				Size:       1.0,
				EntryPrice: 100.0,
				EntryBar:   1,
				EntryTime:  1001,
			})
			pt.UpdatePosition(1.0, 100.0, tt.existingDir)

			handler.HandleReversal(tt.entryDir, 105.0, 2, 1002)

			if len(th.GetClosedTrades()) != tt.wantClosed {
				t.Errorf("Closed trades: expected %d, got %d", tt.wantClosed, len(th.GetClosedTrades()))
			}
			if pt.GetPositionSize() != tt.wantPositionSize {
				t.Errorf("Position size: expected %.1f, got %.1f", tt.wantPositionSize, pt.GetPositionSize())
			}
		})
	}
}

func TestPositionReversalHandler_SingleTrade(t *testing.T) {
	tests := []struct {
		name         string
		existingDir  string
		existingSize float64
		entryDir     string
		exitPrice    float64
		wantProfit   float64
	}{
		{
			name:         "long_reversed_by_short",
			existingDir:  Long,
			existingSize: 1.0,
			entryDir:     Short,
			exitPrice:    110.0,
			wantProfit:   10.0, // (110-100)*1
		},
		{
			name:         "short_reversed_by_long",
			existingDir:  Short,
			existingSize: 1.0,
			entryDir:     Long,
			exitPrice:    95.0,
			wantProfit:   5.0, // (100-95)*1
		},
		{
			name:         "large_position_reversed",
			existingDir:  Long,
			existingSize: 10.0,
			entryDir:     Short,
			exitPrice:    105.0,
			wantProfit:   50.0, // (105-100)*10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTradeHistory()
			pt := NewPositionTracker()
			ec := NewEquityCalculator(10000)
			handler := NewPositionReversalHandler(th, pt, ec)

			th.AddOpenTrade(Trade{
				EntryID:    "existing",
				Direction:  tt.existingDir,
				Size:       tt.existingSize,
				EntryPrice: 100.0,
				EntryBar:   1,
				EntryTime:  1001,
			})
			pt.UpdatePosition(tt.existingSize, 100.0, tt.existingDir)

			handler.HandleReversal(tt.entryDir, tt.exitPrice, 2, 1002)

			closedTrades := th.GetClosedTrades()
			if len(closedTrades) != 1 {
				t.Errorf("Closed trades: expected 1, got %d", len(closedTrades))
			}

			if len(closedTrades) > 0 {
				if closedTrades[0].Profit != tt.wantProfit {
					t.Errorf("Profit: expected %.1f, got %.1f", tt.wantProfit, closedTrades[0].Profit)
				}
				if closedTrades[0].ExitComment != "Position reversal" {
					t.Errorf("Exit comment: expected 'Position reversal', got '%s'", closedTrades[0].ExitComment)
				}
			}

			if pt.GetPositionSize() != 0 {
				t.Errorf("Position size: expected 0, got %.1f", pt.GetPositionSize())
			}

			netProfit := ec.GetNetProfit()
			if netProfit != tt.wantProfit {
				t.Errorf("Net profit: expected %.1f, got %.1f", tt.wantProfit, netProfit)
			}
		})
	}
}

func TestPositionReversalHandler_MultipleTrades(t *testing.T) {
	tests := []struct {
		name       string
		trades     []Trade
		entryDir   string
		exitPrice  float64
		wantClosed int
		wantProfit float64
	}{
		{
			name: "two_long_positions_reversed",
			trades: []Trade{
				{EntryID: "long1", Direction: Long, Size: 1.0, EntryPrice: 100.0, EntryBar: 1, EntryTime: 1001},
				{EntryID: "long2", Direction: Long, Size: 1.0, EntryPrice: 105.0, EntryBar: 2, EntryTime: 1002},
			},
			entryDir:   Short,
			exitPrice:  110.0,
			wantClosed: 2,
			wantProfit: 15.0, // (110-100) + (110-105)
		},
		{
			name: "three_short_positions_reversed",
			trades: []Trade{
				{EntryID: "short1", Direction: Short, Size: 1.0, EntryPrice: 100.0, EntryBar: 1, EntryTime: 1001},
				{EntryID: "short2", Direction: Short, Size: 1.0, EntryPrice: 95.0, EntryBar: 2, EntryTime: 1002},
				{EntryID: "short3", Direction: Short, Size: 1.0, EntryPrice: 90.0, EntryBar: 3, EntryTime: 1003},
			},
			entryDir:   Long,
			exitPrice:  85.0,
			wantClosed: 3,
			wantProfit: 30.0, // (100-85) + (95-85) + (90-85)
		},
		{
			name: "varying_sizes_reversed",
			trades: []Trade{
				{EntryID: "pos1", Direction: Long, Size: 2.0, EntryPrice: 100.0, EntryBar: 1, EntryTime: 1001},
				{EntryID: "pos2", Direction: Long, Size: 3.0, EntryPrice: 105.0, EntryBar: 2, EntryTime: 1002},
			},
			entryDir:   Short,
			exitPrice:  110.0,
			wantClosed: 2,
			wantProfit: 35.0, // (110-100)*2 + (110-105)*3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTradeHistory()
			pt := NewPositionTracker()
			ec := NewEquityCalculator(10000)
			handler := NewPositionReversalHandler(th, pt, ec)

			for _, trade := range tt.trades {
				th.AddOpenTrade(trade)
				pt.UpdatePosition(trade.Size, trade.EntryPrice, trade.Direction)
			}

			handler.HandleReversal(tt.entryDir, tt.exitPrice, 10, 1010)

			closedTrades := th.GetClosedTrades()
			if len(closedTrades) != tt.wantClosed {
				t.Errorf("Closed trades: expected %d, got %d", tt.wantClosed, len(closedTrades))
			}

			netProfit := ec.GetNetProfit()
			if netProfit != tt.wantProfit {
				t.Errorf("Net profit: expected %.1f, got %.1f", tt.wantProfit, netProfit)
			}

			if pt.GetPositionSize() != 0 {
				t.Errorf("Position size: expected 0, got %.1f", pt.GetPositionSize())
			}
		})
	}
}

func TestPositionReversalHandler_MixedDirections(t *testing.T) {
	th := NewTradeHistory()
	pt := NewPositionTracker()
	ec := NewEquityCalculator(10000)
	handler := NewPositionReversalHandler(th, pt, ec)

	th.AddOpenTrade(Trade{
		EntryID:    "long1",
		Direction:  Long,
		Size:       2.0,
		EntryPrice: 100.0,
		EntryBar:   1,
		EntryTime:  1001,
	})
	th.AddOpenTrade(Trade{
		EntryID:    "short1",
		Direction:  Short,
		Size:       1.0,
		EntryPrice: 105.0,
		EntryBar:   2,
		EntryTime:  1002,
	})
	th.AddOpenTrade(Trade{
		EntryID:    "long2",
		Direction:  Long,
		Size:       1.0,
		EntryPrice: 103.0,
		EntryBar:   3,
		EntryTime:  1003,
	})

	pt.UpdatePosition(2.0, 100.0, Long)
	pt.UpdatePosition(1.0, 105.0, Short)
	pt.UpdatePosition(1.0, 103.0, Long)

	handler.HandleReversal(Short, 110.0, 4, 1004)

	closedTrades := th.GetClosedTrades()
	if len(closedTrades) != 2 {
		t.Errorf("Closed trades: expected 2, got %d", len(closedTrades))
	}

	openTrades := th.GetOpenTrades()
	if len(openTrades) != 1 {
		t.Errorf("Open trades: expected 1, got %d", len(openTrades))
	}

	if len(openTrades) > 0 && openTrades[0].Direction != Short {
		t.Errorf("Remaining trade direction: expected Short, got %s", openTrades[0].Direction)
	}
}

func TestPositionReversalHandler_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*TradeHistory, *PositionTracker)
		entryDir   string
		exitPrice  float64
		wantClosed int
	}{
		{
			name: "zero_size_position",
			setup: func(th *TradeHistory, pt *PositionTracker) {
				th.AddOpenTrade(Trade{
					EntryID:    "zero",
					Direction:  Long,
					Size:       0.0,
					EntryPrice: 100.0,
					EntryBar:   1,
					EntryTime:  1001,
				})
			},
			entryDir:   Short,
			exitPrice:  110.0,
			wantClosed: 1,
		},
		{
			name: "fractional_size_position",
			setup: func(th *TradeHistory, pt *PositionTracker) {
				th.AddOpenTrade(Trade{
					EntryID:    "fractional",
					Direction:  Long,
					Size:       0.5,
					EntryPrice: 100.0,
					EntryBar:   1,
					EntryTime:  1001,
				})
				pt.UpdatePosition(0.5, 100.0, Long)
			},
			entryDir:   Short,
			exitPrice:  110.0,
			wantClosed: 1,
		},
		{
			name: "same_bar_reversal",
			setup: func(th *TradeHistory, pt *PositionTracker) {
				th.AddOpenTrade(Trade{
					EntryID:    "same_bar",
					Direction:  Long,
					Size:       1.0,
					EntryPrice: 100.0,
					EntryBar:   1,
					EntryTime:  1001,
				})
				pt.UpdatePosition(1.0, 100.0, Long)
			},
			entryDir:   Short,
			exitPrice:  100.0,
			wantClosed: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTradeHistory()
			pt := NewPositionTracker()
			ec := NewEquityCalculator(10000)
			handler := NewPositionReversalHandler(th, pt, ec)

			tt.setup(th, pt)

			handler.HandleReversal(tt.entryDir, tt.exitPrice, 1, 1001)

			closedTrades := th.GetClosedTrades()
			if len(closedTrades) != tt.wantClosed {
				t.Errorf("Closed trades: expected %d, got %d", tt.wantClosed, len(closedTrades))
			}
		})
	}
}
