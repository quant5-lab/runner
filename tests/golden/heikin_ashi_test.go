package golden

import (
	"testing"
)

func TestHeikinAshi_AAPL_Daily(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Heikin Ashi Test",
		StrategyFile: "heikin-ashi-test.pine",
		Symbol:       "AAPL",
		Timeframe:    "1D",
		DataFile:     "AAPL_1D.json",
		GoldenFile:   "heikin-ashi-aapl-1d.json",
	})
}

func TestHeikinAshi_BTCUSDT_Daily(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Heikin Ashi Test",
		StrategyFile: "heikin-ashi-test.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1D",
		DataFile:     "BTCUSDT_1D.json",
		GoldenFile:   "heikin-ashi-btcusdt-1d.json",
	})
}
