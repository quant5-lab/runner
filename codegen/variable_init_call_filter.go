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
	// TA calls nested inside request.security() are evaluated at runtime by the bar
	// evaluator. Hoisting them to the main context generates incorrect code for the
	// wrong symbol and causes duplicate-declaration compile errors.
	initIsSecurityCall := false
	if callExpr, ok := initExpr.(*ast.CallExpression); ok {
		funcName := ""
		switch c := callExpr.Callee.(type) {
		case *ast.Identifier:
			funcName = c.Name
		case *ast.MemberExpression:
			if obj, ok := c.Object.(*ast.Identifier); ok {
				if prop, ok := c.Property.(*ast.Identifier); ok {
					funcName = obj.Name + "." + prop.Name
				}
			}
		}
		initIsSecurityCall = IsSecurityFunction(funcName)
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
		// Skip TA functions that are arguments of a security() call — they are
		// computed inside the security bar evaluator, not the main bar loop.
		if initIsSecurityCall && !IsSecurityFunction(callInfo.FuncName) {
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
