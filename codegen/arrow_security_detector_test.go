package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestArrowSecurityDetector_ContainsSecurityCall verifies detection across all statement/expression shapes */
func TestArrowSecurityDetector_ContainsSecurityCall(t *testing.T) {
	tests := []struct {
		name     string
		body     []ast.Node
		expected bool
	}{
		{
			name: "expression_statement_direct_call",
			body: []ast.Node{
				&ast.ExpressionStatement{
					Expression: makeSecurityCall("BTCUSDT", "1D", "close"),
				},
			},
			expected: true,
		},
		{
			name: "variable_declaration_init",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID:   &ast.Identifier{Name: "val"},
							Init: makeSecurityCall("BTCUSDT", "1D", "close"),
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "conditional_consequent",
			body: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.ConditionalExpression{
						Test:       &ast.Identifier{Name: "flag"},
						Consequent: makeSecurityCall("BTCUSDT", "1D", "close"),
						Alternate:  &ast.Identifier{Name: "close"},
					},
				},
			},
			expected: true,
		},
		{
			name: "conditional_alternate",
			body: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.ConditionalExpression{
						Test:       &ast.Identifier{Name: "flag"},
						Consequent: &ast.Identifier{Name: "close"},
						Alternate:  makeSecurityCall("ETHUSDT", "1h", "high"),
					},
				},
			},
			expected: true,
		},
		{
			name: "binary_expression_left",
			body: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.BinaryExpression{
						Operator: "+",
						Left:     makeSecurityCall("BTCUSDT", "1D", "close"),
						Right:    &ast.Literal{Value: 5.0},
					},
				},
			},
			expected: true,
		},
		{
			name: "binary_expression_right",
			body: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.BinaryExpression{
						Operator: "*",
						Left:     &ast.Literal{Value: 2.0},
						Right:    makeSecurityCall("BTCUSDT", "1D", "close"),
					},
				},
			},
			expected: true,
		},
		{
			name: "logical_expression",
			body: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.LogicalExpression{
						Operator: "and",
						Left:     &ast.Literal{Value: true},
						Right:    makeSecurityCall("BTCUSDT", "1D", "close"),
					},
				},
			},
			expected: true,
		},
		{
			name: "unary_expression",
			body: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.UnaryExpression{
						Operator: "-",
						Argument: makeSecurityCall("BTCUSDT", "1D", "close"),
					},
				},
			},
			expected: true,
		},
		{
			name: "nested_in_function_argument",
			body: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "ta"},
							Property: &ast.Identifier{Name: "sma"},
						},
						Arguments: []ast.Expression{
							makeSecurityCall("BTCUSDT", "1D", "close"),
							&ast.Literal{Value: 20.0},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "if_statement_test",
			body: []ast.Node{
				&ast.IfStatement{
					Test:       makeSecurityCall("BTCUSDT", "1D", "close"),
					Consequent: []ast.Node{&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "x"}}},
				},
			},
			expected: true,
		},
		{
			name: "if_statement_consequent",
			body: []ast.Node{
				&ast.IfStatement{
					Test: &ast.Identifier{Name: "flag"},
					Consequent: []ast.Node{
						&ast.ExpressionStatement{Expression: makeSecurityCall("BTCUSDT", "1D", "close")},
					},
				},
			},
			expected: true,
		},
		{
			name: "if_statement_alternate",
			body: []ast.Node{
				&ast.IfStatement{
					Test:       &ast.Identifier{Name: "flag"},
					Consequent: []ast.Node{&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "x"}}},
					Alternate: []ast.Node{
						&ast.ExpressionStatement{Expression: makeSecurityCall("BTCUSDT", "1D", "close")},
					},
				},
			},
			expected: true,
		},
		{
			name: "for_loop_body",
			body: []ast.Node{
				&ast.ForStatement{
					Body: []ast.Node{
						&ast.ExpressionStatement{Expression: makeSecurityCall("BTCUSDT", "1D", "close")},
					},
				},
			},
			expected: true,
		},
		{
			name: "multiple_security_calls",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{ID: &ast.Identifier{Name: "a"}, Init: makeSecurityCall("BTCUSDT", "1D", "close")},
					},
				},
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{ID: &ast.Identifier{Name: "b"}, Init: makeSecurityCall("ETHUSDT", "1h", "high")},
					},
				},
			},
			expected: true,
		},
		{
			name: "ta_function_only",
			body: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "ta"},
							Property: &ast.Identifier{Name: "sma"},
						},
						Arguments: []ast.Expression{
							&ast.Identifier{Name: "close"},
							&ast.Literal{Value: float64(20)},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "arithmetic_only",
			body: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.BinaryExpression{
						Operator: "*",
						Left:     &ast.Identifier{Name: "x"},
						Right:    &ast.Literal{Value: 2.0},
					},
				},
			},
			expected: false,
		},
		{
			name:     "empty_body",
			body:     []ast.Node{},
			expected: false,
		},
		{
			name:     "nil_body",
			body:     nil,
			expected: false,
		},
	}

	detector := NewArrowSecurityDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrow := &ast.ArrowFunctionExpression{Body: tt.body}
			result := detector.ContainsSecurityCall(arrow)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

