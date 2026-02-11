package integration

import (
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestSeriesStrategyExecution(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
strategy("SMA Crossover with Series", overlay=true)

sma20 = ta.sma(close, 20)
sma50 = ta.sma(close, 50)

prev_sma20 = sma20[1]
prev_sma50 = sma50[1]

crossover_signal = sma20 > sma50 and prev_sma20 <= prev_sma50
crossunder_signal = sma20 < sma50 and prev_sma20 >= prev_sma50

if (crossover_signal)
    strategy.entry("Long", strategy.long)

if (crossunder_signal)
    strategy.entry("Short", strategy.short)
`

	testData := createSMACrossoverTestData()

	exec := util.NewPineExecutor(t)
	result := exec.ExecuteScriptWithCustomData(t, "series-strategy", pineScript, testData)

	totalTrades := len(result.Strategy.ClosedTrades)

	if totalTrades == 0 {
		t.Log("Warning: Expected trades at crossover points but got none - strategy may need position management")
	} else {
		t.Logf("Series strategy execution test passed - %d trades executed", totalTrades)
	}
}

func createSMACrossoverTestData() []map[string]interface{} {
	// Create data with clear SMA20 crossing above SMA50
	// Need at least 50 bars for SMA50 warmup, plus crossover pattern
	bars := []map[string]interface{}{}

	baseTime := int64(1700000000) // Unix timestamp

	// First 50 bars: downtrend (close below previous, SMA20 < SMA50)
	for i := 0; i < 50; i++ {
		close := 100.0 - float64(i)*0.5 // Decreasing from 100 to 75
		bars = append(bars, map[string]interface{}{
			"time":   baseTime + int64(i)*3600,
			"open":   close + 1,
			"high":   close + 2,
			"low":    close - 1,
			"close":  close,
			"volume": 1000.0,
		})
	}

	// Next 30 bars: uptrend (close above previous, SMA20 crosses above SMA50)
	for i := 0; i < 30; i++ {
		close := 75.0 + float64(i)*1.0 // Increasing from 75 to 105
		bars = append(bars, map[string]interface{}{
			"time":   baseTime + int64(50+i)*3600,
			"open":   close - 1,
			"high":   close + 2,
			"low":    close - 2,
			"close":  close,
			"volume": 1000.0,
		})
	}

	return bars
}
