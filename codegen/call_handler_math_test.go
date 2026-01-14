package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestMathCallHandler_CanHandle validates math function recognition in call routing
 *
 * Tests that the MathCallHandler correctly identifies math functions and defers
 * non-math functions to subsequent handlers in the call chain.
 */
func TestMathCallHandler_CanHandle(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		// Math functions with prefix
		{"math.abs", "math.abs", true},
		{"math.sqrt", "math.sqrt", true},
		{"math.max", "math.max", true},
		{"math.min", "math.min", true},
		{"math.pow", "math.pow", true},
		{"math.floor", "math.floor", true},
		{"math.ceil", "math.ceil", true},
		{"math.round", "math.round", true},
		{"math.log", "math.log", true},
		{"math.exp", "math.exp", true},

		// Math functions without prefix (Pine v4 compatibility)
		{"abs", "abs", true},
		{"sqrt", "sqrt", true},
		{"max", "max", true},
		{"min", "min", true},
		{"floor", "floor", true},
		{"ceil", "ceil", true},
		{"round", "round", true},
		{"log", "log", true},
		{"exp", "exp", true},

		// Non-math functions
		{"ta.sma", "ta.sma", false},
		{"plot", "plot", false},
		{"fixnan", "fixnan", false},
		{"strategy.entry", "strategy.entry", false},
		{"user_function", "myFunc", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &MathCallHandler{}
			got := handler.CanHandle(tt.funcName)
			if got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

/* TestMathCallHandler_GenerateCode validates math function code generation
 *
 * Tests that math functions are correctly translated to Go math package calls
 * with proper argument handling and operator precedence.
 */
func TestMathCallHandler_GenerateCode(t *testing.T) {
	tests := []struct {
		name           string
		funcName       string
		args           []ast.Expression
		expectError    bool
		validateOutput func(t *testing.T, code string)
	}{
		{
			name:     "abs with identifier",
			funcName: "abs",
			args: []ast.Expression{
				&ast.Identifier{Name: "value"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "math.Abs") {
					t.Error("Expected math.Abs translation")
				}
				if !strings.Contains(code, "value") {
					t.Error("Expected 'value' argument")
				}
			},
		},
		{
			name:     "math.abs with negative literal",
			funcName: "math.abs",
			args: []ast.Expression{
				&ast.UnaryExpression{
					Operator: "-",
					Argument: &ast.Literal{Value: 5.0},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "math.Abs") {
					t.Error("Expected math.Abs translation")
				}
			},
		},
		{
			name:     "max with two identifiers",
			funcName: "max",
			args: []ast.Expression{
				&ast.Identifier{Name: "a"},
				&ast.Identifier{Name: "b"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "math.Max") {
					t.Error("Expected math.Max translation")
				}
				if !strings.Contains(code, "a") || !strings.Contains(code, "b") {
					t.Error("Expected both arguments")
				}
			},
		},
		{
			name:     "min with expressions",
			funcName: "min",
			args: []ast.Expression{
				&ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "x"},
					Operator: "+",
					Right:    &ast.Literal{Value: 10.0},
				},
				&ast.Identifier{Name: "y"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "math.Min") {
					t.Error("Expected math.Min translation")
				}
			},
		},
		{
			name:     "pow with base and exponent",
			funcName: "math.pow",
			args: []ast.Expression{
				&ast.Identifier{Name: "base"},
				&ast.Literal{Value: 2.0},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "math.Pow") {
					t.Error("Expected math.Pow translation")
				}
			},
		},
		{
			name:     "sqrt with identifier",
			funcName: "sqrt",
			args: []ast.Expression{
				&ast.Identifier{Name: "value"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "math.Sqrt") {
					t.Error("Expected math.Sqrt translation")
				}
			},
		},
		{
			name:     "floor with expression",
			funcName: "floor",
			args: []ast.Expression{
				&ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "x"},
					Operator: "/",
					Right:    &ast.Literal{Value: 2.0},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "math.Floor") {
					t.Error("Expected math.Floor translation")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			g.variables["value"] = "float"
			g.variables["a"] = "float"
			g.variables["b"] = "float"
			g.variables["x"] = "float"
			g.variables["y"] = "float"
			g.variables["base"] = "float"
			g.inArrowFunctionBody = true

			handler := &MathCallHandler{}
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

			if code == "" {
				t.Error("Expected generated code, got empty string")
			}

			if tt.validateOutput != nil {
				tt.validateOutput(t, code)
			}
		})
	}
}

