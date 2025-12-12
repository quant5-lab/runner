package security

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// extractCallFunctionName retrieves function name from CallExpression callee
func extractCallFunctionName(callee ast.Expression) string {
	if mem, ok := callee.(*ast.MemberExpression); ok {
		obj := ""
		if id, ok := mem.Object.(*ast.Identifier); ok {
			obj = id.Name
		}
		prop := ""
		if id, ok := mem.Property.(*ast.Identifier); ok {
			prop = id.Name
		}
		return obj + "." + prop
	}

	if id, ok := callee.(*ast.Identifier); ok {
		return id.Name
	}

	return ""
}

// extractNumberLiteral converts AST Literal to float64
func extractNumberLiteral(expr ast.Expression) (float64, error) {
	lit, ok := expr.(*ast.Literal)
	if !ok {
		return 0, fmt.Errorf("expected literal, got %T", expr)
	}

	switch v := lit.Value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("expected number literal, got %T", v)
	}
}
