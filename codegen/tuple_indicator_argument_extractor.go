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

	sourceExpr := sourceExprExtractor(call.Arguments[0])

	var periods []int
	for i := 1; i < len(call.Arguments); i++ {
		period := e.extractPeriodValue(call.Arguments[i], constants)
		periods = append(periods, period)
	}

	return &TupleIndicatorArguments{
		SourceExpr: sourceExpr,
		Periods:    periods,
	}, nil
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
