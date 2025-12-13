package request

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestPivotDetector_DetectPivotCall_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		detected bool
		wantType PivotFunctionType
	}{
		{
			name:     "nil_expression",
			expr:     nil,
			detected: false,
		},
		{
			name:     "identifier_only",
			expr:     &ast.Identifier{Name: "pivothigh"},
			detected: false,
		},
		{
			name:     "literal_expression",
			expr:     &ast.Literal{Value: 42.0},
			detected: false,
		},
		{
			name: "binary_expression",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Literal{Value: 1.0},
				Right:    &ast.Literal{Value: 2.0},
			},
			detected: false,
		},
		{
			name: "unary_expression",
			expr: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.Literal{Value: 5.0},
			},
			detected: false,
		},
		{
			name: "non_computed_member_expression",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "obj"},
				Property: &ast.Identifier{Name: "prop"},
				Computed: false,
			},
			detected: false,
		},
		{
			name: "computed_member_non_call_object",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "someArray"},
				Property: &ast.Literal{Value: 0.0},
				Computed: true,
			},
			detected: false,
		},
	}

	detector := NewPivotDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, detected := detector.DetectPivotCall(tt.expr)

			if detected != tt.detected {
				t.Errorf("detected = %v, want %v", detected, tt.detected)
			}

			if tt.detected && info.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", info.Type, tt.wantType)
			}
		})
	}
}

func TestPivotDetector_InvalidArgumentCounts(t *testing.T) {
	tests := []struct {
		name     string
		argCount int
		detected bool
	}{
		{"zero_args", 0, false},
		{"one_arg", 1, false},
		{"two_args", 2, true},
		{"three_args", 3, true},
		{"four_args", 4, false},
		{"five_args", 5, false},
	}

	detector := NewPivotDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := make([]ast.Expression, tt.argCount)
			for i := 0; i < tt.argCount; i++ {
				args[i] = &ast.Literal{Value: float64(5)}
			}

			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "pivothigh"},
				Arguments: args,
			}

			_, detected := detector.DetectPivotCall(call)

			if detected != tt.detected {
				t.Errorf("argCount=%d: detected = %v, want %v", tt.argCount, detected, tt.detected)
			}
		})
	}
}

func TestPivotDetector_ThreeArgumentCall_CustomSource(t *testing.T) {
	detector := NewPivotDetector()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "pivothigh"},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(10)},
			&ast.Literal{Value: float64(10)},
		},
	}

	info, detected := detector.DetectPivotCall(call)

	if !detected {
		t.Fatal("Expected 3-arg pivot call to be detected")
	}

	if info.Type != PivotTypeHigh {
		t.Errorf("Type = %v, want PivotTypeHigh", info.Type)
	}

	if info.Source == nil {
		t.Error("Expected Source to be set for 3-arg call")
	}

	if ident, ok := info.Source.(*ast.Identifier); !ok || ident.Name != "close" {
		t.Errorf("Source = %v, want Identifier{Name: close}", info.Source)
	}

	if info.LeftBars != 10 {
		t.Errorf("LeftBars = %d, want 10", info.LeftBars)
	}

	if info.RightBars != 10 {
		t.Errorf("RightBars = %d, want 10", info.RightBars)
	}
}

func TestPivotDetector_NumericTypes(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected int
	}{
		{"float64", float64(15), 15},
		{"int", int(20), 20},
		{"int64", int64(25), 25},
		{"float64_fractional", float64(10.7), 10},
		{"string", "invalid", 0},
		{"nil", nil, 0},
	}

	detector := NewPivotDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lit := &ast.Literal{Value: tt.value}
			result := detector.extractIntFromLiteral(lit)

			if result != tt.expected {
				t.Errorf("extractIntFromLiteral(%v) = %d, want %d", tt.value, result, tt.expected)
			}
		})
	}
}

