package security

import (
	"strings"

	"github.com/borisquantlab/pinescript-go/ast"
)

/* SecurityCall represents a detected request.security() invocation */
type SecurityCall struct {
	Symbol     string          /* Symbol parameter (e.g., "BTCUSDT", "syminfo.tickerid") */
	Timeframe  string          /* Timeframe parameter (e.g., "1D", "1h") */
	Expression ast.Expression  /* AST node of expression argument for evaluation */
	ExprName   string          /* Optional name from array notation: [expr, "name"] */
}

/* AnalyzeAST scans Pine Script AST for request.security() calls */
func AnalyzeAST(program *ast.Program) []SecurityCall {
	var calls []SecurityCall
	
	/* Walk variable declarations looking for security() calls */
	for _, stmt := range program.Body {
		varDecl, ok := stmt.(*ast.VariableDeclaration)
		if !ok {
			continue
		}
		
		for _, declarator := range varDecl.Declarations {
			if call := extractSecurityCall(declarator.Init); call != nil {
				calls = append(calls, *call)
			}
		}
	}
	
	return calls
}

/* extractSecurityCall checks if expression is request.security() call */
func extractSecurityCall(expr ast.Expression) *SecurityCall {
	callExpr, ok := expr.(*ast.CallExpression)
	if !ok {
		return nil
	}
	
	/* Match: request.security(...) or security(...) */
	funcName := extractFunctionName(callExpr.Callee)
	if funcName != "request.security" && funcName != "security" {
		return nil
	}
	
	/* Require at least 3 arguments: symbol, timeframe, expression */
	if len(callExpr.Arguments) < 3 {
		return nil
	}
	
	return &SecurityCall{
		Symbol:     extractSymbol(callExpr.Arguments[0]),
		Timeframe:  extractTimeframe(callExpr.Arguments[1]),
		Expression: callExpr.Arguments[2],
		ExprName:   extractExpressionName(callExpr.Arguments[2]),
	}
}

/* extractFunctionName gets function name from callee */
func extractFunctionName(callee ast.Expression) string {
	switch c := callee.(type) {
	case *ast.Identifier:
		return c.Name
	case *ast.MemberExpression:
		obj := extractIdentifier(c.Object)
		prop := extractIdentifier(c.Property)
		if obj != "" && prop != "" {
			return obj + "." + prop
		}
	}
	return ""
}

/* extractSymbol gets symbol parameter value */
func extractSymbol(expr ast.Expression) string {
	/* String literal: "BTCUSDT" */
	if lit, ok := expr.(*ast.Literal); ok {
		if s, ok := lit.Value.(string); ok {
			return strings.Trim(s, "\"'")
		}
	}
	
	/* Identifier: syminfo.tickerid */
	if id, ok := expr.(*ast.Identifier); ok {
		return id.Name
	}
	
	/* Member expression: syminfo.tickerid */
	if mem, ok := expr.(*ast.MemberExpression); ok {
		obj := extractIdentifier(mem.Object)
		prop := extractIdentifier(mem.Property)
		if obj != "" && prop != "" {
			return obj + "." + prop
		}
	}
	
	return ""
}

/* extractTimeframe gets timeframe parameter value */
func extractTimeframe(expr ast.Expression) string {
	/* String literal: "1D", "1h" */
	if lit, ok := expr.(*ast.Literal); ok {
		if s, ok := lit.Value.(string); ok {
			/* Strip quotes if present */
			return strings.Trim(s, "\"'")
		}
	}
	
	/* Identifier: timeframe variable */
	if id, ok := expr.(*ast.Identifier); ok {
		return id.Name
	}
	
	return ""
}

/* extractExpressionName gets optional name from array notation */
func extractExpressionName(expr ast.Expression) string {
	/* TODO: Support array expression [expr, "name"] when parser adds support */
	/* For now, return unnamed for all expressions */
	return "unnamed"
}

/* extractIdentifier gets identifier name safely */
func extractIdentifier(expr ast.Expression) string {
	if id, ok := expr.(*ast.Identifier); ok {
		return id.Name
	}
	return ""
}
