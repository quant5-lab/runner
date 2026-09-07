package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/security"
)

// sor_ prefix avoids collision with cod_* helpers in chart_only_udf_detector_test.go.

func sor_prog(stmts ...ast.Node) *ast.Program {
	return &ast.Program{Body: stmts}
}

func sor_varDecl(name string, init ast.Expression) *ast.VariableDeclaration {
	return &ast.VariableDeclaration{
		Declarations: []ast.VariableDeclarator{
			{ID: &ast.Identifier{Name: name}, Init: init},
		},
	}
}

func sor_tupleDecl(names []string, init ast.Expression) *ast.VariableDeclaration {
	elems := make([]ast.Identifier, len(names))
	for i, n := range names {
		elems[i] = ast.Identifier{Name: n}
	}
	return &ast.VariableDeclaration{
		Declarations: []ast.VariableDeclarator{
			{ID: &ast.ArrayPattern{Elements: elems}, Init: init},
		},
	}
}

func sor_call(name string, args ...ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee:    &ast.Identifier{Name: name},
		Arguments: args,
	}
}

// Returns (call, symExpr) — the same pointer used as lhsIndex key and SymbolExpr.
func sor_securityCall(symExpr ast.Expression, tf ast.Expression) (*ast.CallExpression, ast.Expression) {
	call := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "security"},
		Arguments: []ast.Expression{symExpr, tf, &ast.Identifier{Name: "close"}},
	}
	return call, call.Arguments[0]
}

// request.security() member-expression form; same return contract as sor_securityCall.
func sor_requestSecurityCall(symExpr ast.Expression, tf ast.Expression) (*ast.CallExpression, ast.Expression) {
	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "request"},
			Property: &ast.Identifier{Name: "security"},
		},
		Arguments: []ast.Expression{symExpr, tf, &ast.Identifier{Name: "close"}},
	}
	return call, call.Arguments[0]
}

func sor_ident(name string) *ast.Identifier { return &ast.Identifier{Name: name} }
func sor_str(v string) *ast.Literal         { return &ast.Literal{Value: v} }
func sor_num(v float64) *ast.Literal        { return &ast.Literal{Value: v} }
func sor_exprStmt(e ast.Expression) *ast.ExpressionStatement {
	return &ast.ExpressionStatement{Expression: e}
}
func sor_if(test ast.Expression, body ...ast.Node) *ast.IfStatement {
	return &ast.IfStatement{Test: test, Consequent: body}
}
func sor_for(body ...ast.Node) *ast.ForStatement {
	return &ast.ForStatement{
		Counter: "i", From: sor_num(0), To: sor_num(10), Body: body,
	}
}
func sor_while(cond ast.Expression, body ...ast.Node) *ast.WhileStatement {
	return &ast.WhileStatement{Condition: cond, Body: body}
}
func sor_forIn(collection ast.Expression, body ...ast.Node) *ast.ForInStatement {
	return &ast.ForInStatement{ElementVar: "elem", Collection: collection, Body: body}
}
func sor_arrow(body ...ast.Node) *ast.ArrowFunctionExpression {
	return &ast.ArrowFunctionExpression{Body: body}
}
func sor_binary(left ast.Expression, op string, right ast.Expression) *ast.BinaryExpression {
	return &ast.BinaryExpression{Left: left, Operator: op, Right: right}
}
func sor_logical(left ast.Expression, op string, right ast.Expression) *ast.LogicalExpression {
	return &ast.LogicalExpression{Left: left, Operator: op, Right: right}
}
func sor_cond(test, cons, alt ast.Expression) *ast.ConditionalExpression {
	return &ast.ConditionalExpression{Test: test, Consequent: cons, Alternate: alt}
}
func sor_unary(op string, arg ast.Expression) *ast.UnaryExpression {
	return &ast.UnaryExpression{Operator: op, Argument: arg, Prefix: true}
}
func sor_member(obj ast.Expression, prop string) *ast.MemberExpression {
	return &ast.MemberExpression{Object: obj, Property: &ast.Identifier{Name: prop}}
}
func sor_memberComputed(obj, idx ast.Expression) *ast.MemberExpression {
	return &ast.MemberExpression{Object: obj, Property: idx, Computed: true}
}

// SymbolExpr pointer must match the one stored as the lhsIndex key.
func sor_resolved(symExpr ast.Expression, sym, tf string) resolvedSecurityCall {
	return resolvedSecurityCall{
		call:        security.SecurityCall{SymbolExpr: symExpr},
		resolvedSym: sym,
		resolvedTf:  tf,
	}
}

// resolvedDedupKey replaces the flagged dimensions with runtimePlaceholder() — tests must use the same key format.
func sor_resolvedRuntime(symExpr ast.Expression, sym, tf string, symIsRuntime, tfIsRuntime bool) resolvedSecurityCall {
	return resolvedSecurityCall{
		call:         security.SecurityCall{SymbolExpr: symExpr},
		resolvedSym:  sym,
		resolvedTf:   tf,
		isSymRuntime: symIsRuntime,
		isTfRuntime:  tfIsRuntime,
	}
}

// ─── TestExtractPatternNames ──────────────────────────────────────────────────

