package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestMetaFunctionHandler_CanHandle verifies meta function recognition
func TestMetaFunctionHandler_CanHandle(t *testing.T) {
	handler := &MetaFunctionHandler{}

	tests := []struct {
		funcName string
		want     bool
	}{
		{"indicator", true},
		{"strategy", true},
		{"plot", false},
		{"ta.sma", false},
		{"strategy.entry", false},
		{"", false},
		{"INDICATOR", false}, // Case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			got := handler.CanHandle(tt.funcName)
			if got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

// TestMetaFunctionHandler_GenerateCode verifies no code generation for meta functions
func TestMetaFunctionHandler_GenerateCode(t *testing.T) {
	handler := &MetaFunctionHandler{}
	g := newTestGenerator()

	tests := []struct {
		name string
		call *ast.CallExpression
	}{
		{
			name: "indicator call",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "indicator"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "My Indicator"},
				},
			},
		},
		{
			name: "strategy call with arguments",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "strategy"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "My Strategy"},
					&ast.ObjectExpression{},
				},
			},
		},
		{
			name: "indicator with no arguments",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "indicator"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := handler.GenerateCode(g, tt.call)
			if err != nil {
				t.Errorf("GenerateCode() unexpected error: %v", err)
			}
			if code != "" {
				t.Errorf("GenerateCode() should return empty string for meta functions, got: %q", code)
			}
		})
	}
}

// TestMetaFunctionHandler_IntegrationWithGenerator tests meta functions don't affect runtime
func TestMetaFunctionHandler_IntegrationWithGenerator(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "strategy"},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "Test Strategy"},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "x"},
						Init: &ast.Literal{Value: 10.0},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST() error: %v", err)
	}

	// Strategy name should be extracted but no runtime code for strategy() call
	if code.StrategyName != "Test Strategy" {
		t.Errorf("Expected strategy name 'Test Strategy', got %q", code.StrategyName)
	}

	// Should have variable declaration code
	if code.FunctionBody == "" {
		t.Error("Expected non-empty function body")
	}
}
