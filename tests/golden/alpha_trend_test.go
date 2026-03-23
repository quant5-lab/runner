package golden

import (
	"testing"
)

func TestAlphaTrend_AAPL_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "AlphaTrend Strategy",
		StrategyFile: "top10/alpha.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "alpha_trend_aapl_1h.golden.json",
	})
}

func TestAlphaTrend_BTCUSDT_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "AlphaTrend Strategy",
		StrategyFile: "top10/alpha.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "alpha_trend_btcusdt_1h.golden.json",
	})
}

func TestAlphaTrend_SBERP_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "AlphaTrend Strategy",
		StrategyFile: "top10/alpha.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "alpha_trend_sberp_1h.golden.json",
	})
}
