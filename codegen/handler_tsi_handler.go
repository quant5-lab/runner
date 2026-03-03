package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* TsiHandler generates ta.tsi(source, short, long) = 100 * ema2(mom) / ema2(|mom|) */
type TsiHandler struct{}

func (h *TsiHandler) CanHandle(funcName string) bool {
	return funcName == "ta.tsi" || funcName == "tsi"
}

func (h *TsiHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 3 {
		return "", fmt.Errorf("ta.tsi requires 3 arguments: source, shortLength, longLength")
	}

	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.Extract(call, "ta.tsi")
	if err != nil {
		return "", err
	}
	shortLength := comp.Period

	longLength, err := extractor.ExtractConstantPeriodAt(call, 2, "ta.tsi")
	if err != nil {
		return "", err
	}

	var context StatefulIndicatorContext
	if g.inArrowFunctionBody {
		context = NewArrowFunctionIndicatorContext()
	} else {
		context = NewTopLevelIndicatorContext()
	}

	builder := NewTSIIndicatorBuilder(
		varName,
		NewConstantPeriod(shortLength),
		NewConstantPeriod(longLength),
		comp.AccessGen,
		context,
	)
	code := g.indentCode(builder.Build())

	return comp.Preamble + code, nil
}

func (h *TsiHandler) GetInternalSeriesNames(varName string, call *ast.CallExpression) ([]string, error) {
	return []string{
		fmt.Sprintf("_%s_mom", varName),
		fmt.Sprintf("_%s_mom_abs", varName),
		fmt.Sprintf("_%s_ema1_mom", varName),
		fmt.Sprintf("_%s_ema1_abs", varName),
		fmt.Sprintf("_%s_ema2_mom", varName),
		fmt.Sprintf("_%s_ema2_abs", varName),
	}, nil
}
