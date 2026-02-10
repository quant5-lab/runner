package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestArrowLoopModificationAnalyzer_FindLoopModifiedVariables(t *testing.T) {
	tests := []struct {
		name     string
		body     []ast.Node
		expected map[string]bool
	}{
		{
			name: "variable modified in for-loop",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID:   &ast.Identifier{Name: "count"},
							Init: &ast.Literal{Value: 0.0},
						},
					},
				},
				&ast.ForStatement{
					Counter: "i",
					From:    &ast.Literal{Value: 0.0},
					To:      &ast.Literal{Value: 10.0},
					Body: []ast.Node{
						&ast.VariableDeclaration{
							Declarations: []ast.VariableDeclarator{
								{
									ID: &ast.Identifier{Name: "count"},
									Init: &ast.BinaryExpression{
										Operator: "+",
										Left:     &ast.Identifier{Name: "count"},
										Right:    &ast.Literal{Value: 1.0},
									},
								},
							},
						},
					},
				},
			},
			expected: map[string]bool{"count": true},
		},
		{
			name: "no loop modification",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID:   &ast.Identifier{Name: "result"},
							Init: &ast.Literal{Value: 0.0},
						},
					},
				},
			},
			expected: map[string]bool{},
		},
		{
			name: "loop-local variable not counted",
			body: []ast.Node{
				&ast.ForStatement{
					Counter: "i",
					From:    &ast.Literal{Value: 0.0},
					To:      &ast.Literal{Value: 10.0},
					Body: []ast.Node{
						&ast.VariableDeclaration{
							Declarations: []ast.VariableDeclarator{
								{
									ID:   &ast.Identifier{Name: "temp"},
									Init: &ast.Literal{Value: 0.0},
								},
							},
						},
					},
				},
			},
			expected: map[string]bool{},
		},
		{
			name: "multiple variables, one modified in loop",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID:   &ast.Identifier{Name: "sum"},
							Init: &ast.Literal{Value: 0.0},
						},
					},
				},
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID:   &ast.Identifier{Name: "result"},
							Init: &ast.Literal{Value: 0.0},
						},
					},
				},
				&ast.ForStatement{
					Counter: "i",
					From:    &ast.Literal{Value: 0.0},
					To:      &ast.Literal{Value: 10.0},
					Body: []ast.Node{
						&ast.VariableDeclaration{
							Declarations: []ast.VariableDeclarator{
								{
									ID: &ast.Identifier{Name: "sum"},
									Init: &ast.BinaryExpression{
										Operator: "+",
										Left:     &ast.Identifier{Name: "sum"},
										Right:    &ast.Literal{Value: 1.0},
									},
								},
							},
						},
					},
				},
			},
			expected: map[string]bool{"sum": true},
		},
		{
			name: "nested if inside loop",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID:   &ast.Identifier{Name: "count"},
							Init: &ast.Literal{Value: 0.0},
						},
					},
				},
				&ast.ForStatement{
					Counter: "i",
					From:    &ast.Literal{Value: 0.0},
					To:      &ast.Literal{Value: 10.0},
					Body: []ast.Node{
						&ast.IfStatement{
							Test: &ast.Literal{Value: true},
							Consequent: []ast.Node{
								&ast.VariableDeclaration{
									Declarations: []ast.VariableDeclarator{
										{
											ID:   &ast.Identifier{Name: "count"},
											Init: &ast.Literal{Value: 1.0},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: map[string]bool{"count": true},
		},
		{
			name: "variable modified in for-in loop",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID:   &ast.Identifier{Name: "total"},
							Init: &ast.Literal{Value: 0.0},
						},
					},
				},
				&ast.ForInStatement{
					ElementVar: "val",
					Collection: &ast.Identifier{Name: "arr"},
					Body: []ast.Node{
						&ast.VariableDeclaration{
							Declarations: []ast.VariableDeclarator{
								{
									ID: &ast.Identifier{Name: "total"},
									Init: &ast.BinaryExpression{
										Operator: "+",
										Left:     &ast.Identifier{Name: "total"},
										Right:    &ast.Identifier{Name: "val"},
									},
								},
							},
						},
					},
				},
			},
			expected: map[string]bool{"total": true},
		},
		{
			name: "nested for-in inside for loop",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID:   &ast.Identifier{Name: "sum"},
							Init: &ast.Literal{Value: 0.0},
						},
					},
				},
				&ast.ForStatement{
					Counter: "i",
					From:    &ast.Literal{Value: 0.0},
					To:      &ast.Literal{Value: 10.0},
					Body: []ast.Node{
						&ast.ForInStatement{
							ElementVar: "v",
							Collection: &ast.Identifier{Name: "data"},
							Body: []ast.Node{
								&ast.VariableDeclaration{
									Declarations: []ast.VariableDeclarator{
										{
											ID: &ast.Identifier{Name: "sum"},
											Init: &ast.BinaryExpression{
												Operator: "+",
												Left:     &ast.Identifier{Name: "sum"},
												Right:    &ast.Identifier{Name: "v"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: map[string]bool{"sum": true},
		},
		{
			name: "loop-local variable not counted in for-in",
			body: []ast.Node{
				&ast.ForInStatement{
					ElementVar: "val",
					Collection: &ast.Identifier{Name: "arr"},
					Body: []ast.Node{
						&ast.VariableDeclaration{
							Declarations: []ast.VariableDeclarator{
								{
									ID:   &ast.Identifier{Name: "temp"},
									Init: &ast.Literal{Value: 0.0},
								},
							},
						},
					},
				},
			},
			expected: map[string]bool{},
		},
		{
			name: "nested if inside for-in loop",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID:   &ast.Identifier{Name: "count"},
							Init: &ast.Literal{Value: 0.0},
						},
					},
				},
				&ast.ForInStatement{
					ElementVar: "val",
					Collection: &ast.Identifier{Name: "arr"},
					Body: []ast.Node{
						&ast.IfStatement{
							Test: &ast.Literal{Value: true},
							Consequent: []ast.Node{
								&ast.VariableDeclaration{
									Declarations: []ast.VariableDeclarator{
										{
											ID:   &ast.Identifier{Name: "count"},
											Init: &ast.Literal{Value: 1.0},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: map[string]bool{"count": true},
		},
		{
			name: "nested for inside for-in loop",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID:   &ast.Identifier{Name: "total"},
							Init: &ast.Literal{Value: 0.0},
						},
					},
				},
				&ast.ForInStatement{
					ElementVar: "val",
					Collection: &ast.Identifier{Name: "arr"},
					Body: []ast.Node{
						&ast.ForStatement{
							Counter: "j",
							From:    &ast.Literal{Value: 0.0},
							To:      &ast.Literal{Value: 5.0},
							Body: []ast.Node{
								&ast.VariableDeclaration{
									Declarations: []ast.VariableDeclarator{
										{
											ID: &ast.Identifier{Name: "total"},
											Init: &ast.BinaryExpression{
												Operator: "+",
												Left:     &ast.Identifier{Name: "total"},
												Right:    &ast.Literal{Value: 1.0},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: map[string]bool{"total": true},
		},
		{
			name: "nested for-in inside for-in loop",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID:   &ast.Identifier{Name: "sum"},
							Init: &ast.Literal{Value: 0.0},
						},
					},
				},
				&ast.ForInStatement{
					ElementVar: "a",
					Collection: &ast.Identifier{Name: "outer"},
					Body: []ast.Node{
						&ast.ForInStatement{
							ElementVar: "b",
							Collection: &ast.Identifier{Name: "inner"},
							Body: []ast.Node{
								&ast.VariableDeclaration{
									Declarations: []ast.VariableDeclarator{
										{
											ID: &ast.Identifier{Name: "sum"},
											Init: &ast.BinaryExpression{
												Operator: "+",
												Left:     &ast.Identifier{Name: "sum"},
												Right:    &ast.Identifier{Name: "b"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: map[string]bool{"sum": true},
		},
		{
			name: "multiple variables modified in for-in loop",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID:   &ast.Identifier{Name: "sum"},
							Init: &ast.Literal{Value: 0.0},
						},
					},
				},
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID:   &ast.Identifier{Name: "count"},
							Init: &ast.Literal{Value: 0.0},
						},
					},
				},
				&ast.ForInStatement{
					ElementVar: "val",
					Collection: &ast.Identifier{Name: "arr"},
					Body: []ast.Node{
						&ast.VariableDeclaration{
							Declarations: []ast.VariableDeclarator{
								{
									ID: &ast.Identifier{Name: "sum"},
									Init: &ast.BinaryExpression{
										Operator: "+",
										Left:     &ast.Identifier{Name: "sum"},
										Right:    &ast.Identifier{Name: "val"},
									},
								},
							},
						},
						&ast.VariableDeclaration{
							Declarations: []ast.VariableDeclarator{
								{
									ID: &ast.Identifier{Name: "count"},
									Init: &ast.BinaryExpression{
										Operator: "+",
										Left:     &ast.Identifier{Name: "count"},
										Right:    &ast.Literal{Value: 1.0},
									},
								},
							},
						},
					},
				},
			},
			expected: map[string]bool{"sum": true, "count": true},
		},
		{
			name: "ArrayPattern variable modified in nested loop",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID: &ast.ArrayPattern{
								Elements: []ast.Identifier{
									{Name: "high"},
									{Name: "low"},
								},
							},
							Init: &ast.Literal{Value: 0.0},
						},
					},
				},
				&ast.ForStatement{
					Counter: "i",
					From:    &ast.Literal{Value: 0.0},
					To:      &ast.Literal{Value: 5.0},
					Body: []ast.Node{
						&ast.ForInStatement{
							ElementVar: "val",
							Collection: &ast.Identifier{Name: "data"},
							Body: []ast.Node{
								&ast.VariableDeclaration{
									Declarations: []ast.VariableDeclarator{
										{
											ID:   &ast.Identifier{Name: "high"},
											Init: &ast.Literal{Value: 1.0},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: map[string]bool{"high": true},
		},
		{
			name: "empty for-in loop body",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID:   &ast.Identifier{Name: "total"},
							Init: &ast.Literal{Value: 0.0},
						},
					},
				},
				&ast.ForInStatement{
					ElementVar: "val",
					Collection: &ast.Identifier{Name: "arr"},
					Body:       []ast.Node{},
				},
			},
			expected: map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewArrowLoopModificationAnalyzer()
			result := analyzer.FindLoopModifiedVariables(tt.body)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d modified variables, got %d", len(tt.expected), len(result))
			}

			for varName := range tt.expected {
				if !result[varName] {
					t.Errorf("Expected variable %q to be marked as loop-modified", varName)
				}
			}

			for varName := range result {
				if !tt.expected[varName] {
					t.Errorf("Variable %q unexpectedly marked as loop-modified", varName)
				}
			}
		})
	}
}
