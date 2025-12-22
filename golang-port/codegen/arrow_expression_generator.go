package codegen

import "github.com/quant5-lab/runner/ast"

/*
ArrowExpressionGenerator defines contract for generating expressions in arrow function context.

All expressions in arrow functions must resolve variables to Series access patterns:
  - Local variable 'up' → upSeries.GetCurrent()
  - Parameter 'len' → len (scalar parameter)
  - Builtin 'close' → ctx.Data[ctx.BarIndex].Close
*/
type ArrowExpressionGenerator interface {
	Generate(expr ast.Expression) (string, error)
}
