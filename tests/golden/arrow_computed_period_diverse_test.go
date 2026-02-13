package golden

import (
	"testing"
)

/* Regression: RSI, highest, lowest, SMA with computed periods exercise distinct IIFE generators */

func TestArrowComputedPeriodDiverse_BTCUSDT_1h(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Arrow Computed Period Diverse",
		StrategyFile: "test-arrow-computed-period-diverse.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "arrow_computed_period_diverse_btcusdt_1h.golden.json",
	})
}

func TestArrowComputedPeriodDiverse_AAPL_1h(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Arrow Computed Period Diverse",
		StrategyFile: "test-arrow-computed-period-diverse.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "arrow_computed_period_diverse_aapl_1h.golden.json",
	})
}
