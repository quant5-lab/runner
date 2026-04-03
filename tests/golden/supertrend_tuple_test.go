package golden

import (
	"testing"
)

func TestSupertrendTuple_AAPL_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Supertrend Tuple",
		StrategyFile: "supertrend-tuple.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "supertrend-tuple-aapl-1h.json",
	})
}

func TestSupertrendTuple_BTCUSDT_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Supertrend Tuple",
		StrategyFile: "supertrend-tuple.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "supertrend-tuple-btcusdt-1h.json",
	})
}

func TestSupertrendTuple_SBERP_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Supertrend Tuple",
		StrategyFile: "supertrend-tuple.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "supertrend-tuple-sberp-1h.json",
	})
}
