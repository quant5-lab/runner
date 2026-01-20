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
	Preamble      string
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

	// Check for tr builtin (identifier "tr" or member "ta.tr")
	if e.isTrBuiltin(sourceExpr) {
		return &TAArgumentComponents{
			SourceExpr:    sourceExpr,
			Period:        period,
			SourceInfo:    SourceInfo{},
			AccessGen:     NewBuiltinTrueRangeAccessor(),
			NeedsNaNCheck: false,
			Preamble:      "",
		}, nil
	}

	sourceInfo := e.classifier.ClassifyAST(sourceExpr)
	accessGen := CreateAccessGenerator(sourceInfo)
	needsNaN := sourceInfo.IsSeriesVariable()
	preamble := ""

	if e.requiresExpressionAccessor(sourceExpr, sourceInfo) {
		preambleCode, err := e.registerNestedTempVars(sourceExpr)
		if err != nil {
			return nil, err
		}
		preamble += preambleCode
		accessGen = NewSeriesExpressionAccessor(sourceExpr, e.generator.symbolTable, e.generator.tempVarMgr.GetVarNameForCall)
		needsNaN = true
	}

	return &TAArgumentComponents{
		SourceExpr:    sourceExpr,
		Period:        period,
		SourceInfo:    sourceInfo,
		AccessGen:     accessGen,
		NeedsNaNCheck: needsNaN,
		Preamble:      preamble,
	}, nil
}

func (e *TAArgumentExtractor) isTrBuiltin(expr ast.Expression) bool {
	if id, ok := expr.(*ast.Identifier); ok && id.Name == "tr" {
		return true
	}
	if mem, ok := expr.(*ast.MemberExpression); ok {
		if obj, ok := mem.Object.(*ast.Identifier); ok {
			if prop, ok := mem.Property.(*ast.Identifier); ok {
				return obj.Name == "ta" && prop.Name == "tr"
			}
		}
	}
	return false
}

// requiresExpressionAccessor returns true when the source expression is not a simple OHLCV field/series
// and therefore needs expression-aware offset rewriting instead of the default classifier fallback.
func (e *TAArgumentExtractor) requiresExpressionAccessor(sourceExpr ast.Expression, info SourceInfo) bool {
	// Simple series identifier or builtin: use default accessor
	if id, ok := sourceExpr.(*ast.Identifier); ok {
		if info.IsSeriesVariable() {
			return false
		}
		if e.classifier.isBuiltinOHLCVField(id.Name) {
			return false
		}
		if e.classifier.isDerivedPrice(id.Name) {
			return false
		}
		return true
	}

	if mem, ok := sourceExpr.(*ast.MemberExpression); ok {
		if obj, ok := mem.Object.(*ast.Identifier); ok && mem.Computed {
			if e.classifier.isBuiltinOHLCVField(obj.Name) {
				_, isLiteral := mem.Property.(*ast.Literal)
				return !isLiteral
			}
			if e.classifier.isDerivedPrice(obj.Name) {
				_, isLiteral := mem.Property.(*ast.Literal)
				return !isLiteral
			}
		}
		return true
	}

	// Anything else (BinaryExpression, CallExpression, ConditionalExpression, etc.)
	return true
}

// registerNestedTempVars materializes nested TA calls inside complex expressions so they can be referenced with offsets.
func (e *TAArgumentExtractor) registerNestedTempVars(expr ast.Expression) (string, error) {
	nestedCalls := e.generator.exprAnalyzer.FindNestedCalls(expr)
	code := ""

	if len(nestedCalls) == 0 {
		return code, nil
	}

	for i := len(nestedCalls) - 1; i >= 0; i-- {
		callInfo := nestedCalls[i]

		if callInfo.Call == expr {
			continue
		}

		if e.generator.runtimeOnlyFilter.IsRuntimeOnly(callInfo.FuncName) {
			continue
		}

		isTAFunction := e.generator.taRegistry.IsSupported(callInfo.FuncName)
		containsNestedTA := false
		if !isTAFunction {
			mathNestedCalls := e.generator.exprAnalyzer.FindNestedCalls(callInfo.Call)
			for _, mathNested := range mathNestedCalls {
				if mathNested.Call != callInfo.Call && e.generator.taRegistry.IsSupported(mathNested.FuncName) {
					containsNestedTA = true
					break
				}
			}
		}

		if !isTAFunction && !containsNestedTA {
			continue
		}

		tempVarName := e.generator.tempVarMgr.GetOrCreate(callInfo)
		tempCode, err := e.generator.generateVariableFromCall(tempVarName, callInfo.Call)
		if err != nil {
			return "", fmt.Errorf("failed to generate temp var %s: %w", tempVarName, err)
		}
		code += tempCode
	}

	return code, nil
}

func (e *TAArgumentExtractor) extractPeriod(periodArg ast.Expression, funcName string) (int, error) {
	if periodLit, ok := periodArg.(*ast.Literal); ok {
		return extractPeriodFromLiteral(periodLit)
	}

	periodValue := e.generator.constEvaluator.EvaluateConstant(periodArg)
	if math.IsNaN(periodValue) || periodValue <= 0 {
		// Allow runtime periods within arrow functions (use -1 as sentinel)
		if e.generator.inArrowFunctionBody {
			return -1, nil
		}
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
