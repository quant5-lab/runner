package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Validates TA functions require temp vars in variable initializers */
func TestVariableInitCallFilter_TAFunctionsHoisted(t *testing.T) {
	gen := newTestGenerator()
	filter := NewVariableInitCallFilter(
		gen.taRegistry,
		gen.inlineRegistry,
		gen.runtimeOnlyFilter,
		gen.exprAnalyzer,
	)

	initExpr := &ast.BinaryExpression{
		Operator: "+",
		Left: &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "sma"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(14)},
			},
		},
		Right: &ast.Literal{Value: float64(10)},
	}

	nestedCalls := gen.exprAnalyzer.FindNestedCalls(initExpr)
	hoistable := filter.FilterHoistable(nestedCalls, initExpr)

	if len(hoistable) != 1 {
		t.Fatalf("Expected 1 hoistable TA call, got %d", len(hoistable))
	}
	if hoistable[0].FuncName != "ta.sma" {
		t.Errorf("Expected ta.sma, got %s", hoistable[0].FuncName)
	}
}

/* Validates direct init TA calls are NOT hoisted */
func TestVariableInitCallFilter_DirectInitNotHoisted(t *testing.T) {
	gen := newTestGenerator()
	filter := NewVariableInitCallFilter(
		gen.taRegistry,
		gen.inlineRegistry,
		gen.runtimeOnlyFilter,
		gen.exprAnalyzer,
	)

	initExpr := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(14)},
		},
	}

	nestedCalls := gen.exprAnalyzer.FindNestedCalls(initExpr)
	hoistable := filter.FilterHoistable(nestedCalls, initExpr)

	if len(hoistable) != 0 {
		t.Errorf("Expected 0 hoistable calls for direct init, got %d", len(hoistable))
	}
}

/* Validates runtime-only functions are NOT hoisted */
func TestVariableInitCallFilter_RuntimeOnlySkipped(t *testing.T) {
	gen := newTestGenerator()
	filter := NewVariableInitCallFilter(
		gen.taRegistry,
		gen.inlineRegistry,
		gen.runtimeOnlyFilter,
		gen.exprAnalyzer,
	)

	tests := []struct {
		name     string
		funcName string
	}{
		{"fixnan in binary", "fixnan"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initExpr := &ast.BinaryExpression{
				Operator: "+",
				Left: &ast.CallExpression{
					Callee:    &ast.Identifier{Name: tt.funcName},
					Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
				},
				Right: &ast.Literal{Value: float64(10)},
			}

			nestedCalls := gen.exprAnalyzer.FindNestedCalls(initExpr)
			hoistable := filter.FilterHoistable(nestedCalls, initExpr)

			if len(hoistable) != 0 {
				t.Errorf("Expected 0 hoistable calls for runtime-only %s, got %d", tt.funcName, len(hoistable))
			}
		})
	}
}

/* Validates inline-only functions require security context to hoist */
func TestVariableInitCallFilter_InlineOnlySecurityContext(t *testing.T) {
	gen := newTestGenerator()
	filter := NewVariableInitCallFilter(
		gen.taRegistry,
		gen.inlineRegistry,
		gen.runtimeOnlyFilter,
		gen.exprAnalyzer,
	)

	initExpr := &ast.BinaryExpression{
		Operator: "+",
		Left: &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "nz"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		},
		Right: &ast.Literal{Value: float64(10)},
	}

	nestedCalls := gen.exprAnalyzer.FindNestedCalls(initExpr)
	hoistable := filter.FilterHoistable(nestedCalls, initExpr)

	if len(hoistable) != 0 {
		t.Error("Expected nz NOT to be hoisted outside security context")
	}
}

/* Validates math functions with nested TA require temp vars */
func TestVariableInitCallFilter_MathWithNestedTA(t *testing.T) {
	gen := newTestGenerator()
	filter := NewVariableInitCallFilter(
		gen.taRegistry,
		gen.inlineRegistry,
		gen.runtimeOnlyFilter,
		gen.exprAnalyzer,
	)

	initExpr := &ast.BinaryExpression{
		Operator: "*",
		Left: &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "math"},
				Property: &ast.Identifier{Name: "abs"},
			},
			Arguments: []ast.Expression{
				&ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "ema"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: float64(9)},
					},
				},
			},
		},
		Right: &ast.Literal{Value: float64(2)},
	}

	nestedCalls := gen.exprAnalyzer.FindNestedCalls(initExpr)
	hoistable := filter.FilterHoistable(nestedCalls, initExpr)

	if len(hoistable) != 2 {
		t.Fatalf("Expected 2 hoistable calls (ta.ema + math.abs with TA), got %d", len(hoistable))
	}

	funcNames := make(map[string]bool)
	for _, call := range hoistable {
		funcNames[call.FuncName] = true
	}

	if !funcNames["ta.ema"] {
		t.Error("Expected ta.ema to be hoistable")
	}
	if !funcNames["math.abs"] {
		t.Error("Expected math.abs (with nested TA) to be hoistable")
	}
}

