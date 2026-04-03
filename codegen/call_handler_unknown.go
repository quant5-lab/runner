package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* UnknownFunctionHandler handles unrecognized function calls.
 * Single responsibility: Generate TODO comment for unimplemented functions.
 * Position: Last handler in chain (catch-all).
 */
type UnknownFunctionHandler struct{}

func (h *UnknownFunctionHandler) CanHandle(funcName string) bool {
	return true
}

func (h *UnknownFunctionHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)
	return g.ind() + fmt.Sprintf("// %s() - TODO: implement\n", funcName), nil
}
