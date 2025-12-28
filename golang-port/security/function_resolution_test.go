package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

// TestTAFunctionNameResolution_NamespacedVsDirectForms tests that TA functions
// are recognized regardless of whether they use namespace prefix (ta.func) or direct form (func)
func TestTAFunctionNameResolution_NamespacedVsDirectForms(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Low: 90, Close: 95},
			{High: 105, Low: 92, Close: 100},
			{High: 110, Low: 95, Close: 105},
			{High: 108, Low: 93, Close: 102},
			{High: 103, Low: 88, Close: 98},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	testCases := []struct {
		name         string
		callee       ast.Expression
		arguments    []ast.Expression
		barIdx       int
		expectError  bool
		validateFunc func(float64, error) bool
		desc         string
	}{
		{
			name:   "fixnan_direct_form",
			callee: &ast.Identifier{Name: "fixnan"},
			arguments: []ast.Expression{
				&ast.Literal{Value: math.NaN()},
			},
			barIdx:      0,
			expectError: false,
			validateFunc: func(val float64, err error) bool {
				return err == nil && math.IsNaN(val)
			},
			desc: "Direct fixnan() without namespace",
		},
		{
			name: "fixnan_namespaced_form",
			callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "fixnan"},
			},
			arguments: []ast.Expression{
				&ast.Literal{Value: math.NaN()},
			},
			barIdx:      0,
			expectError: false,
			validateFunc: func(val float64, err error) bool {
				return err == nil && math.IsNaN(val)
			},
			desc: "Namespaced ta.fixnan() form",
		},
		{
			name:   "pivothigh_direct_form",
			callee: &ast.Identifier{Name: "pivothigh"},
			arguments: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: 1.0},
				&ast.Literal{Value: 1.0},
			},
			barIdx:      2,
			expectError: false,
			validateFunc: func(val float64, err error) bool {
				return err == nil // May be NaN or valid pivot
			},
			desc: "Direct pivothigh() without namespace",
		},
		{
			name: "pivothigh_namespaced_form",
			callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "pivothigh"},
			},
			arguments: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: 1.0},
				&ast.Literal{Value: 1.0},
			},
			barIdx:      2,
			expectError: false,
			validateFunc: func(val float64, err error) bool {
				return err == nil
			},
			desc: "Namespaced ta.pivothigh() form",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee:    tc.callee,
				Arguments: tc.arguments,
			}

			var result float64
			var err error

			// Route to appropriate evaluator based on function name
			funcName := extractCallFunctionName(tc.callee)
			switch {
			case funcName == "fixnan" || funcName == "ta.fixnan":
				result, err = evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, call, ctx, tc.barIdx)
			case funcName == "pivothigh" || funcName == "ta.pivothigh":
				result, err = evaluator.evaluatePivotHighAtBar(call, ctx, tc.barIdx)
			default:
				t.Fatalf("Unknown function: %s", funcName)
			}

			if tc.expectError && err == nil {
				t.Errorf("%s: expected error but got none", tc.desc)
			}
			if !tc.expectError && err != nil {
				t.Errorf("%s: unexpected error: %v", tc.desc, err)
			}
			if tc.validateFunc != nil && !tc.validateFunc(result, err) {
				t.Errorf("%s: validation failed for result=%.2f, err=%v", tc.desc, result, err)
			}
		})
	}
}

