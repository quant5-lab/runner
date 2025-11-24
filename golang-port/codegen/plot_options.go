package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

type PlotOptions struct {
	Variable string
	Title    string
}

func ParsePlotOptions(call *ast.CallExpression) PlotOptions {
	opts := PlotOptions{}

	if len(call.Arguments) == 0 {
		return opts
	}

	opts.Variable = extractPlotVariable(call.Arguments[0])
	opts.Title = opts.Variable

	if len(call.Arguments) > 1 {
		if obj, ok := call.Arguments[1].(*ast.ObjectExpression); ok {
			parser := NewPropertyParser()
			if title, ok := parser.ParseString(obj, "title"); ok {
				opts.Title = title
			}
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
