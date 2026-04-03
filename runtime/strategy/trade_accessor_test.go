package strategy

import (
	"math"
	"testing"
)

func TestTradeAccessor_ClosedTradePropertyAccess(t *testing.T) {
	history := NewTradeHistory()
	accessor := NewTradeAccessor(history)

	trade := Trade{
		EntryID:      "LONG_ENTRY_001",
		ExitID:       "EXIT_STOP_001",
		Direction:    Long,
		Size:         10.0,
		EntryPrice:   100.0,
		ExitPrice:    105.0,
		EntryBar:     5,
		ExitBar:      10,
		EntryTime:    1000000,
		ExitTime:     2000000,
		EntryComment: "Entry Comment",
		ExitComment:  "Exit Comment",
		Profit:       50.0,
	}
	history.closedTrades = append(history.closedTrades, trade)

	tests := []struct {
		name     string
		accessor func(int) float64
		want     float64
	}{
		{"entry_price", accessor.ClosedTradeEntryPrice, 100.0},
		{"exit_price", accessor.ClosedTradeExitPrice, 105.0},
		{"entry_bar_index", accessor.ClosedTradeEntryBarIndex, 5.0},
		{"exit_bar_index", accessor.ClosedTradeExitBarIndex, 10.0},
		{"entry_time", accessor.ClosedTradeEntryTime, 1000000.0},
		{"exit_time", accessor.ClosedTradeExitTime, 2000000.0},
		{"profit", accessor.ClosedTradeProfit, 50.0},
		{"size", accessor.ClosedTradeSize, 10.0},
		{"commission", accessor.ClosedTradeCommission, 0.0},
		{"max_runup", accessor.ClosedTradeMaxRunup, 0.0},
		{"max_drawdown", accessor.ClosedTradeMaxDrawdown, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.accessor(0)
			if got != tt.want {
				t.Errorf("%s(0) = %f, want %f", tt.name, got, tt.want)
			}
		})
	}

	t.Run("entry_id", func(t *testing.T) {
		if got := accessor.ClosedTradeEntryID(0); got != "LONG_ENTRY_001" {
			t.Errorf("entry_id(0) = %q, want %q", got, "LONG_ENTRY_001")
		}
	})

	t.Run("exit_id", func(t *testing.T) {
		if got := accessor.ClosedTradeExitID(0); got != "EXIT_STOP_001" {
			t.Errorf("exit_id(0) = %q, want %q", got, "EXIT_STOP_001")
		}
	})

	t.Run("entry_comment", func(t *testing.T) {
		if got := accessor.ClosedTradeEntryComment(0); got != "Entry Comment" {
			t.Errorf("entry_comment(0) = %q, want %q", got, "Entry Comment")
		}
	})

	t.Run("exit_comment", func(t *testing.T) {
		if got := accessor.ClosedTradeExitComment(0); got != "Exit Comment" {
			t.Errorf("exit_comment(0) = %q, want %q", got, "Exit Comment")
		}
	})
}

func TestTradeAccessor_OpenTradePropertyAccess(t *testing.T) {
	history := NewTradeHistory()
	accessor := NewTradeAccessor(history)

	trade := Trade{
		EntryID:      "SHORT_ENTRY_001",
		Direction:    Short,
		Size:         5.0,
		EntryPrice:   200.0,
		EntryBar:     15,
		EntryTime:    3000000,
		EntryComment: "Short Entry Comment",
		Profit:       -25.0,
	}
	history.openTrades = append(history.openTrades, trade)

	tests := []struct {
		name     string
		accessor func(int) float64
		want     float64
	}{
		{"entry_price", accessor.OpenTradeEntryPrice, 200.0},
		{"entry_bar_index", accessor.OpenTradeEntryBarIndex, 15.0},
		{"entry_time", accessor.OpenTradeEntryTime, 3000000.0},
		{"profit", accessor.OpenTradeProfit, -25.0},
		{"size", accessor.OpenTradeSize, -5.0},
		{"commission", accessor.OpenTradeCommission, 0.0},
		{"max_runup", accessor.OpenTradeMaxRunup, 0.0},
		{"max_drawdown", accessor.OpenTradeMaxDrawdown, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.accessor(0)
			if got != tt.want {
				t.Errorf("%s(0) = %f, want %f", tt.name, got, tt.want)
			}
		})
	}

	t.Run("entry_id", func(t *testing.T) {
		if got := accessor.OpenTradeEntryID(0); got != "SHORT_ENTRY_001" {
			t.Errorf("entry_id(0) = %q, want %q", got, "SHORT_ENTRY_001")
		}
	})

	t.Run("entry_comment", func(t *testing.T) {
		if got := accessor.OpenTradeEntryComment(0); got != "Short Entry Comment" {
			t.Errorf("entry_comment(0) = %q, want %q", got, "Short Entry Comment")
		}
	})
}

