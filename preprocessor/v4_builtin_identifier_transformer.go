package preprocessor

import "github.com/quant5-lab/runner/parser"

// v4BuiltinIdentifierMappings are global/series identifiers that were bare in v4
// but require namespace qualification in v5.
var v4BuiltinIdentifierMappings = map[string]string{
	"tickerid":   "syminfo.tickerid",
	"ticker":     "syminfo.ticker",
	"period":     "timeframe.period",
	"interval":   "timeframe.multiplier",
	"isintraday": "timeframe.isintraday",
	"isdwm":      "timeframe.isdwm",
	"isdaily":    "timeframe.isdaily",
	"isweekly":   "timeframe.isweekly",
	"ismonthly":  "timeframe.ismonthly",
}

// V4BuiltinIdentifierTransformer qualifies bare v4 identifiers to their v5 namespace forms
// using a full depth-first CST traversal. UDF parameters shadow the global mappings so that
// a param named "period" is left untouched within that function body.
type V4BuiltinIdentifierTransformer struct {
	mappings map[string]string
}

func NewV4BuiltinIdentifierTransformer() *V4BuiltinIdentifierTransformer {
	return &V4BuiltinIdentifierTransformer{mappings: v4BuiltinIdentifierMappings}
}

func (t *V4BuiltinIdentifierTransformer) Transform(script *parser.Script) (*parser.Script, error) {
	shadow := make(map[string]bool)
	for _, stmt := range script.Statements {
		t.visitStatement(stmt, shadow)
	}
	return script, nil
}

func (t *V4BuiltinIdentifierTransformer) visitStatement(stmt *parser.Statement, shadow map[string]bool) {
	if stmt == nil || stmt.Core == nil {
		return
	}
	t.visitStatementCore(stmt.Core, shadow)
}

func (t *V4BuiltinIdentifierTransformer) visitStatementCore(core *parser.StatementCore, shadow map[string]bool) {
	if core == nil {
		return
	}
	if core.TupleAssignment != nil {
		t.visitExpression(core.TupleAssignment.Value, shadow)
	}
	if core.If != nil {
		t.visitIfStatement(core.If, shadow)
	}
	if core.ForIn != nil {
		t.visitArithExpr(core.ForIn.Collection, shadow)
		for _, s := range core.ForIn.Body {
			t.visitStatement(s, shadow)
		}
	}
	if core.For != nil {
		t.visitArithExpr(core.For.From, shadow)
		t.visitArithExpr(core.For.To, shadow)
		t.visitArithExpr(core.For.Step, shadow)
		for _, s := range core.For.Body {
			t.visitStatement(s, shadow)
		}
	}
	if core.While != nil {
		t.visitOrExpr(core.While.Condition, shadow)
		for _, s := range core.While.Body {
			t.visitStatement(s, shadow)
		}
	}
	if core.Switch != nil {
		t.visitSwitchExpr(core.Switch, shadow)
	}
	if core.FunctionDecl != nil {
		t.visitFunctionDecl(core.FunctionDecl, shadow)
	}
	if core.VarAssignment != nil {
		t.visitExpression(core.VarAssignment.Value, shadow)
		if core.VarAssignment.Name != nil {
			shadow[*core.VarAssignment.Name] = true
		}
	}
	if core.TypedAssignment != nil {
		t.visitExpression(core.TypedAssignment.Value, shadow)
		shadow[core.TypedAssignment.Name] = true
	}
	if core.Assignment != nil {
		t.visitExpression(core.Assignment.Value, shadow)
		shadow[core.Assignment.Name] = true
	}
	if core.Reassignment != nil {
		t.visitExpression(core.Reassignment.Value, shadow)
	}
	if core.Expression != nil {
		t.visitExpression(core.Expression.Expr, shadow)
	}
}

