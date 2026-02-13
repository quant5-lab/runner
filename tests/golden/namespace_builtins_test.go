package golden

import (
	"testing"
)

func TestNamespaceBuiltins_AAPL_Daily(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "NamespaceBuiltins",
		StrategyFile: "test-namespace-builtins.pine",
		Symbol:       "AAPL",
		Timeframe:    "1D",
		DataFile:     "AAPL_1D.json",
		GoldenFile:   "namespace-builtins-aapl-1d.json",
	})
}
