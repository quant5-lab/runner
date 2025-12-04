package codegen

import "fmt"

// TAIndicatorFactory creates appropriate components for technical analysis indicators.
//
// This factory encapsulates the logic of selecting the right accumulator strategy
// and configuration for each indicator type, following the Factory pattern.
//
// Usage:
//
//	factory := NewTAIndicatorFactory()
//	builder, err := factory.CreateBuilder("ta.sma", "sma20", 20, accessor)
//	if err != nil {
//	    return "", err
//	}
//	code := builder.Build()
//
// Design:
//   - Factory Pattern: Creates appropriate accumulator for each indicator type
//   - Strategy Pattern: Returns configured builder with correct strategy
//   - Open/Closed: Add new indicators by adding cases, no changes to builder
type TAIndicatorFactory struct{}

// NewTAIndicatorFactory creates a new factory for TA indicators.
func NewTAIndicatorFactory() *TAIndicatorFactory {
	return &TAIndicatorFactory{}
}

// CreateBuilder creates a fully configured TAIndicatorBuilder for the specified indicator type.
//
// Parameters:
//   - indicatorType: The indicator type (e.g., "ta.sma", "ta.ema", "ta.stdev")
//   - varName: Variable name for the output Series
//   - period: Lookback period
//   - accessor: AccessGenerator for data source
//
// Returns a configured builder ready to generate code, or an error if the indicator type is not supported.
func (f *TAIndicatorFactory) CreateBuilder(
	indicatorType string,
	varName string,
	period int,
	accessor AccessGenerator,
) (*TAIndicatorBuilder, error) {
	// Determine if NaN checking is needed based on source type
	needsNaN := f.shouldCheckNaN(accessor)

	// Create base builder
	builder := NewTAIndicatorBuilder(indicatorType, varName, period, accessor, needsNaN)

	// Configure accumulator based on indicator type
	switch indicatorType {
	case "ta.sma":
		builder.WithAccumulator(NewSumAccumulator())
		return builder, nil

	case "ta.ema":
		builder.WithAccumulator(NewEMAAccumulator(period))
		return builder, nil

	case "ta.wma":
		builder.WithAccumulator(NewWeightedSumAccumulator(period))
		return builder, nil

	case "ta.dev":
		// DEV requires special handling like STDEV - return builder without accumulator
		// Caller must handle two-pass calculation (mean then absolute deviation)
		return builder, nil

	case "ta.stdev":
		// STDEV requires special handling - return builder without accumulator
		// Caller must handle two-pass calculation (mean then variance)
		return builder, nil

	default:
		return nil, fmt.Errorf("unsupported indicator type: %s", indicatorType)
	}
}

// CreateSTDEVBuilders creates the two builders needed for STDEV calculation.
//
// STDEV requires two passes:
//  1. Calculate mean (using SumAccumulator)
//  2. Calculate variance from mean (using VarianceAccumulator)
//
// Returns:
//   - meanBuilder: Builder for mean calculation
//   - varianceBuilder: Builder for variance calculation
//   - error: If creation fails
func (f *TAIndicatorFactory) CreateSTDEVBuilders(
	varName string,
	period int,
	accessor AccessGenerator,
) (meanBuilder *TAIndicatorBuilder, varianceBuilder *TAIndicatorBuilder, err error) {
	needsNaN := f.shouldCheckNaN(accessor)

	// Pass 1: Calculate mean
	meanBuilder = NewTAIndicatorBuilder("STDEV_MEAN", varName, period, accessor, needsNaN)
	meanBuilder.WithAccumulator(NewSumAccumulator())

	// Pass 2: Calculate variance (uses mean from pass 1)
	varianceBuilder = NewTAIndicatorBuilder("STDEV", varName, period, accessor, false)
	varianceBuilder.WithAccumulator(NewVarianceAccumulator("mean"))

	return meanBuilder, varianceBuilder, nil
}

// shouldCheckNaN determines if NaN checking is needed based on accessor type.
//
// Series variables need NaN checking because they can contain calculated values
// that might be NaN. OHLCV fields from raw data typically don't need NaN checks.
func (f *TAIndicatorFactory) shouldCheckNaN(accessor AccessGenerator) bool {
	// Check if accessor is a Series variable accessor
	switch accessor.(type) {
	case *SeriesVariableAccessGenerator:
		return true
	case *OHLCVFieldAccessGenerator:
		return false
	default:
		// Conservative default: check for NaN
		return true
	}
}

// SupportedIndicators returns a list of all supported indicator types.
func (f *TAIndicatorFactory) SupportedIndicators() []string {
	return []string{
		"ta.sma",
		"ta.ema",
		"ta.wma",
		"ta.dev",
		"ta.stdev",
	}
}

// IsSupported checks if an indicator type is supported.
func (f *TAIndicatorFactory) IsSupported(indicatorType string) bool {
	for _, supported := range f.SupportedIndicators() {
		if supported == indicatorType {
			return true
		}
	}
	return false
}
