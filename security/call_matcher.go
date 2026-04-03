package security

import (
	"strings"

	"github.com/quant5-lab/runner/ast"
)

/* SecurityCallMatcher extracts security call details from CallExpression */
type SecurityCallMatcher struct{}

func NewSecurityCallMatcher() *SecurityCallMatcher {
	return &SecurityCallMatcher{}
}

func (m *SecurityCallMatcher) Match(call *ast.CallExpression) *SecurityCall {
	return matchSecurityCall(call)
}

func matchSecurityCall(call *ast.CallExpression) *SecurityCall {
	if call == nil {
		return nil
	}

	funcName := extractFunctionName(call.Callee)
	if funcName != "request.security" && funcName != "security" {
		return nil
	}

	if len(call.Arguments) < 3 {
		return nil
	}

	return &SecurityCall{
		Symbol:        extractSymbol(call.Arguments[0]),
		Timeframe:     extractTimeframe(call.Arguments[1]),
		Expression:    call.Arguments[2],
		ExprName:      extractExpressionName(call.Arguments[2]),
		SymbolExpr:    call.Arguments[0],
		TimeframeExpr: call.Arguments[1],
	}
}

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

func extractSymbol(expr ast.Expression) string {
	extractor := NewSymbolExtractor()
	/* Return raw symbol including modifier prefix for cache key matching */
	return extractor.extractRaw(expr)
}

func extractTimeframe(expr ast.Expression) string {
	if lit, ok := expr.(*ast.Literal); ok {
		if s, ok := lit.Value.(string); ok {
			return strings.Trim(s, "\"'")
		}
	}

	if id, ok := expr.(*ast.Identifier); ok {
		return id.Name
	}

	if mem, ok := expr.(*ast.MemberExpression); ok {
		obj := extractIdentifier(mem.Object)
		prop := extractIdentifier(mem.Property)
		if obj != "" && prop != "" {
			return obj + "." + prop
		}
	}

	return ""
}

func extractExpressionName(expr ast.Expression) string {
	return "unnamed"
}

func extractIdentifier(expr ast.Expression) string {
	if id, ok := expr.(*ast.Identifier); ok {
		return id.Name
	}
	return ""
}

/* ExtractMaxPeriod analyzes expression to find maximum indicator period needed */
func ExtractMaxPeriod(expr ast.Expression) int {
	if expr == nil {
		return 0
	}

	switch e := expr.(type) {
	case *ast.CallExpression:
		funcName := extractFunctionName(e.Callee)
		maxPeriod := 0

		if strings.HasPrefix(funcName, "ta.") && len(e.Arguments) >= 2 {
			if lit, ok := e.Arguments[1].(*ast.Literal); ok {
				if period, ok := lit.Value.(float64); ok {
					maxPeriod = int(period)
				}
			}
		}

		for _, arg := range e.Arguments {
			argPeriod := ExtractMaxPeriod(arg)
			if argPeriod > maxPeriod {
				maxPeriod = argPeriod
			}
		}

		return maxPeriod

	case *ast.BinaryExpression:
		leftPeriod := ExtractMaxPeriod(e.Left)
		rightPeriod := ExtractMaxPeriod(e.Right)
		if leftPeriod > rightPeriod {
			return leftPeriod
		}
		return rightPeriod

	case *ast.ConditionalExpression:
		testPeriod := ExtractMaxPeriod(e.Test)
		conseqPeriod := ExtractMaxPeriod(e.Consequent)
		altPeriod := ExtractMaxPeriod(e.Alternate)

		maxPeriod := testPeriod
		if conseqPeriod > maxPeriod {
			maxPeriod = conseqPeriod
		}
		if altPeriod > maxPeriod {
			maxPeriod = altPeriod
		}
		return maxPeriod

	case *ast.MemberExpression:
		return 0

	case *ast.Identifier:
		return 0

	case *ast.Literal:
		return 0

	default:
		return 0
	}
}
