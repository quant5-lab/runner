package codegen

import "github.com/quant5-lab/runner/ast"

type PlotStatement struct {
	node ast.Node
	code string
}

type PlotCollector struct {
	plots []PlotStatement
}

func NewPlotCollector() *PlotCollector {
	return &PlotCollector{
		plots: make([]PlotStatement, 0),
	}
}

func (pc *PlotCollector) IsPlotCall(node ast.Node) bool {
	exprStmt, ok := node.(*ast.ExpressionStatement)
	if !ok {
		return false
	}

	callExpr, ok := exprStmt.Expression.(*ast.CallExpression)
	if !ok {
		return false
	}

	identifier, ok := callExpr.Callee.(*ast.Identifier)
	if !ok {
		return false
	}

	return identifier.Name == "plot"
}

func (pc *PlotCollector) AddPlot(node ast.Node, code string) {
	pc.plots = append(pc.plots, PlotStatement{
		node: node,
		code: code,
	})
}

func (pc *PlotCollector) GetPlots() []PlotStatement {
	return pc.plots
}

func (pc *PlotCollector) HasPlots() bool {
	return len(pc.plots) > 0
}

func (pc *PlotCollector) Clear() {
	pc.plots = make([]PlotStatement, 0)
}
