package security

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestExpressionSecurityWalker_Isolation verifies walker state isolation */
func TestExpressionSecurityWalker_Isolation(t *testing.T) {
	walker1 := NewExpressionSecurityWalker()
	walker2 := NewExpressionSecurityWalker()

	call1 := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "security"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "BTCUSDT"},
			&ast.Literal{Value: "1D"},
			&ast.Identifier{Name: "close"},
		},
	}

	call2 := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "security"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "ETHUSDT"},
			&ast.Literal{Value: "1h"},
			&ast.Identifier{Name: "high"},
		},
	}

	walker1.Walk(call1)
	walker2.Walk(call2)

	calls1 := walker1.GetCalls()
	calls2 := walker2.GetCalls()

	if len(calls1) != 1 {
		t.Errorf("Walker1: expected 1 call, got %d", len(calls1))
	}
	if len(calls2) != 1 {
		t.Errorf("Walker2: expected 1 call, got %d", len(calls2))
	}

	if calls1[0].Symbol == calls2[0].Symbol {
		t.Error("Walker state leaked between instances")
	}
}

/* TestExpressionSecurityWalker_NilExpression verifies nil handling */
func TestExpressionSecurityWalker_NilExpression(t *testing.T) {
	walker := NewExpressionSecurityWalker()
	walker.Walk(nil)

	calls := walker.GetCalls()
	if len(calls) != 0 {
		t.Errorf("Expected 0 calls for nil expression, got %d", len(calls))
	}
}

/* TestExpressionSecurityWalker_UnknownExpressionType verifies graceful handling */
func TestExpressionSecurityWalker_UnknownExpressionType(t *testing.T) {
	walker := NewExpressionSecurityWalker()

	unknownExpr := &ast.Literal{Value: "some value"}
	walker.Walk(unknownExpr)

	calls := walker.GetCalls()
	if len(calls) != 0 {
		t.Errorf("Expected 0 calls for unknown expression, got %d", len(calls))
	}
}

/* TestExpressionSecurityWalker_RecursiveDepth verifies deep recursion handling */
func TestExpressionSecurityWalker_RecursiveDepth(t *testing.T) {
	walker := NewExpressionSecurityWalker()

	depth := 100
	var buildNestedBinary func(int) ast.Expression
	buildNestedBinary = func(level int) ast.Expression {
		if level == 0 {
			return &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
					&ast.Literal{Value: "1D"},
					&ast.Identifier{Name: "close"},
				},
			}
		}
		return &ast.BinaryExpression{
			Operator: "+",
			Left:     buildNestedBinary(level - 1),
			Right:    &ast.Literal{Value: float64(level)},
		}
	}

	expr := buildNestedBinary(depth)
	walker.Walk(expr)

	calls := walker.GetCalls()
	if len(calls) != 1 {
		t.Errorf("Expected 1 call after %d levels of nesting, got %d", depth, len(calls))
	}
}

/* TestStatementSecurityWalker_NilStatement verifies nil handling */
func TestStatementSecurityWalker_NilStatement(t *testing.T) {
	exprWalker := NewExpressionSecurityWalker()
	stmtWalker := NewStatementSecurityWalker(exprWalker)

	stmtWalker.Walk(nil)

	calls := exprWalker.GetCalls()
	if len(calls) != 0 {
		t.Errorf("Expected 0 calls for nil statement, got %d", len(calls))
	}
}

/* TestStatementSecurityWalker_UnknownStatementType verifies graceful handling */
func TestStatementSecurityWalker_UnknownStatementType(t *testing.T) {
	exprWalker := NewExpressionSecurityWalker()
	stmtWalker := NewStatementSecurityWalker(exprWalker)

	unknownStmt := &ast.ExpressionStatement{
		Expression: &ast.Literal{Value: 42},
	}
	stmtWalker.Walk(unknownStmt)

	calls := exprWalker.GetCalls()
	if len(calls) != 0 {
		t.Errorf("Expected 0 calls for non-security statement, got %d", len(calls))
	}
}

/* TestSecurityCallMatcher_InvalidCallExpression verifies rejection of invalid calls */
func TestSecurityCallMatcher_InvalidCallExpression(t *testing.T) {
	tests := []struct {
		name string
		call *ast.CallExpression
	}{
		{
			name: "nil_call",
			call: nil,
		},
		{
			name: "wrong_function_name",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.sma"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
					&ast.Literal{Value: "1D"},
					&ast.Identifier{Name: "close"},
				},
			},
		},
		{
			name: "insufficient_args_0",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{},
			},
		},
		{
			name: "insufficient_args_1",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
				},
			},
		},
		{
			name: "insufficient_args_2",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
					&ast.Literal{Value: "1D"},
				},
			},
		},
	}

	matcher := NewSecurityCallMatcher()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.Match(tt.call)
			if result != nil {
				t.Errorf("Expected nil result for invalid call, got %+v", result)
			}
		})
	}
}

