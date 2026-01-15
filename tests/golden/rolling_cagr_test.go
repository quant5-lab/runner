package golden

import (
	"testing"
)

func TestRollingCAGR_AAPL_Monthly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "RollingCAGR",
		StrategyFile: "rolling-cagr.pine",
		Symbol:       "AAPL",
		Timeframe:    "M",
		DataFile:     "AAPL-M.json",
		GoldenFile:   "rolling-cagr-aapl-m.json",
	})
}

func TestRollingCAGR_BTCUSDT_Monthly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "RollingCAGR",
		StrategyFile: "rolling-cagr.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "M",
		DataFile:     "BTCUSDT-M.json",
		GoldenFile:   "rolling-cagr-btcusdt-m.json",
	})
}

func TestRollingCAGR_SBERP_Monthly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "RollingCAGR",
		StrategyFile: "rolling-cagr.pine",
		Symbol:       "SBERP",
		Timeframe:    "M",
		DataFile:     "SBERP-M.json",
		GoldenFile:   "rolling-cagr-sberp-m.json",
	})
}
