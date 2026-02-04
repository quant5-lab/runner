package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Lexer number literal tokenization tests - verifies Float/Int token patterns */

func TestNumberLiteral_BasicFormats(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		expectValue float64
		description string
	}{
		{
			name:        "integer literal",
			expression:  "42",
			expectValue: 42.0,
			description: "integer",
		},
		{
			name:        "standard float",
			expression:  "3.14",
			expectValue: 3.14,
			description: "float",
		},
		{
			name:        "zero",
			expression:  "0",
			expectValue: 0.0,
			description: "zero int",
		},
		{
			name:        "zero float",
			expression:  "0.0",
			expectValue: 0.0,
			description: "zero float",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runLiteralTest(t, tt.expression, tt.expectValue, tt.description)
		})
	}
}

func TestNumberLiteral_ScientificNotation(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		expectValue float64
		description string
	}{
		{
			name:        "integer base lowercase e",
			expression:  "3e8",
			expectValue: 3e8,
			description: "scientific int",
		},
		{
			name:        "integer base uppercase E",
			expression:  "1E10",
			expectValue: 1e10,
			description: "scientific uppercase",
		},
		{
			name:        "decimal base positive exponent",
			expression:  "6.02e23",
			expectValue: 6.02e23,
			description: "scientific float",
		},
		{
			name:        "decimal base negative exponent",
			expression:  "1.6e-19",
			expectValue: 1.6e-19,
			description: "scientific negative exp",
		},
		{
			name:        "uppercase E negative exponent",
			expression:  "2.5E-5",
			expectValue: 2.5e-5,
			description: "scientific uppercase negative",
		},
		{
			name:        "zero exponent",
			expression:  "5e0",
			expectValue: 5.0,
			description: "scientific e0",
		},
		{
			name:        "positive sign explicit",
			expression:  "1e+10",
			expectValue: 1e10,
			description: "scientific e+",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runLiteralTest(t, tt.expression, tt.expectValue, tt.description)
		})
	}
}

func TestNumberLiteral_DecimalPointVariations(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		expectValue float64
		description string
	}{
		{
			name:        "trailing decimal",
			expression:  "1.",
			expectValue: 1.0,
			description: "trailing decimal",
		},
		{
			name:        "trailing decimal large",
			expression:  "100.",
			expectValue: 100.0,
			description: "trailing decimal",
		},
		{
			name:        "leading decimal",
			expression:  ".5",
			expectValue: 0.5,
			description: "leading decimal",
		},
		{
			name:        "leading decimal small",
			expression:  ".123",
			expectValue: 0.123,
			description: "leading decimal",
		},
		{
			name:        "trailing decimal with exponent",
			expression:  "1.e5",
			expectValue: 1e5,
			description: "trailing decimal + scientific",
		},
		{
			name:        "leading decimal with exponent",
			expression:  ".5e2",
			expectValue: 0.5e2,
			description: "leading decimal + scientific",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runLiteralTest(t, tt.expression, tt.expectValue, tt.description)
		})
	}
}

func TestNumberLiteral_InExpressions(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		expectTree  string
		description string
	}{
		{
			name:        "trailing decimal in subtraction",
			expression:  "10. - 5.",
			expectTree:  "(10 - 5)",
			description: "trailing decimal expr",
		},
		{
			name:        "leading decimal in multiplication",
			expression:  ".5 * 2",
			expectTree:  "(0.5 * 2)",
			description: "leading decimal expr",
		},
		{
			name:        "mixed standard formats",
			expression:  "100 + 2.5 - .5",
			expectTree:  "((100 + 2.5) - 0.5)",
			description: "mixed formats expr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "//@version=5\nindicator(\"Test\")\nresult = " + tt.expression

			p, err := NewParser()
			if err != nil {
				t.Fatalf("NewParser() error: %v", err)
			}

			script, err := p.ParseString("test", input)
			if err != nil {
				t.Fatalf("%s: Parse error: %v", tt.description, err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("%s: Conversion error: %v", tt.description, err)
			}

			varDecl := findVariableDeclaration(program, "result")
			if varDecl == nil {
				t.Fatalf("Variable 'result' not found")
			}

			binExpr := varDecl.Declarations[0].Init.(*ast.BinaryExpression)
			treeStr := binaryExpressionToString(binExpr)

			if treeStr != tt.expectTree {
				t.Errorf("%s\nExpression tree mismatch:\ngot:  %s\nwant: %s",
					tt.description, treeStr, tt.expectTree)
			}
		})
	}
}

