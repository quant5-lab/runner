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
	return stmt.Core != nil && stmt.Core.TypedAssignment != nil
}

func (t *TypedAssignmentConverter) Convert(stmt *Statement) (ast.Node, error) {
	init, err := t.expressionConverter(stmt.Core.TypedAssignment.Value)
	if err != nil {
		return nil, err
	}

	return buildVariableDeclaration(
		buildIdentifier(stmt.Core.TypedAssignment.Name),
		init,
		"let",
	), nil
}
