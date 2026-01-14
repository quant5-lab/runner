package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/*
SubscriptResolver handles variable subscripts like src[nA] where index is computed.

Design: Generate bounds-checked dynamic series access.
Rationale: Safety-first approach prevents runtime panics.
*/
type SubscriptResolver struct{}

func NewSubscriptResolver() *SubscriptResolver {
	return &SubscriptResolver{}
}

/*
ResolveSubscript generates code for series[expression] access.

Returns: Go expression string for series access
*/
func (sr *SubscriptResolver) ResolveSubscript(seriesName string, indexExpr ast.Expression, g *generator) string {
	// Check if seriesName is an input.source alias
	if funcName, isConstant := g.constants[seriesName]; isConstant && funcName == "input.source" {
		// For input.source, treat it as an alias to close (default)
		// TODO: Extract actual source from input.source defval
		seriesName = "close"
	}

	// Check if index is a literal (fast path)
	if lit, ok := indexExpr.(*ast.Literal); ok {
		if floatVal, ok := lit.Value.(float64); ok {
			intVal := int(floatVal)

			// For built-in series, use ctx.Data access
			if seriesName == "close" || seriesName == "open" || seriesName == "high" || seriesName == "low" || seriesName == "volume" {
				if intVal == 0 {
					return fmt.Sprintf("bar.%s", capitalize(seriesName))
				}
				return fmt.Sprintf("ctx.Data[i-%d].%s", intVal, capitalize(seriesName))
			}

			return fmt.Sprintf("%sSeries.Get(%d)", seriesName, intVal)
		}
	}

	// Variable index - evaluate expression using generator's extractSeriesExpression
	indexCode := g.extractSeriesExpression(indexExpr)

	// For built-in series with variable index, need to use ctx.Data[i-index]
	if seriesName == "close" || seriesName == "open" || seriesName == "high" || seriesName == "low" || seriesName == "volume" {
		// Generate bounds-checked access to ctx.Data
		return fmt.Sprintf("func() float64 { idx := i - int(%s); if idx >= 0 && idx < len(ctx.Data) { return ctx.Data[idx].%s } else { return math.NaN() } }()", indexCode, capitalize(seriesName))
	}

	// Generate dynamic access for user-defined series
	return fmt.Sprintf("%sSeries.Get(int(%s))", seriesName, indexCode)
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}
