package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

type PlotOptions struct {
	Variable      string
	Title         string
	ColorExpr     ast.Expression
	OffsetExpr    ast.Expression
	StyleExpr     ast.Expression
	LineWidthExpr ast.Expression
	TranspExpr    ast.Expression
	PaneExpr      ast.Expression
}

func ParsePlotOptions(call *ast.CallExpression) PlotOptions {
	opts := PlotOptions{}

	if len(call.Arguments) == 0 {
		return opts
	}

	opts.Variable = extractPlotVariable(call.Arguments[0])
	opts.Title = opts.Variable

	// Find ObjectExpression in arguments (can be at any position after first)
	var optionsObj *ast.ObjectExpression
	for i := 1; i < len(call.Arguments); i++ {
		if obj, ok := call.Arguments[i].(*ast.ObjectExpression); ok {
			optionsObj = obj
			break
		} else if lit, ok := call.Arguments[i].(*ast.Literal); ok {
			// String literal title
			if strVal, ok := lit.Value.(string); ok {
				opts.Title = strVal
			}
		}
	}

	if optionsObj != nil {
		parser := NewPropertyParser()
		if title, ok := parser.ParseString(optionsObj, "title"); ok {
			opts.Title = title
		}
		if colorExpr, ok := parser.ParseExpression(optionsObj, "color"); ok {
			opts.ColorExpr = colorExpr
		}
		// Store expression for later evaluation (handles literals and compile-time constants)
		if offsetExpr, ok := parser.ParseExpression(optionsObj, "offset"); ok {
			opts.OffsetExpr = offsetExpr
		}
		if styleExpr, ok := parser.ParseExpression(optionsObj, "style"); ok {
			opts.StyleExpr = styleExpr
		}
		if linewidthExpr, ok := parser.ParseExpression(optionsObj, "linewidth"); ok {
			opts.LineWidthExpr = linewidthExpr
		}
		if transpExpr, ok := parser.ParseExpression(optionsObj, "transp"); ok {
			opts.TranspExpr = transpExpr
		}
		if paneExpr, ok := parser.ParseExpression(optionsObj, "pane"); ok {
			opts.PaneExpr = paneExpr
		}
	}

	return opts
}

func extractPlotVariable(arg ast.Expression) string {
	switch expr := arg.(type) {
	case *ast.Identifier:
		return expr.Name
	case *ast.MemberExpression:
		if id, ok := expr.Object.(*ast.Identifier); ok {
			return id.Name
		}
	}
	return ""
}
