package codegen

import (
	"fmt"
	"strconv"

	"github.com/quant5-lab/runner/ast"
)

type ArrowFunctionTACallGenerator struct {
	gen               *generator
	exprGen           ArrowExpressionGenerator
	iifeRegistry      *InlineTAIIFERegistry
	accessorFactory   *ArrowAwareAccessorFactory
	signatureResolver *ArrowTACallSignatureResolver
}

func NewArrowFunctionTACallGenerator(gen *generator, exprGen ArrowExpressionGenerator) *ArrowFunctionTACallGenerator {
	accessResolver := NewArrowSeriesAccessResolver()

	for paramName, paramType := range gen.variables {
		if paramType == "float" {
			accessResolver.RegisterParameter(paramName)
		}
	}

	identifierResolver := NewArrowIdentifierResolver(accessResolver)
	accessorFactory := NewArrowAwareAccessorFactory(identifierResolver, exprGen, gen, gen.symbolTable)
	signatureRegistry := NewTAFunctionSignatureRegistry()

	return &ArrowFunctionTACallGenerator{
		gen:               gen,
		exprGen:           exprGen,
		iifeRegistry:      NewInlineTAIIFERegistry(),
		accessorFactory:   accessorFactory,
		signatureResolver: NewArrowTACallSignatureResolver(signatureRegistry),
	}
}

func (a *ArrowFunctionTACallGenerator) Generate(call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	// Check if this is a user-defined function first
	detector := NewUserDefinedFunctionDetector(a.gen.variables)
	if detector.IsUserDefinedFunction(funcName) {
		handler := &UserDefinedFunctionHandler{}
		return handler.GenerateCode(a.gen, call)
	}

	// Special case: fixnan uses inline IIFE with NaN check
	if funcName == "fixnan" || funcName == "ta.fixnan" {
		return a.generateFixnanIIFE(call)
	}

	if !a.iifeRegistry.IsSupported(funcName) {
		return "", fmt.Errorf("TA function %s not supported in arrow function context", funcName)
	}

	accessor, periodExpr, err := a.extractTAArguments(funcName, call)
	if err != nil {
		return "", fmt.Errorf("failed to extract TA arguments: %w", err)
	}

	// Generate hash from source expression to prevent series name collisions
	sourceHash := ""
	if len(call.Arguments) > 0 {
		hasher := &ExpressionHasher{}
		sourceHash = hasher.Hash(call.Arguments[0])
	}

	code, ok := a.iifeRegistry.Generate(funcName, accessor, periodExpr, sourceHash)
	if !ok {
		return "", fmt.Errorf("failed to generate IIFE for %s", funcName)
	}

	return code, nil
}

/*
generateFixnanIIFE creates inline code for fixnan(source).
Returns: func() float64 { val := source; if math.IsNaN(val) { return 0.0 }; return val }()
*/
func (a *ArrowFunctionTACallGenerator) generateFixnanIIFE(call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 {
		return "", fmt.Errorf("fixnan requires 1 argument")
	}

	sourceArg := call.Arguments[0]
	sourceCode, err := a.gen.generateArrowFunctionExpression(sourceArg)
	if err != nil {
		return "", fmt.Errorf("failed to generate fixnan source: %w", err)
	}

	return fmt.Sprintf("func() float64 { val := %s; if math.IsNaN(val) { return 0.0 }; return val }()", sourceCode), nil
}

func (a *ArrowFunctionTACallGenerator) extractTAArguments(funcName string, call *ast.CallExpression) (AccessGenerator, PeriodExpression, error) {
	if funcName == "ta.change" || funcName == "change" {
		return a.extractChangeArguments(call)
	}

	resolved, err := a.signatureResolver.ResolveCall(funcName, call)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve TA call signature: %w", err)
	}

	var sourceArg ast.Expression
	if resolved.NeedsDefaultSource {
		sourceArg = &ast.Identifier{Name: resolved.DefaultSourceName}
	} else {
		sourceArg = resolved.SourceExpr
	}

	accessor, err := a.accessorFactory.CreateAccessorForExpression(sourceArg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create accessor: %w", err)
	}

	periodExpr, err := a.extractPeriodExpression(resolved.LengthExpr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract period: %w", err)
	}

	return accessor, periodExpr, nil
}

func (a *ArrowFunctionTACallGenerator) extractChangeArguments(call *ast.CallExpression) (AccessGenerator, PeriodExpression, error) {
	if len(call.Arguments) < 1 {
		return nil, nil, fmt.Errorf("change() requires at least 1 argument (source)")
	}

	sourceArg := call.Arguments[0]
	accessor, err := a.accessorFactory.CreateAccessorForExpression(sourceArg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create accessor for change(): %w", err)
	}

	offsetExpr := PeriodExpression(NewConstantPeriod(1))
	if len(call.Arguments) >= 2 {
		extracted, err := a.extractPeriodExpression(call.Arguments[1])
		if err != nil {
			return nil, nil, fmt.Errorf("failed to extract offset for change(): %w", err)
		}
		offsetExpr = extracted
	}

	return accessor, offsetExpr, nil
}

func (a *ArrowFunctionTACallGenerator) extractPeriodExpression(expr ast.Expression) (PeriodExpression, error) {
	switch e := expr.(type) {
	case *ast.Literal:
		if floatVal, ok := e.Value.(float64); ok {
			return NewConstantPeriod(int(floatVal)), nil
		}
		if intVal, ok := e.Value.(int); ok {
			return NewConstantPeriod(intVal), nil
		}
		if strVal, ok := e.Value.(string); ok {
			periodInt, err := strconv.Atoi(strVal)
			if err != nil {
				return nil, fmt.Errorf("period string is not numeric: %s", strVal)
			}
			return NewConstantPeriod(periodInt), nil
		}
		return nil, fmt.Errorf("period literal is not numeric: %v", e.Value)

	case *ast.Identifier:
		/* Check if identifier is a known variable (arrow function parameter) */
		if _, exists := a.gen.variables[e.Name]; !exists {
			return nil, fmt.Errorf("unknown period identifier: %s (not an arrow function parameter)", e.Name)
		}
		return NewRuntimePeriod(e.Name), nil

	default:
		return nil, fmt.Errorf("unsupported period expression type: %T", expr)
	}
}

type ArrowFunctionParameterAccessor struct {
	parameterName string
}

func NewArrowFunctionParameterAccessor(parameterName string) *ArrowFunctionParameterAccessor {
	return &ArrowFunctionParameterAccessor{
		parameterName: parameterName,
	}
}

func (a *ArrowFunctionParameterAccessor) GenerateLoopValueAccess(loopVar string) string {
	return fmt.Sprintf("%sSeries.Get(%s)", a.parameterName, loopVar)
}

func (a *ArrowFunctionParameterAccessor) GenerateInitialValueAccess(period int) string {
	return fmt.Sprintf("%sSeries.Get(%d-1)", a.parameterName, period)
}

func (a *ArrowFunctionParameterAccessor) GenerateCurrentValueAccess() string {
	return fmt.Sprintf("%sSeries.GetCurrent()", a.parameterName)
}

/* GetBaseOffset returns 0 - arrow function parameter access is current bar relative */
func (a *ArrowFunctionParameterAccessor) GetBaseOffset() int {
	return 0
}
