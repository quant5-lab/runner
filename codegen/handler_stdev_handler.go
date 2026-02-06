package codegen

import "github.com/quant5-lab/runner/ast"

type STDEVHandler struct{}

func (h *STDEVHandler) CanHandle(funcName string) bool {
	return funcName == "ta.stdev" || funcName == "stdev"
}

func (h *STDEVHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractWithDynamic(call, "ta.stdev")
	if err != nil {
		return "", err
	}

	if comp.PeriodResult.IsRuntimeDynamic() {
		dynamicGen := NewDynamicPeriodTAGenerator(g)
		code, err := dynamicGen.Generate(varName, "ta.stdev", comp.SourceExpr, comp.PeriodResult)
		if err != nil {
			return "", err
		}
		return g.indentCode(comp.Preamble + code), nil
	}

	builder := NewTAIndicatorBuilder("ta.stdev", varName, comp.PeriodResult.StaticValue, comp.AccessGen, comp.NeedsNaNCheck)
	return g.indentCode(comp.Preamble + builder.BuildSTDEV()), nil
}
