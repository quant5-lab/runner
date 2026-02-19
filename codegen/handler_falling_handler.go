package codegen

import "github.com/quant5-lab/runner/ast"

type FallingHandler struct{}

func (h *FallingHandler) CanHandle(funcName string) bool {
	return funcName == "ta.falling" || funcName == "falling"
}

func (h *FallingHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.Extract(call, "ta.falling")
	if err != nil {
		return "", err
	}
	return comp.Preamble + generateMonotoneSequence(g, varName, comp.AccessGen, comp.Period, ">="), nil
}
