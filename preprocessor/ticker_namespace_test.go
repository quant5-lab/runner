package preprocessor

import (
	"fmt"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

func parseScript(t *testing.T, input string) *parser.Script {
	t.Helper()
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("parser creation failed: %v", err)
	}
	ast, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	return ast
}

func extractFirstCallFromAssignment(t *testing.T, script *parser.Script) *parser.CallExpr {
	t.Helper()
	call := findCallInFactor(script.Statements[0].Core.Assignment.Value.Ternary.Condition.Left.Left.Left.Left.Left)
	if call == nil {
		t.Fatal("expected call expression in assignment value")
	}
	return call
}

func extractCallFromTernaryTrueBranch(t *testing.T, script *parser.Script) *parser.CallExpr {
	t.Helper()
	trueBranch := script.Statements[0].Core.Assignment.Value.Ternary.TrueVal
	call := findCallInFactor(trueBranch.Ternary.Condition.Left.Left.Left.Left.Left)
	if call == nil {
		t.Fatal("expected call expression in ternary true branch")
	}
	return call
}

func TestTickerNamespaceTransformer_OfficialV4Functions(t *testing.T) {
	cases := []struct {
		funcName string
		wantProp string
	}{
		{"heikinashi", "heikinashi"},
		{"renko", "renko"},
		{"kagi", "kagi"},
		{"linebreak", "linebreak"},
		{"pointfigure", "pointfigure"},
	}

	transformer := NewTickerNamespaceTransformer()

	for _, tc := range cases {
		t.Run(tc.funcName, func(t *testing.T) {
			input := fmt.Sprintf("haT = %s(sym)", tc.funcName)
			result, err := transformer.Transform(parseScript(t, input))
			if err != nil {
				t.Fatalf("transform failed: %v", err)
			}
			assertMemberAccessCallee(t, extractFirstCallFromAssignment(t, result), "ticker", tc.wantProp)
		})
	}
}

func TestTickerNamespaceTransformer_HeikenashiTypoAlias(t *testing.T) {
	result, err := NewTickerNamespaceTransformer().Transform(parseScript(t, `haT = heikenashi(sym)`))
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}
	assertMemberAccessCallee(t, extractFirstCallFromAssignment(t, result), "ticker", "heikinashi")
}

func TestTickerNamespaceTransformer_RecursesIntoCallArguments(t *testing.T) {
	input := `src = security(heikinashi(syminfo.tickerid), timeframe.period, close)`
	result, err := NewTickerNamespaceTransformer().Transform(parseScript(t, input))
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}

	outerCall := extractFirstCallFromAssignment(t, result)
	if outerCall.Callee.Ident == nil || *outerCall.Callee.Ident != "security" {
		t.Errorf("outer security must not be rewritten by TickerNamespaceTransformer alone, got: %v", outerCall.Callee)
	}

	innerCall := findCallInFactor(outerCall.Args[0].Value.Ternary.Condition.Left.Left.Left.Left.Left)
	assertMemberAccessCallee(t, innerCall, "ticker", "heikinashi")
}

func TestTickerNamespaceTransformer_RenamesInTernaryTrueBranch(t *testing.T) {
	input := `_ticker = useHA ? heikenashi(tickerid) : tickerid`
	result, err := NewTickerNamespaceTransformer().Transform(parseScript(t, input))
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}
	assertMemberAccessCallee(t, extractCallFromTernaryTrueBranch(t, result), "ticker", "heikinashi")
}

func TestV4ToV5Pipeline_NormalizesTickerFunctions(t *testing.T) {
	result, err := NewV4ToV5Pipeline().Run(parseScript(t, `haT = heikinashi(syminfo.tickerid)`))
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	assertMemberAccessCallee(t, extractFirstCallFromAssignment(t, result), "ticker", "heikinashi")
}

func TestV4ToV5Pipeline_NormalizesNestedTickerAndRequestFunctions(t *testing.T) {
	input := `src = security(heikenashi(syminfo.tickerid), timeframe.period, close)`
	result, err := NewV4ToV5Pipeline().Run(parseScript(t, input))
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}

	outerCall := extractFirstCallFromAssignment(t, result)
	assertMemberAccessCallee(t, outerCall, "request", "security")

	innerCall := findCallInFactor(outerCall.Args[0].Value.Ternary.Condition.Left.Left.Left.Left.Left)
	assertMemberAccessCallee(t, innerCall, "ticker", "heikinashi")
}
