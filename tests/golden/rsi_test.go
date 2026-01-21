package golden

import (
	"testing"
)

func TestRSI_AAPL_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "RSI Strategy",
		StrategyFile: "rsi-strategy.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "rsi-aapl-1h.json",
	})
}

func TestRSI_BTCUSDT_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "RSI Strategy",
		StrategyFile: "rsi-strategy.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "rsi-btcusdt-1h.json",
	})
}

func TestRSI_NVDA_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "RSI Strategy",
		StrategyFile: "rsi-strategy.pine",
		Symbol:       "NVDA",
		Timeframe:    "1h",
		DataFile:     "NVDA-1h.json",
		GoldenFile:   "rsi-nvda-1h.json",
	})
}

func TestRSI_SBERP_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "RSI Strategy",
		StrategyFile: "rsi-strategy.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "rsi-sberp-1h.json",
	})
}
