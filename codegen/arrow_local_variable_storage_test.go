package codegen

import (
	"strings"
	"testing"
)

/*
Arrow function local variable storage tests validate the dual-access pattern
(scalar + ForwardSeriesBuffer) across all operation types. Tests are generalized
to cover algorithmic behavior, not specific bug fixes.

Key patterns tested:
1. Variable lifecycle: declaration → reassignment → series storage
2. Operator semantics: Go := (declare) vs = (assign) correctness
3. Expression normalization: integer literal float64() wrapping
4. Dual-access ordering: scalar operation before Series.Set()
5. Edge cases: empty expressions, complex expressions, tuple destructuring
*/

/* TestArrowLocalVariableStorage_VariableOperationTypes validates operator correctness
 * across variable lifecycle. Generalized for any variable operation.
 */
func TestArrowLocalVariableStorage_VariableOperationTypes(t *testing.T) {
	tests := []struct {
		name         string
		opType       VariableOperationType
		varName      string
		exprCode     string
		wantOperator string
		wantCode     string
		description  string
	}{
		{
			name:         "declaration uses Go :=",
			opType:       VariableDeclaration,
			varName:      "result",
			exprCode:     "close * 2",
			wantOperator: ":=",
			wantCode:     "\tresult := close * 2\n",
			description:  "first assignment to variable uses declaration operator",
		},
		{
			name:         "reassignment uses Go =",
			opType:       VariableReassignment,
			varName:      "result",
			exprCode:     "result + 1",
			wantOperator: "=",
			wantCode:     "\tresult = result + 1\n",
			description:  "subsequent assignment uses reassignment operator",
		},
		{
			name:         "complex expression declaration",
			opType:       VariableDeclaration,
			varName:      "average",
			exprCode:     "ta.sma(close, 20)",
			wantOperator: ":=",
			wantCode:     "\taverage := ta.sma(close, 20)\n",
			description:  "declaration works with complex expressions",
		},
		{
			name:         "complex expression reassignment",
			opType:       VariableReassignment,
			varName:      "average",
			exprCode:     "ta.ema(average, 10)",
			wantOperator: "=",
			wantCode:     "\taverage = ta.ema(average, 10)\n",
			description:  "reassignment works with complex expressions",
		},
		{
			name:         "integer literal declaration normalizes",
			opType:       VariableDeclaration,
			varName:      "count",
			exprCode:     "5",
			wantOperator: ":=",
			wantCode:     "\tcount := float64(5)\n",
			description:  "integer literals wrapped in float64() for type consistency",
		},
		{
			name:         "integer literal reassignment normalizes",
			opType:       VariableReassignment,
			varName:      "count",
			exprCode:     "10",
			wantOperator: "=",
			wantCode:     "\tcount = float64(10)\n",
			description:  "reassignment also normalizes integer literals",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewArrowLocalVariableStorage("\t")
			got := storage.GenerateScalarOperation(tt.varName, tt.exprCode, tt.opType)

			if got != tt.wantCode {
				t.Errorf("GenerateScalarOperation() = %q, want %q", got, tt.wantCode)
			}

			if !strings.Contains(got, tt.wantOperator) {
				t.Errorf("Expected operator %q not found in: %q", tt.wantOperator, got)
			}
		})
	}
}

/* TestArrowLocalVariableStorage_DualAccessLifecycle validates scalar+series pattern
 * across complete variable lifecycle. Generalized for any dual-storage variable.
 */