func TestExtractPatternNames(t *testing.T) {
	tests := []struct {
		name    string
		pattern ast.Pattern
		want    []string
	}{
		{
			name:    "identifier yields single name",
			pattern: &ast.Identifier{Name: "x"},
			want:    []string{"x"},
		},
		{
			name:    "empty ArrayPattern yields empty slice",
			pattern: &ast.ArrayPattern{Elements: []ast.Identifier{}},
			want:    []string{},
		},
		{
			name:    "single-element ArrayPattern",
			pattern: &ast.ArrayPattern{Elements: []ast.Identifier{{Name: "a"}}},
			want:    []string{"a"},
		},
		{
			name: "two-element ArrayPattern preserves order",
			pattern: &ast.ArrayPattern{
				Elements: []ast.Identifier{{Name: "trend"}, {Name: "price"}},
			},
			want: []string{"trend", "price"},
		},
		{
			name: "three-element ArrayPattern",
			pattern: &ast.ArrayPattern{
				Elements: []ast.Identifier{{Name: "a"}, {Name: "b"}, {Name: "c"}},
			},
			want: []string{"a", "b", "c"},
		},
		{
			name:    "nil pattern returns nil",
			pattern: nil,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPatternNames(tt.pattern)
			if len(got) != len(tt.want) {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
			for i, w := range tt.want {
				if got[i] != w {
					t.Errorf("[%d]: want %q, got %q", i, w, got[i])
				}
			}
		})
	}
}

// ─── TestBuildSecurityLHSIndex ────────────────────────────────────────────────

func TestBuildSecurityLHSIndex(t *testing.T) {
	tests := []struct {
		name      string
		buildProg func() (*ast.Program, ast.Expression)
		wantNames []string // nil means the expression should NOT appear in index
	}{
		{
			name: "bare security() with single identifier LHS",
			buildProg: func() (*ast.Program, ast.Expression) {
				sym := sor_str("BTCUSDT")
				call, symExpr := sor_securityCall(sym, sor_str("1D"))
				return sor_prog(sor_varDecl("x", call)), symExpr
			},
			wantNames: []string{"x"},
		},
		{
			name: "request.security() (member-expression callee) with single identifier LHS",
			buildProg: func() (*ast.Program, ast.Expression) {
				sym := sor_str("BTCUSDT")
				call, symExpr := sor_requestSecurityCall(sym, sor_str("1D"))
				return sor_prog(sor_varDecl("x", call)), symExpr
			},
			wantNames: []string{"x"},
		},
		{
			name: "security() with tuple ArrayPattern LHS",
			buildProg: func() (*ast.Program, ast.Expression) {
				sym := sor_str("EURUSD")
				call, symExpr := sor_securityCall(sym, sor_str("1h"))
				return sor_prog(sor_tupleDecl([]string{"trend", "price"}, call)), symExpr
			},
			wantNames: []string{"trend", "price"},
		},
		{
			name: "non-security call is not indexed",
			buildProg: func() (*ast.Program, ast.Expression) {
				sym := sor_str("BTCUSDT")
				// ta.sma is not a security call; symExpr used as key that will be absent
				return sor_prog(sor_varDecl("x", sor_call("ta.sma", sym))), sym
			},
			wantNames: nil,
		},
		{
			name: "security() with no arguments is not indexed",
			buildProg: func() (*ast.Program, ast.Expression) {
				sentinel := sor_str("__sentinel__")
				call := &ast.CallExpression{Callee: &ast.Identifier{Name: "security"}}
				return sor_prog(sor_varDecl("x", call)), sentinel
			},
			wantNames: nil,
		},
		{
			name: "security() nested inside another expression is not indexed",
			buildProg: func() (*ast.Program, ast.Expression) {
				sym := sor_str("NVDA")
				call, symExpr := sor_securityCall(sym, sor_str("1h"))
				// wrapped in nz() — not a direct top-level assignment init
				return sor_prog(sor_varDecl("x", sor_call("nz", call))), symExpr
			},
			wantNames: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog, symExpr := tt.buildProg()
			idx := buildSecurityLHSIndex(prog)
			got := idx[symExpr]

			if tt.wantNames == nil {
				if got != nil {
					t.Errorf("expected no entry for key, got %v", got)
				}
				return
			}
			if len(got) != len(tt.wantNames) {
				t.Fatalf("want %v, got %v", tt.wantNames, got)
			}
			for i, w := range tt.wantNames {
				if got[i] != w {
					t.Errorf("[%d]: want %q, got %q", i, w, got[i])
				}
			}
		})
	}

	t.Run("multiple calls in same program each indexed by their own symExpr", func(t *testing.T) {
		sym1 := sor_str("AAPL")
		call1, key1 := sor_securityCall(sym1, sor_str("1h"))
		sym2 := sor_str("NVDA")
		call2, key2 := sor_securityCall(sym2, sor_str("1D"))
		prog := sor_prog(
			sor_varDecl("aapl", call1),
			sor_varDecl("nvda", call2),
		)
		idx := buildSecurityLHSIndex(prog)
		if names := idx[key1]; len(names) != 1 || names[0] != "aapl" {
			t.Errorf("AAPL entry: want [aapl], got %v", names)
		}
		if names := idx[key2]; len(names) != 1 || names[0] != "nvda" {
			t.Errorf("NVDA entry: want [nvda], got %v", names)
		}
	})
}

// ─── TestExpandVariableTaint ──────────────────────────────────────────────────

