package codegen

// TACalculationCore defines the interface for extracting pure TA calculation logic
// independent of how the result is wrapped (Series.Set() vs IIFE return).
//
// This separates WHAT to calculate from HOW to wrap it, enabling code reuse
// between Series-based indicators (TAIndicatorBuilder) and expression-based
// inline calculations (InlineTAIIFERegistry).
type TACalculationCore interface {
	// GenerateCalculationBody generates the pure calculation logic without wrapper.
	//
	// Parameters:
	//   - accessor: AccessGenerator for retrieving data values
	//   - period: Lookback period for the indicator
	//   - indenter: Optional indenter for multi-line code (nil for single-line IIFE)
	//
	// Returns calculation code and result expression (e.g., "sum / 20.0", "math.Sqrt(variance / 20.0)")
	GenerateCalculationBody(accessor AccessGenerator, period int, indenter *CodeIndenter) (body string, resultExpr string)

	// GetWarmupPeriod returns the minimum number of bars needed before calculation is valid.
	GetWarmupPeriod(period int) int

	NeedsNaNGuard() bool
}
