package codegen

import (
	"testing"

	"github.com/borisquantlab/pinescript-go/ast"
)

func TestGenerateStrategyEntry(t *testing.T) {
	// Create AST with strategy.entry call
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "entry"},
					},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "long"},
						&ast.MemberExpression{
							Object:   &ast.Identifier{Name: "strategy"},
							Property: &ast.Identifier{Name: "long"},
						},
						&ast.Literal{Value: 1.0},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST failed: %v", err)
	}

	// Verify strategy.Entry call
	if !contains(code.FunctionBody, "strat.Entry") {
		t.Error("Missing strategy.Entry call")
	}
	if !contains(code.FunctionBody, "strategy.Long") {
		t.Error("Missing strategy.Long constant")
	}
}

func TestGenerateStrategyClose(t *testing.T) {
	// Create AST with strategy.close call
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "close"},
					},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "long"},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST failed: %v", err)
	}

	// Verify strategy.Close call
	if !contains(code.FunctionBody, "strat.Close") {
		t.Error("Missing strategy.Close call")
	}
}
