package golden

import (
	"testing"
)

func TestMomentumCascade_RBLX_1h(t *testing.T) {
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
