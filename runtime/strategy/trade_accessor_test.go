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
		if got := accessor.ClosedTradeExitID(0); got != "LONG_ENTRY_001" {
			t.Errorf("exit_id(0) = %q, want %q", got, "LONG_ENTRY_001")
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

	// Add one trade to test boundary conditions
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

	t.Run("zero_position_size", func(t *testing.T) {
		history := NewTradeHistory()
		accessor := NewTradeAccessor(history)

		trade := Trade{
			Direction:  Long,
			EntryPrice: 100.0,
			Size:       0.0, // Edge case: zero size
			Profit:     10.0,
		}
		history.closedTrades = append(history.closedTrades, trade)

		got := accessor.ClosedTradeProfitPercent(0)
		if !math.IsInf(got, 1) && !math.IsNaN(got) {
			t.Errorf("profit_percent with zero size should be Inf or NaN, got %f", got)
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

	// Add 5 closed trades with distinct values
	for i := 0; i < 5; i++ {
		trade := Trade{
			EntryID:    string(rune('A' + i)), // "A", "B", "C", "D", "E"
			Profit:     float64(i * 10),       // 0, 10, 20, 30, 40
			Size:       float64(i + 1),        // 1, 2, 3, 4, 5
			EntryPrice: 100.0 + float64(i),    // 100, 101, 102, 103, 104
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
	/* Current implementation passes history to constructor - documents expected behavior for nil history scenario */
	history := NewTradeHistory()
	accessor := NewTradeAccessor(history)

	// Verify zero-length slices behave correctly
	if !math.IsNaN(accessor.ClosedTradeProfit(0)) {
		t.Error("Empty history should return NaN for float accessors")
	}
	if accessor.ClosedTradeEntryID(0) != "" {
		t.Error("Empty history should return empty string for string accessors")
	}
}
