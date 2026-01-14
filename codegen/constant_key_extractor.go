package codegen

import "github.com/quant5-lab/runner/ast"

type ConstantKeyExtractor struct{}

func NewConstantKeyExtractor() *ConstantKeyExtractor {
	return &ConstantKeyExtractor{}
}

func (cke *ConstantKeyExtractor) ExtractFromExpression(expr ast.Expression) (string, bool) {
	if memExpr, ok := expr.(*ast.MemberExpression); ok {
		return cke.extractFromMemberExpression(memExpr)
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
