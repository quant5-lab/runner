package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* ATRHandler generates inline code for Average True Range calculations */
type ATRHandler struct{}

func (h *ATRHandler) CanHandle(funcName string) bool {
	return funcName == "ta.atr" || funcName == "atr"
}

func (h *ATRHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 {
		return "", fmt.Errorf("ta.atr requires 1 argument (period)")
	}

	periodArg, ok := call.Arguments[0].(*ast.Literal)
	if !ok {
		return "", fmt.Errorf("ta.atr period must be literal")
	}

	period, err := extractPeriod(periodArg)
	if err != nil {
		return "", fmt.Errorf("ta.atr: %w", err)
	}

	return g.generateInlineATR(varName, period)
}
