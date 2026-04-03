package codegen

import (
	"fmt"
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

type PeriodEvaluator struct {
	constEvaluator *validation.WarmupAnalyzer
}

func NewPeriodEvaluator(constEvaluator *validation.WarmupAnalyzer) *PeriodEvaluator {
	return &PeriodEvaluator{
		constEvaluator: constEvaluator,
	}
}

func (e *PeriodEvaluator) Evaluate(expr ast.Expression) PeriodEvaluationResult {
	if lit, ok := expr.(*ast.Literal); ok {
		return e.evaluateLiteral(lit)
	}

	if call, ok := expr.(*ast.CallExpression); ok {
		if inputVal := e.tryEvaluateInputCall(call); inputVal > 0 {
			return NewCompileTimeConstantPeriod(inputVal)
		}
	}

	periodValue := e.constEvaluator.EvaluateConstant(expr)
	if !math.IsNaN(periodValue) && periodValue > 0 {
		return NewCompileTimeConstantPeriod(int(periodValue))
	}

	if periodValue <= 0 && !math.IsNaN(periodValue) {
		return NewFailedPeriodEvaluation(fmt.Sprintf("period must be positive, got %.0f", periodValue))
	}

	return NewRuntimeDynamicPeriod(expr)
}

func (e *PeriodEvaluator) tryEvaluateInputCall(call *ast.CallExpression) int {
	funcName := e.extractFunctionName(call)
	if funcName != "input.int" && funcName != "input" {
		return 0
	}

	if len(call.Arguments) == 0 {
		return 0
	}

	firstArg := call.Arguments[0]
	if lit, ok := firstArg.(*ast.Literal); ok {
		switch v := lit.Value.(type) {
		case float64:
			return int(v)
		case int:
			return v
		}
	}

	return 0
}

func (e *PeriodEvaluator) extractFunctionName(call *ast.CallExpression) string {
	switch callee := call.Callee.(type) {
	case *ast.Identifier:
		return callee.Name
	case *ast.MemberExpression:
		if obj, ok := callee.Object.(*ast.Identifier); ok {
			if prop, ok := callee.Property.(*ast.Identifier); ok {
				return obj.Name + "." + prop.Name
			}
		}
	}
	return ""
}

func (e *PeriodEvaluator) evaluateLiteral(lit *ast.Literal) PeriodEvaluationResult {
	switch v := lit.Value.(type) {
	case float64:
		if v <= 0 {
			return NewFailedPeriodEvaluation(fmt.Sprintf("period must be positive, got %.0f", v))
		}
		return NewCompileTimeConstantPeriod(int(v))
	case int:
		if v <= 0 {
			return NewFailedPeriodEvaluation(fmt.Sprintf("period must be positive, got %d", v))
		}
		return NewCompileTimeConstantPeriod(v)
	default:
		return NewFailedPeriodEvaluation(fmt.Sprintf("period must be numeric, got %T", v))
	}
}
