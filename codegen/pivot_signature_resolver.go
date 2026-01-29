package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type ResolvedPivotCall struct {
	SourceExpr  ast.Expression
	LeftPeriod  ast.Expression
	RightPeriod ast.Expression
	UsesDefault bool
}

type PivotSignatureResolver struct{}

func NewPivotSignatureResolver() *PivotSignatureResolver {
	return &PivotSignatureResolver{}
}

func (r *PivotSignatureResolver) Resolve(funcName string, call *ast.CallExpression) (*ResolvedPivotCall, error) {
	argCount := len(call.Arguments)

	if argCount == 2 {
		return r.resolveTwoArgForm(funcName, call)
	}

	if argCount == 3 {
		return r.resolveThreeArgForm(call)
	}

	return nil, fmt.Errorf("pivot functions require 2 or 3 arguments, got %d", argCount)
}

func (r *PivotSignatureResolver) resolveTwoArgForm(funcName string, call *ast.CallExpression) (*ResolvedPivotCall, error) {
	defaultSource := "high"
	if funcName == "ta.pivotlow" {
		defaultSource = "low"
	}

	return &ResolvedPivotCall{
		SourceExpr:  &ast.Identifier{Name: defaultSource},
		LeftPeriod:  call.Arguments[0],
		RightPeriod: call.Arguments[1],
		UsesDefault: true,
	}, nil
}

func (r *PivotSignatureResolver) resolveThreeArgForm(call *ast.CallExpression) (*ResolvedPivotCall, error) {
	return &ResolvedPivotCall{
		SourceExpr:  call.Arguments[0],
		LeftPeriod:  call.Arguments[1],
		RightPeriod: call.Arguments[2],
		UsesDefault: false,
	}, nil
}

func (r *PivotSignatureResolver) IsPivotFunction(funcName string) bool {
	return funcName == "ta.pivothigh" || funcName == "ta.pivotlow"
}
