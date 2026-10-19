package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

func pineToStrategyCode(t *testing.T, src string) *StrategyCode {
	t.Helper()
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("parser init: %v", err)
	}
	script, err := p.ParseBytes("test.pine", []byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	conv := parser.NewConverter()
	program, err := conv.ToESTree(script)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	return code
}

// TestIIFECondition_BoolConversion verifies that the IIFE condition path applies
// addBoolConversionIfNeeded consistently with all other if-condition sites.
//
// Pine if-expression: `result = if <cond>` emits Go `(func() float64 { if <goCondition> { ... } }())`.
// The <goCondition> must be a Go bool; series float64 values must be wrapped via value.IsTrue().
func TestIIFECondition_BoolConversion(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		mustHave []string
		mustNot  []string
	}{
		{
			name: "series variable as IIFE condition gets IsTrue conversion",
			src: `
x = 0.0
y = if x
    1.0
else
    0.0
`,
			mustHave: []string{"if value.IsTrue("},
			mustNot:  []string{"if xSeries.GetCurrent() {"},
		},
		{
			name: "comparison expression as IIFE condition passes through as Go bool",
			src: `
a = 1.0
b = if a > 0
    1.0
else
    0.0
`,
			mustHave: []string{"if ("},
			mustNot:  []string{"!= 0"},
		},
		{
			name: "logical expression as IIFE condition passes through as Go bool",
			src: `
p = 0.0
q = 0.0
r = if p > 0 and q > 0
    1.0
else
    0.0
`,
			mustHave: []string{"if ("},
		},
		{
			name: "else-if chain all branches get IsTrue conversion for series condition",
			src: `
flag = 0.0
result = if flag
    1.0
else if flag
    2.0
else
    3.0
`,
			// Both the initial if and the else-if branch must convert series float64 → IsTrue
			mustHave: []string{"if value.IsTrue("},
		},
		{
			name: "unary not expression as IIFE condition passes through as Go bool",
			src: `
cond = 0.0
r = if not cond > 0
    1.0
else
    0.0
`,
			// Unary not produces a Go bool; the condition does not need IsTrue wrapping
			mustHave: []string{"if "},
			mustNot:  []string{"value.IsTrue(not"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := pineToStrategyCode(t, tt.src)
			body := code.FunctionBody
			for _, want := range tt.mustHave {
				if !strings.Contains(body, want) {
					t.Errorf("missing %q in generated body:\n%s", want, body)
				}
			}
			for _, notWant := range tt.mustNot {
				if strings.Contains(body, notWant) {
					t.Errorf("unexpected %q in generated body:\n%s", notWant, body)
				}
			}
		})
	}
}

// TestIIFEBody_LocalVariableScoping verifies that variables first declared inside an
// IIFE body are emitted as Go-local `:=` declarations rather than series registrations.
//
// A variable is "IIFE-local" when it did not exist in the outer scope before the IIFE
// begins. Three code-generation paths must all resolve the identifier as plain Go (not
// `varNameSeries.GetCurrent()`):
//  1. generateArrowFunctionExpression – direct return of an identifier
//  2. extractSeriesExpression – identifier appears in binary arithmetic
//  3. resolveUserIdentifierAccess – identifier used in series expression context
func TestIIFEBody_LocalVariableScoping(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		mustHave []string
		mustNot  []string
	}{
		{
			name: "new variable in IIFE body becomes Go local (direct return path)",
			src: `
result = if close > open
    tmp = close - open
    tmp
else
    0.0
`,
			mustHave: []string{"tmp :="},
			mustNot:  []string{"tmpSeries.Set(", "tmpSeries.GetCurrent()"},
		},
		{
			name: "IIFE local used in binary arithmetic resolves to plain identifier",
			src: `
result = if close > open
    base = close * 2.0
    base + 1.0
else
    0.0
`,
			// base must appear as a plain Go identifier in the return expression, not as a series
			mustHave: []string{"base :=", "(base +"},
			mustNot:  []string{"baseSeries"},
		},
		{
			name: "outer series variable reassigned inside IIFE body remains series assignment",
			src: `
x = 0.0
y = if close > 0
    x := close + 1.0
    x
else
    0.0
`,
			// x existed before the IIFE → must still go through series
			mustHave: []string{"xSeries.Set("},
		},
		{
			name: "two sequential IIFEs with same local name produce independent Go locals",
			src: `
a = if close > open
    tmp = close - open
    tmp
else
    0.0
b = if high > close
    tmp = high - close
    tmp
else
    0.0
`,
			// Both IIFEs emit tmp := independently; no tmpSeries should exist anywhere
			mustHave: []string{"tmp :="},
			mustNot:  []string{"tmpSeries"},
		},
		{
			name: "IIFE local used in math function argument resolves to plain identifier",
			src: `
result = if close > open
    val = close - open
    math.Abs(val)
else
    0.0
`,
			mustHave: []string{"val :="},
			mustNot:  []string{"valSeries"},
		},
		{
			name: "IIFE local used across multiple sub-expressions in same return",
			src: `
result = if close > open
    mid = (close + open) / 2
    high - mid
else
    0.0
`,
			mustHave: []string{"mid :="},
			mustNot:  []string{"midSeries"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := pineToStrategyCode(t, tt.src)
			body := code.FunctionBody
			for _, want := range tt.mustHave {
				if !strings.Contains(body, want) {
					t.Errorf("missing %q in generated body:\n%s", want, body)
				}
			}
			for _, notWant := range tt.mustNot {
				if strings.Contains(body, notWant) {
					t.Errorf("unexpected %q in generated body:\n%s", notWant, body)
				}
			}
		})
	}
}

// TestIIFEScope_OuterSymbolsUnaffected verifies that introducing a variable in an IIFE body
// does not pollute the outer scope: outer-scoped variables must still use series access
// after the IIFE completes.
func TestIIFEScope_OuterSymbolsUnaffected(t *testing.T) {
	src := `
price = close
shifted = if close > open
    local = close - open
    local + 1.0
else
    0.0
`
	code := pineToStrategyCode(t, src)
	body := code.FunctionBody

	// price is declared in outer scope → must be accessed as series everywhere
	if !strings.Contains(body, "priceSeries") {
		t.Errorf("expected outer variable 'price' to use series access, body:\n%s", body)
	}

	// 'local' is IIFE-scoped → must never appear as a series
	if strings.Contains(body, "localSeries") {
		t.Errorf("IIFE-local variable 'local' must not generate series, body:\n%s", body)
	}
}
