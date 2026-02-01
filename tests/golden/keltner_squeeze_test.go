package golden

import (
	"testing"
)

func TestKeltnerSqueeze_AAPL_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Keltner Squeeze",
		StrategyFile: "keltner-squeeze.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "keltner-squeeze-aapl-1h.json",
	})
}

func TestKeltnerSqueeze_BTCUSDT_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Keltner Squeeze",
		StrategyFile: "keltner-squeeze.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "keltner-squeeze-btcusdt-1h.json",
	})
}

func TestKeltnerSqueeze_SBERP_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Keltner Squeeze",
		StrategyFile: "keltner-squeeze.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "keltner-squeeze-sberp-1h.json",
	})
}

func TestKeltnerSqueeze_PLZL_Hourly(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Keltner Squeeze",
		StrategyFile: "keltner-squeeze.pine",
		Symbol:       "PLZL",
		Timeframe:    "1h",
		DataFile:     "PLZL-1h.json",
		GoldenFile:   "keltner-squeeze-plzl-1h.json",
	})
}
