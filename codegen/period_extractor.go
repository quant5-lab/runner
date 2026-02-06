package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type PeriodExtractor struct {
	evaluator  *PeriodEvaluator
	repository *PeriodRequirementRepository
}

func NewPeriodExtractor(evaluator *PeriodEvaluator, repository *PeriodRequirementRepository) *PeriodExtractor {
	return &PeriodExtractor{
		evaluator:  evaluator,
		repository: repository,
	}
}

func (e *PeriodExtractor) ExtractAndValidate(
	call *ast.CallExpression,
	functionName string,
) (PeriodEvaluationResult, error) {
	spec, exists := e.repository.GetSpec(functionName)
	if !exists {
		return PeriodEvaluationResult{}, fmt.Errorf("unknown TA function: %s", functionName)
	}

	if len(call.Arguments) <= spec.ParameterPosition {
		return PeriodEvaluationResult{}, fmt.Errorf(
			"%s requires period argument at position %d",
			functionName,
			spec.ParameterPosition,
		)
	}

	periodExpr := call.Arguments[spec.ParameterPosition]
	result := e.evaluator.Evaluate(periodExpr)

	if result.IsFailed() {
		return result, fmt.Errorf("%s: %s", functionName, result.FailureReason)
	}

	if result.IsRuntimeDynamic() && !spec.PeriodQualifier.AllowsRuntimeDynamic() {
		return result, fmt.Errorf(
			"%s period must be compile-time constant (PineScript requires %s)",
			functionName,
			spec.PeriodQualifier.String(),
		)
	}

	return result, nil
}
