package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

type StaticPeriodTAGenerator struct {
	indicatorBuilder func(
		name string,
		varName string,
		period int,
		accessor AccessGenerator,
		needsNaN bool,
	) *TAIndicatorBuilder
	accessorFactory func(string) AccessGenerator
}

func NewStaticPeriodTAGenerator() *StaticPeriodTAGenerator {
	return &StaticPeriodTAGenerator{
		indicatorBuilder: func(name, varName string, period int, accessor AccessGenerator, needsNaN bool) *TAIndicatorBuilder {
			return NewTAIndicatorBuilder(name, varName, period, accessor, needsNaN)
		},
		accessorFactory: func(sourceExpr string) AccessGenerator {
			classifier := NewSeriesSourceClassifier()
			sourceInfo := classifier.Classify(sourceExpr)
			return CreateAccessGenerator(sourceInfo)
		},
	}
}

func (g *StaticPeriodTAGenerator) Generate(
	varName string,
	functionName string,
	sourceExpr ast.Expression,
	periodResult PeriodEvaluationResult,
) (string, error) {
	if !periodResult.IsCompileTimeConstant() {
		return "", nil
	}

	sourceExprStr := g.extractSourceExpression(sourceExpr)
	accessor := g.accessorFactory(sourceExprStr)

	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify(sourceExprStr)
	needsNaN := sourceInfo.IsSeriesVariable()

	switch functionName {
	case "ta.sma":
		builder := g.indicatorBuilder(functionName, varName, periodResult.StaticValue, accessor, needsNaN)
		builder.WithAccumulator(NewSumAccumulator())
		return builder.Build(), nil

	case "ta.ema":
		builder := g.indicatorBuilder(functionName, varName, periodResult.StaticValue, accessor, needsNaN)
		return builder.BuildEMA(), nil

	case "ta.stdev":
		builder := g.indicatorBuilder(functionName, varName, periodResult.StaticValue, accessor, needsNaN)
		return builder.BuildSTDEV(), nil

	case "ta.atr":
		builder := g.indicatorBuilder("ta.atr", varName, periodResult.StaticValue, NewTrueRangeAccessGenerator(), false)
		return builder.BuildRMA(), nil

	default:
		return "", nil
	}
}

func (g *StaticPeriodTAGenerator) extractSourceExpression(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Name
	case *ast.MemberExpression:
		if obj, ok := e.Object.(*ast.Identifier); ok {
			if prop, ok := e.Property.(*ast.Identifier); ok {
				return obj.Name + "." + prop.Name
			}
		}
	}
	return "close"
}
