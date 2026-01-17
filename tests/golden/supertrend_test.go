package golden

import (
	"testing"
)

func TestSupertrend_AAPL_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Supertrend",
		StrategyFile: "supertrend.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "supertrend-aapl-1h.json",
	})
}

func TestSupertrend_BTCUSDT_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Supertrend",
		StrategyFile: "supertrend.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "supertrend-btcusdt-1h.json",
	})
}

func TestSupertrend_SBERP_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Supertrend",
		StrategyFile: "supertrend.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "supertrend-sberp-1h.json",
	})
}
