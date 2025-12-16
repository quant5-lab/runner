package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

/* InlineConditionHandler generates inline expressions for use within conditions, plots, and ternary expressions.
 * Unlike TAHandler (which generates variable storage statements), this returns pure expressions like:
 * - "(func() float64 { ... }())" for ta.dev
 * - "(func() bool { ... }())" for ta.crossover
 * - "math.IsNaN(x)" for na(x)
 */
type InlineConditionHandler interface {
	/* CanHandle returns true if this handler supports the given function name */
	CanHandle(funcName string) bool

	/* GenerateInline generates an inline expression (not a statement) that can be embedded in conditions/ternaries */
	GenerateInline(expr *ast.CallExpression, g *generator) (string, error)
}
