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
