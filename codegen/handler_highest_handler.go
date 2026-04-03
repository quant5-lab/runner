package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type HighestHandler struct{}

func (h *HighestHandler) CanHandle(funcName string) bool {
	return funcName == "ta.highest" || funcName == "highest"
}

func (h *HighestHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	var sourceExpr ast.Expression
	var periodResult PeriodEvaluationResult
	var accessGen AccessGenerator

	if len(call.Arguments) == 1 {
		periodArg := call.Arguments[0]
		periodResult = evaluatePeriodExpression(g, periodArg)
		sourceExpr = &ast.Identifier{Name: "high"}
		classifier := NewSeriesSourceClassifier()
		highInfo := classifier.ClassifyAST(sourceExpr)
		accessGen = CreateAccessGenerator(highInfo)
	} else if len(call.Arguments) >= 2 {
		sourceExpr = call.Arguments[0]
		periodArg := call.Arguments[1]
		periodResult = evaluatePeriodExpression(g, periodArg)
		classifier := NewSeriesSourceClassifier()
		sourceInfo := classifier.ClassifyAST(sourceExpr)
		accessGen = CreateAccessGenerator(sourceInfo)
	} else {
		return "", fmt.Errorf("ta.highest requires 1 or 2 arguments")
	}

	if periodResult.IsFailed() {
		return "", fmt.Errorf("ta.highest: %s", periodResult.FailureReason)
	}

	if periodResult.IsRuntimeDynamic() {
		dynamicGen := NewDynamicPeriodTAGenerator(g)
		code, err := dynamicGen.Generate(varName, "ta.highest", sourceExpr, periodResult)
		if err != nil {
			return "", err
		}
		return g.indentCode(code), nil
	}

	registry := NewInlineTAIIFERegistry()
	hasher := &ExpressionHasher{}
	sourceHash := ""
	if len(call.Arguments) > 0 {
		sourceHash = hasher.Hash(call.Arguments[0])
	}
	iifeCode, ok := registry.Generate("ta.highest", accessGen, NewConstantPeriod(periodResult.StaticValue), sourceHash)
	if !ok {
		return "", fmt.Errorf("ta.highest IIFE generation failed")
	}

	return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, iifeCode), nil
}
