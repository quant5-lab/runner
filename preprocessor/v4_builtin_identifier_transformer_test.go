package preprocessor

import (
	"testing"

	"github.com/quant5-lab/runner/parser"
)

// ── CST navigation helpers ────────────────────────────────────────────────────
//
// Bare identifiers always parse through the expression chain:
//
//	Expression.Ternary.Condition  (OrExpr)
//	  .Left                       (AndExpr)
//	  .Left                       (CompExpr)
//	  .Left                       (ArithExpr)
//	  .Left                       (Term)
//	  .Left                       (Factor)
//	  .Postfix.Primary            (PrimaryExpr) ← rewrite site
//
// All helper functions below are derived from this invariant.

// primaryFromOrExpr navigates from an OrExpr to the PrimaryExpr of its leftmost simple expression.
func primaryFromOrExpr(or *parser.OrExpr) *parser.PrimaryExpr {
	if or == nil || or.Left == nil || or.Left.Left == nil || or.Left.Left.Left == nil ||
		or.Left.Left.Left.Left == nil || or.Left.Left.Left.Left.Left == nil {
		return nil
	}
	factor := or.Left.Left.Left.Left.Left
	if factor.Postfix == nil {
		return nil
	}
	return factor.Postfix.Primary
}

// simpleExpressionPrimary navigates to the PrimaryExpr of a simple (non-compound) Expression.
func simpleExpressionPrimary(expr *parser.Expression) *parser.PrimaryExpr {
	if expr == nil || expr.Ternary == nil {
		return nil
	}
	return primaryFromOrExpr(expr.Ternary.Condition)
}

// arithFromExpression returns the ArithExpr node of a simple expression.
// Useful for inspecting both sides of an arithmetic expression (e.g. a + b).
func arithFromExpression(expr *parser.Expression) *parser.ArithExpr {
	if expr == nil || expr.Ternary == nil || expr.Ternary.Condition == nil ||
		expr.Ternary.Condition.Left == nil || expr.Ternary.Condition.Left.Left == nil {
		return nil
	}
	return expr.Ternary.Condition.Left.Left.Left
}

// assignmentRHS returns the RHS Expression of the statement at stmtIndex,
// handling both plain Assignment and VarAssignment forms.
func assignmentRHS(t *testing.T, script *parser.Script, stmtIndex int) *parser.Expression {
	t.Helper()
	stmt := script.Statements[stmtIndex].Core
	if stmt.Assignment != nil {
		return stmt.Assignment.Value
	}
	if stmt.VarAssignment != nil {
		return stmt.VarAssignment.Value
	}
	t.Fatalf("statement %d is neither Assignment nor VarAssignment", stmtIndex)
	return nil
}

// assertPrimaryMemberAccess verifies that a PrimaryExpr carries the expected MemberAccess
// and that its Ident field has been cleared.
func assertPrimaryMemberAccess(t *testing.T, primary *parser.PrimaryExpr, wantObject, wantProperty string) {
	t.Helper()
	if primary == nil {
		t.Fatal("PrimaryExpr is nil")
	}
	if primary.MemberAccess == nil {
		t.Fatalf("expected PrimaryExpr.MemberAccess{%s.%s}: got Ident=%v",
			wantObject, wantProperty, primary.Ident)
	}
	if primary.Ident != nil {
		t.Errorf("PrimaryExpr.Ident must be nil after rewrite, got %q", *primary.Ident)
	}
	if primary.MemberAccess.Object != wantObject {
		t.Errorf("MemberAccess.Object: got %q want %q", primary.MemberAccess.Object, wantObject)
	}
	if len(primary.MemberAccess.Properties) == 0 || primary.MemberAccess.Properties[0] != wantProperty {
		t.Errorf("MemberAccess.Properties[0]: got %v want %q", primary.MemberAccess.Properties, wantProperty)
	}
}

// assertPrimaryIdent verifies that a PrimaryExpr still holds a bare identifier (not rewritten).
func assertPrimaryIdent(t *testing.T, primary *parser.PrimaryExpr, wantIdent string) {
	t.Helper()
	if primary == nil {
		t.Fatal("PrimaryExpr is nil")
	}
	if primary.MemberAccess != nil {
		t.Errorf("PrimaryExpr must not be rewritten: got MemberAccess{%s.%v}",
			primary.MemberAccess.Object, primary.MemberAccess.Properties)
	}
	if primary.Ident == nil || *primary.Ident != wantIdent {
		t.Errorf("PrimaryExpr.Ident: got %v want %q", primary.Ident, wantIdent)
	}
}

