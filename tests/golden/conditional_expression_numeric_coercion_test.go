package golden

import (
	"testing"
)

func TestConditionalNumericCoercion_MainContext_AAPL_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Conditional Numeric Coercion - Main Context",
		StrategyFile: "test-conditional-coercion-main.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "conditional_coercion_main_aapl_1h.golden.json",
	})
}

func TestConditionalNumericCoercion_MainContext_BTCUSDT_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Conditional Numeric Coercion - Main Context",
		StrategyFile: "test-conditional-coercion-main.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "conditional_coercion_main_btcusdt_1h.golden.json",
	})
}

func TestConditionalNumericCoercion_ArrowFunction_AAPL_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Conditional Numeric Coercion - Arrow Function",
		StrategyFile: "test-conditional-coercion-arrow.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "conditional_coercion_arrow_aapl_1h.golden.json",
	})
}

func TestConditionalNumericCoercion_ArrowFunction_BTCUSDT_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Conditional Numeric Coercion - Arrow Function",
		StrategyFile: "test-conditional-coercion-arrow.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "conditional_coercion_arrow_btcusdt_1h.golden.json",
	})
}

func TestConditionalNumericCoercion_ForLoop_AAPL_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Conditional Numeric Coercion - For Loop",
		StrategyFile: "test-conditional-coercion-forloop.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "conditional_coercion_forloop_aapl_1h.golden.json",
	})
}

func TestConditionalNumericCoercion_ForLoop_BTCUSDT_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Conditional Numeric Coercion - For Loop",
		StrategyFile: "test-conditional-coercion-forloop.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "conditional_coercion_forloop_btcusdt_1h.golden.json",
	})
}

func TestConditionalNumericCoercion_MixedTypes_AAPL_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Conditional Numeric Coercion - Mixed Types",
		StrategyFile: "test-conditional-coercion-mixed.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "conditional_coercion_mixed_aapl_1h.golden.json",
	})
}

func TestConditionalNumericCoercion_MixedTypes_BTCUSDT_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Conditional Numeric Coercion - Mixed Types",
		StrategyFile: "test-conditional-coercion-mixed.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "conditional_coercion_mixed_btcusdt_1h.golden.json",
	})
}
