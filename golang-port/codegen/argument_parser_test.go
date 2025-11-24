package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestArgumentParser_ParseString(t *testing.T) {
	parser := NewArgumentParser()

	tests := []struct {
		name          string
		input         ast.Expression
		expectValid   bool
		expectValue   string
		expectLiteral bool
	}{
		{
			name: "double quoted string",
			input: &ast.Literal{
				Value: `"hello world"`,
			},
			expectValid:   true,
			expectValue:   "hello world",
			expectLiteral: true,
		},
		{
			name: "single quoted string",
			input: &ast.Literal{
				Value: `'0950-1645'`,
			},
			expectValid:   true,
			expectValue:   "0950-1645",
			expectLiteral: true,
		},
		{
			name: "string without quotes",
			input: &ast.Literal{
				Value: "plain",
			},
			expectValid:   true,
			expectValue:   "plain",
			expectLiteral: true,
		},
		{
			name: "non-string literal",
			input: &ast.Literal{
				Value: 123,
			},
			expectValid: false,
		},
		{
			name: "identifier (not a string)",
			input: &ast.Identifier{
				Name: "my_var",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ParseString(tt.input)
			if result.IsValid != tt.expectValid {
				t.Errorf("expected IsValid=%v, got %v", tt.expectValid, result.IsValid)
			}
			if tt.expectValid {
				if result.MustBeString() != tt.expectValue {
					t.Errorf("expected value %q, got %q", tt.expectValue, result.MustBeString())
				}
				if result.IsLiteral != tt.expectLiteral {
					t.Errorf("expected IsLiteral=%v, got %v", tt.expectLiteral, result.IsLiteral)
				}
			}
		})
	}
}

func TestArgumentParser_ParseInt(t *testing.T) {
	parser := NewArgumentParser()

	tests := []struct {
		name        string
		input       ast.Expression
		expectValid bool
		expectValue int
	}{
		{
			name: "integer literal",
			input: &ast.Literal{
				Value: 42,
			},
			expectValid: true,
			expectValue: 42,
		},
		{
			name: "float64 literal (converted to int)",
			input: &ast.Literal{
				Value: float64(20),
			},
			expectValid: true,
			expectValue: 20,
		},
		{
			name: "float with decimals (truncated)",
			input: &ast.Literal{
				Value: 3.14,
			},
			expectValid: true,
			expectValue: 3,
		},
		{
			name: "string literal",
			input: &ast.Literal{
				Value: "not a number",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ParseInt(tt.input)
			if result.IsValid != tt.expectValid {
				t.Errorf("expected IsValid=%v, got %v", tt.expectValid, result.IsValid)
			}
			if tt.expectValid {
				if result.MustBeInt() != tt.expectValue {
					t.Errorf("expected value %d, got %d", tt.expectValue, result.MustBeInt())
				}
			}
		})
	}
}

func TestArgumentParser_ParseFloat(t *testing.T) {
	parser := NewArgumentParser()

	tests := []struct {
		name        string
		input       ast.Expression
		expectValid bool
		expectValue float64
	}{
		{
			name: "float64 literal",
			input: &ast.Literal{
				Value: 3.14,
			},
			expectValid: true,
			expectValue: 3.14,
		},
		{
			name: "integer literal (converted to float)",
			input: &ast.Literal{
				Value: 42,
			},
			expectValid: true,
			expectValue: 42.0,
		},
		{
			name: "bool literal",
			input: &ast.Literal{
				Value: true,
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ParseFloat(tt.input)
			if result.IsValid != tt.expectValid {
				t.Errorf("expected IsValid=%v, got %v", tt.expectValid, result.IsValid)
			}
			if tt.expectValid {
				if result.MustBeFloat() != tt.expectValue {
					t.Errorf("expected value %f, got %f", tt.expectValue, result.MustBeFloat())
				}
			}
		})
	}
}

func TestArgumentParser_ParseBool(t *testing.T) {
	parser := NewArgumentParser()

	tests := []struct {
		name        string
		input       ast.Expression
		expectValid bool
		expectValue bool
	}{
		{
			name: "true literal",
			input: &ast.Literal{
				Value: true,
			},
			expectValid: true,
			expectValue: true,
		},
		{
			name: "false literal",
			input: &ast.Literal{
				Value: false,
			},
			expectValid: true,
			expectValue: false,
		},
		{
			name: "integer literal",
			input: &ast.Literal{
				Value: 1,
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ParseBool(tt.input)
			if result.IsValid != tt.expectValid {
				t.Errorf("expected IsValid=%v, got %v", tt.expectValid, result.IsValid)
			}
			if tt.expectValid {
				if result.MustBeBool() != tt.expectValue {
					t.Errorf("expected value %v, got %v", tt.expectValue, result.MustBeBool())
				}
			}
		})
	}
}

