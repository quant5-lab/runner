package golden

import (
	"testing"
)

func TestBB7_AAPL_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "BB7",
		StrategyFile: "bb-strategy-7-rus.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "bb7-aapl-1h.json",
	})
}

func TestBB7_BTCUSDT_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "BB7",
		StrategyFile: "bb-strategy-7-rus.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "bb7-btcusdt-1h.json",
	})
}

func TestBB7_SBERP_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "BB7",
		StrategyFile: "bb-strategy-7-rus.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "bb7-sberp-1h.json",
	})
}
