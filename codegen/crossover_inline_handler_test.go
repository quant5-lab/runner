package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/*
TestCrossoverInlineHandler_ArgumentTypes validates all supported argument expression types.
Ensures crossover handles: identifiers, literals, member expressions, binary/unary/conditional expressions, TA calls.
*/
func TestCrossoverInlineHandler_ArgumentTypes(t *testing.T) {
	tests := []struct {
		name      string
		arg1      ast.Expression
		arg2      ast.Expression
		expectErr bool
		errMsg    string
	}{
		{
			name: "identifier vs identifier",
			arg1: &ast.Identifier{Name: "sma20"},
			arg2: &ast.Identifier{Name: "ema10"},
		},
		{
			name: "identifier vs literal",
			arg1: &ast.Identifier{Name: "close"},
			arg2: &ast.Literal{Value: 100.0},
		},
		{
			name: "literal vs literal",
			arg1: &ast.Literal{Value: 50.0},
			arg2: &ast.Literal{Value: 100.0},
		},
		{
			name: "member expression vs identifier",
			arg1: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 0},
				Computed: true,
			},
			arg2: &ast.Identifier{Name: "sma20"},
		},
		{
			name: "binary expression vs literal",
			arg1: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 10.0},
			},
			arg2: &ast.Literal{Value: 200.0},
		},
		{
			name: "unary expression vs identifier",
			arg1: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.Identifier{Name: "rsi"},
			},
			arg2: &ast.Literal{Value: 0.0},
		},
		{
			name: "conditional expression vs literal",
			arg1: &ast.ConditionalExpression{
				Test:       &ast.Identifier{Name: "condition"},
				Consequent: &ast.Identifier{Name: "close"},
				Alternate:  &ast.Identifier{Name: "open"},
			},
			arg2: &ast.Literal{Value: 100.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewCrossoverInlineHandler()
			gen := createTestGenerator()

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossover"},
				},
				Arguments: []ast.Expression{tt.arg1, tt.arg2},
			}

			code, err := handler.GenerateInline(call, gen)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("Expected error containing %q, got nil", tt.errMsg)
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			/* Validate IIFE structure */
			if !strings.Contains(code, "func() bool {") {
				t.Error("Missing IIFE wrapper")
			}
			if !strings.Contains(code, "if ctx.BarIndex == 0 { return false }") {
				t.Error("Missing warmup check")
			}
			if !strings.Contains(code, "curr1 :=") && !strings.Contains(code, "curr2 :=") {
				t.Error("Missing current value variables")
			}
			if !strings.Contains(code, "prev1 :=") && !strings.Contains(code, "prev2 :=") {
				t.Error("Missing previous value variables")
			}
		})
	}
}

/*
TestCrossoverInlineHandler_StatefulIndicatorDetection validates stateful indicator enforcement.
Ensures EMA/RMA must be extracted to variables before crossover usage.
*/
func TestCrossoverInlineHandler_StatefulIndicatorDetection(t *testing.T) {
	tests := []struct {
		name      string
		arg       ast.Expression
		expectErr bool
		errMsg    string
	}{
		{
			name: "stateful EMA in direct call - should fail",
			arg: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "ema"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 10.0},
				},
			},
			expectErr: true,
			errMsg:    "stateful indicator ta.ema must be assigned to variable",
		},
		{
			name: "stateful RMA in direct call - should fail",
			arg: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "rma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "high"},
					&ast.Literal{Value: 20.0},
				},
			},
			expectErr: true,
			errMsg:    "stateful indicator ta.rma must be assigned to variable",
		},
		{
			name: "stateful EMA in binary expression - should fail",
			arg: &ast.BinaryExpression{
				Operator: "+",
				Left: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "ema"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: 10.0},
					},
				},
				Right: &ast.Literal{Value: 5.0},
			},
			expectErr: true,
			errMsg:    "stateful indicator ta.ema must be assigned to variable",
		},
		{
			name: "stateful EMA in unary expression - should fail",
			arg: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "ema"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: 10.0},
					},
				},
			},
			expectErr: true,
			errMsg:    "stateful indicator ta.ema must be assigned to variable",
		},
		{
			name: "stateful EMA in conditional consequent - should fail",
			arg: &ast.ConditionalExpression{
				Test: &ast.Identifier{Name: "condition"},
				Consequent: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "ema"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: 10.0},
					},
				},
				Alternate: &ast.Identifier{Name: "close"},
			},
			expectErr: true,
			errMsg:    "stateful indicator ta.ema must be assigned to variable",
		},
		{
			name: "stateful EMA nested in SMA - should fail",
			arg: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "ta"},
							Property: &ast.Identifier{Name: "ema"},
						},
						Arguments: []ast.Expression{
							&ast.Identifier{Name: "close"},
							&ast.Literal{Value: 10.0},
						},
					},
					&ast.Literal{Value: 20.0},
				},
			},
			expectErr: true,
			errMsg:    "stateful indicator ta.ema must be assigned to variable",
		},
		{
			name: "window function SMA - should succeed",
			arg: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 20.0},
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewCrossoverInlineHandler()
			gen := createTestGenerator()

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossover"},
				},
				Arguments: []ast.Expression{
					tt.arg,
					&ast.Literal{Value: 50.0},
				},
			}

			_, err := handler.GenerateInline(call, gen)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("Expected error containing %q, got nil", tt.errMsg)
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			}
		})
	}
}

