package parser

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

var binaryOperatorFromCompound = map[string]string{
	"+=": "+",
	"-=": "-",
	"*=": "*",
	"/=": "/",
	"%=": "%",
}

// CompoundAssignmentConverter desugars compound assignment statements into
// standard `:=` reassignments.  The transformation is purely syntactic:
//
//	x += rhs   →   x := x + rhs
//	x -= rhs   →   x := x - rhs
//	x *= rhs   →   x := x * rhs
//	x /= rhs   →   x := x / rhs
//	x %= rhs   →   x := x % rhs
//
// This preserves Pine Script semantics (mutate and persist the series slot)
// while reusing the existing Reassignment codegen path downstream.
type CompoundAssignmentConverter struct {
	expressionConverter func(*Expression) (ast.Expression, error)
}

func NewCompoundAssignmentConverter(expressionConverter func(*Expression) (ast.Expression, error)) *CompoundAssignmentConverter {
	return &CompoundAssignmentConverter{expressionConverter: expressionConverter}
}

func (c *CompoundAssignmentConverter) CanHandle(stmt *Statement) bool {
	return stmt.Core != nil && stmt.Core.CompoundAssignment != nil
}

func (c *CompoundAssignmentConverter) Convert(stmt *Statement) (ast.Node, error) {
	ca := stmt.Core.CompoundAssignment

	binOp, ok := binaryOperatorFromCompound[ca.Operator]
	if !ok {
		return nil, fmt.Errorf("unknown compound assignment operator %q", ca.Operator)
	}

	rhs, err := c.expressionConverter(ca.Value)
	if err != nil {
		return nil, fmt.Errorf("compound assignment rhs: %w", err)
	}

	desugared := &ast.BinaryExpression{
		NodeType: ast.TypeBinaryExpression,
		Operator: binOp,
		Left:     buildIdentifier(ca.Name),
		Right:    rhs,
	}

	return buildVariableDeclaration(buildIdentifier(ca.Name), desugared, "var"), nil
}
