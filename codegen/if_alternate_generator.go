package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

func (g *generator) generateIfAlternate(alternate []ast.Node) (string, error) {
	if len(alternate) == 0 {
		return g.ind() + "}\n", nil
	}

	if elseIfStmt, isChain := extractElseIfChain(alternate); isChain {
		return g.generateElseIfChain(elseIfStmt)
	}

	return g.generateElseBlock(alternate)
}

func (g *generator) generateElseIfChain(ifStmt *ast.IfStatement) (string, error) {
	condition, err := g.generateConditionExpression(ifStmt.Test)
	if err != nil {
		return "", err
	}

	condition = g.addBoolConversionIfNeeded(ifStmt.Test, condition)

	code := g.ind() + fmt.Sprintf("} else if %s {\n", condition)
	g.indent++

	bodyCode, err := g.generateIfBody(ifStmt.Consequent)
	if err != nil {
		return "", err
	}
	code += bodyCode

	g.indent--

	alternateCode, err := g.generateIfAlternate(ifStmt.Alternate)
	if err != nil {
		return "", err
	}

	return code + alternateCode, nil
}

func (g *generator) generateElseBlock(nodes []ast.Node) (string, error) {
	code := g.ind() + "} else {\n"
	g.indent++

	for _, stmt := range nodes {
		stmtCode, err := g.generateStatement(stmt)
		if err != nil {
			return "", err
		}
		if stmtCode != "" {
			code += stmtCode
		}
	}

	g.indent--
	code += g.ind() + "}\n"
	return code, nil
}

func (g *generator) generateIfBody(body []ast.Node) (string, error) {
	var code string
	hasValidBody := false

	for _, stmt := range body {
		if shouldSkipIfBodyStatement(stmt) {
			continue
		}

		stmtCode, err := g.generateStatement(stmt)
		if err != nil {
			return "", err
		}
		if stmtCode != "" {
			code += stmtCode
			hasValidBody = true
		}
	}

	if !hasValidBody {
		code += g.ind() + "// TODO: if body statements\n"
	}

	return code, nil
}

func shouldSkipIfBodyStatement(stmt ast.Node) bool {
	exprStmt, ok := stmt.(*ast.ExpressionStatement)
	if !ok {
		return false
	}

	switch exprStmt.Expression.(type) {
	case *ast.CallExpression:
		return false
	case *ast.Identifier, *ast.Literal:
		return true
	case *ast.BinaryExpression, *ast.LogicalExpression, *ast.ConditionalExpression:
		return true
	}

	return false
}

func extractElseIfChain(alternate []ast.Node) (*ast.IfStatement, bool) {
	if len(alternate) != 1 {
		return nil, false
	}
	ifStmt, ok := alternate[0].(*ast.IfStatement)
	return ifStmt, ok
}
