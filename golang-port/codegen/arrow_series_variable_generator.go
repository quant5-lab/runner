package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/*
ArrowSeriesVariableGenerator generates Series-based variable declarations for arrow functions.

All variables in arrow functions use ForwardSeriesBuffer paradigm via ArrowContext.
This ensures:
  - Historical access: var[1], var[2]
  - TA function compatibility: rma, sma need historical values
  - Universal consistency: same pattern as global variables

Architecture:

	varName := arrowCtx.GetOrCreateSeries("varName")
	varName.Set(expression)
*/
type ArrowSeriesVariableGenerator struct {
	indentation string
	exprGen     ArrowExpressionGenerator
}

func NewArrowSeriesVariableGenerator(indent string, exprGen ArrowExpressionGenerator) *ArrowSeriesVariableGenerator {
	return &ArrowSeriesVariableGenerator{
		indentation: indent,
		exprGen:     exprGen,
	}
}

/*
GenerateDeclaration creates Series initialization for a variable.
Returns code like:

	upSeries := arrowCtx.GetOrCreateSeries("up")
*/
func (g *ArrowSeriesVariableGenerator) GenerateDeclaration(varName string) string {
	return g.indentation + fmt.Sprintf("%sSeries := arrowCtx.GetOrCreateSeries(%q)\n", varName, varName)
}

/*
GenerateAssignment creates Series.Set() statement for a variable.
Returns code like:

	upSeries.Set(change(high))
*/
func (g *ArrowSeriesVariableGenerator) GenerateAssignment(varName string, valueExpr string) string {
	return g.indentation + fmt.Sprintf("%sSeries.Set(%s)\n", varName, valueExpr)
}

/*
GenerateDeclarationAndAssignment combines declaration and assignment.
Used when variable is both declared and initialized in one statement.
*/
func (g *ArrowSeriesVariableGenerator) GenerateDeclarationAndAssignment(varName string, initExpr ast.Expression) (string, error) {
	valueCode, err := g.exprGen.Generate(initExpr)
	if err != nil {
		return "", fmt.Errorf("failed to generate expression for %s: %w", varName, err)
	}

	declaration := g.GenerateDeclaration(varName)
	assignment := g.GenerateAssignment(varName, valueCode)

	return declaration + assignment, nil
}
