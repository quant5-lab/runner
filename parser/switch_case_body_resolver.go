package parser

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type SwitchCaseBodyResolver struct {
	statementConverter  func(*Statement) (ast.Node, error)
	expressionConverter func(*Expression) (ast.Expression, error)
}

func NewSwitchCaseBodyResolver(
	statementConverter func(*Statement) (ast.Node, error),
	expressionConverter func(*Expression) (ast.Expression, error),
) *SwitchCaseBodyResolver {
	return &SwitchCaseBodyResolver{
		statementConverter:  statementConverter,
		expressionConverter: expressionConverter,
	}
}

func (r *SwitchCaseBodyResolver) Resolve(switchCase *SwitchCase) ([]ast.Node, error) {
	if switchCase.InlineBody != nil {
		return r.resolveInlineBody(switchCase.InlineBody)
	}
	return r.resolveMultiLineBody(switchCase.Body)
}

func (r *SwitchCaseBodyResolver) resolveInlineBody(expr *Expression) ([]ast.Node, error) {
	converted, err := r.expressionConverter(expr)
	if err != nil {
		return nil, fmt.Errorf("resolving inline switch case body: %w", err)
	}
	return []ast.Node{
		&ast.ExpressionStatement{
			NodeType:   ast.TypeExpressionStatement,
			Expression: converted,
		},
	}, nil
}

func (r *SwitchCaseBodyResolver) resolveMultiLineBody(statements []*Statement) ([]ast.Node, error) {
	var nodes []ast.Node
	for _, stmt := range statements {
		node, err := r.statementConverter(stmt)
		if err != nil {
			return nil, fmt.Errorf("resolving switch case body statement: %w", err)
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}
