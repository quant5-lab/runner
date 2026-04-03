package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestPlotStatementScope validates plot() execution scope invariants across control flow structures */
func TestPlotStatementScope(t *testing.T) {
	tests := []struct {
		name              string
		program           *ast.Program
		expectedPlotCount int
		scopeValidator    func(t *testing.T, code string)
	}{
		{
			name: "single plot after if statement",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.IfStatement{
						Test: &ast.Identifier{Name: "condition"},
						Consequent: []ast.Node{
							&ast.ExpressionStatement{
								Expression: &ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "strategy"},
										Property: &ast.Identifier{Name: "close_all"},
									},
								},
							},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "plot"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "close"},
							},
						},
					},
				},
			},
			expectedPlotCount: 1,
			scopeValidator: func(t *testing.T, code string) {
				assertPlotAfterClosingBrace(t, code, "CloseAll")
				assertPlotScopeShallowerThan(t, code, "CloseAll")
			},
		},
		{
			name: "multiple plots after nested conditionals",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.IfStatement{
						Test: &ast.Identifier{Name: "condition1"},
						Consequent: []ast.Node{
							&ast.IfStatement{
								Test: &ast.Identifier{Name: "condition2"},
								Consequent: []ast.Node{
									&ast.ExpressionStatement{
										Expression: &ast.CallExpression{
											Callee: &ast.MemberExpression{
												Object:   &ast.Identifier{Name: "strategy"},
												Property: &ast.Identifier{Name: "entry"},
											},
										},
									},
								},
							},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "plot"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "sma20"},
							},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "plot"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "sma50"},
							},
						},
					},
				},
			},
			expectedPlotCount: 2,
			scopeValidator: func(t *testing.T, code string) {
				assertPlotCount(t, code, 2)
				assertAllPlotsAtBarLoopScope(t, code)
			},
		},
		{
			name: "plot with variable declaration in conditional",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.IfStatement{
						Test: &ast.BinaryExpression{
							Operator: ">",
							Left:     &ast.Identifier{Name: "close"},
							Right:    &ast.Identifier{Name: "open"},
						},
						Consequent: []ast.Node{
							&ast.VariableDeclaration{
								Declarations: []ast.VariableDeclarator{
									{
										ID:   &ast.Identifier{Name: "signal"},
										Init: &ast.Literal{Value: 1.0},
									},
								},
							},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "plot"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "signal"},
							},
						},
					},
				},
			},
			expectedPlotCount: 1,
			scopeValidator: func(t *testing.T, code string) {
				assertPlotAfterAllConditionals(t, code)
			},
		},
		{
			name: "no plots produces no collector.Add calls",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.IfStatement{
						Test: &ast.Identifier{Name: "condition"},
						Consequent: []ast.Node{
							&ast.ExpressionStatement{
								Expression: &ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "strategy"},
										Property: &ast.Identifier{Name: "close_all"},
									},
								},
							},
						},
					},
				},
			},
			expectedPlotCount: 0,
			scopeValidator: func(t *testing.T, code string) {
				if strings.Contains(code, "collector.Add") {
					t.Error("No plot() calls in source, but collector.Add found in generated code")
				}
			},
		},
		{
			name: "plot inside conditional is moved outside",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.IfStatement{
						Test: &ast.Identifier{Name: "condition"},
						Consequent: []ast.Node{
							&ast.ExpressionStatement{
								Expression: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "plot"},
									Arguments: []ast.Expression{
										&ast.Identifier{Name: "value"},
									},
								},
							},
						},
					},
				},
			},
			expectedPlotCount: 1,
			scopeValidator: func(t *testing.T, code string) {
				assertPlotAfterClosingBrace(t, code, "condition")
				assertPlotAtBarLoopScope(t, code)
			},
		},
		{
			name: "plots with if-else structure",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.IfStatement{
						Test: &ast.BinaryExpression{
							Operator: ">",
							Left:     &ast.Identifier{Name: "close"},
							Right:    &ast.Identifier{Name: "sma20"},
						},
						Consequent: []ast.Node{
							&ast.VariableDeclaration{
								Declarations: []ast.VariableDeclarator{
									{
										ID:   &ast.Identifier{Name: "signal"},
										Init: &ast.Literal{Value: 1.0},
									},
								},
							},
						},
						Alternate: []ast.Node{
							&ast.VariableDeclaration{
								Declarations: []ast.VariableDeclarator{
									{
										ID:   &ast.Identifier{Name: "signal"},
										Init: &ast.Literal{Value: -1.0},
									},
								},
							},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "plot"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "signal"},
							},
						},
					},
				},
			},
			expectedPlotCount: 1,
			scopeValidator: func(t *testing.T, code string) {
				assertPlotAfterAllConditionals(t, code)
				assertPlotAtBarLoopScope(t, code)
			},
		},
		{
			name: "multiple plots between conditionals",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.IfStatement{
						Test: &ast.Identifier{Name: "condition1"},
						Consequent: []ast.Node{
							&ast.ExpressionStatement{
								Expression: &ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "strategy"},
										Property: &ast.Identifier{Name: "entry"},
									},
								},
							},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "plot"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "value1"},
							},
						},
					},
					&ast.IfStatement{
						Test: &ast.Identifier{Name: "condition2"},
						Consequent: []ast.Node{
							&ast.ExpressionStatement{
								Expression: &ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "strategy"},
										Property: &ast.Identifier{Name: "exit"},
									},
								},
							},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "plot"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "value2"},
							},
						},
					},
				},
			},
			expectedPlotCount: 2,
			scopeValidator: func(t *testing.T, code string) {
				assertPlotCount(t, code, 2)
				assertAllPlotsAtBarLoopScope(t, code)
				assertPlotsAtEndOfBarLoop(t, code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := GenerateStrategyCodeFromAST(tt.program)
			if err != nil {
				t.Fatalf("Code generation failed: %v", err)
			}

			if tt.scopeValidator != nil {
				tt.scopeValidator(t, code.FunctionBody)
			}
		})
	}
}

