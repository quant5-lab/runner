package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type DynamicPeriodTAGenerator struct {
	gen *generator
}

func NewDynamicPeriodTAGenerator(gen *generator) *DynamicPeriodTAGenerator {
	return &DynamicPeriodTAGenerator{
		gen: gen,
	}
}

func (g *DynamicPeriodTAGenerator) renderPeriodExpression(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return fmt.Sprintf("%sSeries.Get(0)", e.Name)
	case *ast.Literal:
		code, _ := g.gen.generateLiteral(e)
		return code
	case *ast.BinaryExpression:
		left := g.renderPeriodExpression(e.Left)
		right := g.renderPeriodExpression(e.Right)
		return fmt.Sprintf("(%s %s %s)", left, e.Operator, right)
	case *ast.ConditionalExpression:
		cond := g.renderPeriodExpression(e.Test)
		cons := g.renderPeriodExpression(e.Consequent)
		alt := g.renderPeriodExpression(e.Alternate)
		return fmt.Sprintf("func() float64 { if %s != 0 { return %s } else { return %s } }()", cond, cons, alt)
	case *ast.CallExpression:
		code, err := g.gen.generateCallExpression(e)
		if err != nil {
			return "0"
		}
		return code
	default:
		code, err := g.gen.generateNumericExpression(expr)
		if err != nil {
			return "0"
		}
		return code
	}
}

type dynamicPeriodHandler func(
	g *DynamicPeriodTAGenerator, varName string, sourceExpr ast.Expression, periodResult PeriodEvaluationResult,
) string

/* Single source of truth — do not duplicate this list */
var dynamicPeriodDispatch = map[string]dynamicPeriodHandler{
	"ta.sma":     (*DynamicPeriodTAGenerator).generateDynamicSMA,
	"ta.ema":     (*DynamicPeriodTAGenerator).generateDynamicEMA,
	"ta.rsi":     (*DynamicPeriodTAGenerator).generateDynamicRSI,
	"ta.stdev":   (*DynamicPeriodTAGenerator).generateDynamicSTDEV,
	"ta.highest": (*DynamicPeriodTAGenerator).generateDynamicHighest,
	"ta.lowest":  (*DynamicPeriodTAGenerator).generateDynamicLowest,
	"ta.atr": func(g *DynamicPeriodTAGenerator, varName string, _ ast.Expression, periodResult PeriodEvaluationResult) string {
		return g.generateDynamicATR(varName, periodResult)
	},
}

func (g *DynamicPeriodTAGenerator) Generate(
	varName string,
	functionName string,
	sourceExpr ast.Expression,
	periodResult PeriodEvaluationResult,
) (string, error) {
	if !periodResult.IsRuntimeDynamic() {
		return "", nil
	}

	handler, ok := dynamicPeriodDispatch[functionName]
	if !ok {
		return "", fmt.Errorf("%s does not support runtime dynamic periods", functionName)
	}
	return handler(g, varName, sourceExpr, periodResult), nil
}

func (g *DynamicPeriodTAGenerator) generateDynamicSMA(
	varName string,
	sourceExpr ast.Expression,
	periodResult PeriodEvaluationResult,
) string {
	periodExprCode := g.renderPeriodExpression(periodResult.DynamicExpr)
	sourceAccessor := g.extractSourceAccessor(sourceExpr)

	code := g.gen.ind() + fmt.Sprintf("period := int(%s)\n", periodExprCode)
	code += g.gen.ind() + "if period <= 0 || ctx.BarIndex < period-1 {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.gen.indent--
	code += g.gen.ind() + "} else {\n"
	g.gen.indent++
	code += g.gen.ind() + "sum := 0.0\n"
	code += g.gen.ind() + "for j := 0; j < period; j++ {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("sum += %s.Get(j)\n", sourceAccessor)
	g.gen.indent--
	code += g.gen.ind() + "}\n"
	code += g.gen.ind() + fmt.Sprintf("%sSeries.Set(sum / float64(period))\n", varName)
	g.gen.indent--
	code += g.gen.ind() + "}\n"

	return code
}