// TestPivotArgumentVariations_TwoVsThreeArgs tests pivot functions with different argument counts
func TestPivotArgumentVariations_TwoVsThreeArgs(t *testing.T) {
	testCases := []struct {
		name      string
		funcName  string
		callExpr  *ast.CallExpression
		testIdx   int
		wantError bool
		desc      string
	}{
		{
			name:     "pivothigh_3args_explicit_source",
			funcName: "ta.pivothigh",
			callExpr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "pivothigh"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "high"},
					&ast.Literal{Value: 1.0},
					&ast.Literal{Value: 1.0},
				},
			},
			testIdx:   2,
			wantError: false,
			desc:      "3-arg form with explicit 'high' source",
		},
		{
			name:     "pivothigh_2args_implicit_source",
			funcName: "ta.pivothigh",
			callExpr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "pivothigh"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 1.0}, // leftBars
					&ast.Literal{Value: 1.0}, // rightBars
				},
			},
			testIdx:   2,
			wantError: false,
			desc:      "2-arg form defaults to 'high' source",
		},
		{
			name:     "pivotlow_3args_explicit_source",
			funcName: "ta.pivotlow",
			callExpr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "pivotlow"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "low"},
					&ast.Literal{Value: 1.0},
					&ast.Literal{Value: 1.0},
				},
			},
			testIdx:   2,
			wantError: false,
			desc:      "3-arg form with explicit 'low' source",
		},
		{
			name:     "pivotlow_2args_implicit_source",
			funcName: "ta.pivotlow",
			callExpr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "pivotlow"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 1.0}, // leftBars
					&ast.Literal{Value: 1.0}, // rightBars
				},
			},
			testIdx:   2,
			wantError: false,
			desc:      "2-arg form defaults to 'low' source",
		},
		{
			name:     "pivot_1arg_invalid",
			funcName: "ta.pivothigh",
			callExpr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "pivothigh"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 1.0}, // Only 1 arg - invalid
				},
			},
			testIdx:   2,
			wantError: true,
			desc:      "1-arg form should error (insufficient arguments)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, err := extractPivotArguments(tc.callExpr)

			if tc.wantError && err == nil {
				t.Errorf("%s: expected error but got none", tc.desc)
			}
			if !tc.wantError && err != nil {
				t.Errorf("%s: unexpected error: %v", tc.desc, err)
			}
		})
	}
}

// TestPivotArgumentVariations_SourceFieldValidation tests different source field identifiers
func TestPivotArgumentVariations_SourceFieldValidation(t *testing.T) {
	data := []context.OHLCV{
		{High: 100, Low: 90, Close: 95, Open: 92},
		{High: 105, Low: 88, Close: 100, Open: 98},
		{High: 110, Low: 95, Close: 105, Open: 103},
		{High: 108, Low: 92, Close: 102, Open: 100},
		{High: 103, Low: 87, Close: 98, Open: 96},
	}

	ctx := &context.Context{Data: data}
	evaluator := NewStreamingBarEvaluator()

	sourceFields := []struct {
		name      string
		fieldName string
		desc      string
	}{
		{"high", "high", "Standard high field"},
		{"low", "low", "Standard low field"},
		{"close", "close", "Close as pivot source"},
		{"open", "open", "Open as pivot source"},
	}

	for _, sf := range sourceFields {
		t.Run(sf.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "pivothigh"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: sf.fieldName},
					&ast.Literal{Value: 1.0},
					&ast.Literal{Value: 1.0},
				},
			}

			result, err := evaluator.evaluatePivotHighAtBar(call, ctx, 2)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", sf.desc, err)
			}

			// Result may be NaN or valid pivot - both are acceptable
			// This test ensures no errors with valid field names
			_ = result // Suppress unused variable warning
		})
	}
}

// TestIdentifierVsLiteralArgumentResolution tests AST node types in function arguments
func TestIdentifierVsLiteralArgumentResolution(t *testing.T) {
	testCases := []struct {
		name      string
		expr      ast.Expression
		wantError bool
		wantValue *float64
		desc      string
	}{
		{
			name:      "literal_float",
			expr:      &ast.Literal{Value: 15.0},
			wantError: false,
			wantValue: floatPtr(15.0),
			desc:      "Direct float64 literal",
		},
		{
			name:      "literal_int",
			expr:      &ast.Literal{Value: 20},
			wantError: false,
			wantValue: floatPtr(20.0),
			desc:      "Integer literal converted to float64",
		},
		{
			name:      "identifier_leftBars",
			expr:      &ast.Identifier{Name: "leftBars"},
			wantError: false,
			wantValue: floatPtr(15.0),
			desc:      "Identifier 'leftBars' resolved to constant",
		},
		{
			name:      "identifier_rightBars",
			expr:      &ast.Identifier{Name: "rightBars"},
			wantError: false,
			wantValue: floatPtr(15.0),
			desc:      "Identifier 'rightBars' resolved to constant",
		},
		{
			name:      "identifier_unknown",
			expr:      &ast.Identifier{Name: "unknownVar"},
			wantError: true,
			wantValue: nil,
			desc:      "Unknown identifier should error",
		},
		{
			name:      "binary_expression",
			expr:      &ast.BinaryExpression{Operator: "+"},
			wantError: true,
			wantValue: nil,
			desc:      "Non-literal/identifier expression should error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := extractNumberLiteral(tc.expr)

			if tc.wantError && err == nil {
				t.Errorf("%s: expected error but got none", tc.desc)
			}
			if !tc.wantError && err != nil {
				t.Errorf("%s: unexpected error: %v", tc.desc, err)
			}
			if tc.wantValue != nil && result != *tc.wantValue {
				t.Errorf("%s: expected %.2f, got %.2f", tc.desc, *tc.wantValue, result)
			}
		})
	}
}

