package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestUnknownFunctionHandler_CanHandle verifies catch-all behavior
func TestUnknownFunctionHandler_CanHandle(t *testing.T) {
	handler := &UnknownFunctionHandler{}

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		// Should handle everything (catch-all)
		{"unrecognized function", "unknown_func", true},
		{"random name", "foo", true},
		{"with namespace", "custom.bar", true},
		{"empty string", "", true},
		{"special chars", "func@123", true},
		{"very long name", strings.Repeat("a", 100), true},

		// Even recognized functions (when reached)
		{"plot", "plot", true},
		{"ta.sma", "ta.sma", true},
		{"strategy.entry", "strategy.entry", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.CanHandle(tt.funcName)
			if got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

// TestUnknownFunctionHandler_GenerateCode verifies TODO comment generation
func TestUnknownFunctionHandler_GenerateCode(t *testing.T) {
	handler := &UnknownFunctionHandler{}
	g := newTestGenerator()

	tests := []struct {
		name     string
		call     *ast.CallExpression
		wantFunc string
	}{
		{
			name: "simple unknown function",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "custom_func"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "x"},
				},
			},
			wantFunc: "custom_func",
		},
		{
			name: "namespaced unknown function",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "custom"},
					Property: &ast.Identifier{Name: "function"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 42.0},
				},
			},
			wantFunc: "custom.function",
		},
		{
			name: "no arguments",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "mystery"},
				Arguments: []ast.Expression{},
			},
			wantFunc: "mystery",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := handler.GenerateCode(g, tt.call)
			if err != nil {
				t.Errorf("GenerateCode() unexpected error: %v", err)
			}

			// Should generate TODO comment with function name
			// Format: // funcName() - TODO: implement\n
			expectedPattern := fmt.Sprintf("// %s() - TODO: implement", tt.wantFunc)
			if !strings.Contains(code, expectedPattern) {
				t.Errorf("GenerateCode() should contain %q, got: %q", expectedPattern, code)
			}
		})
	}
}

// TestUnknownFunctionHandler_TODOFormat verifies TODO comment format
func TestUnknownFunctionHandler_TODOFormat(t *testing.T) {
	handler := &UnknownFunctionHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "unknown_func"},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "arg1"},
			&ast.Literal{Value: 100.0},
		},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("GenerateCode() error: %v", err)
	}

	// TODO comment format expectations
	if !strings.HasPrefix(strings.TrimSpace(code), "//") {
		t.Errorf("Code should start with comment, got: %q", code)
	}

	if !strings.Contains(code, "unknown_func") {
		t.Error("TODO should mention the function name")
	}
}

// TestUnknownFunctionHandler_NilSafety tests handling of nil arguments
func TestUnknownFunctionHandler_NilSafety(t *testing.T) {
	handler := &UnknownFunctionHandler{}
	g := newTestGenerator()

	tests := []struct {
		name    string
		call    *ast.CallExpression
		wantErr bool
	}{
		{
			name: "nil callee",
			call: &ast.CallExpression{
				Callee:    nil,
				Arguments: []ast.Expression{},
			},
			wantErr: true, // extractCallFunctionName should handle gracefully
		},
		{
			name: "valid call",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "func"},
				Arguments: []ast.Expression{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := handler.GenerateCode(g, tt.call)

			if tt.wantErr {
				if err == nil && code == "" {
					t.Error("Expected error or empty code for nil callee")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestUnknownFunctionHandler_IntegrationWithGenerator verifies full pipeline
func TestUnknownFunctionHandler_IntegrationWithGenerator(t *testing.T) {
	g := newTestGenerator()

	// Ensure router has UnknownFunctionHandler as catch-all
	if g.callRouter == nil {
		g.callRouter = NewCallExpressionRouter()
	}

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "unrecognized_builtin"},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "source"},
			&ast.Literal{Value: 10.0},
		},
	}

	code, err := g.generateCallExpression(call)
	if err != nil {
		t.Fatalf("generateCallExpression() error: %v", err)
	}

	// Should generate TODO comment via unknown handler
	// Format: // unrecognized_builtin() - TODO: implement\n
	expectedPattern := "// unrecognized_builtin() - TODO: implement"
	if !strings.Contains(code, expectedPattern) {
		t.Errorf("Integration should generate TODO with pattern %q, got: %q", expectedPattern, code)
	}
}

// TestUnknownFunctionHandler_LastResortBehavior verifies it doesn't intercept known functions
func TestUnknownFunctionHandler_LastResortBehavior(t *testing.T) {
	g := newTestGenerator()

	// Known function should be handled by specific handler, not unknown
	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "indicator"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "Test"},
		},
	}

	code, err := g.generateCallExpression(call)
	if err != nil {
		t.Fatalf("generateCallExpression() error: %v", err)
	}

	// Meta handler returns empty string, not TODO
	if strings.Contains(code, "// TODO") {
		t.Errorf("Known function should not reach unknown handler, got: %q", code)
	}

	if code != "" {
		t.Errorf("Meta function should return empty string, got: %q", code)
	}
}

// TestUnknownFunctionHandler_VariousFunctionNames tests diverse function name formats
func TestUnknownFunctionHandler_VariousFunctionNames(t *testing.T) {
	handler := &UnknownFunctionHandler{}
	g := newTestGenerator()

	functionNames := []string{
		"my_custom_func",
		"library.helper",
		"util.format.number",
		"_private",
		"CONSTANT_FUNC",
		"func123",
		"123func", // unusual but possible
		"a",       // single char
	}

	for _, funcName := range functionNames {
		t.Run(funcName, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: funcName},
				Arguments: []ast.Expression{},
			}

			code, err := handler.GenerateCode(g, call)
			if err != nil {
				t.Errorf("GenerateCode(%q) error: %v", funcName, err)
			}

			if code == "" {
				t.Errorf("GenerateCode(%q) should return TODO comment, got empty", funcName)
			}

			if !strings.Contains(code, funcName) {
				t.Errorf("TODO should mention function %q, got: %q", funcName, code)
			}
		})
	}
}
