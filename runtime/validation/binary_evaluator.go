package validation

import (
	"math"

	"github.com/quant5-lab/runner/ast"
)

type BinaryEvaluator struct {
	evaluator NumericEvaluator
}

func NewBinaryEvaluator(evaluator NumericEvaluator) *BinaryEvaluator {
	return &BinaryEvaluator{
		evaluator: evaluator,
	}
}

func (e *BinaryEvaluator) Evaluate(binary *ast.BinaryExpression) float64 {
	if binary == nil {
		return math.NaN()
	}

	left := e.evaluator.Evaluate(binary.Left)
	right := e.evaluator.Evaluate(binary.Right)

	if math.IsNaN(left) || math.IsNaN(right) {
		return math.NaN()
	}

	switch binary.Operator {
	case "+":
		return left + right
	case "-":
		return left - right
	case "*":
		return left * right
	case "/":
		return e.evaluateDivision(left, right)
	default:
		return math.NaN()
	}
}

func (e *BinaryEvaluator) evaluateDivision(numerator, denominator float64) float64 {
	if denominator == 0 {
		return math.NaN()
	}
	return numerator / denominator
}
