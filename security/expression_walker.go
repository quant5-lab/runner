package security

import "github.com/quant5-lab/runner/ast"

/* ExpressionSecurityWalker recursively walks expressions to find security() calls */
type ExpressionSecurityWalker struct {
	calls []SecurityCall
}

func NewExpressionSecurityWalker() *ExpressionSecurityWalker {
	return &ExpressionSecurityWalker{
		calls: make([]SecurityCall, 0),
	}
}

func (w *ExpressionSecurityWalker) Walk(expr ast.Expression) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.CallExpression:
		w.walkCallExpression(e)
	case *ast.ConditionalExpression:
		w.walkConditionalExpression(e)
	case *ast.BinaryExpression:
		w.walkBinaryExpression(e)
	case *ast.UnaryExpression:
		w.Walk(e.Argument)
	case *ast.MemberExpression:
		w.Walk(e.Object)
	case *ast.LogicalExpression:
		w.walkLogicalExpression(e)
	case *ast.ArrowFunctionExpression:
		w.walkArrowFunctionBody(e)
	}
}

func (w *ExpressionSecurityWalker) GetCalls() []SecurityCall {
	return w.calls
}

func (w *ExpressionSecurityWalker) walkCallExpression(call *ast.CallExpression) {
	if secCall := matchSecurityCall(call); secCall != nil {
		w.calls = append(w.calls, *secCall)
	}

	for _, arg := range call.Arguments {
		w.Walk(arg)
	}
}

func (w *ExpressionSecurityWalker) walkConditionalExpression(cond *ast.ConditionalExpression) {
	w.Walk(cond.Test)
	w.Walk(cond.Consequent)
	w.Walk(cond.Alternate)
}

func (w *ExpressionSecurityWalker) walkBinaryExpression(bin *ast.BinaryExpression) {
	w.Walk(bin.Left)
	w.Walk(bin.Right)
}

func (w *ExpressionSecurityWalker) walkLogicalExpression(logical *ast.LogicalExpression) {
	w.Walk(logical.Left)
	w.Walk(logical.Right)
}

func (w *ExpressionSecurityWalker) walkArrowFunctionBody(arrow *ast.ArrowFunctionExpression) {
	stmtWalker := NewStatementSecurityWalker(w)
	for _, stmt := range arrow.Body {
		stmtWalker.Walk(stmt)
	}
}
