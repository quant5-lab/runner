package codegen

import "github.com/quant5-lab/runner/ast"

// VoidBuiltinHandler absorbs Pine built-ins that are pure notification sinks
// and have no influence on strategy logic or backtesting results.
type VoidBuiltinHandler struct {
	names map[string]struct{}
}

// NewVoidBuiltinHandler seeds the handler with every known void built-in.
// Add a name here when a new such built-in is encountered; no other file needs to change.
func NewVoidBuiltinHandler() *VoidBuiltinHandler {
	names := []string{
		"alert",
		"alertcondition",
	}
	set := make(map[string]struct{}, len(names))
	for _, n := range names {
		set[n] = struct{}{}
	}
	return &VoidBuiltinHandler{names: set}
}

func (h *VoidBuiltinHandler) CanHandle(funcName string) bool {
	_, ok := h.names[funcName]
	return ok
}

func (h *VoidBuiltinHandler) GenerateCode(_ *generator, _ *ast.CallExpression) (string, error) {
	return "", nil
}
