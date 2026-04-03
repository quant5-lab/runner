package strategy

import (
	"math"
	"testing"
)

func TestDefaultQtyCalculator_CalculateQty(t *testing.T) {
	calc := NewDefaultQtyCalculator()

	tests := []struct {
		name          string
		qtyType       string
		qtyValue      float64
		fillPrice     float64
		currentEquity float64
		expected      float64
	}{
		{
			name:          "fixed type returns qty value directly",
			qtyType:       QtyTypeFixed,
			qtyValue:      10,
			fillPrice:     50,
			currentEquity: 10000,
			expected:      10,
		},
		{
			name:          "strategy.fixed prefixed type",
			qtyType:       "strategy.fixed",
			qtyValue:      5,
			fillPrice:     100,
			currentEquity: 10000,
			expected:      5,
		},
		{
			name:          "empty type defaults to fixed",
			qtyType:       "",
			qtyValue:      3,
			fillPrice:     50,
			currentEquity: 10000,
			expected:      3,
		},
		{
			name:          "cash type divides by fill price",
			qtyType:       QtyTypeCash,
			qtyValue:      1000,
			fillPrice:     50,
			currentEquity: 10000,
			expected:      20,
		},
		{
			name:          "strategy.cash prefixed type",
			qtyType:       "strategy.cash",
			qtyValue:      2500,
			fillPrice:     100,
			currentEquity: 10000,
			expected:      25,
		},
		{
			name:          "percent_of_equity calculates from equity",
			qtyType:       QtyTypePercentOfEquity,
			qtyValue:      10,
			fillPrice:     50,
			currentEquity: 10000,
			expected:      20,
		},
		{
			name:          "strategy.percent_of_equity prefixed type",
			qtyType:       "strategy.percent_of_equity",
			qtyValue:      25,
			fillPrice:     100,
			currentEquity: 10000,
			expected:      25,
		},
		{
			name:          "percent_of_equity with fractional result",
			qtyType:       QtyTypePercentOfEquity,
			qtyValue:      15,
			fillPrice:     100,
			currentEquity: 10000,
			expected:      15,
		},
		{
			name:          "zero fill price returns zero",
			qtyType:       QtyTypeCash,
			qtyValue:      1000,
			fillPrice:     0,
			currentEquity: 10000,
			expected:      0,
		},
		{
			name:          "negative fill price returns zero",
			qtyType:       QtyTypeCash,
			qtyValue:      1000,
			fillPrice:     -50,
			currentEquity: 10000,
			expected:      0,
		},
		{
			name:          "NaN fill price returns zero",
			qtyType:       QtyTypeCash,
			qtyValue:      1000,
			fillPrice:     math.NaN(),
			currentEquity: 10000,
			expected:      0,
		},
		{
			name:          "zero equity with percent_of_equity returns zero",
			qtyType:       QtyTypePercentOfEquity,
			qtyValue:      50,
			fillPrice:     100,
			currentEquity: 0,
			expected:      0,
		},
		{
			name:          "negative equity with percent_of_equity returns zero",
			qtyType:       QtyTypePercentOfEquity,
			qtyValue:      50,
			fillPrice:     100,
			currentEquity: -1000,
			expected:      0,
		},
		{
			name:          "unknown type defaults to fixed behavior",
			qtyType:       "unknown_type",
			qtyValue:      7,
			fillPrice:     100,
			currentEquity: 10000,
			expected:      7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateQty(tt.qtyType, tt.qtyValue, tt.fillPrice, tt.currentEquity)

			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("CalculateQty() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
