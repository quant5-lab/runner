package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

var defaultStatefulValueFunctions = map[string]bool{
	"fixnan": true,
}

type HoistableCallClassifier struct {
	gen                    *generator
	inlineTARegistry       *InlineTAIIFERegistry
	taFunctionRegistry     *TAFunctionRegistry
	statefulValueFunctions map[string]bool
}

func NewHoistableCallClassifier(g *generator) HoistableCallClassifier {
	return HoistableCallClassifier{
		gen:                    g,
		inlineTARegistry:       NewInlineTAIIFERegistry(),
		taFunctionRegistry:     g.taRegistry,
		statefulValueFunctions: defaultStatefulValueFunctions,
	}
}

func (c HoistableCallClassifier) IsHoistable(call *ast.CallExpression) bool {
	funcName := c.gen.extractFunctionName(call.Callee)

	if IsSecurityFunction(funcName) {
		return true
	}

	if c.IsStatefulValueFunction(funcName) {
		return true
	}

	if c.isInlineTAFunction(funcName) {
		return c.shouldHoistTACall(call, funcName)
	}

	if c.hasImplementedTAHandler(funcName) {
		return true
	}

	if c.isSignatureRegisteredTA(funcName) {
		return true
	}

	return false
}

func (c HoistableCallClassifier) isSignatureRegisteredTA(funcName string) bool {
	return sharedTASignatures.Contains(funcName)
}

func (c HoistableCallClassifier) IsStatefulValueFunction(funcName string) bool {
	return c.statefulValueFunctions[funcName]
}

func (c HoistableCallClassifier) isInlineTAFunction(funcName string) bool {
	return c.inlineTARegistry.IsSupported(funcName)
}

func (c HoistableCallClassifier) hasImplementedTAHandler(funcName string) bool {
	if c.taFunctionRegistry == nil {
		return false
	}
	return c.taFunctionRegistry.IsSupported(funcName)
}

func (c HoistableCallClassifier) shouldHoistTACall(call *ast.CallExpression, funcName string) bool {
	periodResult := c.extractPeriod(call, funcName)
	if periodResult.IsFailed() {
		return false
	}

	if periodResult.IsRuntimeDynamic() {
		return SupportsDynamicPeriod(funcName)
	}

	return true
}

func (c HoistableCallClassifier) extractPeriod(call *ast.CallExpression, funcName string) PeriodEvaluationResult {
	if funcName == "ta.atr" || funcName == "atr" {
		return extractSinglePeriodWithDynamic(c.gen, call, funcName)
	}

	if len(call.Arguments) < 2 {
		return NewFailedPeriodEvaluation("insufficient arguments for period extraction")
	}

	_, periodResult := extractTAArgumentsWithDynamic(c.gen, call, funcName)
	return periodResult
}
