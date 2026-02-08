package golden

import (
	"testing"
)

/* Regression: nz/na as arguments inside arrow function bodies (BLOCKER #23) */

func TestValueFunctionsArrowArgs_BTCUSDT_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Value Functions Arrow Args",
		StrategyFile: "test-value-functions-arrow-args.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "value_functions_arrow_args_btcusdt_1h.golden.json",
	})
}

func TestValueFunctionsArrowArgs_AAPL_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Value Functions Arrow Args",
		StrategyFile: "test-value-functions-arrow-args.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "value_functions_arrow_args_aapl_1h.golden.json",
	})
}
