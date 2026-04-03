package parser

import "github.com/quant5-lab/runner/ast"

type ReassignmentConverter struct {
	expressionConverter func(*Expression) (ast.Expression, error)
}

func NewReassignmentConverter(expressionConverter func(*Expression) (ast.Expression, error)) *ReassignmentConverter {
	return &ReassignmentConverter{
		expressionConverter: expressionConverter,
	}
}

func (r *ReassignmentConverter) CanHandle(stmt *Statement) bool {
	return stmt.Core != nil && stmt.Core.Reassignment != nil
}

func (r *ReassignmentConverter) Convert(stmt *Statement) (ast.Node, error) {
	init, err := r.expressionConverter(stmt.Core.Reassignment.Value)
	if err != nil {
		return nil, err
	}

	return buildVariableDeclaration(
		buildIdentifier(stmt.Core.Reassignment.Name),
		init,
		"var",
	), nil
}
