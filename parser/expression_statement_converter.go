package parser

import "github.com/quant5-lab/runner/ast"

type ExpressionStatementConverter struct {
	expressionConverter func(*Expression) (ast.Expression, error)
}

func NewExpressionStatementConverter(expressionConverter func(*Expression) (ast.Expression, error)) *ExpressionStatementConverter {
	return &ExpressionStatementConverter{
		expressionConverter: expressionConverter,
	}
}

func (e *ExpressionStatementConverter) CanHandle(stmt *Statement) bool {
	return stmt.Expression != nil
}

func (e *ExpressionStatementConverter) Convert(stmt *Statement) (ast.Node, error) {
	expr, err := e.expressionConverter(stmt.Expression.Expr)
	if err != nil {
		return nil, err
	}

	return &ast.ExpressionStatement{
		NodeType:   ast.TypeExpressionStatement,
		Expression: expr,
	}, nil
}