// ── V4BuiltinIdentifierTransformer unit tests ─────────────────────────────────

// TestV4BuiltinIdentifierTransformer_AllMappingsRewritten verifies that every v4 bare
// identifier in the mapping table is correctly qualified to its v5 namespace form.
func TestV4BuiltinIdentifierTransformer_AllMappingsRewritten(t *testing.T) {
	cases := []struct {
		v4Ident      string
		wantObject   string
		wantProperty string
	}{
		// syminfo namespace
		{"tickerid", "syminfo", "tickerid"},
		{"ticker", "syminfo", "ticker"},
		// timeframe namespace
		{"period", "timeframe", "period"},
		{"interval", "timeframe", "multiplier"},
		{"isintraday", "timeframe", "isintraday"},
		{"isdwm", "timeframe", "isdwm"},
		{"isdaily", "timeframe", "isdaily"},
		{"isweekly", "timeframe", "isweekly"},
		{"ismonthly", "timeframe", "ismonthly"},
	}

	transformer := NewV4BuiltinIdentifierTransformer()
	for _, tc := range cases {
		t.Run(tc.v4Ident, func(t *testing.T) {
			result, err := transformer.Transform(parseScript(t, "x = "+tc.v4Ident))
			if err != nil {
				t.Fatal(err)
			}
			assertPrimaryMemberAccess(t,
				simpleExpressionPrimary(assignmentRHS(t, result, 0)),
				tc.wantObject, tc.wantProperty)
		})
	}
}

// TestV4BuiltinIdentifierTransformer_UnmappedIdentUnchanged verifies that identifiers
// not in the v4 mapping (e.g. built-in series and user variables) are left untouched.
func TestV4BuiltinIdentifierTransformer_UnmappedIdentUnchanged(t *testing.T) {
	unchanged := []string{"close", "open", "high", "low", "volume", "myVar"}
	transformer := NewV4BuiltinIdentifierTransformer()
	for _, ident := range unchanged {
		t.Run(ident, func(t *testing.T) {
			result, err := transformer.Transform(parseScript(t, "x = "+ident))
			if err != nil {
				t.Fatal(err)
			}
			assertPrimaryIdent(t, simpleExpressionPrimary(assignmentRHS(t, result, 0)), ident)
		})
	}
}

// TestV4BuiltinIdentifierTransformer_CaseSensitivity verifies that the mapping is
// case-sensitive: only the exact lowercase v4 form triggers a rewrite.
func TestV4BuiltinIdentifierTransformer_CaseSensitivity(t *testing.T) {
	cases := []struct{ input, wantIdent string }{
		{"Tickerid", "Tickerid"},
		{"TICKERID", "TICKERID"},
		{"Period", "Period"},
		{"PERIOD", "PERIOD"},
	}
	transformer := NewV4BuiltinIdentifierTransformer()
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := transformer.Transform(parseScript(t, "x = "+tc.input))
			if err != nil {
				t.Fatal(err)
			}
			assertPrimaryIdent(t, simpleExpressionPrimary(assignmentRHS(t, result, 0)), tc.wantIdent)
		})
	}
}

// TestV4BuiltinIdentifierTransformer_Idempotency verifies that already-qualified
// identifiers (e.g. syminfo.tickerid) are not modified when the transformer runs again.
// Running the transformer twice must produce the same result as running it once.
func TestV4BuiltinIdentifierTransformer_Idempotency(t *testing.T) {
	cases := []struct {
		input        string
		wantObject   string
		wantProperty string
	}{
		{"x = syminfo.tickerid", "syminfo", "tickerid"},
		{"x = timeframe.period", "timeframe", "period"},
	}
	transformer := NewV4BuiltinIdentifierTransformer()
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			script := parseScript(t, tc.input)
			// First pass
			result, err := transformer.Transform(script)
			if err != nil {
				t.Fatal(err)
			}
			// Second pass — must not corrupt the already-rewritten MemberAccess
			result, err = transformer.Transform(result)
			if err != nil {
				t.Fatal(err)
			}
			assertPrimaryMemberAccess(t,
				simpleExpressionPrimary(assignmentRHS(t, result, 0)),
				tc.wantObject, tc.wantProperty)
		})
	}
}

