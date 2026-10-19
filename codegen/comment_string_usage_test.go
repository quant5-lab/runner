package codegen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestCollectStrategyCommentIdentifiers(t *testing.T) {
	tests := []struct {
		name string
		body []ast.Node
		want map[string]bool
	}{
		{
			name: "entry named comment identifier",
			body: []ast.Node{
				csuExprStmt(csuStrategyCall("entry",
					csuStrLit("L"),
					csuMember("strategy", "long"),
					csuNumLit(1),
					csuObjProp("comment", csuIdent("entry_msg")),
				)),
			},
			want: map[string]bool{"entry_msg": true},
		},
		{
			name: "entry positional comment identifier at real Pine slot",
			body: []ast.Node{
				csuExprStmt(csuStrategyCall("entry",
					csuStrLit("L"),
					csuMember("strategy", "long"),
					csuNumLit(1),
					csuIdent("na"),
					csuIdent("na"),
					csuIdent("na"),
					csuIdent("na"),
					csuIdent("entry_msg"),
				)),
			},
			want: map[string]bool{"entry_msg": true},
		},
		{
			name: "order positional comment identifier at real Pine slot",
			body: []ast.Node{
				csuExprStmt(csuStrategyCall("order",
					csuStrLit("O"),
					csuMember("strategy", "long"),
					csuNumLit(1),
					csuIdent("na"),
					csuIdent("na"),
					csuIdent("na"),
					csuIdent("na"),
					csuIdent("order_msg"),
				)),
			},
			want: map[string]bool{"order_msg": true},
		},
		{
			name: "close_all positional comment identifier at real Pine slot",
			body: []ast.Node{
				csuExprStmt(csuStrategyCall("close_all", csuIdent("close_all_msg"))),
			},
			want: map[string]bool{"close_all_msg": true},
		},
		{
			name: "exit positional comment identifier at real Pine slot",
			body: []ast.Node{
				csuExprStmt(csuStrategyCall("exit",
					csuStrLit("X"),
					csuStrLit("L"),
					csuIdent("na"), csuIdent("na"), csuIdent("na"), csuIdent("na"), csuIdent("na"),
					csuIdent("na"), csuIdent("na"), csuIdent("na"), csuIdent("na"), csuIdent("na"),
					csuIdent("exit_msg"),
				)),
			},
			want: map[string]bool{"exit_msg": true},
		},
		{
			name: "close positional comment identifier",
			body: []ast.Node{
				csuExprStmt(csuStrategyCall("close", csuStrLit("L"), csuIdent("close_msg"))),
			},
			want: map[string]bool{"close_msg": true},
		},
		{
			name: "ternary collects string carriers but not condition",
			body: []ast.Node{
				csuExprStmt(csuStrategyCall("entry",
					csuStrLit("L"),
					csuMember("strategy", "long"),
					csuNumLit(1),
					csuObjProp("comment", &ast.ConditionalExpression{
						Test:       csuIdent("is_long"),
						Consequent: csuIdent("long_msg"),
						Alternate:  csuIdent("short_msg"),
					}),
				)),
			},
			want: map[string]bool{"long_msg": true, "short_msg": true},
		},
		{
			name: "non-comment properties ignored",
			body: []ast.Node{
				csuExprStmt(csuStrategyCall("entry",
					csuStrLit("L"),
					csuMember("strategy", "long"),
					csuObjProp("qty", csuIdent("qty_value")),
				)),
			},
			want: map[string]bool{},
		},
		{
			name: "exit comment variants collect string carriers",
			body: []ast.Node{
				csuExprStmt(csuStrategyCall("exit",
					csuStrLit("X"),
					csuStrLit("L"),
					csuObjProp("comment_profit", csuIdent("profit_msg")),
					csuObjProp("comment_loss", csuIdent("loss_msg")),
					csuObjProp("comment_trailing", csuIdent("trailing_msg")),
				)),
			},
			want: map[string]bool{"profit_msg": true, "loss_msg": true, "trailing_msg": true},
		},
		{
			name: "alert text is not a comment string carrier",
			body: []ast.Node{
				csuExprStmt(csuStrategyCall("exit",
					csuStrLit("X"),
					csuStrLit("L"),
					csuObjProp("alert_message", csuIdent("alert_msg")),
				)),
			},
			want: map[string]bool{},
		},
		{
			name: "sanitizes reserved identifier before registry use",
			body: []ast.Node{
				csuExprStmt(csuStrategyCall("entry",
					csuStrLit("L"),
					csuMember("strategy", "long"),
					csuNumLit(1),
					csuObjProp("comment", csuIdent("type")),
				)),
			},
			want: map[string]bool{"type_": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectStrategyCommentIdentifiers(&ast.Program{NodeType: ast.TypeProgram, Body: tt.body})
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("collectStrategyCommentIdentifiers() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestCollectStrategyCommentIdentifiersNilProgram(t *testing.T) {
	got := collectStrategyCommentIdentifiers(nil)
	if len(got) != 0 {
		t.Fatalf("collectStrategyCommentIdentifiers(nil) = %#v, want empty", got)
	}
}

func TestBareCallName(t *testing.T) {
	tests := []struct {
		name string
		expr ast.Expression
		want string
		ok   bool
	}{
		{name: "bare call", expr: csuCall(csuIdent("unsupported_comment"), csuIdent("close")), want: "unsupported_comment", ok: true},
		{name: "dotted call", expr: csuCall(csuMember("str", "tostring"), csuIdent("close")), ok: false},
		{name: "non call", expr: csuIdent("unsupported_comment"), ok: false},
		{name: "nil", expr: nil, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := bareCallName(tt.expr)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("bareCallName() = (%q, %v), want (%q, %v)", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestInferDeclaredVariableTypePromotesUnknownCommentCarrierOnly(t *testing.T) {
	gen := newTestGenerator()
	gen.commentStringVars = map[string]bool{"comment_msg": true}
	gen.variables["label"] = "function"
	gen.typeSystem.RegisterUDFReturnType("label", "string")

	tests := []struct {
		name    string
		varName string
		expr    ast.Expression
		want    string
	}{
		{name: "unknown bare call used as comment carrier", varName: "comment_msg", expr: csuCall(csuIdent("unsupported_comment"), csuIdent("close")), want: "string"},
		{name: "unknown bare call not used as comment carrier stays numeric fallback", varName: "plot_value", expr: csuCall(csuIdent("unsupported_plot"), csuIdent("close")), want: "float64"},
		{name: "registered string UDF keeps inferred type", varName: "comment_msg", expr: csuCall(csuIdent("label"), csuIdent("close")), want: "string"},
		{name: "dotted call does not use unknown bare fallback", varName: "comment_msg", expr: csuCall(csuMember("str", "tostring"), csuIdent("close")), want: "float64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gen.inferDeclaredVariableType(tt.varName, tt.expr); got != tt.want {
				t.Fatalf("inferDeclaredVariableType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestArgumentExtractorCommentUnsupportedCallUsesStringSentinel(t *testing.T) {
	gen := newTestGenerator()
	extractor := &ArgumentExtractor{generator: gen}

	got := extractor.extractCommentValue(csuCall(csuIdent("unsupported_comment"), csuIdent("close")))
	want := `featuregap.RecordString("unsupported_comment", "generator.comment_unknown", "", 0, 0, ctx.BarIndex)`
	if got != want {
		t.Fatalf("extractCommentValue() = %q, want %q", got, want)
	}
	if !reflect.DeepEqual(gen.featureGaps, []string{"unsupported_comment"}) {
		t.Fatalf("featureGaps = %#v, want unsupported_comment", gen.featureGaps)
	}
}

func TestStrategyCommentDirectUDFCallUsesGeneratedStringExpression(t *testing.T) {
	code, err := compilePineScript(`//@version=5
strategy("Direct UDF comment")
label(x) => x > 0 ? "up" : "down"
strategy.entry("L", strategy.long, comment=label(close))
`)
	if err != nil {
		t.Fatalf("compilePineScript: %v", err)
	}
	if !strings.Contains(code, `strat.Entry("L", strategy.Long, 1, label(`) {
		t.Fatalf("strategy comment did not call string UDF:\n%s", code)
	}
	if strings.Contains(code, `strat.Entry("L", strategy.Long, 1, "")`) {
		t.Fatalf("strategy comment silently dropped direct UDF call:\n%s", code)
	}
}

func TestArgumentExtractorCommentIdentifierTypeBoundaries(t *testing.T) {
	gen := newTestGenerator()
	gen.variables["comment_msg"] = "string"
	gen.variables["risk_value"] = "float64"
	extractor := &ArgumentExtractor{generator: gen}

	tests := []struct {
		name string
		expr ast.Expression
		want string
	}{
		{name: "known string variable", expr: csuIdent("comment_msg"), want: "comment_msg"},
		{name: "known non-string variable", expr: csuIdent("risk_value"), want: `""`},
		{name: "unknown variable keeps legacy series access", expr: csuIdent("legacy_msg"), want: "legacy_msgSeries.GetCurrent()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractor.extractCommentValue(tt.expr); got != tt.want {
				t.Fatalf("extractCommentValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func csuExprStmt(expr ast.Expression) *ast.ExpressionStatement {
	return &ast.ExpressionStatement{Expression: expr}
}

func csuStrategyCall(name string, args ...ast.Expression) *ast.CallExpression {
	return csuCall(csuMember("strategy", name), args...)
}

func csuCall(callee ast.Expression, args ...ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{Callee: callee, Arguments: args}
}

func csuIdent(name string) *ast.Identifier {
	return &ast.Identifier{Name: name}
}

func csuMember(object, property string) *ast.MemberExpression {
	return &ast.MemberExpression{Object: csuIdent(object), Property: csuIdent(property)}
}

func csuStrLit(value string) *ast.Literal {
	return &ast.Literal{Value: value}
}

func csuNumLit(value float64) *ast.Literal {
	return &ast.Literal{Value: value}
}

func csuObjProp(name string, value ast.Expression) *ast.ObjectExpression {
	return &ast.ObjectExpression{
		NodeType: ast.TypeObjectExpression,
		Properties: []ast.Property{
			{NodeType: ast.TypeProperty, Key: csuIdent(name), Value: value},
		},
	}
}
