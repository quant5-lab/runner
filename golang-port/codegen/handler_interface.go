package codegen

import "github.com/quant5-lab/runner/ast"

/* TAHandler interface for indicator-specific code generation */
type TAHandler interface {
	CanHandle(funcName string) bool
	GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error)
}