func TestPivotDetector_OffsetTypes(t *testing.T) {
	tests := []struct {
		name           string
		offsetValue    interface{}
		expectedOffset int
	}{
		{"offset_1", float64(1), 1},
		{"offset_5", float64(5), 5},
		{"offset_0", float64(0), 0},
		{"offset_negative", float64(-2), -2},
		{"offset_int", int(3), 3},
		{"offset_int64", int64(7), 7},
	}

	detector := NewPivotDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memberExpr := &ast.MemberExpression{
				Object: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "pivothigh"},
					Arguments: []ast.Expression{
						&ast.Literal{Value: float64(5)},
						&ast.Literal{Value: float64(5)},
					},
				},
				Property: &ast.Literal{Value: tt.offsetValue},
				Computed: true,
			}

			info, detected := detector.DetectPivotCall(memberExpr)

			if !detected {
				t.Fatal("Expected pivot call with offset to be detected")
			}

			if !info.HasOffset {
				t.Error("Expected HasOffset=true")
			}

			if info.Offset != tt.expectedOffset {
				t.Errorf("Offset = %d, want %d", info.Offset, tt.expectedOffset)
			}
		})
	}
}

func TestPivotDetector_FunctionNameVariations(t *testing.T) {
	tests := []struct {
		name         string
		callee       ast.Expression
		expectedType PivotFunctionType
		detected     bool
	}{
		{
			name:         "simple_pivothigh",
			callee:       &ast.Identifier{Name: "pivothigh"},
			expectedType: PivotTypeHigh,
			detected:     true,
		},
		{
			name: "ta_dot_pivothigh",
			callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "pivothigh"},
			},
			expectedType: PivotTypeHigh,
			detected:     true,
		},
		{
			name:         "simple_pivotlow",
			callee:       &ast.Identifier{Name: "pivotlow"},
			expectedType: PivotTypeLow,
			detected:     true,
		},
		{
			name: "ta_dot_pivotlow",
			callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "pivotlow"},
			},
			expectedType: PivotTypeLow,
			detected:     true,
		},
		{
			name: "wrong_namespace",
			callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "math"},
				Property: &ast.Identifier{Name: "pivothigh"},
			},
			expectedType: PivotTypeNone,
			detected:     false,
		},
		{
			name: "nested_member",
			callee: &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "global"},
					Property: &ast.Identifier{Name: "ta"},
				},
				Property: &ast.Identifier{Name: "pivothigh"},
			},
			expectedType: PivotTypeNone,
			detected:     false,
		},
	}

	detector := NewPivotDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: tt.callee,
				Arguments: []ast.Expression{
					&ast.Literal{Value: float64(5)},
					&ast.Literal{Value: float64(5)},
				},
			}

			info, detected := detector.DetectPivotCall(call)

			if detected != tt.detected {
				t.Errorf("detected = %v, want %v", detected, tt.detected)
			}

			if tt.detected && info.Type != tt.expectedType {
				t.Errorf("Type = %v, want %v", info.Type, tt.expectedType)
			}
		})
	}
}

func TestPivotDetector_BoundaryValues(t *testing.T) {
	tests := []struct {
		name      string
		leftBars  int
		rightBars int
		detected  bool
	}{
		{"zero_zero", 0, 0, true},
		{"zero_positive", 0, 10, true},
		{"positive_zero", 10, 0, true},
		{"small_values", 1, 1, true},
		{"large_values", 100, 100, true},
		{"asymmetric", 5, 15, true},
	}

	detector := NewPivotDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: "pivothigh"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: float64(tt.leftBars)},
					&ast.Literal{Value: float64(tt.rightBars)},
				},
			}

			info, detected := detector.DetectPivotCall(call)

			if detected != tt.detected {
				t.Errorf("detected = %v, want %v", detected, tt.detected)
			}

			if tt.detected {
				if info.LeftBars != tt.leftBars {
					t.Errorf("LeftBars = %d, want %d", info.LeftBars, tt.leftBars)
				}
				if info.RightBars != tt.rightBars {
					t.Errorf("RightBars = %d, want %d", info.RightBars, tt.rightBars)
				}
			}
		})
	}
}
