package strategy

func AggregateGrossProfit(trades []Trade) float64 {
	total := 0.0
	for _, t := range trades {
		if t.Profit > 0 {
			total += t.Profit
		}
	}
	return total
}

func AggregateGrossLoss(trades []Trade) float64 {
	total := 0.0
	for _, t := range trades {
		if t.Profit < 0 {
			total += t.Profit
		}
	}
	return total
}

func CountWinningTrades(trades []Trade) int {
	count := 0
	for _, t := range trades {
		if t.Profit > 0 {
			count++
		}
	}
	return count
}

func CountLosingTrades(trades []Trade) int {
	count := 0
	for _, t := range trades {
		if t.Profit < 0 {
			count++
		}
	}
	return count
}

func CountEvenTrades(trades []Trade) int {
	count := 0
	for _, t := range trades {
		if t.Profit == 0 {
			count++
		}
	}
	return count
}

func AvgTrade(trades []Trade) float64 {
	if len(trades) == 0 {
		return 0
	}
	total := 0.0
	for _, t := range trades {
		total += t.Profit
	}
	return total / float64(len(trades))
}

func AvgWinningTrade(trades []Trade) float64 {
	wins := CountWinningTrades(trades)
	if wins == 0 {
		return 0
	}
	return AggregateGrossProfit(trades) / float64(wins)
}

func AvgLosingTrade(trades []Trade) float64 {
	losses := CountLosingTrades(trades)
	if losses == 0 {
		return 0
	}
	return AggregateGrossLoss(trades) / float64(losses)
}

func CalcOpenProfit(openTrades []Trade, currentPrice float64) float64 {
	total := 0.0
	for _, t := range openTrades {
		multiplier := 1.0
		if t.Direction == Short {
			multiplier = -1.0
		}
		total += (currentPrice - t.EntryPrice) * t.Size * multiplier
	}
	return total
}
