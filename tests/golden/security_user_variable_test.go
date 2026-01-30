package golden

import (
	"testing"
)

func TestSecurityUserVariable_BTCUSDT_Daily(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Security User Variable Test",
		StrategyFile: "test-security-user-variable.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1D",
		DataFile:     "BTCUSDT_1D.json",
		GoldenFile:   "security_user_variable_btcusdt_1d.golden.json",
	})
}
