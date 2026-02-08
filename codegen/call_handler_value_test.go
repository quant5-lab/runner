package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestValueCallHandler_CanHandle validates value function recognition in call routing
 *
 * Tests that ValueCallHandler correctly identifies nz/na functions and defers
 * other functions to subsequent handlers in the chain.
 */
func TestValueCallHandler_CanHandle(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"nz", "nz", true},
		{"na", "na", true},

		{"fixnan not handled", "fixnan", false},
		{"ta.sma", "ta.sma", false},
		{"unprefixed sma", "sma", false},
		{"math.abs", "math.abs", false},
		{"abs", "abs", false},
		{"plot", "plot", false},
		{"user function", "getPoles", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewValueCallHandler()
			got := handler.CanHandle(tt.funcName)
			if got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

/* TestValueCallHandler_GenerateCode validates value function code generation
 *
 * Tests that nz/na functions are correctly translated to runtime calls
 * with proper argument handling and default values.
 */
func TestValueCallHandler_GenerateCode(t *testing.T) {
	tests := []struct {
		name           string
		funcName       string
		args           []ast.Expression
		expectError    bool
		validateOutput func(t *testing.T, code string)
	}{
		{
			name:     "nz with single identifier argument",
			funcName: "nz",
			args: []ast.Expression{
				&ast.Identifier{Name: "pre"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "value.Nz(") {
					t.Error("Expected value.Nz translation")
				}
				if !strings.Contains(code, ", 0)") {
					t.Error("Expected default replacement value 0")
				}
			},
		},
		{
			name:     "nz with explicit replacement value",
			funcName: "nz",
			args: []ast.Expression{
				&ast.Identifier{Name: "pre"},
				&ast.Literal{Value: 100},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "value.Nz(") {
					t.Error("Expected value.Nz translation")
				}
				if !strings.Contains(code, "100") {
					t.Error("Expected explicit replacement value 100")
				}
			},
		},
		{
			name:        "nz with no arguments returns literal zero",
			funcName:    "nz",
			args:        []ast.Expression{},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if code != "0" {
					t.Errorf("Expected literal '0', got %q", code)
				}
			},
		},
		{
			name:     "nz with binary expression argument",
			funcName: "nz",
			args: []ast.Expression{
				&ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "a"},
					Operator: "+",
					Right:    &ast.Identifier{Name: "b"},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "value.Nz(") {
					t.Error("Expected value.Nz translation")
				}
			},
		},
		{
			name:     "na with identifier argument",
			funcName: "na",
			args: []ast.Expression{
				&ast.Identifier{Name: "x"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "math.IsNaN(") {
					t.Error("Expected math.IsNaN translation")
				}
			},
		},
		{
			name:        "na with no arguments returns literal true",
			funcName:    "na",
			args:        []ast.Expression{},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if code != "true" {
					t.Errorf("Expected literal 'true', got %q", code)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			handler := NewValueCallHandler()

			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: tt.funcName},
				Arguments: tt.args,
			}

			code, err := handler.GenerateCode(g, call)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if code == "" && tt.funcName != "" {
				t.Error("Expected generated code, got empty string")
			}

			if tt.validateOutput != nil {
				tt.validateOutput(t, code)
			}
		})
	}
}

/* TestValueCallHandler_GenerateCode_SkipsUnhandled validates passthrough behavior
 *
 * Tests that ValueCallHandler returns empty code for functions it doesn't handle,
 * allowing the router to continue to the next handler.
 */
func TestValueCallHandler_GenerateCode_SkipsUnhandled(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		args     []ast.Expression
	}{
		{
			name:     "ta.sma",
			funcName: "ta.sma",
			args:     []ast.Expression{&ast.Identifier{Name: "close"}},
		},
		{
			name:     "fixnan",
			funcName: "fixnan",
			args:     []ast.Expression{&ast.Identifier{Name: "x"}},
		},
		{
			name:     "user function",
			funcName: "myFunc",
			args:     []ast.Expression{},
		},
		{
			name:     "math.abs",
			funcName: "math.abs",
			args:     []ast.Expression{&ast.Literal{Value: -5}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			handler := NewValueCallHandler()

			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: tt.funcName},
				Arguments: tt.args,
			}

			code, err := handler.GenerateCode(g, call)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if code != "" {
				t.Errorf("expected empty code for unhandled function %q, got %q", tt.funcName, code)
			}
		})
	}
}

/* TestValueCallHandler_RouterIntegration validates handler priority in router chain
 *
 * Tests that ValueCallHandler is positioned correctly in the router chain,
 * handling value functions before TAIndicatorCallHandler.
 */
