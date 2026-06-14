package codegen

import (
	"strings"

	"github.com/quant5-lab/runner/ast"
)

var chartOnlyNamespaces = []string{"label", "line", "box", "table", "linefill"}

// VoidNamespaceCallHandler stubs Pine drawing-object calls with math.NaN() so
// assignments compile in any expression position. Chart-object IDs never reach
// strategy logic, making NaN the correct numeric placeholder throughout.
type VoidNamespaceCallHandler struct{}

func (h *VoidNamespaceCallHandler) CanHandle(funcName string) bool {
	for _, ns := range chartOnlyNamespaces {
		if funcName == ns || strings.HasPrefix(funcName, ns+".") {
			return true
		}
	}
	return false
}

func (h *VoidNamespaceCallHandler) GenerateCode(_ *generator, call *ast.CallExpression) (string, error) {
	if !h.CanHandle(extractCallFunctionName(call)) {
		return "", nil
	}
	return "math.NaN()", nil
}
