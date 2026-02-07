package parser

import "github.com/quant5-lab/runner/ast"

type AssignmentConverter struct {
	expressionConverter func(*Expression) (ast.Expression, error)
}

func NewAssignmentConverter(expressionConverter func(*Expression) (ast.Expression, error)) *AssignmentConverter {
	return &AssignmentConverter{
		expressionConverter: expressionConverter,
	}
}

func (a *AssignmentConverter) CanHandle(stmt *Statement) bool {
	return stmt.Core != nil && stmt.Core.Assignment != nil
}

func (a *AssignmentConverter) Convert(stmt *Statement) (ast.Node, error) {
	init, err := a.expressionConverter(stmt.Core.Assignment.Value)
	if err != nil {
		return nil, err
	}

	return buildVariableDeclaration(
		buildIdentifier(stmt.Core.Assignment.Name),
		init,
		"let",
	), nil
}
