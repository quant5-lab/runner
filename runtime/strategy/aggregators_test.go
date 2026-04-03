package strategy

import "testing"

func TestAggregateGrossProfit(t *testing.T) {
	tests := []struct {
		name     string
		trades   []Trade
		expected float64
	}{
		{
			name: "mixed_profits_and_losses",
			trades: []Trade{
				{Profit: 100},
				{Profit: -50},
				{Profit: 200},
				{Profit: -30},
				{Profit: 0},
			},
			expected: 300.0,
		},
		{
			name:     "empty_trades",
			trades:   []Trade{},
			expected: 0.0,
		},
		{
			name: "only_profits",
			trades: []Trade{
				{Profit: 50},
				{Profit: 100},
				{Profit: 150},
			},
			expected: 300.0,
		},
		{
			name: "only_losses",
			trades: []Trade{
				{Profit: -50},
				{Profit: -100},
			},
			expected: 0.0,
		},
		{
			name: "fractional_profits",
			trades: []Trade{
				{Profit: 0.5},
				{Profit: -0.3},
				{Profit: 1.25},
			},
			expected: 1.75,
		},
		{
			name: "large_values",
			trades: []Trade{
				{Profit: 1000000},
				{Profit: -500000},
				{Profit: 2000000},
			},
			expected: 3000000.0,
		},
		{
			name: "single_positive_trade",
			trades: []Trade{
				{Profit: 123.45},
			},
			expected: 123.45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AggregateGrossProfit(tt.trades)
			if result != tt.expected {
				t.Errorf("Expected %.2f, got %.2f", tt.expected, result)
			}
		})
	}
}

func TestAggregateGrossLoss(t *testing.T) {
	tests := []struct {
		name     string
		trades   []Trade
		expected float64
	}{
		{
			name: "mixed_profits_and_losses",
			trades: []Trade{
				{Profit: 100},
				{Profit: -50},
				{Profit: 200},
				{Profit: -30},
				{Profit: 0},
			},
			expected: -80.0,
		},
		{
			name:     "empty_trades",
			trades:   []Trade{},
			expected: 0.0,
		},
		{
			name: "only_profits",
			trades: []Trade{
				{Profit: 100},
				{Profit: 200},
			},
			expected: 0.0,
		},
		{
			name: "only_losses",
			trades: []Trade{
				{Profit: -50},
				{Profit: -100},
				{Profit: -25},
			},
			expected: -175.0,
		},
		{
			name: "fractional_losses",
			trades: []Trade{
				{Profit: -0.25},
				{Profit: 0.5},
				{Profit: -1.75},
			},
			expected: -2.0,
		},
		{
			name: "large_negative_values",
			trades: []Trade{
				{Profit: -1000000},
				{Profit: 500000},
				{Profit: -250000},
			},
			expected: -1250000.0,
		},
		{
			name: "single_negative_trade",
			trades: []Trade{
				{Profit: -456.78},
			},
			expected: -456.78,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AggregateGrossLoss(tt.trades)
			if result != tt.expected {
				t.Errorf("Expected %.2f, got %.2f", tt.expected, result)
			}
		})
	}
}