func TestArrowLocalVariableStorage_DualAccessLifecycle(t *testing.T) {
	tests := []struct {
		name               string
		opType             VariableOperationType
		varName            string
		exprCode           string
		wantScalarOperator string
		wantScalarLine     string
		wantSeriesLine     string
		description        string
	}{
		{
			name:               "declaration creates scalar+series",
			opType:             VariableDeclaration,
			varName:            "momentum",
			exprCode:           "close - close[1]",
			wantScalarOperator: ":=",
			wantScalarLine:     "momentum := close - close[1]",
			wantSeriesLine:     "momentumSeries.Set(momentum)",
			description:        "first use creates both scalar and series storage",
		},
		{
			name:               "reassignment updates scalar+series",
			opType:             VariableReassignment,
			varName:            "momentum",
			exprCode:           "momentum * 0.9",
			wantScalarOperator: "=",
			wantScalarLine:     "momentum = momentum * 0.9",
			wantSeriesLine:     "momentumSeries.Set(momentum)",
			description:        "subsequent use updates both scalar and series",
		},
		{
			name:               "integer normalization in dual storage",
			opType:             VariableDeclaration,
			varName:            "threshold",
			exprCode:           "100",
			wantScalarOperator: ":=",
			wantScalarLine:     "threshold := float64(100)",
			wantSeriesLine:     "thresholdSeries.Set(threshold)",
			description:        "integer literals normalized in dual pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewArrowLocalVariableStorage("\t")
			got := storage.GenerateScalarAndSeriesStorage(tt.varName, tt.exprCode, tt.opType)

			if !strings.Contains(got, tt.wantScalarOperator) {
				t.Errorf("Expected operator %q not found in: %q", tt.wantScalarOperator, got)
			}

			if !strings.Contains(got, tt.wantScalarLine) {
				t.Errorf("Missing scalar line %q in:\n%s", tt.wantScalarLine, got)
			}

			if !strings.Contains(got, tt.wantSeriesLine) {
				t.Errorf("Missing series line %q in:\n%s", tt.wantSeriesLine, got)
			}

			lines := strings.Split(strings.TrimSpace(got), "\n")
			if len(lines) != 2 {
				t.Errorf("Expected 2 lines (scalar + series), got %d", len(lines))
			}

			// Validate ordering: scalar operation MUST precede series storage
			if !strings.Contains(lines[0], tt.wantScalarOperator) {
				t.Errorf("First line must be scalar operation, got: %s", lines[0])
			}
			if !strings.Contains(lines[1], "Series.Set") {
				t.Errorf("Second line must be series storage, got: %s", lines[1])
			}
		})
	}
}

/* TestArrowLocalVariableStorage_BackwardCompatibility validates wrapper methods
 * maintain existing API contracts. Ensures refactoring didn't break callers.
 */
func TestArrowLocalVariableStorage_BackwardCompatibility(t *testing.T) {
	storage := NewArrowLocalVariableStorage("\t")

	t.Run("GenerateScalarDeclaration wrapper", func(t *testing.T) {
		got := storage.GenerateScalarDeclaration("value", "close * 2")
		want := "\tvalue := close * 2\n"
		if got != want {
			t.Errorf("GenerateScalarDeclaration() = %q, want %q", got, want)
		}
	})

	t.Run("GenerateScalarReassignment wrapper", func(t *testing.T) {
		got := storage.GenerateScalarReassignment("value", "value + 1")
		want := "\tvalue = value + 1\n"
		if got != want {
			t.Errorf("GenerateScalarReassignment() = %q, want %q", got, want)
		}
	})

	t.Run("GenerateDualStorage wrapper", func(t *testing.T) {
		got := storage.GenerateDualStorage("result", "ta.sma(close, 10)")
		if !strings.Contains(got, "result := ta.sma(close, 10)") {
			t.Errorf("GenerateDualStorage() missing scalar declaration")
		}
		if !strings.Contains(got, "resultSeries.Set(result)") {
			t.Errorf("GenerateDualStorage() missing series storage")
		}
	})

	t.Run("GenerateDualReassignment wrapper", func(t *testing.T) {
		got := storage.GenerateDualReassignment("result", "result * 2")
		if !strings.Contains(got, "result = result * 2") {
			t.Errorf("GenerateDualReassignment() missing scalar reassignment")
		}
		if !strings.Contains(got, "resultSeries.Set(result)") {
			t.Errorf("GenerateDualReassignment() missing series storage")
		}
		if strings.Contains(got, ":=") {
			t.Errorf("GenerateDualReassignment() must not use := operator")
		}
	})
}

