package golden

import (
	"testing"
)

func TestMoonPhases_BTCUSDT_Monthly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Moon Phases Strategy [LuxAlgo]",
		StrategyFile: "top10/moon.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1M",
		DataFile:     "BTCUSDT-M.json",
		GoldenFile:   "moon_phases_btcusdt_m.golden.json",
	})
}

func TestMoonPhases_AAPL_Monthly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Moon Phases Strategy [LuxAlgo]",
		StrategyFile: "top10/moon.pine",
		Symbol:       "AAPL",
		Timeframe:    "1M",
		DataFile:     "AAPL-M.json",
		GoldenFile:   "moon_phases_aapl_m.golden.json",
	})
}

func TestMoonPhases_SBERP_Monthly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Moon Phases Strategy [LuxAlgo]",
		StrategyFile: "top10/moon.pine",
		Symbol:       "SBERP",
		Timeframe:    "1M",
		DataFile:     "SBERP-M.json",
		GoldenFile:   "moon_phases_sberp_m.golden.json",
	})
}