func TestExpandVariableTaint(t *testing.T) {
	tests := []struct {
		name      string
		seeds     []string
		program   *ast.Program
		wantIn    []string
		wantNotIn []string
	}{
		{
			name:      "empty seed set yields empty taint",
			seeds:     []string{},
			program:   sor_prog(sor_varDecl("x", sor_ident("y"))),
			wantIn:    nil,
			wantNotIn: []string{"x", "y"},
		},
		{
			name:    "seed with no declarations in program stays seed-only",
			seeds:   []string{"x"},
			program: sor_prog(),
			wantIn:  []string{"x"},
		},
		{
			name:  "direct identifier reference propagates taint",
			seeds: []string{"src"},
			program: sor_prog(
				sor_varDecl("derived", sor_ident("src")),
			),
			wantIn:    []string{"src", "derived"},
			wantNotIn: []string{"unrelated"},
		},
		{
			name:  "transitive chain propagates through three variables",
			seeds: []string{"a"},
			program: sor_prog(
				sor_varDecl("b", sor_ident("a")),
				sor_varDecl("c", sor_ident("b")),
			),
			wantIn: []string{"a", "b", "c"},
		},
		{
			name:  "binary expression operand propagates taint",
			seeds: []string{"t"},
			program: sor_prog(
				sor_varDecl("flag", sor_binary(sor_ident("t"), "!=", sor_num(0))),
			),
			wantIn:    []string{"t", "flag"},
			wantNotIn: []string{"other"},
		},
		{
			name:  "logical expression left operand propagates taint",
			seeds: []string{"t"},
			program: sor_prog(
				sor_varDecl("x", sor_logical(sor_ident("t"), "and", sor_ident("other"))),
			),
			wantIn:    []string{"t", "x"},
			wantNotIn: []string{"other"},
		},
		{
			name:  "logical expression right operand propagates taint",
			seeds: []string{"t"},
			program: sor_prog(
				sor_varDecl("x", sor_logical(sor_ident("other"), "or", sor_ident("t"))),
			),
			wantIn:    []string{"t", "x"},
			wantNotIn: []string{"other"},
		},
		{
			name:  "unary expression operand propagates taint",
			seeds: []string{"t"},
			program: sor_prog(
				sor_varDecl("neg", sor_unary("not", sor_ident("t"))),
			),
			wantIn: []string{"t", "neg"},
		},
		{
			name:  "conditional test propagates taint to result variable",
			seeds: []string{"sec"},
			program: sor_prog(
				sor_varDecl("lbl", sor_cond(sor_ident("sec"), sor_str("yes"), sor_str("no"))),
			),
			wantIn: []string{"sec", "lbl"},
		},
		{
			name:  "conditional consequent propagates taint",
			seeds: []string{"t"},
			program: sor_prog(
				sor_varDecl("x", sor_cond(sor_ident("other"), sor_ident("t"), sor_num(0))),
			),
			wantIn:    []string{"t", "x"},
			wantNotIn: []string{"other"},
		},
		{
			name:  "conditional alternate propagates taint",
			seeds: []string{"t"},
			program: sor_prog(
				sor_varDecl("x", sor_cond(sor_ident("other"), sor_num(0), sor_ident("t"))),
			),
			wantIn:    []string{"t", "x"},
			wantNotIn: []string{"other"},
		},
		{
			name:  "call-argument wrapping a tainted variable propagates taint",
			seeds: []string{"t"},
			program: sor_prog(
				sor_varDecl("safe", sor_call("nz", sor_ident("t"))),
			),
			wantIn: []string{"t", "safe"},
		},
		{
			name:  "member expression object propagates taint",
			seeds: []string{"t"},
			program: sor_prog(
				sor_varDecl("prop", sor_member(sor_ident("t"), "value")),
			),
			wantIn: []string{"t", "prop"},
		},
		{
			name:  "literal-init variable is never tainted",
			seeds: []string{"a"},
			program: sor_prog(
				sor_varDecl("x", sor_str("literal")),
			),
			wantIn:    []string{"a"},
			wantNotIn: []string{"x"},
		},
		{
			name:  "variable declaration inside if-body expands taint",
			seeds: []string{"sig"},
			program: sor_prog(
				sor_if(sor_ident("cond"),
					sor_varDecl("derived", sor_ident("sig")),
				),
			),
			wantIn: []string{"sig", "derived"},
		},
		{
			name:  "variable declaration inside for-loop body expands taint",
			seeds: []string{"sig"},
			program: sor_prog(
				sor_for(sor_varDecl("derived", sor_ident("sig"))),
			),
			wantIn: []string{"sig", "derived"},
		},
		{
			name:  "variable declaration inside while-loop body expands taint",
			seeds: []string{"sig"},
			program: sor_prog(
				sor_while(sor_ident("cond"),
					sor_varDecl("derived", sor_ident("sig")),
				),
			),
			wantIn: []string{"sig", "derived"},
		},
		{
			name:  "arrow function body is not entered for taint propagation",
			seeds: []string{"outer"},
			program: sor_prog(
				sor_varDecl("fn", sor_arrow(
					sor_varDecl("inner", sor_ident("outer")),
				)),
			),
			wantIn:    []string{"outer"},
			wantNotIn: []string{"inner"},
		},
		{
			name:  "multiple seeds each taint their own derived variables",
			seeds: []string{"t1", "s1"},
			program: sor_prog(
				sor_varDecl("flag", sor_ident("t1")),
				sor_varDecl("price", sor_ident("s1")),
			),
			wantIn: []string{"t1", "s1", "flag", "price"},
		},
		{
			name:  "variable declaration inside for-in body expands taint",
			seeds: []string{"sig"},
			program: sor_prog(
				sor_forIn(sor_ident("series"),
					sor_varDecl("derived", sor_ident("sig")),
				),
			),
			wantIn: []string{"sig", "derived"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandVariableTaint(tt.seeds, tt.program)
			for _, name := range tt.wantIn {
				if !got[name] {
					t.Errorf("want %q tainted, was not", name)
				}
			}
			for _, name := range tt.wantNotIn {
				if got[name] {
					t.Errorf("want %q not tainted, but was", name)
				}
			}
		})
	}
}

// ─── TestSecurityChartOnlyClassifier ─────────────────────────────────────────

