package request

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestPivotDetector_DetectPivotCall_PivotHigh(t *testing.T) {
	detector := NewPivotDetector()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "pivothigh"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: float64(15)},
			&ast.Literal{Value: float64(15)},
		},
	}

	info, detected := detector.DetectPivotCall(call)

	if !detected {
		t.Fatal("Expected pivot call to be detected")
	}

	if info.Type != PivotTypeHigh {
		t.Errorf("Expected PivotTypeHigh, got %v", info.Type)
	}

	if info.LeftBars != 15 {
		t.Errorf("Expected LeftBars=15, got %d", info.LeftBars)
	}

	if info.RightBars != 15 {
		t.Errorf("Expected RightBars=15, got %d", info.RightBars)
	}

	if info.HasOffset {
		t.Error("Expected HasOffset=false for direct call")
	}
}

func TestPivotDetector_DetectPivotCall_WithOffset(t *testing.T) {
	detector := NewPivotDetector()

	memberExpr := &ast.MemberExpression{
		Object: &ast.CallExpression{
			Callee: &ast.Identifier{Name: "pivotlow"},
			Arguments: []ast.Expression{
				&ast.Literal{Value: float64(5)},
				&ast.Literal{Value: float64(5)},
			},
		},
		Property: &ast.Literal{Value: float64(1)},
		Computed: true,
	}

	info, detected := detector.DetectPivotCall(memberExpr)

	if !detected {
		t.Fatal("Expected pivot call to be detected")
	}

	if info.Type != PivotTypeLow {
		t.Errorf("Expected PivotTypeLow, got %v", info.Type)
	}

	if !info.HasOffset {
		t.Error("Expected HasOffset=true for subscripted call")
	}

	if info.Offset != 1 {
		t.Errorf("Expected Offset=1, got %d", info.Offset)
	}
}

func TestPivotDetector_DetectPivotCall_NonPivotFunction(t *testing.T) {
	detector := NewPivotDetector()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(20)},
		},
	}

	_, detected := detector.DetectPivotCall(call)

	if detected {
		t.Error("Expected non-pivot function to not be detected as pivot")
	}
}

func TestPivotDetector_IdentifyPivotType(t *testing.T) {
	detector := NewPivotDetector()

	tests := []struct {
		funcName string
		expected PivotFunctionType
	}{
		{"pivothigh", PivotTypeHigh},
		{"ta.pivothigh", PivotTypeHigh},
		{"pivotlow", PivotTypeLow},
		{"ta.pivotlow", PivotTypeLow},
		{"ta.sma", PivotTypeNone},
		{"unknown", PivotTypeNone},
	}

	for _, tt := range tests {
		result := detector.identifyPivotType(tt.funcName)
		if result != tt.expected {
			t.Errorf("identifyPivotType(%q) = %v, want %v", tt.funcName, result, tt.expected)
		}
	}
}
