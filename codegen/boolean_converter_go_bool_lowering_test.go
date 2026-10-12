package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/*
TestBooleanConverter_GoBoolLowering pins the lowering-domain question the series extractor
asks: whether emitted code already evaluates to a Go bool, so a condition position needs no
value.IsTrue coercion. It is narrower than IsAlreadyBoolean because the series extractor
lowers calls and logical expressions to float64 series and float64 IIFEs.
*/
func TestBooleanConverter_GoBoolLowering(t *testing.T) {
	converter := NewBooleanConverter(NewTypeInferenceEngine())

	tests := []struct {
		name          string
		expr          ast.Expression
		code          string
		wantAlreadyGo bool
		justification string
	}{
		{
			name:          "relational comparison lowers to a Go bool",
			expr:          &ast.BinaryExpression{Operator: ">", Left: &ast.Identifier{Name: "close"}, Right: &ast.Literal{Value: 100.0}},
			code:          "(bar.Close > 100)",
			wantAlreadyGo: true,
			justification: "a relational operand is a Go bool; value.IsTrue would reject it",
		},
		{
			name: "arithmetic over a comparison stays float64",
			expr: &ast.BinaryExpression{
				Operator: "*",
				Left:     &ast.BinaryExpression{Operator: ">", Left: &ast.Identifier{Name: "close"}, Right: &ast.Literal{Value: 100.0}},
				Right:    &ast.Literal{Value: 2.0},
			},
			code:          "((bar.Close > 100) * 2)",
			wantAlreadyGo: false,
			justification: "only relational and equality operators are Go bools; an arithmetic result never is",
		},
		{
			name:          "logical expression lowers to a float64 IIFE",
			expr:          &ast.LogicalExpression{Operator: "&&", Left: &ast.Identifier{Name: "a"}, Right: &ast.Identifier{Name: "b"}},
			code:          "func() float64 { if (value.IsTrue(aSeries.GetCurrent()) && value.IsTrue(bSeries.GetCurrent())) { return 1.0 }; return 0.0 }()",
			wantAlreadyGo: false,
			justification: "the series extractor lowers a logical expression to a float64 IIFE",
		},
		{
			name:          "bool-returning call lowers to a float64 series",
			expr:          &ast.CallExpression{Callee: &ast.MemberExpression{Object: &ast.Identifier{Name: "ta"}, Property: &ast.Identifier{Name: "crossover"}}},
			code:          "ta_crossover_4ce657bfSeries.GetCurrent()",
			wantAlreadyGo: false,
			justification: "IsAlreadyBoolean calls ta.crossover boolean, but the emitted binding is float64",
		},
		{
			name:          "series identifier stays float64",
			expr:          &ast.Identifier{Name: "close"},
			code:          "bar.Close",
			wantAlreadyGo: false,
			justification: "numeric truthiness needs the coercion",
		},
		{
			name:          "bool variable stays float64",
			expr:          &ast.Identifier{Name: "enabled"},
			code:          "enabledSeries.GetCurrent()",
			wantAlreadyGo: false,
			justification: "bool variables are stored as 0.0/1.0 series",
		},
		{
			name:          "member expression stays float64",
			expr:          &ast.MemberExpression{Object: &ast.Identifier{Name: "close"}, Property: &ast.Identifier{Name: "value"}},
			code:          "bar.Close",
			wantAlreadyGo: false,
			justification: "without a namespace type for the member its access is float64",
		},
		{
			name:          "unary minus stays float64",
			expr:          &ast.UnaryExpression{Operator: "-", Argument: &ast.Identifier{Name: "close"}},
			code:          "-bar.Close",
			wantAlreadyGo: false,
			justification: "unary minus is arithmetic, not a boolean operator",
		},
		{
			name:          "not-negated comparison lowers to a Go bool",
			expr:          &ast.UnaryExpression{Operator: "not", Argument: &ast.BinaryExpression{Operator: ">", Left: &ast.Identifier{Name: "close"}, Right: &ast.Literal{Value: 100.0}}},
			code:          "!(bar.Close > 100)",
			wantAlreadyGo: true,
			justification: "logical not over a Go bool is still a Go bool",
		},
		{
			name:          "bang-negated comparison lowers to a Go bool",
			expr:          &ast.UnaryExpression{Operator: "!", Argument: &ast.BinaryExpression{Operator: "<", Left: &ast.Identifier{Name: "close"}, Right: &ast.Literal{Value: 100.0}}},
			code:          "!(bar.Close < 100)",
			wantAlreadyGo: true,
			justification: "the bang spelling of logical not is the same decision",
		},
		{
			name:          "negated numeric operand is a Go bool once the lowering coerces it",
			expr:          &ast.UnaryExpression{Operator: "not", Argument: &ast.Identifier{Name: "close"}},
			code:          "!value.IsTrue(bar.Close)",
			wantAlreadyGo: true,
			justification: "the series extractor coerces the negated operand, so the emitted negation is a Go bool",
		},
		{
			name:          "bang-negated numeric operand is a Go bool once the lowering coerces it",
			expr:          &ast.UnaryExpression{Operator: "!", Argument: &ast.Identifier{Name: "close"}},
			code:          "!value.IsTrue(bar.Close)",
			wantAlreadyGo: true,
			justification: "both spellings of logical not reach the same decision",
		},
		{
			name:          "negated bool literal is a Go bool once the lowering coerces it",
			expr:          &ast.UnaryExpression{Operator: "not", Argument: &ast.Literal{Value: true}},
			code:          "!value.IsTrue(1.0)",
			wantAlreadyGo: true,
			justification: "a bool literal becomes 1.0 in the series domain, so the negation coerces it",
		},
		{
			name:          "inline boolean call is a Go bool",
			expr:          &ast.CallExpression{Callee: &ast.Identifier{Name: "na"}},
			code:          "math.IsNaN(bar.Close)",
			wantAlreadyGo: true,
			justification: "the value handler lowers na to a Go bool instead of a series binding",
		},
		{
			name:          "inline boolean call over a series accessor is a Go bool",
			expr:          &ast.CallExpression{Callee: &ast.Identifier{Name: "na"}},
			code:          "math.IsNaN(rsi14Series.GetCurrent())",
			wantAlreadyGo: true,
			justification: "a series accessor nested in the call does not make the call itself a series binding",
		},
		{
			name:          "numeric call lowering to a series accessor stays float64",
			expr:          &ast.CallExpression{Callee: &ast.MemberExpression{Object: &ast.Identifier{Name: "ta"}, Property: &ast.Identifier{Name: "sma"}}},
			code:          "ta_sma_20_4a6a8b06Series.GetCurrent()",
			wantAlreadyGo: false,
			justification: "a numeric call is a float64 series whatever its arguments",
		},
		{
			name:          "bool literal in the series domain stays coerced",
			expr:          &ast.Literal{Value: true},
			code:          "1.0",
			wantAlreadyGo: false,
			justification: "the series extractor stores a bool literal as 1.0, which still needs the coercion",
		},
		{
			name:          "Go bool literal passes through",
			expr:          &ast.Literal{Value: true},
			code:          "true",
			wantAlreadyGo: true,
			justification: "a bool literal is already a Go bool whatever the expression node says",
		},
		{
			name:          "numeric literal stays float64",
			expr:          &ast.Literal{Value: 100.0},
			code:          "100.0",
			wantAlreadyGo: false,
			justification: "a numeric literal is float64 in the series domain, where a bool literal lowers to 1.0/0.0",
		},
		{
			name:          "unresolvable expression falls back to coercion",
			expr:          nil,
			code:          "bar.Close",
			wantAlreadyGo: false,
			justification: "an expression that yields no decision must keep its coercion rather than emit a bare operand",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := converter.IsAlreadyGoBool(tt.expr, tt.code); got != tt.wantAlreadyGo {
				t.Errorf("IsAlreadyGoBool(%q) = %v, want %v: %s", tt.code, got, tt.wantAlreadyGo, tt.justification)
			}

			wantCode := tt.code
			if !tt.wantAlreadyGo {
				wantCode = "value.IsTrue(" + tt.code + ")"
			}
			if got := converter.EnsureGoBool(tt.expr, tt.code); got != wantCode {
				t.Errorf("EnsureGoBool(%q) = %q, want %q: %s", tt.code, got, wantCode, tt.justification)
			}
		})
	}
}

