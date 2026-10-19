package parser

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/*
VarAssignmentConverter transforms VarAssignment grammar nodes into ESTree VariableDeclaration
AST nodes with the Persistence field set to "var" or "varip".

The converter produces Kind="let" (declaration, not reassignment) while Persistence captures
the var/varip modifier. This separation preserves the existing Kind semantics where:
  - Kind="let" → new variable declaration (Go :=)
  - Kind="var" → reassignment of existing variable (Go =)
  - Persistence="var"/"varip" → initialize once on bar 0, carry forward on subsequent bars

Drawing-type annotations (line, label, box, table, linefill, polyline) are display-only.
Their declarations degrade to NaN with the TypeHint preserved in the AST comment so the
codegen can report them via featureGaps without crashing.
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

	if va.Name == nil && len(va.TupleNames) == 0 {
		return nil, fmt.Errorf("var assignment has no variable name")
	}

	init, err := v.expressionConverter(va.Value)
	if err != nil {
		return nil, err
	}

	if isDrawingTypeHint(va.TypeHint) {
		init = nanDegradedDrawingInit(va.TypeHint, true)
	}

	var pattern ast.Pattern
	if len(va.TupleNames) > 0 {
		pattern = buildArrayPattern(va.TupleNames)
	} else {
		pattern = buildIdentifier(*va.Name)
	}

	decl := buildVariableDeclaration(pattern, init, "let")
	decl.Persistence = va.Modifier
	return decl, nil
}

// isDrawingTypeHint reports whether the type hint refers to a Pine drawing object.
// Drawing objects have no numeric value; they degrade to NaN in trade-logic context.
func isDrawingTypeHint(hint *string) bool {
	if hint == nil {
		return false
	}
	switch *hint {
	case "line", "label", "box", "table", "linefill", "polyline":
		return true
	}
	return false
}

// nanDegradedDrawingInit returns a NaN literal as the init expression for drawing-type
// declarations. The TypeHint string is preserved as the raw annotation for featureGap reporting.
func nanDegradedDrawingInit(hint *string, isArray bool) ast.Expression {
	annotation := ""
	if hint != nil {
		annotation = *hint
		if isArray {
			annotation += "[]"
		}
	}
	return &ast.Literal{
		Value: float64(0),
		Raw:   fmt.Sprintf("/* drawing:%s */math.NaN()", annotation),
	}
}
