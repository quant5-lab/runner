package parser

import (
	"encoding/json"
	"testing"
)

func TestParseSimpleIndicator(t *testing.T) {
	input := `//@version=5
indicator("Simple SMA", overlay=true)
sma20 = ta.sma(close, 20)
plot(sma20, color=color.blue, title="SMA20")
`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	script, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(script.Statements) != 3 {
		t.Fatalf("Statements count = %d, want 3", len(script.Statements))
	}
}

func TestConvertToESTree(t *testing.T) {
	input := `//@version=5
indicator("Test", overlay=true)
`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	script, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion error: %v", err)
	}

	if program.NodeType != "Program" {
		t.Errorf("NodeType = %s, want Program", program.NodeType)
	}

	jsonBytes, err := json.Marshal(program)
	if err != nil {
		t.Fatalf("JSON marshal error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}

	if result["type"] != "Program" {
		t.Errorf("JSON type = %s, want Program", result["type"])
	}
}

func TestParseBooleanLiterals(t *testing.T) {
	input := `indicator("Test", overlay=true)`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	script, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion error: %v", err)
	}

	if len(program.Body) == 0 {
		t.Fatal("Empty program body")
	}
}

func TestParseNamedArguments(t *testing.T) {
	input := `plot(close, color=color.blue, title="Test")`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	script, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(script.Statements) != 1 {
		t.Fatalf("Statements count = %d, want 1", len(script.Statements))
	}
}

func TestMultilineExpressionContinuation(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// --- logical operators (all four tokens are distinct in the lexer) ---
		{
			"logical or in if condition",
			`//@version=4
if (a > 0) or
     (b > 1)
    x := 1
`,
		},
		{
			"logical and in if condition",
			`//@version=4
if (a > 0) and
     (b > 1)
    x := 1
`,
		},
		{
			"symbolic || in if condition",
			`//@version=4
if (a > 0) ||
     (b > 1)
    x := 1
`,
		},
		{
			"symbolic && in if condition",
			`//@version=4
if (a > 0) &&
     (b > 1)
    x := 1
`,
		},

		// --- other operator categories (one representative each) ---
		{
			"comparison operator in if condition",
			`//@version=4
if a >=
     0
    x := 1
`,
		},
		{
			"arithmetic operator in if condition",
			`//@version=4
if a +
     b > 0
    x := 1
`,
		},
		{
			"comma in multi-line function call within if condition",
			`//@version=4
if ta.sma(close,
          20) > 0
    x := 1
`,
		},
		{
			"open parenthesis at end of line in if condition",
			`//@version=4
if (
    a > 0)
    x := 1
`,
		},

		// --- chain: multiple consecutive continuation lines ---
		{
			"three consecutive continuation lines",
			`//@version=4
if (a > 0) or
     (b > 1) or
     (c > 2)
    x := 1
`,
		},

		// --- block integrity: INDENT/DEDENT symmetry after continuation ---
		{
			"multi-statement body correctly bounded after continuation",
			`//@version=4
if (a > 0) or
     (b > 1)
    x := 1
    y := 2
    z := 3
`,
		},

		// --- control flow variety ---
		{
			"for loop range expression continuation",
			`//@version=4
for i = 0 to a +
             b
    x := i
`,
		},
		{
			"while loop condition continuation",
			`//@version=4
while (a > 0) or
       (b > 1)
    a := a - 1
`,
		},
		{
			"inline if-expression (IfExpr) in else branch with continuation",
			`//@version=4
result = if condA
    1
else
    if (a > 0) or
         (b > 1)
        2
    else
        3
`,
		},

		// --- regression guards: tokens that must NOT suppress INDENT ---
		{
			"block body after condition ending with closing parenthesis",
			`//@version=4
if (a > 0)
    x := 1
`,
		},
		{
			"block body after condition ending with identifier",
			`//@version=4
if condA
    x := 1
`,
		},
		{
			"function body after => not treated as continuation",
			`//@version=4
f(x) =>
    x + 1
`,
		},

		// --- top-level: different code path (expectingIndent is false) ---
		{
			"top-level assignment with continuation operator",
			`//@version=4
x = (a > 0) or
     (b > 1)
`,
		},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.ParseString("test", tt.input)
			if err != nil {
				t.Errorf("parse error: %v", err)
			}
		})
	}
}

func TestInlineFunctionBody(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// --- expectingIndent not leaked to same-indent statements ---
		{
			"statement at same indent after inline body",
			`//@version=4
f(x) => x + 1
y = 2
`,
		},
		{
			"consecutive inline bodies at same indent",
			`//@version=4
f(x) => x + 1
g(x) => x + 2
y = 1
`,
		},

		// --- expectingIndent not leaked to control-flow blocks ---
		{
			"if block after inline body",
			`//@version=4
f(x) => x + 1
if y > 0
    y := y - 1
`,
		},
		{
			"consecutive inline bodies then if block",
			`//@version=4
f(x) => x + 1
g(x) => x + 2
if y > 0
    y := y - 1
`,
		},

		// --- expectingIndent not leaked to indented continuation expressions ---
		{
			"indented continuation expression after inline body",
			`//@version=4
f(x) => x + 1
y = a and b
 and c
`,
		},

		// --- inline body inside a block does not corrupt outer scope ---
		{
			"inline body inside if block followed by outer statement",
			`//@version=4
if cond
    f(x) => x + 1
y = 2
`,
		},

		// --- regression: block body (multi-line) still works when mixed with inline ---
		{
			"block body and inline body in same script",
			`//@version=4
f(x) =>
    x + 1
g(x) => x + 2
if y > 0
    y := y - 1
`,
		},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.ParseString("test", tt.input)
			if err != nil {
				t.Errorf("parse error: %v", err)
			}
		})
	}
}

func TestParseHexColor8Digit(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"6-digit hex literal", `x = #FF5252`},
		{"8-digit hex literal", `x = #FF525280`},
		{"8-digit in function arg", `plot(close, color=#FF000080)`},
		{"8-digit uppercase", `x = #AABBCCDD`},
		{"8-digit lowercase", `x = #aabbccdd`},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.ParseString("test", tt.input)
			if err != nil {
				t.Errorf("Failed to parse %q: %v", tt.input, err)
			}
		})
	}
}
