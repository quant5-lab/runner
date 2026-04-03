package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTupleIndicatorHandler_MACDSupport(t *testing.T) {
	handler := NewTupleIndicatorHandler()
	g := newTestGenerator()
	g.indent = 1

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "macd"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(12)},
			&ast.Literal{Value: float64(26)},
			&ast.Literal{Value: float64(9)},
		},
	}

	varNames := []string{"macd_line", "signal_line", "hist"}

	code, err := handler.GenerateTupleCode(g, varNames, call)
	if err != nil {
		t.Fatalf("GenerateTupleCode failed: %v", err)
	}

	expectedPatterns := []string{
		"/* Runtime ta.macd(12,26,9) */",
		"if i < 25",
		"macd_lineSeries.Set(math.NaN())",
		"signal_lineSeries.Set(math.NaN())",
		"histSeries.Set(math.NaN())",
		"sourceWindow := make([]float64, i+1)",
		"ta.Macd(sourceWindow, 12, 26, 9)",
		"macd_lineSeries.Set(resultArr[len(resultArr)-1])",
		"signal_lineSeries.Set(resultArr2[len(resultArr2)-1])",
		"histSeries.Set(resultArr3[len(resultArr3)-1])",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(code, pattern) {
			t.Errorf("Generated code missing pattern: %q\nGot:\n%s", pattern, code)
		}
	}
}

func TestTupleIndicatorHandler_BBSupport(t *testing.T) {
	handler := NewTupleIndicatorHandler()
	g := newTestGenerator()
	g.indent = 1

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "bb"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(20)},
			&ast.Literal{Value: float64(2.0)},
		},
	}

	varNames := []string{"upper", "middle", "lower"}

	code, err := handler.GenerateTupleCode(g, varNames, call)
	if err != nil {
		t.Fatalf("GenerateTupleCode failed: %v", err)
	}

	expectedPatterns := []string{
		"/* Runtime ta.bb(20,2) */",
		"if i < 19",
		"upperSeries.Set(math.NaN())",
		"middleSeries.Set(math.NaN())",
		"lowerSeries.Set(math.NaN())",
		"ta.BBands(sourceWindow, 20, 2)",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(code, pattern) {
			t.Errorf("Generated code missing pattern: %q\nGot:\n%s", pattern, code)
		}
	}
}

func TestTupleIndicatorHandler_OutputCountValidation(t *testing.T) {
	handler := NewTupleIndicatorHandler()
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "macd"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(12)},
			&ast.Literal{Value: float64(26)},
			&ast.Literal{Value: float64(9)},
		},
	}

	varNames := []string{"macd", "signal"}

	_, err := handler.GenerateTupleCode(g, varNames, call)
	if err == nil {
		t.Error("Expected error for output count mismatch, got nil")
	}

	if !strings.Contains(err.Error(), "expected 3 outputs, got 2") {
		t.Errorf("Expected output count mismatch error, got: %v", err)
	}
}

func TestTupleIndicatorHandler_UnregisteredIndicator(t *testing.T) {
	handler := NewTupleIndicatorHandler()

	if handler.CanHandle("ta.nonexistent") {
		t.Error("CanHandle should return false for unregistered indicator")
	}
}

func TestTupleIndicatorHandler_PineV4Syntax(t *testing.T) {
	handler := NewTupleIndicatorHandler()
	g := newTestGenerator()
	g.indent = 1

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "macd"},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(12)},
			&ast.Literal{Value: float64(26)},
			&ast.Literal{Value: float64(9)},
		},
	}

	varNames := []string{"macd", "signal", "hist"}

	code, err := handler.GenerateTupleCode(g, varNames, call)
	if err != nil {
		t.Fatalf("Pine v4 syntax failed: %v", err)
	}

	if !strings.Contains(code, "ta.Macd(") {
		t.Error("Pine v4 'macd' should map to ta.Macd runtime function")
	}
}
