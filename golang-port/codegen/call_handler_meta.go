package codegen

import "github.com/quant5-lab/runner/ast"

// MetaFunctionHandler handles Pine Script meta functions.
//
// Handles: indicator(), strategy()
// Behavior: These functions define script metadata and produce no runtime code
type MetaFunctionHandler struct{}

func (h *MetaFunctionHandler) CanHandle(funcName string) bool {
	return funcName == "indicator" || funcName == "strategy"
}

func (h *MetaFunctionHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	// Meta functions produce no runtime code
	return "", nil
}
