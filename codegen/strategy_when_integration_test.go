package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/*
TestStrategyEntryWhenParameter_DirectHandlerIntegration tests when= parameter
through direct handler invocation without full parser pipeline.
*/
func TestStrategyEntryWhenParameter_DirectHandlerIntegration(t *testing.T) {
	tests := []struct {
		name        string
		call        *ast.CallExpression
		qtyType     string
		qtyValue    float64
		expectIf    bool
		expectEntry string
		expectCond  string
	}{
		{
			name: "entry with when condition wraps in if-block",
			call: &ast.CallExpression{
				NodeType: ast.TypeCallExpression,
				Callee: &ast.MemberExpression{
					NodeType: ast.TypeMemberExpression,
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Long"},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "long"},
					},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "qty"},
								Value: &ast.Literal{Value: 1.0},
							},
							{
								Key: &ast.Identifier{Name: "when"},
								Value: &ast.BinaryExpression{
									Left:     &ast.Identifier{Name: "close"},
									Operator: ">",
									Right:    &ast.Identifier{Name: "open"},
								},
							},
						},
					},
				},
			},
			qtyType:     "strategy.fixed",
			qtyValue:    1.0,
			expectIf:    true,
			expectEntry: `strat.Entry("Long"`,
			expectCond:  "if value.IsTrue(",
		},
		{
			name: "entry without when has no wrapper",
			call: &ast.CallExpression{
				NodeType: ast.TypeCallExpression,
				Callee: &ast.MemberExpression{
					NodeType: ast.TypeMemberExpression,
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Short"},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "short"},
					},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "qty"},
								Value: &ast.Literal{Value: 2.0},
							},
						},
					},
				},
			},
			qtyType:     "strategy.fixed",
			qtyValue:    2.0,
			expectIf:    false,
			expectEntry: `strat.Entry("Short"`,
			expectCond:  "",
		},
		{
			name: "entry with when and cash qty type",
			call: &ast.CallExpression{
				NodeType: ast.TypeCallExpression,
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Long"},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "long"},
					},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "when"},
								Value: &ast.Identifier{Name: "buySignal"},
							},
						},
					},
				},
			},
			qtyType:     "strategy.cash",
			qtyValue:    1000.0,
			expectIf:    true,
			expectEntry: "entryQty :=",
			expectCond:  "if value.IsTrue(",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &generator{
				strategyConfig: &StrategyConfig{
					DefaultQtyType:  tt.qtyType,
					DefaultQtyValue: tt.qtyValue,
				},
				indent: 2,
			}

			handler := NewStrategyActionHandler()
			generated, err := handler.generateEntry(g, tt.call)
			if err != nil {
				t.Fatalf("generateEntry failed: %v", err)
			}

			if tt.expectIf {
				if !containsSubstring(generated, tt.expectCond) {
					t.Errorf("Expected if-wrapper with %q, got:\n%s", tt.expectCond, generated)
				}
			} else {
				if containsSubstring(generated, "if value.IsTrue(") {
					t.Errorf("Should NOT have if-wrapper, got:\n%s", generated)
				}
			}

			if !containsSubstring(generated, tt.expectEntry) {
				t.Errorf("Expected %q in generated code, got:\n%s", tt.expectEntry, generated)
			}
		})
	}
}

func TestStrategyEntryWhenParameter_BackwardCompatibility(t *testing.T) {
	/* Verify entries without when= generate identical code as before */
	g := &generator{
		strategyConfig: &StrategyConfig{
			DefaultQtyType:  "strategy.fixed",
			DefaultQtyValue: 1.0,
		},
		indent: 2,
	}

	call := &ast.CallExpression{
		NodeType: ast.TypeCallExpression,
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: "entry"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "Entry1"},
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "long"},
			},
		},
	}

	handler := NewStrategyActionHandler()
	generated, err := handler.generateEntry(g, call)
	if err != nil {
		t.Fatalf("generateEntry failed: %v", err)
	}

	if containsSubstring(generated, "if value.IsTrue(") {
		t.Error("Backward compatibility broken: entry without when= should not have if-wrapper")
	}

	if !containsSubstring(generated, `strat.Entry("Entry1", strategy.Long, 1, "")`) {
		t.Errorf("Expected standard entry call, got:\n%s", generated)
	}
}

func TestStrategyEntryWhenParameter_IndentationPreserved(t *testing.T) {
	indentLevels := []int{0, 1, 2, 3}

	for _, level := range indentLevels {
		t.Run(stringFromInt(level)+" tabs", func(t *testing.T) {
			g := &generator{
				strategyConfig: &StrategyConfig{
					DefaultQtyType:  "strategy.fixed",
					DefaultQtyValue: 1.0,
				},
				indent: level,
			}

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Test"},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "long"},
					},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "when"},
								Value: &ast.Identifier{Name: "condition"},
							},
						},
					},
				},
			}

			handler := NewStrategyActionHandler()
			generated, err := handler.generateEntry(g, call)
			if err != nil {
				t.Fatalf("generateEntry failed: %v", err)
			}

			lines := strings.Split(generated, "\n")
			expectedPrefix := strings.Repeat("\t", level)

			for _, line := range lines {
				if line != "" && !strings.HasPrefix(line, expectedPrefix) {
					t.Errorf("Line not properly indented (expected %d tabs): %q", level, line)
				}
			}
		})
	}
}

func stringFromInt(n int) string {
	return string(rune('0' + n))
}
