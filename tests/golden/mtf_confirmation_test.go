package golden

import (
	"testing"
)

func TestMTF_AAPL_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "MTF Confirmation",
		StrategyFile: "mtf-confirmation-strategy.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "mtf-confirmation-aapl-1h.json",
	})
}

func TestMTF_BTCUSDT_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "MTF Confirmation",
		StrategyFile: "mtf-confirmation-strategy.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "mtf-confirmation-btcusdt-1h.json",
	})
}

func TestMTF_SBERP_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "MTF Confirmation",
		StrategyFile: "mtf-confirmation-strategy.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "mtf-confirmation-sberp-1h.json",
	})
}

func TestMTF_NVDA_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "MTF Confirmation",
		StrategyFile: "mtf-confirmation-strategy.pine",
		Symbol:       "NVDA",
		Timeframe:    "1h",
		DataFile:     "NVDA-1h.json",
		GoldenFile:   "mtf-confirmation-nvda-1h.json",
	})
}
