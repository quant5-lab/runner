package codegen

import "github.com/quant5-lab/runner/ast"

type PeriodEvaluationKind int

const (
	PeriodKindCompileTime PeriodEvaluationKind = iota
	PeriodKindRuntimeDynamic
	PeriodKindFailed
)

type PeriodEvaluationResult struct {
	Kind          PeriodEvaluationKind
	StaticValue   int
	DynamicExpr   ast.Expression
	FailureReason string
}

func NewCompileTimeConstantPeriod(value int) PeriodEvaluationResult {
	return PeriodEvaluationResult{
		Kind:        PeriodKindCompileTime,
		StaticValue: value,
	}
}

func NewRuntimeDynamicPeriod(expr ast.Expression) PeriodEvaluationResult {
	return PeriodEvaluationResult{
		Kind:        PeriodKindRuntimeDynamic,
		DynamicExpr: expr,
	}
}

func NewFailedPeriodEvaluation(reason string) PeriodEvaluationResult {
	return PeriodEvaluationResult{
		Kind:          PeriodKindFailed,
		FailureReason: reason,
	}
}

func (r PeriodEvaluationResult) IsCompileTimeConstant() bool {
	return r.Kind == PeriodKindCompileTime
}

func (r PeriodEvaluationResult) IsRuntimeDynamic() bool {
	return r.Kind == PeriodKindRuntimeDynamic
}

func (r PeriodEvaluationResult) IsFailed() bool {
	return r.Kind == PeriodKindFailed
}
