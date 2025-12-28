package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* SumHandler generates inline code for sum calculations */
type SumHandler struct{}

func (h *SumHandler) CanHandle(funcName string) bool {
	return funcName == "sum" || funcName == "math.sum"
}

func (h *SumHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 2 {
		return "", fmt.Errorf("sum requires 2 arguments")
	}

	var code string
	sourceArg := call.Arguments[0]
	var sourceInfo SourceInfo
	var period int

	if condExpr, ok := sourceArg.(*ast.ConditionalExpression); ok {
		tempVarName := g.tempVarMgr.GetOrCreate(CallInfo{
			FuncName: "ternary",
			Call:     call,
			ArgHash:  fmt.Sprintf("%p", condExpr),
		})

		condCode, err := g.generateConditionExpression(condExpr.Test)
		if err != nil {
			return "", err
		}
		condCode = g.addBoolConversionIfNeeded(condExpr.Test, condCode)

		consequentCode, err := g.generateNumericExpression(condExpr.Consequent)
		if err != nil {
			return "", err
		}
		alternateCode, err := g.generateNumericExpression(condExpr.Alternate)
		if err != nil {
			return "", err
		}

		code += g.ind() + fmt.Sprintf("%sSeries.Set(func() float64 { if %s { return %s } else { return %s } }())\n",
			tempVarName, condCode, consequentCode, alternateCode)

		sourceInfo = SourceInfo{
			Type:         SourceTypeSeriesVariable,
			VariableName: tempVarName,
		}

		extractor := NewTAArgumentExtractor(g)
		extractedPeriod, err := extractor.extractPeriod(call.Arguments[1], "sum")
		if err != nil {
			return "", err
		}
		period = extractedPeriod
	} else {
		extractor := NewTAArgumentExtractor(g)
		comp, err := extractor.Extract(call, "sum")
		if err != nil {
			return "", err
		}
		sourceInfo = comp.SourceInfo
		period = comp.Period
	}

	accessGen := CreateAccessGenerator(sourceInfo)
	needsNaN := sourceInfo.IsSeriesVariable()

	builder := NewTAIndicatorBuilder("sum", varName, period, accessGen, needsNaN)
	builder.WithAccumulator(NewSumAccumulator())
	sumCode := g.indentCode(builder.Build())

	return code + sumCode, nil
}