// TestV4BuiltinIdentifierTransformer_EmptyScript verifies the transformer handles
// an empty (zero-statement) script without panicking.
func TestV4BuiltinIdentifierTransformer_EmptyScript(t *testing.T) {
	result, err := NewV4BuiltinIdentifierTransformer().Transform(parseScript(t, ""))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Statements) != 0 {
		t.Errorf("expected 0 statements, got %d", len(result.Statements))
	}
}

// TestV4BuiltinIdentifierTransformer_TernaryBothBranches verifies that v4 identifiers
// are rewritten in both the true and false branches of a ternary expression.
func TestV4BuiltinIdentifierTransformer_TernaryBothBranches(t *testing.T) {
	script := parseScript(t, `x = isintraday ? tickerid : period`)
	result, err := NewV4BuiltinIdentifierTransformer().Transform(script)
	if err != nil {
		t.Fatal(err)
	}
	ternary := assignmentRHS(t, result, 0).Ternary
	if ternary == nil {
		t.Fatal("expected TernaryExpr")
	}
	// True branch: tickerid → syminfo.tickerid
	assertPrimaryMemberAccess(t, simpleExpressionPrimary(ternary.TrueVal), "syminfo", "tickerid")
	// False branch: period → timeframe.period
	assertPrimaryMemberAccess(t, simpleExpressionPrimary(ternary.FalseVal), "timeframe", "period")
}

// TestV4BuiltinIdentifierTransformer_InArithmeticExpression verifies that v4 identifiers
// used as operands in arithmetic expressions are rewritten at both operand positions.
func TestV4BuiltinIdentifierTransformer_InArithmeticExpression(t *testing.T) {
	result, err := NewV4BuiltinIdentifierTransformer().Transform(
		parseScript(t, `x = tickerid + period`))
	if err != nil {
		t.Fatal(err)
	}
	arith := arithFromExpression(assignmentRHS(t, result, 0))
	if arith == nil {
		t.Fatal("expected ArithExpr")
	}
	// Left operand: tickerid → syminfo.tickerid
	assertPrimaryMemberAccess(t, arith.Left.Left.Postfix.Primary, "syminfo", "tickerid")
	// Right operand: period → timeframe.period
	assertPrimaryMemberAccess(t, arith.Right.Left.Left.Postfix.Primary, "timeframe", "period")
}

// TestV4BuiltinIdentifierTransformer_InIfConditionAndBody verifies that v4 identifiers
// are rewritten both in the condition expression and in the body of an if statement.
func TestV4BuiltinIdentifierTransformer_InIfConditionAndBody(t *testing.T) {
	input := `
if isintraday
    x = tickerid
`
	result, err := NewV4BuiltinIdentifierTransformer().Transform(parseScript(t, input))
	if err != nil {
		t.Fatal(err)
	}
	ifStmt := result.Statements[0].Core.If
	if ifStmt == nil {
		t.Fatal("expected IfStatement")
	}
	// Condition: isintraday → timeframe.isintraday
	assertPrimaryMemberAccess(t, primaryFromOrExpr(ifStmt.Condition), "timeframe", "isintraday")
	// Body: tickerid → syminfo.tickerid
	assertPrimaryMemberAccess(t,
		simpleExpressionPrimary(ifStmt.Body[0].Core.Assignment.Value),
		"syminfo", "tickerid")
}

// TestV4BuiltinIdentifierTransformer_InForLoopBody verifies that v4 identifiers in
// the body of a for loop are rewritten correctly.
func TestV4BuiltinIdentifierTransformer_InForLoopBody(t *testing.T) {
	input := `
for i = 0 to 10
    x = tickerid
`
	result, err := NewV4BuiltinIdentifierTransformer().Transform(parseScript(t, input))
	if err != nil {
		t.Fatal(err)
	}
	forStmt := result.Statements[0].Core.For
	if forStmt == nil {
		t.Fatal("expected ForStatement")
	}
	assertPrimaryMemberAccess(t,
		simpleExpressionPrimary(forStmt.Body[0].Core.Assignment.Value),
		"syminfo", "tickerid")
}

