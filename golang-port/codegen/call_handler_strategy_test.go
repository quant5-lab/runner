package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestStrategyActionHandler_CanHandle verifies strategy action recognition
func TestStrategyActionHandler_CanHandle(t *testing.T) {
	handler := &StrategyActionHandler{}

	tests := []struct {
		funcName string
		want     bool
	}{
		{"strategy.entry", true},
		{"strategy.close", true},
		{"strategy.close_all", true},
		{"strategy", false},
		{"strategy.exit", false},
		{"ta.entry", false},
		{"entry", false},
		{"", false},
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

// TestStrategyActionHandler_EntryValidCases verifies correct entry code generation
func TestStrategyActionHandler_EntryValidCases(t *testing.T) {
	handler := &StrategyActionHandler{}
	g := newTestGenerator()

	tests := []struct {
		name         string
		call         *ast.CallExpression
		wantContains []string
	}{
		{
			name: "entry with 2 args",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Buy"},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "long"},
					},
				},
			},
			wantContains: []string{"strat.Entry", `"Buy"`, "strategy.Long"},
		},
		{
			name: "entry with 3 args (quantity)",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Sell"},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "short"},
					},
					&ast.Literal{Value: 2.0},
				},
			},
			wantContains: []string{"strat.Entry", `"Sell"`, "strategy.Short", "2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := handler.GenerateCode(g, tt.call)
			if err != nil {
				t.Errorf("GenerateCode() unexpected error: %v", err)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("GenerateCode() = %q, want to contain %q", code, want)
				}
			}
		})
	}
}

// TestStrategyActionHandler_EntryInvalidArgs verifies graceful handling of invalid entry args
func TestStrategyActionHandler_EntryInvalidArgs(t *testing.T) {
	handler := &StrategyActionHandler{}
	g := newTestGenerator()

	tests := []struct {
		name         string
		call         *ast.CallExpression
		wantContains string
	}{
		{
			name: "no arguments",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{},
			},
			wantContains: "// strategy.entry() - invalid arguments",
		},
		{
			name: "one argument",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Buy"},
				},
			},
			wantContains: "// strategy.entry() - invalid arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := handler.GenerateCode(g, tt.call)
			if err != nil {
				t.Errorf("GenerateCode() unexpected error: %v", err)
			}

			if !strings.Contains(code, tt.wantContains) {
				t.Errorf("GenerateCode() = %q, want to contain %q", code, tt.wantContains)
			}
		})
	}
}

// TestStrategyActionHandler_CloseValidCases verifies correct close code generation
func TestStrategyActionHandler_CloseValidCases(t *testing.T) {
	handler := &StrategyActionHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: "close"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "Buy"},
		},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Errorf("GenerateCode() unexpected error: %v", err)
	}

	wantContains := []string{"strat.Close", `"Buy"`, "bar.Close", "bar.Time"}
	for _, want := range wantContains {
		if !strings.Contains(code, want) {
			t.Errorf("GenerateCode() = %q, want to contain %q", code, want)
		}
	}
}

// TestStrategyActionHandler_CloseInvalidArgs verifies graceful handling of invalid close args
func TestStrategyActionHandler_CloseInvalidArgs(t *testing.T) {
	handler := &StrategyActionHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: "close"},
		},
		Arguments: []ast.Expression{},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Errorf("GenerateCode() unexpected error: %v", err)
	}

	if !strings.Contains(code, "// strategy.close() - invalid arguments") {
		t.Errorf("GenerateCode() = %q, want TODO comment for invalid args", code)
	}
}

// TestStrategyActionHandler_CloseAll verifies close_all code generation
func TestStrategyActionHandler_CloseAll(t *testing.T) {
	handler := &StrategyActionHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: "close_all"},
		},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Errorf("GenerateCode() unexpected error: %v", err)
	}

	wantContains := []string{"strat.CloseAll", "bar.Close", "bar.Time"}
	for _, want := range wantContains {
		if !strings.Contains(code, want) {
			t.Errorf("GenerateCode() = %q, want to contain %q", code, want)
		}
	}
}

// TestStrategyActionHandler_IntegrationWithGenerator tests strategy actions in full pipeline
func TestStrategyActionHandler_IntegrationWithGenerator(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "signal"},
						Init: &ast.Literal{Value: 1.0},
					},
				},
			},
			&ast.IfStatement{
				Test: &ast.Identifier{Name: "signal"},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "strategy"},
								Property: &ast.Identifier{Name: "entry"},
							},
							Arguments: []ast.Expression{
								&ast.Literal{Value: "Long"},
								&ast.MemberExpression{
									Object:   &ast.Identifier{Name: "strategy"},
									Property: &ast.Identifier{Name: "long"},
								},
							},
						},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST() error: %v", err)
	}

	// Should generate strat.Entry call inside if statement
	if !strings.Contains(code.FunctionBody, "strat.Entry") {
		t.Error("Expected strat.Entry call in generated code")
	}

	if !strings.Contains(code.FunctionBody, "if signal") {
		t.Error("Expected if statement in generated code")
	}
}

// TestStrategyActionHandler_EdgeCases tests unusual but valid scenarios
func TestStrategyActionHandler_EdgeCases(t *testing.T) {
	handler := &StrategyActionHandler{}
	g := newTestGenerator()

	tests := []struct {
		name string
		call *ast.CallExpression
	}{
		{
			name: "entry with expression as ID",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{
					&ast.BinaryExpression{
						Left:     &ast.Literal{Value: "Buy"},
						Operator: "+",
						Right:    &ast.Literal{Value: "1"},
					},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "long"},
					},
				},
			},
		},
		{
			name: "close_all with extra arguments (ignored)",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "close_all"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "ignored"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			code, err := handler.GenerateCode(g, tt.call)
			if err != nil {
				t.Errorf("GenerateCode() unexpected error: %v", err)
			}
			_ = code
		})
	}
}