/* Validates pure math functions without TA are NOT hoisted */
func TestVariableInitCallFilter_PureMathNotHoisted(t *testing.T) {
	gen := newTestGenerator()
	filter := NewVariableInitCallFilter(
		gen.taRegistry,
		gen.inlineRegistry,
		gen.runtimeOnlyFilter,
		gen.exprAnalyzer,
	)

	initExpr := &ast.BinaryExpression{
		Operator: "+",
		Left: &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "math"},
				Property: &ast.Identifier{Name: "abs"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
			},
		},
		Right: &ast.Literal{Value: float64(10)},
	}

	nestedCalls := gen.exprAnalyzer.FindNestedCalls(initExpr)
	hoistable := filter.FilterHoistable(nestedCalls, initExpr)

	if len(hoistable) != 0 {
		t.Errorf("Expected 0 hoistable calls for pure math.abs, got %d", len(hoistable))
	}
}

/* Validates complex nested TA expressions */
func TestVariableInitCallFilter_ComplexNesting(t *testing.T) {
	gen := newTestGenerator()
	filter := NewVariableInitCallFilter(
		gen.taRegistry,
		gen.inlineRegistry,
		gen.runtimeOnlyFilter,
		gen.exprAnalyzer,
	)

	initExpr := &ast.BinaryExpression{
		Operator: "/",
		Left: &ast.CallExpression{
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
						&ast.Literal{Value: float64(5)},
					},
				},
				&ast.Literal{Value: float64(10)},
			},
		},
		Right: &ast.Literal{Value: float64(2)},
	}

	nestedCalls := gen.exprAnalyzer.FindNestedCalls(initExpr)
	hoistable := filter.FilterHoistable(nestedCalls, initExpr)

	if len(hoistable) != 2 {
		t.Fatalf("Expected 2 hoistable calls (nested ta.ema + ta.sma), got %d", len(hoistable))
	}

	funcNames := make(map[string]bool)
	for _, call := range hoistable {
		funcNames[call.FuncName] = true
	}

	if !funcNames["ta.ema"] {
		t.Error("Expected ta.ema to be hoistable")
	}
	if !funcNames["ta.sma"] {
		t.Error("Expected ta.sma to be hoistable")
	}
}

/* Validates edge cases with empty or nil expressions */
func TestVariableInitCallFilter_EdgeCases(t *testing.T) {
	gen := newTestGenerator()
	filter := NewVariableInitCallFilter(
		gen.taRegistry,
		gen.inlineRegistry,
		gen.runtimeOnlyFilter,
		gen.exprAnalyzer,
	)

	tests := []struct {
		name        string
		nestedCalls []CallInfo
		initExpr    ast.Expression
	}{
		{
			"empty nested calls",
			[]CallInfo{},
			&ast.Literal{Value: float64(42)},
		},
		{
			"nil init expression",
			[]CallInfo{},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hoistable := filter.FilterHoistable(tt.nestedCalls, tt.initExpr)
			if len(hoistable) != 0 {
				t.Errorf("Expected 0 hoistable calls for %s, got %d", tt.name, len(hoistable))
			}
		})
	}
}

/* Validates ternary operator in variable init */
func TestVariableInitCallFilter_ConditionalExpression(t *testing.T) {
	gen := newTestGenerator()
	filter := NewVariableInitCallFilter(
		gen.taRegistry,
		gen.inlineRegistry,
		gen.runtimeOnlyFilter,
		gen.exprAnalyzer,
	)

	initExpr := &ast.ConditionalExpression{
		Test: &ast.BinaryExpression{
			Operator: ">",
			Left:     &ast.Identifier{Name: "close"},
			Right:    &ast.Identifier{Name: "open"},
		},
		Consequent: &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "sma"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(10)},
			},
		},
		Alternate: &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "ema"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(10)},
			},
		},
	}

	nestedCalls := gen.exprAnalyzer.FindNestedCalls(initExpr)
	hoistable := filter.FilterHoistable(nestedCalls, initExpr)

	if len(hoistable) != 2 {
		t.Fatalf("Expected 2 hoistable calls (ta.sma + ta.ema in ternary), got %d", len(hoistable))
	}

	funcNames := make(map[string]bool)
	for _, call := range hoistable {
		funcNames[call.FuncName] = true
	}

	if !funcNames["ta.sma"] {
		t.Error("Expected ta.sma in consequent to be hoistable")
	}
	if !funcNames["ta.ema"] {
		t.Error("Expected ta.ema in alternate to be hoistable")
	}
}

/* Validates multiple TA calls in single variable init */
func TestVariableInitCallFilter_MultipleTACalls(t *testing.T) {
	gen := newTestGenerator()
	filter := NewVariableInitCallFilter(
		gen.taRegistry,
		gen.inlineRegistry,
		gen.runtimeOnlyFilter,
		gen.exprAnalyzer,
	)

	initExpr := &ast.BinaryExpression{
		Operator: "+",
		Left: &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "sma"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(14)},
			},
		},
		Right: &ast.BinaryExpression{
			Operator: "*",
			Left: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "rsi"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: float64(14)},
				},
			},
			Right: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "ema"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: float64(9)},
				},
			},
		},
	}

	nestedCalls := gen.exprAnalyzer.FindNestedCalls(initExpr)
	hoistable := filter.FilterHoistable(nestedCalls, initExpr)

	if len(hoistable) != 3 {
		t.Fatalf("Expected 3 hoistable TA calls (sma + rsi + ema), got %d", len(hoistable))
	}

	funcNames := make(map[string]bool)
	for _, call := range hoistable {
		funcNames[call.FuncName] = true
	}

	expectedFuncs := []string{"ta.sma", "ta.rsi", "ta.ema"}
	for _, fn := range expectedFuncs {
		if !funcNames[fn] {
			t.Errorf("Expected %s to be hoistable", fn)
		}
	}
}
