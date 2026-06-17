package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeArrowDecl(name string) *ast.VariableDeclaration {
	return &ast.VariableDeclaration{
		Declarations: []ast.VariableDeclarator{
			{
				ID:   &ast.Identifier{Name: name},
				Init: &ast.ArrowFunctionExpression{},
			},
		},
	}
}

func makeNonArrowDecl(name string) *ast.VariableDeclaration {
	return &ast.VariableDeclaration{
		Declarations: []ast.VariableDeclarator{
			{
				ID:   &ast.Identifier{Name: name},
				Init: &ast.Literal{Value: 0.0},
			},
		},
	}
}

func makeSecurityCallDecl(varName, funcName string) *ast.VariableDeclaration {
	return &ast.VariableDeclaration{
		Declarations: []ast.VariableDeclarator{
			{
				ID: &ast.Identifier{Name: varName},
				Init: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "security"},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "SYM"},
						&ast.Literal{Value: "D"},
						&ast.CallExpression{
							Callee: &ast.Identifier{Name: funcName},
						},
					},
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// TestCollectArrowFunctionNames — unit coverage of the AST pre-analysis helper
// ---------------------------------------------------------------------------

func TestCollectArrowFunctionNames(t *testing.T) {
	tests := []struct {
		name  string
		nodes []ast.Node
		want  map[string]bool
	}{
		{
			name:  "empty program",
			nodes: nil,
			want:  map[string]bool{},
		},
		{
			name:  "single arrow function",
			nodes: []ast.Node{makeArrowDecl("myUDF")},
			want:  map[string]bool{"myUDF": true},
		},
		{
			name:  "multiple arrow functions",
			nodes: []ast.Node{makeArrowDecl("zigzag"), makeArrowDecl("trend"), makeArrowDecl("signal")},
			want:  map[string]bool{"zigzag": true, "trend": true, "signal": true},
		},
		{
			name:  "non-arrow (literal) declaration not collected",
			nodes: []ast.Node{makeNonArrowDecl("seriesVal"), makeArrowDecl("fn")},
			want:  map[string]bool{"fn": true},
		},
		{
			name: "non-VariableDeclaration nodes ignored",
			nodes: []ast.Node{
				&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "close"}},
				makeArrowDecl("fn"),
			},
			want: map[string]bool{"fn": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog := &ast.Program{Body: tt.nodes}
			got := collectArrowFunctionNames(prog)

			for name := range tt.want {
				if !got[name] {
					t.Errorf("expected name %q to be collected", name)
				}
			}
			for name := range got {
				if !tt.want[name] {
					t.Errorf("unexpected name %q collected", name)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestDetectSecurityUDFEvals — unit coverage of the program-level pre-analysis
// ---------------------------------------------------------------------------

func TestDetectSecurityUDFEvals(t *testing.T) {
	tests := []struct {
		name  string
		nodes []ast.Node
		want  bool
	}{
		{
			name:  "UDF defined and passed to security() → true",
			nodes: []ast.Node{makeArrowDecl("myFunc"), makeSecurityCallDecl("v", "myFunc")},
			want:  true,
		},
		{
			name:  "UDF defined but not used in security() → false",
			nodes: []ast.Node{makeArrowDecl("myFunc"), makeNonArrowDecl("v")},
			want:  false,
		},
		{
			name:  "security() present but 3rd arg is not a UDF name → false",
			nodes: []ast.Node{makeArrowDecl("myFunc"), makeSecurityCallDecl("v", "unknownFn")},
			want:  false,
		},
		{
			name:  "no arrow functions at all → false",
			nodes: []ast.Node{makeSecurityCallDecl("v", "myFunc")},
			want:  false,
		},
		{
			name:  "empty program → false",
			nodes: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog := &ast.Program{Body: tt.nodes}
			got := detectSecurityUDFEvals(prog)
			if got != tt.want {
				t.Errorf("detectSecurityUDFEvals = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestSecurityUDFCodegen_BarLoop — integration: bar-loop dispatch path
// ---------------------------------------------------------------------------

func TestSecurityUDFCodegen_BarLoop(t *testing.T) {
	const udfBarLoopScript = `//@version=4
strategy("UDF Security Bar Loop", overlay=true)
myUDF() =>
    _v = 0.0
    _v := nz(_v[1]) + close
    _v
sz = security(syminfo.tickerid, "60", myUDF())
plot(sz)`

	tests := []struct {
		name           string
		script         string
		mustContain    []string
		mustNotContain []string
	}{
		{
			name:   "UDF emits UDFBarEvaluator not TA streaming evaluator",
			script: udfBarLoopScript,
			mustContain: []string{
				"security.NewUDFBarEvaluator",
				"szSeries.Set",
			},
			mustNotContain: []string{
				"ta.myUDF",
				"secBarEvaluator.EvaluateAtBar",
			},
		},
		{
			name:   "var secUDFBarEvaluators declared when UDF present",
			script: udfBarLoopScript,
			mustContain: []string{
				"var secUDFBarEvaluators map[string]security.BarEvaluator",
			},
		},
		{
			name:   "secUDFBarEvaluators suppression prevents unused-var error",
			script: udfBarLoopScript,
			mustContain: []string{
				"_ = secUDFBarEvaluators",
			},
		},
		{
			name:   "UDF bound to secondary-symbol arrow context, not main ctx",
			script: udfBarLoopScript,
			mustContain: []string{
				"secArrowCtx",
				"func(secArrowCtx *context.ArrowContext) float64",
			},
			mustNotContain: []string{
				"myUDF(ctx)",
				"myUDF(arrowCtx)",
			},
		},
		{
			name: "TA-only security does not emit UDF machinery",
			script: `//@version=4
strategy("TA Security", overlay=true)
val = security(syminfo.tickerid, "D", sma(close, 14))
plot(val)`,
			mustNotContain: []string{
				"security.NewUDFBarEvaluator",
				"secUDFBarEvaluators",
			},
		},
		{
			name: "self-referential UDF state persists via arrowCtx history",
			script: `//@version=4
strategy("Self-ref UDF", overlay=true)
zigzag() =>
    _direction = 0
    _direction := close > close[1] ? 1 : (close < close[1] ? -1 : nz(_direction[1]))
    _direction
sz = security(syminfo.tickerid, "60", zigzag())
plot(sz)`,
			mustContain: []string{
				"security.NewUDFBarEvaluator",
				"secArrowCtx",
			},
			mustNotContain: []string{
				"ta.zigzag",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.script)
			if err != nil {
				t.Fatalf("compile failed: %v", err)
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

// ---------------------------------------------------------------------------
// TestSecurityUDFCodegen_InlineIIFE — integration: inline IIFE dispatch path
// (security() used as an expression inside a ternary or condition)
// ---------------------------------------------------------------------------

func TestSecurityUDFCodegen_InlineIIFE(t *testing.T) {
	tests := []struct {
		name           string
		script         string
		mustContain    []string
		mustNotContain []string
	}{
		{
			name: "UDF in ternary branch emits UDFBarEvaluator inside IIFE",
			script: `//@version=4
strategy("UDF Inline IIFE", overlay=true)
myFunc() =>
    _x = 0.0
    _x := nz(_x[1]) + 1.0
    _x
result = myFunc() > 0 ? security(syminfo.tickerid, "D", myFunc()) : 0.0
plot(result)`,
			mustContain: []string{
				"security.NewUDFBarEvaluator",
				"(func() float64 {",
				"secUDFBarEvaluators",
			},
			mustNotContain: []string{
				"ta.myFunc",
			},
		},
		{
			name: "TA-only in ternary does not emit UDF machinery",
			script: `//@version=4
strategy("TA Inline IIFE", overlay=true)
result = close > 0 ? security(syminfo.tickerid, "D", sma(close, 14)) : 0.0
plot(result)`,
			mustNotContain: []string{
				"security.NewUDFBarEvaluator",
				"secUDFBarEvaluators",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.script)
			if err != nil {
				t.Fatalf("compile failed: %v", err)
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

// ---------------------------------------------------------------------------
// TestSecurityUDFCodegen_EvaluatorKeying — evaluator isolation between
// distinct (symbol, function) pairs and between distinct functions
// ---------------------------------------------------------------------------

func TestSecurityUDFCodegen_EvaluatorKeying(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		mustContain []string
	}{
		{
			name: "two distinct UDFs produce two distinct cache keys",
			script: `//@version=4
strategy("Two UDFs", overlay=true)
trend() =>
    close > open ? 1.0 : -1.0
momentum() =>
    close - close[1]
t = security(syminfo.tickerid, "D", trend())
m = security(syminfo.tickerid, "D", momentum())
plot(t + m)`,
			mustContain: []string{
				`":trend"`,
				`":momentum"`,
				"security.NewUDFBarEvaluator",
			},
		},
		{
			name: "same UDF with two different symbols uses distinct key per symbol",
			script: `//@version=4
strategy("Same UDF Two Symbols", overlay=true)
dir() =>
    _d = 0.0
    _d := nz(_d[1]) + 1.0
    _d
a = security("AAPL", "D", dir())
b = security("MSFT", "D", dir())
plot(a + b)`,
			mustContain: []string{
				`":dir"`,
				"security.NewUDFBarEvaluator",
				"udfKey := secKey",
			},
		},
		{
			name: "UDF key never collides with TA builtin key",
			script: `//@version=4
strategy("UDF vs TA keys", overlay=true)
sma14() =>
    _s = 0.0
    _s := nz(_s[1]) + close / 14.0
    _s
v = security(syminfo.tickerid, "D", sma14())
plot(v)`,
			mustContain: []string{
				`":sma14"`,
				"security.NewUDFBarEvaluator",
			},
			// should never route to StreamingBarEvaluator for this UDF
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.script)
			if err != nil {
				t.Fatalf("compile failed: %v", err)
			}
			v := NewCodeVerifier(code, t)
			v.MustContain(tt.mustContain...)
			if !strings.Contains(code, "security.NewUDFBarEvaluator") {
				t.Error("at least one UDFBarEvaluator must be constructed")
			}
		})
	}
}