func TestValueCallHandler_RouterIntegration(t *testing.T) {
	tests := []struct {
		name              string
		funcName          string
		args              []ast.Expression
		inArrowContext    bool
		expectContains    string
		expectNotContains string
	}{
		{
			name:           "nz in series context",
			funcName:       "nz",
			args:           []ast.Expression{&ast.Identifier{Name: "pre"}},
			inArrowContext: false,
			expectContains: "value.Nz",
		},
		{
			name:           "na in series context",
			funcName:       "na",
			args:           []ast.Expression{&ast.Identifier{Name: "x"}},
			inArrowContext: false,
			expectContains: "math.IsNaN",
		},
		{
			name:              "nz in arrow context not hijacked by TAIndicator",
			funcName:          "nz",
			args:              []ast.Expression{&ast.Identifier{Name: "pre"}},
			inArrowContext:    true,
			expectContains:    "value.Nz",
			expectNotContains: "TA function nz requires",
		},
		{
			name:              "na in arrow context not hijacked by TAIndicator",
			funcName:          "na",
			args:              []ast.Expression{&ast.Identifier{Name: "x"}},
			inArrowContext:    true,
			expectContains:    "math.IsNaN",
			expectNotContains: "TA function na requires",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewCallExpressionRouter()
			g := newTestGenerator()
			g.inArrowFunctionBody = tt.inArrowContext

			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: tt.funcName},
				Arguments: tt.args,
			}

			code, err := router.RouteCall(g, call)
			if err != nil {
				t.Fatalf("RouteCall(%s) returned error: %v", tt.funcName, err)
			}

			if tt.expectContains != "" && !strings.Contains(code, tt.expectContains) {
				t.Errorf("expected code to contain %q, got %q", tt.expectContains, code)
			}

			if tt.expectNotContains != "" && strings.Contains(code, tt.expectNotContains) {
				t.Errorf("expected code NOT to contain %q, got %q", tt.expectNotContains, code)
			}
		})
	}
}

/* TestValueCallHandler_InArrowFunctions validates value calls within arrow function bodies
 *
 * Tests that value functions work correctly when used in arrow function context,
 * verifying the handler is properly integrated before TAIndicatorCallHandler.
 */
func TestValueCallHandler_InArrowFunctions(t *testing.T) {
	tests := []struct {
		name           string
		funcName       string
		args           []ast.Expression
		inArrowContext bool
		validateOutput func(t *testing.T, code string)
	}{
		{
			name:     "nz as UDF argument in arrow context",
			funcName: "nz",
			args: []ast.Expression{
				&ast.Identifier{Name: "pre"},
				&ast.Literal{Value: 0},
			},
			inArrowContext: true,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "value.Nz") {
					t.Error("Expected value.Nz in arrow context")
				}
			},
		},
		{
			name:     "na as UDF argument in arrow context",
			funcName: "na",
			args: []ast.Expression{
				&ast.Identifier{Name: "x"},
			},
			inArrowContext: true,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "math.IsNaN") {
					t.Error("Expected math.IsNaN in arrow context")
				}
			},
		},
		{
			name:     "nz with literal argument in arrow context",
			funcName: "nz",
			args: []ast.Expression{
				&ast.Literal{Value: 5.0},
			},
			inArrowContext: true,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "value.Nz") {
					t.Error("Expected value.Nz in arrow context")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewCallExpressionRouter()
			g := newTestGenerator()
			g.inArrowFunctionBody = tt.inArrowContext

			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: tt.funcName},
				Arguments: tt.args,
			}

			code, err := router.RouteCall(g, call)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if code == "" {
				t.Error("Expected generated code, got empty string")
			}

			if tt.validateOutput != nil {
				tt.validateOutput(t, code)
			}
		})
	}
}

/* TestValueCallHandler_EdgeCases validates boundary conditions
 *
 * Tests edge cases like empty function names, nil arguments, and
 * context-specific behavior.
 */
func TestValueCallHandler_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		buildCall   func() *ast.CallExpression
		expectError bool
		expectEmpty bool
	}{
		{
			name: "empty function name",
			buildCall: func() *ast.CallExpression {
				return &ast.CallExpression{
					Callee:    &ast.Identifier{Name: ""},
					Arguments: []ast.Expression{},
				}
			},
			expectError: false,
			expectEmpty: true,
		},
		{
			name: "nz with member expression callee",
			buildCall: func() *ast.CallExpression {
				return &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "nz"},
					},
					Arguments: []ast.Expression{&ast.Identifier{Name: "x"}},
				}
			},
			expectError: false,
			expectEmpty: true,
		},
		{
			name: "nz with multiple replacement values uses second arg",
			buildCall: func() *ast.CallExpression {
				return &ast.CallExpression{
					Callee: &ast.Identifier{Name: "nz"},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "x"},
						&ast.Literal{Value: 42},
						&ast.Literal{Value: 99},
					},
				}
			},
			expectError: false,
			expectEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			handler := NewValueCallHandler()

			call := tt.buildCall()
			code, err := handler.GenerateCode(g, call)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.expectEmpty && code != "" {
				t.Errorf("Expected empty code, got %q", code)
			}

			if !tt.expectEmpty && code == "" {
				t.Error("Expected generated code, got empty string")
			}
		})
	}
}
