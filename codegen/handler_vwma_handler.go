package codegen

import "github.com/quant5-lab/runner/ast"

type VWMAHandler struct{}

func (h *VWMAHandler) CanHandle(funcName string) bool {
	return funcName == "ta.vwma" || funcName == "vwma"
}

func (h *VWMAHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractWithDynamic(call, "ta.vwma")
	if err != nil {
		return "", err
	}

	if comp.PeriodResult.IsRuntimeDynamic() {
		dynamicGen := NewDynamicPeriodTAGenerator(g)
		code, err := dynamicGen.Generate(varName, "ta.vwma", comp.SourceExpr, comp.PeriodResult)
		if err != nil {
			return "", err
		}
		return g.indentCode(comp.Preamble + code), nil
	}

	builder := NewTAIndicatorBuilder("ta.vwma", varName, comp.PeriodResult.StaticValue, comp.AccessGen, comp.NeedsNaNCheck)
	builder.WithAccumulator(NewVolumeWeightedSumAccumulator())
	return g.indentCode(comp.Preamble + builder.Build()), nil
}
