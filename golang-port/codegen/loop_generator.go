package codegen

import "fmt"

// AccessGenerator provides methods to generate code for accessing series values.
//
// This interface abstracts the difference between accessing:
//   - User-defined Series variables: sma20Series.Get(offset)
//   - OHLCV built-in fields: ctx.Data[ctx.BarIndex-offset].Close
//
// Implementations:
//   - SeriesVariableAccessGenerator: For Series variables
//   - OHLCVFieldAccessGenerator: For OHLCV fields (open, high, low, close, volume)
//
// Use CreateAccessGenerator() to automatically create the appropriate implementation.
type AccessGenerator interface {
	// GenerateLoopValueAccess generates code to access a value within a loop
	// Parameter: loopVar is the loop counter variable name (e.g., "j")
	GenerateLoopValueAccess(loopVar string) string

	// GenerateInitialValueAccess generates code to access the initial value
	// Parameter: period is the lookback period
	GenerateInitialValueAccess(period int) string
}

// LoopGenerator creates for-loop structures for iterating over lookback periods.
//
// This component handles:
//   - Forward iteration (0 to period-1) for accumulation
//   - Backward iteration (period-1 to 0) for reverse processing
//   - Integration with AccessGenerator for data retrieval
//   - Optional NaN checking for data validation
//
// Usage:
//
//	accessor := CreateAccessGenerator("close")
//	loopGen := NewLoopGenerator(20, accessor, true)
//
//	indenter := NewCodeIndenter()
//	code := loopGen.GenerateForwardLoop(&indenter)
//	// Output: for j := 0; j < 20; j++ {
//
//	valueAccess := loopGen.GenerateValueAccess()
//	// Output: ctx.Data[ctx.BarIndex-j].Close
type LoopGenerator struct {
	period   int             // Lookback period
	loopVar  string          // Loop counter variable name (default: "j")
	accessor AccessGenerator // Data access strategy
	needsNaN bool            // Whether to add NaN checking
}

func NewLoopGenerator(period int, accessor AccessGenerator, needsNaN bool) *LoopGenerator {
	return &LoopGenerator{
		period:   period,
		loopVar:  "j",
		accessor: accessor,
		needsNaN: needsNaN,
	}
}

func (l *LoopGenerator) GenerateForwardLoop(indenter *CodeIndenter) string {
	return indenter.Line(fmt.Sprintf("for %s := 0; %s < %d; %s++ {",
		l.loopVar, l.loopVar, l.period, l.loopVar))
}

func (l *LoopGenerator) GenerateBackwardLoop(indenter *CodeIndenter) string {
	return indenter.Line(fmt.Sprintf("for %s := %d-2; %s >= 0; %s-- {",
		l.loopVar, l.period, l.loopVar, l.loopVar))
}

func (l *LoopGenerator) GenerateValueAccess() string {
	return l.accessor.GenerateLoopValueAccess(l.loopVar)
}

func (l *LoopGenerator) RequiresNaNCheck() bool {
	return l.needsNaN
}
