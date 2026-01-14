package codegen

import "github.com/quant5-lab/runner/ast"

/* SMAHandler generates inline code for Simple Moving Average calculations */
type SMAHandler struct{}

func (h *SMAHandler) CanHandle(funcName string) bool {
	return funcName == "ta.sma" || funcName == "sma"
}

func (h *SMAHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.Extract(call, "ta.sma")
	if err != nil {
		return "", err
	}

	builder := NewTAIndicatorBuilder("ta.sma", varName, comp.Period, comp.AccessGen, comp.NeedsNaNCheck)
	builder.WithAccumulator(NewSumAccumulator())
	return g.indentCode(comp.Preamble + builder.Build()), nil
}
