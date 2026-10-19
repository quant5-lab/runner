package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// ─── AST helpers (cod_ prefix avoids collision with identically-named helpers in
// nested_variable_scanner_test.go, udf_callsite_argument_type_scanner_test.go,
// and transformer_params_test.go which are all in package codegen) ───────────

func cod_buildProgram(stmts ...ast.Node) *ast.Program {
	return &ast.Program{Body: stmts}
}

func cod_varDecl(name string, init ast.Expression) *ast.VariableDeclaration {
	return &ast.VariableDeclaration{
		Declarations: []ast.VariableDeclarator{
			{ID: &ast.Identifier{Name: name}, Init: init},
		},
	}
}

func cod_arrowFunc(body ...ast.Node) *ast.ArrowFunctionExpression {
	return &ast.ArrowFunctionExpression{Body: body}
}

func cod_callExpr(funcName string, args ...ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee:    &ast.Identifier{Name: funcName},
		Arguments: args,
	}
}

func cod_exprStmt(expr ast.Expression) *ast.ExpressionStatement {
	return &ast.ExpressionStatement{Expression: expr}
}

func cod_ident(name string) *ast.Identifier {
	return &ast.Identifier{Name: name}
}

// ─── TestChartOnlyUDFDetector_BodyClassification ─────────────────────────────

func TestChartOnlyUDFDetector_BodyClassification(t *testing.T) {
	tests := []struct {
		name       string
		buildProg  func() *ast.Program
		wantChart  []string // these UDF names must be in the result
		wantNormal []string // these UDF names must NOT be in the result
	}{
		{
			// label.* calls only → chart-only
			name: "label namespace marks UDF chart-only",
			buildProg: func() *ast.Program {
				body := cod_arrowFunc(
					cod_exprStmt(cod_callExpr("label.new", cod_ident("t"), cod_ident("v"))),
					cod_exprStmt(cod_callExpr("label.delete", cod_ident("lbl"))),
					cod_exprStmt(cod_callExpr("label.set_text", cod_ident("lbl"), cod_ident("txt"))),
				)
				return cod_buildProgram(cod_varDecl("f_label", body))
			},
			wantChart: []string{"f_label"},
		},
		{
			// line.* calls only → chart-only
			name: "line namespace marks UDF chart-only",
			buildProg: func() *ast.Program {
				body := cod_arrowFunc(
					cod_exprStmt(cod_callExpr("line.new", cod_ident("x1"), cod_ident("y1"), cod_ident("x2"), cod_ident("y2"))),
					cod_exprStmt(cod_callExpr("line.delete", cod_ident("ln"))),
				)
				return cod_buildProgram(cod_varDecl("f_line", body))
			},
			wantChart: []string{"f_line"},
		},
		{
			// box.* calls only → chart-only
			name: "box namespace marks UDF chart-only",
			buildProg: func() *ast.Program {
				body := cod_arrowFunc(
					cod_exprStmt(cod_callExpr("box.new", cod_ident("l"), cod_ident("t"), cod_ident("r"), cod_ident("b"))),
				)
				return cod_buildProgram(cod_varDecl("f_box", body))
			},
			wantChart: []string{"f_box"},
		},
		{
			// table.* calls only → chart-only
			name: "table namespace marks UDF chart-only",
			buildProg: func() *ast.Program {
				body := cod_arrowFunc(
					cod_exprStmt(cod_callExpr("table.new", cod_ident("pos"), cod_ident("cols"), cod_ident("rows"))),
					cod_exprStmt(cod_callExpr("table.cell", cod_ident("tbl"), cod_ident("c"), cod_ident("r"), cod_ident("txt"))),
				)
				return cod_buildProgram(cod_varDecl("f_table", body))
			},
			wantChart: []string{"f_table"},
		},
		{
			// linefill.* calls only → chart-only
			name: "linefill namespace marks UDF chart-only",
			buildProg: func() *ast.Program {
				body := cod_arrowFunc(
					cod_exprStmt(cod_callExpr("linefill.new", cod_ident("l1"), cod_ident("l2"), cod_ident("col"))),
				)
				return cod_buildProgram(cod_varDecl("f_fill", body))
			},
			wantChart: []string{"f_fill"},
		},
		{
			// TA call → NOT chart-only
			name: "TA call in body is not chart-only",
			buildProg: func() *ast.Program {
				body := cod_arrowFunc(cod_exprStmt(cod_callExpr("ta.sma", cod_ident("close"), cod_ident("len"))))
				return cod_buildProgram(cod_varDecl("computeMA", body))
			},
			wantNormal: []string{"computeMA"},
		},
		{
			// Empty body → NOT chart-only (avoids silent data loss)
			name: "empty body is not chart-only",
			buildProg: func() *ast.Program {
				return cod_buildProgram(cod_varDecl("empty", cod_arrowFunc()))
			},
			wantNormal: []string{"empty"},
		},
		{
			// Mixed body (chart + TA) → NOT chart-only
			name: "mixed body with chart and TA calls is not chart-only",
			buildProg: func() *ast.Program {
				body := cod_arrowFunc(
					cod_exprStmt(cod_callExpr("label.delete", cod_ident("x"))),
					cod_exprStmt(cod_callExpr("ta.sma", cod_ident("close"), &ast.Literal{Value: float64(14)})),
				)
				return cod_buildProgram(cod_varDecl("f_mixed", body))
			},
			wantNormal: []string{"f_mixed"},
		},
		{
			// VariableDeclaration inside body whose init is a chart call → chart-only
			name: "var decl with chart-call init is chart-only",
			buildProg: func() *ast.Program {
				body := cod_arrowFunc(
					cod_varDecl("lbl", cod_callExpr("label.new", cod_ident("t"), cod_ident("v"))),
				)
				return cod_buildProgram(cod_varDecl("f_varinit", body))
			},
			wantChart: []string{"f_varinit"},
		},
		{
			// VariableDeclaration inside body whose init is NOT a chart call → NOT chart-only
			name: "var decl with non-chart init is not chart-only",
			buildProg: func() *ast.Program {
				body := cod_arrowFunc(
					cod_varDecl("x", &ast.BinaryExpression{
						Left: cod_ident("a"), Operator: "+", Right: cod_ident("b"),
					}),
				)
				return cod_buildProgram(cod_varDecl("f_math", body))
			},
			wantNormal: []string{"f_math"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewChartOnlyUDFDetector().Detect(tt.buildProg())
			for _, name := range tt.wantChart {
				if !got[name] {
					t.Errorf("%s: expected chart-only, got not-chart-only", name)
				}
			}
			for _, name := range tt.wantNormal {
				if got[name] {
					t.Errorf("%s: expected not-chart-only, got chart-only", name)
				}
			}
		})
	}
}

