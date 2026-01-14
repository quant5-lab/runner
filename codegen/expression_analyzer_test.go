package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestExpressionAnalyzer_SimpleCallExpression(t *testing.T) {
	g := &generator{
		variables:      make(map[string]string),
		constants:      make(map[string]interface{}),
		strategyConfig: NewStrategyConfig(),
	}
	analyzer := NewExpressionAnalyzer(g)

	// Create: ta.sma(close, 20)
	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 20},
		},
	}

	calls := analyzer.FindNestedCalls(call)

	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}

	if calls[0].FuncName != "ta.sma" {
		t.Errorf("Expected funcName 'ta.sma', got %q", calls[0].FuncName)
	}

	if calls[0].ArgHash == "" {
		t.Error("Expected non-empty ArgHash")
	}
}

func TestExpressionAnalyzer_NestedCalls(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	analyzer := NewExpressionAnalyzer(g)

	// Create: rma(max(change(close), 0), 9)
	innerCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "change"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
		},
	}

	midCall := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "max"},
		Arguments: []ast.Expression{innerCall, &ast.Literal{Value: 0}},
	}

	outerCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "rma"},
		},
		Arguments: []ast.Expression{midCall, &ast.Literal{Value: 9}},
	}

	calls := analyzer.FindNestedCalls(outerCall)

	if len(calls) != 3 {
		t.Fatalf("Expected 3 calls, got %d", len(calls))
	}

	expectedFuncs := []string{"ta.rma", "max", "ta.change"}
	for i, call := range calls {
		if call.FuncName != expectedFuncs[i] {
			t.Errorf("Call %d: expected %q, got %q", i, expectedFuncs[i], call.FuncName)
		}
	}
}

func TestExpressionAnalyzer_BinaryExpression(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	analyzer := NewExpressionAnalyzer(g)

	// Create: ta.sma(close, 50) > ta.sma(close, 200)
	leftCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 50},
		},
	}

	rightCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 200},
		},
	}

	binExpr := &ast.BinaryExpression{
		Operator: ">",
		Left:     leftCall,
		Right:    rightCall,
	}

	calls := analyzer.FindNestedCalls(binExpr)

	if len(calls) != 2 {
		t.Fatalf("Expected 2 calls, got %d", len(calls))
	}

	if calls[0].FuncName != "ta.sma" || calls[1].FuncName != "ta.sma" {
		t.Error("Expected both calls to be ta.sma")
	}

	if calls[0].ArgHash == calls[1].ArgHash {
		t.Error("Expected different ArgHash for different periods (50 vs 200)")
	}
}

func TestExpressionAnalyzer_HashUniqueness(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	analyzer := NewExpressionAnalyzer(g)

	// sma(close, 50)
	call1 := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 50},
		},
	}

	// sma(close, 200)
	call2 := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 200},
		},
	}

	hash1 := analyzer.ComputeArgHash(call1)
	hash2 := analyzer.ComputeArgHash(call2)

	if hash1 == hash2 {
		t.Error("Expected different hashes for sma(close,50) vs sma(close,200)")
	}
}

func TestExpressionAnalyzer_HashConsistency(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	analyzer := NewExpressionAnalyzer(g)

	// Create same call twice
	createCall := func() *ast.CallExpression {
		return &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "sma"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 50},
			},
		}
	}

	call1 := createCall()
	call2 := createCall()

	hash1 := analyzer.ComputeArgHash(call1)
	hash2 := analyzer.ComputeArgHash(call2)

	if hash1 != hash2 {
		t.Errorf("Expected consistent hash for identical calls, got %q vs %q", hash1, hash2)
	}
}

func TestExpressionAnalyzer_NoCallsInLiterals(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	analyzer := NewExpressionAnalyzer(g)

	literal := &ast.Literal{Value: 42.0}
	calls := analyzer.FindNestedCalls(literal)

	if len(calls) != 0 {
		t.Errorf("Expected 0 calls from literal, got %d", len(calls))
	}
}

func TestExpressionAnalyzer_ConditionalExpression(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	analyzer := NewExpressionAnalyzer(g)

	// Create: condition ? ta.sma(close, 20) : ta.ema(close, 10)
	smaCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 20},
		},
	}

	emaCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "ema"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 10},
		},
	}

	conditional := &ast.ConditionalExpression{
		Test:       &ast.Literal{Value: true},
		Consequent: smaCall,
		Alternate:  emaCall,
	}

	calls := analyzer.FindNestedCalls(conditional)

	if len(calls) != 2 {
		t.Fatalf("Expected 2 calls, got %d", len(calls))
	}

	funcNames := []string{calls[0].FuncName, calls[1].FuncName}
	hasSma := false
	hasEma := false
	for _, fn := range funcNames {
		if fn == "ta.sma" {
			hasSma = true
		}
		if fn == "ta.ema" {
			hasEma = true
		}
	}
	if !hasSma || !hasEma {
		t.Errorf("Expected ta.sma and ta.ema, got %v", funcNames)
	}
}

