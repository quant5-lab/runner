package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type ChangeHandler struct{}

func (h *ChangeHandler) CanHandle(funcName string) bool {
	return funcName == "ta.change" || funcName == "change"
}

func (h *ChangeHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 {
		return "", fmt.Errorf("ta.change requires at least 1 argument")
	}

	sourceExpr := g.extractSeriesExpression(call.Arguments[0])

	offset := 1
	if len(call.Arguments) >= 2 {
		offsetArg, ok := call.Arguments[1].(*ast.Literal)
		if !ok {
			return "", fmt.Errorf("ta.change offset must be literal")
		}
		var err error
		offset, err = extractPeriod(offsetArg)
		if err != nil {
			return "", fmt.Errorf("ta.change: %w", err)
		}
	}

	return g.generateChange(varName, sourceExpr, offset)
}
