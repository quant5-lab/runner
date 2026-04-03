package security

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

// expressionKey returns a stable, human-readable string for any ast.Expression,
// suitable for use as a TA state-manager cache-key component.
func expressionKey(expr ast.Expression) string {
	if expr == nil {
		return "nil"
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Name
	case *ast.Literal:
		return fmt.Sprintf("%v", e.Value)
	case *ast.UnaryExpression:
		return fmt.Sprintf("(%s%s)", e.Operator, expressionKey(e.Argument))
	case *ast.BinaryExpression:
		return fmt.Sprintf("(%s%s%s)", expressionKey(e.Left), e.Operator, expressionKey(e.Right))
	case *ast.MemberExpression:
		if propID, ok := e.Property.(*ast.Identifier); ok {
			return fmt.Sprintf("%s.%s", expressionKey(e.Object), propID.Name)
		}
		return fmt.Sprintf("%s[%s]", expressionKey(e.Object), expressionKey(e.Property))
	case *ast.CallExpression:
		args := make([]string, len(e.Arguments))
		for i, arg := range e.Arguments {
			args[i] = expressionKey(arg)
		}
		return fmt.Sprintf("%s(%s)", expressionKey(e.Callee), strings.Join(args, ","))
	case *ast.ConditionalExpression:
		return fmt.Sprintf("(%s?%s:%s)", expressionKey(e.Test), expressionKey(e.Consequent), expressionKey(e.Alternate))
	default:
		return fmt.Sprintf("expr<%T>", expr)
	}
}
