package golden

import (
	"testing"
)

func TestVWAP_AAPL_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Volume Weighted Strategy",
		StrategyFile: "vwap-strategy.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "vwap-aapl-1h.json",
	})
}

func TestVWAP_BTCUSDT_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Volume Weighted Strategy",
		StrategyFile: "vwap-strategy.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "vwap-btcusdt-1h.json",
	})
}

func TestVWAP_SBERP_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Volume Weighted Strategy",
		StrategyFile: "vwap-strategy.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "vwap-sberp-1h.json",
	})
}
