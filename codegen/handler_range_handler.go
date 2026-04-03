package codegen

import "github.com/quant5-lab/runner/ast"

type RangeHandler struct{}

func (h *RangeHandler) CanHandle(funcName string) bool {
	return funcName == "ta.range" || funcName == "range"
}

func (h *RangeHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractWithDynamic(call, "ta.range")
	if err != nil {
		return "", err
	}

	if comp.PeriodResult.IsRuntimeDynamic() {
		dynamicGen := NewDynamicPeriodTAGenerator(g)
		code, err := dynamicGen.Generate(varName, "ta.range", comp.SourceExpr, comp.PeriodResult)
		if err != nil {
			return "", err
		}
		return g.indentCode(comp.Preamble + code), nil
	}

	builder := NewTAIndicatorBuilder("ta.range", varName, comp.PeriodResult.StaticValue, comp.AccessGen, comp.NeedsNaNCheck)
	builder.WithAccumulator(NewRangeAccumulator())
	return g.indentCode(comp.Preamble + builder.Build()), nil
}
