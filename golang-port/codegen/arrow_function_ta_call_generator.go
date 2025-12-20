package codegen

import (
	"fmt"
	"strconv"

	"github.com/quant5-lab/runner/ast"
)

type ArrowFunctionTACallGenerator struct {
	gen          *generator
	iifeRegistry *InlineTAIIFERegistry
}

func NewArrowFunctionTACallGenerator(gen *generator) *ArrowFunctionTACallGenerator {
	return &ArrowFunctionTACallGenerator{
		gen:          gen,
		iifeRegistry: NewInlineTAIIFERegistry(),
	}
}

func (a *ArrowFunctionTACallGenerator) Generate(call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	if !a.iifeRegistry.IsSupported(funcName) {
		return "", fmt.Errorf("TA function %s not supported in arrow function context", funcName)
	}

	accessor, period, err := a.extractTAArguments(funcName, call)
	if err != nil {
		return "", fmt.Errorf("failed to extract TA arguments: %w", err)
	}

	code, ok := a.iifeRegistry.Generate(funcName, accessor, period)
	if !ok {
		return "", fmt.Errorf("failed to generate IIFE for %s", funcName)
	}

	return code, nil
}

func (a *ArrowFunctionTACallGenerator) extractTAArguments(funcName string, call *ast.CallExpression) (AccessGenerator, int, error) {
	if len(call.Arguments) == 1 {
		return a.extractSingleArgumentForm(funcName, call)
	}

	if len(call.Arguments) < 2 {
		return nil, 0, fmt.Errorf("TA function requires at least 1 argument")
	}

	sourceArg := call.Arguments[0]
	periodArg := call.Arguments[1]

	accessor, err := a.createAccessorFromExpression(sourceArg)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create accessor: %w", err)
	}

	period, err := a.extractPeriodValue(periodArg)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to extract period: %w", err)
	}

	return accessor, period, nil
}

func (a *ArrowFunctionTACallGenerator) extractSingleArgumentForm(funcName string, call *ast.CallExpression) (AccessGenerator, int, error) {
	periodArg := call.Arguments[0]

	period, err := a.extractPeriodValue(periodArg)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to extract period: %w", err)
	}

	accessor := a.getDefaultSourceAccessor(funcName)
	return accessor, period, nil
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

	default:
		return nil, fmt.Errorf("unsupported source expression type: %T", expr)
	}
}

func (a *ArrowFunctionTACallGenerator) extractPeriodValue(expr ast.Expression) (int, error) {
	switch e := expr.(type) {
	case *ast.Literal:
		if floatVal, ok := e.Value.(float64); ok {
			return int(floatVal), nil
		}
		if intVal, ok := e.Value.(int); ok {
			return intVal, nil
		}
		if strVal, ok := e.Value.(string); ok {
			return strconv.Atoi(strVal)
		}
		return 0, fmt.Errorf("period literal is not numeric: %v", e.Value)

	case *ast.Identifier:
		if varType, exists := a.gen.variables[e.Name]; exists && varType == "float" {
			return 20, nil
		}
		return 0, fmt.Errorf("period identifier %s not supported", e.Name)

	default:
		return 0, fmt.Errorf("unsupported period expression type: %T", expr)
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
