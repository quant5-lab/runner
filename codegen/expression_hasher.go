package codegen

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/*
ExpressionHasher generates short unique hashes from AST expressions.
Used to create unique series names for stateful indicators (RMA, EMA) to prevent collisions.

Example: rma(up > down ? up : 0, 18) and rma(down > up ? down : 0, 18) both have period=18
but different sources, so they need different series names: "_rma_18_a1b2c3" vs "_rma_18_d4e5f6"
*/
type ExpressionHasher struct{}

/*
Hash generates a short (8-char) hash from an AST expression.
Returns empty string if expression is nil.
*/
func (h *ExpressionHasher) Hash(expr ast.Expression) string {
	if expr == nil {
		return ""
	}

	// Generate a canonical string representation
	canonical := h.canonicalize(expr)

	// Create SHA256 hash and take first 8 characters
	hash := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(hash[:])[:8]
}

/*
canonicalize converts an AST expression to a stable string representation.
Different expressions produce different strings, same expression produces same string.
*/
func (h *ExpressionHasher) canonicalize(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return fmt.Sprintf("id:%s", e.Name)

	case *ast.Literal:
		return fmt.Sprintf("lit:%v", e.Value)

	case *ast.MemberExpression:
		obj := h.canonicalize(e.Object)
		prop := ""
		if id, ok := e.Property.(*ast.Identifier); ok {
			prop = id.Name
		} else {
			prop = h.canonicalize(e.Property)
		}
		return fmt.Sprintf("mem:%s.%s", obj, prop)

	case *ast.BinaryExpression:
		left := h.canonicalize(e.Left)
		right := h.canonicalize(e.Right)
		return fmt.Sprintf("bin:%s%s%s", left, e.Operator, right)

	case *ast.UnaryExpression:
		arg := h.canonicalize(e.Argument)
		return fmt.Sprintf("unary:%s%s", e.Operator, arg)

	case *ast.ConditionalExpression:
		test := h.canonicalize(e.Test)
		cons := h.canonicalize(e.Consequent)
		alt := h.canonicalize(e.Alternate)
		return fmt.Sprintf("cond:%s?%s:%s", test, cons, alt)

	case *ast.CallExpression:
		callee := h.canonicalize(e.Callee)
		args := ""
		for i, arg := range e.Arguments {
			if i > 0 {
				args += ","
			}
			args += h.canonicalize(arg)
		}
		return fmt.Sprintf("call:%s(%s)", callee, args)

	case *ast.LogicalExpression:
		left := h.canonicalize(e.Left)
		right := h.canonicalize(e.Right)
		return fmt.Sprintf("log:%s%s%s", left, e.Operator, right)

	default:
		// Fallback: use type name
		return fmt.Sprintf("unknown:%T", expr)
	}
}