/*
TestCrossoverInlineHandler_WindowFunctionInlining validates inline IIFE generation for window functions.
Ensures SMA, WMA, STDEV generate proper current and previous bar IIFEs.
*/
func TestCrossoverInlineHandler_WindowFunctionInlining(t *testing.T) {
	tests := []struct {
		name       string
		funcName   string
		period     int
		expectCurr []string
		expectPrev []string
	}{
		{
			name:     "SMA period 20",
			funcName: "sma",
			period:   20,
			expectCurr: []string{
				"func() float64",
				"ctx.BarIndex < 19",
				"for j := 0; j < 20; j++",
				"sum / float64(20)",
			},
			expectPrev: []string{
				"func() float64",
				"ctx.BarIndex < 20", // +1 for previous bar offset
				"for j := 0; j < 20; j++",
			},
		},
		{
			name:     "WMA period 10",
			funcName: "wma",
			period:   10,
			expectCurr: []string{
				"func() float64",
				"ctx.BarIndex < 9",
				"for j := 0; j < 10; j++",
				"weightSum",
			},
			expectPrev: []string{
				"func() float64",
				"ctx.BarIndex < 10",
			},
		},
		{
			name:     "STDEV period 30",
			funcName: "stdev",
			period:   30,
			expectCurr: []string{
				"func() float64",
				"ctx.BarIndex < 29",
				"mean :=",
				"variance :=",
				"math.Sqrt",
			},
			expectPrev: []string{
				"func() float64",
				"ctx.BarIndex < 30",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewCrossoverInlineHandler()
			gen := createTestGenerator()

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossover"},
				},
				Arguments: []ast.Expression{
					&ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "ta"},
							Property: &ast.Identifier{Name: tt.funcName},
						},
						Arguments: []ast.Expression{
							&ast.Identifier{Name: "close"},
							&ast.Literal{Value: float64(tt.period)},
						},
					},
					&ast.Literal{Value: 100.0},
				},
			}

			code, err := handler.GenerateInline(call, gen)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			/* Validate current bar IIFE */
			for _, expected := range tt.expectCurr {
				if !strings.Contains(code, expected) {
					t.Errorf("Current bar IIFE missing: %q\nGenerated:\n%s", expected, code)
				}
			}

			/* Validate previous bar IIFE */
			for _, expected := range tt.expectPrev {
				if !strings.Contains(code, expected) {
					t.Errorf("Previous bar IIFE missing: %q\nGenerated:\n%s", expected, code)
				}
			}
		})
	}
}

/*
TestCrossoverInlineHandler_LogicConditions validates crossover vs crossunder condition logic.
*/
func TestCrossoverInlineHandler_LogicConditions(t *testing.T) {
	tests := []struct {
		name        string
		funcName    string
		expectLogic string
	}{
		{
			name:        "crossover logic",
			funcName:    "crossover",
			expectLogic: "curr1 > curr2 && prev1 <= prev2",
		},
		{
			name:        "crossunder logic",
			funcName:    "crossunder",
			expectLogic: "curr1 < curr2 && prev1 >= prev2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handler *CrossInlineHandler
			if tt.funcName == "crossover" {
				handler = NewCrossoverInlineHandler()
			} else {
				handler = NewCrossunderInlineHandler()
			}

			gen := createTestGenerator()

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: tt.funcName},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Identifier{Name: "sma20"},
				},
			}

			code, err := handler.GenerateInline(call, gen)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !strings.Contains(code, tt.expectLogic) {
				t.Errorf("Missing condition logic: %q\nGenerated:\n%s", tt.expectLogic, code)
			}
		})
	}
}

