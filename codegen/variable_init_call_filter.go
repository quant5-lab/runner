package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

type VariableInitCallFilter struct {
	taRegistry        *TAFunctionRegistry
	inlineRegistry    *InlineFunctionRegistry
	runtimeOnlyFilter *RuntimeOnlyFunctionFilter
	exprAnalyzer      *ExpressionAnalyzer
}

func NewVariableInitCallFilter(
	taRegistry *TAFunctionRegistry,
	inlineRegistry *InlineFunctionRegistry,
	runtimeOnlyFilter *RuntimeOnlyFunctionFilter,
	exprAnalyzer *ExpressionAnalyzer,
) *VariableInitCallFilter {
	return &VariableInitCallFilter{
		taRegistry:        taRegistry,
		inlineRegistry:    inlineRegistry,
		runtimeOnlyFilter: runtimeOnlyFilter,
		exprAnalyzer:      exprAnalyzer,
	}
}

func (f *VariableInitCallFilter) FilterHoistable(
	nestedCalls []CallInfo,
	initExpr ast.Expression,
) []CallInfo {
	// The security() expression is evaluated at runtime by the bar evaluator, which interprets the
	// AST against the foreign symbol. Hoisting a call out of that expression generates incorrect
	// code for the wrong symbol and causes duplicate-declaration compile errors. A security()
	// expression that is itself a user-defined call is the exception: it and the calls in its
	// arguments are compiled to Go, so a TA call below it still needs a hoisted series.
	initIsSecurityCall := false
	if callExpr, ok := initExpr.(*ast.CallExpression); ok {
		initIsSecurityCall = IsSecurityFunction(extractCallFunctionName(callExpr))
	}

	var result []CallInfo
	for i := len(nestedCalls) - 1; i >= 0; i-- {
		callInfo := nestedCalls[i]

		if f.isDirectInit(callInfo, initExpr) {
			continue
		}
		if f.isSkippableInlineOnly(callInfo, initExpr) {
			continue
		}
		if f.runtimeOnlyFilter.IsRuntimeOnly(callInfo.FuncName) {
			continue
		}
		if initIsSecurityCall && !IsSecurityFunction(callInfo.FuncName) &&
			!f.exprAnalyzer.IsCompiledBySecurityExpression(callInfo.Call, initExpr) {
			continue
		}
		if f.requiresTempVar(callInfo) {
			result = append(result, callInfo)
		}
	}
	return result
}

func (f *VariableInitCallFilter) isDirectInit(callInfo CallInfo, initExpr ast.Expression) bool {
	return callInfo.Call == initExpr
}

func (f *VariableInitCallFilter) isSkippableInlineOnly(callInfo CallInfo, initExpr ast.Expression) bool {
	if f.inlineRegistry == nil || !f.inlineRegistry.IsInlineOnly(callInfo.FuncName) {
		return false
	}
	return !f.exprAnalyzer.IsInsideSecurityCall(callInfo.Call, initExpr)
}

func (f *VariableInitCallFilter) requiresTempVar(callInfo CallInfo) bool {
	if IsSecurityFunction(callInfo.FuncName) {
		return true
	}
	if f.taRegistry.IsSupported(callInfo.FuncName) {
		return true
	}
	if sharedTASignatures.Contains(callInfo.FuncName) {
		return true
	}
	return f.containsNestedTA(callInfo)
}

func (f *VariableInitCallFilter) containsNestedTA(callInfo CallInfo) bool {
	innerCalls := f.exprAnalyzer.FindNestedCalls(callInfo.Call)
	for _, inner := range innerCalls {
		if inner.Call != callInfo.Call {
			if f.taRegistry.IsSupported(inner.FuncName) || sharedTASignatures.Contains(inner.FuncName) {
				return true
			}
		}
	}
	return false
}