// classifierInput groups values that must share AST node pointer identity.
type classifierInput struct {
	program  *ast.Program
	lhs      securityLHSIndex
	resolved []resolvedSecurityCall
}

func TestSecurityChartOnlyClassifier(t *testing.T) {
	clf := SecurityChartOnlyClassifier{}

	tests := []struct {
		name             string
		chartOnlyUDFs    map[string]bool
		build            func() classifierInput
		wantChartOnly    []string // dedup keys expected in result
		wantNotChartOnly []string // dedup keys expected NOT in result
	}{
		{
			name: "empty resolved list yields empty result",
			build: func() classifierInput {
				return classifierInput{program: sor_prog(), lhs: securityLHSIndex{}}
			},
			wantChartOnly:    nil,
			wantNotChartOnly: nil,
		},
		{
			name: "output consumed only by label.new is chart-only",
			build: func() classifierInput {
				symExpr := sor_str("EURUSD")
				call, _ := sor_securityCall(symExpr, sor_str("1h"))
				prog := sor_prog(
					sor_varDecl("x", call),
					sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("x"))),
				)
				return classifierInput{
					program:  prog,
					lhs:      securityLHSIndex{symExpr: []string{"x"}},
					resolved: []resolvedSecurityCall{sor_resolved(symExpr, "EURUSD", "1h")},
				}
			},
			wantChartOnly: []string{"EURUSD:1h"},
		},
		{
			name: "output consumed by plot is not chart-only",
			build: func() classifierInput {
				symExpr := sor_str("BTCUSDT")
				call, _ := sor_securityCall(symExpr, sor_str("1D"))
				prog := sor_prog(
					sor_varDecl("price", call),
					sor_exprStmt(sor_call("plot", sor_ident("price"))),
				)
				return classifierInput{
					program:  prog,
					lhs:      securityLHSIndex{symExpr: []string{"price"}},
					resolved: []resolvedSecurityCall{sor_resolved(symExpr, "BTCUSDT", "1D")},
				}
			},
			wantNotChartOnly: []string{"BTCUSDT:1D"},
		},
		{
			name: "tuple with all outputs reaching only chart sinks is chart-only",
			build: func() classifierInput {
				symExpr := sor_str("AAPL")
				call, _ := sor_securityCall(symExpr, sor_str("1D"))
				prog := sor_prog(
					sor_tupleDecl([]string{"trend", "price"}, call),
					sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("trend"))),
					sor_exprStmt(sor_call("line.new", sor_ident("bar_index"), sor_ident("price"))),
				)
				return classifierInput{
					program:  prog,
					lhs:      securityLHSIndex{symExpr: []string{"trend", "price"}},
					resolved: []resolvedSecurityCall{sor_resolved(symExpr, "AAPL", "1D")},
				}
			},
			wantChartOnly: []string{"AAPL:1D"},
		},
		{
			name: "tuple with one output reaching plot makes key not chart-only",
			build: func() classifierInput {
				symExpr := sor_str("AAPL")
				call, _ := sor_securityCall(symExpr, sor_str("1D"))
				prog := sor_prog(
					sor_tupleDecl([]string{"trend", "price"}, call),
					sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("trend"))),
					sor_exprStmt(sor_call("plot", sor_ident("price"))),
				)
				return classifierInput{
					program:  prog,
					lhs:      securityLHSIndex{symExpr: []string{"trend", "price"}},
					resolved: []resolvedSecurityCall{sor_resolved(symExpr, "AAPL", "1D")},
				}
			},
			wantNotChartOnly: []string{"AAPL:1D"},
		},
		{
			name: "missing LHS index entry conservatively keeps the fetch",
			build: func() classifierInput {
				symExpr := sor_str("SBER")
				return classifierInput{
					program:  sor_prog(),
					lhs:      securityLHSIndex{},
					resolved: []resolvedSecurityCall{sor_resolved(symExpr, "SBER", "1h")},
				}
			},
			wantNotChartOnly: []string{"SBER:1h"},
		},
		{
			name:          "call to a chart-only UDF with tainted arg is not golden",
			chartOnlyUDFs: map[string]bool{"f_draw": true},
			build: func() classifierInput {
				symExpr := sor_str("GOOG")
				call, _ := sor_securityCall(symExpr, sor_str("1h"))
				prog := sor_prog(
					sor_varDecl("x", call),
					sor_exprStmt(sor_call("f_draw", sor_ident("x"))),
				)
				return classifierInput{
					program:  prog,
					lhs:      securityLHSIndex{symExpr: []string{"x"}},
					resolved: []resolvedSecurityCall{sor_resolved(symExpr, "GOOG", "1h")},
				}
			},
			wantChartOnly: []string{"GOOG:1h"},
		},
		{
			name: "transitive alias through intermediate variable reaches plot",
			build: func() classifierInput {
				symExpr := sor_str("NVDA")
				call, _ := sor_securityCall(symExpr, sor_str("1h"))
				prog := sor_prog(
					sor_varDecl("raw", call),
					sor_varDecl("alias", sor_ident("raw")),
					sor_exprStmt(sor_call("plot", sor_ident("alias"))),
				)
				return classifierInput{
					program:  prog,
					lhs:      securityLHSIndex{symExpr: []string{"raw"}},
					resolved: []resolvedSecurityCall{sor_resolved(symExpr, "NVDA", "1h")},
				}
			},
			wantNotChartOnly: []string{"NVDA:1h"},
		},
		{
			name: "three-hop transitive chain: raw→alias→label is chart-only",
			build: func() classifierInput {
				symExpr := sor_str("MSFT")
				call, _ := sor_securityCall(symExpr, sor_str("1h"))
				prog := sor_prog(
					sor_varDecl("raw", call),
					sor_varDecl("alias", sor_ident("raw")),
					sor_varDecl("txt", sor_ident("alias")),
					sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("txt"))),
				)
				return classifierInput{
					program:  prog,
					lhs:      securityLHSIndex{symExpr: []string{"raw"}},
					resolved: []resolvedSecurityCall{sor_resolved(symExpr, "MSFT", "1h")},
				}
			},
			wantChartOnly: []string{"MSFT:1h"},
		},
		{
			name: "same dedup key: one chart-only call and one not — key is not chart-only",
			build: func() classifierInput {
				sym1 := sor_str("X")
				call1, _ := sor_securityCall(sym1, sor_str("1h"))
				sym2 := sor_str("X")
				call2, _ := sor_securityCall(sym2, sor_str("1h"))
				prog := sor_prog(
					sor_varDecl("a", call1),
					sor_varDecl("b", call2),
					sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("a"))),
					sor_exprStmt(sor_call("plot", sor_ident("b"))),
				)
				return classifierInput{
					program:  prog,
					lhs:      securityLHSIndex{sym1: []string{"a"}, sym2: []string{"b"}},
					resolved: []resolvedSecurityCall{sor_resolved(sym1, "X", "1h"), sor_resolved(sym2, "X", "1h")},
				}
			},
			wantNotChartOnly: []string{"X:1h"},
		},
		{
			name: "same dedup key: both calls chart-only — key is chart-only",
			build: func() classifierInput {
				sym1 := sor_str("X")
				call1, _ := sor_securityCall(sym1, sor_str("1h"))
				sym2 := sor_str("X")
				call2, _ := sor_securityCall(sym2, sor_str("1h"))
				prog := sor_prog(
					sor_varDecl("a", call1),
					sor_varDecl("b", call2),
					sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("a"))),
					sor_exprStmt(sor_call("line.new", sor_ident("bar_index"), sor_ident("b"))),
				)
				return classifierInput{
					program:  prog,
					lhs:      securityLHSIndex{sym1: []string{"a"}, sym2: []string{"b"}},
					resolved: []resolvedSecurityCall{sor_resolved(sym1, "X", "1h"), sor_resolved(sym2, "X", "1h")},
				}
			},
			wantChartOnly: []string{"X:1h"},
		},
		{
			name: "runtime symbol key uses runtimePlaceholder() in dedup key",
			build: func() classifierInput {
				symExpr := sor_ident("syminfo_tickerid") // runtime — resolved as placeholder
				call, _ := sor_securityCall(symExpr, sor_str("1h"))
				prog := sor_prog(
					sor_varDecl("x", call),
					sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("x"))),
				)
				r := sor_resolvedRuntime(symExpr, runtimePlaceholder(), "1h", true, false)
				return classifierInput{
					program:  prog,
					lhs:      securityLHSIndex{symExpr: []string{"x"}},
					resolved: []resolvedSecurityCall{r},
				}
			},
			wantChartOnly: []string{runtimePlaceholder() + ":1h"},
		},
		{
			name: "runtime timeframe key uses runtimePlaceholder() in dedup key",
			build: func() classifierInput {
				symExpr := sor_str("EURUSD")
				call, _ := sor_securityCall(symExpr, sor_ident("tf"))
				prog := sor_prog(
					sor_varDecl("x", call),
					sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("x"))),
				)
				r := sor_resolvedRuntime(symExpr, "EURUSD", runtimePlaceholder(), false, true)
				return classifierInput{
					program:  prog,
					lhs:      securityLHSIndex{symExpr: []string{"x"}},
					resolved: []resolvedSecurityCall{r},
				}
			},
			wantChartOnly: []string{"EURUSD:" + runtimePlaceholder()},
		},
		{
			name: "request.security() form is indexed by buildSecurityLHSIndex and classified",
			build: func() classifierInput {
				symExpr := sor_str("GOOG")
				call, keyExpr := sor_requestSecurityCall(symExpr, sor_str("1D"))
				prog := sor_prog(
					sor_varDecl("x", call),
					sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("x"))),
				)
				return classifierInput{
					program:  prog,
					lhs:      buildSecurityLHSIndex(prog), // use actual indexer
					resolved: []resolvedSecurityCall{sor_resolved(keyExpr, "GOOG", "1D")},
				}
			},
			wantChartOnly: []string{"GOOG:1D"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inp := tt.build()
			udfs := tt.chartOnlyUDFs
			if udfs == nil {
				udfs = map[string]bool{}
			}
			keys := clf.ChartOnlySecurityKeys(inp.resolved, inp.lhs, udfs, inp.program)

			for _, key := range tt.wantChartOnly {
				if !keys[key] {
					t.Errorf("key %q: want chart-only, was not", key)
				}
			}
			for _, key := range tt.wantNotChartOnly {
				if keys[key] {
					t.Errorf("key %q: want NOT chart-only, but was", key)
				}
			}
		})
	}

	// ── Conservative keep when output is never used anywhere ──────────────
	// No consuming call means the output cannot be proven harmless — conservative keep.
	t.Run("output variable never used in any call — conservative keep", func(t *testing.T) {
		symExpr := sor_str("BTCUSDT")
		call, _ := sor_securityCall(symExpr, sor_str("1D"))
		prog := sor_prog(sor_varDecl("price", call))
		inp := classifierInput{
			program:  prog,
			lhs:      securityLHSIndex{symExpr: []string{"price"}},
			resolved: []resolvedSecurityCall{sor_resolved(symExpr, "BTCUSDT", "1D")},
		}
		keys := clf.ChartOnlySecurityKeys(inp.resolved, inp.lhs, map[string]bool{}, inp.program)
		if keys["BTCUSDT:1D"] {
			t.Error("key BTCUSDT:1D must NOT be chart-only when output is never consumed")
		}
	})

	// ── strategy.entry: the dominant real-world non-chart-only case ──────────────
	t.Run("output flows directly to strategy.entry — not chart-only", func(t *testing.T) {
		symExpr := sor_str("AAPL")
		call, _ := sor_securityCall(symExpr, sor_str("1h"))
		prog := sor_prog(
			sor_varDecl("sig", call),
			sor_exprStmt(sor_call("strategy.entry", sor_str("Long"), sor_ident("sig"))),
		)
		inp := classifierInput{
			program:  prog,
			lhs:      securityLHSIndex{symExpr: []string{"sig"}},
			resolved: []resolvedSecurityCall{sor_resolved(symExpr, "AAPL", "1h")},
		}
		keys := clf.ChartOnlySecurityKeys(inp.resolved, inp.lhs, map[string]bool{}, inp.program)
		if keys["AAPL:1h"] {
			t.Error("AAPL:1h must NOT be chart-only when sig feeds strategy.entry")
		}
	})

	// ── Chart-only call inside a UDF arrow body ─────────────────────────────────
	t.Run("output consumed only by chart-only UDF whose body has label.new — chart-only", func(t *testing.T) {
		symExpr := sor_str("TSLA")
		call, _ := sor_securityCall(symExpr, sor_str("1h"))
		udfBody := sor_arrow(
			sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("x"))),
		)
		prog := sor_prog(
			sor_varDecl("x", call),
			sor_varDecl("f_draw", udfBody),
			sor_exprStmt(sor_call("f_draw", sor_ident("x"))),
		)
		inp := classifierInput{
			program:  prog,
			lhs:      securityLHSIndex{symExpr: []string{"x"}},
			resolved: []resolvedSecurityCall{sor_resolved(symExpr, "TSLA", "1h")},
		}
		keys := clf.ChartOnlySecurityKeys(inp.resolved, inp.lhs, map[string]bool{"f_draw": true}, inp.program)
		if !keys["TSLA:1h"] {
			t.Error("TSLA:1h must be chart-only when x only feeds chart-only UDF f_draw")
		}
	})
}

