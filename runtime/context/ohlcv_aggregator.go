package context

// AggregateToCoarserPeriod uses session-anchored boundaries so that coarser-period
// slot boundaries match TradingView for non-UTC exchanges (NYSE, MOEX, etc.).
//
// timeframe must be coarser than the primary resolution and an integer multiple
// of the primary period — otherwise the slot count will be lower than expected.
func AggregateToCoarserPeriod(bars []OHLCV, timeframe string, anchor PeriodAnchor) []OHLCV {
	if len(bars) == 0 {
		return nil
	}

	result := make([]OHLCV, 0, estimatedSlotCount(len(bars)))
	var current OHLCV
	currentSlotOpen := int64(-1)

	for _, bar := range bars {
		slotOpen := AlignTimestampToPeriodWithAnchor(bar.Time, timeframe, anchor)
		if slotOpen != currentSlotOpen {
			if currentSlotOpen >= 0 {
				result = append(result, current)
			}
			current = OHLCV{
				Time:   slotOpen,
				Open:   bar.Open,
				High:   bar.High,
				Low:    bar.Low,
				Close:  bar.Close,
				Volume: bar.Volume,
			}
			currentSlotOpen = slotOpen
		} else {
			if bar.High > current.High {
				current.High = bar.High
			}
			if bar.Low < current.Low {
				current.Low = bar.Low
			}
			current.Close = bar.Close
			current.Volume += bar.Volume
		}
	}
	if currentSlotOpen >= 0 {
		result = append(result, current)
	}
	return result
}

func estimatedSlotCount(primaryBarCount int) int {
	estimated := primaryBarCount / 4
	if estimated < 1 {
		return 1
	}
	return estimated
}
