package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestLoopNestingValidator_ValidCases validates accepted break/continue placements via AST */
func TestLoopNestingValidator_ValidCases(t *testing.T) {
	validator := NewLoopNestingValidator()

	tests := []struct {
		name    string
		program *ast.Program
	}{
		{
			name: "break inside for",
			program: &ast.Program{Body: []ast.Node{
				&ast.ForStatement{NodeType: ast.TypeForStatement, Counter: "i",
					From: &ast.Literal{Value: 0}, To: &ast.Literal{Value: 9},
					Body: []ast.Node{&ast.BreakStatement{NodeType: ast.TypeBreakStatement}}},
			}},
		},
		{
			name: "continue inside for",
			program: &ast.Program{Body: []ast.Node{
				&ast.ForStatement{NodeType: ast.TypeForStatement, Counter: "i",
					From: &ast.Literal{Value: 0}, To: &ast.Literal{Value: 9},
					Body: []ast.Node{&ast.ContinueStatement{NodeType: ast.TypeContinueStatement}}},
			}},
		},
		{
			name: "break inside for-in",
			program: &ast.Program{Body: []ast.Node{
				&ast.ForInStatement{NodeType: ast.TypeForInStatement, ElementVar: "val",
					Collection: &ast.Identifier{Name: "close"},
					Body:       []ast.Node{&ast.BreakStatement{NodeType: ast.TypeBreakStatement}}},
			}},
		},
		{
			name: "continue inside for-in",
			program: &ast.Program{Body: []ast.Node{
				&ast.ForInStatement{NodeType: ast.TypeForInStatement, ElementVar: "val",
					Collection: &ast.Identifier{Name: "close"},
					Body:       []ast.Node{&ast.ContinueStatement{NodeType: ast.TypeContinueStatement}}},
			}},
		},
		{
			name: "break inside if inside for",
			program: &ast.Program{Body: []ast.Node{
				&ast.ForStatement{NodeType: ast.TypeForStatement, Counter: "i",
					From: &ast.Literal{Value: 0}, To: &ast.Literal{Value: 9},
					Body: []ast.Node{
						&ast.IfStatement{NodeType: ast.TypeIfStatement,
							Test:       &ast.Identifier{Name: "cond"},
							Consequent: []ast.Node{&ast.BreakStatement{NodeType: ast.TypeBreakStatement}}},
					}},
			}},
		},
		{
			name: "continue inside if-else inside for-in",
			program: &ast.Program{Body: []ast.Node{
				&ast.ForInStatement{NodeType: ast.TypeForInStatement, ElementVar: "val",
					Collection: &ast.Identifier{Name: "close"},
					Body: []ast.Node{
						&ast.IfStatement{NodeType: ast.TypeIfStatement,
							Test:       &ast.Identifier{Name: "cond"},
							Consequent: []ast.Node{&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "x"}}},
							Alternate:  []ast.Node{&ast.ContinueStatement{NodeType: ast.TypeContinueStatement}}},
					}},
			}},
		},
		{
			name: "break inside nested for loops",
			program: &ast.Program{Body: []ast.Node{
				&ast.ForStatement{NodeType: ast.TypeForStatement, Counter: "i",
					From: &ast.Literal{Value: 0}, To: &ast.Literal{Value: 5},
					Body: []ast.Node{
						&ast.ForStatement{NodeType: ast.TypeForStatement, Counter: "j",
							From: &ast.Literal{Value: 0}, To: &ast.Literal{Value: 5},
							Body: []ast.Node{&ast.BreakStatement{NodeType: ast.TypeBreakStatement}}},
					}},
			}},
		},
		{
			name: "break inside for-in inside for",
			program: &ast.Program{Body: []ast.Node{
				&ast.ForStatement{NodeType: ast.TypeForStatement, Counter: "i",
					From: &ast.Literal{Value: 0}, To: &ast.Literal{Value: 5},
					Body: []ast.Node{
						&ast.ForInStatement{NodeType: ast.TypeForInStatement, ElementVar: "val",
							Collection: &ast.Identifier{Name: "arr"},
							Body:       []ast.Node{&ast.BreakStatement{NodeType: ast.TypeBreakStatement}}},
					}},
			}},
		},
		{
			name: "break inside for inside arrow inside for",
			program: &ast.Program{Body: []ast.Node{
				&ast.ForStatement{NodeType: ast.TypeForStatement, Counter: "i",
					From: &ast.Literal{Value: 0}, To: &ast.Literal{Value: 9},
					Body: []ast.Node{
						&ast.ExpressionStatement{Expression: &ast.ArrowFunctionExpression{
							NodeType: ast.TypeArrowFunctionExpression,
							Body: []ast.Node{
								&ast.ForStatement{NodeType: ast.TypeForStatement, Counter: "j",
									From: &ast.Literal{Value: 0}, To: &ast.Literal{Value: 3},
									Body: []ast.Node{&ast.BreakStatement{NodeType: ast.TypeBreakStatement}}},
							}}},
					}},
			}},
		},
		{
			name: "break in expression-level for via variable init",
			program: &ast.Program{Body: []ast.Node{
				&ast.VariableDeclaration{NodeType: ast.TypeVariableDeclaration,
					Declarations: []ast.VariableDeclarator{{
						ID: &ast.Identifier{Name: "x"},
						Init: &ast.ForStatement{NodeType: ast.TypeForStatement, Counter: "i",
							From: &ast.Literal{Value: 0}, To: &ast.Literal{Value: 9},
							Body: []ast.Node{&ast.BreakStatement{NodeType: ast.TypeBreakStatement}}},
					}}},
			}},
		},
		{
			name: "break in expression-level for-in via variable init",
			program: &ast.Program{Body: []ast.Node{
				&ast.VariableDeclaration{NodeType: ast.TypeVariableDeclaration,
					Declarations: []ast.VariableDeclarator{{
						ID: &ast.Identifier{Name: "x"},
						Init: &ast.ForInStatement{NodeType: ast.TypeForInStatement, ElementVar: "val",
							Collection: &ast.Identifier{Name: "arr"},
							Body:       []ast.Node{&ast.BreakStatement{NodeType: ast.TypeBreakStatement}}},
					}}},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validator.Validate(tt.program); err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
		})
	}
}

