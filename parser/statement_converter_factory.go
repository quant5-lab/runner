package parser

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type StatementConverterFactory struct {
	converters []StatementConverter
}

func NewStatementConverterFactory(
	expressionConverter func(*Expression) (ast.Expression, error),
	orExprConverter func(*OrExpr) (ast.Expression, error),
	arithExprConverter func(*ArithExpr) (ast.Expression, error),
	statementConverter func(*Statement) (ast.Node, error),
) *StatementConverterFactory {
	return &StatementConverterFactory{
		converters: []StatementConverter{
			NewTupleAssignmentConverter(expressionConverter),
			NewFunctionDeclarationConverter(statementConverter, expressionConverter),
			NewTypedAssignmentConverter(expressionConverter),
			NewAssignmentConverter(expressionConverter),
			NewReassignmentConverter(expressionConverter),
			NewIfStatementConverter(orExprConverter, statementConverter),
			NewForStatementConverter(arithExprConverter, statementConverter),
			NewExpressionStatementConverter(expressionConverter),
		},
	}
}

func (f *StatementConverterFactory) Convert(stmt *Statement) (ast.Node, error) {
	for _, converter := range f.converters {
		if converter.CanHandle(stmt) {
			return converter.Convert(stmt)
		}
	}
	return nil, fmt.Errorf("empty statement")
}
