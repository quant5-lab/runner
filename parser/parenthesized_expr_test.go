package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestParenthesizedExpression_SimpleParentheses(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "numeric literal",
			code: "x = (42)",
		},
		{
			name: "identifier",
			code: "x = (close)",
		},
		{
			name: "arithmetic",
			code: "x = (1 + 2)",
		},
		{
			name: "nested parentheses",
			code: "x = ((1 + 2) * 3)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pineScript := "//@version=5\nindicator(\"Test\")\n" + tt.code + "\n"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(pineScript))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if len(program.Body) < 2 {
				t.Fatalf("Expected at least 2 statements, got %d", len(program.Body))
			}
		})
	}
}

func TestParenthesizedExpression_WithSubscript(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "identifier with subscript",
			code: "x = (close)[1]",
		},
		{
			name: "call with subscript",
			code: "x = (ta.sma(close, 14))[1]",
		},
		{
			name: "arithmetic with subscript",
			code: "x = (close + open)[0]",
		},
		{
			name: "array literal with subscript",
			code: "x = ([1,2,3])[0]",
		},
		{
			name: "nested parentheses with subscript",
			code: "x = ((close))[1]",
		},
		{
			name: "chained subscripts",
			code: "x = ([[1,2],[3,4]])[0][1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pineScript := "//@version=5\nindicator(\"Test\")\n" + tt.code + "\n"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(pineScript))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			varDecl := program.Body[1].(*ast.VariableDeclaration)
			memberExpr, ok := varDecl.Declarations[0].Init.(*ast.MemberExpression)
			if !ok {
				t.Fatalf("Expected MemberExpression for subscript, got %T", varDecl.Declarations[0].Init)
			}

			if !memberExpr.Computed {
				t.Error("Expected computed property (subscript)")
			}
		})
	}
}

func TestParenthesizedExpression_InFunctionArguments(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "simple parenthesized arg",
			code: "result = func((42))",
		},
		{
			name: "parenthesized arithmetic",
			code: "result = func((1 + 2))",
		},
		{
			name: "parenthesized with subscript",
			code: "result = func((close)[1])",
		},
		{
			name: "array with subscript in arg",
			code: "result = func(([1,2,3])[0])",
		},
		{
			name: "multiple parenthesized args",
			code: "result = func((a), (b), (c))",
		},
		{
			name: "mixed parenthesized and non-parenthesized",
			code: "result = func(a, (b + c), d)",
		},
		{
			name: "nested function calls with parentheses",
			code: "result = func((inner(x)))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pineScript := "//@version=5\nindicator(\"Test\")\n" + tt.code + "\n"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(pineScript))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if len(program.Body) < 2 {
				t.Fatalf("Expected at least 2 statements, got %d", len(program.Body))
			}
		})
	}
}

func TestParenthesizedExpression_InTernary(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "parenthesized condition",
			code: "result = (close > open) ? 1 : 0",
		},
		{
			name: "parenthesized true value",
			code: "result = cond ? (a + b) : c",
		},
		{
			name: "parenthesized false value",
			code: "result = cond ? a : (b + c)",
		},
		{
			name: "all parenthesized",
			code: "result = (cond) ? (a) : (b)",
		},
		{
			name: "parenthesized with subscript in ternary",
			code: "result = cond ? ([1,2])[0] : ([3,4])[1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pineScript := "//@version=5\nindicator(\"Test\")\n" + tt.code + "\n"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(pineScript))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if len(program.Body) < 2 {
				t.Fatalf("Expected at least 2 statements, got %d", len(program.Body))
			}
		})
	}
}

func TestParenthesizedExpression_InArithmetic(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "precedence with parentheses",
			code: "x = (1 + 2) * 3",
		},
		{
			name: "nested arithmetic",
			code: "x = ((a + b) * (c + d))",
		},
		{
			name: "parenthesized subscript in arithmetic",
			code: "x = ([1,2,3])[0] + ([4,5,6])[1]",
		},
		{
			name: "complex expression",
			code: "x = (close[1] + open[1]) / 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pineScript := "//@version=5\nindicator(\"Test\")\n" + tt.code + "\n"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(pineScript))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if len(program.Body) < 2 {
				t.Fatalf("Expected at least 2 statements, got %d", len(program.Body))
			}
		})
	}
}

func TestParenthesizedExpression_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		shouldErr bool
	}{
		{
			name:      "empty parentheses",
			code:      "x = ()",
			shouldErr: true,
		},
		{
			name:      "deeply nested",
			code:      "x = ((((42))))",
			shouldErr: false,
		},
		{
			name:      "parentheses in comparison",
			code:      "x = (close) > (open)",
			shouldErr: false,
		},
		{
			name:      "parentheses with member access",
			code:      "x = (strategy.position_size)",
			shouldErr: false,
		},
		{
			name:      "multiple subscripts on parenthesized",
			code:      "x = (arr)[0][1][2]",
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pineScript := "//@version=5\nindicator(\"Test\")\n" + tt.code + "\n"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(pineScript))
			if tt.shouldErr && err == nil {
				t.Fatal("Expected parse error, got none")
			}
			if !tt.shouldErr && err != nil {
				t.Fatalf("Parse failed unexpectedly: %v", err)
			}

			if !tt.shouldErr {
				converter := NewConverter()
				_, err := converter.ToESTree(script)
				if err != nil {
					t.Fatalf("Conversion failed: %v", err)
				}
			}
		})
	}
}
