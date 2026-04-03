package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

type UDFTempVarAnalyzer struct {
	gen *generator
}

func NewUDFTempVarAnalyzer(g *generator) *UDFTempVarAnalyzer {
	return &UDFTempVarAnalyzer{gen: g}
}

func (a *UDFTempVarAnalyzer) Analyze(program *ast.Program) {
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

func (a *UDFTempVarAnalyzer) analyzeExpression(expr ast.Expression) {
	nestedCalls := a.gen.exprAnalyzer.FindNestedCalls(expr)

	for i := len(nestedCalls) - 1; i >= 0; i-- {
		callInfo := nestedCalls[i]

		a.registerConditionals(callInfo)

		if a.shouldSkipTopLevel(callInfo, expr) {
			continue
		}

		if a.gen.taRegistry.IsSupported(callInfo.FuncName) {
			continue
		}

		if a.isUDFCall(callInfo.FuncName) {
			a.gen.tempVarMgr.GetOrCreate(callInfo)
		}
	}
}

func (a *UDFTempVarAnalyzer) shouldSkipTopLevel(callInfo CallInfo, expr ast.Expression) bool {
	if callInfo.Call != expr {
		return false
	}

	if !a.isUDFCall(callInfo.FuncName) {
		return true
	}

	udfNestedCalls := a.gen.exprAnalyzer.FindNestedCalls(callInfo.Call)
	for _, nested := range udfNestedCalls {
		if nested.Call != callInfo.Call && a.gen.taRegistry.IsSupported(nested.FuncName) {
			return false
		}
	}

	return true
}

func (a *UDFTempVarAnalyzer) isUDFCall(funcName string) bool {
	varType, exists := a.gen.variables[funcName]
	return exists && varType == "function"
}

func (a *UDFTempVarAnalyzer) registerConditionals(callInfo CallInfo) {
	conditionals := a.gen.conditionalArgAnalyzer.FindInExpression(callInfo.Call)
	for _, condInfo := range conditionals {
		a.gen.tempVarMgr.RegisterConditional(condInfo.ContentHash, condInfo.Conditional)
	}
}
