package parser

import "github.com/quant5-lab/runner/ast"

type InlineStatementListConverter struct {
	converter *Converter
}

func NewInlineStatementListConverter(
	converter *Converter,
) *InlineStatementListConverter {
	return &InlineStatementListConverter{
		converter: converter,
	}
}

func (c *InlineStatementListConverter) Convert(list *InlineStatementList) ([]ast.Node, error) {
	nodes := []ast.Node{}

	for _, stmt := range list.Statements {
		node, err := c.convertInlineStatement(stmt)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	finalNode, err := c.convertFinalExpression(list.FinalExpression)
	if err != nil {
		return nil, err
	}
	nodes = append(nodes, finalNode)

	return nodes, nil
}

func (c *InlineStatementListConverter) convertInlineStatement(stmt *InlineStatement) (ast.Node, error) {
	value, err := c.converter.convertTernaryExpr(stmt.Value)
	if err != nil {
		return nil, err
	}

	kind := "let"
	if stmt.Op == ":=" {
		kind = "var"
	}

	return buildVariableDeclaration(
		buildIdentifier(stmt.Name),
		value,
		kind,
	), nil
}

func (c *InlineStatementListConverter) convertFinalExpression(expr *TernaryExpr) (ast.Node, error) {
	node, err := c.converter.convertTernaryExpr(expr)
	if err != nil {
		return nil, err
	}

	return &ast.ExpressionStatement{
		NodeType:   ast.TypeExpressionStatement,
		Expression: node,
	}, nil
}
