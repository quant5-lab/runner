package strategy

import "math"

const (
	QtyTypeFixed           = "fixed"
	QtyTypeCash            = "cash"
	QtyTypePercentOfEquity = "percent_of_equity"
)

type DefaultQtyCalculator struct{}

func NewDefaultQtyCalculator() *DefaultQtyCalculator {
	return &DefaultQtyCalculator{}
}

func (c *DefaultQtyCalculator) CalculateQty(qtyType string, qtyValue, fillPrice, currentEquity float64) float64 {
	if math.IsNaN(fillPrice) || fillPrice <= 0 {
		return 0
	}

	switch qtyType {
	case QtyTypeCash, "strategy.cash":
		return qtyValue / fillPrice

	case QtyTypePercentOfEquity, "strategy.percent_of_equity":
		if currentEquity <= 0 {
			return 0
		}
		return (currentEquity * qtyValue / 100.0) / fillPrice

	case QtyTypeFixed, "strategy.fixed", "":
		return qtyValue

	default:
		return qtyValue
	}
}
