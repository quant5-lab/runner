package parser

import "github.com/quant5-lab/runner/ast"

/*
VarAssignmentConverter transforms VarAssignment grammar nodes into ESTree VariableDeclaration
AST nodes with the Persistence field set to "var" or "varip".

The converter produces Kind="let" (declaration, not reassignment) while Persistence captures
the var/varip modifier. This separation preserves the existing Kind semantics where:
  - Kind="let" → new variable declaration (Go :=)
  - Kind="var" → reassignment of existing variable (Go =)
  - Persistence="var"/"varip" → initialize once on bar 0, carry forward on subsequent bars
*/
type VarAssignmentConverter struct {
	expressionConverter func(*Expression) (ast.Expression, error)
}

func NewVarAssignmentConverter(expressionConverter func(*Expression) (ast.Expression, error)) *VarAssignmentConverter {
	return &VarAssignmentConverter{
		expressionConverter: expressionConverter,
	}
}

func (v *VarAssignmentConverter) CanHandle(stmt *Statement) bool {
	return stmt.Core != nil && stmt.Core.VarAssignment != nil
}

func (v *VarAssignmentConverter) Convert(stmt *Statement) (ast.Node, error) {
	va := stmt.Core.VarAssignment

	init, err := v.expressionConverter(va.Value)
	if err != nil {
		return nil, err
	}

	var pattern ast.Pattern
	if len(va.TupleNames) > 0 {
		pattern = buildArrayPattern(va.TupleNames)
	} else {
		pattern = buildIdentifier(*va.Name)
	}

	decl := buildVariableDeclaration(pattern, init, "let")
	decl.Persistence = va.Modifier // "var" or "varip"
	return decl, nil
}
