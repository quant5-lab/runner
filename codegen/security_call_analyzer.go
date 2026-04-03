package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

type SecurityCallAnalyzer struct {
	gen *generator
}

func NewSecurityCallAnalyzer(g *generator) *SecurityCallAnalyzer {
	return &SecurityCallAnalyzer{gen: g}
}

func (a *SecurityCallAnalyzer) Analyze(program *ast.Program) {
	for _, stmt := range program.Body {
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			for _, declarator := range varDecl.Declarations {
				if declarator.Init != nil {
					a.analyzeExpression(declarator.Init)
				}
			}
		}
	}
}

func (a *SecurityCallAnalyzer) analyzeExpression(expr ast.Expression) {
	nestedCalls := a.gen.exprAnalyzer.FindNestedCalls(expr)

	for i := len(nestedCalls) - 1; i >= 0; i-- {
		callInfo := nestedCalls[i]

		if a.shouldSkipTopLevel(callInfo, expr) {
			continue
		}

		if a.shouldSkipRuntimeOnly(callInfo.FuncName) {
			continue
		}

		if a.shouldRegisterTempVar(callInfo) {
			a.gen.tempVarMgr.GetOrCreate(callInfo)
		}

		a.registerConditionals(callInfo)
	}
}

func (a *SecurityCallAnalyzer) shouldSkipTopLevel(callInfo CallInfo, expr ast.Expression) bool {
	if callInfo.Call != expr {
		return false
	}

	isSumWithConditional := (callInfo.FuncName == "sum" || callInfo.FuncName == "math.sum") &&
		len(callInfo.Call.Arguments) >= 1

	if isSumWithConditional {
		if _, ok := callInfo.Call.Arguments[0].(*ast.ConditionalExpression); ok {
			return false
		}
	}

	return true
}

func (a *SecurityCallAnalyzer) shouldSkipRuntimeOnly(funcName string) bool {
	// valuewhen needs temp var even though it's inline-only
	if funcName == "valuewhen" || funcName == "ta.valuewhen" {
		return false
	}

	if a.gen.inlineRegistry != nil && a.gen.inlineRegistry.IsInlineOnly(funcName) {
		return true
	}
	return a.gen.runtimeOnlyFilter.IsRuntimeOnly(funcName)
}

func (a *SecurityCallAnalyzer) shouldRegisterTempVar(callInfo CallInfo) bool {
	isTAFunction := a.gen.taRegistry.IsSupported(callInfo.FuncName)
	if isTAFunction {
		return true
	}

	mathNestedCalls := a.gen.exprAnalyzer.FindNestedCalls(callInfo.Call)
	for _, mathNested := range mathNestedCalls {
		if mathNested.Call != callInfo.Call && a.gen.taRegistry.IsSupported(mathNested.FuncName) {
			return true
		}
	}

	if callInfo.FuncName == "sum" || callInfo.FuncName == "math.sum" {
		if len(callInfo.Call.Arguments) >= 1 {
			if _, ok := callInfo.Call.Arguments[0].(*ast.ConditionalExpression); ok {
				return true
			}
		}
	}

	// valuewhen needs temp var registration for inline Series storage
	if callInfo.FuncName == "valuewhen" || callInfo.FuncName == "ta.valuewhen" {
		return true
	}

	return IsInputFuncName(callInfo.FuncName)
}

func (a *SecurityCallAnalyzer) registerConditionals(callInfo CallInfo) {
	conditionals := a.gen.conditionalArgAnalyzer.FindInExpression(callInfo.Call)
	for _, condInfo := range conditionals {
		a.gen.tempVarMgr.RegisterConditional(condInfo.ContentHash, condInfo.Conditional)
	}
}