/* TestSecurityCallMatcher_ValidCallExpression verifies acceptance of valid calls */
func TestSecurityCallMatcher_ValidCallExpression(t *testing.T) {
	tests := []struct {
		name           string
		call           *ast.CallExpression
		expectedSymbol string
		expectedTF     string
	}{
		{
			name: "security_identifier",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
					&ast.Literal{Value: "1D"},
					&ast.Identifier{Name: "close"},
				},
			},
			expectedSymbol: "BTCUSDT",
			expectedTF:     "1D",
		},
		{
			name: "request_security_member",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "request"},
					Property: &ast.Identifier{Name: "security"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "ETHUSDT"},
					&ast.Literal{Value: "1h"},
					&ast.Identifier{Name: "high"},
				},
			},
			expectedSymbol: "ETHUSDT",
			expectedTF:     "1h",
		},
		{
			name: "runtime_symbol",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "syminfo"},
						Property: &ast.Identifier{Name: "tickerid"},
					},
					&ast.Literal{Value: "1W"},
					&ast.Identifier{Name: "close"},
				},
			},
			expectedSymbol: "syminfo.tickerid",
			expectedTF:     "1W",
		},
	}

	matcher := NewSecurityCallMatcher()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.Match(tt.call)
			if result == nil {
				t.Fatal("Expected non-nil result for valid call")
			}

			if result.Symbol != tt.expectedSymbol {
				t.Errorf("Expected symbol '%s', got '%s'", tt.expectedSymbol, result.Symbol)
			}
			if result.Timeframe != tt.expectedTF {
				t.Errorf("Expected timeframe '%s', got '%s'", tt.expectedTF, result.Timeframe)
			}
			if result.Expression == nil {
				t.Error("Expected non-nil expression")
			}
		})
	}
}

/* TestExpressionSecurityWalker_AllExpressionTypes verifies coverage of all expression types */
func TestExpressionSecurityWalker_AllExpressionTypes(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		hasCalls bool
	}{
		{
			name: "CallExpression_security",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
					&ast.Literal{Value: "1D"},
					&ast.Identifier{Name: "close"},
				},
			},
			hasCalls: true,
		},
		{
			name: "CallExpression_nonsecurity",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 20.0},
				},
			},
			hasCalls: false,
		},
		{
			name: "ConditionalExpression",
			expr: &ast.ConditionalExpression{
				Test:       &ast.Literal{Value: true},
				Consequent: &ast.Literal{Value: 1.0},
				Alternate:  &ast.Literal{Value: 2.0},
			},
			hasCalls: false,
		},
		{
			name: "BinaryExpression",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Literal{Value: 1.0},
				Right:    &ast.Literal{Value: 2.0},
			},
			hasCalls: false,
		},
		{
			name: "UnaryExpression",
			expr: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.Literal{Value: 5.0},
			},
			hasCalls: false,
		},
		{
			name: "MemberExpression",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "syminfo"},
				Property: &ast.Identifier{Name: "tickerid"},
			},
			hasCalls: false,
		},
		{
			name: "LogicalExpression",
			expr: &ast.LogicalExpression{
				Operator: "and",
				Left:     &ast.Literal{Value: true},
				Right:    &ast.Literal{Value: false},
			},
			hasCalls: false,
		},
		{
			name: "ArrowFunctionExpression_with_security",
			expr: &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{{Name: "src"}},
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "request"},
								Property: &ast.Identifier{Name: "security"},
							},
							Arguments: []ast.Expression{
								&ast.Literal{Value: "BTCUSDT"},
								&ast.Literal{Value: "1D"},
								&ast.Identifier{Name: "close"},
							},
						},
					},
				},
			},
			hasCalls: true,
		},
		{
			name: "ArrowFunctionExpression_without_security",
			expr: &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{{Name: "src"}},
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "ta"},
								Property: &ast.Identifier{Name: "sma"},
							},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "src"},
								&ast.Literal{Value: 20.0},
							},
						},
					},
				},
			},
			hasCalls: false,
		},
		{
			name: "ArrowFunctionExpression_empty_body",
			expr: &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{},
				Body:   []ast.Node{},
			},
			hasCalls: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			walker := NewExpressionSecurityWalker()
			walker.Walk(tt.expr)
			calls := walker.GetCalls()

			if tt.hasCalls {
				if len(calls) == 0 {
					t.Error("Expected security calls but got none")
				}
			} else {
				if len(calls) > 0 {
					t.Errorf("Expected no security calls but got %d", len(calls))
				}
			}
		})
	}
}