/* TestLoopNestingValidator_InvalidCases validates rejected break/continue placements via AST */
func TestLoopNestingValidator_InvalidCases(t *testing.T) {
	validator := NewLoopNestingValidator()

	tests := []struct {
		name        string
		program     *ast.Program
		errContains string
	}{
		{
			name: "break at top level",
			program: &ast.Program{Body: []ast.Node{
				&ast.BreakStatement{NodeType: ast.TypeBreakStatement},
			}},
			errContains: "break statement outside loop body",
		},
		{
			name: "continue at top level",
			program: &ast.Program{Body: []ast.Node{
				&ast.ContinueStatement{NodeType: ast.TypeContinueStatement},
			}},
			errContains: "continue statement outside loop body",
		},
		{
			name: "break inside if without loop",
			program: &ast.Program{Body: []ast.Node{
				&ast.IfStatement{NodeType: ast.TypeIfStatement,
					Test:       &ast.Identifier{Name: "cond"},
					Consequent: []ast.Node{&ast.BreakStatement{NodeType: ast.TypeBreakStatement}}},
			}},
			errContains: "break statement outside loop body",
		},
		{
			name: "continue inside if-else without loop",
			program: &ast.Program{Body: []ast.Node{
				&ast.IfStatement{NodeType: ast.TypeIfStatement,
					Test:       &ast.Identifier{Name: "cond"},
					Consequent: []ast.Node{&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "x"}}},
					Alternate:  []ast.Node{&ast.ContinueStatement{NodeType: ast.TypeContinueStatement}}},
			}},
			errContains: "continue statement outside loop body",
		},
		{
			name: "break inside arrow without loop resets context",
			program: &ast.Program{Body: []ast.Node{
				&ast.ForStatement{NodeType: ast.TypeForStatement, Counter: "i",
					From: &ast.Literal{Value: 0}, To: &ast.Literal{Value: 9},
					Body: []ast.Node{
						&ast.ExpressionStatement{Expression: &ast.ArrowFunctionExpression{
							NodeType: ast.TypeArrowFunctionExpression,
							Body:     []ast.Node{&ast.BreakStatement{NodeType: ast.TypeBreakStatement}}}},
					}},
			}},
			errContains: "break statement outside loop body",
		},
		{
			name: "continue inside arrow without loop resets context",
			program: &ast.Program{Body: []ast.Node{
				&ast.ForInStatement{NodeType: ast.TypeForInStatement, ElementVar: "val",
					Collection: &ast.Identifier{Name: "close"},
					Body: []ast.Node{
						&ast.ExpressionStatement{Expression: &ast.ArrowFunctionExpression{
							NodeType: ast.TypeArrowFunctionExpression,
							Body:     []ast.Node{&ast.ContinueStatement{NodeType: ast.TypeContinueStatement}}}},
					}},
			}},
			errContains: "continue statement outside loop body",
		},
		{
			name: "break in nested arrow without inner loop",
			program: &ast.Program{Body: []ast.Node{
				&ast.ExpressionStatement{Expression: &ast.ArrowFunctionExpression{
					NodeType: ast.TypeArrowFunctionExpression,
					Body: []ast.Node{
						&ast.ExpressionStatement{Expression: &ast.ArrowFunctionExpression{
							NodeType: ast.TypeArrowFunctionExpression,
							Body:     []ast.Node{&ast.BreakStatement{NodeType: ast.TypeBreakStatement}}}},
					}}},
			}},
			errContains: "break statement outside loop body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.program)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("expected error containing %q, got: %v", tt.errContains, err)
			}
		})
	}
}

