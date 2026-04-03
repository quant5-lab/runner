package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/*
StatefulIndicatorDetector recursively scans AST expressions for stateful TA functions.
Ensures crossover/crossunder arguments don't contain stateful indicators requiring variable storage.
Supports arbitrary PineScript: binary ops, ternary, nested calls, etc.
*/
type StatefulIndicatorDetector struct {
	statefulFunctions map[string]bool
}

func NewStatefulIndicatorDetector() *StatefulIndicatorDetector {
	return &StatefulIndicatorDetector{
		statefulFunctions: map[string]bool{
			"ta.ema": true,
			"ta.rma": true,
			"ema":    true,
			"rma":    true,
		},
	}
}

func (d *StatefulIndicatorDetector) DetectStateful(expr ast.Expression, g *generator) error {
	switch e := expr.(type) {
	case *ast.Identifier:
		return nil

	case *ast.MemberExpression:
		return nil

	case *ast.Literal:
		return nil

	case *ast.BinaryExpression:
		if err := d.DetectStateful(e.Left, g); err != nil {
			return err
		}
		return d.DetectStateful(e.Right, g)

	case *ast.UnaryExpression:
		return d.DetectStateful(e.Argument, g)

	case *ast.ConditionalExpression:
		if err := d.DetectStateful(e.Test, g); err != nil {
			return err
		}
		if err := d.DetectStateful(e.Consequent, g); err != nil {
			return err
		}
		return d.DetectStateful(e.Alternate, g)

	case *ast.CallExpression:
		funcName := g.extractFunctionName(e.Callee)

		if d.isStatefulFunction(funcName) {
			return fmt.Errorf("stateful indicator %s must be assigned to variable before crossover. Example: var1 = %s(...); if ta.crossover(var1, ...)", funcName, funcName)
		}

		for _, arg := range e.Arguments {
			if err := d.DetectStateful(arg, g); err != nil {
				return err
			}
		}
		return nil

	default:
		return nil
	}
}

func (d *StatefulIndicatorDetector) isStatefulFunction(funcName string) bool {
	return d.statefulFunctions[funcName]
}
