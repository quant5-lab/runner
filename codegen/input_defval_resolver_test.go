package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestIsInputCallExpr(t *testing.T) {
	tests := []struct {
		name     string
		call     *ast.CallExpression
		expected bool
	}{
		{
			name:     "direct input()",
			call:     &ast.CallExpression{Callee: &ast.Identifier{Name: "input"}},
			expected: true,
		},
		{
			name: "input.int()",
			call: &ast.CallExpression{Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "input"},
				Property: &ast.Identifier{Name: "int"},
			}},
			expected: true,
		},
		{
			name: "input.float()",
			call: &ast.CallExpression{Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "input"},
				Property: &ast.Identifier{Name: "float"},
			}},
			expected: true,
		},
		{
			name: "input.bool()",
			call: &ast.CallExpression{Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "input"},
				Property: &ast.Identifier{Name: "bool"},
			}},
			expected: true,
		},
		{
			name: "input.string()",
			call: &ast.CallExpression{Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "input"},
				Property: &ast.Identifier{Name: "string"},
			}},
			expected: true,
		},
		{
			name:     "ta()",
			call:     &ast.CallExpression{Callee: &ast.Identifier{Name: "ta"}},
			expected: false,
		},
		{
			name:     "math()",
			call:     &ast.CallExpression{Callee: &ast.Identifier{Name: "math"}},
			expected: false,
		},
		{
			name: "ta.sma()",
			call: &ast.CallExpression{Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "sma"},
			}},
			expected: false,
		},
		{
			name: "request.security()",
			call: &ast.CallExpression{Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "request"},
				Property: &ast.Identifier{Name: "security"},
			}},
			expected: false,
		},
		{
			name:     "empty name",
			call:     &ast.CallExpression{Callee: &ast.Identifier{Name: ""}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isInputCallExpr(tt.call); got != tt.expected {
				t.Errorf("isInputCallExpr() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExtractInputDefvalLiteral(t *testing.T) {
	tests := []struct {
		name     string
		call     *ast.CallExpression
		expected string
	}{
		{
			name: "positional string defval",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "input"},
				Arguments: []ast.Expression{&ast.Literal{Value: "1D"}},
			},
			expected: "1D",
		},
		{
			name: "named defval string",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "input"},
				Arguments: []ast.Expression{
					&ast.ObjectExpression{Properties: []ast.Property{
						{Key: &ast.Identifier{Name: "title"}, Value: &ast.Literal{Value: "Timeframe"}},
						{Key: &ast.Identifier{Name: "defval"}, Value: &ast.Literal{Value: "1D"}},
					}},
				},
			},
			expected: "1D",
		},
		{
			name: "named defval numeric",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "input"},
				Arguments: []ast.Expression{
					&ast.ObjectExpression{Properties: []ast.Property{
						{Key: &ast.Identifier{Name: "defval"}, Value: &ast.Literal{Value: 14.0}},
					}},
				},
			},
			expected: "",
		},
		{
			name: "named defval boolean",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "input"},
				Arguments: []ast.Expression{
					&ast.ObjectExpression{Properties: []ast.Property{
						{Key: &ast.Identifier{Name: "defval"}, Value: &ast.Literal{Value: true}},
					}},
				},
			},
			expected: "",
		},
		{
			name: "named defval identifier",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "input"},
				Arguments: []ast.Expression{
					&ast.ObjectExpression{Properties: []ast.Property{
						{Key: &ast.Identifier{Name: "defval"}, Value: &ast.Identifier{Name: "someVar"}},
					}},
				},
			},
			expected: "",
		},
		{
			name: "positional numeric defval",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "input"},
				Arguments: []ast.Expression{&ast.Literal{Value: 14.0}},
			},
			expected: "",
		},
		{
			name: "no arguments",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "input"},
				Arguments: []ast.Expression{},
			},
			expected: "",
		},
		{
			name: "nil arguments",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "input"},
			},
			expected: "",
		},
		{
			name: "object without defval key",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "input"},
				Arguments: []ast.Expression{
					&ast.ObjectExpression{Properties: []ast.Property{
						{Key: &ast.Identifier{Name: "title"}, Value: &ast.Literal{Value: "Name"}},
						{Key: &ast.Identifier{Name: "type"}, Value: &ast.Identifier{Name: "resolution"}},
					}},
				},
			},
			expected: "",
		},
		{
			name: "empty string defval",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "input"},
				Arguments: []ast.Expression{&ast.Literal{Value: ""}},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractInputDefvalLiteral(tt.call); got != tt.expected {
				t.Errorf("extractInputDefvalLiteral() = %q, want %q", got, tt.expected)
			}
		})
	}
}
