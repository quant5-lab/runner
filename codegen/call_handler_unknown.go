package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// UnknownFunctionHandler handles unrecognized function calls.
//
// Behavior: Generates TODO comment for unimplemented functions
// Position: Should be last handler in chain (catch-all)
type UnknownFunctionHandler struct{}

func (h *UnknownFunctionHandler) CanHandle(funcName string) bool {
	// Catch-all: handles everything not handled by other handlers
	return true
}

func (h *UnknownFunctionHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)
	return g.ind() + fmt.Sprintf("// %s() - TODO: implement\n", funcName), nil
}