func TestTradeAccessor_BoundsValidation(t *testing.T) {
	history := NewTradeHistory()
	accessor := NewTradeAccessor(history)

	trade := Trade{
		EntryID:    "test",
		Direction:  Long,
		Size:       10.0,
		EntryPrice: 100.0,
		Profit:     50.0,
	}
	history.closedTrades = append(history.closedTrades, trade)
	history.openTrades = append(history.openTrades, trade)

	tests := []struct {
		name           string
		idx            int
		accessorFloat  func(int) float64
		accessorString func(int) string
		description    string
	}{
		{"negative_index", -1, accessor.ClosedTradeProfit, accessor.ClosedTradeEntryID, "negative indices"},
		{"large_positive", 999, accessor.ClosedTradeProfit, accessor.ClosedTradeEntryID, "large out-of-bounds index"},
		{"exact_length", 1, accessor.ClosedTradeProfit, accessor.ClosedTradeEntryID, "index equal to length"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.accessorFloat(tt.idx)
			if !math.IsNaN(got) {
				t.Errorf("Float accessor with %s should return NaN, got %f", tt.description, got)
			}

			gotStr := tt.accessorString(tt.idx)
			if gotStr != "" {
				t.Errorf("String accessor with %s should return empty string, got %q", tt.description, gotStr)
			}
		})
	}

	t.Run("empty_collections", func(t *testing.T) {
		emptyHistory := NewTradeHistory()
		emptyAccessor := NewTradeAccessor(emptyHistory)

		if !math.IsNaN(emptyAccessor.ClosedTradeProfit(0)) {
			t.Error("Accessing empty closed trades should return NaN")
		}
		if !math.IsNaN(emptyAccessor.OpenTradeProfit(0)) {
			t.Error("Accessing empty open trades should return NaN")
		}
		if emptyAccessor.ClosedTradeEntryID(0) != "" {
			t.Error("Accessing empty closed trades should return empty string")
		}
		if emptyAccessor.OpenTradeEntryID(0) != "" {
			t.Error("Accessing empty open trades should return empty string")
		}
	})
}

func TestTradeAccessor_ProfitPercentCalculation(t *testing.T) {
	tests := []struct {
		name       string
		entryPrice float64
		size       float64
		profit     float64
		want       float64
	}{
		{
			name:       "standard_long_profit",
			entryPrice: 100.0,
			size:       10.0,
			profit:     50.0,
			want:       5.0, // (50 / (100 * 10)) * 100 = 5%
		},
		{
			name:       "standard_short_loss",
			entryPrice: 200.0,
			size:       5.0,
			profit:     -25.0,
			want:       -2.5, // (-25 / (200 * 5)) * 100 = -2.5%
		},
		{
			name:       "small_profit",
			entryPrice: 1000.0,
			size:       1.0,
			profit:     10.0,
			want:       1.0, // (10 / 1000) * 100 = 1%
		},
		{
			name:       "large_position",
			entryPrice: 50.0,
			size:       100.0,
			profit:     500.0,
			want:       10.0, // (500 / 5000) * 100 = 10%
		},
		{
			name:       "zero_profit",
			entryPrice: 100.0,
			size:       10.0,
			profit:     0.0,
			want:       0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := NewTradeHistory()
			accessor := NewTradeAccessor(history)

			trade := Trade{
				Direction:  Long,
				EntryPrice: tt.entryPrice,
				Size:       tt.size,
				Profit:     tt.profit,
			}
			history.closedTrades = append(history.closedTrades, trade)

			got := accessor.ClosedTradeProfitPercent(0)
			if math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("profit_percent = %f, want %f (formula: (%.0f / (%.0f * %.0f)) * 100)",
					got, tt.want, tt.profit, tt.entryPrice, tt.size)
			}
		})
	}

	t.Run("zero_entry_price", func(t *testing.T) {
		history := NewTradeHistory()
		accessor := NewTradeAccessor(history)

		trade := Trade{
			Direction:  Long,
			EntryPrice: 0.0,
			Size:       10.0,
			Profit:     10.0,
		}
		history.closedTrades = append(history.closedTrades, trade)

		got := accessor.ClosedTradeProfitPercent(0)
		if !math.IsNaN(got) {
			t.Errorf("profit_percent with zero entry price = %f, want NaN", got)
		}
	})

	t.Run("zero_position_size", func(t *testing.T) {
		history := NewTradeHistory()
		accessor := NewTradeAccessor(history)

		trade := Trade{
			Direction:  Long,
			EntryPrice: 100.0,
			Size:       0.0,
			Profit:     10.0,
		}
		history.closedTrades = append(history.closedTrades, trade)

		got := accessor.ClosedTradeProfitPercent(0)
		if !math.IsNaN(got) {
			t.Errorf("profit_percent with zero size = %f, want NaN", got)
		}
	})
}