/* TestArrowSecurityDetector_FunctionContainsSecurityCall verifies program-level function lookup */
func TestArrowSecurityDetector_FunctionContainsSecurityCall(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			makeArrowFunctionDecl("getHTFClose", []ast.Identifier{}, []ast.Node{
				&ast.ExpressionStatement{
					Expression: makeSecurityCall("syminfo.tickerid", "1D", "close"),
				},
			}),
			makeArrowFunctionDecl("calcSMA", []ast.Identifier{{Name: "src"}}, []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "ta"},
							Property: &ast.Identifier{Name: "sma"},
						},
						Arguments: []ast.Expression{
							&ast.Identifier{Name: "src"},
							&ast.Literal{Value: float64(20)},
						},
					},
				},
			}),
		},
	}

	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"security_function", "getHTFClose", true},
		{"non_security_function", "calcSMA", false},
		{"nonexistent_function", "doesNotExist", false},
	}

	detector := NewArrowSecurityDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.FunctionContainsSecurityCall(tt.funcName, program)
			if result != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.funcName, result)
			}
		})
	}
}

/* TestArrowSecurityDetector_FunctionContainsSecurityCall_EdgeCases verifies boundary conditions */
func TestArrowSecurityDetector_FunctionContainsSecurityCall_EdgeCases(t *testing.T) {
	detector := NewArrowSecurityDetector()

	tests := []struct {
		name     string
		funcName string
		program  *ast.Program
		expected bool
	}{
		{
			name:     "nil_program",
			funcName: "test",
			program:  nil,
			expected: false,
		},
		{
			name:     "empty_program",
			funcName: "test",
			program:  &ast.Program{Body: []ast.Node{}},
			expected: false,
		},
		{
			name:     "non_arrow_variable",
			funcName: "x",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{ID: &ast.Identifier{Name: "x"}, Init: &ast.Literal{Value: 42.0}},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("unexpected panic: %v", r)
				}
			}()

			result := detector.FunctionContainsSecurityCall(tt.funcName, tt.program)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func makeSecurityCall(symbol, timeframe, expr string) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "request"},
			Property: &ast.Identifier{Name: "security"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: symbol},
			&ast.Literal{Value: timeframe},
			&ast.Identifier{Name: expr},
		},
	}
}

func makeArrowFunctionDecl(name string, params []ast.Identifier, body []ast.Node) *ast.VariableDeclaration {
	return &ast.VariableDeclaration{
		Declarations: []ast.VariableDeclarator{
			{
				ID: &ast.Identifier{Name: name},
				Init: &ast.ArrowFunctionExpression{
					Params: params,
					Body:   body,
				},
			},
		},
	}
}
