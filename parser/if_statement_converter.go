package parser

import "github.com/quant5-lab/runner/ast"

type IfStatementConverter struct {
	orExprConverter    func(*OrExpr) (ast.Expression, error)
	statementConverter func(*Statement) (ast.Node, error)
}

func NewIfStatementConverter(
	orExprConverter func(*OrExpr) (ast.Expression, error),
	statementConverter func(*Statement) (ast.Node, error),
) *IfStatementConverter {
	return &IfStatementConverter{
		orExprConverter:    orExprConverter,
		statementConverter: statementConverter,
	}
}

func (i *IfStatementConverter) CanHandle(stmt *Statement) bool {
	return stmt.If != nil
}

func (i *IfStatementConverter) Convert(stmt *Statement) (ast.Node, error) {
	test, err := i.orExprConverter(stmt.If.Condition)
	if err != nil {
		return nil, err
	}

	consequent, err := i.convertBody(stmt.If.Body)
	if err != nil {
		return nil, err
	}

	return &ast.IfStatement{
		NodeType:   ast.TypeIfStatement,
		Test:       test,
		Consequent: consequent,
		Alternate:  []ast.Node{},
	}, nil
}

func (i *IfStatementConverter) convertBody(body []*Statement) ([]ast.Node, error) {
	nodes := []ast.Node{}
	for _, stmt := range body {
		node, err := i.statementConverter(stmt)
		if err != nil {
			return nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}
