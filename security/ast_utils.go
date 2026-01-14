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

// extractNumberLiteral converts AST expression to float64
// Supports input constants via optional inputConstantsMap parameter
func extractNumberLiteral(expr ast.Expression, inputConstantsMap ...map[string]float64) (float64, error) {
	if id, ok := expr.(*ast.Identifier); ok {
		/* Check input constants map first if provided */
		if len(inputConstantsMap) > 0 && inputConstantsMap[0] != nil {
			if val, ok := inputConstantsMap[0][id.Name]; ok {
				return val, nil
			}
		}

		/* Fallback to hardcoded defaults */
		switch id.Name {
		case "leftBars", "rightBars":
			return 15, nil
		default:
			return 0, fmt.Errorf("cannot resolve identifier '%s' to number", id.Name)
		}
	}

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
