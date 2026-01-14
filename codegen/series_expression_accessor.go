package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/*
SeriesExpressionAccessor generates series-aware code for complex expressions.

Responsibility (SRP):
  - Single purpose: convert AST expressions to Go code with series access (varName → varNameSeries.Get(offset))
  - Uses SeriesAccessConverter for transformation logic
  - Implements AccessGenerator interface for TA indicators and arrow functions

Design Rationale:
  - DRY: Reuses SeriesAccessConverter instead of reimplementing conversion logic
  - KISS: Simple delegation pattern - stores AST and symbol table, delegates to converter
  - Unified: Single accessor for both TA context and arrow function context

Usage:
  - TA context: TAArgumentExtractor for complex expressions (ta.sma(close-open, 14))
  - Arrow context: ArrowAwareAccessorFactory for binary/conditional expressions
*/
type SeriesExpressionAccessor struct {
	expr          ast.Expression
	symbolTable   SymbolTable
	lookupCallVar CallVarLookup
}

func NewSeriesExpressionAccessor(
	expr ast.Expression,
	symbolTable SymbolTable,
	lookupCallVar CallVarLookup,
) *SeriesExpressionAccessor {
	return &SeriesExpressionAccessor{
		expr:          expr,
		symbolTable:   symbolTable,
		lookupCallVar: lookupCallVar,
	}
}

/* GenerateLoopValueAccess converts expression to series-aware code for loop iterations */
func (a *SeriesExpressionAccessor) GenerateLoopValueAccess(loopVar string) string {
	if a.symbolTable == nil {
		return "math.NaN()"
	}

	converter := NewSeriesAccessConverter(a.symbolTable, loopVar, a.lookupCallVar)
	code, err := converter.ConvertExpression(a.expr)
	if err != nil {
		return "math.NaN()"
	}

	return code
}

/* GenerateInitialValueAccess converts expression for fixed offset access */
func (a *SeriesExpressionAccessor) GenerateInitialValueAccess(period int) string {
	offset := fmt.Sprintf("%d", period-1)
	if a.symbolTable == nil {
		return "math.NaN()"
	}

	converter := NewSeriesAccessConverter(a.symbolTable, offset, a.lookupCallVar)
	code, err := converter.ConvertExpression(a.expr)
	if err != nil {
		return "math.NaN()"
	}

	return code
}
