package golden

import (
	"testing"
)

// The top-10 SuperTrend member, which entered the set on 2026-10-14 from the
// inbox. It is a different strategy from strategies/supertrend.pine
// ("Supertrend [Alorse]"), so its goldens carry the top10- prefix and do not
// collide with that strategy's goldens.

func TestTop10Supertrend_AAPL_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Top10Supertrend",
		StrategyFile: "top10/supertrend.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "top10-supertrend-aapl-1h.json",
	})
}

func TestTop10Supertrend_BTCUSDT_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Top10Supertrend",
		StrategyFile: "top10/supertrend.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "top10-supertrend-btcusdt-1h.json",
	})
}

func TestTop10Supertrend_SBERP_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Top10Supertrend",
		StrategyFile: "top10/supertrend.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "top10-supertrend-sberp-1h.json",
	})
}
