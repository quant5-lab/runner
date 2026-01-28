package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestArrowTACallSignatureResolver_SingleArgIsLengthPattern(t *testing.T) {
	registry := NewTAFunctionSignatureRegistry()
	resolver := NewArrowTACallSignatureResolver(registry)

	functions := []struct {
		name          string
		defaultSource string
	}{
		{"highest", "high"},
		{"ta.highest", "high"},
		{"lowest", "low"},
		{"ta.lowest", "low"},
		{"highestbars", "high"},
		{"ta.highestbars", "high"},
		{"lowestbars", "low"},
		{"ta.lowestbars", "low"},
	}

	t.Run("single arg resolves to length with default source", func(t *testing.T) {
		for _, fn := range functions {
			t.Run(fn.name, func(t *testing.T) {
				call := &ast.CallExpression{
					Arguments: []ast.Expression{
						&ast.Literal{Value: "10"},
					},
				}

				resolved, err := resolver.ResolveCall(fn.name, call)
				if err != nil {
					t.Fatalf("ResolveCall(%q, single-arg) unexpected error: %v", fn.name, err)
				}
				if resolved == nil {
					t.Fatalf("ResolveCall(%q, single-arg) returned nil", fn.name)
				}

				if resolved.SourceExpr != nil {
					t.Errorf("ResolveCall(%q, single-arg).SourceExpr should be nil, got %T", fn.name, resolved.SourceExpr)
				}

				if resolved.LengthExpr == nil {
					t.Fatalf("ResolveCall(%q, single-arg).LengthExpr is nil", fn.name)
				}

				if !resolved.NeedsDefaultSource {
					t.Errorf("ResolveCall(%q, single-arg).NeedsDefaultSource = false, want true", fn.name)
				}

				if resolved.DefaultSourceName != fn.defaultSource {
					t.Errorf("ResolveCall(%q, single-arg).DefaultSourceName = %q, want %q",
						fn.name, resolved.DefaultSourceName, fn.defaultSource)
				}

				literal, ok := resolved.LengthExpr.(*ast.Literal)
				if !ok {
					t.Fatalf("ResolveCall(%q, single-arg).LengthExpr is not *ast.Literal, got %T", fn.name, resolved.LengthExpr)
				}
				if literal.Value != "10" {
					t.Errorf("ResolveCall(%q, single-arg).LengthExpr.Value = %v, want \"10\"", fn.name, literal.Value)
				}
			})
		}
	})

	t.Run("two args resolves to explicit source and length", func(t *testing.T) {
		for _, fn := range functions {
			t.Run(fn.name, func(t *testing.T) {
				call := &ast.CallExpression{
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: "20"},
					},
				}

				resolved, err := resolver.ResolveCall(fn.name, call)
				if err != nil {
					t.Fatalf("ResolveCall(%q, two-args) unexpected error: %v", fn.name, err)
				}
				if resolved == nil {
					t.Fatalf("ResolveCall(%q, two-args) returned nil", fn.name)
				}

				if resolved.SourceExpr == nil {
					t.Fatalf("ResolveCall(%q, two-args).SourceExpr is nil", fn.name)
				}

				if resolved.LengthExpr == nil {
					t.Fatalf("ResolveCall(%q, two-args).LengthExpr is nil", fn.name)
				}

				if resolved.NeedsDefaultSource {
					t.Errorf("ResolveCall(%q, two-args).NeedsDefaultSource = true, want false", fn.name)
				}

				if resolved.DefaultSourceName != "" {
					t.Errorf("ResolveCall(%q, two-args).DefaultSourceName = %q, want empty string", fn.name, resolved.DefaultSourceName)
				}

				sourceIdent, ok := resolved.SourceExpr.(*ast.Identifier)
				if !ok {
					t.Fatalf("ResolveCall(%q, two-args).SourceExpr is not *ast.Identifier, got %T", fn.name, resolved.SourceExpr)
				}
				if sourceIdent.Name != "close" {
					t.Errorf("ResolveCall(%q, two-args).SourceExpr.Name = %q, want \"close\"", fn.name, sourceIdent.Name)
				}

				lengthLit, ok := resolved.LengthExpr.(*ast.Literal)
				if !ok {
					t.Fatalf("ResolveCall(%q, two-args).LengthExpr is not *ast.Literal, got %T", fn.name, resolved.LengthExpr)
				}
				if lengthLit.Value != "20" {
					t.Errorf("ResolveCall(%q, two-args).LengthExpr.Value = %v, want \"20\"", fn.name, lengthLit.Value)
				}
			})
		}
	})

	t.Run("zero args returns error", func(t *testing.T) {
		for _, fn := range functions {
			t.Run(fn.name, func(t *testing.T) {
				call := &ast.CallExpression{
					Arguments: []ast.Expression{},
				}

				resolved, err := resolver.ResolveCall(fn.name, call)
				if err == nil {
					t.Errorf("ResolveCall(%q, zero-args) expected error, got nil", fn.name)
				}
				if resolved != nil {
					t.Errorf("ResolveCall(%q, zero-args) expected nil result, got %+v", fn.name, resolved)
				}
			})
		}
	})

	t.Run("three+ args returns error", func(t *testing.T) {
		for _, fn := range functions {
			t.Run(fn.name, func(t *testing.T) {
				call := &ast.CallExpression{
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: "10"},
						&ast.Literal{Value: "extra"},
					},
				}

				resolved, err := resolver.ResolveCall(fn.name, call)
				if err == nil {
					t.Errorf("ResolveCall(%q, three-args) expected error, got nil", fn.name)
				}
				if resolved != nil {
					t.Errorf("ResolveCall(%q, three-args) expected nil result, got %+v", fn.name, resolved)
				}
			})
		}
	})
}

