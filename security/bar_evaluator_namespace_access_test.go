package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

/*
TestBarEvaluator_NamespaceAccessPatterns validates MemberExpression evaluation
with namespace access (ta.tr, ta.atr) vs subscript access (close[1], sma()[2]).

Edge cases:
- Namespace identifiers (ta.tr, ta.atr, ta.rma) without parentheses
- Mixed namespace + subscript (ta.sma(close, 14)[1])
- Deeply nested namespace access (object.property.subproperty)
- Invalid namespace combinations
*/
func TestBarEvaluator_NamespaceAccessPatterns(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Open: 100, High: 110, Low: 90, Close: 105},
			{Open: 105, High: 115, Low: 95, Close: 110},
			{Open: 110, High: 120, Low: 100, Close: 115},
			{Open: 115, High: 125, Low: 105, Close: 120},
			{Open: 120, High: 130, Low: 110, Close: 125},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name        string
		expr        *ast.MemberExpression
		barIdx      int
		expectError bool
		validate    func(t *testing.T, value float64, err error)
		desc        string
	}{
		{
			name: "ta.tr_namespace_access",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "tr"},
			},
			barIdx:      2,
			expectError: false,
			validate: func(t *testing.T, value float64, err error) {
				if err != nil {
					t.Fatalf("ta.tr should succeed, got error: %v", err)
				}
				if math.IsNaN(value) {
					t.Error("ta.tr should return numeric value, got NaN")
				}
				if value < 0 {
					t.Errorf("True range cannot be negative, got %.2f", value)
				}
				// True Range = max(high-low, |high-prevClose|, |low-prevClose|)
				expectedTR := 20.0 // max(120-100, |120-110|, |100-110|) = max(20, 10, 10) = 20
				if math.Abs(value-expectedTR) > 0.01 {
					t.Errorf("Expected TR %.2f, got %.2f", expectedTR, value)
				}
			},
			desc: "Namespace access to ta.tr without call expression",
		},
		{
			name: "ta.tr_at_first_bar",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "tr"},
			},
			barIdx:      0,
			expectError: false,
			validate: func(t *testing.T, value float64, err error) {
				if err != nil {
					t.Fatalf("ta.tr at bar 0 should succeed, got error: %v", err)
				}
				// At bar 0, no previous close exists - should use high-low
				expectedTR := 20.0 // high(110) - low(90)
				if math.Abs(value-expectedTR) > 0.01 {
					t.Errorf("Expected TR %.2f at bar 0, got %.2f", expectedTR, value)
				}
			},
			desc: "True range at first bar uses high-low only",
		},
		{
			name: "subscript_access_close",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 1.0},
			},
			barIdx:      3,
			expectError: false,
			validate: func(t *testing.T, value float64, err error) {
				if err != nil {
					t.Fatalf("close[1] should succeed, got error: %v", err)
				}
				expectedValue := 115.0 // close[bar 2]
				if value != expectedValue {
					t.Errorf("Expected close[1] = %.2f, got %.2f", expectedValue, value)
				}
			},
			desc: "Subscript access with literal offset",
		},
		{
			name: "invalid_namespace_unknown_property",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "unknownFunction"},
			},
			barIdx:      2,
			expectError: true,
			validate: func(t *testing.T, value float64, err error) {
				if err == nil {
					t.Error("Invalid ta namespace property should error")
				}
			},
			desc: "Unknown namespace property returns error",
		},
		{
			name: "invalid_namespace_non_ta",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "position_size"},
			},
			barIdx:      2,
			expectError: true,
			validate: func(t *testing.T, value float64, err error) {
				if err == nil {
					t.Error("Non-ta namespace should error in bar evaluator")
				}
			},
			desc: "Non-ta namespace returns error",
		},
		{
			name: "nested_call_with_subscript",
			expr: &ast.MemberExpression{
				Object: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "sma"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: 3.0},
					},
				},
				Property: &ast.Literal{Value: 1.0},
			},
			barIdx:      4,
			expectError: false,
			validate: func(t *testing.T, value float64, err error) {
				if err != nil {
					t.Fatalf("ta.sma()[1] should succeed, got error: %v", err)
				}
				if math.IsNaN(value) {
					t.Error("SMA[1] should return numeric value, got NaN")
				}
			},
			desc: "Namespace call with subscript offset works correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := evaluator.evaluateMemberExpressionAtBar(tt.expr, ctx, tt.barIdx)
			tt.validate(t, value, err)
		})
	}
}

