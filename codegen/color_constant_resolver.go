package codegen

import "github.com/quant5-lab/runner/ast"

type ColorConstantResolver struct {
	pineRegistry *PineConstantRegistry
}

func NewColorConstantResolver() *ColorConstantResolver {
	return &ColorConstantResolver{
		pineRegistry: NewPineConstantRegistry(),
	}
}

func (r *ColorConstantResolver) ResolveIdentifierToHex(name string) (string, bool) {
	return r.pineRegistry.GetColorHex(name)
}

func (r *ColorConstantResolver) ResolveMemberExpressionToHex(expr *ast.MemberExpression) (string, bool) {
	if expr == nil || expr.Object == nil || expr.Property == nil {
		return "", false
	}

	objIdent, objOk := expr.Object.(*ast.Identifier)
	if !objOk || objIdent.Name != "color" {
		return "", false
	}

	propIdent, propOk := expr.Property.(*ast.Identifier)
	if !propOk {
		return "", false
	}

	return r.pineRegistry.GetColorHex(propIdent.Name)
}

func (r *ColorConstantResolver) IsColorIdentifier(name string) bool {
	return r.pineRegistry.IsColorName(name)
}
