package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type ATRHandler struct{}

func (h *ATRHandler) CanHandle(funcName string) bool {
	return funcName == "ta.atr" || funcName == "atr"
}

func (h *ATRHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 {
		return "", fmt.Errorf("ta.atr requires period argument")
	}

	periodResult := evaluatePeriodExpression(g, call.Arguments[0])

	if periodResult.IsFailed() {
		return "", fmt.Errorf("ta.atr: %s", periodResult.FailureReason)
	}

	if periodResult.IsRuntimeDynamic() {
		dynamicGen := NewDynamicPeriodTAGenerator(g)
		code, err := dynamicGen.Generate(varName, "ta.atr", nil, periodResult)
		if err != nil {
			return "", err
		}
		return g.indentCode(code), nil
	}

	return g.generateInlineATR(varName, periodResult.StaticValue)
}
