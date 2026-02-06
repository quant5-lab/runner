package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type TAGeneratorFactory struct {
	staticGenerator  *StaticPeriodTAGenerator
	dynamicGenerator *DynamicPeriodTAGenerator
	periodExtractor  *PeriodExtractor
}

func NewTAGeneratorFactory(
	staticGen *StaticPeriodTAGenerator,
	dynamicGen *DynamicPeriodTAGenerator,
	extractor *PeriodExtractor,
) *TAGeneratorFactory {
	return &TAGeneratorFactory{
		staticGenerator:  staticGen,
		dynamicGenerator: dynamicGen,
		periodExtractor:  extractor,
	}
}

func (f *TAGeneratorFactory) GenerateTA(
	varName string,
	functionName string,
	call *ast.CallExpression,
) (string, error) {
	periodResult, err := f.periodExtractor.ExtractAndValidate(call, functionName)
	if err != nil {
		return "", err
	}

	sourceExpr, err := f.extractSourceExpression(call, functionName)
	if err != nil {
		return "", err
	}

	if periodResult.IsCompileTimeConstant() {
		code, err := f.staticGenerator.Generate(varName, functionName, sourceExpr, periodResult)
		if err != nil {
			return "", err
		}
		if code != "" {
			return code, nil
		}
	}

	if periodResult.IsRuntimeDynamic() {
		code, err := f.dynamicGenerator.Generate(varName, functionName, sourceExpr, periodResult)
		if err != nil {
			return "", err
		}
		if code != "" {
			return code, nil
		}
	}

	return "", fmt.Errorf("no generator available for %s with period kind %d", functionName, periodResult.Kind)
}

func (f *TAGeneratorFactory) extractSourceExpression(
	call *ast.CallExpression,
	functionName string,
) (ast.Expression, error) {
	if functionName == "ta.highest" || functionName == "ta.lowest" {
		if len(call.Arguments) >= 2 {
			return call.Arguments[0], nil
		}
		if len(call.Arguments) == 1 {
			if functionName == "ta.highest" {
				return &ast.Identifier{Name: "high"}, nil
			}
			return &ast.Identifier{Name: "low"}, nil
		}
		return nil, fmt.Errorf("%s requires at least 1 argument", functionName)
	}

	if functionName == "ta.atr" {
		return &ast.Identifier{Name: "high"}, nil
	}

	if len(call.Arguments) < 1 {
		return nil, fmt.Errorf("%s requires source argument", functionName)
	}

	return call.Arguments[0], nil
}