/*
TestBooleanConverter_ComparisonOperatorSet covers the whole relational and equality operator set
the series extractor lowers to a Go bool, so the decision does not depend on which comparison the
script happened to use.
*/
func TestBooleanConverter_ComparisonOperatorSet(t *testing.T) {
	converter := NewBooleanConverter(NewTypeInferenceEngine())

	for _, op := range []string{">", "<", ">=", "<=", "==", "!="} {
		t.Run("operator_"+op, func(t *testing.T) {
			expr := &ast.BinaryExpression{
				Operator: op,
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 100.0},
			}
			code := "(bar.Close " + op + " 100)"

			if !converter.IsAlreadyGoBool(expr, code) {
				t.Errorf("IsAlreadyGoBool(%q) = false, want true: %s lowers to a Go bool", code, op)
			}
			if got := converter.EnsureGoBool(expr, code); got != code {
				t.Errorf("EnsureGoBool(%q) = %q, want the operand unchanged", code, got)
			}
		})
	}
}

/*
TestBooleanConverter_GoBoolLowering_UnaryOperatorSet covers the unary operator set the decision
sees, so an operator that is not logical negation is never mistaken for one. Each case carries the
lowering the series extractor produces for that operator over a float64 operand.
*/
func TestBooleanConverter_GoBoolLowering_UnaryOperatorSet(t *testing.T) {
	converter := NewBooleanConverter(NewTypeInferenceEngine())

	operators := []struct {
		operator      string
		code          string
		wantAlreadyGo bool
	}{
		{"not", "!value.IsTrue(bar.Close)", true},
		{"!", "!value.IsTrue(bar.Close)", true},
		{"-", "-bar.Close", false},
		{"+", "+bar.Close", false},
		{"^", "^bar.Close", false},
	}

	for _, tt := range operators {
		t.Run("operator_"+tt.operator, func(t *testing.T) {
			expr := &ast.UnaryExpression{Operator: tt.operator, Argument: &ast.Identifier{Name: "close"}}

			if got := converter.IsAlreadyGoBool(expr, tt.code); got != tt.wantAlreadyGo {
				t.Errorf("IsAlreadyGoBool(%q) = %v, want %v", tt.code, got, tt.wantAlreadyGo)
			}
		})
	}
}

