package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type SeriesExpressionExtractor interface {
	ExtractSeriesExpression(expr ast.Expression) string
}

type MathExpressionGenerator struct {
	registry  *MathFunctionRegistry
	extractor SeriesExpressionExtractor
}

func NewMathExpressionGenerator(extractor SeriesExpressionExtractor) *MathExpressionGenerator {
	return &MathExpressionGenerator{
		registry:  NewMathFunctionRegistry(),
		extractor: extractor,
	}
}

func (m *MathExpressionGenerator) CanHandle(funcName string) bool {
	_, ok := m.registry.Lookup(funcName)
	return ok
}

func (m *MathExpressionGenerator) GenerateExpression(funcName string, args []ast.Expression) (string, error) {
	spec, ok := m.registry.Lookup(funcName)
	if !ok {
		return "", fmt.Errorf("unsupported math function: %s", funcName)
	}

	if err := m.validateArguments(spec, len(args)); err != nil {
		return "", err
	}

	strategy := m.createStrategy(spec)
	if strategy == nil {
		return "", fmt.Errorf("no generation strategy for %s", funcName)
	}

	return m.generateWithExtractor(strategy, args)
}

func (m *MathExpressionGenerator) validateArguments(spec *MathFunctionSpec, argCount int) error {
	if spec.MinArgs == spec.MaxArgs {
		if argCount != spec.MinArgs {
			plural := ""
			if spec.MinArgs != 1 {
				plural = "s"
			}
			return fmt.Errorf("%s requires exactly %d argument%s, got %d", spec.PineName, spec.MinArgs, plural, argCount)
		}
	} else {
		if spec.MaxArgs >= 0 && argCount > spec.MaxArgs {
			return fmt.Errorf("%s accepts at most %d arguments, got %d", spec.PineName, spec.MaxArgs, argCount)
		}
		if argCount < spec.MinArgs {
			return fmt.Errorf("%s requires at least %d arguments, got %d", spec.PineName, spec.MinArgs, argCount)
		}
	}
	return nil
}

func (m *MathExpressionGenerator) createStrategy(spec *MathFunctionSpec) SeriesAwareMathStrategy {
	switch spec.GeneratorMethod {
	case "unary":
		return NewSeriesAwareUnaryGenerator(spec.GoFunc)
	case "binary":
		return NewSeriesAwareBinaryGenerator(spec.GoFunc)
	case "sign":
		return NewSeriesAwareSignGenerator()
	case "todegrees":
		return NewSeriesAwareToDegreesGenerator()
	case "toradians":
		return NewSeriesAwareToRadiansGenerator()
	case "avg":
		return NewSeriesAwareAvgGenerator()
	case "random":
		return NewSeriesAwareRandomGenerator()
	case "round_to_mintick":
		return NewSeriesAwareRoundToMintickGenerator()
	default:
		return nil
	}
}

func (m *MathExpressionGenerator) generateWithExtractor(strategy SeriesAwareMathStrategy, args []ast.Expression) (string, error) {
	extractedArgs := make([]string, len(args))
	for i, arg := range args {
		extractedArgs[i] = m.extractor.ExtractSeriesExpression(arg)
	}
	return strategy.GenerateFromExtracted(extractedArgs)
}
