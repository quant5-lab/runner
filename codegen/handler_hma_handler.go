package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// HmaHandler generates Hull Moving Average code.
// Implements CompositeIndicatorMetadata because HMA requires 3 intermediate series.
type HmaHandler struct{}

func (h *HmaHandler) CanHandle(funcName string) bool {
	return funcName == "ta.hma" || funcName == "hma"
}

func (h *HmaHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.Extract(call, "ta.hma")
	if err != nil {
		return "", err
	}

	var context StatefulIndicatorContext
	if g.inArrowFunctionBody {
		context = NewArrowFunctionIndicatorContext()
	} else {
		context = NewTopLevelIndicatorContext()
	}

	builder := NewHMAIndicatorBuilder(varName, NewConstantPeriod(comp.Period), comp.AccessGen, context)
	code := g.indentCode(builder.Build())

	return comp.Preamble + code, nil
}

func (h *HmaHandler) GetInternalSeriesNames(varName string, call *ast.CallExpression) ([]string, error) {
	return []string{
		fmt.Sprintf("_%s_wma1", varName),
		fmt.Sprintf("_%s_wma2", varName),
		fmt.Sprintf("_%s_diff", varName),
	}, nil
}

// HMAIIFEGenerator generates HMA code for arrow function context.
type HMAIIFEGenerator struct{}

func (g *HMAIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string {
	if !period.IsConstant() {
		return "math.NaN()"
	}
	context := NewArrowFunctionIndicatorContext()
	resultName := fmt.Sprintf("hma_%s", sourceHash)

	builder := NewHMAIndicatorBuilder(resultName, period, accessor, context)
	statefulCode := builder.Build()
	seriesAccess := fmt.Sprintf("arrowCtx.GetOrCreateSeries(%q).Get(0)", resultName)

	return fmt.Sprintf("func() float64 {\n%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}
