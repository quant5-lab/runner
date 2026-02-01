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
				indent:             2,
				directionExtractor: NewDefaultDirectionExtractor(),
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

/*
TestStrategyCloseWhenParameter verifies when= parameter handling for strategy.close()
Tests conditional wrapping behavior across all edge cases.
*/
func TestStrategyCloseWhenParameter(t *testing.T) {
	tests := []struct {
		name        string
		call        *ast.CallExpression
		expectIf    bool
		expectClose string
		expectCond  string
	}{
		{
			name: "close with when condition wraps in if-block",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "close"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Long"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "when"},
								Value: &ast.Identifier{Name: "exitSignal"},
							},
						},
					},
				},
			},
			expectIf:    true,
			expectClose: `strat.Close("Long"`,
			expectCond:  "if value.IsTrue(",
		},
		{
			name: "close without when has no wrapper",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "close"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Short"},
				},
			},
			expectIf:    false,
			expectClose: `strat.Close("Short"`,
			expectCond:  "",
		},
		{
			name: "close with when and comment parameter",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "close"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Position1"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "comment"},
								Value: &ast.Literal{Value: "Exit signal"},
							},
							{
								Key: &ast.Identifier{Name: "when"},
								Value: &ast.BinaryExpression{
									Left:     &ast.Identifier{Name: "rsi"},
									Operator: ">",
									Right:    &ast.Literal{Value: 70.0},
								},
							},
						},
					},
				},
			},
			expectIf:    true,
			expectClose: `strat.Close("Position1"`,
			expectCond:  "if value.IsTrue(",
		},
		{
			name: "close with complex when condition",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "close"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Entry"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key: &ast.Identifier{Name: "when"},
								Value: &ast.BinaryExpression{
									Left: &ast.BinaryExpression{
										Left:     &ast.Identifier{Name: "close"},
										Operator: "<",
										Right:    &ast.Identifier{Name: "stopLoss"},
									},
									Operator: "or",
									Right: &ast.BinaryExpression{
										Left:     &ast.Identifier{Name: "close"},
										Operator: ">",
										Right:    &ast.Identifier{Name: "takeProfit"},
									},
								},
							},
						},
					},
				},
			},
			expectIf:    true,
			expectClose: `strat.Close("Entry"`,
			expectCond:  "if value.IsTrue(",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &generator{
				strategyConfig: &StrategyConfig{
					DefaultQtyType:  "strategy.fixed",
					DefaultQtyValue: 1.0,
				},
				indent: 2,
			}

			handler := NewStrategyActionHandler()
			generated, err := handler.generateClose(g, tt.call)
			if err != nil {
				t.Fatalf("generateClose failed: %v", err)
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

			if !containsSubstring(generated, tt.expectClose) {
				t.Errorf("Expected %q in generated code, got:\n%s", tt.expectClose, generated)
			}
		})
	}
}

/*
TestStrategyCloseAllWhenParameter verifies when= parameter handling for strategy.close_all()
Tests conditional wrapping for closing all positions.
*/
func TestStrategyCloseAllWhenParameter(t *testing.T) {
	tests := []struct {
		name           string
		call           *ast.CallExpression
		expectIf       bool
		expectCloseAll string
		expectCond     string
	}{
		{
			name: "close_all with when condition wraps in if-block",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "close_all"},
				},
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "when"},
								Value: &ast.Identifier{Name: "panicExit"},
							},
						},
					},
				},
			},
			expectIf:       true,
			expectCloseAll: "strat.CloseAll(",
			expectCond:     "if value.IsTrue(",
		},
		{
			name: "close_all without when has no wrapper",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "close_all"},
				},
				Arguments: []ast.Expression{},
			},
			expectIf:       false,
			expectCloseAll: "strat.CloseAll(",
			expectCond:     "",
		},
		{
			name: "close_all with when and comment",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "close_all"},
				},
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "comment"},
								Value: &ast.Literal{Value: "Emergency exit"},
							},
							{
								Key: &ast.Identifier{Name: "when"},
								Value: &ast.BinaryExpression{
									Left:     &ast.Identifier{Name: "volatility"},
									Operator: ">",
									Right:    &ast.Literal{Value: 100.0},
								},
							},
						},
					},
				},
			},
			expectIf:       true,
			expectCloseAll: "strat.CloseAll(",
			expectCond:     "if value.IsTrue(",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &generator{
				strategyConfig: &StrategyConfig{
					DefaultQtyType:  "strategy.fixed",
					DefaultQtyValue: 1.0,
				},
				indent: 2,
			}

			handler := NewStrategyActionHandler()
			generated, err := handler.generateCloseAll(g, tt.call)
			if err != nil {
				t.Fatalf("generateCloseAll failed: %v", err)
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

			if !containsSubstring(generated, tt.expectCloseAll) {
				t.Errorf("Expected %q in generated code, got:\n%s", tt.expectCloseAll, generated)
			}
		})
	}
}

