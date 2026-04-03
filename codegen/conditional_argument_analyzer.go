package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

type ConditionalArgumentInfo struct {
	Conditional *ast.ConditionalExpression
	ParentCall  *ast.CallExpression
	ArgIndex    int
	ContentHash string
}

type ConditionalArgumentAnalyzer struct {
	hasher *ExpressionHasher
}

func NewConditionalArgumentAnalyzer(hasher *ExpressionHasher) *ConditionalArgumentAnalyzer {
	return &ConditionalArgumentAnalyzer{
		hasher: hasher,
	}
}

func (a *ConditionalArgumentAnalyzer) FindInExpression(expr ast.Expression) []ConditionalArgumentInfo {
	var results []ConditionalArgumentInfo
	a.visitExpression(expr, nil, -1, &results)
	return results
}

func (a *ConditionalArgumentAnalyzer) visitExpression(
	expr ast.Expression,
	parentCall *ast.CallExpression,
	argIndex int,
	results *[]ConditionalArgumentInfo,
) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.CallExpression:
		for i, arg := range e.Arguments {
			if cond, ok := arg.(*ast.ConditionalExpression); ok {
				*results = append(*results, ConditionalArgumentInfo{
					Conditional: cond,
					ParentCall:  e,
					ArgIndex:    i,
					ContentHash: a.computeHash(cond),
				})
			}
			a.visitExpression(arg, e, i, results)
		}
		a.visitExpression(e.Callee, nil, -1, results)

	case *ast.BinaryExpression:
		a.visitExpression(e.Left, parentCall, argIndex, results)
		a.visitExpression(e.Right, parentCall, argIndex, results)

	case *ast.LogicalExpression:
		a.visitExpression(e.Left, parentCall, argIndex, results)
		a.visitExpression(e.Right, parentCall, argIndex, results)

	case *ast.UnaryExpression:
		a.visitExpression(e.Argument, parentCall, argIndex, results)

	case *ast.ConditionalExpression:
		a.visitExpression(e.Test, parentCall, argIndex, results)
		a.visitExpression(e.Consequent, parentCall, argIndex, results)
		a.visitExpression(e.Alternate, parentCall, argIndex, results)

	case *ast.MemberExpression:
		a.visitExpression(e.Object, parentCall, argIndex, results)
		if e.Computed {
			a.visitExpression(e.Property, parentCall, argIndex, results)
		}
	}
}

func (a *ConditionalArgumentAnalyzer) computeHash(cond *ast.ConditionalExpression) string {
	hash := a.hasher.Hash(cond)
	if hash == "" {
		return "00000000"
	}
	if len(hash) > 8 {
		return hash[:8]
	}
	return hash
}
