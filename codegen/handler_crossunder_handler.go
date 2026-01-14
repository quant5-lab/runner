package codegen

import "github.com/quant5-lab/runner/ast"

/* CrossunderHandler generates inline code for crossunder detection (series1 crosses below series2) */
type CrossunderHandler struct{}

func (h *CrossunderHandler) CanHandle(funcName string) bool {
	return funcName == "ta.crossunder"
}

func (h *CrossunderHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	return generateCrossDetection(g, varName, call, true)
}
