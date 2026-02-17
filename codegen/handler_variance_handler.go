package codegen

import "github.com/quant5-lab/runner/ast"

type VarianceHandler struct{}

func (h *VarianceHandler) CanHandle(funcName string) bool {
	return funcName == "ta.variance" || funcName == "variance"
}

func (h *VarianceHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractWithDynamic(call, "ta.variance")
	if err != nil {
		return "", err
	}

	if comp.PeriodResult.IsRuntimeDynamic() {
		dynamicGen := NewDynamicPeriodTAGenerator(g)
		code, err := dynamicGen.Generate(varName, "ta.variance", comp.SourceExpr, comp.PeriodResult)
		if err != nil {
			return "", err
		}
		return g.indentCode(comp.Preamble + code), nil
	}

	builder := NewTAIndicatorBuilder("ta.variance", varName, comp.PeriodResult.StaticValue, comp.AccessGen, comp.NeedsNaNCheck)
	builder.WithAccumulator(NewVarianceAccumulatorForTA())
	return g.indentCode(comp.Preamble + builder.Build()), nil
}
