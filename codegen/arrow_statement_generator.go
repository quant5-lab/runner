package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* ArrowStatementGenerator generates variable declarations with dual scalar+series pattern */
type ArrowStatementGenerator struct {
	gen           *generator
	localStorage  *ArrowLocalVariableStorage
	exprGenerator *ArrowExpressionGeneratorImpl
	symbolTable   SymbolTable
}

func NewArrowStatementGenerator(
	gen *generator,
	localStorage *ArrowLocalVariableStorage,
	exprGen *ArrowExpressionGeneratorImpl,
	symbolTable SymbolTable,
) *ArrowStatementGenerator {
	return &ArrowStatementGenerator{
		gen:           gen,
		localStorage:  localStorage,
		exprGenerator: exprGen,
		symbolTable:   symbolTable,
	}
}

/* GenerateStatement generates arrow-aware statement code with Series.Set() for variables */
func (s *ArrowStatementGenerator) GenerateStatement(stmt ast.Node) (string, error) {
	switch st := stmt.(type) {
	case *ast.VariableDeclaration:
		return s.generateVariableDeclaration(st)

	default:
		return s.gen.generateStatement(stmt)
	}
}

func (s *ArrowStatementGenerator) generateVariableDeclaration(varDecl *ast.VariableDeclaration) (string, error) {
	if len(varDecl.Declarations) == 0 {
		return "", fmt.Errorf("empty variable declaration")
	}

	decl := varDecl.Declarations[0]

	if arrayPattern, ok := decl.ID.(*ast.ArrayPattern); ok {
		return s.generateTupleDeclaration(arrayPattern, decl.Init)
	}

	if id, ok := decl.ID.(*ast.Identifier); ok {
		return s.generateSingleVariableDeclaration(id.Name, decl.Init)
	}

	return "", fmt.Errorf("unsupported variable declarator pattern: %T", decl.ID)
}

func (s *ArrowStatementGenerator) generateSingleVariableDeclaration(varName string, initExpr ast.Expression) (string, error) {
	// Register in symbol table as series (arrow function variables are always series)
	if s.symbolTable != nil {
		s.symbolTable.Register(varName, VariableTypeSeries)
	}

	exprCode, err := s.exprGenerator.Generate(initExpr)
	if err != nil {
		return "", fmt.Errorf("failed to generate init expression for '%s': %w", varName, err)
	}

	return s.localStorage.GenerateDualStorage(varName, exprCode), nil
}

func (s *ArrowStatementGenerator) generateTupleDeclaration(arrayPattern *ast.ArrayPattern, initExpr ast.Expression) (string, error) {
	varNames := make([]string, len(arrayPattern.Elements))
	for i, elem := range arrayPattern.Elements {
		varNames[i] = elem.Name
		// Register each tuple element in symbol table as series
		if s.symbolTable != nil {
			s.symbolTable.Register(varNames[i], VariableTypeSeries)
		}
	}

	exprCode, err := s.exprGenerator.Generate(initExpr)
	if err != nil {
		return "", fmt.Errorf("failed to generate tuple init expression: %w", err)
	}

	return s.localStorage.GenerateTupleDualStorage(varNames, exprCode), nil
}