func TestTradeAccessor_DirectionSignConvention(t *testing.T) {
	tests := []struct {
		name      string
		direction string
		rawSize   float64
		wantSize  float64
	}{
		{"long_positive_size", Long, 10.0, 10.0},
		{"long_fractional_size", Long, 2.5, 2.5},
		{"short_negative_size", Short, 5.0, -5.0},
		{"short_fractional_size", Short, 1.25, -1.25},
		{"long_zero_size", Long, 0.0, 0.0},
		{"short_zero_size", Short, 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := NewTradeHistory()
			accessor := NewTradeAccessor(history)

			trade := Trade{
				Direction: tt.direction,
				Size:      tt.rawSize,
			}
			history.closedTrades = append(history.closedTrades, trade)

			got := accessor.ClosedTradeSize(0)
			if got != tt.wantSize {
				t.Errorf("Size() for %s with raw size %f = %f, want %f",
					tt.direction, tt.rawSize, got, tt.wantSize)
			}
		})
	}
}

func TestTradeAccessor_MultipleTradesIndexing(t *testing.T) {
	history := NewTradeHistory()
	accessor := NewTradeAccessor(history)

	for i := 0; i < 5; i++ {
		trade := Trade{
			EntryID:    string(rune('A' + i)),
			Profit:     float64(i * 10),
			Size:       float64(i + 1),
			EntryPrice: 100.0 + float64(i),
			Direction:  Long,
		}
		history.closedTrades = append(history.closedTrades, trade)
	}

	tests := []struct {
		idx            int
		wantID         string
		wantProfit     float64
		wantSize       float64
		wantEntryPrice float64
	}{
		{0, "A", 0.0, 1.0, 100.0},
		{1, "B", 10.0, 2.0, 101.0},
		{2, "C", 20.0, 3.0, 102.0},
		{3, "D", 30.0, 4.0, 103.0},
		{4, "E", 40.0, 5.0, 104.0},
	}

	for _, tt := range tests {
		t.Run(tt.wantID, func(t *testing.T) {
			if got := accessor.ClosedTradeEntryID(tt.idx); got != tt.wantID {
				t.Errorf("entry_id(%d) = %q, want %q", tt.idx, got, tt.wantID)
			}
			if got := accessor.ClosedTradeProfit(tt.idx); got != tt.wantProfit {
				t.Errorf("profit(%d) = %f, want %f", tt.idx, got, tt.wantProfit)
			}
			if got := accessor.ClosedTradeSize(tt.idx); got != tt.wantSize {
				t.Errorf("size(%d) = %f, want %f", tt.idx, got, tt.wantSize)
			}
			if got := accessor.ClosedTradeEntryPrice(tt.idx); got != tt.wantEntryPrice {
				t.Errorf("entry_price(%d) = %f, want %f", tt.idx, got, tt.wantEntryPrice)
			}
		})
	}
}

func TestTradeAccessor_NilSafety(t *testing.T) {
	history := NewTradeHistory()
	accessor := NewTradeAccessor(history)

	if !math.IsNaN(accessor.ClosedTradeProfit(0)) {
		t.Error("Empty history should return NaN for float accessors")
	}
	if accessor.ClosedTradeEntryID(0) != "" {
		t.Error("Empty history should return empty string for string accessors")
	}
}

