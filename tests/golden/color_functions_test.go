package golden

import (
	"testing"
)

func TestColorFunctions_BTCUSDT_1h(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Color Functions",
		StrategyFile: "test-color-functions.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "color_functions_btcusdt_1h.golden.json",
	})
}

func TestColorFunctions_AAPL_1h(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Color Functions",
		StrategyFile: "test-color-functions.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "color_functions_aapl_1h.golden.json",
	})
}
