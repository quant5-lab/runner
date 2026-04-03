package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestOperatorAssociativity_Multiplication(t *testing.T) {
	tests := []struct {
		name           string
		expression     string
		expectedTree   string
		expectedResult float64
		description    string
	}{
		{
			name:           "two operands division",
			expression:     "10 / 2",
			expectedTree:   "(10 / 2)",
			expectedResult: 5.0,
			description:    "single binary operation baseline",
		},
		{
			name:           "division then multiplication",
			expression:     "10 / 2 * 5",
			expectedTree:   "((10 / 2) * 5)",
			expectedResult: 25.0,
			description:    "mixed operators must evaluate left-to-right",
		},
		{
			name:           "multiplication then division",
			expression:     "10 * 2 / 5",
			expectedTree:   "((10 * 2) / 5)",
			expectedResult: 4.0,
			description:    "reverse order of mixed operators",
		},
		{
			name:           "three divisions",
			expression:     "100 / 10 / 2",
			expectedTree:   "((100 / 10) / 2)",
			expectedResult: 5.0,
			description:    "homogeneous division chain",
		},
		{
			name:           "three multiplications",
			expression:     "2 * 3 * 4",
			expectedTree:   "((2 * 3) * 4)",
			expectedResult: 24.0,
			description:    "homogeneous multiplication chain",
		},
		{
			name:           "modulo then multiplication",
			expression:     "10 % 3 * 2",
			expectedTree:   "((10 % 3) * 2)",
			expectedResult: 2.0,
			description:    "modulo operator associativity",
		},
		{
			name:           "division then modulo",
			expression:     "10 / 2 % 3",
			expectedTree:   "((10 / 2) % 3)",
			expectedResult: 2.0,
			description:    "division before modulo",
		},
		{
			name:           "four operand chain",
			expression:     "100 / 10 * 2 / 4",
			expectedTree:   "(((100 / 10) * 2) / 4)",
			expectedResult: 5.0,
			description:    "long chain alternating operators",
		},
		{
			name:           "five operand chain",
			expression:     "100 * 2 / 4 * 5 / 10",
			expectedTree:   "((((100 * 2) / 4) * 5) / 10)",
			expectedResult: 25.0,
			description:    "maximum operator chain depth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "//@version=5\nindicator(\"Test\")\nresult = " + tt.expression + "\nplot(result)"

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

			varDecl := findVariableDeclaration(program, "result")
			if varDecl == nil {
				t.Fatalf("Variable 'result' not found")
			}

			init := varDecl.Declarations[0].Init
			if binExpr, ok := init.(*ast.BinaryExpression); ok {
				treeStr := binaryExpressionToString(binExpr)
				if treeStr != tt.expectedTree {
					t.Errorf("%s\nTree structure mismatch:\ngot:  %s\nwant: %s",
						tt.description, treeStr, tt.expectedTree)
				}
			} else if _, ok := init.(*ast.Literal); ok {
				t.Errorf("Unexpected literal for expression with operators: %s", tt.expression)
			} else {
				t.Errorf("Unexpected expression type: %T", init)
			}
		})
	}
}

