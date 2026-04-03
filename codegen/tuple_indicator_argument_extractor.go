package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* TupleIndicatorArguments holds extracted indicator parameters */
type TupleIndicatorArguments struct {
	SourceExpr string
	Periods    []int
}

/* TupleIndicatorArgumentExtractor extracts arguments from call expressions */
type TupleIndicatorArgumentExtractor struct{}

func NewTupleIndicatorArgumentExtractor() *TupleIndicatorArgumentExtractor {
	return &TupleIndicatorArgumentExtractor{}
}

func (e *TupleIndicatorArgumentExtractor) Extract(
	call *ast.CallExpression,
	sourceExprExtractor func(ast.Expression) string,
	constants map[string]interface{},
) (*TupleIndicatorArguments, error) {
	if len(call.Arguments) < 1 {
		return nil, fmt.Errorf("insufficient arguments for tuple indicator")
	}

	/* For indicators with ImplicitSources, all args are periods (no source arg) */
	sourceExpr := ""
	argStartIndex := 1
	if call.Arguments[0] != nil {
		/* Check if first arg is numeric literal */
		if lit, ok := call.Arguments[0].(*ast.Literal); ok && isNumeric(lit.Value) {
			/* First arg is numeric → no source arg, all are periods */
			argStartIndex = 0
		} else if ident, ok := call.Arguments[0].(*ast.Identifier); ok {
			/* Check if identifier resolves to numeric constant */
			if val, exists := constants[ident.Name]; exists && isNumeric(val) {
				/* First arg is numeric constant → no source arg, all are periods */
				argStartIndex = 0
			} else {
				sourceExpr = sourceExprExtractor(call.Arguments[0])
			}
		} else {
			sourceExpr = sourceExprExtractor(call.Arguments[0])
		}
	}

	var periods []int
	for i := argStartIndex; i < len(call.Arguments); i++ {
		period := e.extractPeriodValue(call.Arguments[i], constants)
		periods = append(periods, period)
	}

	return &TupleIndicatorArguments{
		SourceExpr: sourceExpr,
		Periods:    periods,
	}, nil
}

func isNumeric(val interface{}) bool {
	switch val.(type) {
	case float64, int, int64, int32:
		return true
	default:
		return false
	}
}

func (e *TupleIndicatorArgumentExtractor) extractPeriodValue(
	expr ast.Expression,
	constants map[string]interface{},
) int {
	if lit, ok := expr.(*ast.Literal); ok {
		switch v := lit.Value.(type) {
		case float64:
			return int(v)
		case int:
			return v
		}
	}

	if id, ok := expr.(*ast.Identifier); ok {
		if val, exists := constants[id.Name]; exists {
			switch v := val.(type) {
			case float64:
				return int(v)
			case int:
				return v
			}
		}
	}

	return 12
}
