package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Expression type handler tests - validates AST expression handling and ObjectExpression architectural invariant */

func TestGenerateExpression_SupportedTypes(t *testing.T) {
	tests := []struct {
		name        string
		expr        ast.Expression
		expectError bool
		setupGen    func(*generator)
	}{
		{
			name: "CallExpression",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.sma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: float64(20)},
				},
			},
			expectError: false,
		},
		{
			name: "BinaryExpression in arrow context",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Literal{Value: float64(10)},
				Right:    &ast.Literal{Value: float64(20)},
			},
			expectError: false,
			setupGen: func(g *generator) {
				g.inArrowFunctionBody = true
			},
		},
		{
			name: "LogicalExpression",
			expr: &ast.LogicalExpression{
				Operator: "and",
				Left:     &ast.Identifier{Name: "condition1"},
				Right:    &ast.Identifier{Name: "condition2"},
			},
			expectError: false,
		},
		{
			name: "ConditionalExpression",
			expr: &ast.ConditionalExpression{
				Test:       &ast.Identifier{Name: "condition"},
				Consequent: &ast.Literal{Value: float64(100)},
				Alternate:  &ast.Literal{Value: float64(200)},
			},
			expectError: false,
		},
		{
			name: "UnaryExpression",
			expr: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.Literal{Value: float64(50)},
			},
			expectError: false,
		},
		{
			name:        "Literal",
			expr:        &ast.Literal{Value: float64(42)},
			expectError: false,
		},
		{
			name: "MemberExpression",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 1},
				Computed: true,
			},
			expectError: false,
		},
		{
			name:        "Identifier in arrow context",
			expr:        &ast.Identifier{Name: "myVar"},
			expectError: false,
			setupGen: func(g *generator) {
				g.inArrowFunctionBody = true
				g.variables["myVar"] = "float"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			if tt.setupGen != nil {
				tt.setupGen(gen)
			}

			_, err := gen.generateExpression(tt.expr)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGenerateExpression_ObjectExpression(t *testing.T) {
	tests := []struct {
		name            string
		expr            ast.Expression
		expectedErrText string
	}{
		{
			name: "ObjectExpression with empty properties",
			expr: &ast.ObjectExpression{
				NodeType:   ast.TypeObjectExpression,
				Properties: []ast.Property{},
			},
			expectedErrText: "ObjectExpression should not reach generateExpression",
		},
		{
			name: "ObjectExpression with named arguments",
			expr: &ast.ObjectExpression{
				NodeType: ast.TypeObjectExpression,
				Properties: []ast.Property{
					{
						Key:   &ast.Identifier{Name: "stop"},
						Value: &ast.Literal{Value: float64(48000)},
					},
					{
						Key:   &ast.Identifier{Name: "limit"},
						Value: &ast.Literal{Value: float64(58000)},
					},
				},
			},
			expectedErrText: "call handlers must use ArgumentExtractor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()

			_, err := gen.generateExpression(tt.expr)

			if err == nil {
				t.Fatal("Expected error for ObjectExpression, got nil")
			}

			if !strings.Contains(err.Error(), tt.expectedErrText) {
				t.Errorf("Error message %q does not contain %q", err.Error(), tt.expectedErrText)
			}
		})
	}
}

func TestExtractSeriesExpression_SupportedTypes(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		expected string
		setupGen func(*generator)
	}{
		{
			name:     "Literal float",
			expr:     &ast.Literal{Value: float64(100.5)},
			expected: "100.5",
		},
		{
			name:     "Literal int",
			expr:     &ast.Literal{Value: 42},
			expected: "42.0",
		},
		{
			name:     "Identifier user variable",
			expr:     &ast.Identifier{Name: "myVar"},
			expected: "myVarSeries.GetCurrent()",
		},
		{
			name: "MemberExpression close[0]",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 0},
				Computed: true,
			},
			expected: "bar.Close",
		},
		{
			name: "MemberExpression user variable subscript",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "sma"},
				Property: &ast.Literal{Value: 1},
				Computed: true,
			},
			expected: "smaSeries.Get(1)",
		},
		{
			name: "BinaryExpression addition",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Literal{Value: float64(10)},
				Right:    &ast.Literal{Value: float64(20)},
			},
			expected: "(10 + 20)",
		},
		{
			name: "UnaryExpression negation",
			expr: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.Literal{Value: float64(50)},
			},
			expected: "-50",
		},
		{
			name: "CallExpression",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.sma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: float64(20)},
				},
			},
			expected: "ta_smaSeries.GetCurrent()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			if tt.setupGen != nil {
				tt.setupGen(gen)
			}

			result := gen.extractSeriesExpression(tt.expr)

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractSeriesExpression_ObjectExpression(t *testing.T) {
	tests := []struct {
		name             string
		expr             ast.Expression
		expectedSentinel string
	}{
		{
			name: "ObjectExpression returns diagnostic sentinel",
			expr: &ast.ObjectExpression{
				NodeType:   ast.TypeObjectExpression,
				Properties: []ast.Property{},
			},
			expectedSentinel: "/* ERROR: ObjectExpression requires ArgumentExtractor */",
		},
		{
			name: "ObjectExpression with properties",
			expr: &ast.ObjectExpression{
				NodeType: ast.TypeObjectExpression,
				Properties: []ast.Property{
					{
						Key:   &ast.Identifier{Name: "length"},
						Value: &ast.Literal{Value: float64(20)},
					},
				},
			},
			expectedSentinel: "/* ERROR: ObjectExpression requires ArgumentExtractor */",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()

			result := gen.extractSeriesExpression(tt.expr)

			if result != tt.expectedSentinel {
				t.Errorf("Expected sentinel %q, got %q", tt.expectedSentinel, result)
			}
		})
	}
}