func TestOperatorAssociativity_Addition(t *testing.T) {
	tests := []struct {
		name           string
		expression     string
		expectedTree   string
		expectedResult float64
		description    string
	}{
		{
			name:           "two operands subtraction",
			expression:     "10 - 2",
			expectedTree:   "(10 - 2)",
			expectedResult: 8.0,
			description:    "single binary operation baseline",
		},
		{
			name:           "subtraction then addition",
			expression:     "10 - 2 + 5",
			expectedTree:   "((10 - 2) + 5)",
			expectedResult: 13.0,
			description:    "mixed operators must evaluate left-to-right",
		},
		{
			name:           "addition then subtraction",
			expression:     "10 + 2 - 5",
			expectedTree:   "((10 + 2) - 5)",
			expectedResult: 7.0,
			description:    "reverse order of mixed operators",
		},
		{
			name:           "three subtractions",
			expression:     "100 - 10 - 5",
			expectedTree:   "((100 - 10) - 5)",
			expectedResult: 85.0,
			description:    "homogeneous subtraction chain - critical for non-commutative ops",
		},
		{
			name:           "three additions",
			expression:     "10 + 20 + 30",
			expectedTree:   "((10 + 20) + 30)",
			expectedResult: 60.0,
			description:    "homogeneous addition chain",
		},
		{
			name:           "four operand chain",
			expression:     "100 - 10 + 20 - 5",
			expectedTree:   "(((100 - 10) + 20) - 5)",
			expectedResult: 105.0,
			description:    "long chain alternating operators",
		},
		{
			name:           "five operand chain",
			expression:     "100 + 50 - 20 + 10 - 5",
			expectedTree:   "((((100 + 50) - 20) + 10) - 5)",
			expectedResult: 135.0,
			description:    "maximum operator chain depth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "//@version=5\nindicator(\"Test\")\nresult = " + tt.expression + "\nplot(result)"

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

			varDecl := findVariableDeclaration(program, "result")
			if varDecl == nil {
				t.Fatalf("Variable 'result' not found")
			}

			init := varDecl.Declarations[0].Init
			if binExpr, ok := init.(*ast.BinaryExpression); ok {
				treeStr := binaryExpressionToString(binExpr)
				if treeStr != tt.expectedTree {
					t.Errorf("%s\nTree structure mismatch:\ngot:  %s\nwant: %s",
						tt.description, treeStr, tt.expectedTree)
				}
			} else if _, ok := init.(*ast.Literal); ok {
				t.Errorf("Unexpected literal for expression with operators: %s", tt.expression)
			} else {
				t.Errorf("Unexpected expression type: %T", init)
			}
		})
	}
}

func TestOperatorAssociativity_MixedPrecedence(t *testing.T) {
	tests := []struct {
		name         string
		expression   string
		expectedTree string
		description  string
	}{
		{
			name:         "multiplication before addition",
			expression:   "2 + 3 * 4",
			expectedTree: "(2 + (3 * 4))",
			description:  "multiplicative operators have higher precedence than additive",
		},
		{
			name:         "division before subtraction",
			expression:   "10 - 6 / 2",
			expectedTree: "(10 - (6 / 2))",
			description:  "division executes before subtraction",
		},
		{
			name:         "left associativity with precedence",
			expression:   "10 + 20 * 2 + 30",
			expectedTree: "((10 + (20 * 2)) + 30)",
			description:  "multiplication has precedence but addition is left-associative",
		},
		{
			name:         "chained operations with precedence",
			expression:   "100 / 10 + 5 * 2 - 3",
			expectedTree: "(((100 / 10) + (5 * 2)) - 3)",
			description:  "mixed precedence with left-to-right within same level",
		},
		{
			name:         "modulo with addition",
			expression:   "10 + 7 % 3",
			expectedTree: "(10 + (7 % 3))",
			description:  "modulo has same precedence as multiplication",
		},
		{
			name:         "complex precedence chain",
			expression:   "2 * 3 + 4 * 5 - 6 / 2",
			expectedTree: "(((2 * 3) + (4 * 5)) - (6 / 2))",
			description:  "multiple precedence levels with left-to-right associativity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "//@version=5\nindicator(\"Test\")\nresult = " + tt.expression + "\nplot(result)"

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

			varDecl := findVariableDeclaration(program, "result")
			if varDecl == nil {
				t.Fatalf("Variable 'result' not found")
			}

			binExpr := varDecl.Declarations[0].Init.(*ast.BinaryExpression)
			treeStr := binaryExpressionToString(binExpr)

			if treeStr != tt.expectedTree {
				t.Errorf("%s\nTree structure mismatch:\ngot:  %s\nwant: %s",
					tt.description, treeStr, tt.expectedTree)
			}
		})
	}
}

