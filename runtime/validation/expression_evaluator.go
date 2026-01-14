package validation

import (
	"math"

	"github.com/quant5-lab/runner/ast"
)

type ConstantStore interface {
	Get(name string) (float64, bool)
}

type ExpressionEvaluator struct {
	constants        ConstantStore
	literalEvaluator *LiteralEvaluator
	unaryEvaluator   *UnaryEvaluator
	binaryEvaluator  *BinaryEvaluator
	mathEvaluator    *MathFunctionEvaluator
	identifierLookup *IdentifierLookup
}

func NewExpressionEvaluator(constants ConstantStore) *ExpressionEvaluator {
	ev := &ExpressionEvaluator{
		constants:        constants,
		literalEvaluator: NewLiteralEvaluator(),
		identifierLookup: NewIdentifierLookup(constants),
	}

	ev.unaryEvaluator = NewUnaryEvaluator(ev)
	ev.binaryEvaluator = NewBinaryEvaluator(ev)
	ev.mathEvaluator = NewMathFunctionEvaluator(ev)

	return ev
}

func (e *ExpressionEvaluator) Evaluate(expr ast.Expression) float64 {
	if expr == nil {
		return math.NaN()
	}

	switch node := expr.(type) {
	case *ast.Literal:
		return e.literalEvaluator.Evaluate(node)

	case *ast.Identifier:
		return e.identifierLookup.Resolve(node)

	case *ast.MemberExpression:
		return e.identifierLookup.ResolveWrappedVariable(node)

	case *ast.UnaryExpression:
		return e.unaryEvaluator.Evaluate(node)

	case *ast.BinaryExpression:
		return e.binaryEvaluator.Evaluate(node)

	case *ast.CallExpression:
		return e.mathEvaluator.Evaluate(node)

	case *ast.ConditionalExpression:
		return math.NaN()

	default:
		return math.NaN()
	}
}
