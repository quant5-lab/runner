package codegen

import "github.com/quant5-lab/runner/ast"

/*
ArrowExpressionGenerator defines contract for generating expressions in arrow function context.

All expressions in arrow functions resolve identifiers using dual-access pattern:
  - Local variable 'up' → up (scalar for current bar)
  - Parameter 'len' → len (scalar parameter)
  - Builtin 'close' → bar.Close
  - Historical access via Series.Get(offset) in TA loops
*/
type ArrowExpressionGenerator interface {
	Generate(expr ast.Expression) (string, error)
}
