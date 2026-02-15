package golden

import (
	"testing"
)

func TestTickerStandard_BTCUSDT_1D(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Ticker Standard Security",
		StrategyFile: "test-ticker-standard.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1D",
		DataFile:     "BTCUSDT_1D.json",
		GoldenFile:   "ticker_standard_btcusdt_1d.golden.json",
	})
}

func TestTickerStandard_AAPL_1D(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Ticker Standard Security",
		StrategyFile: "test-ticker-standard.pine",
		Symbol:       "AAPL",
		Timeframe:    "1D",
		DataFile:     "AAPL_1D.json",
		GoldenFile:   "ticker_standard_aapl_1d.golden.json",
	})
}
