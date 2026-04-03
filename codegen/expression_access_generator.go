package codegen

// ExpressionAccessGenerator rewrites series expressions to support historical offsets.
// It leverages generator helpers to substitute current-bar access with lookback-aware access.
type ExpressionAccessGenerator struct {
	gen      *generator
	exprCode string
}

// NewExpressionAccessGenerator creates accessor for arbitrary expressions.
func NewExpressionAccessGenerator(gen *generator, exprCode string) *ExpressionAccessGenerator {
	return &ExpressionAccessGenerator{gen: gen, exprCode: exprCode}
}

// GenerateLoopValueAccess applies a dynamic offset (loop var) to all series accesses within the expression.
func (a *ExpressionAccessGenerator) GenerateLoopValueAccess(loopVar string) string {
	return a.gen.convertSeriesAccessToOffset(a.exprCode, loopVar)
}

// GenerateInitialValueAccess applies a fixed offset for initial seeding (period-1 lookback).
func (a *ExpressionAccessGenerator) GenerateInitialValueAccess(period int) string {
	return a.gen.convertSeriesAccessToIntOffset(a.exprCode, period-1)
}

// GenerateCurrentValueAccess returns the expression for current bar access (no transformation needed).
func (a *ExpressionAccessGenerator) GenerateCurrentValueAccess() string {
	return a.exprCode
}
