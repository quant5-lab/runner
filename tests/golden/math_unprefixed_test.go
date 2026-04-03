package golden

import (
	"testing"
)

func TestMathUnprefixed_AAPL_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "MathUnprefixed",
		StrategyFile: "test-math-unprefixed.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "math-unprefixed-aapl-1h.json",
	})
}

func TestMathUnprefixed_BTCUSDT_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "MathUnprefixed",
		StrategyFile: "test-math-unprefixed.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "math-unprefixed-btcusdt-1h.json",
	})
}

func TestMathUnprefixed_SBERP_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "MathUnprefixed",
		StrategyFile: "test-math-unprefixed.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "math-unprefixed-sberp-1h.json",
	})
}