func TestArgumentParser_ParseIdentifier(t *testing.T) {
	parser := NewArgumentParser()

	tests := []struct {
		name             string
		input            ast.Expression
		expectValid      bool
		expectIdentifier string
		expectLiteral    bool
	}{
		{
			name: "simple identifier",
			input: &ast.Identifier{
				Name: "my_variable",
			},
			expectValid:      true,
			expectIdentifier: "my_variable",
			expectLiteral:    false,
		},
		{
			name: "identifier with underscores",
			input: &ast.Identifier{
				Name: "entry_time_input",
			},
			expectValid:      true,
			expectIdentifier: "entry_time_input",
			expectLiteral:    false,
		},
		{
			name: "literal (not identifier)",
			input: &ast.Literal{
				Value: "string",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ParseIdentifier(tt.input)
			if result.IsValid != tt.expectValid {
				t.Errorf("expected IsValid=%v, got %v", tt.expectValid, result.IsValid)
			}
			if tt.expectValid {
				if result.Identifier != tt.expectIdentifier {
					t.Errorf("expected identifier %q, got %q", tt.expectIdentifier, result.Identifier)
				}
				if result.IsLiteral != tt.expectLiteral {
					t.Errorf("expected IsLiteral=%v, got %v", tt.expectLiteral, result.IsLiteral)
				}
			}
		})
	}
}

func TestArgumentParser_ParseWrappedIdentifier(t *testing.T) {
	parser := NewArgumentParser()

	tests := []struct {
		name             string
		input            ast.Expression
		expectValid      bool
		expectIdentifier string
	}{
		{
			name: "wrapped identifier with [0]",
			input: &ast.MemberExpression{
				Computed: true,
				Object: &ast.Identifier{
					Name: "my_session",
				},
				Property: &ast.Literal{
					Value: 0,
				},
			},
			expectValid:      true,
			expectIdentifier: "my_session",
		},
		{
			name: "non-computed member expression",
			input: &ast.MemberExpression{
				Computed: false,
				Object: &ast.Identifier{
					Name: "obj",
				},
				Property: &ast.Identifier{
					Name: "prop",
				},
			},
			expectValid: false,
		},
		{
			name: "wrapped with non-zero index",
			input: &ast.MemberExpression{
				Computed: true,
				Object: &ast.Identifier{
					Name: "my_var",
				},
				Property: &ast.Literal{
					Value: 1,
				},
			},
			expectValid: false,
		},
		{
			name: "simple identifier (not wrapped)",
			input: &ast.Identifier{
				Name: "simple",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ParseWrappedIdentifier(tt.input)
			if result.IsValid != tt.expectValid {
				t.Errorf("expected IsValid=%v, got %v", tt.expectValid, result.IsValid)
			}
			if tt.expectValid {
				if result.Identifier != tt.expectIdentifier {
					t.Errorf("expected identifier %q, got %q", tt.expectIdentifier, result.Identifier)
				}
				if result.IsLiteral {
					t.Error("wrapped identifier should not be marked as literal")
				}
			}
		})
	}
}

func TestArgumentParser_ParseStringOrIdentifier(t *testing.T) {
	parser := NewArgumentParser()

	tests := []struct {
		name          string
		input         ast.Expression
		expectValid   bool
		expectLiteral bool
		expectValue   string
	}{
		{
			name: "string literal",
			input: &ast.Literal{
				Value: `"0950-1645"`,
			},
			expectValid:   true,
			expectLiteral: true,
			expectValue:   "0950-1645",
		},
		{
			name: "identifier",
			input: &ast.Identifier{
				Name: "entry_time",
			},
			expectValid:   true,
			expectLiteral: false,
			expectValue:   "entry_time",
		},
		{
			name: "integer (neither string nor identifier)",
			input: &ast.Literal{
				Value: 123,
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ParseStringOrIdentifier(tt.input)
			if result.IsValid != tt.expectValid {
				t.Errorf("expected IsValid=%v, got %v", tt.expectValid, result.IsValid)
			}
			if tt.expectValid {
				if result.IsLiteral != tt.expectLiteral {
					t.Errorf("expected IsLiteral=%v, got %v", tt.expectLiteral, result.IsLiteral)
				}
				actualValue := result.MustBeString()
				if actualValue != tt.expectValue {
					t.Errorf("expected value %q, got %q", tt.expectValue, actualValue)
				}
			}
		})
	}
}

