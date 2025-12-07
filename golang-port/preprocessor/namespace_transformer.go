package preprocessor

import "github.com/quant5-lab/runner/parser"

// NamespaceTransformer provides shared traversal logic for namespace-based transformations.
//
// Design Pattern: Template Method
//   - Base class defines traversal skeleton (visitStatement, visitExpression, etc.)
//   - Derived classes provide mappings (ta.*, math.*, request.*)
//   - CalleeRewriter performs actual transformation (SRP: separation of concerns)
//
// Eliminates: Code duplication across TANamespaceTransformer, MathNamespaceTransformer, RequestNamespaceTransformer
//
// Composition:
//
//	NamespaceTransformer {
//	  mappings: map[string]string      // What to transform (data)
//	  rewriter: *CalleeRewriter        // How to transform (behavior)
//	}
type NamespaceTransformer struct {
	mappings map[string]string
	rewriter *CalleeRewriter
}

// NewNamespaceTransformer creates transformer with function name mappings.
//
// Parameters:
//
//	mappings: Function name conversions (e.g., {"max": "math.max", "sma": "ta.sma"})
func NewNamespaceTransformer(mappings map[string]string) *NamespaceTransformer {
	return &NamespaceTransformer{
		mappings: mappings,
		rewriter: NewCalleeRewriter(),
	}
}

// Transform applies namespace transformations to entire script.
//
// Traversal Strategy: Depth-first recursive descent
// Idempotency: Safe to call multiple times (already-transformed nodes skipped)
func (t *NamespaceTransformer) Transform(script *parser.Script) (*parser.Script, error) {
	for _, stmt := range script.Statements {
		t.visitStatement(stmt)
	}
	return script, nil
}

func (t *NamespaceTransformer) visitStatement(stmt *parser.Statement) {
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

func (t *NamespaceTransformer) visitExpression(expr *parser.Expression) {
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

func (t *NamespaceTransformer) visitCallExpr(call *parser.CallExpr) {
	if call == nil || call.Callee == nil {
		return
	}

	if call.Callee.Ident != nil {
		t.rewriter.RewriteIfMapped(call.Callee, *call.Callee.Ident, t.mappings)
	}

	for _, arg := range call.Args {
		if arg.Value != nil {
			t.visitTernaryExpr(arg.Value)
		}
	}
}

func (t *NamespaceTransformer) visitTernaryExpr(ternary *parser.TernaryExpr) {
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

func (t *NamespaceTransformer) visitOrExpr(or *parser.OrExpr) {
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

func (t *NamespaceTransformer) visitAndExpr(and *parser.AndExpr) {
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

func (t *NamespaceTransformer) visitCompExpr(comp *parser.CompExpr) {
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

func (t *NamespaceTransformer) visitArithExpr(arith *parser.ArithExpr) {
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

func (t *NamespaceTransformer) visitTerm(term *parser.Term) {
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

func (t *NamespaceTransformer) visitFactor(factor *parser.Factor) {
	if factor == nil {
		return
	}

	if factor.Postfix != nil {
		t.visitPostfixExpr(factor.Postfix)
	}
}

func (t *NamespaceTransformer) visitPostfixExpr(postfix *parser.PostfixExpr) {
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

func (t *NamespaceTransformer) visitComparison(comp *parser.Comparison) {
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

func (t *NamespaceTransformer) visitComparisonTerm(term *parser.ComparisonTerm) {
	if term == nil {
		return
	}

	if term.Postfix != nil {
		t.visitPostfixExpr(term.Postfix)
	}
}
