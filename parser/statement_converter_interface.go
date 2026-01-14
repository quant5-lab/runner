package parser

import "github.com/quant5-lab/runner/ast"

type StatementConverter interface {
	Convert(stmt *Statement) (ast.Node, error)
	CanHandle(stmt *Statement) bool
}
