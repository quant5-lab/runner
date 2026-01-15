package golden

import (
	"testing"
)

func TestBB8_AAPL_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "BB8",
		StrategyFile: "bb-strategy-8-rus.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "bb8-aapl-1h.json",
	})
}

func TestBB8_BTCUSDT_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "BB8",
		StrategyFile: "bb-strategy-8-rus.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "bb8-btcusdt-1h.json",
	})
}

func TestBB8_SBERP_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "BB8",
		StrategyFile: "bb-strategy-8-rus.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "bb8-sberp-1h.json",
	})
}
