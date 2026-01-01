package codegen

import (
	"fmt"
	"strconv"

	"github.com/quant5-lab/runner/ast"
)

type ArrowFunctionTACallGenerator struct {
	gen          *generator
	exprGen      ArrowExpressionGenerator
	iifeRegistry *InlineTAIIFERegistry
}

func NewArrowFunctionTACallGenerator(gen *generator, exprGen ArrowExpressionGenerator) *ArrowFunctionTACallGenerator {
	return &ArrowFunctionTACallGenerator{
		gen:          gen,
		exprGen:      exprGen,
		iifeRegistry: NewInlineTAIIFERegistry(),
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

	if len(call.Arguments) < 2 {
		return nil, nil, fmt.Errorf("TA function requires 2 arguments (source, period), got %d", len(call.Arguments))
	}

	sourceArg := call.Arguments[0]
	periodArg := call.Arguments[1]

	accessor, err := a.createAccessorFromExpression(sourceArg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create accessor: %w", err)
	}

	periodExpr, err := a.extractPeriodExpression(periodArg)
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
	accessor, err := a.createAccessorFromExpression(sourceArg)
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

func (a *ArrowFunctionTACallGenerator) extractSingleArgumentForm(funcName string, call *ast.CallExpression) (AccessGenerator, PeriodExpression, error) {
	periodArg := call.Arguments[0]

	periodExpr, err := a.extractPeriodExpression(periodArg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract period: %w", err)
	}

	accessor := a.getDefaultSourceAccessor(funcName)
	return accessor, periodExpr, nil
}

func (a *ArrowFunctionTACallGenerator) getDefaultSourceAccessor(funcName string) AccessGenerator {
	switch funcName {
	case "ta.highest", "highest":
		return NewOHLCVFieldAccessGenerator("High")
	case "ta.lowest", "lowest":
		return NewOHLCVFieldAccessGenerator("Low")
	default:
		return NewOHLCVFieldAccessGenerator("Close")
	}
}

func (a *ArrowFunctionTACallGenerator) createAccessorFromExpression(expr ast.Expression) (AccessGenerator, error) {
	switch e := expr.(type) {
	case *ast.Identifier:
		// tr builtin generates inline calculation
		if e.Name == "tr" {
			return NewBuiltinTrueRangeAccessor(), nil
		}

		if varType, exists := a.gen.variables[e.Name]; exists && varType == "float" {
			return NewArrowFunctionParameterAccessor(e.Name), nil
		}

		classifier := NewSeriesSourceClassifier()
		sourceInfo := classifier.ClassifyAST(e)
		return CreateAccessGenerator(sourceInfo), nil

	case *ast.MemberExpression:
		if obj, ok := e.Object.(*ast.Identifier); ok {
			if obj.Name == "ctx" {
				if prop, ok := e.Property.(*ast.Identifier); ok {
					fieldName := capitalizeFirst(prop.Name)
					return NewOHLCVFieldAccessGenerator(fieldName), nil
				}
			}
		}
		return nil, fmt.Errorf("unsupported member expression in TA call")

	case *ast.ConditionalExpression:
		if a.gen.symbolTable != nil {
			return NewSeriesExpressionAccessor(e, a.gen.symbolTable), nil
		}

		tempVarName := "ternary_source_temp"
		condCode, err := a.exprGen.Generate(e)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ternary expression: %w", err)
		}

		return &FixnanCallExpressionAccessor{
			tempVarName: tempVarName,
			tempVarCode: fmt.Sprintf("%s := %s", tempVarName, condCode),
			exprCode:    condCode,
		}, nil

	case *ast.BinaryExpression:
		if a.gen.symbolTable != nil {
			return NewSeriesExpressionAccessor(e, a.gen.symbolTable), nil
		}

		tempVarName := "binary_source_temp"
		binaryCode, err := a.exprGen.Generate(e)
		if err != nil {
			return nil, fmt.Errorf("failed to generate binary expression: %w", err)
		}

		return &FixnanCallExpressionAccessor{
			tempVarName: tempVarName,
			tempVarCode: fmt.Sprintf("%s := %s", tempVarName, binaryCode),
			exprCode:    binaryCode,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported source expression type: %T", expr)
	}
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

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-32) + s[1:]
	}
	return s
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
