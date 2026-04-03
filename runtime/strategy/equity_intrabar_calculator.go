package strategy

type IntrabarEquityCalculator struct{}

func NewIntrabarEquityCalculator() *IntrabarEquityCalculator {
	return &IntrabarEquityCalculator{}
}

func (c *IntrabarEquityCalculator) CalculateAdverseEquity(
	openTrades []Trade,
	realizedProfit float64,
	initialCapital float64,
	barHigh float64,
	barLow float64,
) float64 {
	unrealizedPL := 0.0

	for _, trade := range openTrades {
		adversePrice := c.selectAdversePrice(trade.Direction, barHigh, barLow)
		priceDiff := adversePrice - trade.EntryPrice
		multiplier := directionMultiplier(trade.Direction)
		unrealizedPL += priceDiff * trade.Size * multiplier
	}

	return initialCapital + realizedProfit + unrealizedPL
}

func (c *IntrabarEquityCalculator) CalculateFavorableEquity(
	openTrades []Trade,
	realizedProfit float64,
	initialCapital float64,
	barHigh float64,
	barLow float64,
) float64 {
	unrealizedPL := 0.0

	for _, trade := range openTrades {
		favorablePrice := c.selectFavorablePrice(trade.Direction, barHigh, barLow)
		priceDiff := favorablePrice - trade.EntryPrice
		multiplier := directionMultiplier(trade.Direction)
		unrealizedPL += priceDiff * trade.Size * multiplier
	}

	return initialCapital + realizedProfit + unrealizedPL
}

func (c *IntrabarEquityCalculator) selectAdversePrice(direction string, barHigh, barLow float64) float64 {
	if direction == Long {
		return barLow
	}
	return barHigh
}

func (c *IntrabarEquityCalculator) selectFavorablePrice(direction string, barHigh, barLow float64) float64 {
	if direction == Long {
		return barHigh
	}
	return barLow
}
