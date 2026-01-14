package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestArrowFunctionVariableInit_PreambleSeparation(t *testing.T) {
	tests := []struct {
		name              string
		varName           string
		initExpr          ast.Expression
		expectPreamble    bool
		preamblePattern   string
		assignmentPattern string
		description       string
	}{
		{
			name:    "simple identifier - no preamble",
			varName: "result",
			initExpr: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     "source",
			},
			expectPreamble:    false,
			assignmentPattern: "result := source",
			description:       "Direct identifier assignment without preamble",
		},
		{
			name:    "literal value - no preamble",
			varName: "constant",
			initExpr: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    42.0,
			},
			expectPreamble:    false,
			assignmentPattern: "constant := 42",
			description:       "Literal value assignment without preamble",
		},
		{
			name:    "conditional expression - no preamble",
			varName: "choice",
			initExpr: &ast.ConditionalExpression{
				NodeType: ast.TypeConditionalExpression,
				Test: &ast.Identifier{
					NodeType: ast.TypeIdentifier,
					Name:     "condition",
				},
				Consequent: &ast.Literal{
					NodeType: ast.TypeLiteral,
					Value:    1.0,
				},
				Alternate: &ast.Literal{
					NodeType: ast.TypeLiteral,
					Value:    0.0,
				},
			},
			expectPreamble:    false,
			assignmentPattern: "choice := func() float64",
			description:       "Conditional expression generates IIFE without preamble",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createTestGenerator()

			result, err := gen.generateArrowFunctionVariableInit(tt.varName, tt.initExpr)
			if err != nil {
				t.Fatalf("[%s] Unexpected error: %v", tt.description, err)
			}

			if result == nil {
				t.Fatalf("[%s] Result is nil", tt.description)
			}

			if tt.expectPreamble && !result.HasPreamble() {
				t.Errorf("[%s] Expected preamble but none found", tt.description)
			}

			if !tt.expectPreamble && result.HasPreamble() {
				t.Errorf("[%s] Unexpected preamble: %s", tt.description, result.Preamble)
			}

			if tt.preamblePattern != "" && !strings.Contains(result.Preamble, tt.preamblePattern) {
				t.Errorf("[%s] Preamble missing pattern %q\nGot: %s",
					tt.description, tt.preamblePattern, result.Preamble)
			}

			if tt.assignmentPattern != "" && !strings.Contains(result.Assignment, tt.assignmentPattern) {
				t.Errorf("[%s] Assignment missing pattern %q\nGot: %s",
					tt.description, tt.assignmentPattern, result.Assignment)
			}

			combined := result.CombinedCode()
			if tt.expectPreamble {
				preambleIdx := strings.Index(combined, result.Preamble)
				assignmentIdx := strings.Index(combined, result.Assignment)
				if preambleIdx < 0 || assignmentIdx < 0 {
					t.Errorf("[%s] Missing preamble or assignment in combined code", tt.description)
				} else if preambleIdx > assignmentIdx {
					t.Errorf("[%s] Preamble must appear before assignment", tt.description)
				}
			}
		})
	}
}

func TestArrowFunctionVariableInit_NestedPreambles(t *testing.T) {
	t.Run("nested call expressions accumulate preambles", func(t *testing.T) {
		gen := createTestGenerator()

		nestedCall := &ast.CallExpression{
			NodeType: ast.TypeCallExpression,
			Callee: &ast.MemberExpression{
				NodeType: ast.TypeMemberExpression,
				Object: &ast.Identifier{
					NodeType: ast.TypeIdentifier,
					Name:     "ta",
				},
				Property: &ast.Identifier{
					NodeType: ast.TypeIdentifier,
					Name:     "sma",
				},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{
					NodeType: ast.TypeIdentifier,
					Name:     "close",
				},
				&ast.Literal{
					NodeType: ast.TypeLiteral,
					Value:    20.0,
				},
			},
		}

		result, err := gen.generateArrowFunctionVariableInit("average", nestedCall)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Assignment == "" {
			t.Error("Expected assignment code")
		}
	})
}

func TestArrowFunctionVariableInit_FormatConsistency(t *testing.T) {
	tests := []struct {
		name        string
		varName     string
		initExpr    ast.Expression
		checkFormat func(result *ArrowVarInitResult) error
	}{
		{
			name:    "assignment has proper indentation",
			varName: "value",
			initExpr: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    42.0,
			},
			checkFormat: func(result *ArrowVarInitResult) error {
				if !strings.HasPrefix(result.Assignment, "\t") {
					return &testError{"Assignment should start with tab indentation"}
				}
				return nil
			},
		},
		{
			name:    "assignment ends with newline",
			varName: "value",
			initExpr: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    42.0,
			},
			checkFormat: func(result *ArrowVarInitResult) error {
				if !strings.HasSuffix(result.Assignment, "\n") {
					return &testError{"Assignment should end with newline"}
				}
				return nil
			},
		},
		{
			name:    "preamble has proper format when present",
			varName: "value",
			initExpr: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     "source",
			},
			checkFormat: func(result *ArrowVarInitResult) error {
				if result.HasPreamble() && !strings.HasSuffix(result.Preamble, "\n") {
					return &testError{"Preamble should end with newline when present"}
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createTestGenerator()

			result, err := gen.generateArrowFunctionVariableInit(tt.varName, tt.initExpr)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if err := tt.checkFormat(result); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestArrowFunctionVariableInit_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		varName     string
		initExpr    ast.Expression
		expectError bool
		description string
	}{
		{
			name:    "empty variable name",
			varName: "",
			initExpr: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    42.0,
			},
			expectError: false,
			description: "Should handle empty variable name",
		},
		{
			name:        "nil expression",
			varName:     "test",
			initExpr:    nil,
			expectError: true,
			description: "Should handle nil expression gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createTestGenerator()

			result, err := gen.generateArrowFunctionVariableInit(tt.varName, tt.initExpr)

			if tt.expectError {
				if err == nil {
					t.Errorf("[%s] Expected error but got none", tt.description)
				}
				if result != nil {
					t.Errorf("[%s] Expected nil result on error", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("[%s] Unexpected error: %v", tt.description, err)
				}
				if result == nil {
					t.Errorf("[%s] Expected non-nil result", tt.description)
				}
			}
		})
	}
}

type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
