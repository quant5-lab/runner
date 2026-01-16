package codegen

import (
	"fmt"
	"math"

	"github.com/quant5-lab/runner/ast"
)

/* LowestInlineHandler generates inline expressions for ta.lowest (minimum value over period) */
type LowestInlineHandler struct{}

func NewLowestInlineHandler() *LowestInlineHandler {
	return &LowestInlineHandler{}
}

func (h *LowestInlineHandler) CanHandle(funcName string) bool {
	return funcName == "ta.lowest" || funcName == "lowest"
}

func (h *LowestInlineHandler) GenerateInline(expr *ast.CallExpression, g *generator) (string, error) {
	var accessGen AccessGenerator
	var period int

	if len(expr.Arguments) == 1 {
		periodArg := expr.Arguments[0]
		periodLit, ok := periodArg.(*ast.Literal)
		if !ok {
			periodValue := g.constEvaluator.EvaluateConstant(periodArg)
			if math.IsNaN(periodValue) || periodValue <= 0 {
				if g.inArrowFunctionBody {
					period = -1
				} else {
					return "", fmt.Errorf("ta.lowest period must be compile-time constant")
				}
			} else {
				period = int(periodValue)
			}
		} else {
			var err error
			period, err = extractPeriod(periodLit)
			if err != nil {
				return "", err
			}
		}

		lowIdent := &ast.Identifier{Name: "low"}
		classifier := NewSeriesSourceClassifier()
		lowInfo := classifier.ClassifyAST(lowIdent)
		accessGen = CreateAccessGenerator(lowInfo)
	} else if len(expr.Arguments) >= 2 {
		extractor := NewTAArgumentExtractor(g)
		comp, err := extractor.Extract(expr, "ta.lowest")
		if err != nil {
			return "", err
		}
		accessGen = comp.AccessGen
		period = comp.Period
	} else {
		return "", fmt.Errorf("ta.lowest requires 1 or 2 arguments")
	}

	registry := NewInlineTAIIFERegistry()
	hasher := &ExpressionHasher{}
	sourceHash := ""
	if len(expr.Arguments) > 0 {
		sourceHash = hasher.Hash(expr.Arguments[0])
	}
	iifeCode, ok := registry.Generate("ta.lowest", accessGen, NewConstantPeriod(period), sourceHash)
	if !ok {
		return "", fmt.Errorf("ta.lowest IIFE generation failed")
	}

	return iifeCode, nil
}
