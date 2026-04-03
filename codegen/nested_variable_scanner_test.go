package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/*
TestNestedVariableScanner_ScopePromotion validates PineScript scope boundary behavior:

	if-block variables promote to function scope, loop variables remain loop-local.
*/
func TestNestedVariableScanner_ScopePromotion(t *testing.T) {
	tests := []struct {
		name        string
		ifStmt      *ast.IfStatement
		wantVars    []string
		wantAbsent  []string
		description string
	}{
		{
			name: "consequent block variable",
			ifStmt: &ast.IfStatement{
				Test: &ast.Literal{Value: true},
				Consequent: []ast.Node{
					varDecl("x", &ast.Literal{Value: 1.0}),
				},
			},
			wantVars:    []string{"x"},
			description: "variable in if-body promoted to function scope",
		},
		{
			name: "alternate block variable",
			ifStmt: &ast.IfStatement{
				Test:       &ast.Literal{Value: true},
				Consequent: []ast.Node{},
				Alternate: []ast.Node{
					varDecl("y", &ast.Literal{Value: 2.0}),
				},
			},
			wantVars:    []string{"y"},
			description: "variable in else-body promoted to function scope",
		},
		{
			name: "both branches",
			ifStmt: &ast.IfStatement{
				Test: &ast.Literal{Value: true},
				Consequent: []ast.Node{
					varDecl("a", &ast.Literal{Value: 1.0}),
				},
				Alternate: []ast.Node{
					varDecl("b", &ast.Literal{Value: 2.0}),
				},
			},
			wantVars:    []string{"a", "b"},
			description: "variables in both if and else branches promoted",
		},
		{
			name: "nested if-blocks recurse",
			ifStmt: &ast.IfStatement{
				Test: &ast.Literal{Value: true},
				Consequent: []ast.Node{
					&ast.IfStatement{
						Test: &ast.Literal{Value: false},
						Consequent: []ast.Node{
							varDecl("deep", &ast.Literal{Value: 99.0}),
						},
					},
				},
			},
			wantVars:    []string{"deep"},
			description: "nested if-block variables promoted through recursion",
		},
		{
			name: "if-else-if chain",
			ifStmt: &ast.IfStatement{
				Test: &ast.Literal{Value: true},
				Consequent: []ast.Node{
					varDecl("first", &ast.Literal{Value: 1.0}),
				},
				Alternate: []ast.Node{
					&ast.IfStatement{
						Test: &ast.Literal{Value: false},
						Consequent: []ast.Node{
							varDecl("second", &ast.Literal{Value: 2.0}),
						},
						Alternate: []ast.Node{
							varDecl("third", &ast.Literal{Value: 3.0}),
						},
					},
				},
			},
			wantVars:    []string{"first", "second", "third"},
			description: "all branches of if-else-if chain promoted",
		},
		{
			name: "multiple variables in same block",
			ifStmt: &ast.IfStatement{
				Test: &ast.Literal{Value: true},
				Consequent: []ast.Node{
					varDecl("m", &ast.Literal{Value: 1.0}),
					varDecl("n", &ast.Literal{Value: 2.0}),
				},
			},
			wantVars:    []string{"m", "n"},
			description: "all declarations within same if-block promoted",
		},
		{
			name: "empty blocks produce no variables",
			ifStmt: &ast.IfStatement{
				Test:       &ast.Literal{Value: true},
				Consequent: []ast.Node{},
				Alternate:  []ast.Node{},
			},
			wantVars:    nil,
			description: "empty if/else blocks register nothing",
		},
		{
			name: "for-loop inside if excluded",
			ifStmt: &ast.IfStatement{
				Test: &ast.Literal{Value: true},
				Consequent: []ast.Node{
					&ast.ForStatement{
						Counter: "i",
						From:    &ast.Literal{Value: 0},
						To:      &ast.Literal{Value: 10},
						Body: []ast.Node{
							varDecl("loopVar", &ast.Literal{Value: 1.0}),
						},
					},
				},
			},
			wantAbsent:  []string{"loopVar"},
			description: "for-loop body variables are loop-local, not promoted",
		},
		{
			name: "for-in loop inside if excluded",
			ifStmt: &ast.IfStatement{
				Test: &ast.Literal{Value: true},
				Consequent: []ast.Node{
					&ast.ForInStatement{
						ElementVar: "val",
						Collection: &ast.Identifier{Name: "arr"},
						Body: []ast.Node{
							varDecl("forInVar", &ast.Literal{Value: 1.0}),
						},
					},
				},
			},
			wantAbsent:  []string{"forInVar"},
			description: "for-in loop body variables are loop-local, not promoted",
		},
		{
			name: "while-loop inside if excluded",
			ifStmt: &ast.IfStatement{
				Test: &ast.Literal{Value: true},
				Consequent: []ast.Node{
					&ast.WhileStatement{
						Condition: &ast.Literal{Value: true},
						Body: []ast.Node{
							varDecl("whileVar", &ast.Literal{Value: 1.0}),
						},
					},
				},
			},
			wantAbsent:  []string{"whileVar"},
			description: "while-loop body variables are loop-local, not promoted",
		},
		{
			name: "if after loop inside if — only if-variable promoted",
			ifStmt: &ast.IfStatement{
				Test: &ast.Literal{Value: true},
				Consequent: []ast.Node{
					&ast.ForStatement{
						Counter: "i",
						From:    &ast.Literal{Value: 0},
						To:      &ast.Literal{Value: 5},
						Body: []ast.Node{
							varDecl("loopOnly", &ast.Literal{Value: 0.0}),
						},
					},
					varDecl("promoted", &ast.Literal{Value: 1.0}),
				},
			},
			wantVars:    []string{"promoted"},
			wantAbsent:  []string{"loopOnly"},
			description: "loop variable excluded while sibling if-block variable promoted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newDeclarationTestGenerator()
			scanner := NewNestedVariableScanner(gen)

			scanner.ScanIfBlock(tt.ifStmt)

			for _, name := range tt.wantVars {
				if _, ok := gen.variables[name]; !ok {
					t.Errorf("expected %q promoted to scope (%s)", name, tt.description)
				}
			}
			for _, name := range tt.wantAbsent {
				if _, ok := gen.variables[name]; ok {
					t.Errorf("expected %q to remain loop-local (%s)", name, tt.description)
				}
			}
		})
	}
}

