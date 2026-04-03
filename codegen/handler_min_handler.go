package codegen

import "github.com/quant5-lab/runner/ast"

type MinHandler struct{}

func (h *MinHandler) CanHandle(funcName string) bool {
	return funcName == "ta.min"
}

func (h *MinHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractWithDynamic(call, "ta.min")
	if err != nil {
		return "", err
	}

	if comp.PeriodResult.IsRuntimeDynamic() {
		dynamicGen := NewDynamicPeriodTAGenerator(g)
		code, err := dynamicGen.Generate(varName, "ta.min", comp.SourceExpr, comp.PeriodResult)
		if err != nil {
			return "", err
		}
		return g.indentCode(comp.Preamble + code), nil
	}

	builder := NewTAIndicatorBuilder("ta.min", varName, comp.PeriodResult.StaticValue, comp.AccessGen, comp.NeedsNaNCheck)
	builder.WithAccumulator(NewMinAccumulator())
	return g.indentCode(comp.Preamble + builder.Build()), nil
}
