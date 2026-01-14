package codegen

import "github.com/quant5-lab/runner/ast"

/* EMAHandler generates inline code for Exponential Moving Average calculations */
type EMAHandler struct{}

func (h *EMAHandler) CanHandle(funcName string) bool {
	return funcName == "ta.ema" || funcName == "ema"
}

func (h *EMAHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.Extract(call, "ta.ema")
	if err != nil {
		return "", err
	}

	builder := NewTAIndicatorBuilder("ta.ema", varName, comp.Period, comp.AccessGen, comp.NeedsNaNCheck)
	return g.indentCode(comp.Preamble + builder.BuildEMA()), nil
}
