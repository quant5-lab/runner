package codegen

import "github.com/quant5-lab/runner/ast"

/* RMAHandler generates inline code for RMA (Relative Moving Average) calculations */
type RMAHandler struct{}

func (h *RMAHandler) CanHandle(funcName string) bool {
	return funcName == "ta.rma" || funcName == "rma"
}

func (h *RMAHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.Extract(call, "ta.rma")
	if err != nil {
		return "", err
	}

	code, err := g.generateRMA(varName, comp.Period, comp.AccessGen, comp.NeedsNaNCheck)
	if err != nil {
		return "", err
	}
	return comp.Preamble + code, nil
}