func (t *V4BuiltinIdentifierTransformer) visitIfStatement(stmt *parser.IfStatement, shadow map[string]bool) {
	if stmt == nil {
		return
	}
	t.visitOrExpr(stmt.Condition, shadow)
	for _, s := range stmt.Body {
		t.visitStatement(s, shadow)
	}
	if stmt.ElseClause != nil {
		if stmt.ElseClause.ElseIf != nil {
			t.visitIfStatement(stmt.ElseClause.ElseIf, shadow)
		}
		for _, s := range stmt.ElseClause.ElseBody {
			t.visitStatement(s, shadow)
		}
	}
}

// visitFunctionDecl copies the shadow set before adding params so UDF scope cannot
// leak into the caller's scope.
func (t *V4BuiltinIdentifierTransformer) visitFunctionDecl(fn *parser.FunctionDecl, outerShadow map[string]bool) {
	if fn == nil {
		return
	}
	fnShadow := make(map[string]bool, len(outerShadow)+len(fn.Params))
	for k := range outerShadow {
		fnShadow[k] = true
	}
	for _, p := range fn.Params {
		fnShadow[p] = true
	}

	for _, s := range fn.MultiLineBody {
		t.visitStatement(s, fnShadow)
	}
	if fn.InlineBody != nil {
		t.visitExpression(fn.InlineBody, fnShadow)
	}
	if fn.InlineStatementList != nil {
		for _, is := range fn.InlineStatementList.Statements {
			t.visitTernaryExpr(is.Value, fnShadow)
		}
		t.visitTernaryExpr(fn.InlineStatementList.FinalExpression, fnShadow)
	}
}

func (t *V4BuiltinIdentifierTransformer) visitSwitchExpr(sw *parser.SwitchExpr, shadow map[string]bool) {
	if sw == nil {
		return
	}
	t.visitOrExpr(sw.Subject, shadow)
	for _, c := range sw.Cases {
		if c != nil {
			t.visitOrExpr(c.Condition, shadow)
			for _, s := range c.Body {
				t.visitStatement(s, shadow)
			}
			if c.InlineBody != nil {
				t.visitExpression(c.InlineBody, shadow)
			}
		}
	}
}

func (t *V4BuiltinIdentifierTransformer) visitExpression(expr *parser.Expression, shadow map[string]bool) {
	if expr == nil {
		return
	}
	if expr.ForInExpr != nil {
		t.visitArithExpr(expr.ForInExpr.Collection, shadow)
		for _, s := range expr.ForInExpr.Body {
			t.visitStatement(s, shadow)
		}
	}
	if expr.ForExpr != nil {
		t.visitArithExpr(expr.ForExpr.From, shadow)
		t.visitArithExpr(expr.ForExpr.To, shadow)
		t.visitArithExpr(expr.ForExpr.Step, shadow)
		for _, s := range expr.ForExpr.Body {
			t.visitStatement(s, shadow)
		}
	}
	if expr.WhileExpr != nil {
		t.visitOrExpr(expr.WhileExpr.Condition, shadow)
		for _, s := range expr.WhileExpr.Body {
			t.visitStatement(s, shadow)
		}
	}
	if expr.IfExpr != nil {
		t.visitOrExpr(expr.IfExpr.Condition, shadow)
		for _, s := range expr.IfExpr.Body {
			t.visitStatement(s, shadow)
		}
		if expr.IfExpr.ElseClause != nil {
			if expr.IfExpr.ElseClause.ElseIf != nil {
				t.visitIfStatement(expr.IfExpr.ElseClause.ElseIf, shadow)
			}
			for _, s := range expr.IfExpr.ElseClause.ElseBody {
				t.visitStatement(s, shadow)
			}
		}
	}
	if expr.SwitchExpr != nil {
		t.visitSwitchExpr(expr.SwitchExpr, shadow)
	}
	if expr.Ternary != nil {
		t.visitTernaryExpr(expr.Ternary, shadow)
	}
	if expr.Array != nil {
		for _, elem := range expr.Array.Elements {
			t.visitTernaryExpr(elem, shadow)
		}
	}
	if expr.Call != nil {
		t.visitCallArgs(expr.Call, shadow)
	}
}

