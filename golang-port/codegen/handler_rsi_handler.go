package codegen

import "github.com/quant5-lab/runner/ast"

/* RSIHandler generates inline code for Relative Strength Index calculations */
type RSIHandler struct{}

func (h *RSIHandler) CanHandle(funcName string) bool {
	return funcName == "ta.rsi" || funcName == "rsi"
}

func (h *RSIHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.Extract(call, "ta.rsi")
	if err != nil {
		return "", err
	}

	code, err := g.generateRSI(varName, comp.Period, comp.AccessGen, comp.NeedsNaNCheck)
	if err != nil {
		return "", err
	}
	return comp.Preamble + code, nil
}
