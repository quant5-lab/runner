package golden

import (
	"testing"
)

/* Regression: math.abs/max/min/sqrt in arrow function bodies resolve scalar params via RouteCall delegation */

func TestArrowDelegatedMath_BTCUSDT_1h(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Arrow Delegated Math",
		StrategyFile: "test-arrow-delegated-math.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "arrow_delegated_math_btcusdt_1h.golden.json",
	})
}

func TestArrowDelegatedMath_AAPL_1h(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Arrow Delegated Math",
		StrategyFile: "test-arrow-delegated-math.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "arrow_delegated_math_aapl_1h.golden.json",
	})
}
