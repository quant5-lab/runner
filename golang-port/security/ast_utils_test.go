package security

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestExtractCallFunctionName_ValidCases(t *testing.T) {
	tests := []struct {
		name     string
		callee   ast.Expression
		expected string
	}{
		{
			name: "member_expression_ta_sma",
			callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "sma"},
			},
			expected: "ta.sma",
		},
		{
			name: "member_expression_ta_ema",
			callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "ema"},
			},
			expected: "ta.ema",
		},
		{
			name: "member_expression_math_max",
			callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "math"},
				Property: &ast.Identifier{Name: "max"},
			},
			expected: "math.max",
		},
		{
			name:     "identifier_simple_function",
			callee:   &ast.Identifier{Name: "customFunc"},
			expected: "customFunc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractCallFunctionName(tt.callee)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractCallFunctionName_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		callee   ast.Expression
		expected string
	}{
		{
			name:     "nil_object",
			callee:   &ast.MemberExpression{Object: nil, Property: &ast.Identifier{Name: "func"}},
			expected: ".func",
		},
		{
			name:     "nil_property",
			callee:   &ast.MemberExpression{Object: &ast.Identifier{Name: "obj"}, Property: nil},
			expected: "obj.",
		},
		{
			name:     "literal_callee",
			callee:   &ast.Literal{Value: 42.0},
			expected: "",
		},
		{
			name:     "binary_expression_callee",
			callee:   &ast.BinaryExpression{Operator: "+"},
			expected: "",
		},
		{
			name:     "empty_object_name",
			callee:   &ast.MemberExpression{Object: &ast.Identifier{Name: ""}, Property: &ast.Identifier{Name: "func"}},
			expected: ".func",
		},
		{
			name:     "empty_property_name",
			callee:   &ast.MemberExpression{Object: &ast.Identifier{Name: "obj"}, Property: &ast.Identifier{Name: ""}},
			expected: "obj.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractCallFunctionName(tt.callee)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractNumberLiteral_ValidTypes(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		expected float64
	}{
		{
			name:     "float64_literal",
			expr:     &ast.Literal{Value: 42.5},
			expected: 42.5,
		},
		{
			name:     "int_literal",
			expr:     &ast.Literal{Value: 20},
			expected: 20.0,
		},
		{
			name:     "int64_literal",
			expr:     &ast.Literal{Value: int64(100)},
			expected: 100.0,
		},
		{
			name:     "zero_float",
			expr:     &ast.Literal{Value: 0.0},
			expected: 0.0,
		},
		{
			name:     "zero_int",
			expr:     &ast.Literal{Value: 0},
			expected: 0.0,
		},
		{
			name:     "negative_float",
			expr:     &ast.Literal{Value: -15.75},
			expected: -15.75,
		},
		{
			name:     "negative_int",
			expr:     &ast.Literal{Value: -50},
			expected: -50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractNumberLiteral(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %.2f, got %.2f", tt.expected, result)
			}
		})
	}
}

func TestExtractNumberLiteral_InvalidTypes(t *testing.T) {
	tests := []struct {
		name    string
		expr    ast.Expression
		wantErr bool
	}{
		{
			name:    "non_literal_identifier",
			expr:    &ast.Identifier{Name: "variable"},
			wantErr: true,
		},
		{
			name:    "string_literal",
			expr:    &ast.Literal{Value: "text"},
			wantErr: true,
		},
		{
			name:    "bool_literal",
			expr:    &ast.Literal{Value: true},
			wantErr: true,
		},
		{
			name:    "nil_literal_value",
			expr:    &ast.Literal{Value: nil},
			wantErr: true,
		},
		{
			name:    "binary_expression",
			expr:    &ast.BinaryExpression{Operator: "+"},
			wantErr: true,
		},
		{
			name:    "call_expression",
			expr:    &ast.CallExpression{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extractNumberLiteral(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got error=%v", tt.wantErr, err)
			}
		})
	}
}

func TestExtractNumberLiteral_BoundaryValues(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		expected float64
	}{
		{
			name:     "very_large_float",
			expr:     &ast.Literal{Value: 1e308},
			expected: 1e308,
		},
		{
			name:     "very_small_float",
			expr:     &ast.Literal{Value: 1e-308},
			expected: 1e-308,
		},
		{
			name:     "max_int",
			expr:     &ast.Literal{Value: int(2147483647)},
			expected: 2147483647.0,
		},
		{
			name:     "min_int",
			expr:     &ast.Literal{Value: int(-2147483648)},
			expected: -2147483648.0,
		},
		{
			name:     "fractional_precision",
			expr:     &ast.Literal{Value: 0.123456789},
			expected: 0.123456789,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractNumberLiteral(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %.15f, got %.15f", tt.expected, result)
			}
		})
	}
}
