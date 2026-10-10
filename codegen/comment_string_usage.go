package codegen

import (
	"strings"

	"github.com/quant5-lab/runner/ast"
)

func collectStrategyCommentIdentifiers(program *ast.Program) map[string]bool {
	names := map[string]bool{}
	if program == nil {
		return names
	}
	walkCallExpressions(program.Body, func(call *ast.CallExpression) {
		funcName := extractCallFunctionName(call)
		for _, expr := range strategyCommentExpressions(funcName, call.Arguments) {
			collectIdentifiers(expr, names)
		}
	})
	return names
}

func strategyCommentExpressions(funcName string, args []ast.Expression) []ast.Expression {
	slots := strategyAnnotationSlots(funcName, args)
	expressions := make([]ast.Expression, 0, len(slots))
	for _, slot := range slots {
		expressions = append(expressions, slot.expression)
	}
	return expressions
}

func collectIdentifiers(expr ast.Expression, out map[string]bool) {
	switch e := expr.(type) {
	case *ast.Identifier:
		out[SanitizeGoIdentifier(e.Name)] = true
	case *ast.ConditionalExpression:
		collectIdentifiers(e.Consequent, out)
		collectIdentifiers(e.Alternate, out)
	}
}

func bareCallName(expr ast.Expression) (string, bool) {
	call, ok := expr.(*ast.CallExpression)
	if !ok {
		return "", false
	}
	name := extractFunctionName(call.Callee)
	if name == "" || strings.Contains(name, ".") {
		return "", false
	}
	return name, true
}