func TestArrowTACallSignatureResolver_SingleArgIsSourcePattern(t *testing.T) {
	registry := NewTAFunctionSignatureRegistry()
	resolver := NewArrowTACallSignatureResolver(registry)

	functions := []string{"change", "ta.change"}

	t.Run("single arg resolves to source with default length", func(t *testing.T) {
		for _, fn := range functions {
			t.Run(fn, func(t *testing.T) {
				call := &ast.CallExpression{
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
					},
				}

				resolved, err := resolver.ResolveCall(fn, call)
				if err != nil {
					t.Fatalf("ResolveCall(%q, single-arg) unexpected error: %v", fn, err)
				}
				if resolved == nil {
					t.Fatalf("ResolveCall(%q, single-arg) returned nil", fn)
				}

				if resolved.SourceExpr == nil {
					t.Fatalf("ResolveCall(%q, single-arg).SourceExpr is nil", fn)
				}

				if resolved.LengthExpr == nil {
					t.Fatalf("ResolveCall(%q, single-arg).LengthExpr is nil", fn)
				}

				if resolved.NeedsDefaultSource {
					t.Errorf("ResolveCall(%q, single-arg).NeedsDefaultSource = true, want false", fn)
				}

				sourceIdent, ok := resolved.SourceExpr.(*ast.Identifier)
				if !ok {
					t.Fatalf("ResolveCall(%q, single-arg).SourceExpr is not *ast.Identifier, got %T", fn, resolved.SourceExpr)
				}
				if sourceIdent.Name != "close" {
					t.Errorf("ResolveCall(%q, single-arg).SourceExpr.Name = %q, want \"close\"", fn, sourceIdent.Name)
				}

				lengthLit, ok := resolved.LengthExpr.(*ast.Literal)
				if !ok {
					t.Fatalf("ResolveCall(%q, single-arg).LengthExpr is not *ast.Literal, got %T", fn, resolved.LengthExpr)
				}
				if lengthLit.Value != "1" {
					t.Errorf("ResolveCall(%q, single-arg).LengthExpr.Value = %v, want \"1\" (default)", fn, lengthLit.Value)
				}
			})
		}
	})

	t.Run("two args resolves to explicit source and length", func(t *testing.T) {
		for _, fn := range functions {
			t.Run(fn, func(t *testing.T) {
				call := &ast.CallExpression{
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: "5"},
					},
				}

				resolved, err := resolver.ResolveCall(fn, call)
				if err != nil {
					t.Fatalf("ResolveCall(%q, two-args) unexpected error: %v", fn, err)
				}
				if resolved == nil {
					t.Fatalf("ResolveCall(%q, two-args) returned nil", fn)
				}

				if resolved.SourceExpr == nil {
					t.Fatalf("ResolveCall(%q, two-args).SourceExpr is nil", fn)
				}

				if resolved.LengthExpr == nil {
					t.Fatalf("ResolveCall(%q, two-args).LengthExpr is nil", fn)
				}

				if resolved.NeedsDefaultSource {
					t.Errorf("ResolveCall(%q, two-args).NeedsDefaultSource = true, want false", fn)
				}

				sourceIdent, ok := resolved.SourceExpr.(*ast.Identifier)
				if !ok {
					t.Fatalf("ResolveCall(%q, two-args).SourceExpr is not *ast.Identifier, got %T", fn, resolved.SourceExpr)
				}
				if sourceIdent.Name != "close" {
					t.Errorf("ResolveCall(%q, two-args).SourceExpr.Name = %q, want \"close\"", fn, sourceIdent.Name)
				}

				lengthLit, ok := resolved.LengthExpr.(*ast.Literal)
				if !ok {
					t.Fatalf("ResolveCall(%q, two-args).LengthExpr is not *ast.Literal, got %T", fn, resolved.LengthExpr)
				}
				if lengthLit.Value != "5" {
					t.Errorf("ResolveCall(%q, two-args).LengthExpr.Value = %v, want \"5\"", fn, lengthLit.Value)
				}
			})
		}
	})

	t.Run("zero args returns error", func(t *testing.T) {
		for _, fn := range functions {
			t.Run(fn, func(t *testing.T) {
				call := &ast.CallExpression{
					Arguments: []ast.Expression{},
				}

				resolved, err := resolver.ResolveCall(fn, call)
				if err == nil {
					t.Errorf("ResolveCall(%q, zero-args) expected error, got nil", fn)
				}
				if resolved != nil {
					t.Errorf("ResolveCall(%q, zero-args) expected nil result, got %+v", fn, resolved)
				}
			})
		}
	})

	t.Run("three+ args returns error", func(t *testing.T) {
		for _, fn := range functions {
			t.Run(fn, func(t *testing.T) {
				call := &ast.CallExpression{
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: "1"},
						&ast.Literal{Value: "extra"},
					},
				}

				resolved, err := resolver.ResolveCall(fn, call)
				if err == nil {
					t.Errorf("ResolveCall(%q, three-args) expected error, got nil", fn)
				}
				if resolved != nil {
					t.Errorf("ResolveCall(%q, three-args) expected nil result, got %+v", fn, resolved)
				}
			})
		}
	})
}