/* TestLoopNestingValidator_Integration validates full Pine→AST→validator pipeline */
func TestLoopNestingValidator_Integration(t *testing.T) {
	tests := []struct {
		name        string
		pine        string
		shouldError bool
		errContains string
		description string
	}{
		{
			name: "break inside for passes validation",
			pine: `
//@version=5
strategy("Test")
for i = 0 to 10
    if i > 5
        break
`,
			shouldError: false,
			description: "break inside for accepted by validator in full pipeline",
		},
		{
			name: "continue inside for-in passes validation",
			pine: `
//@version=5
strategy("Test")
for val in close
    if val < 0
        continue
`,
			shouldError: false,
			description: "continue inside for-in accepted by validator in full pipeline",
		},
		{
			name: "break inside expression-level for passes validation",
			pine: `
//@version=5
indicator("Test")
x = for i = 0 to 9
    if i > 5
        break
    i
plot(x)
`,
			shouldError: false,
			description: "break inside IIFE for-expression accepted by validator",
		},
		{
			name: "top-level break rejected",
			pine: `
//@version=5
strategy("Test")
break
`,
			shouldError: true,
			errContains: "break statement outside loop body",
			description: "top-level break caught by validator through full pipeline",
		},
		{
			name: "top-level continue rejected",
			pine: `
//@version=5
strategy("Test")
continue
`,
			shouldError: true,
			errContains: "continue statement outside loop body",
			description: "top-level continue caught by validator through full pipeline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := compilePineScript(tt.pine)
			if tt.shouldError {
				if err == nil {
					t.Fatalf("Expected error but compilation succeeded\nDescription: %s", tt.description)
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got: %v\nDescription: %s",
						tt.errContains, err, tt.description)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected success but got error: %v\nDescription: %s", err, tt.description)
				}
			}
		})
	}
}
