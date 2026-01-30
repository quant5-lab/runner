package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* LinregHandler generates inline code for linear regression calculations */
type LinregHandler struct{}

func (h *LinregHandler) CanHandle(funcName string) bool {
	return funcName == "ta.linreg" || funcName == "linreg"
}

func (h *LinregHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 3 {
		return "", fmt.Errorf("ta.linreg requires 3 arguments: source, length, offset")
	}

	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.Extract(call, "ta.linreg")
	if err != nil {
		return "", err
	}

	offset := 0
	if len(call.Arguments) >= 3 {
		if lit, ok := call.Arguments[2].(*ast.Literal); ok {
			if v, ok := lit.Value.(float64); ok {
				offset = int(v)
			}
		}
	}

	registry := NewInlineTAIIFERegistry()
	hasher := &ExpressionHasher{}
	sourceHash := hasher.Hash(call.Arguments[0])

	iifeCode, ok := registry.GenerateLinreg(comp.AccessGen, NewConstantPeriod(comp.Period), offset, sourceHash)
	if !ok {
		return "", fmt.Errorf("ta.linreg IIFE generation failed")
	}

	return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, iifeCode), nil
}