// ─── TestChartOnlyUDFDetector_TransitiveClassification ───────────────────────

func TestChartOnlyUDFDetector_TransitiveClassification(t *testing.T) {
	tests := []struct {
		name       string
		buildProg  func() *ast.Program
		wantChart  []string
		wantNormal []string
	}{
		{
			// f_colorscr (non-chart body) only used as arg to label.new → transitively chart-only
			name: "UDF only used as arg to chart call is transitively chart-only",
			buildProg: func() *ast.Program {
				colorBody := cod_arrowFunc(cod_exprStmt(&ast.ConditionalExpression{
					Test: cod_ident("v"), Consequent: &ast.Literal{Value: "#FF0000"}, Alternate: cod_ident("na"),
				}))
				printBody := cod_arrowFunc(
					cod_exprStmt(cod_callExpr("label.new", cod_ident("t"), cod_callExpr("f_colorscr", cod_ident("showscr")))),
				)
				return cod_buildProgram(
					cod_varDecl("f_colorscr", colorBody),
					cod_varDecl("f_print", printBody),
				)
			},
			wantChart:  []string{"f_print", "f_colorscr"},
			wantNormal: nil,
		},
		{
			// udf_B's body is all chart calls (direct Phase 1), so udf_A (only ever called
			// as an arg inside udf_B) is promoted to chart-only in Phase 2.
			name: "helper only called as arg inside chart-only UDF body is promoted via phase 2",
			buildProg: func() *ast.Program {
				aBody := cod_arrowFunc(cod_exprStmt(&ast.Literal{Value: "#AABBCC"}))
				bBody := cod_arrowFunc(
					cod_exprStmt(cod_callExpr("line.new",
						cod_ident("x1"), cod_ident("y1"), cod_ident("x2"),
						cod_callExpr("udf_A"),
					)),
				)
				return cod_buildProgram(
					cod_varDecl("udf_A", aBody),
					cod_varDecl("udf_B", bBody),
				)
			},
			wantChart: []string{"udf_A", "udf_B"},
		},
		{
			// A UDF reachable from both chart and non-chart call sites must stay non-chart-only;
			// eliding it would silently drop its numeric return value at the non-chart site.
			name: "UDF called in both chart and non-chart position is not chart-only",
			buildProg: func() *ast.Program {
				helperBody := cod_arrowFunc(cod_exprStmt(&ast.BinaryExpression{
					Left: cod_ident("v"), Operator: "+", Right: &ast.Literal{Value: float64(1)},
				}))
				drawBody := cod_arrowFunc(
					cod_exprStmt(cod_callExpr("label.new", cod_ident("t"), cod_callExpr("helper"))),
				)
				return cod_buildProgram(
					cod_varDecl("helper", helperBody),
					cod_varDecl("draw", drawBody),
					cod_varDecl("x", cod_callExpr("helper")),
				)
			},
			wantNormal: []string{"helper"},
			wantChart:  []string{"draw"},
		},
		{
			// Dead UDFs with no call sites must not be falsely promoted to chart-only,
			// as they may be called from generated code paths invisible to the AST walk.
			name: "UDF with no call sites is not classified as chart-only",
			buildProg: func() *ast.Program {
				body := cod_arrowFunc(cod_exprStmt(&ast.Literal{Value: float64(42)}))
				return cod_buildProgram(cod_varDecl("dead_fn", body))
			},
			wantNormal: []string{"dead_fn"},
		},
		{
			// Phase 1 marks alpha chart-only; Phase 2 must then see that colorFn's only
			// call site is inside alpha's body (already chart-only) and promote it too.
			name: "phase 1 chart-only UDF body call site promotes helper via phase 2",
			buildProg: func() *ast.Program {
				// colorFn: non-chart body
				colorFnBody := cod_arrowFunc(cod_exprStmt(&ast.ConditionalExpression{
					Test: cod_ident("v"), Consequent: &ast.Literal{Value: "#FF0000"}, Alternate: cod_ident("na"),
				}))
				// alpha: direct chart-only; calls colorFn only as arg to label.new
				alphaBody := cod_arrowFunc(
					cod_exprStmt(cod_callExpr("label.new", cod_ident("t"), cod_callExpr("colorFn", cod_ident("flag")))),
				)
				return cod_buildProgram(
					cod_varDecl("colorFn", colorFnBody),
					cod_varDecl("alpha", alphaBody),
				)
			},
			wantChart: []string{"alpha", "colorFn"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewChartOnlyUDFDetector().Detect(tt.buildProg())
			for _, name := range tt.wantChart {
				if !got[name] {
					t.Errorf("%s: expected chart-only, got not-chart-only", name)
				}
			}
			for _, name := range tt.wantNormal {
				if got[name] {
					t.Errorf("%s: expected not-chart-only, got chart-only", name)
				}
			}
		})
	}
}

