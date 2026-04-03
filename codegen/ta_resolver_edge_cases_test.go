package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTA_ArgumentCountBoundaries(t *testing.T) {
	resolver := NewArrowTACallSignatureResolver()

	tests := []struct {
		name      string
		functions []string
		argCounts []int
		wantError []bool
	}{
		{
			name:      "length-first pattern (highest/lowest)",
			functions: []string{"highest", "ta.highest", "lowest", "ta.lowest"},
			argCounts: []int{0, 1, 2, 3},
			wantError: []bool{true, false, false, true},
		},
		{
			name:      "source-first pattern (change)",
			functions: []string{"change", "ta.change"},
			argCounts: []int{0, 1, 2, 3},
			wantError: []bool{true, false, false, true},
		},
		{
			name:      "two-arg pattern with overload (sma)",
			functions: []string{"sma", "ta.sma"},
			argCounts: []int{0, 1, 2, 3},
			wantError: []bool{true, false, false, true},
		},
		{
			name:      "implicit OHLC pattern (atr)",
			functions: []string{"ta.atr"},
			argCounts: []int{0, 1, 2},
			wantError: []bool{true, false, true},
		},
		{
			name:      "multi-arg pattern (pivothigh)",
			functions: []string{"ta.pivothigh", "ta.pivotlow"},
			argCounts: []int{0, 1, 2, 3, 4},
			wantError: []bool{true, true, false, false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, fn := range tt.functions {
				for i, argCount := range tt.argCounts {
					args := make([]ast.Expression, argCount)
					for j := 0; j < argCount; j++ {
						args[j] = &ast.Literal{Value: "10"}
					}

					call := &ast.CallExpression{Arguments: args}
					resolved, err := resolver.ResolveCall(fn, call)

					hasError := err != nil
					if hasError != tt.wantError[i] {
						t.Errorf("%s(%d args): error=%v, want error=%v",
							fn, argCount, hasError, tt.wantError[i])
					}

					if !hasError && resolved == nil {
						t.Errorf("%s(%d args): expected non-nil result when no error", fn, argCount)
					}
					if hasError && resolved != nil {
						t.Errorf("%s(%d args): expected nil result when error, got %+v", fn, argCount, resolved)
					}
				}
			}
		})
	}
}

func TestTA_ExpressionTypePreservation(t *testing.T) {
	resolver := NewArrowTACallSignatureResolver()

	tests := []struct {
		name       string
		function   string
		sourceExpr ast.Expression
		lengthExpr ast.Expression
	}{
		{
			name:       "identifier expressions",
			function:   "ta.sma",
			sourceExpr: &ast.Identifier{Name: "mySource"},
			lengthExpr: &ast.Identifier{Name: "myPeriod"},
		},
		{
			name:       "literal expressions",
			function:   "ta.sma",
			sourceExpr: &ast.Literal{Value: 100.5},
			lengthExpr: &ast.Literal{Value: "20"},
		},
		{
			name:     "binary expression source",
			function: "ta.ema",
			sourceExpr: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "high"},
				Operator: "+",
				Right:    &ast.Identifier{Name: "low"},
			},
			lengthExpr: &ast.Literal{Value: "14"},
		},
		{
			name:     "conditional expression source",
			function: "ta.rma",
			sourceExpr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "x"},
					Operator: ">",
					Right:    &ast.Literal{Value: "0"},
				},
				Consequent: &ast.Identifier{Name: "x"},
				Alternate:  &ast.Literal{Value: "0"},
			},
			lengthExpr: &ast.Literal{Value: "10"},
		},
		{
			name:     "call expression source",
			function: "ta.wma",
			sourceExpr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "abs"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "change"},
				},
			},
			lengthExpr: &ast.Literal{Value: "5"},
		},
		{
			name:     "member expression source",
			function: "ta.sma",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "data"},
				Property: &ast.Identifier{Name: "price"},
			},
			lengthExpr: &ast.Literal{Value: "20"},
		},
		{
			name:       "binary expression length",
			function:   "highest",
			sourceExpr: &ast.Identifier{Name: "close"},
			lengthExpr: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "period"},
				Operator: "*",
				Right:    &ast.Literal{Value: "2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{tt.sourceExpr, tt.lengthExpr},
			}

			resolved, err := resolver.ResolveCall(tt.function, call)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resolved == nil {
				t.Fatal("returned nil")
			}

			if resolved.SourceExpr != tt.sourceExpr {
				t.Errorf("SourceExpr not preserved: got %T, want %T",
					resolved.SourceExpr, tt.sourceExpr)
			}
			if resolved.LengthExpr != tt.lengthExpr {
				t.Errorf("LengthExpr not preserved: got %T, want %T",
					resolved.LengthExpr, tt.lengthExpr)
			}
		})
	}
}

