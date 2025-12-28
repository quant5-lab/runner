package codegen

import "github.com/quant5-lab/runner/ast"

/* WMAHandler generates inline code for Weighted Moving Average calculations */
type WMAHandler struct{}

func (h *WMAHandler) CanHandle(funcName string) bool {
	return funcName == "ta.wma" || funcName == "wma"
}

func (h *WMAHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.Extract(call, "ta.wma")
	if err != nil {
		return "", err
	}

	builder := NewTAIndicatorBuilder("ta.wma", varName, comp.Period, comp.AccessGen, comp.NeedsNaNCheck)
	builder.WithAccumulator(NewWeightedSumAccumulator(comp.Period))
	return g.indentCode(builder.Build()), nil
}