func TestNumberLiteral_ScientificInExpressions(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		description string
	}{
		{
			name:        "scientific in addition",
			expression:  "1e6 + 5e5",
			description: "scientific binary expr",
		},
		{
			name:        "scientific with standard",
			expression:  "1e3 + 100",
			description: "scientific mixed expr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "//@version=5\nindicator(\"Test\")\nresult = " + tt.expression

			p, err := NewParser()
			if err != nil {
				t.Fatalf("NewParser() error: %v", err)
			}

			script, err := p.ParseString("test", input)
			if err != nil {
				t.Fatalf("%s: Parse error: %v", tt.description, err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("%s: Conversion error: %v", tt.description, err)
			}

			varDecl := findVariableDeclaration(program, "result")
			if varDecl == nil {
				t.Fatalf("Variable 'result' not found")
			}

			/* Must be binary expression */
			if _, ok := varDecl.Declarations[0].Init.(*ast.BinaryExpression); !ok {
				t.Errorf("%s: Expected binary expression, got %T",
					tt.description, varDecl.Declarations[0].Init)
			}
		})
	}
}

func TestNumberLiteral_TokenBoundaries(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		expectStmts int
		description string
	}{
		{
			name: "scientific not split into identifier",
			script: `//@version=5
indicator("Test")
x = 6.02e23
y = 1`,
			expectStmts: 3,
			description: "scientific single token",
		},
		{
			name: "trailing decimal not split",
			script: `//@version=5
indicator("Test")
x = 10.
y = 20`,
			expectStmts: 3,
			description: "trailing decimal single token",
		},
		{
			name: "negative exponent not confused with subtraction",
			script: `//@version=5
indicator("Test")
x = 1.6e-19`,
			expectStmts: 2,
			description: "negative exp single token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("NewParser() error: %v", err)
			}

			script, err := p.ParseString("test", tt.script)
			if err != nil {
				t.Fatalf("%s: Parse error: %v", tt.description, err)
			}

			if len(script.Statements) != tt.expectStmts {
				t.Errorf("%s\nStatement count mismatch: got %d, want %d\n"+
					"This suggests lexer incorrectly tokenized number literal",
					tt.description, len(script.Statements), tt.expectStmts)
			}
		})
	}
}

func TestNumberLiteral_InArrays(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectCount int
		description string
	}{
		{
			name:        "scientific in array",
			code:        "arr = [1e6, 2e6, 3e6]",
			expectCount: 3,
			description: "scientific array",
		},
		{
			name:        "trailing decimals in array",
			code:        "arr = [1., 2., 3.]",
			expectCount: 3,
			description: "trailing decimal array",
		},
		{
			name:        "leading decimals in array",
			code:        "arr = [.1, .2, .3]",
			expectCount: 3,
			description: "leading decimal array",
		},
		{
			name:        "mixed formats in array",
			code:        "arr = [1e3, 2.5, .5, 10.]",
			expectCount: 4,
			description: "mixed formats array",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pineScript := "//@version=5\nindicator(\"Test\")\n" + tt.code

			p, err := NewParser()
			if err != nil {
				t.Fatalf("NewParser() error: %v", err)
			}

			script, err := p.ParseString("test", pineScript)
			if err != nil {
				t.Fatalf("%s: Parse error: %v", tt.description, err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("%s: Conversion error: %v", tt.description, err)
			}

			varDecl := program.Body[1].(*ast.VariableDeclaration)
			literal := varDecl.Declarations[0].Init.(*ast.Literal)
			elements := literal.Value.([]ast.Expression)

			if len(elements) != tt.expectCount {
				t.Errorf("%s\nArray element count: got %d, want %d",
					tt.description, len(elements), tt.expectCount)
			}
		})
	}
}

/* Helper functions */

func runLiteralTest(t *testing.T, expression string, expectValue float64, description string) {
	input := "//@version=5\nindicator(\"Test\")\nresult = " + expression

	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	script, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("%s: Parse error: %v", description, err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("%s: Conversion error: %v", description, err)
	}

	varDecl := findVariableDeclaration(program, "result")
	if varDecl == nil {
		t.Fatalf("Variable 'result' not found")
	}

	lit, ok := varDecl.Declarations[0].Init.(*ast.Literal)
	if !ok {
		t.Fatalf("%s: Expected literal, got %T", description, varDecl.Declarations[0].Init)
	}

	got := lit.Value.(float64)
	if got != expectValue {
		t.Errorf("%s\nValue mismatch: got %v, want %v", description, got, expectValue)
	}
}
