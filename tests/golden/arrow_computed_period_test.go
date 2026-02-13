package golden

import (
	"testing"
)

/* Regression: computed period expressions (_length/2, math.round(math.sqrt(_length))) in arrow function TA calls */

func TestArrowComputedPeriod_BTCUSDT_1h(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Arrow Computed Period",
		StrategyFile: "test-arrow-computed-period.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "arrow_computed_period_btcusdt_1h.golden.json",
	})
}

func TestArrowComputedPeriod_AAPL_1h(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Arrow Computed Period",
		StrategyFile: "test-arrow-computed-period.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "arrow_computed_period_aapl_1h.golden.json",
	})
}
