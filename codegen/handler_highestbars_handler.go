package codegen

import "github.com/quant5-lab/runner/ast"

type HighestbarsHandler struct{}

func (h *HighestbarsHandler) CanHandle(funcName string) bool {
	return funcName == "ta.highestbars" || funcName == "highestbars"
}

func (h *HighestbarsHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.Extract(call, "ta.highestbars")
	if err != nil {
		return "", err
	}
	return comp.Preamble + generateBarsToExtremum(g, varName, comp.AccessGen, comp.Period, ">"), nil
}
