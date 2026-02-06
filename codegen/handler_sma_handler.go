package codegen

import "github.com/quant5-lab/runner/ast"

type SMAHandler struct{}

func (h *SMAHandler) CanHandle(funcName string) bool {
	return funcName == "ta.sma" || funcName == "sma"
}

func (h *SMAHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractWithDynamic(call, "ta.sma")
	if err != nil {
		return "", err
	}

	if comp.PeriodResult.IsRuntimeDynamic() {
		dynamicGen := NewDynamicPeriodTAGenerator(g)
		code, err := dynamicGen.Generate(varName, "ta.sma", comp.SourceExpr, comp.PeriodResult)
		if err != nil {
			return "", err
		}
		return g.indentCode(comp.Preamble + code), nil
	}

	builder := NewTAIndicatorBuilder("ta.sma", varName, comp.PeriodResult.StaticValue, comp.AccessGen, comp.NeedsNaNCheck)
	builder.WithAccumulator(NewSumAccumulator())
	return g.indentCode(comp.Preamble + builder.Build()), nil
}
