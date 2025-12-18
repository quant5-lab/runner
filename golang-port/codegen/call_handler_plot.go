package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// PlotFunctionHandler generates code for Pine Script plot() calls.
//
// Handles: plot()
// Generates: collector.Add() calls for visualization output
type PlotFunctionHandler struct{}

func (h *PlotFunctionHandler) CanHandle(funcName string) bool {
	return funcName == "plot"
}

func (h *PlotFunctionHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	opts := ParsePlotOptions(call)

	var plotExpr string
	if len(call.Arguments) > 0 {
		// Always use generatePlotExpression for proper builtin resolution
		exprCode, err := g.generatePlotExpression(call.Arguments[0])
		if err != nil {
			return "", err
		}
		plotExpr = exprCode
	}

	if plotExpr != "" {
		options := g.buildPlotOptions(opts)
		plotCode := fmt.Sprintf("collector.Add(%q, bar.Time, %s, %s)\n", opts.Title, plotExpr, options)
		if g.plotCollector != nil {
			g.plotCollector.AddPlot(call, plotCode)
		}
	}

	return "", nil
}