/*
TestBarEvaluator_TrueRangeCalculation validates ta.tr algorithm correctness
across all edge cases: first bar, gaps, overlaps, extreme values.
*/
func TestBarEvaluator_TrueRangeCalculation(t *testing.T) {
	tests := []struct {
		name      string
		ctx       *context.Context
		barIdx    int
		expected  float64
		expectNaN bool
		desc      string
	}{
		{
			name: "first_bar_no_previous_close",
			ctx: &context.Context{
				Data: []context.OHLCV{
					{High: 110, Low: 90, Close: 100},
				},
			},
			barIdx:    0,
			expectNaN: true,
			desc:      "First bar returns NaN (ta.tr variable = ta.tr(false))",
		},
		{
			name: "gap_up_from_previous_close",
			ctx: &context.Context{
				Data: []context.OHLCV{
					{High: 110, Low: 90, Close: 100},
					{High: 130, Low: 120, Close: 125}, // Gap up from 100 to 120
				},
			},
			barIdx:   1,
			expected: 30.0, // max(130-120=10, |130-100|=30, |120-100|=20) = 30
			desc:     "Gap up increases true range via high-prevClose",
		},
		{
			name: "gap_down_from_previous_close",
			ctx: &context.Context{
				Data: []context.OHLCV{
					{High: 110, Low: 90, Close: 100},
					{High: 85, Low: 70, Close: 75}, // Gap down from 100 to 85
				},
			},
			barIdx:   1,
			expected: 30.0, // max(85-70=15, |85-100|=15, |70-100|=30) = 30
			desc:     "Gap down increases true range via prevClose-low",
		},
		{
			name: "inside_bar_contained_range",
			ctx: &context.Context{
				Data: []context.OHLCV{
					{High: 120, Low: 80, Close: 100},
					{High: 110, Low: 90, Close: 105}, // Inside previous bar
				},
			},
			barIdx:   1,
			expected: 20.0, // max(110-90=20, |110-100|=10, |90-100|=10) = 20
			desc:     "Inside bar uses high-low when no gap",
		},
		{
			name: "extreme_volatility",
			ctx: &context.Context{
				Data: []context.OHLCV{
					{High: 100, Low: 100, Close: 100},
					{High: 200, Low: 50, Close: 150}, // 100% range expansion
				},
			},
			barIdx:   1,
			expected: 150.0, // max(200-50=150, |200-100|=100, |50-100|=50) = 150
			desc:     "Extreme volatility handled correctly",
		},
		{
			name: "zero_range_doji",
			ctx: &context.Context{
				Data: []context.OHLCV{
					{High: 100, Low: 100, Close: 100},
					{High: 100, Low: 100, Close: 100}, // Perfect doji
				},
			},
			barIdx:   1,
			expected: 0.0, // max(0, 0, 0) = 0
			desc:     "Zero range when all prices equal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := NewStreamingBarEvaluator()

			expr := &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "tr"},
			}

			value, err := evaluator.evaluateMemberExpressionAtBar(expr, tt.ctx, tt.barIdx)

			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tt.desc, err)
			}

			if tt.expectNaN {
				if !math.IsNaN(value) {
					t.Errorf("%s: expected NaN, got %.2f", tt.desc, value)
				}
			} else if math.Abs(value-tt.expected) > 0.01 {
				t.Errorf("%s: expected %.2f, got %.2f", tt.desc, tt.expected, value)
			}
		})
	}
}

/*
TestBarEvaluator_TrFunctionCall validates ta.tr(handle_na) function call evaluation.

The handleNA parameter controls first-bar behaviour:
  - ta.tr(false) or ta.tr() → NaN on bar 0 (default; PineScript ta.tr variable semantics)
  - ta.tr(true)             → High-Low on bar 0 (ATR seed semantics; no prevClose gap correction)

For all bars after the first, both variants return identical full true-range values.
*/
func TestBarEvaluator_TrFunctionCall(t *testing.T) {
	twoBarCtx := &context.Context{
		Data: []context.OHLCV{
			{High: 110, Low: 90, Close: 100},
			{High: 130, Low: 120, Close: 125},
		},
	}
	oneBarCtx := &context.Context{
		Data: []context.OHLCV{
			{High: 110, Low: 90, Close: 100},
		},
	}

	makeTRCall := func(handleNA bool) *ast.CallExpression {
		return &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "tr"},
			},
			Arguments: []ast.Expression{
				&ast.Literal{Value: handleNA},
			},
		}
	}

	tests := []struct {
		name      string
		ctx       *context.Context
		barIdx    int
		handleNA  bool
		expectNaN bool
		expected  float64
	}{
		{
			name: "handleNA_false_first_bar_yields_NaN",
			ctx:  oneBarCtx, barIdx: 0, handleNA: false,
			expectNaN: true,
		},
		{
			name: "handleNA_true_first_bar_yields_HL",
			ctx:  oneBarCtx, barIdx: 0, handleNA: true,
			expected: 20.0,
		},
		{
			name: "handleNA_false_normal_bar_yields_true_range",
			ctx:  twoBarCtx, barIdx: 1, handleNA: false,
			expected: 30.0, // max(130-120=10, |130-100|=30, |120-100|=20) = 30
		},
		{
			name: "handleNA_true_normal_bar_yields_same_true_range",
			ctx:  twoBarCtx, barIdx: 1, handleNA: true,
			expected: 30.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := NewStreamingBarEvaluator()
			call := makeTRCall(tt.handleNA)

			value, err := evaluator.evaluateTRFuncAtBar(call, tt.ctx, tt.barIdx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectNaN {
				if !math.IsNaN(value) {
					t.Errorf("expected NaN, got %.4f", value)
				}
			} else if math.Abs(value-tt.expected) > 0.001 {
				t.Errorf("expected %.4f, got %.4f", tt.expected, value)
			}
		})
	}
}

