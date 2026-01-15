package golden

import (
	"testing"
)

func TestBB9_AAPL_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "BB9",
		StrategyFile: "bb-strategy-9-rus.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "bb9-aapl-1h.json",
	})
}

func TestBB9_BTCUSDT_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "BB9",
		StrategyFile: "bb-strategy-9-rus.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "bb9-btcusdt-1h.json",
	})
}

func TestBB9_SBERP_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "BB9",
		StrategyFile: "bb-strategy-9-rus.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "bb9-sberp-1h.json",
	})
}
