package golden

import (
	"testing"

	"github.com/quant5-lab/runner/tests/golden/testutil"
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

var moonSymbols = []struct {
	name     string
	symbol   string
	dataFile string
}{
	{name: "btcusdt_monthly", symbol: "BTCUSDT", dataFile: "BTCUSDT-M.json"},
	{name: "aapl_monthly", symbol: "AAPL", dataFile: "AAPL-M.json"},
	{name: "sberp_monthly", symbol: "SBERP", dataFile: "SBERP-M.json"},
}

// buy=FullMoon and sell=NewMoon are mutually exclusive events, so consecutive trades must alternate.
func TestMoonPhases_AlternatingDirections(t *testing.T) {
	t.Parallel()

	for _, tc := range moonSymbols {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			suite := NewTestSuite(t)

			result := suite.runner.Execute(t,
				suite.StrategyPath("top10/moon.pine"),
				suite.DataPath(tc.dataFile),
				tc.symbol, "1M")

			validateAlternatingDirections(t, result)
		})
	}
}

// process_orders_on_close forces every reversal onto the same bar, so exit and entry timestamps must coincide.
func TestMoonPhases_PositionReversalTiming(t *testing.T) {
	t.Parallel()

	for _, tc := range moonSymbols {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			suite := NewTestSuite(t)

			result := suite.runner.Execute(t,
				suite.StrategyPath("top10/moon.pine"),
				suite.DataPath(tc.dataFile),
				tc.symbol, "1M")

			validatePositionReversalBehavior(t, result)
		})
	}
}

func TestMoonPhases_EquityConsistency(t *testing.T) {
	t.Parallel()

	for _, tc := range moonSymbols {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			suite := NewTestSuite(t)

			result := suite.runner.Execute(t,
				suite.StrategyPath("top10/moon.pine"),
				suite.DataPath(tc.dataFile),
				tc.symbol, "1M")

			testutil.ValidateEquityConsistency(t, result)
		})
	}
}
