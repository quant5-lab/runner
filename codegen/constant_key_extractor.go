package codegen

import "github.com/quant5-lab/runner/ast"

type ConstantKeyExtractor struct {
	pineRegistry *PineConstantRegistry
}

func NewConstantKeyExtractor() *ConstantKeyExtractor {
	return &ConstantKeyExtractor{
		pineRegistry: NewPineConstantRegistry(),
	}
}

func (cke *ConstantKeyExtractor) ExtractFromExpression(expr ast.Expression) (string, bool) {
	if memExpr, ok := expr.(*ast.MemberExpression); ok {
		return cke.extractFromMemberExpression(memExpr)
	}

	if ident, ok := expr.(*ast.Identifier); ok {
		if cke.pineRegistry.IsColorName(ident.Name) {
			return "color." + ident.Name, true
		}
	}

	return "", false
}

func (cke *ConstantKeyExtractor) extractFromMemberExpression(memExpr *ast.MemberExpression) (string, bool) {
	obj, objOk := memExpr.Object.(*ast.Identifier)
	if !objOk {
		return "", false
	}

	prop, propOk := memExpr.Property.(*ast.Identifier)
	if !propOk {
		return "", false
	}

	return obj.Name + "." + prop.Name, true
}
