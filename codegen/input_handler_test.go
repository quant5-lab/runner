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
		/* All recognized input.* types */
		{name: "input.float", call: memberCall("input", "float"), expected: true},
		{name: "input.int", call: memberCall("input", "int"), expected: true},
		{name: "input.bool", call: memberCall("input", "bool"), expected: true},
		{name: "input.string", call: memberCall("input", "string"), expected: true},
		{name: "input.session", call: memberCall("input", "session"), expected: true},
		{name: "input.source", call: memberCall("input", "source"), expected: true},
		{name: "input.symbol", call: memberCall("input", "symbol"), expected: true},
		{name: "input.timeframe", call: memberCall("input", "timeframe"), expected: true},
		{name: "input.text_area", call: memberCall("input", "text_area"), expected: true},
		{name: "input.price", call: memberCall("input", "price"), expected: true},
		{name: "input.time", call: memberCall("input", "time"), expected: true},
		{name: "input.color", call: memberCall("input", "color"), expected: true},
		/* Negative: non-input member expressions */
		{name: "ta.sma", call: memberCall("ta", "sma"), expected: false},
		{name: "strategy.entry", call: memberCall("strategy", "entry"), expected: false},
		/* Negative: bare identifier (not member expression) */
		{
			name: "bare input",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "input"},
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

func memberCall(obj, prop string) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: obj},
			Property: &ast.Identifier{Name: prop},
		},
	}
}

func TestInputHandler_GenerateInputSession(t *testing.T) {
	tests := []struct {
		name     string
		call     *ast.CallExpression
		varName  string
		expected string
	}{
		{
			name: "positional session",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: "0930-1600"},
				},
			},
			varName:  "sess",
			expected: "const sess = \"0930-1600\"\n",
		},
		{
			name: "named defval",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "defval"},
								Value: &ast.Literal{Value: "0950-1345"},
							},
						},
					},
				},
			},
			varName:  "tradingSession",
			expected: "const tradingSession = \"0950-1345\"\n",
		},
		{
			name: "no arguments defaults to full day",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{},
			},
			varName:  "window",
			expected: "const window = \"0000-2359\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ih := NewInputHandler()
			result, err := ih.GenerateInputSession(tt.call, tt.varName)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInputHandler_GenerateInputColor(t *testing.T) {
	tests := []struct {
		name     string
		call     *ast.CallExpression
		varName  string
		expected string
	}{
		{
			name: "positional named color",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "color"},
						Property: &ast.Identifier{Name: "red"},
					},
				},
			},
			varName:  "lineColor",
			expected: "const lineColor = \"#FF5252\"\n",
		},
		{
			name: "positional hex literal",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: "#00FF00"},
				},
			},
			varName:  "greenHex",
			expected: "const greenHex = \"#00FF00\"\n",
		},
		{
			name: "positional color.new",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "color"},
							Property: &ast.Identifier{Name: "new"},
						},
						Arguments: []ast.Expression{
							&ast.MemberExpression{
								Object:   &ast.Identifier{Name: "color"},
								Property: &ast.Identifier{Name: "blue"},
							},
							&ast.Literal{Value: float64(50)},
						},
					},
				},
			},
			varName:  "fadeBlue",
			expected: "const fadeBlue = \"#2962FF\"\n",
		},
		{
			name: "named defval",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key: &ast.Identifier{Name: "defval"},
								Value: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "color"},
									Property: &ast.Identifier{Name: "blue"},
								},
							},
							{
								Key:   &ast.Identifier{Name: "title"},
								Value: &ast.Literal{Value: "Color"},
							},
						},
					},
				},
			},
			varName:  "bgColor",
			expected: "const bgColor = \"#2962FF\"\n",
		},
		{
			name: "named defval with color.rgb",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key: &ast.Identifier{Name: "defval"},
								Value: &ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "color"},
										Property: &ast.Identifier{Name: "rgb"},
									},
									Arguments: []ast.Expression{
										&ast.Literal{Value: float64(255)},
										&ast.Literal{Value: float64(255)},
										&ast.Literal{Value: float64(255)},
									},
								},
							},
						},
					},
				},
			},
			varName:  "whiteColor",
			expected: "const whiteColor = \"#FFFFFF\"\n",
		},
		{
			name: "named defval with hex literal",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "defval"},
								Value: &ast.Literal{Value: "#AABB00"},
							},
						},
					},
				},
			},
			varName:  "custom",
			expected: "const custom = \"#AABB00\"\n",
		},
		{
			name: "unresolvable defval defaults to empty",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "defval"},
								Value: &ast.Identifier{Name: "myColorVar"},
							},
						},
					},
				},
			},
			varName:  "dynColor",
			expected: "const dynColor = \"\"\n",
		},
		{
			name: "object without defval defaults to empty",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "title"},
								Value: &ast.Literal{Value: "Bar Color"},
							},
						},
					},
				},
			},
			varName:  "barColor",
			expected: "const barColor = \"\"\n",
		},
		{
			name: "no arguments defaults to empty",
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
			result, err := ih.GenerateInputColor(tt.call, tt.varName)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

/* Multiple input types stored independently in constant map */
func TestInputHandler_Integration(t *testing.T) {
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
