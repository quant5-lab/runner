package codegen

import "github.com/quant5-lab/runner/ast"

/* ArrowSecurityDetector scans arrow function AST bodies for security() calls */
type ArrowSecurityDetector struct{}

func NewArrowSecurityDetector() *ArrowSecurityDetector {
	return &ArrowSecurityDetector{}
}

func (d *ArrowSecurityDetector) ContainsSecurityCall(arrowFunc *ast.ArrowFunctionExpression) bool {
	for _, stmt := range arrowFunc.Body {
		if d.scanStatement(stmt) {
			return true
		}
	}
	return false
}

func (d *ArrowSecurityDetector) FunctionContainsSecurityCall(funcName string, program *ast.Program) bool {
	arrowFunc := d.findArrowFunction(funcName, program)
	if arrowFunc == nil {
		return false
	}
	return d.ContainsSecurityCall(arrowFunc)
}

func (d *ArrowSecurityDetector) findArrowFunction(funcName string, program *ast.Program) *ast.ArrowFunctionExpression {
	if program == nil {
		return nil
	}
	for _, stmt := range program.Body {
		varDecl, ok := stmt.(*ast.VariableDeclaration)
		if !ok {
			continue
		}
		for _, declarator := range varDecl.Declarations {
			id, ok := declarator.ID.(*ast.Identifier)
			if !ok || id.Name != funcName {
				continue
			}
			if arrow, ok := declarator.Init.(*ast.ArrowFunctionExpression); ok {
				return arrow
			}
		}
	}
	return nil
}

func (d *ArrowSecurityDetector) scanStatement(stmt ast.Node) bool {
	switch s := stmt.(type) {
	case *ast.VariableDeclaration:
		for _, decl := range s.Declarations {
			if d.scanExpression(decl.Init) {
				return true
			}
		}
	case *ast.ExpressionStatement:
		return d.scanExpression(s.Expression)
	case *ast.IfStatement:
		if d.scanExpression(s.Test) {
			return true
		}
		for _, c := range s.Consequent {
			if d.scanStatement(c) {
				return true
			}
		}
		for _, a := range s.Alternate {
			if d.scanStatement(a) {
				return true
			}
		}
	case *ast.ForStatement:
		for _, b := range s.Body {
			if d.scanStatement(b) {
				return true
			}
		}
	case *ast.ForInStatement:
		for _, b := range s.Body {
			if d.scanStatement(b) {
				return true
			}
		}
	case *ast.WhileStatement:
		if d.scanExpression(s.Condition) {
			return true
		}
		for _, b := range s.Body {
			if d.scanStatement(b) {
				return true
			}
		}
	}
	return false
}

func (d *ArrowSecurityDetector) scanExpression(expr ast.Expression) bool {
	if expr == nil {
		return false
	}

	switch e := expr.(type) {
	case *ast.CallExpression:
		if isSecurityCallExpression(e) {
			return true
		}
		for _, arg := range e.Arguments {
			if d.scanExpression(arg) {
				return true
			}
		}
	case *ast.BinaryExpression:
		return d.scanExpression(e.Left) || d.scanExpression(e.Right)
	case *ast.LogicalExpression:
		return d.scanExpression(e.Left) || d.scanExpression(e.Right)
	case *ast.ConditionalExpression:
		return d.scanExpression(e.Test) || d.scanExpression(e.Consequent) || d.scanExpression(e.Alternate)
	case *ast.UnaryExpression:
		return d.scanExpression(e.Argument)
	}
	return false
}

func isSecurityCallExpression(call *ast.CallExpression) bool {
	funcName := extractCallFunctionName(call)
	return funcName == "request.security" || funcName == "security"
}
