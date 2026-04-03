package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type LowestHandler struct{}

func (h *LowestHandler) CanHandle(funcName string) bool {
	return funcName == "ta.lowest" || funcName == "lowest"
}

func (h *LowestHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	var sourceExpr ast.Expression
	var periodResult PeriodEvaluationResult
	var accessGen AccessGenerator

	if len(call.Arguments) == 1 {
		periodArg := call.Arguments[0]
		periodResult = evaluatePeriodExpression(g, periodArg)
		sourceExpr = &ast.Identifier{Name: "low"}
		classifier := NewSeriesSourceClassifier()
		lowInfo := classifier.ClassifyAST(sourceExpr)
		accessGen = CreateAccessGenerator(lowInfo)
	} else if len(call.Arguments) >= 2 {
		sourceExpr = call.Arguments[0]
		periodArg := call.Arguments[1]
		periodResult = evaluatePeriodExpression(g, periodArg)
		classifier := NewSeriesSourceClassifier()
		sourceInfo := classifier.ClassifyAST(sourceExpr)
		accessGen = CreateAccessGenerator(sourceInfo)
	} else {
		return "", fmt.Errorf("ta.lowest requires 1 or 2 arguments")
	}

	if periodResult.IsFailed() {
		return "", fmt.Errorf("ta.lowest: %s", periodResult.FailureReason)
	}

	if periodResult.IsRuntimeDynamic() {
		dynamicGen := NewDynamicPeriodTAGenerator(g)
		code, err := dynamicGen.Generate(varName, "ta.lowest", sourceExpr, periodResult)
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
	iifeCode, ok := registry.Generate("ta.lowest", accessGen, NewConstantPeriod(periodResult.StaticValue), sourceHash)
	if !ok {
		return "", fmt.Errorf("ta.lowest IIFE generation failed")
	}

	return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, iifeCode), nil
}
