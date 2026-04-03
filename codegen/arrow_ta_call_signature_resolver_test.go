package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestArrowTACallSignatureResolver_SingleArgIsLengthPattern(t *testing.T) {
	resolver := NewArrowTACallSignatureResolver()

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
	resolver := NewArrowTACallSignatureResolver()

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

func TestArrowTACallSignatureResolver_TwoArgumentPattern(t *testing.T) {
	resolver := NewArrowTACallSignatureResolver()

	tests := []struct {
		name                 string
		functions            []string
		oneArgBehavior       string
		oneArgDefaultSource  string
		twoArgSourcePosition int
		twoArgLengthPosition int
	}{
		{
			name:                 "explicit source and length (SMA family)",
			functions:            []string{"sma", "ta.sma", "ema", "ta.ema", "rma", "ta.rma", "wma", "ta.wma"},
			oneArgBehavior:       "applies_default_source",
			oneArgDefaultSource:  "close",
			twoArgSourcePosition: 0,
			twoArgLengthPosition: 1,
		},
		{
			name:                 "statistical functions",
			functions:            []string{"ta.stdev"},
			oneArgBehavior:       "applies_default_source",
			oneArgDefaultSource:  "close",
			twoArgSourcePosition: 0,
			twoArgLengthPosition: 1,
		},
		{
			name:                 "oscillator functions",
			functions:            []string{"ta.rsi"},
			oneArgBehavior:       "applies_default_source",
			oneArgDefaultSource:  "close",
			twoArgSourcePosition: 0,
			twoArgLengthPosition: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("two args resolves to explicit source and length", func(t *testing.T) {
				for _, fn := range tt.functions {
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
					})
				}
			})

			t.Run("one arg behavior", func(t *testing.T) {
				for _, fn := range tt.functions {
					t.Run(fn, func(t *testing.T) {
						call := &ast.CallExpression{
							Arguments: []ast.Expression{
								&ast.Literal{Value: "14"},
							},
						}

						resolved, err := resolver.ResolveCall(fn, call)

						if tt.oneArgBehavior == "applies_default_source" {
							if err != nil {
								t.Fatalf("ResolveCall(%q, one-arg) should apply default source, got error: %v", fn, err)
							}
							if resolved == nil {
								t.Fatalf("ResolveCall(%q, one-arg) returned nil", fn)
							}
							if !resolved.NeedsDefaultSource {
								t.Errorf("ResolveCall(%q, one-arg).NeedsDefaultSource = false, want true", fn)
							}
							if resolved.DefaultSourceName != tt.oneArgDefaultSource {
								t.Errorf("ResolveCall(%q, one-arg).DefaultSourceName = %q, want %q",
									fn, resolved.DefaultSourceName, tt.oneArgDefaultSource)
							}
						} else {
							if err == nil {
								t.Errorf("ResolveCall(%q, one-arg) expected error, got nil", fn)
							}
							if resolved != nil {
								t.Errorf("ResolveCall(%q, one-arg) expected nil result, got %+v", fn, resolved)
							}
						}
					})
				}
			})

			t.Run("zero args returns error", func(t *testing.T) {
				for _, fn := range tt.functions {
					t.Run(fn, func(t *testing.T) {
						call := &ast.CallExpression{Arguments: []ast.Expression{}}
						resolved, err := resolver.ResolveCall(fn, call)
						if err == nil {
							t.Errorf("ResolveCall(%q, zero-args) expected error, got nil", fn)
						}
						if resolved != nil {
							t.Errorf("ResolveCall(%q, zero-args) expected nil, got %+v", fn, resolved)
						}
					})
				}
			})

			t.Run("three+ args returns error", func(t *testing.T) {
				for _, fn := range tt.functions {
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
							t.Errorf("ResolveCall(%q, three-args) expected nil, got %+v", fn, resolved)
						}
					})
				}
			})
		})
	}
}