// TestV4BuiltinIdentifierTransformer_VarAssignment verifies that v4 identifiers in
// var/varip declarations are rewritten correctly.
func TestV4BuiltinIdentifierTransformer_VarAssignment(t *testing.T) {
	cases := []struct{ input, wantObj, wantProp string }{
		{"var x = tickerid", "syminfo", "tickerid"},
		{"varip x = period", "timeframe", "period"},
	}
	transformer := NewV4BuiltinIdentifierTransformer()
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := transformer.Transform(parseScript(t, tc.input))
			if err != nil {
				t.Fatal(err)
			}
			assertPrimaryMemberAccess(t,
				simpleExpressionPrimary(assignmentRHS(t, result, 0)),
				tc.wantObj, tc.wantProp)
		})
	}
}

// TestV4BuiltinIdentifierTransformer_UDFBodyShadowScoping verifies the UDF parameter
// shadowing rules: a param named after a v4 ident suppresses its rewrite within the body,
// while params with different names leave the v4 ident free to be rewritten.
func TestV4BuiltinIdentifierTransformer_UDFBodyShadowScoping(t *testing.T) {
	cases := []struct {
		name          string
		param         string
		bodyIdent     string
		wantRewritten bool
		wantObj       string
		wantProp      string
	}{
		// param name matches the mapped ident → shadow, body ident left as-is
		{"tickerid_shadowed_by_param", "tickerid", "tickerid", false, "", ""},
		{"period_shadowed_by_param", "period", "period", false, "", ""},
		// param name is different → no shadow, body ident rewritten
		{"tickerid_not_shadowed", "src", "tickerid", true, "syminfo", "tickerid"},
		{"period_not_shadowed", "src", "period", true, "timeframe", "period"},
		{"interval_not_shadowed", "length", "interval", true, "timeframe", "multiplier"},
	}

	transformer := NewV4BuiltinIdentifierTransformer()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := "\nf(" + tc.param + ") =>\n    " + tc.bodyIdent + "\n"
			result, err := transformer.Transform(parseScript(t, input))
			if err != nil {
				t.Fatal(err)
			}
			fn := result.Statements[0].Core.FunctionDecl
			if fn == nil {
				t.Fatal("expected FunctionDecl")
			}
			primary := simpleExpressionPrimary(fn.MultiLineBody[0].Core.Expression.Expr)
			if tc.wantRewritten {
				assertPrimaryMemberAccess(t, primary, tc.wantObj, tc.wantProp)
			} else {
				assertPrimaryIdent(t, primary, tc.bodyIdent)
			}
		})
	}
}

// TestV4BuiltinIdentifierTransformer_OuterScopeUnaffectedByUDFShadow verifies that a
// UDF parameter shadow is scoped to the function body and does not affect the outer script.
func TestV4BuiltinIdentifierTransformer_OuterScopeUnaffectedByUDFShadow(t *testing.T) {
	input := `
f(tickerid) =>
    tickerid
x = tickerid
`
	result, err := NewV4BuiltinIdentifierTransformer().Transform(parseScript(t, input))
	if err != nil {
		t.Fatal(err)
	}
	// Outer-scope `x = tickerid` (statement index 1) must still be rewritten.
	assertPrimaryMemberAccess(t,
		simpleExpressionPrimary(assignmentRHS(t, result, 1)),
		"syminfo", "tickerid")
}

// TestV4BuiltinIdentifierTransformer_UDFMultipleShadowParams verifies that when multiple
// UDF parameters collectively shadow several v4 idents, all are suppressed in the body.
func TestV4BuiltinIdentifierTransformer_UDFMultipleShadowParams(t *testing.T) {
	input := `
f(tickerid, period) =>
    tickerid + period
`
	result, err := NewV4BuiltinIdentifierTransformer().Transform(parseScript(t, input))
	if err != nil {
		t.Fatal(err)
	}
	fn := result.Statements[0].Core.FunctionDecl
	if fn == nil {
		t.Fatal("expected FunctionDecl")
	}
	arith := arithFromExpression(fn.MultiLineBody[0].Core.Expression.Expr)
	if arith == nil {
		t.Fatal("expected ArithExpr in body")
	}
	// Both operands must remain as bare idents (both params shadow the mappings).
	assertPrimaryIdent(t, arith.Left.Left.Postfix.Primary, "tickerid")
	assertPrimaryIdent(t, arith.Right.Left.Left.Postfix.Primary, "period")
}

