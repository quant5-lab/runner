package parser

import "github.com/quant5-lab/runner/ast"

type BreakStatementConverter struct{}

func NewBreakStatementConverter() *BreakStatementConverter {
	return &BreakStatementConverter{}
}

func (b *BreakStatementConverter) CanHandle(stmt *Statement) bool {
	return stmt.Core != nil && stmt.Core.Break != nil
}

func (b *BreakStatementConverter) Convert(stmt *Statement) (ast.Node, error) {
	return &ast.BreakStatement{NodeType: ast.TypeBreakStatement}, nil
}

type ContinueStatementConverter struct{}

func NewContinueStatementConverter() *ContinueStatementConverter {
	return &ContinueStatementConverter{}
}

func (c *ContinueStatementConverter) CanHandle(stmt *Statement) bool {
	return stmt.Core != nil && stmt.Core.Continue != nil
}

func (c *ContinueStatementConverter) Convert(stmt *Statement) (ast.Node, error) {
	return &ast.ContinueStatement{NodeType: ast.TypeContinueStatement}, nil
}