/*
TestCrossoverInlineHandler_EdgeCases validates boundary conditions and error handling.
*/
func TestCrossoverInlineHandler_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (*ast.CallExpression, *generator)
		expectErr bool
		errMsg    string
	}{
		{
			name: "missing second argument",
			setup: func() (*ast.CallExpression, *generator) {
				return &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "crossover"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
					},
				}, createTestGenerator()
			},
			expectErr: true,
			errMsg:    "requires 2 arguments",
		},
		{
			name: "empty arguments",
			setup: func() (*ast.CallExpression, *generator) {
				return &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "crossover"},
					},
					Arguments: []ast.Expression{},
				}, createTestGenerator()
			},
			expectErr: true,
			errMsg:    "requires 2 arguments",
		},
		{
			name: "RSI inline TA function now supported",
			setup: func() (*ast.CallExpression, *generator) {
				return &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "crossover"},
					},
					Arguments: []ast.Expression{
						&ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "ta"},
								Property: &ast.Identifier{Name: "rsi"},
							},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "close"},
								&ast.Literal{Value: 14.0},
							},
						},
						&ast.Literal{Value: 70.0},
					},
				}, createTestGenerator()
			},
			expectErr: false,
		},
		{
			name: "nested window functions",
			setup: func() (*ast.CallExpression, *generator) {
				return &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "crossover"},
					},
					Arguments: []ast.Expression{
						&ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "ta"},
								Property: &ast.Identifier{Name: "sma"},
							},
							Arguments: []ast.Expression{
								&ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "ta"},
										Property: &ast.Identifier{Name: "wma"},
									},
									Arguments: []ast.Expression{
										&ast.Identifier{Name: "close"},
										&ast.Literal{Value: 10.0},
									},
								},
								&ast.Literal{Value: 20.0},
							},
						},
						&ast.Identifier{Name: "high"},
					},
				}, createTestGenerator()
			},
			expectErr: false,
		},
		{
			name: "complex arithmetic both sides",
			setup: func() (*ast.CallExpression, *generator) {
				return &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "crossover"},
					},
					Arguments: []ast.Expression{
						&ast.BinaryExpression{
							Operator: "+",
							Left:     &ast.Identifier{Name: "close"},
							Right:    &ast.Literal{Value: 10.0},
						},
						&ast.BinaryExpression{
							Operator: "*",
							Left:     &ast.Identifier{Name: "sma20"},
							Right:    &ast.Literal{Value: 1.05},
						},
					},
				}, createTestGenerator()
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewCrossoverInlineHandler()
			call, gen := tt.setup()

			_, err := handler.GenerateInline(call, gen)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("Expected error containing %q, got nil", tt.errMsg)
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			}
		})
	}
}

/*
TestCrossoverInlineHandler_LiteralTypes validates all literal value types.
*/
func TestCrossoverInlineHandler_LiteralTypes(t *testing.T) {
	tests := []struct {
		name       string
		value      interface{}
		expectCode string
		expectErr  bool
	}{
		{
			name:       "float64 literal",
			value:      100.5,
			expectCode: "100.5",
		},
		{
			name:       "int literal",
			value:      50,
			expectCode: "50",
		},
		{
			name:       "bool true literal",
			value:      true,
			expectCode: "true",
		},
		{
			name:       "bool false literal",
			value:      false,
			expectCode: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewCrossoverInlineHandler()
			gen := createTestGenerator()

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossover"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: tt.value},
				},
			}

			code, err := handler.GenerateInline(call, gen)

			if tt.expectErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !strings.Contains(code, tt.expectCode) {
				t.Errorf("Expected literal %q in code\nGenerated:\n%s", tt.expectCode, code)
			}
		})
	}
}