func TestArgumentParser_ParseIdentifierOrWrapped(t *testing.T) {
	parser := NewArgumentParser()

	tests := []struct {
		name             string
		input            ast.Expression
		expectValid      bool
		expectIdentifier string
	}{
		{
			name: "simple identifier",
			input: &ast.Identifier{
				Name: "my_var",
			},
			expectValid:      true,
			expectIdentifier: "my_var",
		},
		{
			name: "wrapped identifier",
			input: &ast.MemberExpression{
				Computed: true,
				Object: &ast.Identifier{
					Name: "wrapped_var",
				},
				Property: &ast.Literal{
					Value: 0,
				},
			},
			expectValid:      true,
			expectIdentifier: "wrapped_var",
		},
		{
			name: "string literal (not identifier)",
			input: &ast.Literal{
				Value: "string",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ParseIdentifierOrWrapped(tt.input)
			if result.IsValid != tt.expectValid {
				t.Errorf("expected IsValid=%v, got %v", tt.expectValid, result.IsValid)
			}
			if tt.expectValid {
				if result.Identifier != tt.expectIdentifier {
					t.Errorf("expected identifier %q, got %q", tt.expectIdentifier, result.Identifier)
				}
				if result.IsLiteral {
					t.Error("identifier should not be marked as literal")
				}
			}
		})
	}
}

func TestArgumentParser_ParseSession(t *testing.T) {
	parser := NewArgumentParser()

	tests := []struct {
		name          string
		input         ast.Expression
		expectValid   bool
		expectLiteral bool
		expectValue   string
	}{
		{
			name: "string literal session",
			input: &ast.Literal{
				Value: `"0950-1645"`,
			},
			expectValid:   true,
			expectLiteral: true,
			expectValue:   "0950-1645",
		},
		{
			name: "identifier session",
			input: &ast.Identifier{
				Name: "entry_time_input",
			},
			expectValid:   true,
			expectLiteral: false,
			expectValue:   "entry_time_input",
		},
		{
			name: "wrapped identifier session",
			input: &ast.MemberExpression{
				Computed: true,
				Object: &ast.Identifier{
					Name: "my_session",
				},
				Property: &ast.Literal{
					Value: 0,
				},
			},
			expectValid:   true,
			expectLiteral: false,
			expectValue:   "my_session",
		},
		{
			name: "invalid type",
			input: &ast.Literal{
				Value: 123,
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ParseSession(tt.input)
			if result.IsValid != tt.expectValid {
				t.Errorf("expected IsValid=%v, got %v", tt.expectValid, result.IsValid)
			}
			if tt.expectValid {
				if result.IsLiteral != tt.expectLiteral {
					t.Errorf("expected IsLiteral=%v, got %v", tt.expectLiteral, result.IsLiteral)
				}
				actualValue := result.MustBeString()
				if actualValue != tt.expectValue {
					t.Errorf("expected value %q, got %q", tt.expectValue, actualValue)
				}
			}
		})
	}
}

func TestParsedArgument_MustMethods(t *testing.T) {
	t.Run("MustBeString", func(t *testing.T) {
		arg := ParsedArgument{IsValid: true, IsLiteral: true, Value: "hello"}
		if arg.MustBeString() != "hello" {
			t.Errorf("expected 'hello', got %q", arg.MustBeString())
		}

		arg = ParsedArgument{IsValid: true, IsLiteral: false, Identifier: "my_var"}
		if arg.MustBeString() != "my_var" {
			t.Errorf("expected 'my_var', got %q", arg.MustBeString())
		}

		arg = ParsedArgument{IsValid: false}
		if arg.MustBeString() != "" {
			t.Errorf("expected empty string for invalid arg, got %q", arg.MustBeString())
		}
	})

	t.Run("MustBeInt", func(t *testing.T) {
		arg := ParsedArgument{IsValid: true, IsLiteral: true, Value: 42}
		if arg.MustBeInt() != 42 {
			t.Errorf("expected 42, got %d", arg.MustBeInt())
		}

		arg = ParsedArgument{IsValid: false}
		if arg.MustBeInt() != 0 {
			t.Errorf("expected 0 for invalid arg, got %d", arg.MustBeInt())
		}
	})

	t.Run("MustBeFloat", func(t *testing.T) {
		arg := ParsedArgument{IsValid: true, IsLiteral: true, Value: 3.14}
		if arg.MustBeFloat() != 3.14 {
			t.Errorf("expected 3.14, got %f", arg.MustBeFloat())
		}

		arg = ParsedArgument{IsValid: false}
		if arg.MustBeFloat() != 0.0 {
			t.Errorf("expected 0.0 for invalid arg, got %f", arg.MustBeFloat())
		}
	})

	t.Run("MustBeBool", func(t *testing.T) {
		arg := ParsedArgument{IsValid: true, IsLiteral: true, Value: true}
		if !arg.MustBeBool() {
			t.Error("expected true")
		}

		arg = ParsedArgument{IsValid: false}
		if arg.MustBeBool() {
			t.Error("expected false for invalid arg")
		}
	})
}
