package codegen

// TACalculationCore defines the interface for extracting pure TA calculation logic
// independent of how the result is wrapped (Series.Set() vs IIFE return).
//
// This separates WHAT to calculate from HOW to wrap it, enabling code reuse
// between Series-based indicators (TAIndicatorBuilder) and expression-based
// inline calculations (InlineTAIIFERegistry).
type TACalculationCore interface {
	// GenerateCalculationBody generates the pure calculation logic without wrapper.
	// Returns the calculation code that produces a result variable or expression.
	//
	// Parameters:
	//   - accessor: AccessGenerator for retrieving data values
	//   - period: Lookback period for the indicator
	//   - indenter: Optional indenter for multi-line code (nil for single-line IIFE)
	//
	// Returns:
	//   - Calculation code without Series.Set() or return statement
	//   - Result expression (e.g., "sum / 20.0", "ema", "math.Sqrt(variance / 20.0)")
	GenerateCalculationBody(accessor AccessGenerator, period int, indenter *CodeIndenter) (body string, resultExpr string)

	// GetWarmupPeriod returns the minimum number of bars needed before calculation is valid.
	// Defaults to period-1 for most indicators.
	GetWarmupPeriod(period int) int

	// NeedsNaNGuard returns true if accumulation loop should check for NaN values.
	NeedsNaNGuard() bool
}
