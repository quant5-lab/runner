package parser

import (
	"github.com/quant5-lab/runner/ast"
)

/* Shared test helper functions */

func findVariableDeclaration(program *ast.Program, name string) *ast.VariableDeclaration {
	for _, stmt := range program.Body {
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			if len(varDecl.Declarations) > 0 {
				if id, ok := varDecl.Declarations[0].ID.(*ast.Identifier); ok {
					if id.Name == name {
						return varDecl
					}
				}
			}
		}
	}
	return nil
}

func binaryExpressionToString(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.BinaryExpression:
		left := binaryExpressionToString(e.Left)
		right := binaryExpressionToString(e.Right)
		return "(" + left + " " + e.Operator + " " + right + ")"
	case *ast.Literal:
		return formatLiteral(e.Value)
	case *ast.Identifier:
		return e.Name
	default:
		return "?"
	}
}

func formatLiteral(v interface{}) string {
	if f, ok := v.(float64); ok {
		if f == float64(int(f)) && f >= 0 && f < 1000000 {
			return formatInt(int(f))
		}
		return formatFloat(f)
	}
	return "?"
}

func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + formatInt(-n)
	}
	result := ""
	for n > 0 {
		result = string(rune(n%10+'0')) + result
		n /= 10
	}
	return result
}

func formatFloat(f float64) string {
	switch f {
	case 0.5:
		return "0.5"
	case 2.5:
		return "2.5"
	default:
		return "?"
	}
}