func TestArrowTACallSignatureResolver_UnknownFunctionHandling(t *testing.T) {
	resolver := NewArrowTACallSignatureResolver()

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
				t.Errorf("ResolveCall(%q) expected error for unknown function with 1 arg, got nil", fn)
			}
			if resolved != nil {
				t.Errorf("ResolveCall(%q) expected nil result for unknown function with 1 arg, got %+v", fn, resolved)
			}
			if err != nil && err.Error() != "" {
				expectedSubstring := "requires exactly 2 arguments"
				if !stringContains(err.Error(), expectedSubstring) {
					t.Errorf("ResolveCall(%q) error message should contain %q, got: %v", fn, expectedSubstring, err)
				}
			}
		})
	}
}

func TestArrowTACallSignatureResolver_ExpressionTypeVariety(t *testing.T) {
	resolver := NewArrowTACallSignatureResolver()

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

func TestArrowTACallSignatureResolver_UnknownFunctionFallback(t *testing.T) {
	resolver := NewArrowTACallSignatureResolver()

	t.Run("truly unknown functions with two args use fallback", func(t *testing.T) {
		unknownFunctions := []string{
			"ta.cmo",
			"ta.cog",
			"ta.mom",
			"ta.roc",
			"ta.xyz_unknown",
			"custom_indicator",
			"my_ta_function",
		}

		for _, fn := range unknownFunctions {
			t.Run(fn, func(t *testing.T) {
				call := &ast.CallExpression{
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: "14"},
					},
				}

				resolved, err := resolver.ResolveCall(fn, call)
				if err != nil {
					t.Fatalf("unknown function %q with 2 args should not error: %v", fn, err)
				}
				if resolved == nil {
					t.Fatalf("unknown function %q returned nil", fn)
				}
				if resolved.SourceExpr == nil {
					t.Errorf("unknown function %q should have SourceExpr", fn)
				}
				if resolved.LengthExpr == nil {
					t.Errorf("unknown function %q should have LengthExpr", fn)
				}
				if resolved.NeedsDefaultSource {
					t.Errorf("unknown function %q should not need default source", fn)
				}
				if resolved.DefaultSourceName != "" {
					t.Errorf("unknown function %q should have empty DefaultSourceName, got %q", fn, resolved.DefaultSourceName)
				}

				sourceIdent, ok := resolved.SourceExpr.(*ast.Identifier)
				if !ok {
					t.Errorf("unknown function %q SourceExpr should be *ast.Identifier, got %T", fn, resolved.SourceExpr)
				} else if sourceIdent.Name != "close" {
					t.Errorf("unknown function %q SourceExpr.Name = %q, want \"close\"", fn, sourceIdent.Name)
				}

				lengthLit, ok := resolved.LengthExpr.(*ast.Literal)
				if !ok {
					t.Errorf("unknown function %q LengthExpr should be *ast.Literal, got %T", fn, resolved.LengthExpr)
				} else if lengthLit.Value != "14" {
					t.Errorf("unknown function %q LengthExpr.Value = %v, want \"14\"", fn, lengthLit.Value)
				}
			})
		}
	})

	t.Run("complex expressions preserved in fallback", func(t *testing.T) {
		tests := []struct {
			name       string
			sourceExpr ast.Expression
			lengthExpr ast.Expression
		}{
			{
				name:       "binary expression source",
				sourceExpr: &ast.BinaryExpression{Left: &ast.Identifier{Name: "high"}, Operator: "+", Right: &ast.Identifier{Name: "low"}},
				lengthExpr: &ast.Literal{Value: "20"},
			},
			{
				name: "conditional expression source",
				sourceExpr: &ast.ConditionalExpression{
					Test:       &ast.BinaryExpression{Left: &ast.Identifier{Name: "x"}, Operator: ">", Right: &ast.Literal{Value: "0"}},
					Consequent: &ast.Identifier{Name: "x"},
					Alternate:  &ast.Literal{Value: "0"},
				},
				lengthExpr: &ast.Literal{Value: "14"},
			},
			{
				name:       "call expression source",
				sourceExpr: &ast.CallExpression{Callee: &ast.Identifier{Name: "abs"}, Arguments: []ast.Expression{&ast.Identifier{Name: "diff"}}},
				lengthExpr: &ast.Literal{Value: "10"},
			},
			{
				name:       "identifier length",
				sourceExpr: &ast.Identifier{Name: "volume"},
				lengthExpr: &ast.Identifier{Name: "period"},
			},
			{
				name:       "binary expression length",
				sourceExpr: &ast.Identifier{Name: "close"},
				lengthExpr: &ast.BinaryExpression{Left: &ast.Identifier{Name: "len"}, Operator: "*", Right: &ast.Literal{Value: "2"}},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				call := &ast.CallExpression{
					Arguments: []ast.Expression{tt.sourceExpr, tt.lengthExpr},
				}

				resolved, err := resolver.ResolveCall("ta.unknown", call)
				if err != nil {
					t.Fatalf("fallback with complex expressions should not error: %v", err)
				}
				if resolved == nil {
					t.Fatal("fallback returned nil")
				}

				if resolved.SourceExpr != tt.sourceExpr {
					t.Errorf("SourceExpr not preserved: got %T, want %T", resolved.SourceExpr, tt.sourceExpr)
				}
				if resolved.LengthExpr != tt.lengthExpr {
					t.Errorf("LengthExpr not preserved: got %T, want %T", resolved.LengthExpr, tt.lengthExpr)
				}
			})
		}
	})

	t.Run("invalid argument counts return clear errors", func(t *testing.T) {
		tests := []struct {
			name     string
			argCount int
			args     []ast.Expression
		}{
			{
				name:     "zero args",
				argCount: 0,
				args:     []ast.Expression{},
			},
			{
				name:     "one arg",
				argCount: 1,
				args:     []ast.Expression{&ast.Literal{Value: "14"}},
			},
			{
				name:     "three args",
				argCount: 3,
				args:     []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: "14"}, &ast.Literal{Value: "2"}},
			},
			{
				name:     "four args",
				argCount: 4,
				args: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: "14"},
					&ast.Literal{Value: "2"},
					&ast.Literal{Value: "1"},
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				call := &ast.CallExpression{Arguments: tt.args}

				resolved, err := resolver.ResolveCall("ta.unknown", call)
				if err == nil {
					t.Errorf("expected error for %d args, got nil", tt.argCount)
				}
				if resolved != nil {
					t.Errorf("expected nil result for %d args, got %+v", tt.argCount, resolved)
				}
				if err != nil && !stringContains(err.Error(), "requires exactly 2 arguments") {
					t.Errorf("error message should mention 2 arguments requirement, got: %v", err)
				}
				if err != nil && !stringContains(err.Error(), "source, length") {
					t.Errorf("error message should mention source and length, got: %v", err)
				}
			})
		}
	})

	t.Run("fallback does not interfere with registered functions", func(t *testing.T) {
		registeredFunctions := []struct {
			name string
		}{
			{"highest"},
			{"ta.sma"},
			{"ta.ema"},
			{"change"},
		}

		for _, fn := range registeredFunctions {
			t.Run(fn.name, func(t *testing.T) {
				call := &ast.CallExpression{
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: "14"},
					},
				}

				resolved, err := resolver.ResolveCall(fn.name, call)
				if err != nil {
					t.Fatalf("registered function %q should resolve: %v", fn.name, err)
				}
				if resolved == nil {
					t.Fatalf("registered function %q returned nil", fn.name)
				}
			})
		}
	})
}

