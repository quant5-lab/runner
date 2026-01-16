package golden

import (
	"testing"
)

func TestMTF_AAPL_Hourly(t *testing.T) {
	t.Skip("Runtime bug: ta.crossover IIFE incorrect previous bar access - see mtf-confirmation-strategy.pine.skip")
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
	t.Skip("Runtime bug: ta.crossover IIFE incorrect previous bar access - see mtf-confirmation-strategy.pine.skip")
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
	t.Skip("Runtime bug: ta.crossover IIFE incorrect previous bar access - see mtf-confirmation-strategy.pine.skip")
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
