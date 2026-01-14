package validation

import (
	"math"

	"github.com/quant5-lab/runner/ast"
)

type LiteralEvaluator struct{}

func NewLiteralEvaluator() *LiteralEvaluator {
	return &LiteralEvaluator{}
}

func (e *LiteralEvaluator) Evaluate(literal *ast.Literal) float64 {
	if literal == nil {
		return math.NaN()
	}

	switch value := literal.Value.(type) {
	case float64:
		return value
	case int:
		return float64(value)
	default:
		return math.NaN()
	}
}
