package codegen

import "github.com/quant5-lab/runner/ast"

type LowestbarsHandler struct{}

func (h *LowestbarsHandler) CanHandle(funcName string) bool {
	return funcName == "ta.lowestbars" || funcName == "lowestbars"
}

func (h *LowestbarsHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.Extract(call, "ta.lowestbars")
	if err != nil {
		return "", err
	}
	return comp.Preamble + generateBarsToExtremum(g, varName, comp.AccessGen, comp.Period, "<"), nil
}
