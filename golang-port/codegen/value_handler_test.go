package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestValueHandlerCanHandle(t *testing.T) {
	handler := NewValueHandler()

	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"na function", "na", true},
		{"nz function", "nz", true},
		{"fixnan function", "fixnan", true},
		{"ta.sma function", "sma", false},
		{"close builtin", "close", false},
		{"math.abs function", "math.abs", false},
		{"empty string", "", false},
		{"random string", "xyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.CanHandle(tt.funcName)
			if result != tt.expected {
				t.Errorf("CanHandle(%s) = %v, want %v", tt.funcName, result, tt.expected)
			}
		})
	}
}

func TestValueHandlerGenerateNa(t *testing.T) {
	handler := NewValueHandler()
	gen := &generator{
		variables: make(map[string]string),
		varInits:  make(map[string]ast.Expression),
	}

	tests := []struct {
		name     string
		args     []ast.Expression
		expected string
		wantErr  bool
	}{
		{
			name:     "no arguments returns true",
			args:     []ast.Expression{},
			expected: "true",
		},
		{
			name: "identifier argument",
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
			},
			expected: "math.IsNaN(bar.Close)",
		},
		{
			name: "literal argument",
			args: []ast.Expression{
				&ast.Literal{Value: 42.0},
			},
			expected: "math.IsNaN(42.00)",
		},
		{
			name: "series historical access",
			args: []ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "value"},
					Property: &ast.Literal{Value: 1},
					Computed: true,
				},
			},
			expected: "math.IsNaN(valueSeries.Get(1))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.generateNa(tt.args, gen)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("generateNa() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestValueHandlerGenerateNz(t *testing.T) {
	handler := NewValueHandler()
	gen := &generator{
		variables: make(map[string]string),
		varInits:  make(map[string]ast.Expression),
	}

	tests := []struct {
		name     string
		args     []ast.Expression
		expected string
		wantErr  bool
	}{
		{
			name:     "no arguments returns zero",
			args:     []ast.Expression{},
			expected: "0",
		},
		{
			name: "single identifier argument",
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
			},
			expected: "value.Nz(bar.Close, 0)",
		},
		{
			name: "identifier with literal replacement",
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 100.0},
			},
			expected: "value.Nz(bar.Close, 100.00)",
		},
		{
			name: "series historical access with default",
			args: []ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "sl_inp"},
					Property: &ast.Literal{Value: 1},
					Computed: true,
				},
			},
			expected: "value.Nz(sl_inpSeries.Get(1), 0)",
		},
		{
			name: "series historical access with replacement",
			args: []ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "value"},
					Property: &ast.Literal{Value: 2},
					Computed: true,
				},
				&ast.Literal{Value: -1.0},
			},
			expected: "value.Nz(valueSeries.Get(2), -1.00)",
		},
		{
			name: "literal with zero replacement",
			args: []ast.Expression{
				&ast.Literal{Value: 42.0},
				&ast.Literal{Value: 0.0},
			},
			expected: "value.Nz(42.00, 0.00)",
		},
		{
			name: "negative literal replacement",
			args: []ast.Expression{
				&ast.Identifier{Name: "x"},
				&ast.Literal{Value: -999.0},
			},
			expected: "value.Nz(xSeries.GetCurrent(), -999.00)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.generateNz(tt.args, gen)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("generateNz() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestValueHandlerGenerateInlineCall(t *testing.T) {
	handler := NewValueHandler()
	gen := &generator{
		variables: make(map[string]string),
		varInits:  make(map[string]ast.Expression),
	}

	tests := []struct {
		name     string
		funcName string
		args     []ast.Expression
		expected string
		wantErr  bool
	}{
		{
			name:     "na function dispatch",
			funcName: "na",
			args:     []ast.Expression{&ast.Identifier{Name: "close"}},
			expected: "math.IsNaN(bar.Close)",
		},
		{
			name:     "nz function dispatch",
			funcName: "nz",
			args:     []ast.Expression{&ast.Identifier{Name: "value"}},
			expected: "value.Nz(valueSeries.GetCurrent(), 0)",
		},
		{
			name:     "unsupported function",
			funcName: "fixnan",
			args:     []ast.Expression{},
			wantErr:  true,
		},
		{
			name:     "unknown function",
			funcName: "unknown",
			args:     []ast.Expression{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.GenerateInlineCall(tt.funcName, tt.args, gen)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("GenerateInlineCall() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestValueHandlerComplexExpressionArguments(t *testing.T) {
	handler := NewValueHandler()
	gen := &generator{
		variables: make(map[string]string),
		varInits:  make(map[string]ast.Expression),
	}

	tests := []struct {
		name        string
		funcName    string
		args        []ast.Expression
		expectStart string
	}{
		{
			name:     "na with binary expression",
			funcName: "na",
			args: []ast.Expression{
				&ast.BinaryExpression{
					Operator: "-",
					Left:     &ast.Identifier{Name: "close"},
					Right:    &ast.Identifier{Name: "open"},
				},
			},
			expectStart: "math.IsNaN(",
		},
		{
			name:     "nz with ternary result",
			funcName: "nz",
			args: []ast.Expression{
				&ast.ConditionalExpression{
					Test:       &ast.Identifier{Name: "condition"},
					Consequent: &ast.Literal{Value: 1.0},
					Alternate:  &ast.Literal{Value: 0.0},
				},
			},
			expectStart: "value.Nz(",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.GenerateInlineCall(tt.funcName, tt.args, gen)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.HasPrefix(result, tt.expectStart) {
				t.Errorf("expected result to start with %q, got %q", tt.expectStart, result)
			}
		})
	}
}

func TestValueHandlerIntegrationWithGenerator(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		args     []ast.Expression
	}{
		{
			name:     "nz with series access",
			funcName: "nz",
			args: []ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "value"},
					Property: &ast.Literal{Value: 1},
					Computed: true,
				},
			},
		},
		{
			name:     "na with identifier",
			funcName: "na",
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: ast.Identifier{Name: "test_var"},
								Init: &ast.CallExpression{
									Callee:    &ast.Identifier{Name: tt.funcName},
									Arguments: tt.args,
								},
							},
						},
					},
				},
			}

			_, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("GenerateStrategyCodeFromAST() error: %v", err)
			}
		})
	}
}
