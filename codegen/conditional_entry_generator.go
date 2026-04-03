package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/*
ConditionalEntryGenerator wraps strategy entry calls with conditional execution.

Single Responsibility: Generate conditional wrappers for strategy.entry() when parameter
Design: Pure function - takes condition expression, returns wrapped code
*/
type ConditionalEntryGenerator struct {
	indentation string
}

func NewConditionalEntryGenerator(indentation string) *ConditionalEntryGenerator {
	return &ConditionalEntryGenerator{
		indentation: indentation,
	}
}

/*
WrapWithCondition wraps entry code with if statement when condition exists.

Input:  entryCode="strat.Entry(...)", condition="buySignalSeries.GetCurrent()"
Output: "if value.IsTrue(buySignalSeries.GetCurrent()) {\n    strat.Entry(...)\n}\n"

Input:  entryCode="strat.Entry(...)", condition=""
Output: "strat.Entry(...)"
*/
func (w *ConditionalEntryGenerator) WrapWithCondition(entryCode, conditionExpr string) string {
	if conditionExpr == "" {
		return entryCode
	}

	// PineScript boolean semantics: non-zero = true, zero/NaN = false
	// value.IsTrue() handles float→bool conversion with NaN safety
	return fmt.Sprintf("%sif value.IsTrue(%s) {\n%s    %s%s}\n",
		w.indentation,
		conditionExpr,
		w.indentation,
		entryCode,
		w.indentation,
	)
}

/*
ExtractCondition extracts when parameter from strategy.entry() arguments.
Returns condition expression or empty string if not present.
*/
func (w *ConditionalEntryGenerator) ExtractCondition(args []ast.Expression, extractor *ArgumentExtractor) string {
	conditionExpr, found := extractor.ExtractConditionArgument(args, "when")
	if !found {
		return ""
	}
	return conditionExpr
}