func (t *V4BuiltinIdentifierTransformer) visitTernaryExpr(ternary *parser.TernaryExpr, shadow map[string]bool) {
	if ternary == nil {
		return
	}
	t.visitOrExpr(ternary.Condition, shadow)
	t.visitExpression(ternary.TrueVal, shadow)
	t.visitExpression(ternary.FalseVal, shadow)
}

func (t *V4BuiltinIdentifierTransformer) visitOrExpr(or *parser.OrExpr, shadow map[string]bool) {
	if or == nil {
		return
	}
	t.visitAndExpr(or.Left, shadow)
	t.visitOrExpr(or.Right, shadow)
}

func (t *V4BuiltinIdentifierTransformer) visitAndExpr(and *parser.AndExpr, shadow map[string]bool) {
	if and == nil {
		return
	}
	t.visitCompExpr(and.Left, shadow)
	t.visitAndExpr(and.Right, shadow)
}

func (t *V4BuiltinIdentifierTransformer) visitCompExpr(comp *parser.CompExpr, shadow map[string]bool) {
	if comp == nil {
		return
	}
	t.visitArithExpr(comp.Left, shadow)
	t.visitCompExpr(comp.Right, shadow)
}

func (t *V4BuiltinIdentifierTransformer) visitArithExpr(arith *parser.ArithExpr, shadow map[string]bool) {
	if arith == nil {
		return
	}
	t.visitTerm(arith.Left, shadow)
	t.visitArithExpr(arith.Right, shadow)
}

func (t *V4BuiltinIdentifierTransformer) visitTerm(term *parser.Term, shadow map[string]bool) {
	if term == nil {
		return
	}
	t.visitFactor(term.Left, shadow)
	t.visitTerm(term.Right, shadow)
}

func (t *V4BuiltinIdentifierTransformer) visitFactor(factor *parser.Factor, shadow map[string]bool) {
	if factor == nil {
		return
	}
	// The parser always prefers Postfix over Ident for bare identifiers, so this is a
	// defensive guard for any grammar edge case where Factor.Ident is set directly.
	rewriteFactorIdent(factor, t.mappings, shadow)
	if factor.Ident != nil {
		return
	}

	if factor.Array != nil {
		for _, elem := range factor.Array.Elements {
			t.visitTernaryExpr(elem, shadow)
		}
	}
	if factor.Unary != nil {
		t.visitFactor(factor.Unary.Operand, shadow)
	}
	if factor.Postfix != nil {
		t.visitPostfixExpr(factor.Postfix, shadow)
	}
}

func (t *V4BuiltinIdentifierTransformer) visitPostfixExpr(postfix *parser.PostfixExpr, shadow map[string]bool) {
	if postfix == nil {
		return
	}
	if postfix.Primary != nil {
		t.visitPrimaryExpr(postfix.Primary, shadow)
	}
	t.visitArithExpr(postfix.Subscript, shadow)
}

func (t *V4BuiltinIdentifierTransformer) visitPrimaryExpr(primary *parser.PrimaryExpr, shadow map[string]bool) {
	if primary == nil {
		return
	}
	rewritePrimaryExprIdent(primary, t.mappings, shadow)
	if primary.Ident != nil {
		return
	}

	if primary.Paren != nil {
		t.visitExpression(primary.Paren, shadow)
	}
	if primary.Call != nil {
		t.visitCallArgs(primary.Call, shadow)
	}
}

// visitCallArgs intentionally skips the callee — function-name rewrites belong to
// TickerNamespaceTransformer / TANamespaceTransformer, not here.
func (t *V4BuiltinIdentifierTransformer) visitCallArgs(call *parser.CallExpr, shadow map[string]bool) {
	if call == nil {
		return
	}
	for _, arg := range call.Args {
		if arg != nil {
			t.visitExpression(arg.Value, shadow)
		}
	}
}
