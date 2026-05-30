package parser

import "github.com/quant5-lab/runner/ast"

// TypedAssignmentConverter transforms bare typed declarations (e.g. float x = expr)
// into VariableDeclaration AST nodes. Drawing-type declarations degrade to NaN.
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
	ta := stmt.Core.TypedAssignment

	init, err := t.expressionConverter(ta.Value)
	if err != nil {
		return nil, err
	}

	if isDrawingTypeHint(&ta.TypeHint) {
		init = nanDegradedDrawingInit(&ta.TypeHint, false)
	}

	return buildVariableDeclaration(
		buildIdentifier(ta.Name),
		init,
		"let",
	), nil
}
