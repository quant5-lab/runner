package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func makeUDFCallExpr(name string, args ...ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee:    &ast.Identifier{Name: name},
		Arguments: args,
	}
}

// TestSecurityUDFCallGenerator_IsUDFCall verifies that only registered user-defined
// functions are recognised as UDF calls.
func TestSecurityUDFCallGenerator_IsUDFCall(t *testing.T) {
	g := newTestGenerator()
	g.variables["myFunc"] = "function"
	gen := NewSecurityUDFCallGenerator(g)

	tests := []struct {
		name string
		expr ast.Expression
		want bool
	}{
		{
			name: "registered UDF is recognised",
			expr: makeUDFCallExpr("myFunc"),
			want: true,
		},
		{
			name: "unknown function is not a UDF",
			expr: makeUDFCallExpr("ta.sma"),
			want: false,
		},
		{
			name: "non-call expression is never a UDF",
			expr: &ast.Identifier{Name: "myFunc"},
			want: false,
		},
		{
			name: "built-in identifier is not a UDF",
			expr: makeUDFCallExpr("plot"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gen.IsUDFCall(tt.expr)
			if got != tt.want {
				t.Errorf("IsUDFCall = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSecurityUDFCallGenerator_EmitSingleCall verifies that a single-return UDF call
// in security context creates an arrow context bound to secCtx (not ctx) and stores
// the result in varNameSeries.
func TestSecurityUDFCallGenerator_EmitSingleCall(t *testing.T) {
	g := newTestGenerator()
	g.variables["zigzag"] = "function"
	gen := NewSecurityUDFCallGenerator(g)

	code, err := gen.EmitSingleCall("sz", makeUDFCallExpr("zigzag"))
	if err != nil {
		t.Fatalf("EmitSingleCall: %v", err)
	}

	v := NewCodeVerifier(code, t)
	v.MustContain(
		"context.NewArrowContext(secCtx)",
		"zigzag(",
		"szSeries.Set(",
	)
	v.MustNotContain(
		"context.NewArrowContext(ctx)",
	)
}

// TestSecurityUDFCallGenerator_EmitTupleCall verifies that a multi-return UDF call
// in security context unpacks all return values and stores each in its series.
func TestSecurityUDFCallGenerator_EmitTupleCall(t *testing.T) {
	g := newTestGenerator()
	g.variables["Pmax"] = "function"
	gen := NewSecurityUDFCallGenerator(g)

	varNames := []string{"trend", "tsl"}
	code, err := gen.EmitTupleCall(varNames, makeUDFCallExpr("Pmax",
		&ast.Literal{Value: 3.0},
		&ast.Literal{Value: 10.0},
	))
	if err != nil {
		t.Fatalf("EmitTupleCall: %v", err)
	}

	v := NewCodeVerifier(code, t)
	v.MustContain(
		"context.NewArrowContext(secCtx)",
		"Pmax(",
		"trend, tsl :=",
		"trendSeries.Set(",
		"tslSeries.Set(",
	)
	v.MustNotContain("context.NewArrowContext(ctx)")
}

// TestSecurityUDFCallGenerator_SecContextBinding verifies the generator uses "secCtx"
// as the parent context regardless of how many UDF calls are emitted.
func TestSecurityUDFCallGenerator_SecContextBinding(t *testing.T) {
	g := newTestGenerator()
	g.variables["fn"] = "function"
	gen := NewSecurityUDFCallGenerator(g)

	for i, varName := range []string{"a", "b", "c"} {
		code, err := gen.EmitSingleCall(varName, makeUDFCallExpr("fn"))
		if err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
		if strings.Contains(code, "NewArrowContext(ctx)") {
			t.Errorf("call %d: bound to main ctx instead of secCtx:\n%s", i, code)
		}
		if !strings.Contains(code, "NewArrowContext(secCtx)") {
			t.Errorf("call %d: missing secCtx binding:\n%s", i, code)
		}
	}
}

// TestSecurityUDFCallGenerator_EmitBarLoopUDFEval verifies code emitted for a UDF
// security() call in the bar-loop path (variable declaration position).
func TestSecurityUDFCallGenerator_EmitBarLoopUDFEval(t *testing.T) {
	tests := []struct {
		name        string
		funcName    string
		varName     string
		mustContain []string
	}{
		{
			name:     "emits UDFBarEvaluator construction",
			funcName: "zigzag",
			varName:  "sz",
			mustContain: []string{
				"security.NewUDFBarEvaluator",
				"len(secCtx.Data)",
				"secCtx",
			},
		},
		{
			name:     "emits per-func cache key",
			funcName: "trend",
			varName:  "val",
			mustContain: []string{
				`":trend"`,
				"udfKey := secKey",
				"secUDFBarEvaluators",
			},
		},
		{
			name:     "emits lazy map initialisation",
			funcName: "signal",
			varName:  "s",
			mustContain: []string{
				"secUDFBarEvaluators == nil",
				"make(map[string]security.BarEvaluator)",
			},
		},
		{
			name:     "calls UDF with secArrowCtx and stores result",
			funcName: "myUDF",
			varName:  "out",
			mustContain: []string{
				"func(secArrowCtx *context.ArrowContext) float64",
				"myUDF(secArrowCtx)",
				"EvaluateAtBar(nil, nil, secBarIdx)",
				"outSeries.Set(secUDFVal)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			gen := NewSecurityUDFCallGenerator(g)

			code, err := gen.EmitBarLoopUDFEval(tt.varName, makeUDFCallExpr(tt.funcName))
			if err != nil {
				t.Fatalf("EmitBarLoopUDFEval: %v", err)
			}

			v := NewCodeVerifier(code, t)
			v.MustContain(tt.mustContain...)
		})
	}
}

// TestSecurityUDFCallGenerator_EmitBarLoopUDFEval_SetsFlag verifies that emitting
// bar-loop UDF code marks the generator so the secUDFBarEvaluators var is declared.
func TestSecurityUDFCallGenerator_EmitBarLoopUDFEval_SetsFlag(t *testing.T) {
	g := newTestGenerator()
	if g.hasSecurityUDFEvals {
		t.Fatal("flag must be clear before emission")
	}

	gen := NewSecurityUDFCallGenerator(g)
	_, err := gen.EmitBarLoopUDFEval("v", makeUDFCallExpr("fn"))
	if err != nil {
		t.Fatalf("EmitBarLoopUDFEval: %v", err)
	}

	if !g.hasSecurityUDFEvals {
		t.Error("hasSecurityUDFEvals must be set after EmitBarLoopUDFEval")
	}
}

// TestSecurityUDFCallGenerator_EmitArrowContextUDFEval verifies code emitted for a UDF
// security() call inside an arrow-function IIFE.
func TestSecurityUDFCallGenerator_EmitArrowContextUDFEval(t *testing.T) {
	tests := []struct {
		name           string
		funcName       string
		mustContain    []string
		mustNotContain []string
	}{
		{
			name:     "stores evaluator in arrowCtx evaluators map",
			funcName: "zigzag",
			mustContain: []string{
				"GetOrCreateSecurityEvaluators()",
				"arrowUDFMap",
			},
		},
		{
			name:     "keys evaluator with udf: prefix",
			funcName: "trend",
			mustContain: []string{
				`"udf:" + secKey`,
				`":trend"`,
			},
		},
		{
			name:     "creates UDFBarEvaluator with secArrowCtx closure",
			funcName: "myFunc",
			mustContain: []string{
				"security.NewUDFBarEvaluator",
				"func(secArrowCtx *context.ArrowContext) float64",
				"myFunc(secArrowCtx)",
			},
		},
		{
			name:     "type-asserts stored evaluator and evaluates at bar",
			funcName: "dir",
			mustContain: []string{
				".(security.BarEvaluator).EvaluateAtBar(nil, nil, secBarIdx)",
				"return secUDFVal",
			},
		},
		{
			name:     "does not reference bar-scope secUDFBarEvaluators",
			funcName: "fn",
			mustNotContain: []string{
				"secUDFBarEvaluators",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			gen := NewSecurityUDFCallGenerator(g)

			code, err := gen.EmitArrowContextUDFEval(makeUDFCallExpr(tt.funcName))
			if err != nil {
				t.Fatalf("EmitArrowContextUDFEval: %v", err)
			}

			v := NewCodeVerifier(code, t)
			if len(tt.mustContain) > 0 {
				v.MustContain(tt.mustContain...)
			}
			if len(tt.mustNotContain) > 0 {
				v.MustNotContain(tt.mustNotContain...)
			}
		})
	}
}
