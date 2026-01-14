package security

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestHistoricalOffsetExtractor_Extract_SimpleSubscript(t *testing.T) {
	extractor := NewHistoricalOffsetExtractor()

	// close[1]
	expr := &ast.MemberExpression{
		Object:   &ast.Identifier{Name: "close"},
		Property: &ast.Literal{Value: 1.0},
	}

	inner, offset := extractor.Extract(expr)

	if offset != 1 {
		t.Errorf("Expected offset 1, got %d", offset)
	}

	if ident, ok := inner.(*ast.Identifier); !ok || ident.Name != "close" {
		t.Errorf("Expected inner expression to be 'close', got %T", inner)
	}
}

func TestHistoricalOffsetExtractor_Extract_NoSubscript(t *testing.T) {
	extractor := NewHistoricalOffsetExtractor()

	// close (no subscript)
	expr := &ast.Identifier{Name: "close"}

	inner, offset := extractor.Extract(expr)

	if offset != 0 {
		t.Errorf("Expected offset 0, got %d", offset)
	}

	if inner != expr {
		t.Errorf("Expected inner to be original expression")
	}
}

func TestHistoricalOffsetExtractor_ExtractRecursive_FixnanPivot(t *testing.T) {
	extractor := NewHistoricalOffsetExtractor()

	// fixnan(pivothigh(5, 5)[1])
	pivotCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "pivothigh"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: 5.0},
			&ast.Literal{Value: 5.0},
		},
	}

	pivotWithOffset := &ast.MemberExpression{
		Object:   pivotCall,
		Property: &ast.Literal{Value: 1.0},
	}

	fixnanCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "fixnan"},
		},
		Arguments: []ast.Expression{pivotWithOffset},
	}

	inner, offset := extractor.ExtractRecursive(fixnanCall)

	if offset != 1 {
		t.Errorf("Expected offset 1, got %d", offset)
	}

	// Should return fixnan(pivothigh(5,5)) without [1]
	innerCall, ok := inner.(*ast.CallExpression)
	if !ok {
		t.Fatalf("Expected CallExpression, got %T", inner)
	}

	if len(innerCall.Arguments) != 1 {
		t.Fatalf("Expected 1 argument, got %d", len(innerCall.Arguments))
	}

	// Argument should be pivothigh(5,5) without subscript
	pivotArg, ok := innerCall.Arguments[0].(*ast.CallExpression)
	if !ok {
		t.Errorf("Expected pivothigh call as argument, got %T", innerCall.Arguments[0])
	}

	if callee, ok := pivotArg.Callee.(*ast.MemberExpression); ok {
		if prop, ok := callee.Property.(*ast.Identifier); !ok || prop.Name != "pivothigh" {
			t.Errorf("Expected pivothigh, got %v", prop.Name)
		}
	}
}

func TestHistoricalOffsetExtractor_ExtractRecursive_DirectSubscript(t *testing.T) {
	extractor := NewHistoricalOffsetExtractor()

	// pivothigh(5, 5)[2]
	pivotCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "pivothigh"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: 5.0},
			&ast.Literal{Value: 5.0},
		},
	}

	pivotWithOffset := &ast.MemberExpression{
		Object:   pivotCall,
		Property: &ast.Literal{Value: 2.0},
	}

	inner, offset := extractor.ExtractRecursive(pivotWithOffset)

	if offset != 2 {
		t.Errorf("Expected offset 2, got %d", offset)
	}

	if _, ok := inner.(*ast.CallExpression); !ok {
		t.Errorf("Expected CallExpression without subscript, got %T", inner)
	}
}

func TestHistoricalOffsetExtractor_ExtractRecursive_NoOffset(t *testing.T) {
	extractor := NewHistoricalOffsetExtractor()

	// sma(close, 20) - no subscript
	smaCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 20.0},
		},
	}

	inner, offset := extractor.ExtractRecursive(smaCall)

	if offset != 0 {
		t.Errorf("Expected offset 0, got %d", offset)
	}

	if inner != smaCall {
		t.Errorf("Expected inner to be original expression")
	}
}
