package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

var timeBuiltinNames = map[string]bool{
	"time":       true,
	"time_close": true,
}

// extractTimeBuiltin dispatches time() and time_close() before extractDefaultSeries
// so version-sensitive calls never silently degrade to featuregap.Record.
func (g *generator) extractTimeBuiltin(call *ast.CallExpression) string {
	if call == nil || g.inlineConditionRegistry == nil {
		return ""
	}
	funcName := g.extractFunctionName(call.Callee)
	if !timeBuiltinNames[funcName] {
		return ""
	}

	if funcName == "time_close" {
		return g.extractTimeClose(call)
	}

	code, err := g.inlineConditionRegistry.GenerateInline(funcName, call, g)
	if err != nil || code == "" {
		return ""
	}
	return code
}

// extractTimeClose stubs the session-argument form as a featuregap and falls back to
// the bar-close timestamp series for the no-argument form.
func (g *generator) extractTimeClose(call *ast.CallExpression) string {
	if len(call.Arguments) >= 2 {
		g.featureGaps = append(g.featureGaps, "time_close_session")
		return fmt.Sprintf("featuregap.Record(%q, %q, ctx.BarIndex)", "time_close_session", "time_series_extractor")
	}
	return "time_closeSeries.GetCurrent()"
}
