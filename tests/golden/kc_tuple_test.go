package golden

import (
	"testing"
)

func TestKCTuple_AAPL_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "KC Tuple",
		StrategyFile: "kc-tuple.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "kc-tuple-aapl-1h.json",
	})
}

func TestKCTuple_BTCUSDT_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "KC Tuple",
		StrategyFile: "kc-tuple.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "kc-tuple-btcusdt-1h.json",
	})
}

func TestKCTuple_SBERP_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "KC Tuple",
		StrategyFile: "kc-tuple.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "kc-tuple-sberp-1h.json",
	})
}
