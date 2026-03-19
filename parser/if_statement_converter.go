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
	return stmt.Core != nil && stmt.Core.If != nil
}

func (i *IfStatementConverter) Convert(stmt *Statement) (ast.Node, error) {
	test, err := i.orExprConverter(stmt.Core.If.Condition)
	if err != nil {
		return nil, err
	}

	consequent, err := i.convertBody(stmt.Core.If.Body)
	if err != nil {
		return nil, err
	}

	alternate, err := i.convertElseClause(stmt.Core.If.ElseClause)
	if err != nil {
		return nil, err
	}

	return &ast.IfStatement{
		NodeType:   ast.TypeIfStatement,
		Test:       test,
		Consequent: consequent,
		Alternate:  alternate,
	}, nil
}

func (i *IfStatementConverter) convertElseClause(ec *ElseClause) ([]ast.Node, error) {
	if ec == nil {
		return []ast.Node{}, nil
	}
	if ec.ElseIf != nil {
		node, err := i.convertIfGrammarNode(ec.ElseIf)
		if err != nil {
			return nil, err
		}
		return []ast.Node{node}, nil
	}
	return i.convertBody(ec.ElseBody)
}

func (i *IfStatementConverter) convertIfGrammarNode(ifGram *IfStatement) (ast.Node, error) {
	test, err := i.orExprConverter(ifGram.Condition)
	if err != nil {
		return nil, err
	}
	consequent, err := i.convertBody(ifGram.Body)
	if err != nil {
		return nil, err
	}
	alternate, err := i.convertElseClause(ifGram.ElseClause)
	if err != nil {
		return nil, err
	}
	return &ast.IfStatement{
		NodeType:   ast.TypeIfStatement,
		Test:       test,
		Consequent: consequent,
		Alternate:  alternate,
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
