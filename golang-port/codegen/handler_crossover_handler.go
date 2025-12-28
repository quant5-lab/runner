package codegen

import "github.com/quant5-lab/runner/ast"

/* CrossoverHandler generates inline code for crossover detection (series1 crosses above series2) */
type CrossoverHandler struct{}

func (h *CrossoverHandler) CanHandle(funcName string) bool {
	return funcName == "ta.crossover"
}

func (h *CrossoverHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	return generateCrossDetection(g, varName, call, false)
}