func TestArrowTACallSignatureResolver_ExplicitSourceAndLengthPattern(t *testing.T) {
	registry := NewTAFunctionSignatureRegistry()
	resolver := NewArrowTACallSignatureResolver(registry)

	functions := []string{
		"sma", "ta.sma",
		"ema", "ta.ema",
		"rma", "ta.rma",
		"wma", "ta.wma",
		"stdev", "ta.stdev",
		"rsi", "ta.rsi",
	}

	t.Run("two args resolves to explicit source and length", func(t *testing.T) {
		for _, fn := range functions {
			t.Run(fn, func(t *testing.T) {
				call := &ast.CallExpression{
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: "14"},
					},
				}

				resolved, err := resolver.ResolveCall(fn, call)
				if err != nil {
					t.Fatalf("ResolveCall(%q, two-args) unexpected error: %v", fn, err)
				}
				if resolved == nil {
					t.Fatalf("ResolveCall(%q, two-args) returned nil", fn)
				}

				if resolved.SourceExpr == nil {
					t.Fatalf("ResolveCall(%q, two-args).SourceExpr is nil", fn)
				}

				if resolved.LengthExpr == nil {
					t.Fatalf("ResolveCall(%q, two-args).LengthExpr is nil", fn)
				}

				if resolved.NeedsDefaultSource {
					t.Errorf("ResolveCall(%q, two-args).NeedsDefaultSource = true, want false", fn)
				}

				if resolved.DefaultSourceName != "" {
					t.Errorf("ResolveCall(%q, two-args).DefaultSourceName = %q, want empty string", fn, resolved.DefaultSourceName)
				}

				sourceIdent, ok := resolved.SourceExpr.(*ast.Identifier)
				if !ok {
					t.Fatalf("ResolveCall(%q, two-args).SourceExpr is not *ast.Identifier, got %T", fn, resolved.SourceExpr)
				}
				if sourceIdent.Name != "close" {
					t.Errorf("ResolveCall(%q, two-args).SourceExpr.Name = %q, want \"close\"", fn, sourceIdent.Name)
				}

				lengthLit, ok := resolved.LengthExpr.(*ast.Literal)
				if !ok {
					t.Fatalf("ResolveCall(%q, two-args).LengthExpr is not *ast.Literal, got %T", fn, resolved.LengthExpr)
				}
				if lengthLit.Value != "14" {
					t.Errorf("ResolveCall(%q, two-args).LengthExpr.Value = %v, want \"14\"", fn, lengthLit.Value)
				}
			})
		}
	})

	t.Run("zero args returns error", func(t *testing.T) {
		for _, fn := range functions {
			t.Run(fn, func(t *testing.T) {
				call := &ast.CallExpression{
					Arguments: []ast.Expression{},
				}

				resolved, err := resolver.ResolveCall(fn, call)
				if err == nil {
					t.Errorf("ResolveCall(%q, zero-args) expected error, got nil", fn)
				}
				if resolved != nil {
					t.Errorf("ResolveCall(%q, zero-args) expected nil result, got %+v", fn, resolved)
				}
			})
		}
	})

	t.Run("one arg returns error", func(t *testing.T) {
		for _, fn := range functions {
			t.Run(fn, func(t *testing.T) {
				call := &ast.CallExpression{
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
					},
				}

				resolved, err := resolver.ResolveCall(fn, call)
				if err == nil {
					t.Errorf("ResolveCall(%q, one-arg) expected error, got nil", fn)
				}
				if resolved != nil {
					t.Errorf("ResolveCall(%q, one-arg) expected nil result, got %+v", fn, resolved)
				}
			})
		}
	})

	t.Run("three+ args returns error", func(t *testing.T) {
		for _, fn := range functions {
			t.Run(fn, func(t *testing.T) {
				call := &ast.CallExpression{
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: "14"},
						&ast.Literal{Value: "extra"},
					},
				}

				resolved, err := resolver.ResolveCall(fn, call)
				if err == nil {
					t.Errorf("ResolveCall(%q, three-args) expected error, got nil", fn)
				}
				if resolved != nil {
					t.Errorf("ResolveCall(%q, three-args) expected nil result, got %+v", fn, resolved)
				}
			})
		}
	})
}