func TestTradeHistory_UpdateOpenTradeMetrics(t *testing.T) {
	tests := []struct {
		name         string
		direction    string
		entryPrice   float64
		size         float64
		barHigh      float64
		barLow       float64
		wantRunup    float64
		wantDrawdown float64
	}{
		{
			name:         "long trade favorable bar",
			direction:    Long,
			entryPrice:   100.0,
			size:         10.0,
			barHigh:      108.0,
			barLow:       97.0,
			wantRunup:    80.0, // (108-100)*10
			wantDrawdown: 30.0, // (100-97)*10
		},
		{
			name:         "short trade favorable bar",
			direction:    Short,
			entryPrice:   100.0,
			size:         10.0,
			barHigh:      103.0,
			barLow:       92.0,
			wantRunup:    80.0, // (100-92)*10
			wantDrawdown: 30.0, // (103-100)*10
		},
		{
			name:         "long trade no adverse excursion",
			direction:    Long,
			entryPrice:   100.0,
			size:         5.0,
			barHigh:      110.0,
			barLow:       100.0, // Low == entry - no adverse excursion
			wantRunup:    50.0,  // (110-100)*5
			wantDrawdown: 0.0,   // (100-100)*5 = 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTradeHistory()
			th.AddOpenTrade(Trade{
				EntryID:    "test",
				Direction:  tt.direction,
				Size:       tt.size,
				EntryPrice: tt.entryPrice,
			})

			th.UpdateOpenTradeMetrics(tt.barHigh, tt.barLow)

			trades := th.GetOpenTrades()
			if len(trades) != 1 {
				t.Fatalf("Expected 1 open trade, got %d", len(trades))
			}
			if trades[0].MaxRunup != tt.wantRunup {
				t.Errorf("MaxRunup = %f, want %f", trades[0].MaxRunup, tt.wantRunup)
			}
			if trades[0].MaxDrawdown != tt.wantDrawdown {
				t.Errorf("MaxDrawdown = %f, want %f", trades[0].MaxDrawdown, tt.wantDrawdown)
			}
		})
	}

	t.Run("running max across multiple bars", func(t *testing.T) {
		th := NewTradeHistory()
		th.AddOpenTrade(Trade{
			EntryID:    "long1",
			Direction:  Long,
			Size:       10.0,
			EntryPrice: 100.0,
		})

		th.UpdateOpenTradeMetrics(105.0, 98.0)  // runup=50, drawdown=20
		th.UpdateOpenTradeMetrics(112.0, 101.0) // runup=120 (new high), drawdown still 20
		th.UpdateOpenTradeMetrics(103.0, 95.0)  // runup still 120, drawdown=50 (new low)

		trades := th.GetOpenTrades()
		if trades[0].MaxRunup != 120.0 {
			t.Errorf("MaxRunup = %f, want 120.0 (running max)", trades[0].MaxRunup)
		}
		if trades[0].MaxDrawdown != 50.0 {
			t.Errorf("MaxDrawdown = %f, want 50.0 (running max)", trades[0].MaxDrawdown)
		}
	})
}

