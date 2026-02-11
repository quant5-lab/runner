package parser

import "github.com/quant5-lab/runner/ast"

type WhileStatementConverter struct {
	orExprConverter    func(*OrExpr) (ast.Expression, error)
	statementConverter func(*Statement) (ast.Node, error)
}

func NewWhileStatementConverter(
	orExprConverter func(*OrExpr) (ast.Expression, error),
	statementConverter func(*Statement) (ast.Node, error),
) *WhileStatementConverter {
	return &WhileStatementConverter{
		orExprConverter:    orExprConverter,
		statementConverter: statementConverter,
	}
}

func (w *WhileStatementConverter) CanHandle(stmt *Statement) bool {
	return stmt.Core != nil && stmt.Core.While != nil
}

func (w *WhileStatementConverter) Convert(stmt *Statement) (ast.Node, error) {
	condition, err := w.orExprConverter(stmt.Core.While.Condition)
	if err != nil {
		return nil, err
	}

	body := []ast.Node{}
	for _, bodyStmt := range stmt.Core.While.Body {
		node, err := w.statementConverter(bodyStmt)
		if err != nil {
			return nil, err
		}
		if node != nil {
			body = append(body, node)
		}
	}

	return &ast.WhileStatement{
		NodeType:  ast.TypeWhileStatement,
		Condition: condition,
		Body:      body,
	}, nil
}
