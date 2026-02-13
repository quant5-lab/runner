package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

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

func (r *ColorConstantResolver) ResolveExpression(expr ast.Expression) (string, bool) {
	switch e := expr.(type) {
	case *ast.MemberExpression:
		return r.ResolveMemberExpressionToHex(e)
	case *ast.CallExpression:
		return r.resolveCallExpression(e)
	case *ast.Literal:
		return r.resolveLiteralHex(e)
	default:
		return "", false
	}
}

func (r *ColorConstantResolver) resolveCallExpression(call *ast.CallExpression) (string, bool) {
	member, ok := call.Callee.(*ast.MemberExpression)
	if !ok {
		return "", false
	}
	obj, objOk := member.Object.(*ast.Identifier)
	prop, propOk := member.Property.(*ast.Identifier)
	if !objOk || !propOk || obj.Name != "color" {
		return "", false
	}

	switch prop.Name {
	case "rgb":
		return r.resolveColorRGB(call.Arguments)
	case "new":
		return r.resolveColorNew(call.Arguments)
	default:
		return "", false
	}
}

func (r *ColorConstantResolver) resolveColorRGB(args []ast.Expression) (string, bool) {
	if len(args) < 3 {
		return "", false
	}
	red, ok := extractNumericLiteral(args[0])
	if !ok {
		return "", false
	}
	green, ok := extractNumericLiteral(args[1])
	if !ok {
		return "", false
	}
	blue, ok := extractNumericLiteral(args[2])
	if !ok {
		return "", false
	}
	return fmt.Sprintf("#%02X%02X%02X", clampByte(red), clampByte(green), clampByte(blue)), true
}

func (r *ColorConstantResolver) resolveColorNew(args []ast.Expression) (string, bool) {
	if len(args) < 1 {
		return "", false
	}
	return r.ResolveExpression(args[0])
}

/* Pine hex color literals: #RRGGBB (6-digit) or #RRGGBBAA (8-digit RGBA) */
func (r *ColorConstantResolver) resolveLiteralHex(lit *ast.Literal) (string, bool) {
	str, ok := lit.Value.(string)
	if !ok || (len(str) != 7 && len(str) != 9) || str[0] != '#' {
		return "", false
	}
	for _, c := range str[1:] {
		if !isHexDigit(c) {
			return "", false
		}
	}
	return str, true
}

func isHexDigit(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func extractNumericLiteral(expr ast.Expression) (int, bool) {
	lit, ok := expr.(*ast.Literal)
	if !ok {
		return 0, false
	}
	switch v := lit.Value.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	default:
		return 0, false
	}
}

func clampByte(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}