// TestExpressionAnalyzer_ContextDetection validates nested call detection algorithm.
// Tests if target call is nested inside parent call (security, fixnan, etc).
func TestExpressionAnalyzer_ContextDetection(t *testing.T) {
	tests := []struct {
		name         string
		expr         ast.Expression
		targetFunc   string
		parentFunc   string
		expectInside bool
	}{
		{
			name: "direct child call in parent",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "tickerid"},
					&ast.Literal{Value: "1D"},
					&ast.CallExpression{
						Callee:    &ast.Identifier{Name: "valuewhen"},
						Arguments: []ast.Expression{},
					},
				},
			},
			targetFunc:   "valuewhen",
			parentFunc:   "security",
			expectInside: true,
		},
		{
			name: "call nested in binary expression within parent",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "tickerid"},
					&ast.Literal{Value: "1D"},
					&ast.BinaryExpression{
						Left: &ast.CallExpression{
							Callee:    &ast.Identifier{Name: "valuewhen"},
							Arguments: []ast.Expression{},
						},
						Right:    &ast.Literal{Value: 1.0},
						Operator: "+",
					},
				},
			},
			targetFunc:   "valuewhen",
			parentFunc:   "security",
			expectInside: true,
		},
		{
			name: "call nested in logical expression within parent",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "tickerid"},
					&ast.Literal{Value: "1D"},
					&ast.LogicalExpression{
						Left: &ast.Identifier{Name: "condition"},
						Right: &ast.CallExpression{
							Callee:    &ast.Identifier{Name: "valuewhen"},
							Arguments: []ast.Expression{},
						},
						Operator: "and",
					},
				},
			},
			targetFunc:   "valuewhen",
			parentFunc:   "security",
			expectInside: true,
		},
		{
			name: "call nested in unary expression within parent",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "tickerid"},
					&ast.Literal{Value: "1D"},
					&ast.UnaryExpression{
						Operator: "not",
						Argument: &ast.CallExpression{
							Callee:    &ast.Identifier{Name: "valuewhen"},
							Arguments: []ast.Expression{},
						},
					},
				},
			},
			targetFunc:   "valuewhen",
			parentFunc:   "security",
			expectInside: true,
		},
		{
			name: "call nested in conditional within parent",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "tickerid"},
					&ast.Literal{Value: "1D"},
					&ast.ConditionalExpression{
						Test: &ast.Identifier{Name: "condition"},
						Consequent: &ast.CallExpression{
							Callee:    &ast.Identifier{Name: "valuewhen"},
							Arguments: []ast.Expression{},
						},
						Alternate: &ast.Identifier{Name: "na"},
					},
				},
			},
			targetFunc:   "valuewhen",
			parentFunc:   "security",
			expectInside: true,
		},
		{
			name: "deeply nested call 5 levels",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "tickerid"},
					&ast.Literal{Value: "1D"},
					&ast.BinaryExpression{
						Left: &ast.ConditionalExpression{
							Test: &ast.Identifier{Name: "cond"},
							Consequent: &ast.BinaryExpression{
								Left: &ast.CallExpression{
									Callee:    &ast.Identifier{Name: "max"},
									Arguments: []ast.Expression{},
								},
								Right: &ast.CallExpression{
									Callee:    &ast.Identifier{Name: "valuewhen"},
									Arguments: []ast.Expression{},
								},
								Operator: "+",
							},
							Alternate: &ast.Literal{Value: 0.0},
						},
						Right:    &ast.Literal{Value: 1.0},
						Operator: "*",
					},
				},
			},
			targetFunc:   "valuewhen",
			parentFunc:   "security",
			expectInside: true,
		},
		{
			name: "call NOT inside parent - standalone",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "valuewhen"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "condition"},
					&ast.Identifier{Name: "high"},
					&ast.Literal{Value: 0.0},
				},
			},
			targetFunc:   "valuewhen",
			parentFunc:   "security",
			expectInside: false,
		},
		{
			name: "call NOT inside parent - in different context",
			expr: &ast.ConditionalExpression{
				Test: &ast.Identifier{Name: "condition"},
				Consequent: &ast.CallExpression{
					Callee:    &ast.Identifier{Name: "valuewhen"},
					Arguments: []ast.Expression{},
				},
				Alternate: &ast.Identifier{Name: "na"},
			},
			targetFunc:   "valuewhen",
			parentFunc:   "security",
			expectInside: false,
		},
		{
			name: "parent with namespace variant",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "request"},
					Property: &ast.Identifier{Name: "security"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "tickerid"},
					&ast.Literal{Value: "1D"},
					&ast.CallExpression{
						Callee:    &ast.Identifier{Name: "valuewhen"},
						Arguments: []ast.Expression{},
					},
				},
			},
			targetFunc:   "valuewhen",
			parentFunc:   "request.security",
			expectInside: true,
		},
		{
			name: "nested parent calls - target in inner parent",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "ticker1"},
					&ast.Literal{Value: "1D"},
					&ast.CallExpression{
						Callee: &ast.Identifier{Name: "security"},
						Arguments: []ast.Expression{
							&ast.Identifier{Name: "ticker2"},
							&ast.Literal{Value: "1H"},
							&ast.CallExpression{
								Callee:    &ast.Identifier{Name: "valuewhen"},
								Arguments: []ast.Expression{},
							},
						},
					},
				},
			},
			targetFunc:   "valuewhen",
			parentFunc:   "security",
			expectInside: true,
		},
		{
			name: "call chain - target inside parent inside another parent",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "tickerid"},
					&ast.Literal{Value: "1D"},
					&ast.CallExpression{
						Callee: &ast.Identifier{Name: "fixnan"},
						Arguments: []ast.Expression{
							&ast.CallExpression{
								Callee:    &ast.Identifier{Name: "valuewhen"},
								Arguments: []ast.Expression{},
							},
						},
					},
				},
			},
			targetFunc:   "valuewhen",
			parentFunc:   "security",
			expectInside: true,
		},
		{
			name: "different inline function - barcolor in security",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "tickerid"},
					&ast.Literal{Value: "1D"},
					&ast.CallExpression{
						Callee:    &ast.Identifier{Name: "barcolor"},
						Arguments: []ast.Expression{},
					},
				},
			},
			targetFunc:   "barcolor",
			parentFunc:   "security",
			expectInside: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &generator{
				variables: make(map[string]string),
			}
			analyzer := NewExpressionAnalyzer(gen)

			calls := analyzer.FindNestedCalls(tt.expr)

			var targetCall *ast.CallExpression
			for _, callInfo := range calls {
				funcName := gen.extractFunctionName(callInfo.Call.Callee)
				if funcName == tt.targetFunc {
					targetCall = callInfo.Call
					break
				}
			}

			if targetCall == nil {
				t.Fatalf("Target function %q not found in expression", tt.targetFunc)
			}

			isInside := analyzer.IsInsideSecurityCall(targetCall, tt.expr)

			if isInside != tt.expectInside {
				t.Errorf("IsInsideSecurityCall() = %v, want %v", isInside, tt.expectInside)
			}
		})
	}
}

