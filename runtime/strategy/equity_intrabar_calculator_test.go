package strategy

import (
	"testing"
)

func TestIntrabarEquityCalculator_DirectionalPriceSelection(t *testing.T) {
	tests := []struct {
		name               string
		direction          string
		entryPrice         float64
		barHigh            float64
		barLow             float64
		wantAdversePrice   float64
		wantFavorablePrice float64
	}{
		{
			name:               "long_uses_low_for_adverse_high_for_favorable",
			direction:          Long,
			entryPrice:         100,
			barHigh:            110,
			barLow:             90,
			wantAdversePrice:   90,
			wantFavorablePrice: 110,
		},
		{
			name:               "short_uses_high_for_adverse_low_for_favorable",
			direction:          Short,
			entryPrice:         100,
			barHigh:            110,
			barLow:             90,
			wantAdversePrice:   110,
			wantFavorablePrice: 90,
		},
		{
			name:               "long_doji_bar_same_price",
			direction:          Long,
			entryPrice:         100,
			barHigh:            100,
			barLow:             100,
			wantAdversePrice:   100,
			wantFavorablePrice: 100,
		},
		{
			name:               "short_doji_bar_same_price",
			direction:          Short,
			entryPrice:         100,
			barHigh:            100,
			barLow:             100,
			wantAdversePrice:   100,
			wantFavorablePrice: 100,
		},
	}

	calc := NewIntrabarEquityCalculator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adversePrice := calc.selectAdversePrice(tt.direction, tt.barHigh, tt.barLow)
			if adversePrice != tt.wantAdversePrice {
				t.Errorf("adversePrice = %.2f, want %.2f", adversePrice, tt.wantAdversePrice)
			}

			favorablePrice := calc.selectFavorablePrice(tt.direction, tt.barHigh, tt.barLow)
			if favorablePrice != tt.wantFavorablePrice {
				t.Errorf("favorablePrice = %.2f, want %.2f", favorablePrice, tt.wantFavorablePrice)
			}
		})
	}
}

