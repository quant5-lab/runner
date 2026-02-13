package golden

import (
	"testing"
)

func TestSecurityUserVariable_SBERP_Weekly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "Security User Variable Test",
		StrategyFile: "test-security-user-variable.pine",
		Symbol:       "SBERP",
		Timeframe:    "1D",
		DataFile:     "SBERP_1D.json",
		GoldenFile:   "security_user_variable_sberp_1d.golden.json",
	})
}
