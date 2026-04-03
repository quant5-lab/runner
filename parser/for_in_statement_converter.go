package parser

import "github.com/quant5-lab/runner/ast"

type ForInStatementConverter struct {
	arithExprConverter func(*ArithExpr) (ast.Expression, error)
	statementConverter func(*Statement) (ast.Node, error)
}

func NewForInStatementConverter(
	arithExprConverter func(*ArithExpr) (ast.Expression, error),
	statementConverter func(*Statement) (ast.Node, error),
) *ForInStatementConverter {
	return &ForInStatementConverter{
		arithExprConverter: arithExprConverter,
		statementConverter: statementConverter,
	}
}

func (f *ForInStatementConverter) CanHandle(stmt *Statement) bool {
	return stmt.Core != nil && stmt.Core.ForIn != nil
}

func (f *ForInStatementConverter) Convert(stmt *Statement) (ast.Node, error) {
	forIn := stmt.Core.ForIn
	return convertForInToAST(forIn.Vars, forIn.Collection, forIn.Body, f.arithExprConverter, f.statementConverter)
}

/* Shared conversion logic for both statement and expression contexts */
func convertForInToAST(
	vars *ForInVars,
	collection *ArithExpr,
	body []*Statement,
	arithExprConverter func(*ArithExpr) (ast.Expression, error),
	statementConverter func(*Statement) (ast.Node, error),
) (*ast.ForInStatement, error) {
	collExpr, err := arithExprConverter(collection)
	if err != nil {
		return nil, err
	}

	bodyNodes := []ast.Node{}
	for _, bodyStmt := range body {
		node, err := statementConverter(bodyStmt)
		if err != nil {
			return nil, err
		}
		if node != nil {
			bodyNodes = append(bodyNodes, node)
		}
	}

	indexVar := ""
	elementVar := ""
	if vars.TupleIndex != nil {
		indexVar = *vars.TupleIndex
		elementVar = *vars.TupleElement
	} else {
		elementVar = *vars.SingleElement
	}

	return &ast.ForInStatement{
		NodeType:   ast.TypeForInStatement,
		IndexVar:   indexVar,
		ElementVar: elementVar,
		Collection: collExpr,
		Body:       bodyNodes,
	}, nil
}
