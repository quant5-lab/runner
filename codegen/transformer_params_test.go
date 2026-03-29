package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func callExpr(args ...ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "ticker.x"},
		Arguments: args,
	}
}

func strLit(v string) *ast.Literal      { return &ast.Literal{Value: v} }
func fltLit(v float64) *ast.Literal     { return &ast.Literal{Value: v} }
func ident(name string) *ast.Identifier { return &ast.Identifier{Name: name} }

func TestTransformerConstructorCode_NonCallExpression_FallsBackToNewTransformer(t *testing.T) {
	tests := []struct {
		prefix string
		want   string
	}{
		{"HEIKINASHI", `ticker.NewTransformer("HEIKINASHI")`},
		{"RENKO", `ticker.NewTransformer("RENKO")`},
		{"KAGI", `ticker.NewTransformer("KAGI")`},
		{"LINEBREAK", `ticker.NewTransformer("LINEBREAK")`},
		{"POINTFIG", `ticker.NewTransformer("POINTFIG")`},
		{"RANGE", `ticker.NewTransformer("RANGE")`},
		{"UNKNOWN", `ticker.NewTransformer("UNKNOWN")`},
	}

	expr := &ast.Literal{Value: "BTCUSDT"}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			got := transformerConstructorCode(expr, tt.prefix)
			if got != tt.want {
				t.Errorf("transformerConstructorCode(literal, %q) = %q, want %q", tt.prefix, got, tt.want)
			}
		})
	}
}

func TestTransformerConstructorCode_HeikinAshi_ReturnsInlineStruct(t *testing.T) {
	got := transformerConstructorCode(callExpr(), "HEIKINASHI")
	if got != "&ticker.HeikinAshiTransformer{}" {
		t.Errorf("HEIKINASHI = %q, want &ticker.HeikinAshiTransformer{}", got)
	}
}

func TestTransformerConstructorCode_Range_ReturnsIdentityStruct(t *testing.T) {
	got := transformerConstructorCode(callExpr(), "RANGE")
	if got != "&ticker.IdentityTransformer{}" {
		t.Errorf("RANGE = %q, want &ticker.IdentityTransformer{}", got)
	}
}

func TestTransformerConstructorCode_Renko_LiteralsForwardToTypedConstructor(t *testing.T) {
	tests := []struct {
		name  string
		style string
		param float64
		want  string
	}{
		{"ATR style", "ATR", 14.0, `ticker.NewRenkoTransformer("ATR", 14)`},
		{"Traditional style", "Traditional", 5.0, `ticker.NewRenkoTransformer("Traditional", 5)`},
		{"fractional box size", "ATR", 0.5, `ticker.NewRenkoTransformer("ATR", 0.5)`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := callExpr(strLit("BTCUSDT"), strLit(tt.style), fltLit(tt.param))
			got := transformerConstructorCode(expr, "RENKO")
			if got != tt.want {
				t.Errorf("RENKO literal = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTransformerConstructorCode_Renko_NonLiteralFallsBack(t *testing.T) {
	expr := callExpr(strLit("BTCUSDT"), ident("myStyle"), fltLit(14.0))
	got := transformerConstructorCode(expr, "RENKO")
	if !strings.HasPrefix(got, "ticker.NewTransformer(") {
		t.Errorf("RENKO non-literal = %q, want NewTransformer fallback", got)
	}
}

func TestTransformerConstructorCode_Kagi_LiteralForwardsToTypedConstructor(t *testing.T) {
	expr := callExpr(strLit("BTCUSDT"), fltLit(3.5))
	got := transformerConstructorCode(expr, "KAGI")
	want := "ticker.NewKagiTransformer(3.5)"
	if got != want {
		t.Errorf("KAGI literal = %q, want %q", got, want)
	}
}

func TestTransformerConstructorCode_Kagi_NonLiteralFallsBack(t *testing.T) {
	expr := callExpr(strLit("BTCUSDT"), ident("myReversal"))
	got := transformerConstructorCode(expr, "KAGI")
	if !strings.HasPrefix(got, "ticker.NewTransformer(") {
		t.Errorf("KAGI non-literal = %q, want NewTransformer fallback", got)
	}
}

func TestTransformerConstructorCode_LineBreak_LiteralForwardsToTypedConstructor(t *testing.T) {
	expr := callExpr(strLit("BTCUSDT"), fltLit(3.0))
	got := transformerConstructorCode(expr, "LINEBREAK")
	want := "ticker.NewLineBreakTransformer(3)"
	if got != want {
		t.Errorf("LINEBREAK literal = %q, want %q", got, want)
	}
}

func TestTransformerConstructorCode_LineBreak_NonLiteralFallsBack(t *testing.T) {
	expr := callExpr(strLit("BTCUSDT"), ident("lines"))
	got := transformerConstructorCode(expr, "LINEBREAK")
	if !strings.HasPrefix(got, "ticker.NewTransformer(") {
		t.Errorf("LINEBREAK non-literal = %q, want NewTransformer fallback", got)
	}
}

func TestTransformerConstructorCode_PointFig_AllLiteralsForwardToTypedConstructor(t *testing.T) {
	expr := callExpr(strLit("BTCUSDT"), strLit("close"), strLit("ATR"), fltLit(14.0), fltLit(3.0))
	got := transformerConstructorCode(expr, "POINTFIG")
	want := `ticker.NewPointFigureTransformer("close", "ATR", 14, 3)`
	if got != want {
		t.Errorf("POINTFIG literal = %q, want %q", got, want)
	}
}

func TestTransformerConstructorCode_PointFig_PartialLiteralsFallsBack(t *testing.T) {
	expr := callExpr(strLit("BTCUSDT"), strLit("close"), ident("style"), fltLit(14.0), fltLit(3.0))
	got := transformerConstructorCode(expr, "POINTFIG")
	if !strings.HasPrefix(got, "ticker.NewTransformer(") {
		t.Errorf("POINTFIG partial-literal = %q, want NewTransformer fallback", got)
	}
}

func TestTransformerConstructorCode_CaseInsensitive(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"heikinashi", "&ticker.HeikinAshiTransformer{}"},
		{"Heikinashi", "&ticker.HeikinAshiTransformer{}"},
		{"range", "&ticker.IdentityTransformer{}"},
		{"Range", "&ticker.IdentityTransformer{}"},
	}
	expr := callExpr()
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := transformerConstructorCode(expr, tt.input)
			if got != tt.want {
				t.Errorf("transformerConstructorCode(call, %q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsVariableBarCountModifier(t *testing.T) {
	tests := []struct {
		prefix   string
		variable bool
	}{
		{"RENKO", true},
		{"renko", true},
		{"Renko", true},
		{"KAGI", true},
		{"kagi", true},
		{"LINEBREAK", true},
		{"linebreak", true},
		{"POINTFIG", true},
		{"pointfig", true},

		{"HEIKINASHI", false},
		{"heikinashi", false},
		{"RANGE", false},
		{"range", false},
		{"", false},
		{"UNKNOWN", false},
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			got := isVariableBarCountModifier(tt.prefix)
			if got != tt.variable {
				t.Errorf("isVariableBarCountModifier(%q) = %v, want %v", tt.prefix, got, tt.variable)
			}
		})
	}
}
