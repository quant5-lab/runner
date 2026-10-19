package codegen

import (
	"strings"
	"testing"
)

// TestStringVariable_UDFCallInit validates that string variables initialized by a
// user-defined arrow function returning string are emitted as scalar Go variables
// (not *series.Series) and that the UDF call site is generated correctly.
func TestStringVariable_UDFCallInit(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		mustHave    []string
		mustNotHave []string
		description string
	}{
		{
			name: "simple UDF returning string",
			source: `//@version=5
strategy("t")
label(tf) =>
    tf == "D" ? "Daily" : tf
result = label(timeframe.period)
plot(close)
`,
			mustHave: []string{
				"var result string",
				"func label(",
				"result = label(",
				"_ = result",
			},
			mustNotHave: []string{
				"var resultSeries",
				"resultSeries = series.New",
				"resultSeries.Set(",
				"resultSeries.Next()",
			},
			description: "UDF-call-initialized string variable must use scalar, not Series",
		},
		{
			name: "var-prefixed UDF returning string",
			source: `//@version=5
strategy("t")
fmt(tf) =>
    tf == "1" ? "1min" : tf
var tag = fmt(timeframe.period)
plot(close)
`,
			mustHave: []string{
				"var tag string",
				"func fmt(",
				"tag = fmt(",
				"_ = tag",
			},
			mustNotHave: []string{
				"var tagSeries",
				"tagSeries = series.New",
				"tagSeries.Next()",
			},
			description: "var-prefixed string UDF init also produces scalar",
		},
		{
			name: "UDF returning string used in strategy",
			source: `//@version=5
strategy("t")
directionLabel(side) =>
    side > 0 ? "long" : "short"
dir = directionLabel(close - open)
strategy.entry(dir, strategy.long)
`,
			mustHave: []string{
				"var dir string",
				"func directionLabel(",
				"dir = directionLabel(",
				"_ = dir",
			},
			mustNotHave: []string{
				"var dirSeries",
				"dirSeries.Next()",
			},
			description: "String UDF result used downstream stays scalar",
		},
		{
			name: "multiple string UDF calls coexist with float variables",
			source: `//@version=5
strategy("t")
tfLabel(tf) =>
    tf == "D" ? "daily" : "intraday"
tag = tfLabel(timeframe.period)
smaVal = ta.sma(close, 20)
plot(smaVal)
`,
			mustHave: []string{
				"var tag string",
				"tag = tfLabel(",
				"var smaValSeries",
			},
			mustNotHave: []string{
				"var tagSeries",
				"tagSeries.Set(",
			},
			description: "String UDF variable coexists with float64 Series variable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.source)
			if err != nil {
				t.Fatalf("compilePineScript: %v", err)
			}

			for _, pat := range tt.mustHave {
				if !strings.Contains(code, pat) {
					t.Errorf("%s: missing %q\n--- generated ---\n%s", tt.description, pat, code)
				}
			}
			for _, pat := range tt.mustNotHave {
				if strings.Contains(code, pat) {
					t.Errorf("%s: should NOT contain %q\n--- generated ---\n%s", tt.description, pat, code)
				}
			}
		})
	}
}

// TestStringVariableInit_TypeRescanPromotesUDFVariable verifies that a variable
// initialized by a UDF call whose return type is only known after arrow codegen
// is allocated as a scalar string, not a *series.Series.
func TestStringVariableInit_TypeRescanPromotesUDFVariable(t *testing.T) {
	const source = `//@version=5
strategy("t")
toLabel(x) =>
    x > 0 ? "pos" : "neg"
myVar = toLabel(close)
plot(close)
`
	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilePineScript: %v", err)
	}

	// Scalar declaration must be present.
	if !strings.Contains(code, "var myVar string") {
		t.Errorf("expected scalar string declaration 'var myVar string' in:\n%s", code)
	}
	// Series form must be absent — the rescan must have promoted the type.
	if strings.Contains(code, "myVarSeries") {
		t.Errorf("Series form 'myVarSeries' must not appear after type rescan in:\n%s", code)
	}
	// Call site must be emitted.
	if !strings.Contains(code, "myVar = toLabel(") {
		t.Errorf("expected UDF call 'myVar = toLabel(' in:\n%s", code)
	}
}

// TestStringVariableInit_ExistingCasesUnchanged confirms that the UDF case in
// generateStringVariableInit does not disturb the already-working cases: string
// literals, member expressions, ternary, and na().
func TestStringVariableInit_ExistingCasesUnchanged(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		mustHave    []string
		mustNotHave []string
	}{
		{
			name: "string literal init — scalar declaration",
			source: `//@version=5
strategy("t")
myLabel = "hello"
plot(close)
`,
			// Regardless of whether the init appears in the bar loop, the variable
			// must be declared as a scalar string — never as *series.Series.
			mustHave:    []string{"var myLabel string"},
			mustNotHave: []string{"myLabelSeries"},
		},
		{
			name: "strategy.long member expression",
			source: `//@version=5
strategy("t")
dir = strategy.long
strategy.entry(dir, strategy.long)
`,
			mustHave:    []string{"var dir string", "dir = strategy.Long"},
			mustNotHave: []string{"dirSeries"},
		},
		{
			name: "ternary string init",
			source: `//@version=5
strategy("t")
side = close > open ? strategy.long : strategy.short
strategy.entry(side, strategy.long)
`,
			mustHave:    []string{"var side string", "func() string {"},
			mustNotHave: []string{"sideSeries"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.source)
			if err != nil {
				t.Fatalf("compilePineScript: %v", err)
			}
			for _, pat := range tt.mustHave {
				if !strings.Contains(code, pat) {
					t.Errorf("missing %q\n--- generated ---\n%s", pat, code)
				}
			}
			for _, pat := range tt.mustNotHave {
				if strings.Contains(code, pat) {
					t.Errorf("should NOT contain %q\n--- generated ---\n%s", pat, code)
				}
			}
		})
	}
}
