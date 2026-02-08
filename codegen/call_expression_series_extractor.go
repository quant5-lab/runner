package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

type callExtractor func(call *ast.CallExpression) string

func (g *generator) extractCallExpression(call *ast.CallExpression) string {
	extractors := []callExtractor{
		g.extractInputConstant,
		g.extractTempVariable,
		g.extractValueFunction,
		g.extractMathFunction,
		g.extractUserDefinedFunction,
		g.extractDefaultSeries,
	}

	for _, extract := range extractors {
		if result := extract(call); result != "" {
			return result
		}
	}
	return ""
}

func (g *generator) extractTempVariable(call *ast.CallExpression) string {
	if call == nil {
		return ""
	}
	varName := g.tempVarMgr.GetVarNameForCall(call)
	if varName == "" {
		return ""
	}
	return fmt.Sprintf("%sSeries.GetCurrent()", varName)
}

func (g *generator) extractValueFunction(call *ast.CallExpression) string {
	if call == nil || g.valueHandler == nil {
		return ""
	}
	funcName := g.extractFunctionName(call.Callee)
	if !g.valueHandler.CanHandle(funcName) {
		return ""
	}
	code, err := g.valueHandler.GenerateInlineCall(funcName, call.Arguments, g)
	if err == nil && code != "" {
		return code
	}
	return ""
}

func (g *generator) extractMathFunction(call *ast.CallExpression) string {
	if call == nil {
		return ""
	}
	funcName := g.extractFunctionName(call.Callee)
	if !g.mathHandler.CanHandle(funcName) {
		return ""
	}
	code, err := g.mathHandler.GenerateMathCall(funcName, call.Arguments, g)
	if err == nil && code != "" {
		return code
	}
	return ""
}

func (g *generator) extractUserDefinedFunction(call *ast.CallExpression) string {
	if call == nil {
		return ""
	}
	funcName := g.extractFunctionName(call.Callee)
	detector := NewUserDefinedFunctionDetector(g.variables)
	if !detector.IsUserDefinedFunction(funcName) {
		return ""
	}

	arrowCtxVar := g.arrowContextLifecycle.AllocateContextVariable(funcName)

	argsCode := []string{arrowCtxVar}
	for _, arg := range call.Arguments {
		argsCode = append(argsCode, g.extractSeriesExpression(arg))
	}

	return fmt.Sprintf("%s(%s)", funcName, strings.Join(argsCode, ", "))
}

func (g *generator) extractDefaultSeries(call *ast.CallExpression) string {
	if call == nil {
		return "Series.GetCurrent()"
	}
	funcName := g.extractFunctionName(call.Callee)
	varName := strings.ReplaceAll(funcName, ".", "_")
	return fmt.Sprintf("%sSeries.GetCurrent()", varName)
}

func (g *generator) extractInputConstant(call *ast.CallExpression) string {
	if call == nil {
		return ""
	}
	funcName := g.extractFunctionName(call.Callee)
	if g.inputConstExtractor == nil {
		return ""
	}
	return g.inputConstExtractor.ExtractInputConstant(call, funcName)
}
