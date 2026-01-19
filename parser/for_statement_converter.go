package parser

import "github.com/quant5-lab/runner/ast"

type ForStatementConverter struct {
	arithExprConverter func(*ArithExpr) (ast.Expression, error)
	statementConverter func(*Statement) (ast.Node, error)
}

func NewForStatementConverter(
	arithExprConverter func(*ArithExpr) (ast.Expression, error),
	statementConverter func(*Statement) (ast.Node, error),
) *ForStatementConverter {
	return &ForStatementConverter{
		arithExprConverter: arithExprConverter,
		statementConverter: statementConverter,
	}
}

func (f *ForStatementConverter) CanHandle(stmt *Statement) bool {
	return stmt.For != nil
}

func (f *ForStatementConverter) Convert(stmt *Statement) (ast.Node, error) {
	forStmt := stmt.For

	fromExpr, err := f.arithExprConverter(forStmt.From)
	if err != nil {
		return nil, err
	}

	toExpr, err := f.arithExprConverter(forStmt.To)
	if err != nil {
		return nil, err
	}

	var stepExpr ast.Expression
	if forStmt.Step != nil {
		stepExpr, err = f.arithExprConverter(forStmt.Step)
		if err != nil {
			return nil, err
		}
	}

	body := []ast.Node{}
	for _, bodyStmt := range forStmt.Body {
		node, err := f.statementConverter(bodyStmt)
		if err != nil {
			return nil, err
		}
		if node != nil {
			body = append(body, node)
		}
	}

	return &ast.ForStatement{
		NodeType: ast.TypeForStatement,
		Counter:  forStmt.Counter,
		From:     fromExpr,
		To:       toExpr,
		Step:     stepExpr,
		Body:     body,
	}, nil
}