func TestIntrabarEquityCalculator_EquityCalculation(t *testing.T) {
	tests := []struct {
		name           string
		trades         []Trade
		realizedProfit float64
		initialCapital float64
		barHigh        float64
		barLow         float64
		wantAdverse    float64
		wantFavorable  float64
	}{
		{
			name: "single_long_in_profit",
			trades: []Trade{
				{Direction: Long, Size: 10, EntryPrice: 100},
			},
			realizedProfit: 0,
			initialCapital: 10000,
			barHigh:        110,
			barLow:         90,
			wantAdverse:    10000 + (90-100)*10,
			wantFavorable:  10000 + (110-100)*10,
		},
		{
			name: "single_long_in_loss",
			trades: []Trade{
				{Direction: Long, Size: 10, EntryPrice: 100},
			},
			realizedProfit: 0,
			initialCapital: 10000,
			barHigh:        95,
			barLow:         85,
			wantAdverse:    10000 + (85-100)*10,
			wantFavorable:  10000 + (95-100)*10,
		},
		{
			name: "single_short_in_profit",
			trades: []Trade{
				{Direction: Short, Size: 10, EntryPrice: 100},
			},
			realizedProfit: 0,
			initialCapital: 10000,
			barHigh:        110,
			barLow:         90,
			wantAdverse:    10000 + (100-110)*10,
			wantFavorable:  10000 + (100-90)*10,
		},
		{
			name: "single_short_in_loss",
			trades: []Trade{
				{Direction: Short, Size: 10, EntryPrice: 100},
			},
			realizedProfit: 0,
			initialCapital: 10000,
			barHigh:        115,
			barLow:         105,
			wantAdverse:    10000 + (100-115)*10,
			wantFavorable:  10000 + (100-105)*10,
		},
		{
			name: "multiple_longs_different_entries",
			trades: []Trade{
				{Direction: Long, Size: 10, EntryPrice: 100},
				{Direction: Long, Size: 5, EntryPrice: 95},
			},
			realizedProfit: 0,
			initialCapital: 10000,
			barHigh:        110,
			barLow:         90,
			wantAdverse:    10000 + (90-100)*10 + (90-95)*5,
			wantFavorable:  10000 + (110-100)*10 + (110-95)*5,
		},
		{
			name: "multiple_shorts_different_entries",
			trades: []Trade{
				{Direction: Short, Size: 10, EntryPrice: 100},
				{Direction: Short, Size: 5, EntryPrice: 105},
			},
			realizedProfit: 0,
			initialCapital: 10000,
			barHigh:        110,
			barLow:         90,
			wantAdverse:    10000 + (100-110)*10 + (105-110)*5,
			wantFavorable:  10000 + (100-90)*10 + (105-90)*5,
		},
		{
			name: "mixed_directions_hedged",
			trades: []Trade{
				{Direction: Long, Size: 10, EntryPrice: 100},
				{Direction: Short, Size: 5, EntryPrice: 100},
			},
			realizedProfit: 0,
			initialCapital: 10000,
			barHigh:        110,
			barLow:         90,
			wantAdverse:    10000 + (90-100)*10 + (100-110)*5,
			wantFavorable:  10000 + (110-100)*10 + (100-90)*5,
		},
		{
			name:           "no_open_trades_flat_equity",
			trades:         []Trade{},
			realizedProfit: 250,
			initialCapital: 10000,
			barHigh:        110,
			barLow:         90,
			wantAdverse:    10250,
			wantFavorable:  10250,
		},
		{
			name: "with_realized_profit",
			trades: []Trade{
				{Direction: Long, Size: 10, EntryPrice: 100},
			},
			realizedProfit: 500,
			initialCapital: 10000,
			barHigh:        110,
			barLow:         90,
			wantAdverse:    10500 + (90-100)*10,
			wantFavorable:  10500 + (110-100)*10,
		},
		{
			name: "with_realized_loss",
			trades: []Trade{
				{Direction: Long, Size: 10, EntryPrice: 100},
			},
			realizedProfit: -300,
			initialCapital: 10000,
			barHigh:        110,
			barLow:         90,
			wantAdverse:    9700 + (90-100)*10,
			wantFavorable:  9700 + (110-100)*10,
		},
		{
			name: "fractional_sizes",
			trades: []Trade{
				{Direction: Long, Size: 0.5, EntryPrice: 100},
				{Direction: Short, Size: 0.25, EntryPrice: 100},
			},
			realizedProfit: 0,
			initialCapital: 10000,
			barHigh:        110,
			barLow:         90,
			wantAdverse:    10000 + (90-100)*0.5 + (100-110)*0.25,
			wantFavorable:  10000 + (110-100)*0.5 + (100-90)*0.25,
		},
	}

	calc := NewIntrabarEquityCalculator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adverse := calc.CalculateAdverseEquity(
				tt.trades,
				tt.realizedProfit,
				tt.initialCapital,
				tt.barHigh,
				tt.barLow,
			)

			if adverse != tt.wantAdverse {
				t.Errorf("adverse equity = %.2f, want %.2f", adverse, tt.wantAdverse)
			}

			favorable := calc.CalculateFavorableEquity(
				tt.trades,
				tt.realizedProfit,
				tt.initialCapital,
				tt.barHigh,
				tt.barLow,
			)

			if favorable != tt.wantFavorable {
				t.Errorf("favorable equity = %.2f, want %.2f", favorable, tt.wantFavorable)
			}
		})
	}
}

func TestIntrabarEquityCalculator_AdverseFavorableRelationship(t *testing.T) {
	calc := NewIntrabarEquityCalculator()

	tests := []struct {
		name      string
		trades    []Trade
		barHigh   float64
		barLow    float64
		wantOrder string
	}{
		{
			name:      "long_position_adverse_less_than_favorable",
			trades:    []Trade{{Direction: Long, Size: 10, EntryPrice: 100}},
			barHigh:   110,
			barLow:    90,
			wantOrder: "adverse_less_favorable",
		},
		{
			name:      "short_position_adverse_less_than_favorable",
			trades:    []Trade{{Direction: Short, Size: 10, EntryPrice: 100}},
			barHigh:   110,
			barLow:    90,
			wantOrder: "adverse_less_favorable",
		},
		{
			name:      "doji_bar_adverse_equals_favorable",
			trades:    []Trade{{Direction: Long, Size: 10, EntryPrice: 100}},
			barHigh:   100,
			barLow:    100,
			wantOrder: "adverse_equals_favorable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adverse := calc.CalculateAdverseEquity(tt.trades, 0, 10000, tt.barHigh, tt.barLow)
			favorable := calc.CalculateFavorableEquity(tt.trades, 0, 10000, tt.barHigh, tt.barLow)

			switch tt.wantOrder {
			case "adverse_less_favorable":
				if adverse >= favorable {
					t.Errorf("adverse (%.2f) should be < favorable (%.2f)", adverse, favorable)
				}
			case "adverse_equals_favorable":
				if adverse != favorable {
					t.Errorf("adverse (%.2f) should equal favorable (%.2f)", adverse, favorable)
				}
			}
		})
	}
}
