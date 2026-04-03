package codegen

import "github.com/quant5-lab/runner/ast"

/* PivotHighHandler generates inline code for pivot high detection */
type PivotHighHandler struct{}

func (h *PivotHighHandler) CanHandle(funcName string) bool {
	return funcName == "ta.pivothigh" || funcName == "pivothigh"
}

func (h *PivotHighHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	return g.generatePivot(varName, call, true)
}
