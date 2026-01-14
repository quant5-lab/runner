package validation

import (
	"math"

	"github.com/quant5-lab/runner/ast"
)

type NumericEvaluator interface {
	Evaluate(ast.Expression) float64
}

type UnaryEvaluator struct {
	evaluator NumericEvaluator
}

func NewUnaryEvaluator(evaluator NumericEvaluator) *UnaryEvaluator {
	return &UnaryEvaluator{
		evaluator: evaluator,
	}
}

func (e *UnaryEvaluator) Evaluate(unary *ast.UnaryExpression) float64 {
	if unary == nil {
		return math.NaN()
	}

	operand := e.evaluator.Evaluate(unary.Argument)
	if math.IsNaN(operand) {
		return math.NaN()
	}

	switch unary.Operator {
	case "-":
		return -operand
	case "+":
		return operand
	case "!":
		return e.evaluateLogicalNot(operand)
	default:
		return math.NaN()
	}
}

func (e *UnaryEvaluator) evaluateLogicalNot(operand float64) float64 {
	if operand == 0 {
		return 1
	}
	return 0
}
