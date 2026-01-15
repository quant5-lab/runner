package golden

import (
	"testing"
)

func TestADXDI_AAPL_Hourly(t *testing.T) {
	t.Skip("Parser limitation: v4 tooltip with newlines - see adx-di-strategy.pine.skip")
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "ADX + DI Strategy",
		StrategyFile: "adx-di-strategy.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "adx-di-aapl-1h.json",
	})
}

func TestADXDI_BTCUSDT_Hourly(t *testing.T) {
	t.Skip("Parser limitation: v4 tooltip with newlines - see adx-di-strategy.pine.skip")
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "ADX + DI Strategy",
		StrategyFile: "adx-di-strategy.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "adx-di-btcusdt-1h.json",
	})
}

func TestADXDI_SBERP_Hourly(t *testing.T) {
	t.Skip("Parser limitation: v4 tooltip with newlines - see adx-di-strategy.pine.skip")
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "ADX + DI Strategy",
		StrategyFile: "adx-di-strategy.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "adx-di-sberp-1h.json",
	})
}
