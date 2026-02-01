package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTACallResolver_ATR_SingleArgument(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Literal{Value: "14"},
		},
	}

	resolved, err := resolver.Resolve("ta.atr", call)
	if err != nil {
		t.Fatalf("Failed to resolve ta.atr(14): %v", err)
	}

	if !resolved.RequiresOHLC {
		t.Error("ta.atr should require implicit OHLC")
	}

	if len(resolved.SeriesArguments) != 0 {
		t.Errorf("ta.atr(14) should have 0 series arguments, got %d", len(resolved.SeriesArguments))
	}

	if len(resolved.ScalarArguments) != 1 {
		t.Errorf("ta.atr(14) should have 1 scalar argument, got %d", len(resolved.ScalarArguments))
	}
}

func TestTACallResolver_Pivothigh_TwoArguments(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Literal{Value: "5"},
			&ast.Literal{Value: "5"},
		},
	}

	resolved, err := resolver.Resolve("ta.pivothigh", call)
	if err != nil {
		t.Fatalf("Failed to resolve ta.pivothigh(5, 5): %v", err)
	}

	if !resolved.DefaultSourceApplied {
		t.Error("ta.pivothigh(5, 5) should apply default source 'high'")
	}

	if resolved.DefaultSourceName != "high" {
		t.Errorf("ta.pivothigh default source = %s, want 'high'", resolved.DefaultSourceName)
	}

	if len(resolved.ScalarArguments) != 2 {
		t.Errorf("ta.pivothigh(5, 5) should have 2 scalar arguments, got %d", len(resolved.ScalarArguments))
	}
}

func TestTACallResolver_Pivothigh_ThreeArguments(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: "5"},
			&ast.Literal{Value: "5"},
		},
	}

	resolved, err := resolver.Resolve("ta.pivothigh", call)
	if err != nil {
		t.Fatalf("Failed to resolve ta.pivothigh(close, 5, 5): %v", err)
	}

	if resolved.DefaultSourceApplied {
		t.Error("ta.pivothigh(close, 5, 5) should not apply default source")
	}

	if len(resolved.SeriesArguments) != 1 {
		t.Errorf("ta.pivothigh(close, 5, 5) should have 1 series argument, got %d", len(resolved.SeriesArguments))
	}

	if len(resolved.ScalarArguments) != 2 {
		t.Errorf("ta.pivothigh(close, 5, 5) should have 2 scalar arguments, got %d", len(resolved.ScalarArguments))
	}
}

func TestTACallResolver_SMA_SingleArgument(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Literal{Value: "14"},
		},
	}

	resolved, err := resolver.Resolve("ta.sma", call)
	if err != nil {
		t.Fatalf("Failed to resolve ta.sma(14): %v", err)
	}

	if !resolved.DefaultSourceApplied {
		t.Error("ta.sma(14) should apply default source 'close'")
	}

	if resolved.DefaultSourceName != "close" {
		t.Errorf("ta.sma default source = %s, want 'close'", resolved.DefaultSourceName)
	}
}

func TestTACallResolver_SMA_TwoArguments(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "open"},
			&ast.Literal{Value: "14"},
		},
	}

	resolved, err := resolver.Resolve("ta.sma", call)
	if err != nil {
		t.Fatalf("Failed to resolve ta.sma(open, 14): %v", err)
	}

	if resolved.DefaultSourceApplied {
		t.Error("ta.sma(open, 14) should not apply default source")
	}

	if len(resolved.SeriesArguments) != 1 {
		t.Errorf("ta.sma(open, 14) should have 1 series argument, got %d", len(resolved.SeriesArguments))
	}

	if len(resolved.ScalarArguments) != 1 {
		t.Errorf("ta.sma(open, 14) should have 1 scalar argument, got %d", len(resolved.ScalarArguments))
	}
}

func TestTACallResolver_UnknownFunction_FallbackPattern(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: "20"},
		},
	}

	resolved, err := resolver.Resolve("ta.unknown", call)
	if err != nil {
		t.Fatalf("Failed to resolve unknown function with 2 args: %v", err)
	}

	if len(resolved.SeriesArguments) != 1 {
		t.Errorf("Fallback pattern should have 1 series argument, got %d", len(resolved.SeriesArguments))
	}

	if len(resolved.ScalarArguments) != 1 {
		t.Errorf("Fallback pattern should have 1 scalar argument, got %d", len(resolved.ScalarArguments))
	}
}

func TestTACallResolver_UnknownFunction_InvalidArgCount(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Literal{Value: "14"},
		},
	}

	_, err := resolver.Resolve("ta.unknown", call)
	if err == nil {
		t.Error("Unknown function with 1 argument should fail")
	}
}

func TestTACallResolver_Change_SingleArgument(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
		},
	}

	resolved, err := resolver.Resolve("ta.change", call)
	if err != nil {
		t.Fatalf("Failed to resolve ta.change(close): %v", err)
	}

	if len(resolved.SeriesArguments) != 1 {
		t.Errorf("ta.change(close) should have 1 series argument, got %d", len(resolved.SeriesArguments))
	}
}

func TestTACallResolver_Change_TwoArguments(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: "2"},
		},
	}

	resolved, err := resolver.Resolve("ta.change", call)
	if err != nil {
		t.Fatalf("Failed to resolve ta.change(close, 2): %v", err)
	}

	if len(resolved.SeriesArguments) != 1 {
		t.Errorf("ta.change(close, 2) should have 1 series argument, got %d", len(resolved.SeriesArguments))
	}

	if len(resolved.ScalarArguments) != 1 {
		t.Errorf("ta.change(close, 2) should have 1 scalar argument, got %d", len(resolved.ScalarArguments))
	}
}
