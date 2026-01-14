package source_identity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type IdentifierFactory struct{}

func NewIdentifierFactory() *IdentifierFactory {
	return &IdentifierFactory{}
}

func (f *IdentifierFactory) CreateFromExpression(expr ast.Expression) SourceIdentifier {
	if expr == nil {
		return NewSourceIdentifier("")
	}

	canonical := f.canonicalize(expr)
	hash := sha256.Sum256([]byte(canonical))
	return NewSourceIdentifier(hex.EncodeToString(hash[:])[:8])
}

func (f *IdentifierFactory) canonicalize(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return fmt.Sprintf("id:%s", e.Name)

	case *ast.Literal:
		return fmt.Sprintf("lit:%v", e.Value)

	case *ast.MemberExpression:
		obj := f.canonicalize(e.Object)
		prop := ""
		if id, ok := e.Property.(*ast.Identifier); ok {
			prop = id.Name
		} else {
			prop = f.canonicalize(e.Property)
		}
		return fmt.Sprintf("mem:%s.%s", obj, prop)

	case *ast.BinaryExpression:
		left := f.canonicalize(e.Left)
		right := f.canonicalize(e.Right)
		return fmt.Sprintf("bin:%s%s%s", left, e.Operator, right)

	case *ast.UnaryExpression:
		arg := f.canonicalize(e.Argument)
		return fmt.Sprintf("unary:%s%s", e.Operator, arg)

	case *ast.ConditionalExpression:
		test := f.canonicalize(e.Test)
		cons := f.canonicalize(e.Consequent)
		alt := f.canonicalize(e.Alternate)
		return fmt.Sprintf("cond:%s?%s:%s", test, cons, alt)

	case *ast.CallExpression:
		callee := f.canonicalize(e.Callee)
		args := ""
		for i, arg := range e.Arguments {
			if i > 0 {
				args += ","
			}
			args += f.canonicalize(arg)
		}
		return fmt.Sprintf("call:%s(%s)", callee, args)

	case *ast.LogicalExpression:
		left := f.canonicalize(e.Left)
		right := f.canonicalize(e.Right)
		return fmt.Sprintf("log:%s%s%s", left, e.Operator, right)

	default:
		return fmt.Sprintf("unknown:%T", expr)
	}
}