// TestFixnanWithNestedTAFunctions tests fixnan wrapping various TA function calls
func TestFixnanWithNestedTAFunctions(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Low: 90}, {High: 105, Low: 88}, {High: 110, Low: 95},
			{High: 108, Low: 92}, {High: 103, Low: 87}, {High: 102, Low: 89},
			{High: 107, Low: 91}, {High: 106, Low: 90}, {High: 101, Low: 85},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	testCases := []struct {
		name       string
		nestedFunc ast.Expression
		nestedArgs []ast.Expression
		barIdx     int
		shouldPass bool
		desc       string
	}{
		{
			name: "fixnan_wraps_pivothigh_namespaced",
			nestedFunc: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "pivothigh"},
			},
			nestedArgs: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: 2.0},
				&ast.Literal{Value: 2.0},
			},
			barIdx:     6,
			shouldPass: true,
			desc:       "fixnan(ta.pivothigh(...)) - namespaced form supported",
		},
		{
			name: "fixnan_wraps_pivotlow_namespaced",
			nestedFunc: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "pivotlow"},
			},
			nestedArgs: []ast.Expression{
				&ast.Identifier{Name: "low"},
				&ast.Literal{Value: 2.0},
				&ast.Literal{Value: 2.0},
			},
			barIdx:     6,
			shouldPass: true,
			desc:       "fixnan(ta.pivotlow(...)) - namespaced form supported",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nestedCall := &ast.CallExpression{
				Callee:    tc.nestedFunc,
				Arguments: tc.nestedArgs,
			}

			fixnanCall := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "fixnan"},
				Arguments: []ast.Expression{nestedCall},
			}

			result, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, tc.barIdx)

			if tc.shouldPass && err != nil {
				t.Fatalf("%s: unexpected error: %v", tc.desc, err)
			}
			if !tc.shouldPass && err == nil {
				t.Fatalf("%s: expected error but got none", tc.desc)
			}

			if tc.shouldPass {
				// Result may be NaN or valid pivot - both acceptable
				// Test ensures no errors with nested function forms
				_ = result
			}
		})
	}
}

// TestFixnanWithNestedTAFunctions_Namespaced tests ta.fixnan wrapping various functions
func TestFixnanWithNestedTAFunctions_Namespaced(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Low: 90}, {High: 105, Low: 88}, {High: 110, Low: 95},
			{High: 108, Low: 92}, {High: 103, Low: 87}, {High: 102, Low: 89},
			{High: 107, Low: 91}, {High: 106, Low: 90}, {High: 101, Low: 85},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	testCases := []struct {
		name       string
		nestedFunc ast.Expression
		nestedArgs []ast.Expression
		barIdx     int
		shouldPass bool
		desc       string
	}{
		{
			name: "ta_fixnan_wraps_ta_pivothigh",
			nestedFunc: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "pivothigh"},
			},
			nestedArgs: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: 2.0},
				&ast.Literal{Value: 2.0},
			},
			barIdx:     6,
			shouldPass: true,
			desc:       "ta.fixnan(ta.pivothigh(...)) - both namespaced",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nestedCall := &ast.CallExpression{
				Callee:    tc.nestedFunc,
				Arguments: tc.nestedArgs,
			}

			fixnanCall := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "fixnan"},
				},
				Arguments: []ast.Expression{nestedCall},
			}

			result, err := evaluator.fixnanEvaluator.EvaluateAtBar(evaluator, fixnanCall, ctx, tc.barIdx)

			if tc.shouldPass && err != nil {
				t.Fatalf("%s: unexpected error: %v", tc.desc, err)
			}
			if !tc.shouldPass && err == nil {
				t.Fatalf("%s: expected error but got none", tc.desc)
			}

			if tc.shouldPass {
				// Test ensures namespaced fixnan works correctly
				_ = result
			}
		})
	}
}

// Helper function for test cases
func floatPtr(f float64) *float64 {
	return &f
}