func TestArrowTACallSignatureResolver_UnknownFunctionHandling(t *testing.T) {
	registry := NewTAFunctionSignatureRegistry()
	resolver := NewArrowTACallSignatureResolver(registry)

	unknownFunctions := []string{
		"unknown_function",
		"custom_indicator",
		"",
		"ta.nonexistent",
		"strategy.entry",
		"plot",
	}

	for _, fn := range unknownFunctions {
		t.Run(fn, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: "10"},
				},
			}

			resolved, err := resolver.ResolveCall(fn, call)
			if err == nil {
				t.Errorf("ResolveCall(%q) expected error for unknown function, got nil", fn)
			}
			if resolved != nil {
				t.Errorf("ResolveCall(%q) expected nil result for unknown function, got %+v", fn, resolved)
			}
			if err != nil && err.Error() != "" {
				expectedSubstring := "unknown TA function"
				if !stringContains(err.Error(), expectedSubstring) {
					t.Errorf("ResolveCall(%q) error message should contain %q, got: %v", fn, expectedSubstring, err)
				}
			}
		})
	}
}

func TestArrowTACallSignatureResolver_ExpressionTypeVariety(t *testing.T) {
	registry := NewTAFunctionSignatureRegistry()
	resolver := NewArrowTACallSignatureResolver(registry)

	t.Run("literal length expressions", func(t *testing.T) {
		literalTypes := []interface{}{
			"10",
			10,
			10.0,
		}

		for i, litVal := range literalTypes {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: litVal},
				},
			}

			resolved, err := resolver.ResolveCall("highest", call)
			if err != nil {
				t.Errorf("Test case %d: ResolveCall with literal %v unexpected error: %v", i, litVal, err)
			}
			if resolved == nil {
				t.Errorf("Test case %d: ResolveCall with literal %v returned nil", i, litVal)
			}
			if resolved != nil && resolved.LengthExpr == nil {
				t.Errorf("Test case %d: ResolveCall with literal %v has nil LengthExpr", i, litVal)
			}
		}
	})

	t.Run("identifier source expressions", func(t *testing.T) {
		identifiers := []string{"close", "high", "low", "open", "volume", "myCustomSeries"}

		for _, ident := range identifiers {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: ident},
					&ast.Literal{Value: "10"},
				},
			}

			resolved, err := resolver.ResolveCall("highest", call)
			if err != nil {
				t.Errorf("ResolveCall with identifier %q unexpected error: %v", ident, err)
			}
			if resolved == nil {
				t.Errorf("ResolveCall with identifier %q returned nil", ident)
			}
			if resolved != nil && resolved.SourceExpr == nil {
				t.Errorf("ResolveCall with identifier %q has nil SourceExpr", ident)
			}
		}
	})

	t.Run("complex expression types preserved", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "high"},
					Operator: "+",
					Right:    &ast.Identifier{Name: "low"},
				},
				&ast.Literal{Value: "5"},
			},
		}

		resolved, err := resolver.ResolveCall("sma", call)
		if err != nil {
			t.Fatalf("ResolveCall with BinaryExpression unexpected error: %v", err)
		}
		if resolved == nil {
			t.Fatal("ResolveCall with BinaryExpression returned nil")
		}

		if _, ok := resolved.SourceExpr.(*ast.BinaryExpression); !ok {
			t.Errorf("ResolveCall should preserve BinaryExpression, got %T", resolved.SourceExpr)
		}
	})
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
