package codegen

import (
	"fmt"
	"math"

	"github.com/quant5-lab/runner/ast"
)

/* HighestInlineHandler generates inline expressions for ta.highest (maximum value over period) */
type HighestInlineHandler struct{}

func NewHighestInlineHandler() *HighestInlineHandler {
	return &HighestInlineHandler{}
}

func (h *HighestInlineHandler) CanHandle(funcName string) bool {
	return funcName == "ta.highest" || funcName == "highest"
}

func (h *HighestInlineHandler) GenerateInline(expr *ast.CallExpression, g *generator) (string, error) {
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
					return "", fmt.Errorf("ta.highest period must be compile-time constant")
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

		highIdent := &ast.Identifier{Name: "high"}
		classifier := NewSeriesSourceClassifier()
		highInfo := classifier.ClassifyAST(highIdent)
		accessGen = CreateAccessGenerator(highInfo)
	} else if len(expr.Arguments) >= 2 {
		extractor := NewTAArgumentExtractor(g)
		comp, err := extractor.Extract(expr, "ta.highest")
		if err != nil {
			return "", err
		}
		accessGen = comp.AccessGen
		period = comp.Period
	} else {
		return "", fmt.Errorf("ta.highest requires 1 or 2 arguments")
	}

	registry := NewInlineTAIIFERegistry()
	hasher := &ExpressionHasher{}
	sourceHash := ""
	if len(expr.Arguments) > 0 {
		sourceHash = hasher.Hash(expr.Arguments[0])
	}
	iifeCode, ok := registry.Generate("ta.highest", accessGen, NewConstantPeriod(period), sourceHash)
	if !ok {
		return "", fmt.Errorf("ta.highest IIFE generation failed")
	}

	return iifeCode, nil
}
