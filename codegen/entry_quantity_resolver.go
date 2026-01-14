package codegen

import "github.com/quant5-lab/runner/ast"

// EntryQuantityResolver determines quantity for strategy.entry() calls.
type EntryQuantityResolver struct{}

// NewEntryQuantityResolver creates a resolver.
func NewEntryQuantityResolver() *EntryQuantityResolver {
	return &EntryQuantityResolver{}
}

// ResolveQuantity determines entry quantity from call arguments and config.
func (r *EntryQuantityResolver) ResolveQuantity(
	args []ast.Expression,
	defaultQty float64,
	extractLiteral func(ast.Expression) float64,
) float64 {
	if len(args) < 3 {
		return defaultQty
	}

	thirdArg := args[2]

	if r.isNamedParameter(thirdArg) {
		return defaultQty
	}

	explicitQty := extractLiteral(thirdArg)
	if explicitQty > 0 {
		return explicitQty
	}

	return defaultQty
}

func (r *EntryQuantityResolver) isNamedParameter(expr ast.Expression) bool {
	_, ok := expr.(*ast.ObjectExpression)
	return ok
}