func TestTA_DefaultSourceApplication(t *testing.T) {
	resolver := NewArrowTACallSignatureResolver()

	tests := []struct {
		name              string
		function          string
		args              []ast.Expression
		wantDefaultSource bool
		defaultSourceName string
	}{
		{
			name:              "sma one arg applies default source",
			function:          "ta.sma",
			args:              []ast.Expression{&ast.Literal{Value: "14"}},
			wantDefaultSource: true,
			defaultSourceName: "close",
		},
		{
			name:              "sma two args no default",
			function:          "ta.sma",
			args:              []ast.Expression{&ast.Identifier{Name: "open"}, &ast.Literal{Value: "14"}},
			wantDefaultSource: false,
		},
		{
			name:              "highest one arg applies default high",
			function:          "ta.highest",
			args:              []ast.Expression{&ast.Literal{Value: "10"}},
			wantDefaultSource: true,
			defaultSourceName: "high",
		},
		{
			name:              "lowest one arg applies default low",
			function:          "ta.lowest",
			args:              []ast.Expression{&ast.Literal{Value: "10"}},
			wantDefaultSource: true,
			defaultSourceName: "low",
		},
		{
			name:              "pivothigh two args applies default high",
			function:          "ta.pivothigh",
			args:              []ast.Expression{&ast.Literal{Value: "5"}, &ast.Literal{Value: "5"}},
			wantDefaultSource: true,
			defaultSourceName: "high",
		},
		{
			name:              "pivotlow two args applies default low",
			function:          "ta.pivotlow",
			args:              []ast.Expression{&ast.Literal{Value: "3"}, &ast.Literal{Value: "3"}},
			wantDefaultSource: true,
			defaultSourceName: "low",
		},
		{
			name:              "change one arg no default",
			function:          "ta.change",
			args:              []ast.Expression{&ast.Identifier{Name: "volume"}},
			wantDefaultSource: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Arguments: tt.args}
			resolved, err := resolver.ResolveCall(tt.function, call)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resolved == nil {
				t.Fatal("returned nil")
			}

			if resolved.NeedsDefaultSource != tt.wantDefaultSource {
				t.Errorf("NeedsDefaultSource = %v, want %v",
					resolved.NeedsDefaultSource, tt.wantDefaultSource)
			}

			if tt.defaultSourceName != "" && resolved.DefaultSourceName != tt.defaultSourceName {
				t.Errorf("DefaultSourceName = %q, want %q",
					resolved.DefaultSourceName, tt.defaultSourceName)
			}
		})
	}
}

func TestTA_FallbackBehavior(t *testing.T) {
	resolver := NewArrowTACallSignatureResolver()

	t.Run("unknown functions with two args use fallback", func(t *testing.T) {
		unknownFunctions := []string{
			"ta.unknown",
			"custom_indicator",
			"my_ta_function",
			"",
		}

		for _, fn := range unknownFunctions {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: "20"},
				},
			}

			resolved, err := resolver.ResolveCall(fn, call)
			if err != nil {
				t.Errorf("%q with 2 args should use fallback, got error: %v", fn, err)
				continue
			}
			if resolved == nil {
				t.Errorf("%q with 2 args returned nil", fn)
				continue
			}

			if resolved.SourceExpr == nil {
				t.Errorf("%q fallback should have SourceExpr", fn)
			}
			if resolved.LengthExpr == nil {
				t.Errorf("%q fallback should have LengthExpr", fn)
			}
			if resolved.NeedsDefaultSource {
				t.Errorf("%q fallback should not need default source", fn)
			}
		}
	})

	t.Run("unknown functions with invalid arg count error", func(t *testing.T) {
		invalidArgCounts := []int{0, 1, 3, 4, 5}

		for _, argCount := range invalidArgCounts {
			args := make([]ast.Expression, argCount)
			for i := 0; i < argCount; i++ {
				args[i] = &ast.Literal{Value: "10"}
			}

			call := &ast.CallExpression{Arguments: args}
			resolved, err := resolver.ResolveCall("ta.unknown", call)

			if err == nil {
				t.Errorf("unknown function with %d args should error", argCount)
			}
			if resolved != nil {
				t.Errorf("unknown function with %d args should return nil, got %+v", argCount, resolved)
			}
		}
	})

	t.Run("fallback preserves complex expressions", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "high"},
					Operator: "/",
					Right:    &ast.Literal{Value: "2"},
				},
				&ast.ConditionalExpression{
					Test:       &ast.Identifier{Name: "cond"},
					Consequent: &ast.Literal{Value: "10"},
					Alternate:  &ast.Literal{Value: "20"},
				},
			},
		}

		resolved, err := resolver.ResolveCall("custom_func", call)
		if err != nil {
			t.Fatalf("fallback with complex expressions failed: %v", err)
		}

		if _, ok := resolved.SourceExpr.(*ast.BinaryExpression); !ok {
			t.Errorf("fallback should preserve BinaryExpression, got %T", resolved.SourceExpr)
		}
		if _, ok := resolved.LengthExpr.(*ast.ConditionalExpression); !ok {
			t.Errorf("fallback should preserve ConditionalExpression, got %T", resolved.LengthExpr)
		}
	})
}

