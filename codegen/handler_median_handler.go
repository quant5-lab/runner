package codegen

import "github.com/quant5-lab/runner/ast"

type MedianHandler struct{}

func (h *MedianHandler) CanHandle(funcName string) bool {
	return funcName == "ta.median" || funcName == "median"
}

func (h *MedianHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractWithDynamic(call, "ta.median")
	if err != nil {
		return "", err
	}

	g.hasSortUsage = true

	if comp.PeriodResult.IsRuntimeDynamic() {
		dynamicGen := NewDynamicPeriodTAGenerator(g)
		code, err := dynamicGen.Generate(varName, "ta.median", comp.SourceExpr, comp.PeriodResult)
		if err != nil {
			return "", err
		}
		return g.indentCode(comp.Preamble + code), nil
	}

	builder := NewTAIndicatorBuilder("ta.median", varName, comp.PeriodResult.StaticValue, comp.AccessGen, comp.NeedsNaNCheck)
	builder.WithAccumulator(NewMedianAccumulator())
	return g.indentCode(comp.Preamble + builder.Build()), nil
}
