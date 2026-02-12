package codegen

import (
	"strings"
	"testing"
)

/*
TestVariableScopingCodegen validates PineScript variable scoping semantics in generated Go code.

	PineScript rules:
	- Variables declared inside if/else blocks are promoted to function scope (Series storage)
	- Variables declared inside loops (for/for-in/while) remain loop-local (scalar)
	- var-keyword declarations enable cross-bar persistence
*/
func TestVariableScopingCodegen(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "if-block variable promoted to Series",
			pine: `
//@version=5
strategy("Test")
if close > open
    x = close + 1
`,
			mustContainAll: []string{"xSeries"},
			description:    "variable in if-body gets Series declaration at function scope",
		},
		{
			name: "else-block variable promoted to Series",
			pine: `
//@version=5
strategy("Test")
if close > open
    x = 1
else
    y = 2
`,
			mustContainAll: []string{"xSeries", "ySeries"},
			description:    "variables in both if and else branches promoted to function scope",
		},
		{
			name: "if-else-if chain all promoted",
			pine: `
//@version=5
strategy("Test")
if close > open
    a = close
else if close < open
    b = open
else
    c = high
`,
			mustContainAll: []string{"aSeries", "bSeries", "cSeries"},
			description:    "all branches of if-else-if chain get Series declarations",
		},
		{
			name: "nested if-blocks promoted",
			pine: `
//@version=5
strategy("Test")
if close > open
    if high > low
        z = close
`,
			mustContainAll: []string{"zSeries"},
			description:    "nested if-block variables promoted through recursion",
		},
		{
			name: "multiple variables in same if-block",
			pine: `
//@version=5
strategy("Test")
if close > open
    m = close * 2
    n = open / 2
`,
			mustContainAll: []string{"mSeries", "nSeries"},
			description:    "all declarations within same if-block promoted",
		},
		{
			name: "while-loop variable stays loop-local",
			pine: `
//@version=5
strategy("Test")
i = 0
while i < 10
    wVar = i + 1
    i := i + 1
`,
			forbiddenPattern: []string{"wVarSeries"},
			description:      "while-loop body variables remain loop-local scalars",
		},
		{
			name: "loop inside if does not promote loop variable",
			pine: `
//@version=5
strategy("Test")
if close > open
    for i = 0 to 5
        temp = i * 2
`,
			forbiddenPattern: []string{"tempSeries"},
			description:      "loop body inside if-block stays loop-local",
		},
		{
			name: "if-variable promoted alongside sibling loop exclusion",
			pine: `
//@version=5
strategy("Test")
if close > open
    promoted = close
    for i = 0 to 5
        excluded = i
`,
			mustContainAll:   []string{"promotedSeries"},
			forbiddenPattern: []string{"excludedSeries"},
			description:      "if-body variable promoted while sibling loop variable excluded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goCode, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("compilation failed: %v\n%s", err, tt.description)
			}

			for _, pattern := range tt.mustContainAll {
				if !strings.Contains(goCode, pattern) {
					t.Errorf("missing required pattern: %q\ndescription: %s\ngenerated code:\n%s",
						pattern, tt.description, goCode)
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(goCode, forbidden) {
					t.Errorf("found forbidden pattern: %q\ndescription: %s\ngenerated code:\n%s",
						forbidden, tt.description, goCode)
				}
			}
		})
	}
}

/*
TestUnaryVariableInitCodegen validates expression-level code generation for unary variable assignments.

	The generator must produce clean inline expressions (no embedded newlines or statement decorations)
	when initializing variables with unary operators.
*/
func TestUnaryVariableInitCodegen(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "negation of literal",
			pine: `
//@version=5
strategy("Test")
x = -1
`,
			mustContainAll:   []string{"xSeries.Set(-(1))"},
			forbiddenPattern: []string{"-(\\n", "-(\n"},
			description:      "unary negation of literal produces clean inline expression",
		},
		{
			name: "negation of identifier",
			pine: `
//@version=5
strategy("Test")
x = -close
`,
			mustContainAll:   []string{"xSeries.Set(-("},
			forbiddenPattern: []string{"-(\\n", "-(\n"},
			description:      "unary negation of identifier produces clean inline expression",
		},
		{
			name: "negation of parenthesized binary",
			pine: `
//@version=5
strategy("Test")
x = -(close + open)
`,
			mustContainAll:   []string{"xSeries.Set(-("},
			forbiddenPattern: []string{"-(\\n", "-(\n"},
			description:      "unary negation of grouped binary expression produces clean inline",
		},
		{
			name: "positive unary",
			pine: `
//@version=5
strategy("Test")
x = +close
`,
			mustContainAll:   []string{"xSeries.Set(+("},
			forbiddenPattern: []string{"+(\\n", "+(\n"},
			description:      "unary positive operator produces clean inline expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goCode, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("compilation failed: %v\n%s", err, tt.description)
			}

			for _, pattern := range tt.mustContainAll {
				if !strings.Contains(goCode, pattern) {
					t.Errorf("missing required pattern: %q\ndescription: %s\ngenerated code:\n%s",
						pattern, tt.description, goCode)
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(goCode, forbidden) {
					t.Errorf("found forbidden pattern: %q\ndescription: %s\ngenerated code:\n%s",
						forbidden, tt.description, goCode)
				}
			}
		})
	}
}