/*
TestStrategyExitWhenParameter verifies when= parameter handling for strategy.exit()
Tests conditional wrapping with stop/limit levels.
*/
func TestStrategyExitWhenParameter(t *testing.T) {
	tests := []struct {
		name       string
		call       *ast.CallExpression
		expectIf   bool
		expectExit string
		expectCond string
	}{
		{
			name: "exit with when condition wraps in if-block",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "exit"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "ExitOrder"},
					&ast.Literal{Value: "Long"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "stop"},
								Value: &ast.Literal{Value: 95.0},
							},
							{
								Key:   &ast.Identifier{Name: "limit"},
								Value: &ast.Literal{Value: 105.0},
							},
							{
								Key:   &ast.Identifier{Name: "when"},
								Value: &ast.Identifier{Name: "exitCondition"},
							},
						},
					},
				},
			},
			expectIf:   true,
			expectExit: `strat.ExitWithLevels("ExitOrder"`,
			expectCond: "if value.IsTrue(",
		},
		{
			name: "exit without when has no wrapper",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "exit"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Exit1"},
					&ast.Literal{Value: "Short"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "stop"},
								Value: &ast.Literal{Value: 100.0},
							},
						},
					},
				},
			},
			expectIf:   false,
			expectExit: `strat.ExitWithLevels("Exit1"`,
			expectCond: "",
		},
		{
			name: "exit with when and dynamic stop/limit",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "exit"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "DynamicExit"},
					&ast.Literal{Value: "Entry"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "stop"},
								Value: &ast.Identifier{Name: "stopLevel"},
							},
							{
								Key:   &ast.Identifier{Name: "limit"},
								Value: &ast.Identifier{Name: "profitTarget"},
							},
							{
								Key: &ast.Identifier{Name: "when"},
								Value: &ast.BinaryExpression{
									Left:     &ast.Identifier{Name: "barIndex"},
									Operator: ">",
									Right:    &ast.Literal{Value: 10.0},
								},
							},
						},
					},
				},
			},
			expectIf:   true,
			expectExit: `strat.ExitWithLevels("DynamicExit"`,
			expectCond: "if value.IsTrue(",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &generator{
				strategyConfig: &StrategyConfig{
					DefaultQtyType:  "strategy.fixed",
					DefaultQtyValue: 1.0,
				},
				indent: 2,
			}

			handler := NewStrategyActionHandler()
			generated, err := handler.generateExit(g, tt.call)
			if err != nil {
				t.Fatalf("generateExit failed: %v", err)
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

			if !containsSubstring(generated, tt.expectExit) {
				t.Errorf("Expected %q in generated code, got:\n%s", tt.expectExit, generated)
			}
		})
	}
}