func (g *DynamicPeriodTAGenerator) generateDynamicSTDEV(
	varName string,
	sourceExpr ast.Expression,
	periodResult PeriodEvaluationResult,
) string {
	periodExprCode := g.renderPeriodExpression(periodResult.DynamicExpr)
	sourceAccessor := g.extractSourceAccessor(sourceExpr)

	code := g.gen.ind() + fmt.Sprintf("period := int(%s)\n", periodExprCode)
	code += g.gen.ind() + "if period <= 0 || ctx.BarIndex < period-1 {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.gen.indent--
	code += g.gen.ind() + "} else {\n"
	g.gen.indent++
	code += g.gen.ind() + "sum := 0.0\n"
	code += g.gen.ind() + "for j := 0; j < period; j++ {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("sum += %s.Get(j)\n", sourceAccessor)
	g.gen.indent--
	code += g.gen.ind() + "}\n"
	code += g.gen.ind() + "mean := sum / float64(period)\n"
	code += g.gen.ind() + "variance := 0.0\n"
	code += g.gen.ind() + "for j := 0; j < period; j++ {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("diff := %s.Get(j) - mean\n", sourceAccessor)
	code += g.gen.ind() + "variance += diff * diff\n"
	g.gen.indent--
	code += g.gen.ind() + "}\n"
	code += g.gen.ind() + fmt.Sprintf("%sSeries.Set(math.Sqrt(variance / float64(period)))\n", varName)
	g.gen.indent--
	code += g.gen.ind() + "}\n"

	return code
}

func (g *DynamicPeriodTAGenerator) generateDynamicHighest(
	varName string,
	sourceExpr ast.Expression,
	periodResult PeriodEvaluationResult,
) string {
	periodExprCode := g.renderPeriodExpression(periodResult.DynamicExpr)
	sourceAccessor := g.extractSourceAccessor(sourceExpr)

	code := g.gen.ind() + fmt.Sprintf("period := int(%s)\n", periodExprCode)
	code += g.gen.ind() + "if period <= 0 || ctx.BarIndex < period-1 {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.gen.indent--
	code += g.gen.ind() + "} else {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("maxVal := %s.Get(0)\n", sourceAccessor)
	code += g.gen.ind() + "for j := 1; j < period; j++ {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("val := %s.Get(j)\n", sourceAccessor)
	code += g.gen.ind() + "if val > maxVal {\n"
	g.gen.indent++
	code += g.gen.ind() + "maxVal = val\n"
	g.gen.indent--
	code += g.gen.ind() + "}\n"
	g.gen.indent--
	code += g.gen.ind() + "}\n"
	code += g.gen.ind() + fmt.Sprintf("%sSeries.Set(maxVal)\n", varName)
	g.gen.indent--
	code += g.gen.ind() + "}\n"

	return code
}

func (g *DynamicPeriodTAGenerator) generateDynamicLowest(
	varName string,
	sourceExpr ast.Expression,
	periodResult PeriodEvaluationResult,
) string {
	periodExprCode := g.renderPeriodExpression(periodResult.DynamicExpr)
	sourceAccessor := g.extractSourceAccessor(sourceExpr)

	code := g.gen.ind() + fmt.Sprintf("period := int(%s)\n", periodExprCode)
	code += g.gen.ind() + "if period <= 0 || ctx.BarIndex < period-1 {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.gen.indent--
	code += g.gen.ind() + "} else {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("minVal := %s.Get(0)\n", sourceAccessor)
	code += g.gen.ind() + "for j := 1; j < period; j++ {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("val := %s.Get(j)\n", sourceAccessor)
	code += g.gen.ind() + "if val < minVal {\n"
	g.gen.indent++
	code += g.gen.ind() + "minVal = val\n"
	g.gen.indent--
	code += g.gen.ind() + "}\n"
	g.gen.indent--
	code += g.gen.ind() + "}\n"
	code += g.gen.ind() + fmt.Sprintf("%sSeries.Set(minVal)\n", varName)
	g.gen.indent--
	code += g.gen.ind() + "}\n"

	return code
}

func (g *DynamicPeriodTAGenerator) extractSourceAccessor(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Name + "Series"
	case *ast.MemberExpression:
		if obj, ok := e.Object.(*ast.Identifier); ok {
			if prop, ok := e.Property.(*ast.Identifier); ok {
				return obj.Name + prop.Name + "Series"
			}
		}
	}
	return "closeSeries"
}