/*
TestBarEvaluator_TrVariableEqualsHandleNAFalse verifies that the ta.tr variable
(member-expression access) is semantically equivalent to ta.tr(false): both yield
NaN on the first bar and identical full true-range values on all subsequent bars.
*/
func TestBarEvaluator_TrVariableEqualsHandleNAFalse(t *testing.T) {
	tests := []struct {
		name   string
		ctx    *context.Context
		barIdx int
	}{
		{
			name:   "first_bar",
			ctx:    &context.Context{Data: []context.OHLCV{{High: 110, Low: 90, Close: 100}}},
			barIdx: 0,
		},
		{
			name: "second_bar_gap_up",
			ctx: &context.Context{Data: []context.OHLCV{
				{High: 110, Low: 90, Close: 100},
				{High: 130, Low: 120, Close: 125},
			}},
			barIdx: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := NewStreamingBarEvaluator()

			memberExpr := &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "tr"},
			}
			varValue, err := evaluator.evaluateMemberExpressionAtBar(memberExpr, tt.ctx, tt.barIdx)
			if err != nil {
				t.Fatalf("member expr error: %v", err)
			}

			funcCall := &ast.CallExpression{
				Callee: memberExpr,
				Arguments: []ast.Expression{
					&ast.Literal{Value: false},
				},
			}
			funcValue, err := evaluator.evaluateTRFuncAtBar(funcCall, tt.ctx, tt.barIdx)
			if err != nil {
				t.Fatalf("func call error: %v", err)
			}

			bothNaN := math.IsNaN(varValue) && math.IsNaN(funcValue)
			if !bothNaN && varValue != funcValue {
				t.Errorf("ta.tr variable (%.4f) != ta.tr(false) (%.4f)", varValue, funcValue)
			}
		})
	}
}

/*
TestBarEvaluator_MemberExpressionPropertyTypes validates Property field handling
as both *ast.Identifier (namespace) and *ast.Literal (subscript).
*/
func TestBarEvaluator_MemberExpressionPropertyTypes(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 105},
			{Close: 110},
			{Close: 115},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name         string
		buildExpr    func() *ast.MemberExpression
		barIdx       int
		expectError  bool
		validateType func(t *testing.T, expr *ast.MemberExpression)
		desc         string
	}{
		{
			name: "property_as_identifier_namespace",
			buildExpr: func() *ast.MemberExpression {
				return &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "tr"},
				}
			},
			barIdx:      2,
			expectError: false,
			validateType: func(t *testing.T, expr *ast.MemberExpression) {
				if _, ok := expr.Property.(*ast.Identifier); !ok {
					t.Errorf("Property should be *ast.Identifier, got %T", expr.Property)
				}
			},
			desc: "Property as Identifier for namespace access",
		},
		{
			name: "property_as_literal_subscript",
			buildExpr: func() *ast.MemberExpression {
				return &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "close"},
					Property: &ast.Literal{Value: 1.0},
				}
			},
			barIdx:      3,
			expectError: false,
			validateType: func(t *testing.T, expr *ast.MemberExpression) {
				if _, ok := expr.Property.(*ast.Literal); !ok {
					t.Errorf("Property should be *ast.Literal, got %T", expr.Property)
				}
			},
			desc: "Property as Literal for subscript access",
		},
		{
			name: "property_as_literal_float_subscript",
			buildExpr: func() *ast.MemberExpression {
				return &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "close"},
					Property: &ast.Literal{Value: 2.0},
				}
			},
			barIdx:      3,
			expectError: false,
			validateType: func(t *testing.T, expr *ast.MemberExpression) {
				lit, ok := expr.Property.(*ast.Literal)
				if !ok {
					t.Fatalf("Property should be *ast.Literal, got %T", expr.Property)
				}
				if _, ok := lit.Value.(float64); !ok {
					t.Errorf("Literal value should be float64, got %T", lit.Value)
				}
			},
			desc: "Subscript offset as float literal",
		},
		{
			name: "property_as_literal_int_subscript",
			buildExpr: func() *ast.MemberExpression {
				return &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "close"},
					Property: &ast.Literal{Value: int(1)},
				}
			},
			barIdx:      2,
			expectError: true,
			validateType: func(t *testing.T, expr *ast.MemberExpression) {
				lit, ok := expr.Property.(*ast.Literal)
				if !ok {
					t.Fatalf("Property should be *ast.Literal, got %T", expr.Property)
				}
				if _, ok := lit.Value.(int); !ok {
					t.Errorf("Literal value should be int, got %T", lit.Value)
				}
			},
			desc: "Subscript offset as int literal currently unsupported (expects float64)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := tt.buildExpr()
			tt.validateType(t, expr)

			_, err := evaluator.evaluateMemberExpressionAtBar(expr, ctx, tt.barIdx)

			if tt.expectError && err == nil {
				t.Errorf("%s: expected error but got none", tt.desc)
			}
			if !tt.expectError && err != nil {
				t.Errorf("%s: unexpected error: %v", tt.desc, err)
			}
		})
	}
}
