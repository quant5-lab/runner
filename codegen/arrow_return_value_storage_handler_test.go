package codegen

import (
	"strings"
	"testing"
)

func TestReturnValueSeriesStorageHandler_GenerateStorageStatements(t *testing.T) {
	tests := []struct {
		name         string
		varNames     []string
		wantContains []string
	}{
		{
			name:         "empty return values",
			varNames:     []string{},
			wantContains: []string{},
		},
		{
			name:     "single return value",
			varNames: []string{"result"},
			wantContains: []string{
				"resultSeries.Set(result)",
			},
		},
		{
			name:     "multiple return values",
			varNames: []string{"ADX", "up", "down"},
			wantContains: []string{
				"ADXSeries.Set(ADX)",
				"upSeries.Set(up)",
				"downSeries.Set(down)",
			},
		},
		{
			name:     "two return values",
			varNames: []string{"plus", "minus"},
			wantContains: []string{
				"plusSeries.Set(plus)",
				"minusSeries.Set(minus)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewReturnValueSeriesStorageHandler("\t")
			got := handler.GenerateStorageStatements(tt.varNames)

			if len(tt.wantContains) == 0 {
				if got != "" {
					t.Errorf("Expected empty string, got %q", got)
				}
				return
			}

			for _, expected := range tt.wantContains {
				if !strings.Contains(got, expected) {
					t.Errorf("Generated code missing expected statement %q\nGot:\n%s", expected, got)
				}
			}
		})
	}
}

func TestReturnValueSeriesStorageHandler_StatementOrder(t *testing.T) {
	handler := NewReturnValueSeriesStorageHandler("\t")
	varNames := []string{"first", "second", "third"}

	got := handler.GenerateStorageStatements(varNames)

	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}

	expectedOrder := []string{
		"firstSeries.Set(first)",
		"secondSeries.Set(second)",
		"thirdSeries.Set(third)",
	}

	for i, expected := range expectedOrder {
		if !strings.Contains(lines[i], expected) {
			t.Errorf("Line %d: expected %q, got %q", i, expected, lines[i])
		}
	}
}

func TestReturnValueSeriesStorageHandler_Indentation(t *testing.T) {
	tests := []struct {
		name        string
		indentation string
		varNames    []string
		wantPrefix  string
	}{
		{
			name:        "tab indentation",
			indentation: "\t",
			varNames:    []string{"value"},
			wantPrefix:  "\t",
		},
		{
			name:        "double tab indentation",
			indentation: "\t\t",
			varNames:    []string{"value"},
			wantPrefix:  "\t\t",
		},
		{
			name:        "four spaces indentation",
			indentation: "    ",
			varNames:    []string{"value"},
			wantPrefix:  "    ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewReturnValueSeriesStorageHandler(tt.indentation)
			got := handler.GenerateStorageStatements(tt.varNames)

			if !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("Expected prefix %q, got: %q", tt.wantPrefix, got[:len(tt.wantPrefix)])
			}
		})
	}
}

