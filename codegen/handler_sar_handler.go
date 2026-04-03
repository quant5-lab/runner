package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// SarHandler generates Parabolic SAR code.
// Implements CompositeIndicatorMetadata because SAR requires 3 internal state series.
type SarHandler struct{}

func (h *SarHandler) CanHandle(funcName string) bool {
	return funcName == "ta.sar" || funcName == "sar"
}

func (h *SarHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	start, inc, maxAF, err := extractSARArguments(call)
	if err != nil {
		return "", err
	}

	var context StatefulIndicatorContext
	if g.inArrowFunctionBody {
		context = NewArrowFunctionIndicatorContext()
	} else {
		context = NewTopLevelIndicatorContext()
	}

	builder := NewSARIndicatorBuilder(varName, start, inc, maxAF, context)
	code := g.indentCode(builder.Build())

	return code, nil
}

func (h *SarHandler) GetInternalSeriesNames(varName string, call *ast.CallExpression) ([]string, error) {
	return []string{
		fmt.Sprintf("_%s_ep", varName),
		fmt.Sprintf("_%s_af", varName),
		fmt.Sprintf("_%s_trend", varName),
	}, nil
}

func extractSARArguments(call *ast.CallExpression) (start, inc, maxAF float64, err error) {
	start = 0.02
	inc = 0.02
	maxAF = 0.2

	if len(call.Arguments) >= 1 {
		if lit, ok := call.Arguments[0].(*ast.Literal); ok {
			if v, ok := lit.Value.(float64); ok {
				start = v
			}
		}
	}
	if len(call.Arguments) >= 2 {
		if lit, ok := call.Arguments[1].(*ast.Literal); ok {
			if v, ok := lit.Value.(float64); ok {
				inc = v
			}
		}
	}
	if len(call.Arguments) >= 3 {
		if lit, ok := call.Arguments[2].(*ast.Literal); ok {
			if v, ok := lit.Value.(float64); ok {
				maxAF = v
			}
		}
	}
	return
}
