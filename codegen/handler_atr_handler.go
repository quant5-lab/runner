package codegen

import "github.com/quant5-lab/runner/ast"

/* ATRHandler generates inline code for Average True Range calculations */
type ATRHandler struct{}

func (h *ATRHandler) CanHandle(funcName string) bool {
	return funcName == "ta.atr" || funcName == "atr"
}

func (h *ATRHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	period, err := extractSinglePeriodArgument(g, call, "ta.atr")
	if err != nil {
		return "", err
	}

	return g.generateInlineATR(varName, period)
}
