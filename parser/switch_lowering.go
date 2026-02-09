package parser

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type SwitchLowering struct {
	orExprConverter    func(*OrExpr) (ast.Expression, error)
	statementConverter func(*Statement) (ast.Node, error)
}

func NewSwitchLowering(
	orExprConverter func(*OrExpr) (ast.Expression, error),
	statementConverter func(*Statement) (ast.Node, error),
) *SwitchLowering {
	return &SwitchLowering{
		orExprConverter:    orExprConverter,
		statementConverter: statementConverter,
	}
}

func (l *SwitchLowering) Lower(switchExpr *SwitchExpr) (*ast.IfStatement, error) {
	subject, err := l.resolveSubject(switchExpr.Subject)
	if err != nil {
		return nil, fmt.Errorf("lowering switch subject: %w", err)
	}

	regularCases, defaultCase := l.partitionCases(switchExpr.Cases)

	if len(regularCases) == 0 && defaultCase == nil {
		return nil, fmt.Errorf("switch expression has no cases")
	}

	if len(regularCases) == 0 {
		return l.lowerDefaultOnlySwitch(defaultCase)
	}

	return l.buildIfChain(subject, regularCases, defaultCase)
}

func (l *SwitchLowering) resolveSubject(subjectExpr *OrExpr) (ast.Expression, error) {
	if subjectExpr == nil {
		return nil, nil
	}
	return l.orExprConverter(subjectExpr)
}

func (l *SwitchLowering) partitionCases(cases []*SwitchCase) ([]*SwitchCase, *SwitchCase) {
	var regular []*SwitchCase
	var defaultCase *SwitchCase

	for _, c := range cases {
		if c.Condition == nil {
			defaultCase = c
		} else {
			regular = append(regular, c)
		}
	}

	return regular, defaultCase
}

func (l *SwitchLowering) lowerDefaultOnlySwitch(defaultCase *SwitchCase) (*ast.IfStatement, error) {
	body, err := l.convertBody(defaultCase.Body)
	if err != nil {
		return nil, err
	}

	return &ast.IfStatement{
		NodeType: ast.TypeIfStatement,
		Test: &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    true,
			Raw:      "true",
		},
		Consequent: body,
		Alternate:  []ast.Node{},
	}, nil
}

func (l *SwitchLowering) buildIfChain(subject ast.Expression, cases []*SwitchCase, defaultCase *SwitchCase) (*ast.IfStatement, error) {
	defaultAlternate, err := l.resolveDefaultAlternate(defaultCase)
	if err != nil {
		return nil, err
	}

	current, err := l.buildTerminalCase(subject, cases[len(cases)-1], defaultAlternate)
	if err != nil {
		return nil, err
	}

	for i := len(cases) - 2; i >= 0; i-- {
		current, err = l.wrapWithCase(subject, cases[i], current)
		if err != nil {
			return nil, err
		}
	}

	return current, nil
}

func (l *SwitchLowering) buildTerminalCase(subject ast.Expression, switchCase *SwitchCase, alternate []ast.Node) (*ast.IfStatement, error) {
	test, err := l.buildTestExpression(subject, switchCase.Condition)
	if err != nil {
		return nil, err
	}

	consequent, err := l.convertBody(switchCase.Body)
	if err != nil {
		return nil, err
	}

	return &ast.IfStatement{
		NodeType:   ast.TypeIfStatement,
		Test:       test,
		Consequent: consequent,
		Alternate:  alternate,
	}, nil
}

func (l *SwitchLowering) wrapWithCase(subject ast.Expression, switchCase *SwitchCase, nested *ast.IfStatement) (*ast.IfStatement, error) {
	test, err := l.buildTestExpression(subject, switchCase.Condition)
	if err != nil {
		return nil, err
	}

	consequent, err := l.convertBody(switchCase.Body)
	if err != nil {
		return nil, err
	}

	return &ast.IfStatement{
		NodeType:   ast.TypeIfStatement,
		Test:       test,
		Consequent: consequent,
		Alternate:  []ast.Node{nested},
	}, nil
}

func (l *SwitchLowering) buildTestExpression(subject ast.Expression, condition *OrExpr) (ast.Expression, error) {
	condExpr, err := l.orExprConverter(condition)
	if err != nil {
		return nil, fmt.Errorf("lowering switch case condition: %w", err)
	}

	if subject == nil {
		return condExpr, nil
	}

	return &ast.BinaryExpression{
		NodeType: ast.TypeBinaryExpression,
		Operator: "==",
		Left:     subject,
		Right:    condExpr,
	}, nil
}

func (l *SwitchLowering) resolveDefaultAlternate(defaultCase *SwitchCase) ([]ast.Node, error) {
	if defaultCase == nil {
		return []ast.Node{}, nil
	}
	return l.convertBody(defaultCase.Body)
}

func (l *SwitchLowering) convertBody(body []*Statement) ([]ast.Node, error) {
	nodes := []ast.Node{}
	for _, stmt := range body {
		node, err := l.statementConverter(stmt)
		if err != nil {
			return nil, fmt.Errorf("lowering switch case body: %w", err)
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}