func TestExpressionAnalyzer_MultipleCallsInContext(t *testing.T) {
	gen := &generator{
		variables: make(map[string]string),
	}
	analyzer := NewExpressionAnalyzer(gen)

	call1 := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "valuewhen"},
		Arguments: []ast.Expression{&ast.Literal{Value: 1}},
	}
	call2 := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "valuewhen"},
		Arguments: []ast.Expression{&ast.Literal{Value: 2}},
	}

	expr := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "security"},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "tickerid"},
			&ast.Literal{Value: "1D"},
			&ast.BinaryExpression{
				Left:     call1,
				Right:    call2,
				Operator: "+",
			},
		},
	}

	isInside1 := analyzer.IsInsideSecurityCall(call1, expr)
	isInside2 := analyzer.IsInsideSecurityCall(call2, expr)

	if !isInside1 {
		t.Error("First valuewhen should be inside security")
	}
	if !isInside2 {
		t.Error("Second valuewhen should be inside security")
	}
}

func TestExpressionAnalyzer_ContextDetectionEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		expr         ast.Expression
		targetFunc   string
		expectInside bool
	}{
		{
			name: "security with only 2 args - no expression arg",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "tickerid"},
					&ast.Literal{Value: "1D"},
				},
			},
			targetFunc:   "valuewhen",
			expectInside: false,
		},
		{
			name: "target call is security itself",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "tickerid"},
					&ast.Literal{Value: "1D"},
					&ast.Literal{Value: 0},
				},
			},
			targetFunc:   "security",
			expectInside: false,
		},
		{
			name:         "empty expression tree",
			expr:         &ast.Literal{Value: 42},
			targetFunc:   "valuewhen",
			expectInside: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &generator{
				variables: make(map[string]string),
			}
			analyzer := NewExpressionAnalyzer(gen)

			calls := analyzer.FindNestedCalls(tt.expr)

			var targetCall *ast.CallExpression
			for _, callInfo := range calls {
				funcName := gen.extractFunctionName(callInfo.Call.Callee)
				if funcName == tt.targetFunc {
					targetCall = callInfo.Call
					break
				}
			}

			if targetCall == nil && !tt.expectInside {
				return
			}

			if targetCall == nil {
				t.Fatalf("Target function %q not found but was expected", tt.targetFunc)
			}

			isInside := analyzer.IsInsideSecurityCall(targetCall, tt.expr)

			if isInside != tt.expectInside {
				t.Errorf("IsInsideSecurityCall() = %v, want %v", isInside, tt.expectInside)
			}
		})
	}
}