func TestCountWinningTrades(t *testing.T) {
	tests := []struct {
		name     string
		trades   []Trade
		expected int
	}{
		{
			name: "mixed_outcomes",
			trades: []Trade{
				{Profit: 100},
				{Profit: -50},
				{Profit: 200},
				{Profit: -30},
				{Profit: 0},
				{Profit: 10},
			},
			expected: 3,
		},
		{
			name:     "empty_trades",
			trades:   []Trade{},
			expected: 0,
		},
		{
			name: "all_winning",
			trades: []Trade{
				{Profit: 50},
				{Profit: 100},
				{Profit: 0.01},
			},
			expected: 3,
		},
		{
			name: "all_losing",
			trades: []Trade{
				{Profit: -50},
				{Profit: -100},
			},
			expected: 0,
		},
		{
			name: "all_breakeven",
			trades: []Trade{
				{Profit: 0},
				{Profit: 0},
				{Profit: 0},
			},
			expected: 0,
		},
		{
			name: "minimal_profit",
			trades: []Trade{
				{Profit: 0.001},
				{Profit: 0.0001},
			},
			expected: 2,
		},
		{
			name: "single_winning_trade",
			trades: []Trade{
				{Profit: 1},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountWinningTrades(tt.trades)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestCountLosingTrades(t *testing.T) {
	tests := []struct {
		name     string
		trades   []Trade
		expected int
	}{
		{
			name: "mixed_outcomes",
			trades: []Trade{
				{Profit: 100},
				{Profit: -50},
				{Profit: 200},
				{Profit: -30},
				{Profit: 0},
			},
			expected: 2,
		},
		{
			name:     "empty_trades",
			trades:   []Trade{},
			expected: 0,
		},
		{
			name: "all_winning",
			trades: []Trade{
				{Profit: 50},
				{Profit: 100},
			},
			expected: 0,
		},
		{
			name: "all_losing",
			trades: []Trade{
				{Profit: -50},
				{Profit: -100},
				{Profit: -0.01},
			},
			expected: 3,
		},
		{
			name: "minimal_loss",
			trades: []Trade{
				{Profit: -0.001},
				{Profit: -0.0001},
			},
			expected: 2,
		},
		{
			name: "single_losing_trade",
			trades: []Trade{
				{Profit: -1},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountLosingTrades(tt.trades)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestCountEvenTrades(t *testing.T) {
	tests := []struct {
		name     string
		trades   []Trade
		expected int
	}{
		{
			name: "mixed_with_breakevens",
			trades: []Trade{
				{Profit: 100},
				{Profit: 0},
				{Profit: 200},
				{Profit: 0},
				{Profit: -50},
			},
			expected: 2,
		},
		{
			name:     "empty_trades",
			trades:   []Trade{},
			expected: 0,
		},
		{
			name: "all_breakeven",
			trades: []Trade{
				{Profit: 0},
				{Profit: 0},
				{Profit: 0},
			},
			expected: 3,
		},
		{
			name: "no_breakeven",
			trades: []Trade{
				{Profit: 50},
				{Profit: -50},
				{Profit: 100},
			},
			expected: 0,
		},
		{
			name: "single_breakeven",
			trades: []Trade{
				{Profit: 0},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountEvenTrades(tt.trades)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestAggregatorsWithMixedScenarios(t *testing.T) {
	tests := []struct {
		name                string
		trades              []Trade
		expectedGrossProfit float64
		expectedGrossLoss   float64
		expectedWinCount    int
		expectedLossCount   int
		expectedEvenCount   int
	}{
		{
			name: "balanced_trading",
			trades: []Trade{
				{Profit: 100},
				{Profit: -100},
				{Profit: 50},
				{Profit: -50},
				{Profit: 0},
			},
			expectedGrossProfit: 150,
			expectedGrossLoss:   -150,
			expectedWinCount:    2,
			expectedLossCount:   2,
			expectedEvenCount:   1,
		},
		{
			name: "profitable_strategy",
			trades: []Trade{
				{Profit: 1000},
				{Profit: 500},
				{Profit: -200},
				{Profit: 300},
			},
			expectedGrossProfit: 1800,
			expectedGrossLoss:   -200,
			expectedWinCount:    3,
			expectedLossCount:   1,
			expectedEvenCount:   0,
		},
		{
			name: "losing_strategy",
			trades: []Trade{
				{Profit: 50},
				{Profit: -500},
				{Profit: -300},
				{Profit: -100},
			},
			expectedGrossProfit: 50,
			expectedGrossLoss:   -900,
			expectedWinCount:    1,
			expectedLossCount:   3,
			expectedEvenCount:   0,
		},
		{
			name:                "no_trades",
			trades:              []Trade{},
			expectedGrossProfit: 0,
			expectedGrossLoss:   0,
			expectedWinCount:    0,
			expectedLossCount:   0,
			expectedEvenCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grossProfit := AggregateGrossProfit(tt.trades)
			if grossProfit != tt.expectedGrossProfit {
				t.Errorf("GrossProfit: expected %.2f, got %.2f", tt.expectedGrossProfit, grossProfit)
			}

			grossLoss := AggregateGrossLoss(tt.trades)
			if grossLoss != tt.expectedGrossLoss {
				t.Errorf("GrossLoss: expected %.2f, got %.2f", tt.expectedGrossLoss, grossLoss)
			}

			winCount := CountWinningTrades(tt.trades)
			if winCount != tt.expectedWinCount {
				t.Errorf("WinCount: expected %d, got %d", tt.expectedWinCount, winCount)
			}

			lossCount := CountLosingTrades(tt.trades)
			if lossCount != tt.expectedLossCount {
				t.Errorf("LossCount: expected %d, got %d", tt.expectedLossCount, lossCount)
			}

			evenCount := CountEvenTrades(tt.trades)
			if evenCount != tt.expectedEvenCount {
				t.Errorf("EvenCount: expected %d, got %d", tt.expectedEvenCount, evenCount)
			}
		})
	}
}

func TestAvgTrade(t *testing.T) {
	tests := []struct {
		name     string
		trades   []Trade
		expected float64
	}{
		{"empty", []Trade{}, 0},
		{"all_winning", []Trade{{Profit: 100}, {Profit: 200}, {Profit: 300}}, 200},
		{"all_losing", []Trade{{Profit: -100}, {Profit: -200}}, -150},
		{"mixed", []Trade{{Profit: 300}, {Profit: -100}, {Profit: 200}, {Profit: -200}}, 50},
		{"single_trade", []Trade{{Profit: 75}}, 75},
		{"breakeven_only", []Trade{{Profit: 0}, {Profit: 0}}, 0},
		{"fractional", []Trade{{Profit: 0.5}, {Profit: 1.5}}, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AvgTrade(tt.trades)
			if result != tt.expected {
				t.Errorf("AvgTrade = %.6f, want %.6f", result, tt.expected)
			}
		})
	}
}

func TestAvgWinningTrade(t *testing.T) {
	tests := []struct {
		name     string
		trades   []Trade
		expected float64
	}{
		{"empty", []Trade{}, 0},
		{"no_wins", []Trade{{Profit: -100}, {Profit: 0}, {Profit: -50}}, 0},
		{"single_win", []Trade{{Profit: 150}}, 150},
		{"multiple_wins", []Trade{{Profit: 100}, {Profit: -50}, {Profit: 300}, {Profit: -25}}, 200},
		{"all_winning", []Trade{{Profit: 50}, {Profit: 100}, {Profit: 150}}, 100},
		{"wins_with_breakevens", []Trade{{Profit: 200}, {Profit: 0}, {Profit: 400}}, 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AvgWinningTrade(tt.trades)
			if result != tt.expected {
				t.Errorf("AvgWinningTrade = %.6f, want %.6f", result, tt.expected)
			}
		})
	}
}

func TestAvgLosingTrade(t *testing.T) {
	tests := []struct {
		name     string
		trades   []Trade
		expected float64
	}{
		{"empty", []Trade{}, 0},
		{"no_losses", []Trade{{Profit: 100}, {Profit: 0}, {Profit: 50}}, 0},
		{"single_loss", []Trade{{Profit: -150}}, -150},
		{"multiple_losses", []Trade{{Profit: 100}, {Profit: -50}, {Profit: 200}, {Profit: -150}}, -100},
		{"all_losing", []Trade{{Profit: -50}, {Profit: -100}, {Profit: -150}}, -100},
		{"losses_with_breakevens", []Trade{{Profit: -200}, {Profit: 0}, {Profit: -400}}, -300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AvgLosingTrade(tt.trades)
			if result != tt.expected {
				t.Errorf("AvgLosingTrade = %.6f, want %.6f", result, tt.expected)
			}
		})
	}
}

func TestCalcOpenProfit(t *testing.T) {
	tests := []struct {
		name         string
		trades       []Trade
		currentPrice float64
		expected     float64
	}{
		{"no_trades", []Trade{}, 110.0, 0},
		{"long_at_cost", []Trade{{Direction: Long, EntryPrice: 100, Size: 10}}, 100.0, 0},
		{"long_profit", []Trade{{Direction: Long, EntryPrice: 100, Size: 10}}, 110.0, 100},
		{"long_loss", []Trade{{Direction: Long, EntryPrice: 100, Size: 10}}, 90.0, -100},
		{"short_profit", []Trade{{Direction: Short, EntryPrice: 100, Size: 5}}, 90.0, 50},
		{"short_loss", []Trade{{Direction: Short, EntryPrice: 100, Size: 5}}, 110.0, -50},
		{
			"mixed_long_short",
			[]Trade{
				{Direction: Long, EntryPrice: 100, Size: 10},
				{Direction: Short, EntryPrice: 100, Size: 5},
			},
			110.0,
			50, // long: +100, short: -50
		},
		{"fractional_size", []Trade{{Direction: Long, EntryPrice: 100, Size: 0.5}}, 120.0, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalcOpenProfit(tt.trades, tt.currentPrice)
			if result != tt.expected {
				t.Errorf("CalcOpenProfit = %.6f, want %.6f", result, tt.expected)
			}
		})
	}
}

func TestAggregatorInvariants(t *testing.T) {
	tests := []struct {
		name   string
		trades []Trade
	}{
		{"empty", []Trade{}},
		{"all_winning", []Trade{{Profit: 100}, {Profit: 200}, {Profit: 300}}},
		{"all_losing", []Trade{{Profit: -100}, {Profit: -200}}},
		{"mixed", []Trade{{Profit: 300}, {Profit: -100}, {Profit: 0}, {Profit: 200}, {Profit: -50}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wins := CountWinningTrades(tt.trades)
			losses := CountLosingTrades(tt.trades)

			// AvgWinningTrade > 0 iff wins > 0
			avgWin := AvgWinningTrade(tt.trades)
			if wins > 0 && avgWin <= 0 {
				t.Errorf("AvgWinningTrade must be > 0 when wins=%d, got %.2f", wins, avgWin)
			}
			if wins == 0 && avgWin != 0 {
				t.Errorf("AvgWinningTrade must be 0 when no wins, got %.2f", avgWin)
			}

			// AvgLosingTrade < 0 iff losses > 0
			avgLoss := AvgLosingTrade(tt.trades)
			if losses > 0 && avgLoss >= 0 {
				t.Errorf("AvgLosingTrade must be < 0 when losses=%d, got %.2f", losses, avgLoss)
			}
			if losses == 0 && avgLoss != 0 {
				t.Errorf("AvgLosingTrade must be 0 when no losses, got %.2f", avgLoss)
			}

			// AvgTrade equals sum / count
			if len(tt.trades) > 0 {
				total := 0.0
				for _, tr := range tt.trades {
					total += tr.Profit
				}
				expected := total / float64(len(tt.trades))
				if AvgTrade(tt.trades) != expected {
					t.Errorf("AvgTrade invariant failed: got %.6f, want %.6f", AvgTrade(tt.trades), expected)
				}
			}

			// GrossProfit + GrossLoss == sum of all profits
			total := AggregateGrossProfit(tt.trades) + AggregateGrossLoss(tt.trades)
			sum := 0.0
			for _, tr := range tt.trades {
				sum += tr.Profit
			}
			if total != sum {
				t.Errorf("GrossProfit+GrossLoss invariant: got %.6f, want %.6f", total, sum)
			}
		})
	}
}
