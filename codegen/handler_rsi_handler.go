package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type RSIHandler struct{}

func (h *RSIHandler) CanHandle(funcName string) bool {
	return funcName == "ta.rsi" || funcName == "rsi"
}

func (h *RSIHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	extractor := NewTAArgumentExtractor(g)
	comp, err := extractor.ExtractWithDynamic(call, "ta.rsi")
	if err != nil {
		return "", err
	}

	if comp.PeriodResult.IsFailed() {
		return "", fmt.Errorf("ta.rsi: %s", comp.PeriodResult.FailureReason)
	}

	if comp.PeriodResult.IsRuntimeDynamic() {
		return "", fmt.Errorf("ta.rsi period must be compile-time constant (PineScript requires simple int)")
	}

	code, err := g.generateRSI(varName, comp.PeriodResult.StaticValue, comp.AccessGen, comp.NeedsNaNCheck)
	if err != nil {
		return "", err
	}
	return comp.Preamble + code, nil
}

/* GetInternalSeriesNames implements CompositeIndicatorMetadata interface.
 * RSI requires 4 internal series for gains, losses, and their RMA smoothing.
 */
func (h *RSIHandler) GetInternalSeriesNames(varName string, call *ast.CallExpression) ([]string, error) {
	return []string{
		fmt.Sprintf("_%s_gains", varName),
		fmt.Sprintf("_%s_losses", varName),
		fmt.Sprintf("_%s_rma_gains", varName),
		fmt.Sprintf("_%s_rma_losses", varName),
	}, nil
}
