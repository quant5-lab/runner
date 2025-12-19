package parser

import "github.com/quant5-lab/runner/ast"

type FunctionDeclarationConverter struct {
	statementConverter func(*Statement) (ast.Node, error)
}

func NewFunctionDeclarationConverter(statementConverter func(*Statement) (ast.Node, error)) *FunctionDeclarationConverter {
	return &FunctionDeclarationConverter{
		statementConverter: statementConverter,
	}
}

func (f *FunctionDeclarationConverter) CanHandle(stmt *Statement) bool {
	return stmt.FunctionDecl != nil
}

func (f *FunctionDeclarationConverter) Convert(stmt *Statement) (ast.Node, error) {
	funcDecl := stmt.FunctionDecl

	params := buildIdentifiers(funcDecl.Params)
	body, err := f.convertBody(funcDecl.Body)
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

func (f *FunctionDeclarationConverter) convertBody(body []*Statement) ([]ast.Node, error) {
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