/* TestNestedVariableScanner_ReassignmentTracking validates var-keyword reassignment discovery across all control flow */
func TestNestedVariableScanner_ReassignmentTracking(t *testing.T) {
	tests := []struct {
		name        string
		stmt        ast.Node
		wantTracked []string
		wantAbsent  []string
		description string
	}{
		{
			name: "var keyword in if-block",
			stmt: &ast.IfStatement{
				Test: &ast.Literal{Value: true},
				Consequent: []ast.Node{
					&ast.VariableDeclaration{
						Kind:         "var",
						Declarations: []ast.VariableDeclarator{{ID: &ast.Identifier{Name: "counter"}, Init: &ast.Literal{Value: 0.0}}},
					},
				},
			},
			wantTracked: []string{"counter"},
			description: "var keyword inside if-block tracked as reassignment",
		},
		{
			name: "let keyword ignored",
			stmt: &ast.IfStatement{
				Test: &ast.Literal{Value: true},
				Consequent: []ast.Node{
					&ast.VariableDeclaration{
						Kind:         "let",
						Declarations: []ast.VariableDeclarator{{ID: &ast.Identifier{Name: "temp"}, Init: &ast.Literal{Value: 0.0}}},
					},
				},
			},
			wantAbsent:  []string{"temp"},
			description: "let keyword declarations not tracked as reassignments",
		},
		{
			name: "var in for-loop body",
			stmt: &ast.ForStatement{
				Counter: "i",
				From:    &ast.Literal{Value: 0},
				To:      &ast.Literal{Value: 10},
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Kind:         "var",
						Declarations: []ast.VariableDeclarator{{ID: &ast.Identifier{Name: "sum"}, Init: &ast.Literal{Value: 0.0}}},
					},
				},
			},
			wantTracked: []string{"sum"},
			description: "var keyword inside for-loop body tracked for reassignment",
		},
		{
			name: "var in for-in body",
			stmt: &ast.ForInStatement{
				ElementVar: "val",
				Collection: &ast.Identifier{Name: "arr"},
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Kind:         "var",
						Declarations: []ast.VariableDeclarator{{ID: &ast.Identifier{Name: "acc"}, Init: &ast.Literal{Value: 0.0}}},
					},
				},
			},
			wantTracked: []string{"acc"},
			description: "var keyword inside for-in body tracked for reassignment",
		},
		{
			name: "var in while body",
			stmt: &ast.WhileStatement{
				Condition: &ast.Literal{Value: true},
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Kind:         "var",
						Declarations: []ast.VariableDeclarator{{ID: &ast.Identifier{Name: "state"}, Init: &ast.Literal{Value: 0.0}}},
					},
				},
			},
			wantTracked: []string{"state"},
			description: "var keyword inside while body tracked for reassignment",
		},
		{
			name: "deep nesting: var in if inside for",
			stmt: &ast.ForStatement{
				Counter: "i",
				From:    &ast.Literal{Value: 0},
				To:      &ast.Literal{Value: 10},
				Body: []ast.Node{
					&ast.IfStatement{
						Test: &ast.Literal{Value: true},
						Consequent: []ast.Node{
							&ast.VariableDeclaration{
								Kind:         "var",
								Declarations: []ast.VariableDeclarator{{ID: &ast.Identifier{Name: "nested"}, Init: &ast.Literal{Value: 0.0}}},
							},
						},
					},
				},
			},
			wantTracked: []string{"nested"},
			description: "var keyword found through deep nesting across different control flow types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newDeclarationTestGenerator()
			scanner := NewNestedVariableScanner(gen)

			scanner.ScanReassignments(tt.stmt)

			for _, name := range tt.wantTracked {
				if !gen.reassignedVars[name] {
					t.Errorf("expected %q tracked as reassignment (%s)", name, tt.description)
				}
			}
			for _, name := range tt.wantAbsent {
				if gen.reassignedVars[name] {
					t.Errorf("expected %q NOT tracked as reassignment (%s)", name, tt.description)
				}
			}
		})
	}
}

/* varDecl creates a single-declarator VariableDeclaration for test brevity */
func varDecl(name string, init ast.Expression) *ast.VariableDeclaration {
	return &ast.VariableDeclaration{
		Declarations: []ast.VariableDeclarator{
			{ID: &ast.Identifier{Name: name}, Init: init},
		},
	}
}
