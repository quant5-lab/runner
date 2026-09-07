package codegen

import "github.com/quant5-lab/runner/ast"

// namedArgResolver unpacks Pine all-named-arg calls into positional ast.Expression slices.
//
// Pine allows calling any function with all named arguments, e.g.:
//
//	valuewhen(condition=x, source=y, occurrence=0)
//
// The parser encodes this as Arguments: [ObjectExpression{condition:x, source:y, occurrence:0}].
// This resolver expands that back into [x, y, 0] using a declared parameter-name order.
type namedArgResolver struct {
	paramOrder []string
}

func newNamedArgResolver(paramOrder ...string) *namedArgResolver {
	return &namedArgResolver{paramOrder: paramOrder}
}

// resolve returns positional expressions from args.
// When args is already positional (len > 1, or len == 1 and not ObjectExpression),
// it is returned as-is. When args is [ObjectExpression], properties are extracted
// in paramOrder. Missing properties produce nil entries.
func (r *namedArgResolver) resolve(args []ast.Expression) []ast.Expression {
	if !r.isAllNamed(args) {
		return args
	}
	obj := args[0].(*ast.ObjectExpression)
	named := make(map[string]ast.Expression, len(obj.Properties))
	for _, prop := range obj.Properties {
		if key, ok := prop.Key.(*ast.Identifier); ok {
			named[key.Name] = prop.Value
		}
	}
	out := make([]ast.Expression, len(r.paramOrder))
	for i, name := range r.paramOrder {
		out[i] = named[name]
	}
	return out
}

func (r *namedArgResolver) isAllNamed(args []ast.Expression) bool {
	if len(args) != 1 {
		return false
	}
	_, ok := args[0].(*ast.ObjectExpression)
	return ok
}
