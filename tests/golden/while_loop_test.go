package golden

import (
	"testing"
)

func TestWhileLoopFactorial_AAPL_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "While Loop Factorial",
		StrategyFile: "test-while-loop-factorial.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "while_loop_factorial_aapl_1h.golden.json",
	})
}

func TestWhileLoopFactorial_BTCUSDT_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "While Loop Factorial",
		StrategyFile: "test-while-loop-factorial.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "while_loop_factorial_btcusdt_1h.golden.json",
	})
}

func TestWhileLoopSum_AAPL_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "While Loop Sum",
		StrategyFile: "test-while-loop-sum.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "while_loop_sum_aapl_1h.golden.json",
	})
}

func TestWhileLoopSum_BTCUSDT_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "While Loop Sum",
		StrategyFile: "test-while-loop-sum.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "while_loop_sum_btcusdt_1h.golden.json",
	})
}

/* While Break Strategy Tests */

func TestWhileBreak_AAPL_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "While Break",
		StrategyFile: "test-while-break.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "while_break_aapl_1h.golden.json",
	})
}

func TestWhileBreak_BTCUSDT_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "While Break",
		StrategyFile: "test-while-break.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "while_break_btcusdt_1h.golden.json",
	})
}

/* While Continue Strategy Tests */

func TestWhileContinue_AAPL_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "While Continue",
		StrategyFile: "test-while-continue.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "while_continue_aapl_1h.golden.json",
	})
}

func TestWhileContinue_BTCUSDT_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "While Continue",
		StrategyFile: "test-while-continue.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "while_continue_btcusdt_1h.golden.json",
	})
}

/* While Expression Strategy Tests */

func TestWhileExpression_AAPL_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "While Expression",
		StrategyFile: "test-while-expression.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "while_expression_aapl_1h.golden.json",
	})
}

func TestWhileExpression_BTCUSDT_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "While Expression",
		StrategyFile: "test-while-expression.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "while_expression_btcusdt_1h.golden.json",
	})
}