func TestArrowTACallSignatureResolver_ArgumentCountBoundaries(t *testing.T) {
	resolver := NewArrowTACallSignatureResolver()

	t.Run("zero arguments across all patterns", func(t *testing.T) {
		testFunctions := []string{
			"highest",
			"ta.sma",
			"change",
			"ta.unknown",
		}

		for _, fn := range testFunctions {
			t.Run(fn, func(t *testing.T) {
				call := &ast.CallExpression{Arguments: []ast.Expression{}}
				resolved, err := resolver.ResolveCall(fn, call)

				if err == nil {
					t.Errorf("%q with 0 args should error", fn)
				}
				if resolved != nil {
					t.Errorf("%q with 0 args should return nil, got %+v", fn, resolved)
				}
			})
		}
	})

	t.Run("excessive arguments across all patterns", func(t *testing.T) {
		excessiveArgs := []ast.Expression{
			&ast.Identifier{Name: "arg1"},
			&ast.Identifier{Name: "arg2"},
			&ast.Identifier{Name: "arg3"},
			&ast.Identifier{Name: "arg4"},
			&ast.Identifier{Name: "arg5"},
		}

		testFunctions := []string{
			"highest",
			"ta.sma",
			"change",
			"ta.unknown",
		}

		for _, fn := range testFunctions {
			t.Run(fn, func(t *testing.T) {
				call := &ast.CallExpression{Arguments: excessiveArgs}
				resolved, err := resolver.ResolveCall(fn, call)

				if err == nil {
					t.Errorf("%q with 5 args should error", fn)
				}
				if resolved != nil {
					t.Errorf("%q with 5 args should return nil, got %+v", fn, resolved)
				}
			})
		}
	})
}