func TestTradeAccessor_MaxExcursionPercent(t *testing.T) {
	tests := []struct {
		name            string
		entryPrice      float64
		size            float64
		maxDrawdown     float64
		maxRunup        float64
		wantDrawdownPct float64
		wantRunupPct    float64
	}{
		{
			name:            "standard values",
			entryPrice:      100.0,
			size:            10.0,
			maxDrawdown:     50.0,
			maxRunup:        80.0,
			wantDrawdownPct: 5.0, // 50/(100*10)*100
			wantRunupPct:    8.0, // 80/(100*10)*100
		},
		{
			name:            "small position",
			entryPrice:      1000.0,
			size:            1.0,
			maxDrawdown:     10.0,
			maxRunup:        25.0,
			wantDrawdownPct: 1.0, // 10/1000*100
			wantRunupPct:    2.5, // 25/1000*100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := NewTradeHistory()
			accessor := NewTradeAccessor(history)

			trade := Trade{
				Direction:   Long,
				EntryPrice:  tt.entryPrice,
				Size:        tt.size,
				MaxDrawdown: tt.maxDrawdown,
				MaxRunup:    tt.maxRunup,
			}
			history.closedTrades = append(history.closedTrades, trade)

			gotDrawdown := accessor.ClosedTradeMaxDrawdownPercent(0)
			if math.Abs(gotDrawdown-tt.wantDrawdownPct) > 0.0001 {
				t.Errorf("MaxDrawdownPercent = %f, want %f", gotDrawdown, tt.wantDrawdownPct)
			}

			gotRunup := accessor.ClosedTradeMaxRunupPercent(0)
			if math.Abs(gotRunup-tt.wantRunupPct) > 0.0001 {
				t.Errorf("MaxRunupPercent = %f, want %f", gotRunup, tt.wantRunupPct)
			}
		})
	}

	t.Run("zero_basis_returns_nan", func(t *testing.T) {
		history := NewTradeHistory()
		accessor := NewTradeAccessor(history)

		trade := Trade{Direction: Long, EntryPrice: 0.0, Size: 10.0, MaxDrawdown: 50.0}
		history.closedTrades = append(history.closedTrades, trade)

		if !math.IsNaN(accessor.ClosedTradeMaxDrawdownPercent(0)) {
			t.Error("MaxDrawdownPercent with zero entryPrice should return NaN")
		}
	})

	t.Run("zero_size_returns_nan", func(t *testing.T) {
		history := NewTradeHistory()
		accessor := NewTradeAccessor(history)

		trade := Trade{Direction: Long, EntryPrice: 100.0, Size: 0.0, MaxDrawdown: 50.0, MaxRunup: 30.0}
		history.closedTrades = append(history.closedTrades, trade)

		if !math.IsNaN(accessor.ClosedTradeMaxDrawdownPercent(0)) {
			t.Error("MaxDrawdownPercent with zero size should return NaN")
		}
		if !math.IsNaN(accessor.ClosedTradeMaxRunupPercent(0)) {
			t.Error("MaxRunupPercent with zero size should return NaN")
		}
	})

	t.Run("open_trade_standard_values", func(t *testing.T) {
		history := NewTradeHistory()
		accessor := NewTradeAccessor(history)

		trade := Trade{
			Direction:   Long,
			EntryPrice:  100.0,
			Size:        10.0,
			MaxDrawdown: 50.0,
			MaxRunup:    80.0,
		}
		history.openTrades = append(history.openTrades, trade)

		gotDrawdown := accessor.OpenTradeMaxDrawdownPercent(0)
		if math.Abs(gotDrawdown-5.0) > 0.0001 {
			t.Errorf("OpenTradeMaxDrawdownPercent = %f, want 5.0", gotDrawdown)
		}

		gotRunup := accessor.OpenTradeMaxRunupPercent(0)
		if math.Abs(gotRunup-8.0) > 0.0001 {
			t.Errorf("OpenTradeMaxRunupPercent = %f, want 8.0", gotRunup)
		}
	})

	t.Run("open_trade_zero_basis_returns_nan", func(t *testing.T) {
		history := NewTradeHistory()
		accessor := NewTradeAccessor(history)

		trade := Trade{Direction: Long, EntryPrice: 0.0, Size: 10.0, MaxDrawdown: 50.0}
		history.openTrades = append(history.openTrades, trade)

		if !math.IsNaN(accessor.OpenTradeMaxDrawdownPercent(0)) {
			t.Error("OpenTradeMaxDrawdownPercent with zero basis should return NaN")
		}
		if !math.IsNaN(accessor.OpenTradeMaxRunupPercent(0)) {
			t.Error("OpenTradeMaxRunupPercent with zero basis should return NaN")
		}
	})
}

