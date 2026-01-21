package golden

import (
	"testing"
)

/* Simple Counter Strategy Tests */

func TestForLoopSimpleCounter_AAPL_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Simple For Loop Counter",
		StrategyFile: "test-for-loop-simple-counter.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "for_loop_simple_counter_aapl_1h.golden.json",
	})
}

func TestForLoopSimpleCounter_BTCUSDT_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Simple For Loop Counter",
		StrategyFile: "test-for-loop-simple-counter.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "for_loop_simple_counter_btcusdt_1h.golden.json",
	})
}

/* Moving Average Strategy Tests */

func TestForLoopMovingAverage_AAPL_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "For Loop Moving Average",
		StrategyFile: "test-for-loop-moving-average.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "for_loop_moving_average_aapl_1h.golden.json",
	})
}

func TestForLoopMovingAverage_BTCUSDT_1D(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "For Loop Moving Average",
		StrategyFile: "test-for-loop-moving-average.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "D",
		DataFile:     "BTCUSDT_1D.json",
		GoldenFile:   "for_loop_moving_average_btcusdt_1d.golden.json",
	})
}

/* Conditional Accumulation Strategy Tests */

func TestForLoopConditionalAccumulation_AAPL_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Conditional Accumulation",
		StrategyFile: "test-for-loop-conditional-accumulation.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "for_loop_conditional_accumulation_aapl_1h.golden.json",
	})
}

func TestForLoopConditionalAccumulation_BTCUSDT_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Conditional Accumulation",
		StrategyFile: "test-for-loop-conditional-accumulation.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "for_loop_conditional_accumulation_btcusdt_1h.golden.json",
	})
}

/* Correlation Strategy Tests */

func TestForLoopCorrelation_AAPL_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Price Correlation",
		StrategyFile: "test-for-loop-correlation.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "for_loop_correlation_aapl_1h.golden.json",
	})
}

func TestForLoopCorrelation_BTCUSDT_1D(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Price Correlation",
		StrategyFile: "test-for-loop-correlation.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "D",
		DataFile:     "BTCUSDT_1D.json",
		GoldenFile:   "for_loop_correlation_btcusdt_1d.golden.json",
	})
}

/* Volatility Bands Strategy Tests */

func TestForLoopVolatilityBands_AAPL_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Custom Volatility Bands",
		StrategyFile: "test-for-loop-volatility-bands.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "for_loop_volatility_bands_aapl_1h.golden.json",
	})
}

func TestForLoopVolatilityBands_BTCUSDT_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Custom Volatility Bands",
		StrategyFile: "test-for-loop-volatility-bands.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "for_loop_volatility_bands_btcusdt_1h.golden.json",
	})
}

/* Weighted Average Strategy Tests */

func TestForLoopWeightedAverage_AAPL_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Weighted Average",
		StrategyFile: "test-for-loop-weighted-average.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "for_loop_weighted_average_aapl_1h.golden.json",
	})
}

func TestForLoopWeightedAverage_BTCUSDT_1D(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "Weighted Average",
		StrategyFile: "test-for-loop-weighted-average.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "D",
		DataFile:     "BTCUSDT_1D.json",
		GoldenFile:   "for_loop_weighted_average_btcusdt_1d.golden.json",
	})
}

/* Step Variations Strategy Tests */

func TestForLoopStepVariations_AAPL_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "For Loop Step Variations",
		StrategyFile: "test-for-loop-step-variations.pine",
		Symbol:       "AAPL",
		Timeframe:    "1h",
		DataFile:     "AAPL-1h.json",
		GoldenFile:   "for_loop_step_variations_aapl_1h.golden.json",
	})
}

func TestForLoopStepVariations_BTCUSDT_1h(t *testing.T) {
	suite := NewTestSuite(t)

	suite.RunTestFixtureAndValidate(t, TestConfig{
		StrategyName: "For Loop Step Variations",
		StrategyFile: "test-for-loop-step-variations.pine",
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		DataFile:     "BTCUSDT-1h.json",
		GoldenFile:   "for_loop_step_variations_btcusdt_1h.golden.json",
	})
}
