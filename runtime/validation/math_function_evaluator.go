package validation

import (
	"math"

	"github.com/quant5-lab/runner/ast"
)

type MathFunctionEvaluator struct {
	evaluator NumericEvaluator
}

func NewMathFunctionEvaluator(evaluator NumericEvaluator) *MathFunctionEvaluator {
	return &MathFunctionEvaluator{
		evaluator: evaluator,
	}
}

func (e *MathFunctionEvaluator) Evaluate(call *ast.CallExpression) float64 {
	if call == nil {
		return math.NaN()
	}

	functionName := e.extractFunctionName(call.Callee)
	if functionName == "" {
		return math.NaN()
	}

	switch functionName {
	case "pow", "math.pow":
		return e.evaluatePower(call.Arguments)
	case "round", "math.round":
		return e.evaluateRound(call.Arguments)
	case "sqrt", "math.sqrt":
		return e.evaluateSquareRoot(call.Arguments)
	case "floor", "math.floor":
		return e.evaluateFloor(call.Arguments)
	case "ceil", "math.ceil":
		return e.evaluateCeiling(call.Arguments)
	default:
		return math.NaN()
	}
}

func (e *MathFunctionEvaluator) extractFunctionName(callee ast.Expression) string {
	if member, ok := callee.(*ast.MemberExpression); ok {
		return e.extractMemberFunctionName(member)
	}

	if identifier, ok := callee.(*ast.Identifier); ok {
		return identifier.Name
	}

	return ""
}

func (e *MathFunctionEvaluator) extractMemberFunctionName(member *ast.MemberExpression) string {
	object, ok := member.Object.(*ast.Identifier)
	if !ok || object.Name != "math" {
		return ""
	}

	property, ok := member.Property.(*ast.Identifier)
	if !ok {
		return ""
	}

	return property.Name
}

func (e *MathFunctionEvaluator) evaluatePower(args []ast.Expression) float64 {
	if len(args) != 2 {
		return math.NaN()
	}

	base := e.evaluator.Evaluate(args[0])
	exponent := e.evaluator.Evaluate(args[1])

	if math.IsNaN(base) || math.IsNaN(exponent) {
		return math.NaN()
	}

	return math.Pow(base, exponent)
}

func (e *MathFunctionEvaluator) evaluateRound(args []ast.Expression) float64 {
	if len(args) < 1 {
		return math.NaN()
	}

	value := e.evaluator.Evaluate(args[0])
	if math.IsNaN(value) {
		return math.NaN()
	}

	return math.Round(value)
}

func (e *MathFunctionEvaluator) evaluateSquareRoot(args []ast.Expression) float64 {
	if len(args) != 1 {
		return math.NaN()
	}

	value := e.evaluator.Evaluate(args[0])
	if math.IsNaN(value) {
		return math.NaN()
	}

	return math.Sqrt(value)
}

func (e *MathFunctionEvaluator) evaluateFloor(args []ast.Expression) float64 {
	if len(args) != 1 {
		return math.NaN()
	}

	value := e.evaluator.Evaluate(args[0])
	if math.IsNaN(value) {
		return math.NaN()
	}

	return math.Floor(value)
}

func (e *MathFunctionEvaluator) evaluateCeiling(args []ast.Expression) float64 {
	if len(args) != 1 {
		return math.NaN()
	}

	value := e.evaluator.Evaluate(args[0])
	if math.IsNaN(value) {
		return math.NaN()
	}

	return math.Ceil(value)
}
