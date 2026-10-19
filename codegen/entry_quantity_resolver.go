package codegen

import "github.com/quant5-lab/runner/ast"

// QtyEvaluator resolves an AST expression to a compile-time fixed quantity.
type QtyEvaluator func(expr ast.Expression) (qty float64, resolved bool)

// EntryQuantityResolver determines the fixed quantity for strategy.entry/order calls.
type EntryQuantityResolver struct{}

func NewEntryQuantityResolver() *EntryQuantityResolver {
	return &EntryQuantityResolver{}
}

func (r *EntryQuantityResolver) ResolveQuantity(
	args []ast.Expression,
	defaultQty float64,
	evaluate QtyEvaluator,
) float64 {
	if qty, ok := r.resolveFromNamedArgs(args, evaluate); ok {
		return qty
	}
	if qty, ok := r.resolveFromPositionalArg(args, evaluate); ok {
		return qty
	}
	return defaultQty
}

// Pine packs all named arguments into one ObjectExpression appended after positional args,
// so scanning all args handles every call shape without assuming a fixed index.
func (r *EntryQuantityResolver) resolveFromNamedArgs(args []ast.Expression, evaluate QtyEvaluator) (float64, bool) {
	for _, arg := range args {
		obj, ok := arg.(*ast.ObjectExpression)
		if !ok {
			continue
		}
		for _, prop := range obj.Properties {
			key, ok := prop.Key.(*ast.Identifier)
			if !ok || key.Name != "qty" {
				continue
			}
			return evaluate(prop.Value)
		}
	}
	return 0, false
}

// ObjectExpression at args[2] means no positional qty was provided (it is the named-arg blob).
func (r *EntryQuantityResolver) resolveFromPositionalArg(args []ast.Expression, evaluate QtyEvaluator) (float64, bool) {
	if len(args) < 3 {
		return 0, false
	}
	if _, isNamedBlob := args[2].(*ast.ObjectExpression); isNamedBlob {
		return 0, false
	}
	return evaluate(args[2])
}
