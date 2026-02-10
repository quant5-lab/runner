package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* LoopNestingValidator rejects break/continue outside loop bodies */
type LoopNestingValidator struct{}

func NewLoopNestingValidator() *LoopNestingValidator {
	return &LoopNestingValidator{}
}

func (v *LoopNestingValidator) Validate(program *ast.Program) error {
	for _, node := range program.Body {
		if err := v.validateNode(node, false); err != nil {
			return err
		}
	}
	return nil
}

func (v *LoopNestingValidator) validateNode(node ast.Node, inLoop bool) error {
	switch n := node.(type) {
	case *ast.BreakStatement:
		if !inLoop {
			return fmt.Errorf("break statement outside loop body")
		}
	case *ast.ContinueStatement:
		if !inLoop {
			return fmt.Errorf("continue statement outside loop body")
		}
	case *ast.ForStatement:
		return v.validateBody(n.Body, true)
	case *ast.ForInStatement:
		return v.validateBody(n.Body, true)
	case *ast.IfStatement:
		if err := v.validateBody(n.Consequent, inLoop); err != nil {
			return err
		}
		return v.validateBody(n.Alternate, inLoop)
	case *ast.ArrowFunctionExpression:
		return v.validateBody(n.Body, false)
	case *ast.ExpressionStatement:
		return v.validateExpression(n.Expression, inLoop)
	case *ast.VariableDeclaration:
		for _, decl := range n.Declarations {
			if decl.Init != nil {
				if err := v.validateExpression(decl.Init, inLoop); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (v *LoopNestingValidator) validateExpression(expr ast.Expression, inLoop bool) error {
	switch e := expr.(type) {
	case *ast.ForStatement:
		return v.validateBody(e.Body, true)
	case *ast.ForInStatement:
		return v.validateBody(e.Body, true)
	case *ast.IfStatement:
		if err := v.validateBody(e.Consequent, inLoop); err != nil {
			return err
		}
		return v.validateBody(e.Alternate, inLoop)
	case *ast.ArrowFunctionExpression:
		return v.validateBody(e.Body, false)
	case *ast.CallExpression:
		for _, arg := range e.Arguments {
			if err := v.validateExpression(arg, inLoop); err != nil {
				return err
			}
		}
	case *ast.BinaryExpression:
		if err := v.validateExpression(e.Left, inLoop); err != nil {
			return err
		}
		return v.validateExpression(e.Right, inLoop)
	case *ast.LogicalExpression:
		if err := v.validateExpression(e.Left, inLoop); err != nil {
			return err
		}
		return v.validateExpression(e.Right, inLoop)
	case *ast.UnaryExpression:
		return v.validateExpression(e.Argument, inLoop)
	case *ast.ConditionalExpression:
		if err := v.validateExpression(e.Test, inLoop); err != nil {
			return err
		}
		if err := v.validateExpression(e.Consequent, inLoop); err != nil {
			return err
		}
		return v.validateExpression(e.Alternate, inLoop)
	}
	return nil
}

func (v *LoopNestingValidator) validateBody(nodes []ast.Node, inLoop bool) error {
	for _, node := range nodes {
		if err := v.validateNode(node, inLoop); err != nil {
			return err
		}
	}
	return nil
}
