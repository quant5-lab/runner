package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// KcwHandler generates Keltner Channel Width code.
// Implements CompositeIndicatorMetadata because KCW requires 2 intermediate series.
type KcwHandler struct{}

func (h *KcwHandler) CanHandle(funcName string) bool {
	return funcName == "ta.kcw" || funcName == "kcw"
}

func (h *KcwHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractWithDynamic(call, "ta.kcw")
	if err != nil {
		return "", err
	}

	multExpr, err := kcwMultExpr(g, call)
	if err != nil {
		return "", err
	}

	if comp.PeriodResult.IsRuntimeDynamic() {
		return "", fmt.Errorf("ta.kcw does not support runtime dynamic period")
	}

	var context StatefulIndicatorContext
	if g.inArrowFunctionBody {
		context = NewArrowFunctionIndicatorContext()
	} else {
		context = NewTopLevelIndicatorContext()
	}

	builder := NewKCWIndicatorBuilder(
		varName,
		NewConstantPeriod(comp.PeriodResult.StaticValue),
		comp.AccessGen,
		multExpr,
		context,
	)
	code := g.indentCode(builder.Build())

	return comp.Preamble + code, nil
}

func kcwMultExpr(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 3 {
		return "1.5", nil
	}
	return g.generateNumericExpression(call.Arguments[2])
}

func (h *KcwHandler) GetInternalSeriesNames(varName string, call *ast.CallExpression) ([]string, error) {
	return []string{
		fmt.Sprintf("_%s_ema", varName),
		fmt.Sprintf("_%s_atr", varName),
	}, nil
}

// KCWIIFEGenerator generates KCW code for arrow function context.
type KCWIIFEGenerator struct {
	multExpr string
}

func (g *KCWIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string {
	if !period.IsConstant() {
		return "math.NaN()"
	}

	context := NewArrowFunctionIndicatorContext()
	resultName := fmt.Sprintf("kcw_%s", sourceHash)

	builder := NewKCWIndicatorBuilder(resultName, period, accessor, g.multExpr, context)
	statefulCode := builder.Build()
	seriesAccess := fmt.Sprintf("arrowCtx.GetOrCreateSeries(%q).Get(0)", resultName)

	return fmt.Sprintf("func() float64 {\n%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}
