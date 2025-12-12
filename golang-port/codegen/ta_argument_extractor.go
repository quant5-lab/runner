package codegen

import (
	"fmt"
	"math"

	"github.com/quant5-lab/runner/ast"
)

/* TAArgumentComponents contains prepared components for TA indicator generation */
type TAArgumentComponents struct {
	SourceExpr    ast.Expression
	Period        int
	SourceInfo    SourceInfo
	AccessGen     AccessGenerator
	NeedsNaNCheck bool
}

/* TAArgumentExtractor prepares TA function arguments for code generation.
 * Centralizes extraction, classification, and accessor creation.
 * Eliminates duplication across all TA handlers.
 *
 * Usage:
 *   extractor := NewTAArgumentExtractor(g)
 *   comp, err := extractor.Extract(call, "ta.sma")
 *   builder := NewTAIndicatorBuilder(name, varName, comp.Period, comp.AccessGen, comp.NeedsNaNCheck)
 */
type TAArgumentExtractor struct {
	generator  *generator
	classifier *SeriesSourceClassifier
}

func NewTAArgumentExtractor(g *generator) *TAArgumentExtractor {
	return &TAArgumentExtractor{
		generator:  g,
		classifier: NewSeriesSourceClassifier(),
	}
}

/* Extract prepares components needed for TA indicator generation */
func (e *TAArgumentExtractor) Extract(call *ast.CallExpression, funcName string) (*TAArgumentComponents, error) {
	if len(call.Arguments) < 2 {
		return nil, fmt.Errorf("%s requires at least 2 arguments", funcName)
	}

	sourceExpr := call.Arguments[0]
	period, err := e.extractPeriod(call.Arguments[1], funcName)
	if err != nil {
		return nil, err
	}

	sourceInfo := e.classifier.ClassifyAST(sourceExpr)
	accessGen := CreateAccessGenerator(sourceInfo)
	needsNaN := sourceInfo.IsSeriesVariable()

	return &TAArgumentComponents{
		SourceExpr:    sourceExpr,
		Period:        period,
		SourceInfo:    sourceInfo,
		AccessGen:     accessGen,
		NeedsNaNCheck: needsNaN,
	}, nil
}

func (e *TAArgumentExtractor) extractPeriod(periodArg ast.Expression, funcName string) (int, error) {
	if periodLit, ok := periodArg.(*ast.Literal); ok {
		return extractPeriodFromLiteral(periodLit)
	}

	periodValue := e.generator.constEvaluator.EvaluateConstant(periodArg)
	if math.IsNaN(periodValue) || periodValue <= 0 {
		return 0, fmt.Errorf("%s period must be compile-time constant (got %T that evaluates to NaN)", funcName, periodArg)
	}

	return int(periodValue), nil
}

func extractPeriodFromLiteral(lit *ast.Literal) (int, error) {
	switch v := lit.Value.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	default:
		return 0, fmt.Errorf("period must be numeric, got %T", v)
	}
}
