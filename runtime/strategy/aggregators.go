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
