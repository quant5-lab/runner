package codegen

import "github.com/quant5-lab/runner/ast"

/* PivotLowHandler generates inline code for pivot low detection */
type PivotLowHandler struct{}

func (h *PivotLowHandler) CanHandle(funcName string) bool {
	return funcName == "ta.pivotlow"
}

func (h *PivotLowHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	return g.generatePivot(varName, call, false)
}