func TestReturnValueSeriesStorageHandler_ValidateReturnValueNames(t *testing.T) {
	tests := []struct {
		name     string
		varNames []string
		wantErr  bool
	}{
		{
			name:     "valid simple names",
			varNames: []string{"ADX", "up", "down"},
			wantErr:  false,
		},
		{
			name:     "valid with underscore",
			varNames: []string{"var_name", "_private", "value_123"},
			wantErr:  false,
		},
		{
			name:     "valid mixed case",
			varNames: []string{"myVar", "MyVar", "MYVAR"},
			wantErr:  false,
		},
		{
			name:     "empty name",
			varNames: []string{"valid", ""},
			wantErr:  true,
		},
		{
			name:     "starts with number",
			varNames: []string{"123invalid"},
			wantErr:  true,
		},
		{
			name:     "contains hyphen",
			varNames: []string{"invalid-name"},
			wantErr:  true,
		},
		{
			name:     "contains space",
			varNames: []string{"invalid name"},
			wantErr:  true,
		},
		{
			name:     "contains special chars",
			varNames: []string{"invalid$name"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewReturnValueSeriesStorageHandler("\t")
			err := handler.ValidateReturnValueNames(tt.varNames)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateReturnValueNames() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReturnValueSeriesStorageHandler_EmptyInput(t *testing.T) {
	handler := NewReturnValueSeriesStorageHandler("\t")

	got := handler.GenerateStorageStatements([]string{})
	if got != "" {
		t.Errorf("Expected empty string for empty input, got %q", got)
	}

	got = handler.GenerateStorageStatements(nil)
	if got != "" {
		t.Errorf("Expected empty string for nil input, got %q", got)
	}
}

func TestReturnValueSeriesStorageHandler_LargeInputs(t *testing.T) {
	tests := []struct {
		name     string
		varCount int
	}{
		{"moderate size", 10},
		{"large size", 50},
		{"very large size", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewReturnValueSeriesStorageHandler("\t")

			varNames := make([]string, tt.varCount)
			for i := 0; i < tt.varCount; i++ {
				varNames[i] = "var" + string(rune('A'+i%26))
			}

			got := handler.GenerateStorageStatements(varNames)

			lines := strings.Split(strings.TrimSpace(got), "\n")
			if len(lines) != tt.varCount {
				t.Errorf("Expected %d statements, got %d", tt.varCount, len(lines))
			}

			for i, varName := range varNames {
				expected := varName + "Series.Set(" + varName + ")"
				if !strings.Contains(lines[i], expected) {
					t.Errorf("Line %d missing expected statement %q", i, expected)
				}
			}
		})
	}
}

func TestReturnValueSeriesStorageHandler_VariableNameEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		varNames  []string
		wantValid bool
	}{
		{
			name:      "single character names",
			varNames:  []string{"a", "b", "c"},
			wantValid: true,
		},
		{
			name:      "very long name",
			varNames:  []string{"veryLongVariableNameThatExceedsNormalLengthButStillValidGoIdentifier"},
			wantValid: true,
		},
		{
			name:      "consecutive underscores",
			varNames:  []string{"var__name", "___triple"},
			wantValid: true,
		},
		{
			name:      "leading underscore",
			varNames:  []string{"_private", "_internal"},
			wantValid: true,
		},
		{
			name:      "trailing numbers",
			varNames:  []string{"var1", "var2", "var999"},
			wantValid: true,
		},
		{
			name:      "mixed valid invalid",
			varNames:  []string{"valid", "123invalid"},
			wantValid: false,
		},
		{
			name:      "all caps",
			varNames:  []string{"CONSTANT", "VALUE"},
			wantValid: true,
		},
		{
			name:      "camelCase",
			varNames:  []string{"myVariable", "anotherOne"},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewReturnValueSeriesStorageHandler("\t")
			err := handler.ValidateReturnValueNames(tt.varNames)

			if tt.wantValid && err != nil {
				t.Errorf("Expected valid, got error: %v", err)
			}
			if !tt.wantValid && err == nil {
				t.Error("Expected error for invalid names, got nil")
			}

			if tt.wantValid {
				got := handler.GenerateStorageStatements(tt.varNames)
				for _, varName := range tt.varNames {
					expected := varName + "Series.Set(" + varName + ")"
					if !strings.Contains(got, expected) {
						t.Errorf("Missing expected statement for %q", varName)
					}
				}
			}
		})
	}
}

func TestReturnValueSeriesStorageHandler_IndentationVariations(t *testing.T) {
	tests := []struct {
		name        string
		indentation string
		varNames    []string
	}{
		{
			name:        "no indentation",
			indentation: "",
			varNames:    []string{"result"},
		},
		{
			name:        "single tab",
			indentation: "\t",
			varNames:    []string{"result"},
		},
		{
			name:        "deep indentation",
			indentation: "\t\t\t\t\t\t\t\t\t\t",
			varNames:    []string{"result"},
		},
		{
			name:        "spaces",
			indentation: "  ",
			varNames:    []string{"result"},
		},
		{
			name:        "many spaces",
			indentation: "        ",
			varNames:    []string{"result"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewReturnValueSeriesStorageHandler(tt.indentation)
			got := handler.GenerateStorageStatements(tt.varNames)

			if tt.indentation == "" {
				if strings.HasPrefix(got, "\t") || strings.HasPrefix(got, " ") {
					t.Error("Expected no indentation, but got indented output")
				}
			} else {
				maxLen := len(tt.indentation)
				if len(got) < maxLen {
					maxLen = len(got)
				}
				if !strings.HasPrefix(got, tt.indentation) {
					t.Errorf("Expected prefix %q, got: %q", tt.indentation, got[:maxLen])
				}
			}
		})
	}
}

func TestReturnValueSeriesStorageHandler_StatementIntegrity(t *testing.T) {
	t.Run("single return preserves format", func(t *testing.T) {
		handler := NewReturnValueSeriesStorageHandler("\t")
		got := handler.GenerateStorageStatements([]string{"value"})

		if !strings.Contains(got, "valueSeries.Set(value)") {
			t.Errorf("Statement format corrupted: %q", got)
		}

		if strings.Count(got, "valueSeries.Set") != 1 {
			t.Error("Statement duplicated or missing")
		}
	})

	t.Run("trailing newline consistency", func(t *testing.T) {
		handler := NewReturnValueSeriesStorageHandler("\t")
		got := handler.GenerateStorageStatements([]string{"a", "b", "c"})

		if !strings.HasSuffix(got, "\n") {
			t.Error("Expected trailing newline")
		}

		if strings.HasSuffix(got, "\n\n") {
			t.Error("Unexpected double newline")
		}
	})

	t.Run("no statement leakage", func(t *testing.T) {
		handler := NewReturnValueSeriesStorageHandler("\t")
		got := handler.GenerateStorageStatements([]string{"secure"})

		forbiddenPatterns := []string{
			"undefined",
			"null",
			"Series.Set(Series",
			"Set(Set",
		}

		for _, pattern := range forbiddenPatterns {
			if strings.Contains(got, pattern) {
				t.Errorf("Generated code contains forbidden pattern: %q", pattern)
			}
		}
	})
}