func TestArrowTACallSignatureResolver_EdgeCases(t *testing.T) {
	resolver := NewArrowTACallSignatureResolver()

	t.Run("empty function name", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: "14"},
			},
		}

		resolved, err := resolver.ResolveCall("", call)
		if err == nil && resolved != nil {
			if resolved.SourceExpr == nil || resolved.LengthExpr == nil {
				t.Error("fallback should populate both SourceExpr and LengthExpr")
			}
		}
	})

	t.Run("nil expression in arguments", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{nil, &ast.Literal{Value: "14"}},
		}

		_, err := resolver.ResolveCall("ta.sma", call)
		if err != nil {
			return
		}
	})

	t.Run("mixed namespace and bare function names", func(t *testing.T) {
		pairs := [][2]string{
			{"highest", "ta.highest"},
			{"sma", "ta.sma"},
			{"change", "ta.change"},
		}

		for _, pair := range pairs {
			bare, namespaced := pair[0], pair[1]
			t.Run(bare+" vs "+namespaced, func(t *testing.T) {
				call := &ast.CallExpression{
					Arguments: []ast.Expression{
						&ast.Literal{Value: "10"},
					},
				}

				resolved1, err1 := resolver.ResolveCall(bare, call)
				resolved2, err2 := resolver.ResolveCall(namespaced, call)

				if (err1 == nil) != (err2 == nil) {
					t.Errorf("bare and namespaced should have same error state: %v vs %v", err1, err2)
				}

				if resolved1 != nil && resolved2 != nil {
					if resolved1.NeedsDefaultSource != resolved2.NeedsDefaultSource {
						t.Errorf("bare and namespaced should have same NeedsDefaultSource: %v vs %v",
							resolved1.NeedsDefaultSource, resolved2.NeedsDefaultSource)
					}
					if resolved1.DefaultSourceName != resolved2.DefaultSourceName {
						t.Errorf("bare and namespaced should have same DefaultSourceName: %q vs %q",
							resolved1.DefaultSourceName, resolved2.DefaultSourceName)
					}
				}
			})
		}
	})

	t.Run("special characters in function names", func(t *testing.T) {
		specialNames := []string{
			"ta._internal",
			"ta.123invalid",
			"ta.",
			".function",
			"ta..double",
		}

		for _, fn := range specialNames {
			t.Run(fn, func(t *testing.T) {
				call := &ast.CallExpression{
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: "14"},
					},
				}

				_, err := resolver.ResolveCall(fn, call)
				if err != nil {
					return
				}
			})
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
