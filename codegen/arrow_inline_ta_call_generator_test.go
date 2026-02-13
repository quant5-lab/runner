package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestArrowInlineTACallGenerator_PeriodDelegation verifies inline IIFE gracefully delegates non-constant periods */
func TestArrowInlineTACallGenerator_PeriodDelegation(t *testing.T) {
	tests := []struct {
		name           string
		expr           ast.Expression
		expectedPeriod int
		description    string
	}{
		{
			name:           "Integer literal returns value",
			expr:           &ast.Literal{Value: int(14)},
			expectedPeriod: 14,
			description:    "Compile-time constant integer",
		},
		{
			name:           "Float literal returns truncated value",
			expr:           &ast.Literal{Value: 20.0},
			expectedPeriod: 20,
			description:    "Compile-time constant float",
		},
		{
			name:           "String numeric literal returns parsed value",
			expr:           &ast.Literal{Value: "50"},
			expectedPeriod: 50,
			description:    "Compile-time constant string",
		},
		{
			name:           "Identifier signals delegation",
			expr:           &ast.Identifier{Name: "len"},
			expectedPeriod: 0,
			description:    "Runtime parameter — IIFE cannot handle",
		},
		{
			name: "Binary expression signals delegation",
			expr: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "len"},
				Operator: "/",
				Right:    &ast.Literal{Value: 2.0},
			},
			expectedPeriod: 0,
			description:    "Computed expression — IIFE cannot handle",
		},
		{
			name: "Call expression signals delegation",
			expr: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "math.round"},
				Arguments: []ast.Expression{&ast.Identifier{Name: "len"}},
			},
			expectedPeriod: 0,
			description:    "Nested function call — IIFE cannot handle",
		},
		{
			name: "Unary expression signals delegation",
			expr: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.Identifier{Name: "len"},
			},
			expectedPeriod: 0,
			description:    "Unary expression — IIFE cannot handle",
		},
		{
			name: "Member expression signals delegation",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "arr"},
				Property: &ast.Identifier{Name: "length"},
			},
			expectedPeriod: 0,
			description:    "Member access — IIFE cannot handle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &ArrowInlineTACallGenerator{}

			period, err := gen.extractPeriod(tt.expr)

			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tt.description, err)
			}

			if period != tt.expectedPeriod {
				t.Errorf("%s: extractPeriod() = %d, want %d", tt.description, period, tt.expectedPeriod)
			}
		})
	}
}

/* TestArrowInlineTACallGenerator_NonNumericLiteralError verifies non-numeric literals are not silently accepted */
func TestArrowInlineTACallGenerator_NonNumericLiteralError(t *testing.T) {
	tests := []struct {
		name string
		expr ast.Expression
	}{
		{
			name: "Non-numeric string literal",
			expr: &ast.Literal{Value: "abc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &ArrowInlineTACallGenerator{}

			_, err := gen.extractPeriod(tt.expr)

			if err == nil {
				t.Error("Expected error for non-numeric string literal")
			}
		})
	}
}