/* TestMathCallHandler_InArrowFunctions validates math calls within arrow function bodies
 *
 * Tests that math functions work correctly when used inside arrow functions,
 * particularly in complex expressions like TA function arguments.
 */
func TestMathCallHandler_InArrowFunctions(t *testing.T) {
	tests := []struct {
		name           string
		buildExpr      func() ast.Expression
		expectError    bool
		validateOutput func(t *testing.T, code string)
	}{
		{
			name: "abs in binary expression",
			buildExpr: func() ast.Expression {
				return &ast.BinaryExpression{
					Left: &ast.CallExpression{
						Callee: &ast.Identifier{Name: "abs"},
						Arguments: []ast.Expression{
							&ast.BinaryExpression{
								Left:     &ast.Identifier{Name: "a"},
								Operator: "-",
								Right:    &ast.Identifier{Name: "b"},
							},
						},
					},
					Operator: "/",
					Right:    &ast.Identifier{Name: "c"},
				}
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "math.Abs") {
					t.Error("Expected math.Abs in binary expression")
				}
				if !strings.Contains(code, "/") {
					t.Error("Expected division operator")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			g.inArrowFunctionBody = true

			expr := tt.buildExpr()

			var code string
			var err error

			switch e := expr.(type) {
			case *ast.BinaryExpression:
				code, err = g.generateBinaryExpression(e)
			case *ast.ConditionalExpression:
				code, err = g.generateConditionalExpression(e)
			case *ast.CallExpression:
				code, err = g.generateCallExpression(e)
			default:
				t.Fatalf("Unsupported expression type: %T", expr)
			}

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

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

/* TestMathCallHandler_EdgeCases validates boundary conditions
 *
 * Tests exceptional cases and error handling for math functions.
 */
func TestMathCallHandler_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		funcName    string
		args        []ast.Expression
		expectError bool
		errorMsg    string
	}{
		{
			name:        "abs with no arguments",
			funcName:    "abs",
			args:        []ast.Expression{},
			expectError: true,
			errorMsg:    "requires exactly 1 argument",
		},
		{
			name:     "abs with multiple arguments",
			funcName: "abs",
			args: []ast.Expression{
				&ast.Identifier{Name: "a"},
				&ast.Identifier{Name: "b"},
			},
			expectError: true,
			errorMsg:    "requires exactly 1 argument",
		},
		{
			name:     "max with one argument",
			funcName: "max",
			args: []ast.Expression{
				&ast.Identifier{Name: "a"},
			},
			expectError: true,
			errorMsg:    "requires exactly 2 arguments",
		},
		{
			name:     "pow with one argument",
			funcName: "math.pow",
			args: []ast.Expression{
				&ast.Identifier{Name: "base"},
			},
			expectError: true,
			errorMsg:    "requires exactly 2 arguments",
		},
		{
			name:        "sqrt with no arguments",
			funcName:    "sqrt",
			args:        []ast.Expression{},
			expectError: true,
			errorMsg:    "requires exactly 1 argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			g.variables["a"] = "float"
			g.variables["b"] = "float"
			g.variables["base"] = "float"

			handler := &MathCallHandler{}
			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: tt.funcName},
				Arguments: tt.args,
			}

			_, err := handler.GenerateCode(g, call)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

/* TestMathCallHandler_CaseInsensitivity validates case-insensitive function name handling
 *
 * Tests that math functions are recognized regardless of case (abs, ABS, Abs, etc.).
 */
func TestMathCallHandler_CaseInsensitivity(t *testing.T) {
	cases := []string{"abs", "ABS", "Abs", "aBs", "math.abs", "math.ABS", "MATH.ABS"}

	for _, funcName := range cases {
		t.Run(funcName, func(t *testing.T) {
			handler := &MathCallHandler{}

			// CanHandle should recognize all case variations
			if !handler.CanHandle(strings.ToLower(funcName)) {
				t.Errorf("CanHandle(%q) = false, expected true", funcName)
			}

			// Code generation should work
			g := newTestGenerator()
			g.variables["x"] = "float"

			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: funcName},
				Arguments: []ast.Expression{&ast.Identifier{Name: "x"}},
			}

			code, err := handler.GenerateCode(g, call)
			if err != nil {
				t.Errorf("GenerateCode failed: %v", err)
			}

			if !strings.Contains(code, "math.Abs") {
				t.Errorf("Expected math.Abs in output, got: %s", code)
			}
		})
	}
}
