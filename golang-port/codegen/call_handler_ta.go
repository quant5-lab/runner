package codegen

import "github.com/quant5-lab/runner/ast"

// TAIndicatorCallHandler handles TA indicator calls in expression context.
//
// Handles: ta.sma(), ta.ema(), ta.stdev(), ta.crossover(), etc.
// Behavior: These are handled in variable declarations, not as statements
//
// Note: This is separate from TAIndicatorBuilder which generates declaration code
type TAIndicatorCallHandler struct{}

func (h *TAIndicatorCallHandler) CanHandle(funcName string) bool {
	switch funcName {
	case "ta.sma", "ta.ema", "ta.stdev", "ta.rma", "ta.wma",
		"ta.crossover", "ta.crossunder",
		"ta.change", "ta.pivothigh", "ta.pivotlow",
		"fixnan", "valuewhen":
		return true
	default:
		return false
	}
}

func (h *TAIndicatorCallHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	// TA indicator calls are handled in variable declarations
	// No immediate statement code generated
	return "", nil
}
