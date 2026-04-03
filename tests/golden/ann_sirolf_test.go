package golden

import (
	"testing"
)

func TestANNSirolf_BTCUSDT_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "ANN Strategy v2",
		StrategyFile: "ann-sirolf.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "ann_sirolf_btcusdt_1h.golden.json",
	})
}
