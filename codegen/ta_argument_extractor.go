package codegen

import (
	"fmt"

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

type TAArgumentComponentsWithDynamic struct {
	SourceExpr    ast.Expression
	PeriodResult  PeriodEvaluationResult
	SourceInfo    SourceInfo
	AccessGen     AccessGenerator
	NeedsNaNCheck bool
	Preamble      string
}

func (e *TAArgumentExtractor) ExtractWithDynamic(call *ast.CallExpression, funcName string) (*TAArgumentComponentsWithDynamic, error) {
	if len(call.Arguments) < 2 {
		return nil, fmt.Errorf("%s requires at least 2 arguments", funcName)
	}

	sourceExpr := call.Arguments[0]
	periodResult := e.extractPeriodResult(call.Arguments[1], funcName)

	if periodResult.IsFailed() {
		return nil, fmt.Errorf("%s: %s", funcName, periodResult.FailureReason)
	}

	if e.isTrBuiltin(sourceExpr) {
		return &TAArgumentComponentsWithDynamic{
			SourceExpr:    sourceExpr,
			PeriodResult:  periodResult,
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

	return &TAArgumentComponentsWithDynamic{
		SourceExpr:    sourceExpr,
		PeriodResult:  periodResult,
		SourceInfo:    sourceInfo,
		AccessGen:     accessGen,
		NeedsNaNCheck: needsNaN,
		Preamble:      preamble,
	}, nil
}

/* Extract prepares components for TA indicators requiring compile-time constant period.
 * Delegates to ExtractWithDynamic and converts PeriodEvaluationResult to int period.
 * Arrow function context returns sentinel period=-1 for deferred resolution.
 * Returns error for runtime dynamic periods outside arrow context. */
func (e *TAArgumentExtractor) Extract(call *ast.CallExpression, funcName string) (*TAArgumentComponents, error) {
	dynComp, err := e.ExtractWithDynamic(call, funcName)
	if err != nil {
		return nil, err
	}

	if dynComp.PeriodResult.IsRuntimeDynamic() {
		if e.generator.inArrowFunctionBody {
			return &TAArgumentComponents{
				SourceExpr:    dynComp.SourceExpr,
				Period:        -1,
				SourceInfo:    dynComp.SourceInfo,
				AccessGen:     dynComp.AccessGen,
				NeedsNaNCheck: dynComp.NeedsNaNCheck,
				Preamble:      dynComp.Preamble,
			}, nil
		}
		return nil, fmt.Errorf("%s period must be compile-time constant (got dynamic expression)", funcName)
	}

	return &TAArgumentComponents{
		SourceExpr:    dynComp.SourceExpr,
		Period:        dynComp.PeriodResult.StaticValue,
		SourceInfo:    dynComp.SourceInfo,
		AccessGen:     dynComp.AccessGen,
		NeedsNaNCheck: dynComp.NeedsNaNCheck,
		Preamble:      dynComp.Preamble,
	}, nil
}

/* ExtractSourceOnly prepares components for fixed-period TA indicators that take only a source argument (e.g. ta.swma). */
func (e *TAArgumentExtractor) ExtractSourceOnly(call *ast.CallExpression, funcName string) (*TAArgumentComponents, error) {
	if len(call.Arguments) < 1 {
		return nil, fmt.Errorf("%s requires at least 1 argument", funcName)
	}

	sourceExpr := call.Arguments[0]
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
		Period:        0,
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

/* ExtractConstantPeriodAt evaluates a compile-time constant period from an arbitrary argument position.
 * For parameters typed as Pine Script "simple int" — rejects runtime-dynamic expressions.
 * Use when a function has multiple period parameters (e.g. ta.tsi shortLength + longLength). */
func (e *TAArgumentExtractor) ExtractConstantPeriodAt(call *ast.CallExpression, argPosition int, funcName string) (int, error) {
	if len(call.Arguments) <= argPosition {
		return 0, fmt.Errorf("%s requires argument at position %d", funcName, argPosition)
	}
	result := e.extractPeriodResult(call.Arguments[argPosition], funcName)
	if result.IsFailed() {
		return 0, fmt.Errorf("%s: %s", funcName, result.FailureReason)
	}
	if result.IsRuntimeDynamic() {
		return 0, fmt.Errorf("%s period at position %d must be compile-time constant (Pine Script 'simple int')", funcName, argPosition)
	}
	if result.StaticValue <= 0 {
		return 0, fmt.Errorf("%s period at position %d must be positive, got %d", funcName, argPosition, result.StaticValue)
	}
	return result.StaticValue, nil
}

func (e *TAArgumentExtractor) extractPeriodResult(periodArg ast.Expression, funcName string) PeriodEvaluationResult {
	return evaluatePeriodExpression(e.generator, periodArg)
}

func extractPeriodFromLiteral(lit *ast.Literal) (int, error) {
	switch v := lit.Value.(type) {
	case float64:
		if v <= 0 {
			return 0, fmt.Errorf("period must be positive, got %.0f", v)
		}
		return int(v), nil
	case int:
		if v <= 0 {
			return 0, fmt.Errorf("period must be positive, got %d", v)
		}
		return v, nil
	default:
		return 0, fmt.Errorf("period must be numeric, got %T", v)
	}
}
