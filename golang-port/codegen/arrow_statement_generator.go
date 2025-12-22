package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/*
ArrowStatementGenerator generates statements with arrow-context awareness.

Responsibility (SRP):
  - Single purpose: generate variable declaration statements in arrow functions
  - Uses ArrowSeriesVariableGenerator for Series.Set() generation
  - Delegates expression generation to ArrowExpressionGeneratorImpl
  - No knowledge of function-level code organization

Design:
  - Composition: uses expression generator for RHS evaluation
  - DRY: reuses existing Series variable generator
  - KISS: simple delegation, minimal logic
*/
type ArrowStatementGenerator struct {
	gen           *generator
	seriesVarGen  *ArrowSeriesVariableGenerator
	exprGenerator *ArrowExpressionGeneratorImpl
}

func NewArrowStatementGenerator(
	gen *generator,
	seriesVarGen *ArrowSeriesVariableGenerator,
	exprGen *ArrowExpressionGeneratorImpl,
) *ArrowStatementGenerator {
	return &ArrowStatementGenerator{
		gen:           gen,
		seriesVarGen:  seriesVarGen,
		exprGenerator: exprGen,
	}
}

/*
GenerateStatement generates arrow-aware statement code.
Handles variable declarations with Series.Set(), delegates other statements to standard generator.
*/
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
	exprCode, err := s.exprGenerator.Generate(initExpr)
	if err != nil {
		return "", fmt.Errorf("failed to generate init expression for '%s': %w", varName, err)
	}

	return s.seriesVarGen.GenerateAssignment(varName, exprCode), nil
}

func (s *ArrowStatementGenerator) generateTupleDeclaration(arrayPattern *ast.ArrayPattern, initExpr ast.Expression) (string, error) {
	varNames := make([]string, len(arrayPattern.Elements))
	for i, elem := range arrayPattern.Elements {
		varNames[i] = elem.Name
	}

	exprCode, err := s.exprGenerator.Generate(initExpr)
	if err != nil {
		return "", fmt.Errorf("failed to generate tuple init expression: %w", err)
	}

	// Generate tuple unpacking and Series.Set() for each variable
	tempVarNames := make([]string, len(varNames))
	for i, varName := range varNames {
		tempVarNames[i] = "temp_" + varName
	}

	code := s.gen.ind() + fmt.Sprintf("%s := %s\n", join(tempVarNames, ", "), exprCode)
	for i, varName := range varNames {
		code += s.seriesVarGen.GenerateAssignment(varName, tempVarNames[i])
	}

	return code, nil
}

func join(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
