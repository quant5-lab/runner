package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestInputConstantExtractor_ExtractInputConstant_Float(t *testing.T) {
	extractor := NewInputConstantExtractor()

	tests := []struct {
		name     string
		funcName string
		call     *ast.CallExpression
		expected string
	}{
		{
			name:     "positional float argument",
			funcName: "input.float",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: 1.5},
				},
			},
			expected: "1.5",
		},
		{
			name:     "positional integer converts to float",
			funcName: "input.float",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: 10},
				},
			},
			expected: "10",
		},
		{
			name:     "named parameter defval in object",
			funcName: "input.float",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "defval"},
								Value: &ast.Literal{Value: 2.5},
							},
							{
								Key:   &ast.Identifier{Name: "title"},
								Value: &ast.Literal{Value: "Test"},
							},
						},
					},
				},
			},
			expected: "2.5",
		},
		{
			name:     "object without defval returns default zero",
			funcName: "input.float",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "title"},
								Value: &ast.Literal{Value: "Test"},
							},
						},
					},
				},
			},
			expected: "0.0",
		},
		{
			name:     "no arguments returns default zero",
			funcName: "input.float",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{},
			},
			expected: "0.0",
		},
		{
			name:     "nil arguments returns default zero",
			funcName: "input.float",
			call: &ast.CallExpression{
				Arguments: nil,
			},
			expected: "0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.ExtractInputConstant(tt.call, tt.funcName)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInputConstantExtractor_ExtractInputConstant_Price(t *testing.T) {
	extractor := NewInputConstantExtractor()

	tests := []struct {
		name     string
		funcName string
		call     *ast.CallExpression
		expected string
	}{
		{
			name:     "price uses float extraction",
			funcName: "input.price",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: 100.5},
				},
			},
			expected: "100.5",
		},
		{
			name:     "price with object defval",
			funcName: "input.price",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "defval"},
								Value: &ast.Literal{Value: 50.25},
							},
						},
					},
				},
			},
			expected: "50.25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.ExtractInputConstant(tt.call, tt.funcName)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInputConstantExtractor_ExtractInputConstant_Int(t *testing.T) {
	extractor := NewInputConstantExtractor()

	tests := []struct {
		name     string
		funcName string
		call     *ast.CallExpression
		expected string
	}{
		{
			name:     "positional int argument",
			funcName: "input.int",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: 42},
				},
			},
			expected: "42",
		},
		{
			name:     "float truncates to int",
			funcName: "input.int",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: 3.99},
				},
			},
			expected: "3",
		},
		{
			name:     "named parameter defval in object",
			funcName: "input.int",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "defval"},
								Value: &ast.Literal{Value: 20.0},
							},
						},
					},
				},
			},
			expected: "20",
		},
		{
			name:     "no arguments returns default zero",
			funcName: "input.int",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{},
			},
			expected: "0",
		},
		{
			name:     "negative integer",
			funcName: "input.int",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: -15},
				},
			},
			expected: "-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.ExtractInputConstant(tt.call, tt.funcName)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInputConstantExtractor_ExtractInputConstant_Bool(t *testing.T) {
	extractor := NewInputConstantExtractor()

	tests := []struct {
		name     string
		funcName string
		call     *ast.CallExpression
		expected string
	}{
		{
			name:     "positional true argument",
			funcName: "input.bool",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: true},
				},
			},
			expected: "true",
		},
		{
			name:     "positional false argument",
			funcName: "input.bool",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: false},
				},
			},
			expected: "false",
		},
		{
			name:     "named parameter defval true in object",
			funcName: "input.bool",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "defval"},
								Value: &ast.Literal{Value: true},
							},
							{
								Key:   &ast.Identifier{Name: "title"},
								Value: &ast.Literal{Value: "Enable"},
							},
						},
					},
				},
			},
			expected: "true",
		},
		{
			name:     "named parameter defval false in object",
			funcName: "input.bool",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "defval"},
								Value: &ast.Literal{Value: false},
							},
						},
					},
				},
			},
			expected: "false",
		},
		{
			name:     "no arguments returns default false",
			funcName: "input.bool",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{},
			},
			expected: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.ExtractInputConstant(tt.call, tt.funcName)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInputConstantExtractor_ExtractInputConstant_String(t *testing.T) {
	extractor := NewInputConstantExtractor()

	tests := []struct {
		name     string
		funcName string
		call     *ast.CallExpression
		expected string
	}{
		{
			name:     "positional string argument",
			funcName: "input.string",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: "EMA"},
				},
			},
			expected: `"EMA"`,
		},
		{
			name:     "empty string",
			funcName: "input.string",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: ""},
				},
			},
			expected: `""`,
		},
		{
			name:     "string with spaces",
			funcName: "input.string",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Moving Average"},
				},
			},
			expected: `"Moving Average"`,
		},
		{
			name:     "named parameter defval in object",
			funcName: "input.string",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "defval"},
								Value: &ast.Literal{Value: "SMA"},
							},
						},
					},
				},
			},
			expected: `"SMA"`,
		},
		{
			name:     "no arguments returns empty string",
			funcName: "input.string",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{},
			},
			expected: `""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.ExtractInputConstant(tt.call, tt.funcName)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInputConstantExtractor_ExtractInputConstant_FallThrough(t *testing.T) {
	extractor := NewInputConstantExtractor()

	tests := []struct {
		name     string
		funcName string
		call     *ast.CallExpression
		expected string
	}{
		{
			name:     "input.source returns empty (fall through)",
			funcName: "input.source",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			expected: "",
		},
		{
			name:     "input.color returns empty (not implemented)",
			funcName: "input.color",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: "red"},
				},
			},
			expected: "",
		},
		{
			name:     "input.timeframe returns empty (not implemented)",
			funcName: "input.timeframe",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: "D"},
				},
			},
			expected: "",
		},
		{
			name:     "ta.sma returns empty (not input function)",
			funcName: "ta.sma",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 20},
				},
			},
			expected: "",
		},
		{
			name:     "unknown function returns empty",
			funcName: "custom.func",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.ExtractInputConstant(tt.call, tt.funcName)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInputConstantExtractor_ExtractInputConstant_NilHandling(t *testing.T) {
	extractor := NewInputConstantExtractor()

	tests := []struct {
		name     string
		funcName string
		call     *ast.CallExpression
		expected string
	}{
		{
			name:     "nil call returns empty",
			funcName: "input.float",
			call:     nil,
			expected: "",
		},
		{
			name:     "nil call with any funcName returns empty",
			funcName: "input.int",
			call:     nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.ExtractInputConstant(tt.call, tt.funcName)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInputConstantExtractor_ExtractInputConstant_EdgeCases(t *testing.T) {
	extractor := NewInputConstantExtractor()

	tests := []struct {
		name     string
		funcName string
		call     *ast.CallExpression
		expected string
	}{
		{
			name:     "object with multiple properties uses defval",
			funcName: "input.float",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "title"},
								Value: &ast.Literal{Value: "Length"},
							},
							{
								Key:   &ast.Identifier{Name: "defval"},
								Value: &ast.Literal{Value: 14.0},
							},
							{
								Key:   &ast.Identifier{Name: "minval"},
								Value: &ast.Literal{Value: 1.0},
							},
						},
					},
				},
			},
			expected: "14",
		},
		{
			name:     "object with non-identifier key skips property",
			funcName: "input.int",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Literal{Value: "defval"},
								Value: &ast.Literal{Value: 10},
							},
						},
					},
				},
			},
			expected: "0",
		},
		{
			name:     "object with invalid value type returns default",
			funcName: "input.float",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "defval"},
								Value: &ast.Identifier{Name: "close"},
							},
						},
					},
				},
			},
			expected: "0.0",
		},
		{
			name:     "zero value float",
			funcName: "input.float",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: 0.0},
				},
			},
			expected: "0",
		},
		{
			name:     "zero value int",
			funcName: "input.int",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: 0},
				},
			},
			expected: "0",
		},
		{
			name:     "large float value",
			funcName: "input.float",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: 999999.123456},
				},
			},
			expected: "999999.123456",
		},
		{
			name:     "large int value",
			funcName: "input.int",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: 1000000},
				},
			},
			expected: "1000000",
		},
		{
			name:     "negative float value",
			funcName: "input.float",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: -5.5},
				},
			},
			expected: "-5.5",
		},
		{
			name:     "small decimal precision",
			funcName: "input.float",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: 0.0001},
				},
			},
			expected: "0.0001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.ExtractInputConstant(tt.call, tt.funcName)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInputConstantExtractor_ObjectPropertyExtraction(t *testing.T) {
	extractor := NewInputConstantExtractor()

	tests := []struct {
		name     string
		funcName string
		call     *ast.CallExpression
		expected string
	}{
		{
			name:     "first matching defval property used",
			funcName: "input.float",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "defval"},
								Value: &ast.Literal{Value: 1.0},
							},
							{
								Key:   &ast.Identifier{Name: "defval"},
								Value: &ast.Literal{Value: 2.0},
							},
						},
					},
				},
			},
			expected: "1",
		},
		{
			name:     "property order doesn't matter for defval extraction",
			funcName: "input.bool",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "inline"},
								Value: &ast.Literal{Value: true},
							},
							{
								Key:   &ast.Identifier{Name: "group"},
								Value: &ast.Literal{Value: "Settings"},
							},
							{
								Key:   &ast.Identifier{Name: "defval"},
								Value: &ast.Literal{Value: false},
							},
							{
								Key:   &ast.Identifier{Name: "title"},
								Value: &ast.Literal{Value: "Show Labels"},
							},
						},
					},
				},
			},
			expected: "false",
		},
		{
			name:     "empty object properties returns default",
			funcName: "input.string",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{},
					},
				},
			},
			expected: `""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.ExtractInputConstant(tt.call, tt.funcName)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
