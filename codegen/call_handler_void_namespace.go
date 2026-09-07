package codegen

import (
	"strings"

	"github.com/quant5-lab/runner/ast"
)

var chartOnlyNamespaces = []string{"label", "line", "box", "table", "linefill"}

// VoidNamespaceCallHandler keeps drawing-object calls compilable while recording
// getter calls that can feed calculations.
type VoidNamespaceCallHandler struct{}

func (h *VoidNamespaceCallHandler) CanHandle(funcName string) bool {
	if isCalculationBearingChartCall(funcName) {
		return true
	}
	for _, ns := range chartOnlyNamespaces {
		if funcName == ns || strings.HasPrefix(funcName, ns+".") {
			return true
		}
	}
	return false
}

func (h *VoidNamespaceCallHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)
	if !h.CanHandle(funcName) {
		return "", nil
	}
	if isCalculationBearingChartCall(funcName) {
		g.featureGaps = append(g.featureGaps, funcName)
		return "featuregap.Record(" + quote(funcName) + `, "chart_namespace_getter", ctx.BarIndex)`, nil
	}
	return "math.NaN()", nil
}

func isCalculationBearingChartCall(funcName string) bool {
	parts := strings.Split(funcName, ".")
	if len(parts) != 2 {
		return false
	}
	switch parts[0] {
	case "label", "line", "box", "table", "linefill":
		return strings.HasPrefix(parts[1], "get_")
	default:
		return false
	}
}

func quote(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
}
