package golden

import (
	"testing"
)

/* UT Bot Strategy - Pine v4 boolean direction syntax validation */

func TestUTBot_BTCUSDT_Daily(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "UT Bot Strategy",
		StrategyFile: "utbot-quantnomad.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1D",
		DataFile:     "BTCUSDT_1D.json",
		GoldenFile:   "utbot-btcusdt-1d.json",
	})
}

func TestUTBot_AAPL_Daily(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "UT Bot Strategy",
		StrategyFile: "utbot-quantnomad.pine",
		Symbol:       "AAPL",
		Timeframe:    "1D",
		DataFile:     "AAPL_1D.json",
		GoldenFile:   "utbot-aapl-1d.json",
	})
}
