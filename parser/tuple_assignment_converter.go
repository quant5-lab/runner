package parser

import "github.com/quant5-lab/runner/ast"

type TupleAssignmentConverter struct {
	expressionConverter func(*Expression) (ast.Expression, error)
}

func NewTupleAssignmentConverter(expressionConverter func(*Expression) (ast.Expression, error)) *TupleAssignmentConverter {
	return &TupleAssignmentConverter{
		expressionConverter: expressionConverter,
	}
}

func (t *TupleAssignmentConverter) CanHandle(stmt *Statement) bool {
	return stmt.Core != nil && stmt.Core.TupleAssignment != nil
}

func (t *TupleAssignmentConverter) Convert(stmt *Statement) (ast.Node, error) {
	tuple := stmt.Core.TupleAssignment

	if tuple.Value == nil {
		return t.convertArrayLiteralStatement(tuple.Names)
	}

	return t.convertTupleDestructuring(tuple.Names, tuple.Value)
}

func (t *TupleAssignmentConverter) convertArrayLiteralStatement(names []string) (ast.Node, error) {
	elements := make([]ast.Expression, len(names))
	for i, name := range names {
		elements[i] = buildIdentifier(name)
	}

	return &ast.ExpressionStatement{
		NodeType: ast.TypeExpressionStatement,
		Expression: &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    elements,
			Raw:      "[...]",
		},
	}, nil
}

func (t *TupleAssignmentConverter) convertTupleDestructuring(names []string, value *Expression) (ast.Node, error) {
	init, err := t.expressionConverter(value)
	if err != nil {
		return nil, err
	}

	return buildVariableDeclaration(
		buildArrayPattern(names),
		init,
		"let",
	), nil
}
