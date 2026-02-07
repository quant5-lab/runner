package golden

import (
	"testing"
)

func TestMathFunctions_AAPL_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "MathFunctions",
		StrategyFile: "test-math-functions.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "math-functions-aapl-1h.json",
	})
}

func TestMathFunctions_BTCUSDT_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "MathFunctions",
		StrategyFile: "test-math-functions.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "math-functions-btcusdt-1h.json",
	})
}

func TestMathFunctions_SBERP_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "MathFunctions",
		StrategyFile: "test-math-functions.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "math-functions-sberp-1h.json",
	})
}
