package codegen

import "github.com/quant5-lab/runner/ast"

func strategyCallNode() *ast.ExpressionStatement {
	return &ast.ExpressionStatement{
		Expression: &ast.CallExpression{
			Callee: &ast.Identifier{Name: "strategy"},
			Arguments: []ast.Expression{
				&ast.Literal{Value: "Test"},
			},
		},
	}
}

func securityVariableNode(varName string, expression ast.Expression) *ast.VariableDeclaration {
	return &ast.VariableDeclaration{
		Declarations: []ast.VariableDeclarator{
			{
				ID: ast.Identifier{Name: varName},
				Init: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "security"},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "BTCUSD"},
						&ast.Literal{Value: "1D"},
						expression,
					},
				},
			},
		},
	}
}

func plotCallNode(expression ast.Expression) *ast.ExpressionStatement {
	return &ast.ExpressionStatement{
		Expression: &ast.CallExpression{
			Callee: &ast.Identifier{Name: "plot"},
			Arguments: []ast.Expression{
				expression,
			},
		},
	}
}

func buildSecurityTestProgram(varName string, expression ast.Expression) *ast.Program {
	return &ast.Program{
		Body: []ast.Node{
			strategyCallNode(),
			securityVariableNode(varName, expression),
		},
	}
}

func buildPlotTestProgram(expression ast.Expression) *ast.Program {
	return &ast.Program{
		Body: []ast.Node{
			plotCallNode(expression),
		},
	}
}

func buildMultiSecurityTestProgram(vars map[string]ast.Expression) *ast.Program {
	body := []ast.Node{strategyCallNode()}
	for varName, expr := range vars {
		body = append(body, securityVariableNode(varName, expr))
	}
	return &ast.Program{Body: body}
}