/*
TestBooleanConverter_GoBoolLowering_NamespaceOracle covers the member half of the decision: the
namespace resolver's Go type decides, so a member typed GoBool passes through while a member
typed float64 or string keeps its coercion. Each code sample is the text the resolver emits for
that member.
*/
func TestBooleanConverter_GoBoolLowering_NamespaceOracle(t *testing.T) {
	converter := NewBooleanConverter(NewTypeInferenceEngine())
	converter.SetGoTypeResolver(NewBuiltinIdentifierHandler())

	tests := []struct {
		name          string
		object        string
		property      string
		code          string
		wantAlreadyGo bool
	}{
		{"barstate.islast is a Go bool", "barstate", "islast", "(ctx.BarIndex == len(ctx.Data)-1)", true},
		{"timeframe.isintraday is a Go bool", "timeframe", "isintraday", "ctx.IsIntraday", true},
		{"session.isfirstbar compares a bool series", "session", "isfirstbar", "session_isfirstbarSeries.GetCurrent() == 1.0", true},
		{"timeframe.multiplier is float64", "timeframe", "multiplier", "float64(context.TimeframeMultiplier(ctx.Timeframe))", false},
		{"dayofweek.monday is float64", "dayofweek", "monday", "2.0", false},
		{"syminfo.tickerid is a string", "syminfo", "tickerid", "syminfo_tickerid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.MemberExpression{
				Object:   &ast.Identifier{Name: tt.object},
				Property: &ast.Identifier{Name: tt.property},
			}

			if got := converter.IsAlreadyGoBool(expr, tt.code); got != tt.wantAlreadyGo {
				t.Errorf("IsAlreadyGoBool(%s.%s) = %v, want %v", tt.object, tt.property, got, tt.wantAlreadyGo)
			}

			wantCode := tt.code
			if !tt.wantAlreadyGo {
				wantCode = "value.IsTrue(" + tt.code + ")"
			}
			if got := converter.EnsureGoBool(expr, tt.code); got != wantCode {
				t.Errorf("EnsureGoBool(%s.%s) = %q, want %q", tt.object, tt.property, got, wantCode)
			}
		})
	}
}