// ─── TestFilterChartOnlyCallSites ────────────────────────────────────────────

func TestFilterChartOnlyCallSites(t *testing.T) {
	tests := []struct {
		name      string
		sites     []ArrowCallSite
		chartOnly map[string]bool
		wantNames []string // FunctionName values expected in output, in order
	}{
		{
			name:      "empty input returns empty",
			sites:     nil,
			chartOnly: map[string]bool{"f": true},
			wantNames: nil,
		},
		{
			name:      "empty chart-only set returns all sites unchanged",
			sites:     []ArrowCallSite{{FunctionName: "f1"}, {FunctionName: "f2"}},
			chartOnly: map[string]bool{},
			wantNames: []string{"f1", "f2"},
		},
		{
			name:      "nil chart-only set returns all sites unchanged",
			sites:     []ArrowCallSite{{FunctionName: "f1"}},
			chartOnly: nil,
			wantNames: []string{"f1"},
		},
		{
			name: "removes only chart-only sites preserving order",
			sites: []ArrowCallSite{
				{FunctionName: "draw"},
				{FunctionName: "compute"},
				{FunctionName: "render"},
			},
			chartOnly: map[string]bool{"draw": true, "render": true},
			wantNames: []string{"compute"},
		},
		{
			name: "all chart-only returns empty",
			sites: []ArrowCallSite{
				{FunctionName: "f_print"},
				{FunctionName: "f_color"},
			},
			chartOnly: map[string]bool{"f_print": true, "f_color": true},
			wantNames: nil,
		},
		{
			name: "no chart-only sites returns all",
			sites: []ArrowCallSite{
				{FunctionName: "trend"},
				{FunctionName: "signal"},
			},
			chartOnly: map[string]bool{"draw": true},
			wantNames: []string{"trend", "signal"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterChartOnlyCallSites(tt.sites, tt.chartOnly)
			if len(got) != len(tt.wantNames) {
				t.Fatalf("len: want %d, got %d — names: %v", len(tt.wantNames), len(got), got)
			}
			for i, want := range tt.wantNames {
				if got[i].FunctionName != want {
					t.Errorf("[%d]: want %q, got %q", i, want, got[i].FunctionName)
				}
			}
		})
	}
}

