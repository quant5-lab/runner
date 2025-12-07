package preprocessor

import "github.com/quant5-lab/runner/parser"

/* SimpleRenameTransformer renames function identifiers without namespace changes */
type SimpleRenameTransformer struct {
	mappings map[string]string
}

func NewSimpleRenameTransformer(mappings map[string]string) *SimpleRenameTransformer {
	return &SimpleRenameTransformer{
		mappings: mappings,
	}
}

/* Transform renames function Idents directly (study→indicator, not study→ta.indicator) */
func (t *SimpleRenameTransformer) Transform(script *parser.Script) (*parser.Script, error) {
	for _, stmt := range script.Statements {
		t.visitStatement(stmt)
	}
	return script, nil
}

func (t *SimpleRenameTransformer) visitStatement(stmt *parser.Statement) {
	if stmt == nil {
		return
	}

	if stmt.Assignment != nil {
		t.visitExpression(stmt.Assignment.Value)
	}

	if stmt.If != nil {
		t.visitComparison(stmt.If.Condition)
		for _, bodyStmt := range stmt.If.Body {
			t.visitStatement(bodyStmt)
		}
	}

	if stmt.Expression != nil {
		t.visitExpression(stmt.Expression.Expr)
	}
}

func (t *SimpleRenameTransformer) visitExpression(expr *parser.Expression) {
	if expr == nil {
		return
	}

	if expr.Call != nil {
		t.visitCallExpr(expr.Call)
	}

	if expr.Ternary != nil {
		t.visitTernaryExpr(expr.Ternary)
	}
}

func (t *SimpleRenameTransformer) visitCallExpr(call *parser.CallExpr) {
	if call == nil || call.Callee == nil {
		return
	}

	/* Simple rename: only modify Ident, leave MemberAccess unchanged */
	if call.Callee.Ident != nil {
		funcName := *call.Callee.Ident
		if newName, exists := t.mappings[funcName]; exists {
			call.Callee.Ident = &newName
		}
	}

	/* Visit arguments */
	for _, arg := range call.Args {
		if arg.Value != nil {
			t.visitTernaryExpr(arg.Value)
		}
	}
}

func (t *SimpleRenameTransformer) visitTernaryExpr(ternary *parser.TernaryExpr) {
	if ternary == nil {
		return
	}

	if ternary.Condition != nil {
		t.visitOrExpr(ternary.Condition)
	}

	if ternary.TrueVal != nil {
		t.visitExpression(ternary.TrueVal)
	}

	if ternary.FalseVal != nil {
		t.visitExpression(ternary.FalseVal)
	}
}

func (t *SimpleRenameTransformer) visitOrExpr(or *parser.OrExpr) {
	if or == nil {
		return
	}

	if or.Left != nil {
		t.visitAndExpr(or.Left)
	}

	if or.Right != nil {
		t.visitOrExpr(or.Right)
	}
}

func (t *SimpleRenameTransformer) visitAndExpr(and *parser.AndExpr) {
	if and == nil {
		return
	}

	if and.Left != nil {
		t.visitCompExpr(and.Left)
	}

	if and.Right != nil {
		t.visitAndExpr(and.Right)
	}
}

func (t *SimpleRenameTransformer) visitCompExpr(comp *parser.CompExpr) {
	if comp == nil {
		return
	}

	if comp.Left != nil {
		t.visitArithExpr(comp.Left)
	}

	if comp.Right != nil {
		t.visitCompExpr(comp.Right)
	}
}

func (t *SimpleRenameTransformer) visitArithExpr(arith *parser.ArithExpr) {
	if arith == nil {
		return
	}

	if arith.Left != nil {
		t.visitTerm(arith.Left)
	}

	if arith.Right != nil {
		t.visitArithExpr(arith.Right)
	}
}

func (t *SimpleRenameTransformer) visitTerm(term *parser.Term) {
	if term == nil {
		return
	}

	if term.Left != nil {
		t.visitFactor(term.Left)
	}

	if term.Right != nil {
		t.visitTerm(term.Right)
	}
}

func (t *SimpleRenameTransformer) visitFactor(factor *parser.Factor) {
	if factor == nil {
		return
	}

	if factor.Postfix != nil {
		t.visitPostfixExpr(factor.Postfix)
	}
}

func (t *SimpleRenameTransformer) visitPostfixExpr(postfix *parser.PostfixExpr) {
	if postfix == nil {
		return
	}

	if postfix.Primary != nil && postfix.Primary.Call != nil {
		t.visitCallExpr(postfix.Primary.Call)
	}

	if postfix.Subscript != nil {
		t.visitArithExpr(postfix.Subscript)
	}
}

func (t *SimpleRenameTransformer) visitComparison(comp *parser.Comparison) {
	if comp == nil {
		return
	}

	if comp.Left != nil {
		t.visitComparisonTerm(comp.Left)
	}

	if comp.Right != nil {
		t.visitComparisonTerm(comp.Right)
	}
}

func (t *SimpleRenameTransformer) visitComparisonTerm(term *parser.ComparisonTerm) {
	if term == nil {
		return
	}

	if term.Postfix != nil {
		t.visitPostfixExpr(term.Postfix)
	}
}