func TestArrowLocalVariableStorage_GenerateScalarDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		exprCode string
		want     string
	}{
		{
			name:     "simple assignment",
			varName:  "up",
			exprCode: "change(high)",
			want:     "\tup := change(high)\n",
		},
		{
			name:     "complex expression",
			varName:  "truerange",
			exprCode: "rma(tr, len)",
			want:     "\ttruerange := rma(tr, len)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewArrowLocalVariableStorage("\t")
			got := storage.GenerateScalarDeclaration(tt.varName, tt.exprCode)

			if got != tt.want {
				t.Errorf("GenerateScalarDeclaration() = %q, want %q", got, tt.want)
			}
		})
	}
}

/* TestArrowLocalVariableStorage_SeriesStorageNaming validates series buffer naming
 * convention. Generalized for any variable name pattern.
 */
func TestArrowLocalVariableStorage_GenerateSeriesStorage(t *testing.T) {
	tests := []struct {
		name        string
		varName     string
		want        string
		description string
	}{
		{
			name:        "simple variable name",
			varName:     "up",
			want:        "\tupSeries.Set(up)\n",
			description: "single word variable gets 'Series' suffix",
		},
		{
			name:        "variable with underscore",
			varName:     "true_range",
			want:        "\ttrue_rangeSeries.Set(true_range)\n",
			description: "snake_case preserved in series buffer name",
		},
		{
			name:        "camelCase variable",
			varName:     "smoothedValue",
			want:        "\tsmoothedValueSeries.Set(smoothedValue)\n",
			description: "camelCase preserved in series buffer name",
		},
		{
			name:        "uppercase variable",
			varName:     "ADX",
			want:        "\tADXSeries.Set(ADX)\n",
			description: "uppercase variable gets series suffix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewArrowLocalVariableStorage("\t")
			got := storage.GenerateSeriesStorage(tt.varName)

			if got != tt.want {
				t.Errorf("GenerateSeriesStorage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestArrowLocalVariableStorage_GenerateTupleDualStorage(t *testing.T) {
	tests := []struct {
		name        string
		varNames    []string
		exprCode    string
		want        []string
		lineCount   int
		description string
	}{
		{
			name:     "two variable tuple",
			varNames: []string{"plus", "minus"},
			exprCode: "dirmov(len)",
			want: []string{
				"temp_plus, temp_minus := dirmov(len)",
				"plus := temp_plus",
				"plusSeries.Set(plus)",
				"minus := temp_minus",
				"minusSeries.Set(minus)",
			},
			lineCount:   5,
			description: "binary tuple uses temp vars then assigns to final vars",
		},
		{
			name:     "three variable tuple",
			varNames: []string{"adx", "up", "down"},
			exprCode: "calculateADX(period)",
			want: []string{
				"temp_adx, temp_up, temp_down := calculateADX(period)",
				"adx := temp_adx",
				"adxSeries.Set(adx)",
				"up := temp_up",
				"upSeries.Set(up)",
				"down := temp_down",
				"downSeries.Set(down)",
			},
			lineCount:   7,
			description: "ternary tuple creates 3 temp vars + 3 final vars + 3 series",
		},
		{
			name:     "single variable tuple edge case",
			varNames: []string{"result"},
			exprCode: "compute()",
			want: []string{
				"temp_result := compute()",
				"result := temp_result",
				"resultSeries.Set(result)",
			},
			lineCount:   3,
			description: "single-value tuple (edge case) still uses temp var pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewArrowLocalVariableStorage("\t")
			got := storage.GenerateTupleDualStorage(tt.varNames, tt.exprCode)

			for _, expectedLine := range tt.want {
				if !strings.Contains(got, expectedLine) {
					t.Errorf("GenerateTupleDualStorage() missing expected line: %q\nGot: %q", expectedLine, got)
				}
			}

			gotLines := strings.Split(strings.TrimSpace(got), "\n")
			if len(gotLines) != tt.lineCount {
				t.Errorf("GenerateTupleDualStorage() expected %d lines, got %d", tt.lineCount, len(gotLines))
			}

			// Validate pattern: temp assignment, then N × (var := temp, series.Set)
			if !strings.HasPrefix(strings.TrimSpace(gotLines[0]), "temp_") {
				t.Errorf("First line must be temp variable assignment, got: %s", gotLines[0])
			}
		})
	}
}

/* TestArrowLocalVariableStorage_DualAccessOrderingInvariant validates temporal
 * correctness: scalar must be computed before series storage. Critical for
 * ForwardSeriesBuffer paradigm. Generalized for any dual-access scenario.
 */
func TestArrowLocalVariableStorage_DualAccessPattern(t *testing.T) {
	tests := []struct {
		name        string
		varName     string
		exprCode    string
		description string
	}{
		{
			name:        "simple expression ordering",
			varName:     "up",
			exprCode:    "change(high)",
			description: "scalar declaration precedes series storage",
		},
		{
			name:        "complex expression ordering",
			varName:     "momentum",
			exprCode:    "ta.rsi(close, 14) - 50",
			description: "complex expressions follow same ordering",
		},
		{
			name:        "self-referential expression",
			varName:     "accumulator",
			exprCode:    "accumulator * 0.95",
			description: "self-references require scalar-first pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewArrowLocalVariableStorage("\t")
			code := storage.GenerateDualStorage(tt.varName, tt.exprCode)

			lines := strings.Split(strings.TrimSpace(code), "\n")
			if len(lines) != 2 {
				t.Fatalf("Expected exactly 2 lines, got %d", len(lines))
			}

			// Line 0: scalar operation
			if !strings.Contains(lines[0], tt.varName) || !strings.Contains(lines[0], ":=") {
				t.Errorf("First line must be scalar declaration, got: %s", lines[0])
			}

			// Line 1: series storage
			if !strings.Contains(lines[1], tt.varName+"Series.Set("+tt.varName+")") {
				t.Errorf("Second line must store scalar to series, got: %s", lines[1])
			}

			// Ordering invariant: scalar position < series position
			scalarPos := strings.Index(code, tt.varName+" :=")
			seriesPos := strings.Index(code, tt.varName+"Series.Set")
			if scalarPos == -1 || seriesPos == -1 || scalarPos >= seriesPos {
				t.Errorf("Scalar declaration must precede series storage. Scalar pos=%d, Series pos=%d", scalarPos, seriesPos)
			}
		})
	}
}

/* TestArrowLocalVariableStorage_IndentationConsistency validates formatting
 * across all indentation styles. Generalized for any indent string.
 */
func TestArrowLocalVariableStorage_Indentation(t *testing.T) {
	tests := []struct {
		name        string
		indent      string
		description string
	}{
		{
			name:        "single tab",
			indent:      "\t",
			description: "standard Go indentation",
		},
		{
			name:        "two tabs",
			indent:      "\t\t",
			description: "nested block indentation",
		},
		{
			name:        "four spaces",
			indent:      "    ",
			description: "space-based indentation",
		},
		{
			name:        "no indent",
			indent:      "",
			description: "edge case: root level code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewArrowLocalVariableStorage(tt.indent)

			// Test scalar declaration indentation
			scalarCode := storage.GenerateScalarDeclaration("test", "value")
			if tt.indent != "" && !strings.HasPrefix(scalarCode, tt.indent) {
				t.Errorf("Scalar code should start with indent %q, got: %q", tt.indent, scalarCode)
			}

			// Test series storage indentation
			seriesCode := storage.GenerateSeriesStorage("test")
			if tt.indent != "" && !strings.HasPrefix(seriesCode, tt.indent) {
				t.Errorf("Series code should start with indent %q, got: %q", tt.indent, seriesCode)
			}

			// Test dual storage indentation consistency
			dualCode := storage.GenerateDualStorage("test", "value")
			lines := strings.Split(dualCode, "\n")
			for i, line := range lines {
				if line == "" {
					continue
				}
				if tt.indent != "" && !strings.HasPrefix(line, tt.indent) {
					t.Errorf("Line %d should start with indent %q, got: %q", i, tt.indent, line)
				}
			}
		})
	}
}

/* TestArrowLocalVariableStorage_EdgeCases validates boundary conditions and
 * unusual inputs. Generalized for algorithmic robustness.
 */
func TestArrowLocalVariableStorage_EdgeCases(t *testing.T) {
	storage := NewArrowLocalVariableStorage("\t")

	t.Run("empty expression", func(t *testing.T) {
		got := storage.GenerateScalarDeclaration("empty", "")
		want := "\tempty := \n"
		if got != want {
			t.Errorf("Empty expression: got %q, want %q", got, want)
		}
	})

	t.Run("whitespace-only expression", func(t *testing.T) {
		got := storage.GenerateScalarDeclaration("whitespace", "   ")
		if !strings.Contains(got, "whitespace :=") {
			t.Errorf("Whitespace expression should preserve pattern: %q", got)
		}
	})

	t.Run("expression with newlines", func(t *testing.T) {
		expr := "ta.sma(\n\tclose,\n\t20\n)"
		got := storage.GenerateScalarDeclaration("multiline", expr)
		if !strings.Contains(got, "multiline :=") {
			t.Errorf("Multiline expression should preserve pattern: %q", got)
		}
	})

	t.Run("very long variable name", func(t *testing.T) {
		longName := "thisIsAnExtremelyLongVariableNameThatSomeoneM" +
			"ightActuallyUseInTheirPineScriptCode"
		got := storage.GenerateSeriesStorage(longName)
		expectedSuffix := longName + "Series.Set(" + longName + ")"
		if !strings.Contains(got, expectedSuffix) {
			t.Errorf("Long variable name not handled correctly: %q", got)
		}
	})

	t.Run("special characters in expression", func(t *testing.T) {
		expr := `(close > open ? 1 : -1) * 100.0`
		got := storage.GenerateScalarDeclaration("signal", expr)
		if !strings.Contains(got, expr) {
			t.Errorf("Special characters should be preserved: %q", got)
		}
	})

	t.Run("zero-length tuple", func(t *testing.T) {
		got := storage.GenerateTupleDualStorage([]string{}, "emptyFunc()")
		// Should handle gracefully (likely empty or minimal output)
		if strings.Contains(got, "temp_") && len(strings.Split(strings.TrimSpace(got), "\n")) > 1 {
			t.Errorf("Zero-length tuple should not generate temp vars: %q", got)
		}
	})
}

/* TestArrowLocalVariableStorage_ExpressionNormalization validates integer literal
 * wrapping across all operation types. Generalized for type system correctness.
 */
func TestArrowLocalVariableStorage_ExpressionNormalization(t *testing.T) {
	tests := []struct {
		name        string
		exprCode    string
		wantWrapped bool
		wantExpr    string
		description string
	}{
		{
			name:        "integer literal wrapped",
			exprCode:    "42",
			wantWrapped: true,
			wantExpr:    "float64(42)",
			description: "plain integer becomes float64(N)",
		},
		{
			name:        "negative integer wrapped",
			exprCode:    "-100",
			wantWrapped: true,
			wantExpr:    "float64(-100)",
			description: "negative integers are also wrapped for type consistency",
		},
		{
			name:        "float literal unwrapped",
			exprCode:    "3.14",
			wantWrapped: false,
			wantExpr:    "3.14",
			description: "float literals pass through unchanged",
		},
		{
			name:        "expression with integer unwrapped",
			exprCode:    "value + 10",
			wantWrapped: false,
			wantExpr:    "value + 10",
			description: "integer in expression context not wrapped",
		},
		{
			name:        "zero wrapped",
			exprCode:    "0",
			wantWrapped: true,
			wantExpr:    "float64(0)",
			description: "zero is an integer literal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewArrowLocalVariableStorage("\t")
			got := storage.GenerateScalarDeclaration("test", tt.exprCode)

			if tt.wantWrapped {
				if !strings.Contains(got, tt.wantExpr) {
					t.Errorf("Expected wrapped expression %q in: %q", tt.wantExpr, got)
				}
			} else {
				if strings.Contains(got, "float64(") && !strings.Contains(tt.exprCode, "float64") {
					t.Errorf("Expression should not be wrapped, got: %q", got)
				}
			}
		})
	}
}
