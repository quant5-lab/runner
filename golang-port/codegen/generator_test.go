package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestGenerateStrategyCodeFromAST(t *testing.T) {
	// Create minimal AST
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body:     []ast.Node{},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST failed: %v", err)
	}

	if code == nil {
		t.Fatal("Generated code is nil")
	}

	if len(code.FunctionBody) == 0 {
		t.Error("Function body is empty")
	}

	// Verify placeholder code
	if !contains(code.FunctionBody, "strat.Call") {
		t.Error("Missing strategy initialization")
	}
	if !contains(code.FunctionBody, "for i := 0") {
		t.Error("Missing bar loop")
	}
}

func TestGenerateProgramWithStatements(t *testing.T) {
	// Create AST with indicator call
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "indicator"},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "Test Strategy"},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST failed: %v", err)
	}

	// Verify strategy initialization
	if !contains(code.FunctionBody, "strat.Call") {
		t.Error("Missing strategy call")
	}
}

func TestGeneratorIndentation(t *testing.T) {
	gen := &generator{
		imports:   make(map[string]bool),
		variables: make(map[string]string),
		indent:    0,
	}

	// Test indentation levels
	if gen.ind() != "" {
		t.Error("Indent level 0 should be empty")
	}

	gen.indent = 1
	if gen.ind() != "\t" {
		t.Error("Indent level 1 should be one tab")
	}

	gen.indent = 2
	if gen.ind() != "\t\t" {
		t.Error("Indent level 2 should be two tabs")
	}
}
