package codegen

import "github.com/quant5-lab/runner/ast"

type RisingHandler struct{}

func (h *RisingHandler) CanHandle(funcName string) bool {
	return funcName == "ta.rising" || funcName == "rising"
}

func (h *RisingHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.Extract(call, "ta.rising")
	if err != nil {
		return "", err
	}
	return comp.Preamble + generateMonotoneSequence(g, varName, comp.AccessGen, comp.Period, "<="), nil
}