func TestOperatorAssociativity_Parentheses(t *testing.T) {
	tests := []struct {
		name         string
		expression   string
		expectedTree string
		description  string
	}{
		{
			name:         "force right association with parens",
			expression:   "10 / (2 * 5)",
			expectedTree: "(10 / (2 * 5))",
			description:  "parentheses override left-to-right associativity",
		},
		{
			name:         "nested parentheses",
			expression:   "100 / (10 / (2 * 5))",
			expectedTree: "(100 / (10 / (2 * 5)))",
			description:  "multiple levels of explicit grouping",
		},
		{
			name:         "left parens redundant",
			expression:   "(10 / 2) * 5",
			expectedTree: "((10 / 2) * 5)",
			description:  "parentheses matching natural left-to-right order",
		},
		{
			name:         "override precedence with parens",
			expression:   "(2 + 3) * 4",
			expectedTree: "((2 + 3) * 4)",
			description:  "force addition before multiplication",
		},
		{
			name:         "complex nested grouping",
			expression:   "((10 - 2) * (3 + 4)) / 7",
			expectedTree: "(((10 - 2) * (3 + 4)) / 7)",
			description:  "multiple grouped sub-expressions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "//@version=5\nindicator(\"Test\")\nresult = " + tt.expression + "\nplot(result)"

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

			varDecl := findVariableDeclaration(program, "result")
			if varDecl == nil {
				t.Fatalf("Variable 'result' not found")
			}

			binExpr := varDecl.Declarations[0].Init.(*ast.BinaryExpression)
			treeStr := binaryExpressionToString(binExpr)

			if treeStr != tt.expectedTree {
				t.Errorf("%s\nTree structure mismatch:\ngot:  %s\nwant: %s",
					tt.description, treeStr, tt.expectedTree)
			}
		})
	}
}

func TestOperatorAssociativity_WithVariables(t *testing.T) {
	tests := []struct {
		name         string
		script       string
		varName      string
		expectedTree string
		description  string
	}{
		{
			name: "variables in division chain",
			script: `//@version=5
indicator("Test")
a = 100.0
b = 10.0
c = 2.0
result = a / b / c`,
			varName:      "result",
			expectedTree: "((a / b) / c)",
			description:  "identifiers follow same left-to-right rules",
		},
		{
			name: "mixed literals and variables",
			script: `//@version=5
indicator("Test")
multiplier = 5.0
result = 10 / 2 * multiplier`,
			varName:      "result",
			expectedTree: "((10 / 2) * multiplier)",
			description:  "literals and identifiers mixed in chain",
		},
		{
			name: "builtin variables",
			script: `//@version=5
indicator("Test")
result = close + open - high`,
			varName:      "result",
			expectedTree: "((close + open) - high)",
			description:  "builtin series variables follow associativity rules",
		},
		{
			name: "complex expression with builtins",
			script: `//@version=5
indicator("Test")
result = (close - open) / open * 100`,
			varName:      "result",
			expectedTree: "(((close - open) / open) * 100)",
			description:  "realistic momentum calculation pattern",
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
				t.Fatalf("Parse error: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion error: %v", err)
			}

			varDecl := findVariableDeclaration(program, tt.varName)
			if varDecl == nil {
				t.Fatalf("Variable '%s' not found", tt.varName)
			}

			binExpr := varDecl.Declarations[0].Init.(*ast.BinaryExpression)
			treeStr := binaryExpressionToString(binExpr)

			if treeStr != tt.expectedTree {
				t.Errorf("%s\nTree structure mismatch:\ngot:  %s\nwant: %s",
					tt.description, treeStr, tt.expectedTree)
			}
		})
	}
}

func TestOperatorAssociativity_SingleOperand(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		expectValue interface{}
		description string
	}{
		{
			name:        "single integer literal",
			expression:  "42",
			expectValue: 42.0,
			description: "no operators - literal passthrough",
		},
		{
			name:        "single float literal",
			expression:  "3.14",
			expectValue: 3.14,
			description: "float literal without operators",
		},
		{
			name:        "single identifier",
			expression:  "close",
			expectValue: nil, /* Identifier, not Literal */
			description: "identifier without operators",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "//@version=5\nindicator(\"Test\")\nresult = " + tt.expression + "\nplot(result)"

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

			varDecl := findVariableDeclaration(program, "result")
			if varDecl == nil {
				t.Fatalf("Variable 'result' not found")
			}

			init := varDecl.Declarations[0].Init

			if tt.expectValue != nil {
				/* Expect Literal */
				lit, ok := init.(*ast.Literal)
				if !ok {
					t.Fatalf("%s: Expected literal, got %T", tt.description, init)
				}
				if lit.Value != tt.expectValue {
					t.Errorf("Expected %v, got %v", tt.expectValue, lit.Value)
				}
			} else {
				/* Expect Identifier */
				id, ok := init.(*ast.Identifier)
				if !ok {
					t.Fatalf("%s: Expected identifier, got %T", tt.description, init)
				}
				if id.Name != tt.expression {
					t.Errorf("Expected identifier '%s', got '%s'", tt.expression, id.Name)
				}
			}
		})
	}
}