func TestArgumentExtractor_NamedArguments(t *testing.T) {
	tests := []struct {
		name      string
		args      []ast.Expression
		argName   string
		wantValue string
		wantFound bool
	}{
		{
			name: "Extract named argument from ObjectExpression",
			args: []ast.Expression{
				&ast.Literal{Value: "Exit"},
				&ast.Literal{Value: "Long"},
				&ast.ObjectExpression{
					Properties: []ast.Property{
						{
							Key:   &ast.Identifier{Name: "stop"},
							Value: &ast.Literal{Value: float64(48000)},
						},
						{
							Key:   &ast.Identifier{Name: "limit"},
							Value: &ast.Literal{Value: float64(58000)},
						},
					},
				},
			},
			argName:   "stop",
			wantValue: "48000",
			wantFound: true,
		},
		{
			name: "Named argument not found",
			args: []ast.Expression{
				&ast.Literal{Value: "Exit"},
				&ast.ObjectExpression{
					Properties: []ast.Property{
						{
							Key:   &ast.Identifier{Name: "stop"},
							Value: &ast.Literal{Value: float64(48000)},
						},
					},
				},
			},
			argName:   "profit",
			wantValue: "",
			wantFound: false,
		},
		{
			name: "No ObjectExpression in arguments",
			args: []ast.Expression{
				&ast.Literal{Value: "Exit"},
				&ast.Literal{Value: "Long"},
			},
			argName:   "stop",
			wantValue: "",
			wantFound: false,
		},
		{
			name: "ObjectExpression with identifier value",
			args: []ast.Expression{
				&ast.ObjectExpression{
					Properties: []ast.Property{
						{
							Key:   &ast.Identifier{Name: "source"},
							Value: &ast.Identifier{Name: "close"},
						},
					},
				},
			},
			argName:   "source",
			wantValue: "bar.Close",
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			extractor := &ArgumentExtractor{generator: gen}

			value, found := extractor.ExtractNamedArgument(tt.args, tt.argName)

			if found != tt.wantFound {
				t.Errorf("Expected found=%v, got %v", tt.wantFound, found)
			}

			if tt.wantFound && value != tt.wantValue {
				t.Errorf("Expected value %q, got %q", tt.wantValue, value)
			}
		})
	}
}

func TestGenerateExpression_UnsupportedExpressionType(t *testing.T) {
	gen := newTestGenerator()

	/* Mock unsupported expression type */
	type unsupportedExpression struct {
		ast.Expression
	}
	expr := &unsupportedExpression{}

	_, err := gen.generateExpression(expr)

	if err == nil {
		t.Fatal("Expected error for unsupported expression type, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported expression type") {
		t.Errorf("Expected 'unsupported expression type' in error, got: %v", err)
	}
}