// ─── TestArrowCaptureRegistry_AppendCallArgs ─────────────────────────────────
// Register/Get behaviour is covered by TestArrowCaptureRegistry in
// outer_scope_capture_test.go — this test covers only AppendCallArgs.

func TestArrowCaptureRegistry_AppendCallArgs(t *testing.T) {
	tests := []struct {
		name       string
		captures   []OuterScopeCapture
		constants  map[string]interface{}
		prefix     []string // existing args before the call
		wantSuffix []string // the appended portion only
	}{
		{
			name:       "no captures leaves prefix unchanged",
			captures:   []OuterScopeCapture{},
			constants:  nil,
			prefix:     []string{"arrowCtx"},
			wantSuffix: nil,
		},
		{
			name:       "SeriesFloat capture appends <name>Series",
			captures:   []OuterScopeCapture{{Name: "price", Kind: OuterScopeCaptureSeriesFloat}},
			constants:  nil,
			prefix:     []string{"arrowCtx"},
			wantSuffix: []string{"priceSeries"},
		},
		{
			name:       "ArraySeries capture appends <name>ArraySeries",
			captures:   []OuterScopeCapture{{Name: "levels", Kind: OuterScopeCaptureArraySeries}},
			constants:  nil,
			prefix:     []string{"arrowCtx"},
			wantSuffix: []string{"levelsArraySeries"},
		},
		{
			name:       "StringArraySeries capture appends <name>StringArraySeries",
			captures:   []OuterScopeCapture{{Name: "labels", Kind: OuterScopeCaptureStringArraySeries}},
			constants:  nil,
			prefix:     []string{"arrowCtx"},
			wantSuffix: []string{"labelsStringArraySeries"},
		},
		{
			name:       "String capture appends bare name",
			captures:   []OuterScopeCapture{{Name: "ticker", Kind: OuterScopeCaptureString}},
			constants:  nil,
			prefix:     []string{"arrowCtx"},
			wantSuffix: []string{"ticker"},
		},
		{
			name:       "Scalar non-bool constant appends bare name",
			captures:   []OuterScopeCapture{{Name: "length", Kind: OuterScopeCaptureScalar}},
			constants:  map[string]interface{}{"length": float64(14)},
			prefix:     []string{"arrowCtx"},
			wantSuffix: []string{"length"},
		},
		{
			// bool constants cannot be passed as float64 directly; the IIFE bridges the type gap
			// without introducing a temporary variable at every call site.
			name:       "Scalar bool constant is bridged to float64 IIFE",
			captures:   []OuterScopeCapture{{Name: "showLabel", Kind: OuterScopeCaptureScalar}},
			constants:  map[string]interface{}{"showLabel": false},
			prefix:     []string{"arrowCtx"},
			wantSuffix: []string{"func() float64 { if showLabel { return 1.0 } else { return 0.0 } }()"},
		},
		{
			name:       "Scalar capture with no constants entry appends bare name",
			captures:   []OuterScopeCapture{{Name: "mult", Kind: OuterScopeCaptureScalar}},
			constants:  map[string]interface{}{},
			prefix:     []string{"arrowCtx"},
			wantSuffix: []string{"mult"},
		},
		{
			name:       "nil constants map does not panic",
			captures:   []OuterScopeCapture{{Name: "period", Kind: OuterScopeCaptureScalar}},
			constants:  nil,
			prefix:     []string{"arrowCtx"},
			wantSuffix: []string{"period"},
		},
		{
			// Multiple captures of different kinds preserve order
			name: "multiple mixed captures appended in registration order",
			captures: []OuterScopeCapture{
				{Name: "src", Kind: OuterScopeCaptureSeriesFloat},
				{Name: "len", Kind: OuterScopeCaptureScalar},
				{Name: "sym", Kind: OuterScopeCaptureString},
			},
			constants:  map[string]interface{}{"len": float64(20)},
			prefix:     []string{"ctx"},
			wantSuffix: []string{"srcSeries", "len", "sym"},
		},
		{
			// Unknown function: original slice returned unchanged (len==prefix len)
			// — tested separately below to keep this table clean
		},
	}

	for _, tt := range tests {
		if tt.name == "" {
			continue // skip placeholder
		}
		t.Run(tt.name, func(t *testing.T) {
			r := NewArrowCaptureRegistry()
			r.Register("fn", tt.captures)

			got := r.AppendCallArgs(append([]string(nil), tt.prefix...), "fn", tt.constants)

			wantLen := len(tt.prefix) + len(tt.wantSuffix)
			if len(got) != wantLen {
				t.Fatalf("len: want %d, got %d — %v", wantLen, len(got), got)
			}
			for i, want := range tt.wantSuffix {
				idx := len(tt.prefix) + i
				if got[idx] != want {
					t.Errorf("suffix[%d]: want %q, got %q", i, want, got[idx])
				}
			}
		})
	}

	t.Run("unknown function leaves prefix unchanged", func(t *testing.T) {
		r := NewArrowCaptureRegistry()
		got := r.AppendCallArgs([]string{"arrowCtx", "x"}, "unknown", nil)
		if len(got) != 2 {
			t.Errorf("want 2 args unchanged, got %d: %v", len(got), got)
		}
	})

	t.Run("multiple registered functions do not interfere", func(t *testing.T) {
		r := NewArrowCaptureRegistry()
		r.Register("f1", []OuterScopeCapture{{Name: "a", Kind: OuterScopeCaptureSeriesFloat}})
		r.Register("f2", []OuterScopeCapture{{Name: "b", Kind: OuterScopeCaptureString}})

		got1 := r.AppendCallArgs([]string{"ctx"}, "f1", nil)
		got2 := r.AppendCallArgs([]string{"ctx"}, "f2", nil)

		if len(got1) != 2 || got1[1] != "aSeries" {
			t.Errorf("f1: want [ctx aSeries], got %v", got1)
		}
		if len(got2) != 2 || got2[1] != "b" {
			t.Errorf("f2: want [ctx b], got %v", got2)
		}
	})

	t.Run("bool IIFE is distinguishable from bare name", func(t *testing.T) {
		// Explicitly guard the semantic contract: bool-scalar must produce an IIFE,
		// not the raw variable name "showLabel".
		r := NewArrowCaptureRegistry()
		r.Register("fn", []OuterScopeCapture{{Name: "showLabel", Kind: OuterScopeCaptureScalar}})

		withBool := r.AppendCallArgs([]string{}, "fn", map[string]interface{}{"showLabel": true})
		withFloat := r.AppendCallArgs([]string{}, "fn", map[string]interface{}{"showLabel": float64(1)})

		if withBool[0] == "showLabel" {
			t.Error("bool-scalar must emit IIFE, not bare name — GoParamName vs GoCallSiteExpression contract broken")
		}
		if !strings.Contains(withBool[0], "func()") {
			t.Errorf("bool-scalar IIFE expected 'func()' in expression, got: %q", withBool[0])
		}
		if withFloat[0] != "showLabel" {
			t.Errorf("float-scalar should emit bare name, got: %q", withFloat[0])
		}
	})
}