func TestTradeAccessor_OpenTradeProfitPercent(t *testing.T) {
	tests := []struct {
		name       string
		entryPrice float64
		size       float64
		profit     float64
		want       float64
	}{
		{"standard_profit", 100.0, 10.0, 50.0, 5.0},
		{"standard_loss", 200.0, 5.0, -25.0, -2.5},
		{"zero_profit", 100.0, 10.0, 0.0, 0.0},
		{"fractional_size", 100.0, 0.5, 10.0, 20.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := NewTradeHistory()
			accessor := NewTradeAccessor(history)

			history.openTrades = append(history.openTrades, Trade{
				Direction:  Long,
				EntryPrice: tt.entryPrice,
				Size:       tt.size,
				Profit:     tt.profit,
			})

			got := accessor.OpenTradeProfitPercent(0)
			if math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("OpenTradeProfitPercent = %f, want %f", got, tt.want)
			}
		})
	}

	t.Run("zero_entry_price", func(t *testing.T) {
		history := NewTradeHistory()
		accessor := NewTradeAccessor(history)
		history.openTrades = append(history.openTrades, Trade{Direction: Long, EntryPrice: 0.0, Size: 10.0, Profit: 10.0})
		if !math.IsNaN(accessor.OpenTradeProfitPercent(0)) {
			t.Error("OpenTradeProfitPercent with zero entry price should return NaN")
		}
	})

	t.Run("zero_size", func(t *testing.T) {
		history := NewTradeHistory()
		accessor := NewTradeAccessor(history)
		history.openTrades = append(history.openTrades, Trade{Direction: Long, EntryPrice: 100.0, Size: 0.0, Profit: 10.0})
		if !math.IsNaN(accessor.OpenTradeProfitPercent(0)) {
			t.Error("OpenTradeProfitPercent with zero size should return NaN")
		}
	})
}

func TestTradeHistory_MetricsFrozenAfterClose(t *testing.T) {
	th := NewTradeHistory()
	th.AddOpenTrade(Trade{
		EntryID:    "long1",
		Direction:  Long,
		Size:       10.0,
		EntryPrice: 100.0,
	})

	th.UpdateOpenTradeMetrics(110.0, 95.0)

	th.CloseTrade("long1", "exit1", 108.0, 5, 1000, "", 0)

	closed := th.GetClosedTrades()
	if len(closed) != 1 {
		t.Fatalf("Expected 1 closed trade, got %d", len(closed))
	}
	if closed[0].MaxRunup != 100.0 {
		t.Errorf("Closed trade MaxRunup = %f, want 100.0 (frozen at close)", closed[0].MaxRunup)
	}
	if closed[0].MaxDrawdown != 50.0 {
		t.Errorf("Closed trade MaxDrawdown = %f, want 50.0 (frozen at close)", closed[0].MaxDrawdown)
	}

	th.UpdateOpenTradeMetrics(120.0, 80.0)

	closed = th.GetClosedTrades()
	if closed[0].MaxRunup != 100.0 {
		t.Errorf("Closed trade MaxRunup changed after close: %f, want 100.0", closed[0].MaxRunup)
	}
	if closed[0].MaxDrawdown != 50.0 {
		t.Errorf("Closed trade MaxDrawdown changed after close: %f, want 50.0", closed[0].MaxDrawdown)
	}
}

func TestTradeHistory_MultipleOpenTradesMetrics(t *testing.T) {
	th := NewTradeHistory()
	th.AddOpenTrade(Trade{EntryID: "long1", Direction: Long, Size: 10.0, EntryPrice: 100.0})
	th.AddOpenTrade(Trade{EntryID: "short1", Direction: Short, Size: 5.0, EntryPrice: 200.0})

	// barHigh=105, barLow=95
	// Long:  runup=(105-100)*10=50, drawdown=(100-95)*10=50
	// Short: runup=(200-95)*5=525,  drawdown=(105-200)*5=-475 → clamped to 0
	th.UpdateOpenTradeMetrics(105.0, 95.0)

	open := th.GetOpenTrades()
	if len(open) != 2 {
		t.Fatalf("Expected 2 open trades, got %d", len(open))
	}

	longTrade := open[0]
	if longTrade.MaxRunup != 50.0 {
		t.Errorf("Long MaxRunup = %f, want 50.0", longTrade.MaxRunup)
	}
	if longTrade.MaxDrawdown != 50.0 {
		t.Errorf("Long MaxDrawdown = %f, want 50.0", longTrade.MaxDrawdown)
	}

	shortTrade := open[1]
	if shortTrade.MaxRunup != 525.0 {
		t.Errorf("Short MaxRunup = %f, want 525.0 ((200-95)*5)", shortTrade.MaxRunup)
	}
	if shortTrade.MaxDrawdown != 0.0 {
		t.Errorf("Short MaxDrawdown = %f, want 0.0 (barHigh=105 below entry=200)", shortTrade.MaxDrawdown)
	}
}
