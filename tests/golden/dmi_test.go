package golden

import (
	"testing"
)

func TestDMI_Basic_AAPL(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "DMI Basic",
		StrategyFile: "test-dmi-basic.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "dmi-basic-aapl-1h.json",
	})
}

func TestDMI_Basic_BTCUSDT(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "DMI Basic",
		StrategyFile: "test-dmi-basic.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "dmi-basic-btcusdt-1h.json",
	})
}

func TestDMI_Periods_AAPL(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "DMI Periods",
		StrategyFile: "test-dmi-periods.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "dmi-periods-aapl-1h.json",
	})
}

func TestDMI_Periods_SBERP(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "DMI Periods",
		StrategyFile: "test-dmi-periods.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "dmi-periods-sberp-1h.json",
	})
}

func TestDMI_Strategy_AAPL(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "DMI Strategy",
		StrategyFile: "test-dmi-strategy.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "dmi-strategy-aapl-1h.json",
	})
}

func TestDMI_Strategy_BTCUSDT(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "DMI Strategy",
		StrategyFile: "test-dmi-strategy.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "dmi-strategy-btcusdt-1h.json",
	})
}

func TestDMI_Crossover_AAPL(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "DMI Crossover",
		StrategyFile: "test-dmi-crossover.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "dmi-crossover-aapl-1h.json",
	})
}

func TestDMI_Crossover_CNRU(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "DMI Crossover",
		StrategyFile: "test-dmi-crossover.pine",
		Symbol:       "CNRU",
		Timeframe:    "1h",
		DataFile:     "CNRU-1h.json",
		GoldenFile:   "dmi-crossover-cnru-1h.json",
	})
}

func TestDMI_Comparison_AAPL(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "DMI Comparison",
		StrategyFile: "test-dmi-comparison.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "dmi-comparison-aapl-1h.json",
	})
}

func TestDMI_Comparison_SBERP(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "DMI Comparison",
		StrategyFile: "test-dmi-comparison.pine",
		Symbol:       "SBERP",
		Timeframe:    "1h",
		DataFile:     "SBERP-1h.json",
		GoldenFile:   "dmi-comparison-sberp-1h.json",
	})
}

func TestDMI_Historical_AAPL(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "DMI Historical",
		StrategyFile: "test-dmi-historical.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "dmi-historical-aapl-1h.json",
	})
}

func TestDMI_Historical_BTCUSDT(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "DMI Historical",
		StrategyFile: "test-dmi-historical.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "dmi-historical-btcusdt-1h.json",
	})
}
