package security

import "github.com/quant5-lab/runner/ast"

// HistoricalOffsetExtractor extracts historical lookback offset from AST expressions
// Handles patterns: expr[1], expr[2], nested: fixnan(pivothigh()[1])
type HistoricalOffsetExtractor struct{}

func NewHistoricalOffsetExtractor() *HistoricalOffsetExtractor {
	return &HistoricalOffsetExtractor{}
}

// Extract returns (innerExpression, offset) or (originalExpression, 0) if no offset
// Example: pivothigh(5,5)[1] → (pivothigh(5,5), 1)
func (e *HistoricalOffsetExtractor) Extract(expr ast.Expression) (ast.Expression, int) {
	memberExpr, isMember := expr.(*ast.MemberExpression)
	if !isMember {
		return expr, 0
	}

	offsetLit, isLiteral := memberExpr.Property.(*ast.Literal)
	if !isLiteral {
		return expr, 0
	}

	offsetValue, isFloat := offsetLit.Value.(float64)
	if !isFloat {
		return expr, 0
	}

	return memberExpr.Object, int(offsetValue)
}

// ExtractRecursive handles nested patterns: fixnan(pivothigh()[1])
// Returns deepest inner expression and accumulated offset
func (e *HistoricalOffsetExtractor) ExtractRecursive(expr ast.Expression) (ast.Expression, int) {
	switch exp := expr.(type) {
	case *ast.MemberExpression:
		// Direct subscript: expr[N]
		inner, offset := e.Extract(expr)
		if offset > 0 {
			return inner, offset
		}
		return expr, 0

	case *ast.CallExpression:
		// Check if any argument contains subscripted expression
		for i, arg := range exp.Arguments {
			innerArg, offset := e.ExtractRecursive(arg)
			if offset > 0 {
				// Rebuild call with inner argument (without subscript)
				newArgs := make([]ast.Expression, len(exp.Arguments))
				copy(newArgs, exp.Arguments)
				newArgs[i] = innerArg
				newCall := &ast.CallExpression{
					Callee:    exp.Callee,
					Arguments: newArgs,
				}
				return newCall, offset
			}
		}
		return expr, 0

	default:
		return expr, 0
	}
}
