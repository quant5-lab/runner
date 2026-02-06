package codegen

import "github.com/quant5-lab/runner/ast"

type EMAHandler struct{}

func (h *EMAHandler) CanHandle(funcName string) bool {
	return funcName == "ta.ema" || funcName == "ema"
}

func (h *EMAHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractWithDynamic(call, "ta.ema")
	if err != nil {
		return "", err
	}

	if comp.PeriodResult.IsRuntimeDynamic() {
		dynamicGen := NewDynamicPeriodTAGenerator(g)
		code, err := dynamicGen.Generate(varName, "ta.ema", comp.SourceExpr, comp.PeriodResult)
		if err != nil {
			return "", err
		}
		return g.indentCode(comp.Preamble + code), nil
	}

	builder := NewTAIndicatorBuilder("ta.ema", varName, comp.PeriodResult.StaticValue, comp.AccessGen, comp.NeedsNaNCheck)
	return g.indentCode(comp.Preamble + builder.BuildEMA()), nil
}
