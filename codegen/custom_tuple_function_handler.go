package codegen

import "github.com/quant5-lab/runner/ast"

// CustomTupleFunctionHandler handles tuple-returning TA functions requiring stateful
// ForwardSeriesBuffer computation, as opposed to the windowed-array runtime delegation
// used by TupleIndicatorRegistry (ta.macd, ta.bb, ta.stoch, ta.dmi).
//
// Implement this interface for any tuple indicator whose sub-computations are
// inherently stateful (EMA, RMA, carry-forward bands) and cannot be expressed
// as a stateless function of a source window.
type CustomTupleFunctionHandler interface {
	CanHandle(funcName string) bool

	// varNames are in PineScript return order.
	GenerateTupleCode(g *generator, varNames []string, call *ast.CallExpression) (string, error)

	// firstOutputVar is the naming prefix (matches CompositeIndicatorMetadata convention, e.g. "_upper_ema").
	InternalSeriesNames(firstOutputVar string, call *ast.CallExpression) []string
}
