package golden

import (
	"testing"
)

func TestMACD_AAPL_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "MACD Crossover",
		StrategyFile: "macd-crossover.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "macd-aapl-1h.json",
	})
}

func TestMACD_BTCUSDT_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "MACD Crossover",
		StrategyFile: "macd-crossover.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "macd-btcusdt-1h.json",
	})
}

func TestMACD_SBERP_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "MACD Crossover",
		StrategyFile: "macd-crossover.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "macd-sberp-1h.json",
	})
}
