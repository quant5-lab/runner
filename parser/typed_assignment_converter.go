package parser

import "github.com/quant5-lab/runner/ast"

type TypedAssignmentConverter struct {
	expressionConverter func(*Expression) (ast.Expression, error)
}

func NewTypedAssignmentConverter(expressionConverter func(*Expression) (ast.Expression, error)) *TypedAssignmentConverter {
	return &TypedAssignmentConverter{
		expressionConverter: expressionConverter,
	}
}

func (t *TypedAssignmentConverter) CanHandle(stmt *Statement) bool {
	return stmt.TypedAssignment != nil
}

func (t *TypedAssignmentConverter) Convert(stmt *Statement) (ast.Node, error) {
	init, err := t.expressionConverter(stmt.TypedAssignment.Value)
	if err != nil {
		return nil, err
	}

	return buildVariableDeclaration(
		buildIdentifier(stmt.TypedAssignment.Name),
		init,
		"let",
	), nil
}
