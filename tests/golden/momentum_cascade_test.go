package golden

import (
	"testing"
)

func TestMomentumCascade_RBLX_1h(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "MomentumCascade",
		StrategyFile: "test-for-loop-momentum-cascade.pine",
		Symbol:       "RBLX",
		Timeframe:    "1h",
		DataFile:     "RBLX_1h.json",
		GoldenFile:   "momentum_cascade_rblx_1h.golden.json",
	})
}

func TestMomentumCascade_AAPL_1h(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Momentum Cascade Strategy",
		StrategyFile: "test-for-loop-momentum-cascade.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "momentum_cascade_aapl_1h.golden.json",
	})
}

func TestMomentumCascade_BTCUSDT_1h(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Momentum Cascade Strategy",
		StrategyFile: "test-for-loop-momentum-cascade.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "momentum_cascade_btcusdt_1h.golden.json",
	})
}

func TestMomentumCascade_BTCUSDT_1D(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Momentum Cascade Strategy",
		StrategyFile: "test-for-loop-momentum-cascade.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "D",
		DataFile:     "BTCUSDT_1D.json",
		GoldenFile:   "momentum_cascade_btcusdt_1d.golden.json",
	})
}