// ─── TestOutputIsChartOnly ────────────────────────────────────────────────────

func TestOutputIsChartOnly(t *testing.T) {
	tests := []struct {
		name          string
		tainted       map[string]bool
		chartOnlyUDFs map[string]bool
		program       *ast.Program
		want          bool
	}{
		// ── No uses at all ───────────────────────────────────────────────────
		{
			name:    "output never used in any call — conservative false",
			tainted: map[string]bool{"x": true},
			program: sor_prog(sor_varDecl("unused", sor_str("literal"))),
			want:    false,
		},
		{
			name:    "empty program — conservative false",
			tainted: map[string]bool{"x": true},
			program: sor_prog(),
			want:    false,
		},
		{
			name:    "empty taint set — false (nothing can be chart-only)",
			tainted: map[string]bool{},
			program: sor_prog(sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("x")))),
			want:    false,
		},
		// ── Chart-only namespace calls ────────────────────────────────────────
		{
			name:    "used only in label.new — chart-only",
			tainted: map[string]bool{"x": true},
			program: sor_prog(sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("x")))),
			want:    true,
		},
		{
			name:    "used only in label.set_text — chart-only",
			tainted: map[string]bool{"txt": true},
			program: sor_prog(sor_exprStmt(sor_call("label.set_text", sor_ident("lbl"), sor_ident("txt")))),
			want:    true,
		},
		{
			name:    "used only in line.new — chart-only",
			tainted: map[string]bool{"x": true},
			program: sor_prog(sor_exprStmt(sor_call("line.new", sor_ident("x"), sor_str("y")))),
			want:    true,
		},
		{
			name:    "used only in box.new — chart-only",
			tainted: map[string]bool{"x": true},
			program: sor_prog(sor_exprStmt(sor_call("box.new", sor_ident("x"), sor_str("t"), sor_str("r"), sor_str("b")))),
			want:    true,
		},
		{
			name:    "used only in table.new — chart-only",
			tainted: map[string]bool{"pos": true},
			program: sor_prog(sor_exprStmt(sor_call("table.new", sor_ident("pos"), sor_num(2), sor_num(2)))),
			want:    true,
		},
		{
			name:    "used only in linefill.new — chart-only",
			tainted: map[string]bool{"col": true},
			program: sor_prog(sor_exprStmt(sor_call("linefill.new", sor_ident("l1"), sor_ident("l2"), sor_ident("col")))),
			want:    true,
		},
		{
			name:          "used only in a recognised chart-only UDF — chart-only",
			tainted:       map[string]bool{"x": true},
			chartOnlyUDFs: map[string]bool{"f_draw": true},
			program:       sor_prog(sor_exprStmt(sor_call("f_draw", sor_ident("x")))),
			want:          true,
		},
		{
			name:    "used in multiple chart-only calls — chart-only",
			tainted: map[string]bool{"x": true},
			program: sor_prog(
				sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("x"))),
				sor_exprStmt(sor_call("line.new", sor_ident("bar_index"), sor_ident("x"))),
			),
			want: true,
		},
		// ── Golden sink calls ─────────────────────────────────────────────────
		{
			name:    "used in plot — not chart-only",
			tainted: map[string]bool{"x": true},
			program: sor_prog(sor_exprStmt(sor_call("plot", sor_ident("x")))),
			want:    false,
		},
		{
			name:    "used in strategy.entry — not chart-only",
			tainted: map[string]bool{"sig": true},
			program: sor_prog(sor_exprStmt(sor_call("strategy.entry", sor_str("Long"), sor_ident("sig")))),
			want:    false,
		},
		{
			name:    "used in bgcolor — not chart-only",
			tainted: map[string]bool{"col": true},
			program: sor_prog(sor_exprStmt(sor_call("bgcolor", sor_ident("col")))),
			want:    false,
		},
		{
			name:    "used in alertcondition — not chart-only",
			tainted: map[string]bool{"cond": true},
			program: sor_prog(sor_exprStmt(sor_call("alertcondition", sor_ident("cond")))),
			want:    false,
		},
		// ── Unknown UDF — conservative false ─────────────────────────────────
		{
			name:    "used in unknown UDF — conservative false",
			tainted: map[string]bool{"x": true},
			program: sor_prog(sor_exprStmt(sor_call("myHelper", sor_ident("x")))),
			want:    false,
		},
		{
			name:          "used in unknown UDF even when chart-only UDF map is populated — conservative false",
			tainted:       map[string]bool{"x": true},
			chartOnlyUDFs: map[string]bool{"f_draw": true},
			program:       sor_prog(sor_exprStmt(sor_call("otherUDF", sor_ident("x")))),
			want:          false,
		},
		// ── Mixed uses ───────────────────────────────────────────────────────
		{
			name:    "used in both chart-only and golden sink — not chart-only",
			tainted: map[string]bool{"x": true},
			program: sor_prog(
				sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("x"))),
				sor_exprStmt(sor_call("plot", sor_ident("x"))),
			),
			want: false,
		},
		{
			name:          "used in chart-only UDF and chart-only namespace — chart-only",
			tainted:       map[string]bool{"x": true},
			chartOnlyUDFs: map[string]bool{"f_print": true},
			program: sor_prog(
				sor_exprStmt(sor_call("f_print", sor_ident("x"))),
				sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("x"))),
			),
			want: true,
		},
		// ── Untainted arg in golden sink — untainted arg does not trigger ─────
		{
			name:    "golden sink call but arg is not tainted — chart-only not blocked",
			tainted: map[string]bool{"x": true},
			program: sor_prog(
				sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("x"))),
				sor_exprStmt(sor_call("plot", sor_ident("untainted_close"))),
			),
			want: true,
		},
		// ── Control-flow body scanning ────────────────────────────────────────
		{
			name:    "chart-only call inside if-body counts as a use",
			tainted: map[string]bool{"x": true},
			program: sor_prog(sor_if(sor_ident("cond"),
				sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("x"))),
			)),
			want: true,
		},
		{
			name:    "golden call inside for-loop body makes it not chart-only",
			tainted: map[string]bool{"x": true},
			program: sor_prog(sor_for(
				sor_exprStmt(sor_call("plot", sor_ident("x"))),
			)),
			want: false,
		},
		// ── WhileStatement and ForInStatement body scanning ──────────────────────
		{
			name:    "chart-only call inside while-body counts as a use",
			tainted: map[string]bool{"x": true},
			program: sor_prog(sor_while(sor_ident("cond"),
				sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("x"))),
			)),
			want: true,
		},
		{
			name:    "golden call inside while-body makes it not chart-only",
			tainted: map[string]bool{"x": true},
			program: sor_prog(sor_while(sor_ident("cond"),
				sor_exprStmt(sor_call("strategy.entry", sor_str("Long"), sor_ident("x"))),
			)),
			want: false,
		},
		{
			name:    "chart-only call inside for-in body counts as a use",
			tainted: map[string]bool{"x": true},
			program: sor_prog(sor_forIn(sor_ident("series"),
				sor_exprStmt(sor_call("line.new", sor_ident("x"), sor_num(0))),
			)),
			want: true,
		},
		{
			name:    "golden call inside for-in body makes it not chart-only",
			tainted: map[string]bool{"x": true},
			program: sor_prog(sor_forIn(sor_ident("series"),
				sor_exprStmt(sor_call("plot", sor_ident("x"))),
			)),
			want: false,
		},
		// ── Arrow body is entered for call scanning (unlike walkControlFlow) ────────
		{
			name:    "chart-only call inside arrow body counts as a use",
			tainted: map[string]bool{"x": true},
			program: sor_prog(sor_varDecl("f", sor_arrow(
				sor_exprStmt(sor_call("label.new", sor_ident("bar_index"), sor_ident("x"))),
			))),
			want: true,
		},
		{
			name:    "golden call inside arrow body makes it not chart-only",
			tainted: map[string]bool{"x": true},
			program: sor_prog(sor_varDecl("f", sor_arrow(
				sor_exprStmt(sor_call("strategy.close", sor_str("Long"))),
				sor_exprStmt(sor_call("plot", sor_ident("x"))),
			))),
			want: false,
		},
		// ── Non-namespace names are golden sinks (not eliminated) ──────────────────
		{
			name:    "plotshape is a golden sink",
			tainted: map[string]bool{"x": true},
			program: sor_prog(sor_exprStmt(sor_call("plotshape", sor_ident("x")))),
			want:    false,
		},
		{
			name:    "hline is a golden sink",
			tainted: map[string]bool{"x": true},
			program: sor_prog(sor_exprStmt(sor_call("hline", sor_ident("x")))),
			want:    false,
		},
		{
			name:    "fill is a golden sink",
			tainted: map[string]bool{"x": true},
			program: sor_prog(sor_exprStmt(sor_call("fill", sor_ident("x"), sor_ident("y")))),
			want:    false,
		},
		{
			name:    "barcolor is a golden sink",
			tainted: map[string]bool{"col": true},
			program: sor_prog(sor_exprStmt(sor_call("barcolor", sor_ident("col")))),
			want:    false,
		},
		{
			name:    "plotcandle is a golden sink",
			tainted: map[string]bool{"o": true},
			program: sor_prog(sor_exprStmt(sor_call("plotcandle", sor_ident("o"), sor_str("h"), sor_str("l"), sor_str("c")))),
			want:    false,
		},
		// ── strategy.* prefix variants ───────────────────────────────────────────
		{
			name:    "strategy.close is a golden sink",
			tainted: map[string]bool{"id": true},
			program: sor_prog(sor_exprStmt(sor_call("strategy.close", sor_ident("id")))),
			want:    false,
		},
		{
			name:    "strategy.exit is a golden sink",
			tainted: map[string]bool{"id": true},
			program: sor_prog(sor_exprStmt(sor_call("strategy.exit", sor_ident("id")))),
			want:    false,
		},
		{
			name:    "strategy.order is a golden sink",
			tainted: map[string]bool{"sig": true},
			program: sor_prog(sor_exprStmt(sor_call("strategy.order", sor_str("Long"), sor_ident("sig")))),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udfs := tt.chartOnlyUDFs
			if udfs == nil {
				udfs = map[string]bool{}
			}
			got := outputIsChartOnly(tt.tainted, udfs, tt.program)
			if got != tt.want {
				t.Errorf("want %v, got %v", tt.want, got)
			}
		})
	}
}

