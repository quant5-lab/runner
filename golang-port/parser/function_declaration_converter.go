package parser

import "github.com/quant5-lab/runner/ast"

type FunctionDeclarationConverter struct {
	statementConverter  func(*Statement) (ast.Node, error)
	expressionConverter func(*Expression) (ast.Expression, error)
}

func NewFunctionDeclarationConverter(
	statementConverter func(*Statement) (ast.Node, error),
	expressionConverter func(*Expression) (ast.Expression, error),
) *FunctionDeclarationConverter {
	return &FunctionDeclarationConverter{
		statementConverter:  statementConverter,
		expressionConverter: expressionConverter,
	}
}

func (f *FunctionDeclarationConverter) CanHandle(stmt *Statement) bool {
	return stmt.FunctionDecl != nil
}

func (f *FunctionDeclarationConverter) Convert(stmt *Statement) (ast.Node, error) {
	funcDecl := stmt.FunctionDecl

	params := buildIdentifiers(funcDecl.Params)
	body, err := f.convertFunctionBody(funcDecl)
	if err != nil {
		return nil, err
	}

	arrowFunc := &ast.ArrowFunctionExpression{
		NodeType: ast.TypeArrowFunctionExpression,
		Params:   params,
		Body:     body,
	}

	return buildVariableDeclaration(
		buildIdentifier(funcDecl.Name),
		arrowFunc,
		"let",
	), nil
}

func (f *FunctionDeclarationConverter) convertFunctionBody(funcDecl *FunctionDecl) ([]ast.Node, error) {
	if funcDecl.InlineBody != nil {
		return f.convertInlineBody(funcDecl.InlineBody)
	}
	return f.convertMultiLineBody(funcDecl.MultiLineBody)
}

func (f *FunctionDeclarationConverter) convertInlineBody(expr *Expression) ([]ast.Node, error) {
	node, err := f.expressionConverter(expr)
	if err != nil {
		return nil, err
	}
	returnStmt := &ast.ExpressionStatement{
		NodeType:   ast.TypeExpressionStatement,
		Expression: node,
	}
	return []ast.Node{returnStmt}, nil
}

func (f *FunctionDeclarationConverter) convertMultiLineBody(body []*Statement) ([]ast.Node, error) {
	nodes := []ast.Node{}
	for _, stmt := range body {
		node, err := f.statementConverter(stmt)
		if err != nil {
			return nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}
