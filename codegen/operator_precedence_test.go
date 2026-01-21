package codegen

import (
	"testing"
)

func TestGetOperatorPrecedence(t *testing.T) {
	tests := []struct {
		operator string
		expected OperatorPrecedence
	}{
		{"+", PrecAdditive},
		{"-", PrecAdditive},
		{"*", PrecMultiplicative},
		{"/", PrecMultiplicative},
		{"%", PrecMultiplicative},
		{"<", PrecComparison},
		{"==", PrecEquality},
		{"&&", PrecLogicalAnd},
		{"||", PrecLogicalOr},
	}

	for _, tt := range tests {
		t.Run(tt.operator, func(t *testing.T) {
			got := GetOperatorPrecedence(tt.operator)
			if got != tt.expected {
				t.Errorf("GetOperatorPrecedence(%q) = %v, want %v", tt.operator, got, tt.expected)
			}
		})
	}
}

func TestNeedsParentheses(t *testing.T) {
	tests := []struct {
		name         string
		childOp      string
		parentOp     string
		isRightChild bool
		expected     bool
		description  string
	}{
		{
			name:         "addition child, multiplication parent",
			childOp:      "+",
			parentOp:     "*",
			isRightChild: false,
			expected:     true,
			description:  "lower precedence child needs parens",
		},
		{
			name:         "multiplication child, addition parent",
			childOp:      "*",
			parentOp:     "+",
			isRightChild: false,
			expected:     false,
			description:  "higher precedence child no parens needed",
		},
		{
			name:         "same precedence left child",
			childOp:      "+",
			parentOp:     "-",
			isRightChild: false,
			expected:     false,
			description:  "left-associative: left child no parens",
		},
		{
			name:         "same precedence right child",
			childOp:      "+",
			parentOp:     "-",
			isRightChild: true,
			expected:     true,
			description:  "left-associative: right child needs parens",
		},
		{
			name:         "division left, multiplication parent",
			childOp:      "/",
			parentOp:     "*",
			isRightChild: false,
			expected:     false,
			description:  "same precedence left child no parens",
		},
		{
			name:         "division right, multiplication parent",
			childOp:      "/",
			parentOp:     "*",
			isRightChild: true,
			expected:     true,
			description:  "same precedence right child needs parens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NeedsParentheses(tt.childOp, tt.parentOp, tt.isRightChild)
			if got != tt.expected {
				t.Errorf("%s\nNeedsParentheses(%q, %q, %v) = %v, want %v",
					tt.description, tt.childOp, tt.parentOp, tt.isRightChild, got, tt.expected)
			}
		})
	}
}
