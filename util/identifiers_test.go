package util

import "testing"

func TestSanitizeGoIdentifier_Keywords(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"len", "len_"},
		{"type", "type_"},
		{"map", "map_"},
		{"var", "var_"},
		{"func", "func_"},
		{"return", "return_"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeGoIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSanitizeGoIdentifier_NonKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"length", "length"},
		{"src", "src"},
		{"myVar", "myVar"},
		{"value123", "value123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeGoIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSanitizeGoIdentifier_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", "_unnamed"},
		{"uppercase keyword", "LEN", "LEN_"},
		{"mixed case keyword", "Type", "Type_"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeGoIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
