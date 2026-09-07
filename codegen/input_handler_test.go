package codegen

import (
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
			expected: "const mult = 1.5\n",
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
			expected: "const factor = 2.5\n",
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
			expected: "const offset = -2.5\n",
		},
		{
			name: "zero value",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: 0.0},
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
		{
			name: "named defval with multi-decimal value",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "defval"},
								Value: &ast.Literal{Value: 0.618},
							},
							{
								Key:   &ast.Identifier{Name: "title"},
								Value: &ast.Literal{Value: "Fib Rate"},
							},
						},
					},
				},
			},
			varName:  "fibRate",
			expected: "const fibRate = 0.618\n",
		},
		{
			name: "non-literal argument falls back to zero",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "someVar"},
				},
			},
			varName:  "rate",
			expected: "const rate = 0\n",
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

func TestInputHandler_GenerateInputFloat_RoundtripFidelity(t *testing.T) {
	cases := []struct {
		name   string
		defval float64
		want   string
	}{
		{"zero", 0.0, "const v = 0\n"},
		{"integer-valued", 10000.0, "const v = 10000\n"},
		{"one-decimal", 1.5, "const v = 1.5\n"},
		{"two-decimal", 3.14, "const v = 3.14\n"},
		{"three-decimal", 0.236, "const v = 0.236\n"},
		{"four-decimal", 1.618, "const v = 1.618\n"},
		{"sub-milli", 0.0001, "const v = 0.0001\n"},
		{"negative-one-decimal", -2.5, "const v = -2.5\n"},
		{"negative-three-decimal", -0.236, "const v = -0.236\n"},
		{"price-one-decimal", 99.5, "const v = 99.5\n"},
		{"integer-valued-large", 100.0, "const v = 100\n"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ih := NewInputHandler()
			call := &ast.CallExpression{Arguments: []ast.Expression{&ast.Literal{Value: c.defval}}}

			code, err := ih.GenerateInputFloat(call, "v")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if code != c.want {
				t.Errorf("emitted %q, want %q", code, c.want)
			}

			if got := ih.GetInputConstantsMap()["v"]; got != c.defval {
				t.Errorf("GetInputConstantsMap round-trip: got %v, want %v", got, c.defval)
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
		{name: "ta.sma", call: memberCall("ta", "sma"), expected: false},
		{name: "strategy.entry", call: memberCall("strategy", "entry"), expected: false},
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

	if ih.inputConstants["mult"] != "const mult = 1.5\n" {
		t.Errorf("mult: got %q", ih.inputConstants["mult"])
	}
	if ih.inputConstants["length"] != "const length = 20\n" {
		t.Errorf("length: got %q", ih.inputConstants["length"])
	}
}

func TestInputHandler_GetInputConstantsMap(t *testing.T) {
	numericCall := func(v float64) *ast.CallExpression {
		return &ast.CallExpression{Arguments: []ast.Expression{&ast.Literal{Value: v}}}
	}
	boolCall := func(v bool) *ast.CallExpression {
		return &ast.CallExpression{Arguments: []ast.Expression{&ast.Literal{Value: v}}}
	}

	ih := NewInputHandler()
	ih.GenerateInputFloat(numericCall(0.236), "rate")
	ih.GenerateInputInt(numericCall(14), "period")
	ih.GenerateInputBool(boolCall(true), "enabled")
	ih.GenerateInputBool(boolCall(false), "disabled")

	m := ih.GetInputConstantsMap()

	if got := m["rate"]; got != 0.236 {
		t.Errorf("float: got %v, want 0.236", got)
	}
	if got := m["period"]; got != 14.0 {
		t.Errorf("int: got %v, want 14", got)
	}
	if got := m["enabled"]; got != 1.0 {
		t.Errorf("bool true: got %v, want 1", got)
	}
	if got := m["disabled"]; got != 0.0 {
		t.Errorf("bool false: got %v, want 0", got)
	}
}

func TestInputHandler_IsInputConstant(t *testing.T) {
	ih := NewInputHandler()
	call := &ast.CallExpression{Arguments: []ast.Expression{&ast.Literal{Value: 1.5}}}

	if ih.IsInputConstant("rate") {
		t.Error("before registration: expected false")
	}
	ih.GenerateInputFloat(call, "rate")
	if !ih.IsInputConstant("rate") {
		t.Error("after registration: expected true")
	}
	if ih.IsInputConstant("other") {
		t.Error("unregistered name: expected false")
	}
}
