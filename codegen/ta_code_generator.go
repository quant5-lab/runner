package codegen

import (
	"fmt"

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
		return g.generateATR(varName, periodResult.StaticValue), nil

	default:
		return "", nil
	}
}

func (g *StaticPeriodTAGenerator) generateATR(varName string, period int) string {
	code := fmt.Sprintf("if ctx.BarIndex < 1 {\n")
	code += fmt.Sprintf("    %sSeries.Set(math.NaN())\n", varName)
	code += "} else {\n"
	code += "    hl := highSeries.GetCurrent() - lowSeries.GetCurrent()\n"
	code += "    hc := math.Abs(highSeries.GetCurrent() - closeSeries.Get(1))\n"
	code += "    lc := math.Abs(lowSeries.GetCurrent() - closeSeries.Get(1))\n"
	code += "    tr := math.Max(hl, math.Max(hc, lc))\n"
	code += fmt.Sprintf("    if ctx.BarIndex < %d {\n", period)
	code += fmt.Sprintf("        sum := tr\n")
	code += fmt.Sprintf("        for i := 1; i < ctx.BarIndex+1 && i < %d; i++ {\n", period)
	code += "            prevHL := highSeries.Get(i) - lowSeries.Get(i)\n"
	code += "            prevHC := math.Abs(highSeries.Get(i) - closeSeries.Get(i+1))\n"
	code += "            prevLC := math.Abs(lowSeries.Get(i) - closeSeries.Get(i+1))\n"
	code += "            sum += math.Max(prevHL, math.Max(prevHC, prevLC))\n"
	code += "        }\n"
	code += fmt.Sprintf("        %sSeries.Set(sum / float64(ctx.BarIndex+1))\n", varName)
	code += "    } else {\n"
	code += fmt.Sprintf("        prevATR := %sSeries.Get(1)\n", varName)
	code += fmt.Sprintf("        alpha := 1.0 / float64(%d)\n", period)
	code += "        newATR := alpha*tr + (1-alpha)*prevATR\n"
	code += fmt.Sprintf("        %sSeries.Set(newATR)\n", varName)
	code += "    }\n"
	code += "}\n"
	return code
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
