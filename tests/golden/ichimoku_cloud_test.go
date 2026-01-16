package golden

import (
	"testing"
)

func TestIchimokuCloud_AAPL_Hourly(t *testing.T) {
	t.Skip("Codegen limitation: ta.ichimoku() not implemented - see ichimoku-cloud-strategy.pine.skip")
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Ichimoku Cloud",
		StrategyFile: "ichimoku-cloud-strategy.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "ichimoku-cloud-aapl-1h.json",
	})
}

func TestIchimokuCloud_BTCUSDT_Hourly(t *testing.T) {
	t.Skip("Codegen limitation: ta.ichimoku() not implemented - see ichimoku-cloud-strategy.pine.skip")
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Ichimoku Cloud",
		StrategyFile: "ichimoku-cloud-strategy.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "ichimoku-cloud-btcusdt-1h.json",
	})
}

func TestIchimokuCloud_SBERP_Hourly(t *testing.T) {
	t.Skip("Codegen limitation: ta.ichimoku() not implemented - see ichimoku-cloud-strategy.pine.skip")
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Ichimoku Cloud",
		StrategyFile: "ichimoku-cloud-strategy.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "ichimoku-cloud-sberp-1h.json",
	})
}
