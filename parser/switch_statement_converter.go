package parser

import "github.com/quant5-lab/runner/ast"

type SwitchStatementConverter struct {
	lowering *SwitchLowering
}

func NewSwitchStatementConverter(
	orExprConverter func(*OrExpr) (ast.Expression, error),
	bodyResolver *SwitchCaseBodyResolver,
) *SwitchStatementConverter {
	return &SwitchStatementConverter{
		lowering: NewSwitchLowering(orExprConverter, bodyResolver),
	}
}

func (s *SwitchStatementConverter) CanHandle(stmt *Statement) bool {
	return stmt.Core != nil && stmt.Core.Switch != nil
}

func (s *SwitchStatementConverter) Convert(stmt *Statement) (ast.Node, error) {
	return s.lowering.Lower(stmt.Core.Switch)
}
