package codegen

import "github.com/quant5-lab/runner/ast"

func Ident(name string) *ast.Identifier {
	return &ast.Identifier{Name: name}
}

func Lit(value interface{}) *ast.Literal {
	return &ast.Literal{Value: value}
}

func BinaryExpr(op string, left, right ast.Expression) *ast.BinaryExpression {
	return &ast.BinaryExpression{
		Operator: op,
		Left:     left,
		Right:    right,
	}
}

func TACall(method string, source ast.Expression, period float64) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   Ident("ta"),
			Property: Ident(method),
		},
		Arguments: []ast.Expression{
			source,
			Lit(period),
		},
	}
}

func TACallPeriodOnly(method string, period float64) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   Ident("ta"),
			Property: Ident(method),
		},
		Arguments: []ast.Expression{
			Lit(period),
		},
	}
}

func MathCall(method string, args ...ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   Ident("math"),
			Property: Ident(method),
		},
		Arguments: args,
	}
}

func MemberExpr(object, property string) *ast.MemberExpression {
	return &ast.MemberExpression{
		Object:   Ident(object),
		Property: Ident(property),
	}
}

func CallExpr(callee ast.Expression, args ...ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee:    callee,
		Arguments: args,
	}
}
