package codegen

import "github.com/quant5-lab/runner/ast"

type ModeHandler struct{}

func (h *ModeHandler) CanHandle(funcName string) bool {
	return funcName == "ta.mode" || funcName == "mode"
}

func (h *ModeHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractWithDynamic(call, "ta.mode")
	if err != nil {
		return "", err
	}

	g.hasSortUsage = true

	if comp.PeriodResult.IsRuntimeDynamic() {
		dynamicGen := NewDynamicPeriodTAGenerator(g)
		code, err := dynamicGen.Generate(varName, "ta.mode", comp.SourceExpr, comp.PeriodResult)
		if err != nil {
			return "", err
		}
		return g.indentCode(comp.Preamble + code), nil
	}

	builder := NewTAIndicatorBuilder("ta.mode", varName, comp.PeriodResult.StaticValue, comp.AccessGen, comp.NeedsNaNCheck)
	builder.WithAccumulator(NewModeAccumulator())
	return g.indentCode(comp.Preamble + builder.Build()), nil
}
