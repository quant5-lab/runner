package preprocessor

import "github.com/borisquantlab/pinescript-go/parser"

// functionRenamer is a shared visitor for simple function name replacements
// DRY principle: reuse traversal logic across multiple transformers
type functionRenamer struct {
	mappings map[string]string
}

func (v *functionRenamer) visitStatement(stmt *parser.Statement) {
	if stmt.Assignment != nil {
		v.visitExpression(stmt.Assignment.Value)
	}
	if stmt.If != nil {
		v.visitComparison(stmt.If.Condition)
		if stmt.If.Body != nil {
			v.visitStatement(stmt.If.Body)
		}
	}
	if stmt.Expression != nil {
		v.visitExpression(stmt.Expression.Expr)
	}
}

func (v *functionRenamer) visitExpression(expr *parser.Expression) {
	if expr == nil {
		return
	}

	if expr.Call != nil {
		v.visitCallExpr(expr.Call)
	}
	if expr.Ternary != nil {
		v.visitTernaryExpr(expr.Ternary)
	}
}

func (v *functionRenamer) visitCallExpr(call *parser.CallExpr) {
	// Rename function if in mappings (only for simple identifiers)
	if call.Callee != nil && call.Callee.Ident != nil {
		if newName, ok := v.mappings[*call.Callee.Ident]; ok {
			call.Callee.Ident = &newName
		}
	}

	// Recurse into arguments
	for _, arg := range call.Args {
		if arg.Value != nil {
			v.visitTernaryExpr(arg.Value)
		}
	}
}

func (v *functionRenamer) visitTernaryExpr(ternary *parser.TernaryExpr) {
	if ternary.Condition != nil {
		v.visitOrExpr(ternary.Condition)
	}
	if ternary.TrueVal != nil {
		v.visitExpression(ternary.TrueVal)
	}
	if ternary.FalseVal != nil {
		v.visitExpression(ternary.FalseVal)
	}
}

func (v *functionRenamer) visitOrExpr(or *parser.OrExpr) {
	if or.Left != nil {
		v.visitAndExpr(or.Left)
	}
	if or.Right != nil {
		v.visitOrExpr(or.Right)
	}
}

func (v *functionRenamer) visitAndExpr(and *parser.AndExpr) {
	if and.Left != nil {
		v.visitCompExpr(and.Left)
	}
	if and.Right != nil {
		v.visitAndExpr(and.Right)
	}
}

func (v *functionRenamer) visitCompExpr(comp *parser.CompExpr) {
	if comp.Left != nil {
		v.visitArithExpr(comp.Left)
	}
	if comp.Right != nil {
		v.visitCompExpr(comp.Right)
	}
}

func (v *functionRenamer) visitArithExpr(arith *parser.ArithExpr) {
	if arith.Left != nil {
		v.visitTerm(arith.Left)
	}
	if arith.Right != nil {
		v.visitArithExpr(arith.Right)
	}
}

func (v *functionRenamer) visitTerm(term *parser.Term) {
	if term.Left != nil {
		v.visitFactor(term.Left)
	}
	if term.Right != nil {
		v.visitTerm(term.Right)
	}
}

func (v *functionRenamer) visitFactor(factor *parser.Factor) {
	if factor.Call != nil {
		v.visitCallExpr(factor.Call)
	}
	if factor.Subscript != nil && factor.Subscript.Index != nil {
		v.visitArithExpr(factor.Subscript.Index)
	}
}

func (v *functionRenamer) visitComparison(comp *parser.Comparison) {
	if comp.Left != nil {
		v.visitComparisonTerm(comp.Left)
	}
	if comp.Right != nil {
		v.visitComparisonTerm(comp.Right)
	}
}

func (v *functionRenamer) visitComparisonTerm(term *parser.ComparisonTerm) {
	if term.Call != nil {
		v.visitCallExpr(term.Call)
	}
	if term.Subscript != nil && term.Subscript.Index != nil {
		v.visitArithExpr(term.Subscript.Index)
	}
}

func (v *functionRenamer) visitValue(val *parser.Value) {
	if val == nil {
		return
	}
	
	if val.Subscript != nil && val.Subscript.Index != nil {
		v.visitArithExpr(val.Subscript.Index)
	}
}
