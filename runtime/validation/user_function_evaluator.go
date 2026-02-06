package validation

import (
	"math"

	"github.com/quant5-lab/runner/ast"
)

type ScopedConstantStore interface {
	ConstantStore
	Set(name string, value float64)
	Delete(name string)
}

type UserFunctionEvaluator struct {
	evaluator NumericEvaluator
	functions FunctionStore
	constants ScopedConstantStore
	depth     int
}

const maxEvalDepth = 10

func NewUserFunctionEvaluator(evaluator NumericEvaluator, functions FunctionStore, constants ScopedConstantStore) *UserFunctionEvaluator {
	return &UserFunctionEvaluator{
		evaluator: evaluator,
		functions: functions,
		constants: constants,
	}
}

func (e *UserFunctionEvaluator) Evaluate(call *ast.CallExpression) float64 {
	if call == nil || e.functions == nil {
		return math.NaN()
	}

	funcName := e.extractCalleeName(call.Callee)
	if funcName == "" {
		return math.NaN()
	}

	fn, exists := e.functions.GetFunction(funcName)
	if !exists {
		return math.NaN()
	}

	if e.depth >= maxEvalDepth {
		return math.NaN()
	}

	if len(fn.Params) != len(call.Arguments) {
		return math.NaN()
	}

	e.depth++
	defer func() { e.depth-- }()

	if len(fn.Params) > 0 {
		return e.evaluateWithParams(fn, call.Arguments)
	}

	return e.evaluateBody(fn.Body)
}

func (e *UserFunctionEvaluator) evaluateWithParams(fn *ast.ArrowFunctionExpression, args []ast.Expression) float64 {
	bindings := make([]paramBinding, 0, len(fn.Params))
	for i, param := range fn.Params {
		argVal := e.evaluator.Evaluate(args[i])
		if math.IsNaN(argVal) {
			return math.NaN()
		}
		prev, existed := e.constants.Get(param.Name)
		bindings = append(bindings, paramBinding{name: param.Name, prev: prev, existed: existed})
		e.constants.Set(param.Name, argVal)
	}

	result := e.evaluateBody(fn.Body)

	for i := len(bindings) - 1; i >= 0; i-- {
		b := bindings[i]
		if b.existed {
			e.constants.Set(b.name, b.prev)
		} else {
			e.constants.Delete(b.name)
		}
	}

	return result
}

type paramBinding struct {
	name    string
	prev    float64
	existed bool
}

func (e *UserFunctionEvaluator) evaluateBody(body []ast.Node) float64 {
	if len(body) == 0 {
		return math.NaN()
	}

	lastNode := body[len(body)-1]

	if exprStmt, ok := lastNode.(*ast.ExpressionStatement); ok {
		return e.evaluator.Evaluate(exprStmt.Expression)
	}

	return math.NaN()
}

func (e *UserFunctionEvaluator) extractCalleeName(callee ast.Expression) string {
	if id, ok := callee.(*ast.Identifier); ok {
		return id.Name
	}
	return ""
}
