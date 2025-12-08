package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestMathHandler_GenerateMathPow(t *testing.T) {
	mh := NewMathHandler()
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}

	tests := []struct {
		name     string
		args     []ast.Expression
		expected string
		wantErr  bool
	}{
		{
			name: "literal arguments",
			args: []ast.Expression{
				&ast.Literal{Value: 2.0},
				&ast.Literal{Value: 3.0},
			},
			expected: "math.Pow(2.00, 3.00)",
		},
		{
			name: "identifier arguments",
			args: []ast.Expression{
				&ast.Identifier{Name: "base"},
				&ast.Identifier{Name: "exp"},
			},
			expected: "math.Pow(baseSeries.GetCurrent(), expSeries.GetCurrent())",
		},
		{
			name: "mixed arguments",
			args: []ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "vf"},
					Property: &ast.Literal{Value: float64(0)},
					Computed: true,
				},
				&ast.Literal{Value: -1.0},
			},
			expected: "math.Pow(vfSeries.Get(0), -1.00)",
		},
		{
			name:    "wrong number of args",
			args:    []ast.Expression{&ast.Literal{Value: 2.0}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mh.GenerateMathCall("math.pow", tt.args, g)
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
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMathHandler_GenerateUnaryMath(t *testing.T) {
	mh := NewMathHandler()
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}

	tests := []struct {
		name     string
		funcName string
		args     []ast.Expression
		expected string
	}{
		{
			name:     "math.abs with literal",
			funcName: "math.abs",
			args: []ast.Expression{
				&ast.Literal{Value: -5.0},
			},
			expected: "math.Abs(-5.00)",
		},
		{
			name:     "math.sqrt with identifier",
			funcName: "math.sqrt",
			args: []ast.Expression{
				&ast.Identifier{Name: "value"},
			},
			expected: "math.Sqrt(valueSeries.GetCurrent())",
		},
		{
			name:     "math.floor with expression",
			funcName: "math.floor",
			args: []ast.Expression{
				&ast.BinaryExpression{
					Operator: "*",
					Left:     &ast.Literal{Value: 1.5},
					Right:    &ast.Literal{Value: 10.0},
				},
			},
			expected: "math.Floor((1.50 * 10.00))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mh.GenerateMathCall(tt.funcName, tt.args, g)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMathHandler_GenerateBinaryMath(t *testing.T) {
	mh := NewMathHandler()
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}

	tests := []struct {
		name     string
		funcName string
		args     []ast.Expression
		expected string
	}{
		{
			name:     "math.max with literals",
			funcName: "math.max",
			args: []ast.Expression{
				&ast.Literal{Value: 5.0},
				&ast.Literal{Value: 10.0},
			},
			expected: "math.Max(5.00, 10.00)",
		},
		{
			name:     "math.min with identifiers",
			funcName: "math.min",
			args: []ast.Expression{
				&ast.Identifier{Name: "a"},
				&ast.Identifier{Name: "b"},
			},
			expected: "math.Min(aSeries.GetCurrent(), bSeries.GetCurrent())",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mh.GenerateMathCall(tt.funcName, tt.args, g)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMathHandler_WithInputConstants(t *testing.T) {
	// Test that math.pow works with input constants
	mh := NewMathHandler()
	g := &generator{
		variables: make(map[string]string),
		constants: map[string]interface{}{
			"yA": "input.float",
		},
	}

	args := []ast.Expression{
		&ast.MemberExpression{
			Object:   &ast.Identifier{Name: "yA"},
			Property: &ast.Literal{Value: float64(0)},
			Computed: true,
		},
		&ast.UnaryExpression{
			Operator: "-",
			Argument: &ast.Literal{Value: 1.0},
		},
	}

	result, err := mh.GenerateMathCall("math.pow", args, g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use constant name directly, not Series access
	if !strings.Contains(result, "yA") {
		t.Errorf("expected result to contain 'yA' constant, got %q", result)
	}
	if strings.Contains(result, "yASeries") {
		t.Errorf("result should not use Series access for constant, got %q", result)
	}
}

func TestMathHandler_UnsupportedFunction(t *testing.T) {
	mh := NewMathHandler()
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}

	args := []ast.Expression{
		&ast.Literal{Value: 1.0},
	}

	_, err := mh.GenerateMathCall("math.unsupported", args, g)
	if err == nil {
		t.Error("expected error for unsupported function, got nil")
	}
}

func TestMathHandler_NormalizationEdgeCases(t *testing.T) {
	mh := NewMathHandler()
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}

	tests := []struct {
		name        string
		funcName    string
		args        []ast.Expression
		expectFunc  string
		description string
	}{
		{
			name:        "unprefixed abs normalized",
			funcName:    "abs",
			args:        []ast.Expression{&ast.Literal{Value: -5.0}},
			expectFunc:  "math.Abs",
			description: "Pine 'abs' → Go 'math.Abs'",
		},
		{
			name:        "prefixed math.abs normalized",
			funcName:    "math.abs",
			args:        []ast.Expression{&ast.Literal{Value: -5.0}},
			expectFunc:  "math.Abs",
			description: "Pine 'math.abs' → Go 'math.Abs'",
		},
		{
			name:        "unprefixed sqrt normalized",
			funcName:    "sqrt",
			args:        []ast.Expression{&ast.Literal{Value: 16.0}},
			expectFunc:  "math.Sqrt",
			description: "Pine 'sqrt' → Go 'math.Sqrt'",
		},
		{
			name:        "prefixed math.sqrt normalized",
			funcName:    "math.sqrt",
			args:        []ast.Expression{&ast.Literal{Value: 16.0}},
			expectFunc:  "math.Sqrt",
			description: "Pine 'math.sqrt' → Go 'math.Sqrt'",
		},
		{
			name:        "unprefixed max normalized",
			funcName:    "max",
			args:        []ast.Expression{&ast.Literal{Value: 5.0}, &ast.Literal{Value: 10.0}},
			expectFunc:  "math.Max",
			description: "Pine 'max' → Go 'math.Max'",
		},
		{
			name:        "prefixed math.max normalized",
			funcName:    "math.max",
			args:        []ast.Expression{&ast.Literal{Value: 5.0}, &ast.Literal{Value: 10.0}},
			expectFunc:  "math.Max",
			description: "Pine 'math.max' → Go 'math.Max'",
		},
		{
			name:        "unprefixed min normalized",
			funcName:    "min",
			args:        []ast.Expression{&ast.Literal{Value: 5.0}, &ast.Literal{Value: 10.0}},
			expectFunc:  "math.Min",
			description: "Pine 'min' → Go 'math.Min'",
		},
		{
			name:        "prefixed math.min normalized",
			funcName:    "math.min",
			args:        []ast.Expression{&ast.Literal{Value: 5.0}, &ast.Literal{Value: 10.0}},
			expectFunc:  "math.Min",
			description: "Pine 'math.min' → Go 'math.Min'",
		},
		{
			name:        "unprefixed floor normalized",
			funcName:    "floor",
			args:        []ast.Expression{&ast.Literal{Value: 3.7}},
			expectFunc:  "math.Floor",
			description: "Pine 'floor' → Go 'math.Floor'",
		},
		{
			name:        "unprefixed ceil normalized",
			funcName:    "ceil",
			args:        []ast.Expression{&ast.Literal{Value: 3.2}},
			expectFunc:  "math.Ceil",
			description: "Pine 'ceil' → Go 'math.Ceil'",
		},
		{
			name:        "unprefixed round normalized",
			funcName:    "round",
			args:        []ast.Expression{&ast.Literal{Value: 3.5}},
			expectFunc:  "math.Round",
			description: "Pine 'round' → Go 'math.Round'",
		},
		{
			name:        "unprefixed log normalized",
			funcName:    "log",
			args:        []ast.Expression{&ast.Literal{Value: 10.0}},
			expectFunc:  "math.Log",
			description: "Pine 'log' → Go 'math.Log'",
		},
		{
			name:        "unprefixed exp normalized",
			funcName:    "exp",
			args:        []ast.Expression{&ast.Literal{Value: 2.0}},
			expectFunc:  "math.Exp",
			description: "Pine 'exp' → Go 'math.Exp'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mh.GenerateMathCall(tt.funcName, tt.args, g)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.HasPrefix(result, tt.expectFunc+"(") {
				t.Errorf("%s: expected result to start with %q, got %q", tt.description, tt.expectFunc+"(", result)
			}
		})
	}
}

func TestMathHandler_CaseInsensitiveMatching(t *testing.T) {
	mh := NewMathHandler()
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}

	tests := []struct {
		name        string
		funcName    string
		args        []ast.Expression
		expectStart string
	}{
		{
			name:        "uppercase ABS normalized",
			funcName:    "ABS",
			args:        []ast.Expression{&ast.Literal{Value: -5.0}},
			expectStart: "math.Abs(",
		},
		{
			name:        "mixed case Sqrt normalized",
			funcName:    "Sqrt",
			args:        []ast.Expression{&ast.Literal{Value: 16.0}},
			expectStart: "math.Sqrt(",
		},
		{
			name:        "uppercase MATH.MAX normalized",
			funcName:    "MATH.MAX",
			args:        []ast.Expression{&ast.Literal{Value: 5.0}, &ast.Literal{Value: 10.0}},
			expectStart: "math.Max(",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mh.GenerateMathCall(tt.funcName, tt.args, g)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.HasPrefix(result, tt.expectStart) {
				t.Errorf("expected result to start with %q, got %q", tt.expectStart, result)
			}
		})
	}
}

func TestMathHandler_ArgumentCountValidation(t *testing.T) {
	mh := NewMathHandler()
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}

	tests := []struct {
		name     string
		funcName string
		argCount int
		wantErr  bool
	}{
		{
			name:     "pow with 1 arg fails",
			funcName: "math.pow",
			argCount: 1,
			wantErr:  true,
		},
		{
			name:     "pow with 3 args fails",
			funcName: "math.pow",
			argCount: 3,
			wantErr:  true,
		},
		{
			name:     "abs with 0 args fails",
			funcName: "abs",
			argCount: 0,
			wantErr:  true,
		},
		{
			name:     "abs with 2 args fails",
			funcName: "abs",
			argCount: 2,
			wantErr:  true,
		},
		{
			name:     "max with 1 arg fails",
			funcName: "max",
			argCount: 1,
			wantErr:  true,
		},
		{
			name:     "max with 3 args fails",
			funcName: "max",
			argCount: 3,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := make([]ast.Expression, tt.argCount)
			for i := 0; i < tt.argCount; i++ {
				args[i] = &ast.Literal{Value: float64(i)}
			}

			_, err := mh.GenerateMathCall(tt.funcName, args, g)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestMathHandler_ComplexExpressionArguments(t *testing.T) {
	mh := NewMathHandler()
	g := &generator{
		variables:      make(map[string]string),
		constants:      make(map[string]interface{}),
		tempVarMgr:     NewTempVariableManager(nil),
		builtinHandler: NewBuiltinIdentifierHandler(),
	}
	g.tempVarMgr.gen = g

	tests := []struct {
		name        string
		funcName    string
		args        []ast.Expression
		expectStart string
	}{
		{
			name:     "abs with binary expression",
			funcName: "abs",
			args: []ast.Expression{
				&ast.BinaryExpression{
					Operator: "-",
					Left:     &ast.Identifier{Name: "close"},
					Right:    &ast.Identifier{Name: "open"},
				},
			},
			expectStart: "math.Abs(",
		},
		{
			name:     "max with literals",
			funcName: "max",
			args: []ast.Expression{
				&ast.Literal{Value: 5.0},
				&ast.Literal{Value: 0.0},
			},
			expectStart: "math.Max(",
		},
		{
			name:     "sqrt with identifier",
			funcName: "sqrt",
			args: []ast.Expression{
				&ast.Identifier{Name: "value"},
			},
			expectStart: "math.Sqrt(",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mh.GenerateMathCall(tt.funcName, tt.args, g)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.HasPrefix(result, tt.expectStart) {
				t.Errorf("expected result to start with %q, got %q", tt.expectStart, result)
			}
		})
	}
}
