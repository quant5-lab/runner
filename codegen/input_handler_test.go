package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestInputHandler_GenerateInputFloat(t *testing.T) {
	tests := []struct {
		name     string
		call     *ast.CallExpression
		varName  string
		expected string
	}{
		{
			name: "positional defval",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: 1.5},
				},
			},
			varName:  "mult",
			expected: "const mult = 1.50\n",
		},
		{
			name: "named defval",
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
								Value: &ast.Literal{Value: "Multiplier"},
							},
						},
					},
				},
			},
			varName:  "factor",
			expected: "const factor = 2.50\n",
		},
		{
			name: "positional with metadata properties",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: 3.14},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "title"},
								Value: &ast.Literal{Value: "Pi Value"},
							},
							{
								Key:   &ast.Identifier{Name: "minval"},
								Value: &ast.Literal{Value: 0.0},
							},
						},
					},
				},
			},
			varName:  "pi",
			expected: "const pi = 3.14\n",
		},
		{
			name: "negative value",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: -2.5},
				},
			},
			varName:  "offset",
			expected: "const offset = -2.50\n",
		},
		{
			name: "zero value",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: 0.0},
				},
			},
			varName:  "baseline",
			expected: "const baseline = 0.00\n",
		},
		{
			name: "no arguments defaults to 0",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{},
			},
			varName:  "value",
			expected: "const value = 0.00\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ih := NewInputHandler()
			result, err := ih.GenerateInputFloat(tt.call, tt.varName)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInputHandler_GenerateInputInt(t *testing.T) {
	tests := []struct {
		name     string
		call     *ast.CallExpression
		varName  string
		expected string
	}{
		{
			name: "positional defval",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: float64(20)},
				},
			},
			varName:  "length",
			expected: "const length = 20\n",
		},
		{
			name: "named defval",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "defval"},
								Value: &ast.Literal{Value: float64(14)},
							},
						},
					},
				},
			},
			varName:  "period",
			expected: "const period = 14\n",
		},
		{
			name: "positional with metadata properties",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: float64(14)},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "title"},
								Value: &ast.Literal{Value: "DI Length"},
							},
							{
								Key:   &ast.Identifier{Name: "minval"},
								Value: &ast.Literal{Value: float64(1)},
							},
							{
								Key:   &ast.Identifier{Name: "group"},
								Value: &ast.Identifier{Name: "dmiGroup"},
							},
						},
					},
				},
			},
			varName:  "diLen",
			expected: "const diLen = 14\n",
		},
		{
			name: "negative value",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: float64(-5)},
				},
			},
			varName:  "offset",
			expected: "const offset = -5\n",
		},
		{
			name: "zero value",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: float64(0)},
				},
			},
			varName:  "baseline",
			expected: "const baseline = 0\n",
		},
		{
			name: "no arguments defaults to 0",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{},
			},
			varName:  "value",
			expected: "const value = 0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ih := NewInputHandler()
			result, err := ih.GenerateInputInt(tt.call, tt.varName)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInputHandler_GenerateInputBool(t *testing.T) {
	tests := []struct {
		name     string
		call     *ast.CallExpression
		varName  string
		expected string
	}{
		{
			name: "positional true",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: true},
				},
			},
			varName:  "enabled",
			expected: "const enabled = true\n",
		},
		{
			name: "positional false",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: false},
				},
			},
			varName:  "disabled",
			expected: "const disabled = false\n",
		},
		{
			name: "named defval true",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "defval"},
								Value: &ast.Literal{Value: true},
							},
						},
					},
				},
			},
			varName:  "active",
			expected: "const active = true\n",
		},
		{
			name: "named defval false",
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
			varName:  "showTrades",
			expected: "const showTrades = false\n",
		},
		{
			name: "positional with metadata properties",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: true},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "title"},
								Value: &ast.Literal{Value: "Long entries"},
							},
							{
								Key:   &ast.Identifier{Name: "group"},
								Value: &ast.Identifier{Name: "stratGroup"},
							},
						},
					},
				},
			},
			varName:  "showLong",
			expected: "const showLong = true\n",
		},
		{
			name: "no arguments defaults to false",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{},
			},
			varName:  "flag",
			expected: "const flag = false\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ih := NewInputHandler()
			result, err := ih.GenerateInputBool(tt.call, tt.varName)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInputHandler_GenerateInputString(t *testing.T) {
	tests := []struct {
		name     string
		call     *ast.CallExpression
		varName  string
		expected string
	}{
		{
			name: "positional string",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
				},
			},
			varName:  "symbol",
			expected: "const symbol = \"BTCUSDT\"\n",
		},
		{
			name: "named defval",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "defval"},
								Value: &ast.Literal{Value: "1D"},
							},
						},
					},
				},
			},
			varName:  "timeframe",
			expected: "const timeframe = \"1D\"\n",
		},
		{
			name: "positional with metadata properties",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: "EMA"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "title"},
								Value: &ast.Literal{Value: "MA Type"},
							},
							{
								Key:   &ast.Identifier{Name: "group"},
								Value: &ast.Identifier{Name: "maGroup"},
							},
						},
					},
				},
			},
			varName:  "maType",
			expected: "const maType = \"EMA\"\n",
		},
		{
			name: "empty string",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: ""},
				},
			},
			varName:  "label",
			expected: "const label = \"\"\n",
		},
		{
			name: "no arguments defaults to empty string",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{},
			},
			varName:  "value",
			expected: "const value = \"\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ih := NewInputHandler()
			result, err := ih.GenerateInputString(tt.call, tt.varName)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInputHandler_DetectInputFunction(t *testing.T) {
	tests := []struct {
		name     string
		call     *ast.CallExpression
		expected bool
	}{
		{
			name: "input.float detected",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "float"},
				},
			},
			expected: true,
		},
		{
			name: "input.int detected",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
			},
			expected: true,
		},
		{
			name: "ta.sma not detected",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ih := NewInputHandler()
			result := ih.DetectInputFunction(tt.call)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestInputHandler_Integration(t *testing.T) {
	// Test that multiple input constants are stored correctly
	ih := NewInputHandler()

	call1 := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Literal{Value: 1.5},
		},
	}
	call2 := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Literal{Value: float64(20)},
		},
	}

	ih.GenerateInputFloat(call1, "mult")
	ih.GenerateInputInt(call2, "length")

	if len(ih.inputConstants) != 2 {
		t.Errorf("expected 2 constants, got %d", len(ih.inputConstants))
	}

	if !strings.Contains(ih.inputConstants["mult"], "1.50") {
		t.Errorf("mult constant not stored correctly: %s", ih.inputConstants["mult"])
	}
	if !strings.Contains(ih.inputConstants["length"], "20") {
		t.Errorf("length constant not stored correctly: %s", ih.inputConstants["length"])
	}
}
