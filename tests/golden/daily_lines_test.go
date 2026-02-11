package golden

import (
	"testing"
)

func TestDailyLines_Hourly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "DailyLines",
		StrategyFile: "daily-lines.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "daily-lines-aapl-1h.json",
	})
}

func TestDailyLines_Daily(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "DailyLines",
		StrategyFile: "daily-lines.pine",
		Symbol:       "AAPL",
		Timeframe:    "D",
		DataFile:     "AAPL-D.json",
		GoldenFile:   "daily-lines-aapl-d.json",
	})
}

func TestDailyLines_Weekly(t *testing.T) {
	t.Parallel()
	suite := NewTestSuite(t)

	suite.RunAndValidate(t, TestConfig{
		StrategyName: "DailyLines",
		StrategyFile: "daily-lines.pine",
		Symbol:       "AAPL",
		Timeframe:    "W",
		DataFile:     "AAPL-W.json",
		GoldenFile:   "daily-lines-aapl-w.json",
	})
}
