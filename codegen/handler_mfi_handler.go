package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* MFIHandler generates inline code for Money Flow Index calculations.
 * Delegates to MFIIndicatorBuilder for code generation. */
type MFIHandler struct{}

func (h *MFIHandler) CanHandle(funcName string) bool {
	return funcName == "ta.mfi" || funcName == "mfi"
}

func (h *MFIHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractWithDynamic(call, "ta.mfi")
	if err != nil {
		return "", err
	}

	if comp.PeriodResult.IsFailed() {
		return "", fmt.Errorf("ta.mfi: %s", comp.PeriodResult.FailureReason)
	}

	if comp.PeriodResult.IsRuntimeDynamic() {
		return "", fmt.Errorf("ta.mfi: runtime dynamic period not yet supported")
	}

	context := NewTopLevelIndicatorContext()
	builder := NewMFIIndicatorBuilder(varName, NewConstantPeriod(comp.PeriodResult.StaticValue), comp.AccessGen, context)
	return g.indentCode(comp.Preamble + builder.Build()), nil
}

func (h *MFIHandler) GetInternalSeriesNames(varName string, call *ast.CallExpression) ([]string, error) {
	return []string{
		fmt.Sprintf("_%s_positive_mf", varName),
		fmt.Sprintf("_%s_negative_mf", varName),
	}, nil
}