/* Assertion helpers */

func assertPlotAfterClosingBrace(t *testing.T, code, marker string) {
	t.Helper()
	lines := strings.Split(code, "\n")
	foundMarker := false
	foundClosingBrace := false
	foundPlot := false

	for _, line := range lines {
		if strings.Contains(line, marker) {
			foundMarker = true
		}
		if foundMarker && strings.Contains(line, "}") && !foundPlot {
			foundClosingBrace = true
		}
		if strings.Contains(line, "collector.Add") {
			foundPlot = true
			if !foundClosingBrace {
				t.Errorf("Plot found before closing brace of conditional containing %q", marker)
			}
			break
		}
	}
}

func assertPlotScopeShallowerThan(t *testing.T, code, marker string) {
	t.Helper()
	markerIdx := strings.Index(code, marker)
	plotIdx := strings.Index(code, "collector.Add")

	if markerIdx == -1 || plotIdx == -1 {
		return
	}

	markerIndent := countTrailingIndentation(code[:markerIdx])
	plotIndent := countTrailingIndentation(code[:plotIdx])

	if plotIndent >= markerIndent {
		t.Errorf("Plot indentation (%d tabs) >= marker %q indentation (%d tabs), should be shallower",
			plotIndent, marker, markerIndent)
	}
}

func assertPlotCount(t *testing.T, code string, expected int) {
	t.Helper()
	actual := strings.Count(code, "collector.Add")
	if actual != expected {
		t.Errorf("Expected %d plot calls, got %d", expected, actual)
	}
}

func assertAllPlotsAtBarLoopScope(t *testing.T, code string) {
	t.Helper()
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		if strings.Contains(line, "collector.Add") {
			indent := countLeadingIndentation(line)
			if indent != 1 {
				t.Errorf("Plot at wrong scope: expected 1 tab (bar loop), got %d tabs: %q", indent, line)
			}
		}
	}
}

func assertPlotAtBarLoopScope(t *testing.T, code string) {
	t.Helper()
	assertAllPlotsAtBarLoopScope(t, code)
}

func assertPlotAfterAllConditionals(t *testing.T, code string) {
	t.Helper()
	lines := strings.Split(code, "\n")
	lastIfBlock := -1
	firstPlot := -1

	for i, line := range lines {
		if strings.Contains(line, "if ") {
			lastIfBlock = i
		}
		if strings.Contains(line, "collector.Add") && firstPlot == -1 {
			firstPlot = i
			break
		}
	}

	if lastIfBlock != -1 && firstPlot != -1 {
		foundClosing := false
		for i := lastIfBlock; i < firstPlot; i++ {
			if strings.Contains(lines[i], "}") {
				foundClosing = true
				break
			}
		}
		if !foundClosing {
			t.Error("Plot found before conditional block closed")
		}
	}
}

func assertPlotsAtEndOfBarLoop(t *testing.T, code string) {
	t.Helper()
	lines := strings.Split(code, "\n")
	lastPlot := -1
	barLoopEnd := -1

	for i, line := range lines {
		if strings.Contains(line, "collector.Add") {
			lastPlot = i
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "}" && lastPlot != -1 && barLoopEnd == -1 {
			indent := countLeadingIndentation(line)
			if indent == 0 {
				barLoopEnd = i
				break
			}
		}
	}

	if lastPlot != -1 && barLoopEnd != -1 {
		nonCommentLinesBetween := 0
		for i := lastPlot + 1; i < barLoopEnd; i++ {
			trimmed := strings.TrimSpace(lines[i])
			if trimmed != "" && !strings.HasPrefix(trimmed, "//") {
				nonCommentLinesBetween++
			}
		}
		if nonCommentLinesBetween > 6 {
			t.Errorf("Found %d non-empty lines between last plot and bar loop end - plots should be at end",
				nonCommentLinesBetween)
		}
	}
}

/* Helper functions */

func countTrailingIndentation(s string) int {
	lastNewline := strings.LastIndex(s, "\n")
	if lastNewline == -1 {
		return 0
	}
	indent := 0
	for i := lastNewline + 1; i < len(s); i++ {
		if s[i] == '\t' {
			indent++
		} else if s[i] != ' ' {
			break
		}
	}
	return indent
}

func countLeadingIndentation(line string) int {
	indent := 0
	for _, ch := range line {
		if ch == '\t' {
			indent++
		} else if ch != ' ' {
			break
		}
	}
	return indent
}