/*
TestStrategyActionsWhenParameter_BackwardCompatibility verifies all strategy actions
without when= parameter generate identical code as before (no regression).
*/
func TestStrategyActionsWhenParameter_BackwardCompatibility(t *testing.T) {
	g := &generator{
		strategyConfig: &StrategyConfig{
			DefaultQtyType:  "strategy.fixed",
			DefaultQtyValue: 1.0,
		},
		indent: 2,
	}

	handler := NewStrategyActionHandler()

	tests := []struct {
		name     string
		call     *ast.CallExpression
		generate func(*generator, *ast.CallExpression) (string, error)
		expect   string
	}{
		{
			name: "entry without when",
			call: &ast.CallExpression{
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
				},
			},
			generate: handler.generateEntry,
			expect:   `strat.Entry("Long", strategy.Long, 1, "")`,
		},
		{
			name: "close without when",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "close"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Long"},
				},
			},
			generate: handler.generateClose,
			expect:   `strat.Close("Long", bar.Close, bar.Time, "")`,
		},
		{
			name: "close_all without when",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "close_all"},
				},
				Arguments: []ast.Expression{},
			},
			generate: handler.generateCloseAll,
			expect:   `strat.CloseAll(bar.Close, bar.Time, "")`,
		},
		{
			name: "exit without when",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "exit"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Exit"},
					&ast.Literal{Value: "Long"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "stop"},
								Value: &ast.Literal{Value: 100.0},
							},
						},
					},
				},
			},
			generate: handler.generateExit,
			expect:   `strat.ExitWithLevels("Exit", "Long"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generated, err := tt.generate(g, tt.call)
			if err != nil {
				t.Fatalf("generate failed: %v", err)
			}

			if containsSubstring(generated, "if value.IsTrue(") {
				t.Errorf("Backward compatibility broken: should NOT have if-wrapper without when=, got:\n%s", generated)
			}

			if !containsSubstring(generated, tt.expect) {
				t.Errorf("Expected %q in generated code, got:\n%s", tt.expect, generated)
			}
		})
	}
}

/*
TestStrategyActionsWhenParameter_IndentationConsistency verifies all strategy actions
preserve proper indentation when wrapped in if-blocks.
*/
func TestStrategyActionsWhenParameter_IndentationConsistency(t *testing.T) {
	indentLevels := []int{0, 1, 2, 3, 4}

	for _, level := range indentLevels {
		t.Run(stringFromInt(level)+" tabs", func(t *testing.T) {
			g := &generator{
				strategyConfig: &StrategyConfig{
					DefaultQtyType:  "strategy.fixed",
					DefaultQtyValue: 1.0,
				},
				indent: level,
			}

			handler := NewStrategyActionHandler()

			actions := []struct {
				name     string
				call     *ast.CallExpression
				generate func(*generator, *ast.CallExpression) (string, error)
			}{
				{
					name: "entry",
					call: &ast.CallExpression{
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
									{Key: &ast.Identifier{Name: "when"}, Value: &ast.Identifier{Name: "cond"}},
								},
							},
						},
					},
					generate: handler.generateEntry,
				},
				{
					name: "close",
					call: &ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "strategy"},
							Property: &ast.Identifier{Name: "close"},
						},
						Arguments: []ast.Expression{
							&ast.Literal{Value: "Test"},
							&ast.ObjectExpression{
								Properties: []ast.Property{
									{Key: &ast.Identifier{Name: "when"}, Value: &ast.Identifier{Name: "cond"}},
								},
							},
						},
					},
					generate: handler.generateClose,
				},
				{
					name: "close_all",
					call: &ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "strategy"},
							Property: &ast.Identifier{Name: "close_all"},
						},
						Arguments: []ast.Expression{
							&ast.ObjectExpression{
								Properties: []ast.Property{
									{Key: &ast.Identifier{Name: "when"}, Value: &ast.Identifier{Name: "cond"}},
								},
							},
						},
					},
					generate: handler.generateCloseAll,
				},
				{
					name: "exit",
					call: &ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "strategy"},
							Property: &ast.Identifier{Name: "exit"},
						},
						Arguments: []ast.Expression{
							&ast.Literal{Value: "Exit"},
							&ast.Literal{Value: "Test"},
							&ast.ObjectExpression{
								Properties: []ast.Property{
									{Key: &ast.Identifier{Name: "stop"}, Value: &ast.Literal{Value: 100.0}},
									{Key: &ast.Identifier{Name: "when"}, Value: &ast.Identifier{Name: "cond"}},
								},
							},
						},
					},
					generate: handler.generateExit,
				},
			}

			for _, action := range actions {
				t.Run(action.name, func(t *testing.T) {
					generated, err := action.generate(g, action.call)
					if err != nil {
						t.Fatalf("generate failed: %v", err)
					}

					lines := strings.Split(generated, "\n")
					expectedPrefix := strings.Repeat("\t", level)

					for i, line := range lines {
						if line != "" && !strings.HasPrefix(line, expectedPrefix) {
							t.Errorf("Line %d not properly indented (expected %d tabs): %q\nFull output:\n%s",
								i, level, line, generated)
						}
					}
				})
			}
		})
	}
}