// TestV4BuiltinIdentifierTransformer_UDFInlineBody verifies that v4 identifiers in
// a UDF's single-expression inline body (f(src) => expr) are rewritten when not shadowed.
func TestV4BuiltinIdentifierTransformer_UDFInlineBody(t *testing.T) {
	input := `f(src) => tickerid`
	result, err := NewV4BuiltinIdentifierTransformer().Transform(parseScript(t, input))
	if err != nil {
		t.Fatal(err)
	}
	fn := result.Statements[0].Core.FunctionDecl
	if fn == nil || fn.InlineBody == nil {
		t.Fatal("expected FunctionDecl with InlineBody")
	}
	assertPrimaryMemberAccess(t, simpleExpressionPrimary(fn.InlineBody), "syminfo", "tickerid")
}

// ── V4ToV5Pipeline integration tests ─────────────────────────────────────────

// TestV4ToV5Pipeline_V4IdentsRewritten verifies that every v4 bare identifier is
// correctly rewritten when the full migration pipeline is applied. This ensures
// the transformer is wired correctly and no earlier pipeline stage interferes.
func TestV4ToV5Pipeline_V4IdentsRewritten(t *testing.T) {
	cases := []struct {
		v4Ident      string
		wantObject   string
		wantProperty string
	}{
		{"tickerid", "syminfo", "tickerid"},
		{"ticker", "syminfo", "ticker"},
		{"period", "timeframe", "period"},
		{"interval", "timeframe", "multiplier"},
		{"isintraday", "timeframe", "isintraday"},
		{"isdwm", "timeframe", "isdwm"},
		{"isdaily", "timeframe", "isdaily"},
		{"isweekly", "timeframe", "isweekly"},
		{"ismonthly", "timeframe", "ismonthly"},
	}

	pipeline := NewV4ToV5Pipeline()
	for _, tc := range cases {
		t.Run(tc.v4Ident, func(t *testing.T) {
			result, err := pipeline.Run(parseScript(t, "x = "+tc.v4Ident))
			if err != nil {
				t.Fatal(err)
			}
			assertPrimaryMemberAccess(t,
				simpleExpressionPrimary(assignmentRHS(t, result, 0)),
				tc.wantObject, tc.wantProperty)
		})
	}
}

// TestV4ToV5Pipeline_RewritesTickeridInTernary is an integration test for the
// combined effect of TickerNamespaceTransformer and V4BuiltinIdentifierTransformer.
// It exercises the pattern found in zigzag/zigzag-pa: a ternary whose true branch
// calls a v4 ticker constructor (heikenashi) with a bare tickerid argument, and whose
// false branch is a bare tickerid.
func TestV4ToV5Pipeline_RewritesTickeridInTernary(t *testing.T) {
	input := `_ticker = useHA ? heikenashi(tickerid) : tickerid`
	result, err := NewV4ToV5Pipeline().Run(parseScript(t, input))
	if err != nil {
		t.Fatal(err)
	}

	ternary := assignmentRHS(t, result, 0).Ternary
	if ternary == nil {
		t.Fatal("expected TernaryExpr")
	}

	// True branch: heikenashi(tickerid) → ticker.heikinashi(syminfo.tickerid)
	trueFactor := ternary.TrueVal.Ternary.Condition.Left.Left.Left.Left.Left
	if trueFactor == nil || trueFactor.Postfix == nil || trueFactor.Postfix.Primary == nil {
		t.Fatal("expected Factor.Postfix.Primary in true branch")
	}
	trueBranchCall := trueFactor.Postfix.Primary.Call
	if trueBranchCall == nil {
		t.Fatal("expected Call in true branch")
	}
	assertMemberAccessCallee(t, trueBranchCall, "ticker", "heikinashi")
	assertPrimaryMemberAccess(t, simpleExpressionPrimary(trueBranchCall.Args[0].Value), "syminfo", "tickerid")

	// False branch: tickerid → syminfo.tickerid
	assertPrimaryMemberAccess(t, simpleExpressionPrimary(ternary.FalseVal), "syminfo", "tickerid")
}
