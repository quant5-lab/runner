package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* KcwHandler generates Keltner Channel Width:
 * 2 * mult * range(length) / EMA(source, length)
 * where range = ATR(length) when useTrueRange=true, or SMA(high-low, length) otherwise */
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
		extractUseTrueRangeArg(call, 3),
		context,
	)
	code := g.indentCode(builder.Build())

	return comp.Preamble + code, nil
}

func extractUseTrueRangeArg(call *ast.CallExpression, argIndex int) bool {
	if len(call.Arguments) <= argIndex {
		return true
	}
	lit, ok := call.Arguments[argIndex].(*ast.Literal)
	if !ok {
		return true
	}
	boolVal, ok := lit.Value.(bool)
	if !ok {
		return true
	}
	return boolVal
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

type KCWIIFEGenerator struct {
	multExpr     string
	useTrueRange bool
}

func (g *KCWIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string {
	if !period.IsConstant() {
		return "math.NaN()"
	}

	context := NewArrowFunctionIndicatorContext()
	resultName := fmt.Sprintf("kcw_%s", sourceHash)

	builder := NewKCWIndicatorBuilder(resultName, period, accessor, g.multExpr, g.useTrueRange, context)
	statefulCode := builder.Build()
	seriesAccess := fmt.Sprintf("arrowCtx.GetOrCreateSeries(%q).Get(0)", resultName)

	return fmt.Sprintf("func() float64 {\n%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}