// ─── TestExprMentionsTainted ────────────────────────────────────────────────

func TestExprMentionsTainted(t *testing.T) {
	ta := map[string]bool{"x": true} // tainted: x

	tests := []struct {
		name    string
		expr    ast.Expression
		tainted map[string]bool
		want    bool
	}{
		// ── nil ───────────────────────────────────────────────────────────────
		{"nil expr", nil, ta, false},

		// ── Literal — never tainted ───────────────────────────────────────────
		{"literal float", sor_num(42), ta, false},
		{"literal string", sor_str("hello"), ta, false},

		// ── Identifier ────────────────────────────────────────────────────────
		{"identifier tainted", sor_ident("x"), ta, true},
		{"identifier untainted", sor_ident("z"), ta, false},

		// ── CallExpression — taint propagates through arguments only ──────────
		{"call no args", sor_call("f"), ta, false},
		{"call tainted arg", sor_call("f", sor_ident("x")), ta, true},
		{"call untainted arg", sor_call("f", sor_ident("z")), ta, false},
		{"call multiple args one tainted", sor_call("f", sor_ident("z"), sor_ident("x")), ta, true},

		// ── BinaryExpression ──────────────────────────────────────────────────
		{"binary left tainted", sor_binary(sor_ident("x"), "+", sor_ident("z")), ta, true},
		{"binary right tainted", sor_binary(sor_ident("z"), "+", sor_ident("x")), ta, true},
		{"binary neither tainted", sor_binary(sor_ident("z"), "+", sor_ident("z")), ta, false},

		// ── LogicalExpression ─────────────────────────────────────────────────
		{"logical left tainted", sor_logical(sor_ident("x"), "and", sor_ident("z")), ta, true},
		{"logical right tainted", sor_logical(sor_ident("z"), "and", sor_ident("x")), ta, true},
		{"logical neither tainted", sor_logical(sor_ident("z"), "and", sor_ident("z")), ta, false},

		// ── ConditionalExpression ─────────────────────────────────────────────
		{"conditional test tainted", sor_cond(sor_ident("x"), sor_ident("z"), sor_ident("z")), ta, true},
		{"conditional consequent tainted", sor_cond(sor_ident("z"), sor_ident("x"), sor_ident("z")), ta, true},
		{"conditional alternate tainted", sor_cond(sor_ident("z"), sor_ident("z"), sor_ident("x")), ta, true},
		{"conditional none tainted", sor_cond(sor_ident("z"), sor_ident("z"), sor_ident("z")), ta, false},

		// ── UnaryExpression ───────────────────────────────────────────────────
		{"unary argument tainted", sor_unary("not", sor_ident("x")), ta, true},
		{"unary argument untainted", sor_unary("not", sor_ident("z")), ta, false},

		// ── MemberExpression dot-access: only Object checked ─────────────────
		{"dot-access object tainted", sor_member(sor_ident("x"), "prop"), ta, true},
		{"dot-access object untainted property named tainted", sor_member(sor_ident("z"), "x"), ta, false},
		{"dot-access neither tainted", sor_member(sor_ident("z"), "prop"), ta, false},

		// ── MemberExpression computed: Object OR index checked ────────────────
		{"computed-access object tainted", sor_memberComputed(sor_ident("x"), sor_ident("z")), ta, true},
		{"computed-access index tainted", sor_memberComputed(sor_ident("z"), sor_ident("x")), ta, true},
		{"computed-access neither tainted", sor_memberComputed(sor_ident("z"), sor_ident("z")), ta, false},

		// ── Nested: taint propagates through composite expressions ────────────
		{"nested tainted inside binary inside call",
			sor_call("f", sor_binary(sor_ident("z"), "+", sor_ident("x"))), ta, true},
		{"nested untainted inside binary inside call",
			sor_call("f", sor_binary(sor_ident("z"), "+", sor_ident("z"))), ta, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exprMentionsTainted(tt.expr, tt.tainted)
			if got != tt.want {
				t.Errorf("want %v, got %v", tt.want, got)
			}
		})
	}
}