func (g *DynamicPeriodTAGenerator) generateDynamicEMA(
	varName string,
	sourceExpr ast.Expression,
	periodResult PeriodEvaluationResult,
) string {
	periodExprCode := g.renderPeriodExpression(periodResult.DynamicExpr)
	sourceAccessor := g.extractSourceAccessor(sourceExpr)

	code := g.gen.ind() + fmt.Sprintf("period := int(%s)\n", periodExprCode)
	code += g.gen.ind() + "if period <= 0 {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.gen.indent--
	code += g.gen.ind() + "} else {\n"
	g.gen.indent++
	code += g.gen.ind() + "alpha := 2.0 / (float64(period) + 1.0)\n"
	code += g.gen.ind() + fmt.Sprintf("src := %s.Get(0)\n", sourceAccessor)
	code += g.gen.ind() + "if ctx.BarIndex == 0 {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("%sSeries.Set(src)\n", varName)
	g.gen.indent--
	code += g.gen.ind() + "} else {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("prev := %sSeries.Get(1)\n", varName)
	code += g.gen.ind() + "if math.IsNaN(prev) {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("%sSeries.Set(src)\n", varName)
	g.gen.indent--
	code += g.gen.ind() + "} else {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("%sSeries.Set(alpha*src + (1.0-alpha)*prev)\n", varName)
	g.gen.indent--
	code += g.gen.ind() + "}\n"
	g.gen.indent--
	code += g.gen.ind() + "}\n"
	g.gen.indent--
	code += g.gen.ind() + "}\n"

	return code
}

func (g *DynamicPeriodTAGenerator) generateDynamicRSI(
	varName string,
	sourceExpr ast.Expression,
	periodResult PeriodEvaluationResult,
) string {
	periodExprCode := g.renderPeriodExpression(periodResult.DynamicExpr)
	sourceAccessor := g.extractSourceAccessor(sourceExpr)

	code := g.gen.ind() + fmt.Sprintf("period := int(%s)\n", periodExprCode)
	code += g.gen.ind() + "if period <= 0 || ctx.BarIndex < period {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.gen.indent--
	code += g.gen.ind() + "} else {\n"
	g.gen.indent++
	code += g.gen.ind() + "var gains, losses float64\n"
	code += g.gen.ind() + "for j := 0; j < period; j++ {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("change := %s.Get(j) - %s.Get(j+1)\n", sourceAccessor, sourceAccessor)
	code += g.gen.ind() + "if change > 0 {\n"
	g.gen.indent++
	code += g.gen.ind() + "gains += change\n"
	g.gen.indent--
	code += g.gen.ind() + "} else {\n"
	g.gen.indent++
	code += g.gen.ind() + "losses -= change\n"
	g.gen.indent--
	code += g.gen.ind() + "}\n"
	g.gen.indent--
	code += g.gen.ind() + "}\n"
	code += g.gen.ind() + "avgGain := gains / float64(period)\n"
	code += g.gen.ind() + "avgLoss := losses / float64(period)\n"
	code += g.gen.ind() + "if avgLoss == 0 {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("%sSeries.Set(100.0)\n", varName)
	g.gen.indent--
	code += g.gen.ind() + "} else {\n"
	g.gen.indent++
	code += g.gen.ind() + "rs := avgGain / avgLoss\n"
	code += g.gen.ind() + fmt.Sprintf("%sSeries.Set(100.0 - 100.0/(1.0+rs))\n", varName)
	g.gen.indent--
	code += g.gen.ind() + "}\n"
	g.gen.indent--
	code += g.gen.ind() + "}\n"

	return code
}

func (g *DynamicPeriodTAGenerator) generateDynamicATR(
	varName string,
	periodResult PeriodEvaluationResult,
) string {
	periodExprCode := g.renderPeriodExpression(periodResult.DynamicExpr)

	code := g.gen.ind() + fmt.Sprintf("period := int(%s)\n", periodExprCode)
	code += g.gen.ind() + "if period <= 0 || ctx.BarIndex < period {\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.gen.indent--
	code += g.gen.ind() + "} else {\n"
	g.gen.indent++
	code += g.gen.ind() + "sum := 0.0\n"
	code += g.gen.ind() + "for j := 0; j < period; j++ {\n"
	g.gen.indent++
	code += g.gen.ind() + "high := bar.High(j)\n"
	code += g.gen.ind() + "low := bar.Low(j)\n"
	code += g.gen.ind() + "prevClose := bar.Close(j + 1)\n"
	code += g.gen.ind() + "tr := math.Max(high-low, math.Max(math.Abs(high-prevClose), math.Abs(low-prevClose)))\n"
	code += g.gen.ind() + "sum += tr\n"
	g.gen.indent--
	code += g.gen.ind() + "}\n"
	code += g.gen.ind() + fmt.Sprintf("%sSeries.Set(sum / float64(period))\n", varName)
	g.gen.indent--
	code += g.gen.ind() + "}\n"

	return code
}
