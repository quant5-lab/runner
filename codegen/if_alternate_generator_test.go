package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestIfAlternateGenerator_EmptyAlternate(t *testing.T) {
	g := newTestGenerator()

	code, err := g.generateIfAlternate([]ast.Node{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "}") {
		t.Fatalf("expected closing brace, got: %s", code)
	}

	if strings.Contains(code, "else") {
		t.Fatalf("empty alternate should not contain else, got: %s", code)
	}
}

func TestIfAlternateGenerator_ElseIfChain(t *testing.T) {
	g := newTestGenerator()

	nested := &ast.IfStatement{
		NodeType: ast.TypeIfStatement,
		Test: &ast.BinaryExpression{
			NodeType: ast.TypeBinaryExpression,
			Operator: "==",
			Left:     &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "x"},
			Right:    &ast.Literal{NodeType: ast.TypeLiteral, Value: 2.0, Raw: "2"},
		},
		Consequent: []ast.Node{
			&ast.ExpressionStatement{
				NodeType: ast.TypeExpressionStatement,
				Expression: &ast.CallExpression{
					NodeType: ast.TypeCallExpression,
					Callee:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "doB"},
				},
			},
		},
		Alternate: []ast.Node{},
	}

	code, err := g.generateIfAlternate([]ast.Node{nested})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "} else if") {
		t.Fatalf("expected else-if chain, got: %s", code)
	}
}

func TestIfAlternateGenerator_ElseBlock(t *testing.T) {
	g := newTestGenerator()

	alternate := []ast.Node{
		&ast.ExpressionStatement{
			NodeType: ast.TypeExpressionStatement,
			Expression: &ast.CallExpression{
				NodeType: ast.TypeCallExpression,
				Callee:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "fallback"},
			},
		},
	}

	code, err := g.generateIfAlternate(alternate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "} else {") {
		t.Fatalf("expected else block, got: %s", code)
	}
}

func TestIfAlternateGenerator_DeepChain(t *testing.T) {
	g := newTestGenerator()

	innermost := &ast.IfStatement{
		NodeType: ast.TypeIfStatement,
		Test:     &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "c"},
		Consequent: []ast.Node{
			&ast.ExpressionStatement{
				NodeType:   ast.TypeExpressionStatement,
				Expression: &ast.CallExpression{NodeType: ast.TypeCallExpression, Callee: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "doC"}},
			},
		},
		Alternate: []ast.Node{
			&ast.ExpressionStatement{
				NodeType:   ast.TypeExpressionStatement,
				Expression: &ast.CallExpression{NodeType: ast.TypeCallExpression, Callee: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "doDefault"}},
			},
		},
	}

	middle := &ast.IfStatement{
		NodeType: ast.TypeIfStatement,
		Test:     &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "b"},
		Consequent: []ast.Node{
			&ast.ExpressionStatement{
				NodeType:   ast.TypeExpressionStatement,
				Expression: &ast.CallExpression{NodeType: ast.TypeCallExpression, Callee: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "doB"}},
			},
		},
		Alternate: []ast.Node{innermost},
	}

	root := &ast.IfStatement{
		NodeType: ast.TypeIfStatement,
		Test:     &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "a"},
		Consequent: []ast.Node{
			&ast.ExpressionStatement{
				NodeType:   ast.TypeExpressionStatement,
				Expression: &ast.CallExpression{NodeType: ast.TypeCallExpression, Callee: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "doA"}},
			},
		},
		Alternate: []ast.Node{middle},
	}

	code, err := g.generateIfStatement(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "} else if") {
		t.Fatalf("expected else-if chains in output, got: %s", code)
	}

	if !strings.Contains(code, "} else {") {
		t.Fatalf("expected final else block, got: %s", code)
	}

	elseIfCount := strings.Count(code, "else if")
	if elseIfCount != 2 {
		t.Fatalf("expected 2 else-if clauses (3-case chain), got %d in:\n%s", elseIfCount, code)
	}
}

func TestShouldSkipIfBodyStatement(t *testing.T) {
	tests := []struct {
		name   string
		node   ast.Node
		expect bool
	}{
		{
			name: "call expression passes through",
			node: &ast.ExpressionStatement{
				NodeType:   ast.TypeExpressionStatement,
				Expression: &ast.CallExpression{NodeType: ast.TypeCallExpression, Callee: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "f"}},
			},
			expect: false,
		},
		{
			name: "identifier is skipped",
			node: &ast.ExpressionStatement{
				NodeType:   ast.TypeExpressionStatement,
				Expression: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "x"},
			},
			expect: true,
		},
		{
			name: "literal is skipped",
			node: &ast.ExpressionStatement{
				NodeType:   ast.TypeExpressionStatement,
				Expression: &ast.Literal{NodeType: ast.TypeLiteral, Value: 1.0, Raw: "1"},
			},
			expect: true,
		},
		{
			name: "binary expression is skipped",
			node: &ast.ExpressionStatement{
				NodeType: ast.TypeExpressionStatement,
				Expression: &ast.BinaryExpression{
					NodeType: ast.TypeBinaryExpression,
					Operator: ">",
					Left:     &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "a"},
					Right:    &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "b"},
				},
			},
			expect: true,
		},
		{
			name:   "variable declaration passes through",
			node:   &ast.VariableDeclaration{NodeType: ast.TypeVariableDeclaration, Kind: "var"},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSkipIfBodyStatement(tt.node)
			if result != tt.expect {
				t.Errorf("expected %v, got %v", tt.expect, result)
			}
		})
	}
}
