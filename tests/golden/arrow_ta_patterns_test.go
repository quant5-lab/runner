package golden

import (
	"testing"
)

/* Pattern: ta.atr, ta.tr - functions with implicit OHLC sources */
func TestArrow_ImplicitOHLC_BTCUSDT(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Arrow-ImplicitOHLC",
		StrategyFile: "test-arrow-implicit-ohlc.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1D",
		DataFile:     "BTCUSDT_1D.json",
		GoldenFile:   "arrow-implicit-ohlc-btcusdt-1d.json",
	})
}

/* Pattern: ta.pivothigh, ta.pivotlow - 2-arg vs 3-arg overloads */
func TestArrow_MultiargOverload_BTCUSDT(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Arrow-MultiargOverload",
		StrategyFile: "test-arrow-multiarg-overload.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1D",
		DataFile:     "BTCUSDT_1D.json",
		GoldenFile:   "arrow-multiarg-overload-btcusdt-1d.json",
	})
}