func TestTA_NamespaceConsistency(t *testing.T) {
	resolver := NewArrowTACallSignatureResolver()

	functionPairs := []struct {
		bare       string
		namespaced string
	}{
		{"highest", "ta.highest"},
		{"lowest", "ta.lowest"},
		{"sma", "ta.sma"},
		{"ema", "ta.ema"},
		{"rma", "ta.rma"},
		{"wma", "ta.wma"},
		{"change", "ta.change"},
	}

	for _, pair := range functionPairs {
		t.Run(pair.bare+" vs "+pair.namespaced, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: "14"},
				},
			}

			resolved1, err1 := resolver.ResolveCall(pair.bare, call)
			resolved2, err2 := resolver.ResolveCall(pair.namespaced, call)

			if (err1 == nil) != (err2 == nil) {
				t.Errorf("error state mismatch: bare=%v, namespaced=%v", err1, err2)
			}

			if resolved1 != nil && resolved2 != nil {
				if resolved1.NeedsDefaultSource != resolved2.NeedsDefaultSource {
					t.Errorf("NeedsDefaultSource mismatch: bare=%v, namespaced=%v",
						resolved1.NeedsDefaultSource, resolved2.NeedsDefaultSource)
				}
				if resolved1.DefaultSourceName != resolved2.DefaultSourceName {
					t.Errorf("DefaultSourceName mismatch: bare=%q, namespaced=%q",
						resolved1.DefaultSourceName, resolved2.DefaultSourceName)
				}
			}
		})
	}
}

func TestTA_ErrorMessageClarity(t *testing.T) {
	resolver := NewArrowTACallSignatureResolver()

	tests := []struct {
		name            string
		function        string
		args            []ast.Expression
		errorSubstrings []string
	}{
		{
			name:     "unknown function invalid args",
			function: "ta.unknown",
			args:     []ast.Expression{&ast.Literal{Value: "10"}},
			errorSubstrings: []string{
				"requires exactly 2 arguments",
				"source",
				"length",
			},
		},
		{
			name:            "zero args for known function",
			function:        "ta.sma",
			args:            []ast.Expression{},
			errorSubstrings: []string{"ta.sma"},
		},
		{
			name:     "too many args",
			function: "ta.ema",
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: "14"},
				&ast.Literal{Value: "extra"},
				&ast.Literal{Value: "more"},
			},
			errorSubstrings: []string{"ta.ema"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Arguments: tt.args}
			_, err := resolver.ResolveCall(tt.function, call)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			errMsg := err.Error()
			for _, substr := range tt.errorSubstrings {
				if !stringContains(errMsg, substr) {
					t.Errorf("error message should contain %q, got: %s", substr, errMsg)
				}
			}
		})
	}
}

func TestTA_NilAndEmptyHandling(t *testing.T) {
	resolver := NewArrowTACallSignatureResolver()

	t.Run("nil arguments array", func(t *testing.T) {
		call := &ast.CallExpression{Arguments: nil}
		_, err := resolver.ResolveCall("ta.sma", call)

		if err == nil {
			t.Error("nil arguments should be treated as zero arguments and error")
		}
	})

	t.Run("empty function name with valid args", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: "14"},
			},
		}

		resolved, err := resolver.ResolveCall("", call)
		if err != nil {
			return
		}
		if resolved == nil {
			t.Error("empty function name with 2 args should use fallback")
		}
	})

	t.Run("arguments with nil expressions", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{nil, &ast.Literal{Value: "14"}},
		}

		_, err := resolver.ResolveCall("ta.sma", call)
		if err != nil {
			return
		}
	})
}
