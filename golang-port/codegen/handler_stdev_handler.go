package codegen

import "github.com/quant5-lab/runner/ast"

/* STDEVHandler generates inline code for Standard Deviation calculations */
type STDEVHandler struct{}

func (h *STDEVHandler) CanHandle(funcName string) bool {
	return funcName == "ta.stdev" || funcName == "stdev"
}

func (h *STDEVHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.Extract(call, "ta.stdev")
	if err != nil {
		return "", err
	}

	builder := NewTAIndicatorBuilder("ta.stdev", varName, comp.Period, comp.AccessGen, comp.NeedsNaNCheck)
	return g.indentCode(comp.Preamble + builder.BuildSTDEV()), nil
}
