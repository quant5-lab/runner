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
