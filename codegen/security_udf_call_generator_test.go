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
