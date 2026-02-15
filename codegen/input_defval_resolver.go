package codegen

import "github.com/quant5-lab/runner/ast"

func isInputCallExpr(call *ast.CallExpression) bool {
	if id, ok := call.Callee.(*ast.Identifier); ok {
		return id.Name == "input"
	}
	if mem, ok := call.Callee.(*ast.MemberExpression); ok {
		if obj, ok := mem.Object.(*ast.Identifier); ok {
			return obj.Name == "input"
		}
	}
	return false
}

func extractInputDefvalLiteral(call *ast.CallExpression) string {
	defvalExpr := extractDefvalExpression(call)
	if defvalExpr == nil {
		return ""
	}
	if lit, ok := defvalExpr.(*ast.Literal); ok {
		if s, ok := lit.Value.(string); ok {
			return s
		}
	}
	return ""
}
