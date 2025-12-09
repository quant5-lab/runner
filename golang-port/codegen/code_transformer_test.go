package codegen

import "testing"

func TestAddNotEqualZeroTransformer_Transform(t *testing.T) {
	transformer := NewAddNotEqualZeroTransformer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple Series access", "priceSeries.GetCurrent()", "priceSeries.GetCurrent() != 0"},
		{"identifier", "enabled", "enabled != 0"},
		{"bar property", "bar.Close", "bar.Close != 0"},
		{"empty string", "", " != 0"},
		{"already has comparison", "price > 100", "price > 100 != 0"},
		{"expression with spaces", "  value  ", "  value   != 0"},
		{"complex expression", "(a + b) * 2", "(a + b) * 2 != 0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := transformer.Transform(tt.input); result != tt.expected {
				t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, result)
			}
		})
	}
}

func TestAddParenthesesTransformer_Transform(t *testing.T) {
	transformer := NewAddParenthesesTransformer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"comparison expression", "price > 100", "(price > 100)"},
		{"boolean conversion", "enabled != 0", "(enabled != 0)"},
		{"logical expression", "a && b", "(a && b)"},
		{"empty string", "", "()"},
		{"already parenthesized", "(expr)", "((expr))"},
		{"complex expression", "a > 10 && b < 20", "(a > 10 && b < 20)"},
		{"single value", "true", "(true)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := transformer.Transform(tt.input); result != tt.expected {
				t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, result)
			}
		})
	}
}

func TestTransformer_Composition(t *testing.T) {
	notEqualZero := NewAddNotEqualZeroTransformer()
	parentheses := NewAddParenthesesTransformer()

	tests := []struct {
		name   string
		input  string
		order1 string
		order2 string
	}{
		{
			name:   "parentheses then != 0",
			input:  "value",
			order1: "(value) != 0",
			order2: "(value != 0)",
		},
		{
			name:   "!= 0 then parentheses",
			input:  "enabled",
			order1: "(enabled) != 0",
			order2: "(enabled != 0)",
		},
		{
			name:   "empty string composition",
			input:  "",
			order1: "() != 0",
			order2: "( != 0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result1 := notEqualZero.Transform(parentheses.Transform(tt.input))
			if result1 != tt.order1 {
				t.Errorf("parentheses→notEqualZero: expected %q, got %q", tt.order1, result1)
			}

			result2 := parentheses.Transform(notEqualZero.Transform(tt.input))
			if result2 != tt.order2 {
				t.Errorf("notEqualZero→parentheses: expected %q, got %q", tt.order2, result2)
			}
		})
	}
}

func TestTransformer_Idempotency(t *testing.T) {
	notEqualZero := NewAddNotEqualZeroTransformer()
	parentheses := NewAddParenthesesTransformer()

	tests := []struct {
		name        string
		transformer CodeTransformer
		input       string
		idempotent  bool
		firstPass   string
		secondPass  string
	}{
		{
			name:        "!= 0 is not idempotent",
			transformer: notEqualZero,
			input:       "value",
			idempotent:  false,
			firstPass:   "value != 0",
			secondPass:  "value != 0 != 0",
		},
		{
			name:        "parentheses is not idempotent",
			transformer: parentheses,
			input:       "expr",
			idempotent:  false,
			firstPass:   "(expr)",
			secondPass:  "((expr))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first := tt.transformer.Transform(tt.input)
			if first != tt.firstPass {
				t.Errorf("first pass: expected %q, got %q", tt.firstPass, first)
			}

			second := tt.transformer.Transform(first)
			if second != tt.secondPass {
				t.Errorf("second pass: expected %q, got %q", tt.secondPass, second)
			}

			if tt.idempotent && first != second {
				t.Errorf("expected idempotent but got %q != %q", first, second)
			}
		})
	}
}
