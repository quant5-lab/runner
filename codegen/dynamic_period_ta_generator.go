package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type DynamicPeriodEmitter interface {
	EmitCalculation(g *generator, varName, sourceAccessor string) string
}

var dynamicPeriodDispatch = map[string]DynamicPeriodEmitter{
	"ta.sma":     DynamicSMAEmitter{},
	"ta.ema":     DynamicEMAEmitter{},
	"ta.rsi":     DynamicRSIEmitter{},
	"ta.stdev":   DynamicSTDEVEmitter{},
	"ta.highest": DynamicHighestEmitter{},
	"ta.lowest":  DynamicLowestEmitter{},
	"ta.atr":     DynamicATREmitter{},
	"ta.cci":     DynamicCCIEmitter{},
	"ta.cog":     DynamicCOGEmitter{},
}

type DynamicPeriodTAGenerator struct {
	gen *generator
}

func NewDynamicPeriodTAGenerator(gen *generator) *DynamicPeriodTAGenerator {
	return &DynamicPeriodTAGenerator{gen: gen}
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

	emitter, ok := dynamicPeriodDispatch[functionName]
	if !ok {
		return "", fmt.Errorf("%s does not support runtime dynamic periods", functionName)
	}

	periodExpr := g.renderPeriodExpression(periodResult.DynamicExpr)
	sourceAccessor := g.extractSourceAccessor(sourceExpr)

	code := g.gen.ind() + "{\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("period := int(%s)\n", periodExpr)
	code += emitter.EmitCalculation(g.gen, varName, sourceAccessor)
	g.gen.indent--
	code += g.gen.ind() + "}\n"

	return code, nil
}

/* GenerateWithEmitter produces runtime-dynamic period code using a caller-supplied emitter.
 * Use when the emitter requires construction-time parameters not expressible in the shared
 * EmitCalculation signature (e.g. DynamicBBWEmitter captures mult). */
func (g *DynamicPeriodTAGenerator) GenerateWithEmitter(
	varName string,
	emitter DynamicPeriodEmitter,
	sourceExpr ast.Expression,
	periodResult PeriodEvaluationResult,
) (string, error) {
	if !periodResult.IsRuntimeDynamic() {
		return "", nil
	}

	periodExpr := g.renderPeriodExpression(periodResult.DynamicExpr)
	sourceAccessor := g.extractSourceAccessor(sourceExpr)

	code := g.gen.ind() + "{\n"
	g.gen.indent++
	code += g.gen.ind() + fmt.Sprintf("period := int(%s)\n", periodExpr)
	code += emitter.EmitCalculation(g.gen, varName, sourceAccessor)
	g.gen.indent--
	code += g.gen.ind() + "}\n"

	return code, nil
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

func (g *DynamicPeriodTAGenerator) extractSourceAccessor(expr ast.Expression) string {
	if expr == nil {
		return ""
	}
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
