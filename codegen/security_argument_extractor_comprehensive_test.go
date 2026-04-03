package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestExtractSymbol_AllExpressionTypes(t *testing.T) {
	tests := []struct {
		name        string
		expr        ast.Expression
		generator   *generator
		wantCode    string
		wantRuntime bool
		wantError   bool
	}{
		{
			name:      "nil expression",
			expr:      nil,
			wantError: true,
		},
		{
			name:        "identifier - tickerid builtin",
			expr:        &ast.Identifier{Name: "tickerid"},
			wantCode:    "ctx.Symbol",
			wantRuntime: true,
		},
		{
			name:        "identifier - unknown becomes literal",
			expr:        &ast.Identifier{Name: "unknownVar"},
			wantCode:    `"unknownVar"`,
			wantRuntime: false,
		},
		{
			name: "identifier - string variable without generator",
			expr: &ast.Identifier{Name: "mySymbol"},
			generator: &generator{
				variables: map[string]string{"mySymbol": "string"},
			},
			wantCode:    "mySymbol",
			wantRuntime: true,
		},
		{
			name: "identifier - string constant resolved",
			expr: &ast.Identifier{Name: "MY_CONST"},
			generator: &generator{
				variables: map[string]string{"MY_CONST": "string"},
				constants: map[string]interface{}{"MY_CONST": "BTCUSDT"},
			},
			wantCode:    `"BTCUSDT"`,
			wantRuntime: false,
		},
		{
			name: "identifier - non-string variable becomes literal",
			expr: &ast.Identifier{Name: "intVar"},
			generator: &generator{
				variables: map[string]string{"intVar": "int"},
			},
			wantCode:    `"intVar"`,
			wantRuntime: false,
		},
		{
			name: "member - syminfo.tickerid",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "syminfo"},
				Property: &ast.Identifier{Name: "tickerid"},
			},
			wantCode:    "ctx.Symbol",
			wantRuntime: true,
		},
		{
			name: "member - syminfo.ticker",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "syminfo"},
				Property: &ast.Identifier{Name: "ticker"},
			},
			wantCode:    "ctx.Symbol",
			wantRuntime: true,
		},
		{
			name: "member - unsupported object",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "custom"},
				Property: &ast.Identifier{Name: "symbol"},
			},
			wantError: true,
		},
		{
			name: "member - nil object",
			expr: &ast.MemberExpression{
				Object:   nil,
				Property: &ast.Identifier{Name: "tickerid"},
			},
			wantError: true,
		},
		{
			name: "member - nil property",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "syminfo"},
				Property: nil,
			},
			wantError: true,
		},
		{
			name:        "literal - string",
			expr:        &ast.Literal{Value: "BTCUSDT"},
			wantCode:    `"BTCUSDT"`,
			wantRuntime: false,
		},
		{
			name:        "literal - string with special chars",
			expr:        &ast.Literal{Value: "BINANCE:BTCUSDT"},
			wantCode:    `"BINANCE:BTCUSDT"`,
			wantRuntime: false,
		},
		{
			name:      "literal - non-string type",
			expr:      &ast.Literal{Value: 123},
			wantError: true,
		},
		{
			name:      "call expression without generator",
			expr:      &ast.CallExpression{Callee: &ast.Identifier{Name: "input.symbol"}},
			wantError: true,
		},
		{
			name:      "unsupported expression type",
			expr:      &ast.BinaryExpression{Operator: "+"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewSecurityArgumentExtractor(tt.generator)
			result, err := extractor.ExtractSymbol(tt.expr)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got result: %+v", result)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", result.Code, tt.wantCode)
			}

			if result.IsRuntime != tt.wantRuntime {
				t.Errorf("IsRuntime = %v, want %v", result.IsRuntime, tt.wantRuntime)
			}
		})
	}
}

func TestExtractTimeframe_AllExpressionTypes(t *testing.T) {
	tests := []struct {
		name        string
		expr        ast.Expression
		generator   *generator
		wantCode    string
		wantRuntime bool
		wantError   bool
	}{
		{
			name:      "nil expression",
			expr:      nil,
			wantError: true,
		},
		{
			name: "identifier - unknown",
			expr: &ast.Identifier{Name: "unknownTf"},
			generator: &generator{
				variables: map[string]string{},
			},
			wantError: true,
		},
		{
			name: "identifier - string variable",
			expr: &ast.Identifier{Name: "myTimeframe"},
			generator: &generator{
				variables: map[string]string{"myTimeframe": "string"},
			},
			wantCode:    "myTimeframe",
			wantRuntime: true,
		},
		{
			name: "identifier - string constant",
			expr: &ast.Identifier{Name: "TF_CONST"},
			generator: &generator{
				variables: map[string]string{"TF_CONST": "string"},
				constants: map[string]interface{}{"TF_CONST": "1H"},
			},
			wantCode:    `"1H"`,
			wantRuntime: false,
		},
		{
			name: "identifier - non-string variable",
			expr: &ast.Identifier{Name: "intVar"},
			generator: &generator{
				variables: map[string]string{"intVar": "int"},
			},
			wantError: true,
		},
		{
			name: "member - timeframe.period",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "timeframe"},
				Property: &ast.Identifier{Name: "period"},
			},
			wantCode:    "ctx.Timeframe",
			wantRuntime: true,
		},
		{
			name: "member - unsupported",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "timeframe"},
				Property: &ast.Identifier{Name: "multiplier"},
			},
			wantError: true,
		},
		{
			name:        "literal - D normalized to 1D",
			expr:        &ast.Literal{Value: "D"},
			wantCode:    `"1D"`,
			wantRuntime: false,
		},
		{
			name:        "literal - W normalized to 1W",
			expr:        &ast.Literal{Value: "W"},
			wantCode:    `"1W"`,
			wantRuntime: false,
		},
		{
			name:        "literal - M normalized to 1M",
			expr:        &ast.Literal{Value: "M"},
			wantCode:    `"1M"`,
			wantRuntime: false,
		},
		{
			name:        "literal - already normalized",
			expr:        &ast.Literal{Value: "1D"},
			wantCode:    `"1D"`,
			wantRuntime: false,
		},
		{
			name:        "literal - intraday format",
			expr:        &ast.Literal{Value: "5m"},
			wantCode:    `"5m"`,
			wantRuntime: false,
		},
		{
			name:        "literal - quoted string stripped",
			expr:        &ast.Literal{Value: `"1D"`},
			wantCode:    `"1D"`,
			wantRuntime: false,
		},
		{
			name:      "literal - non-string type",
			expr:      &ast.Literal{Value: 60},
			wantError: true,
		},
		{
			name:      "call expression without generator",
			expr:      &ast.CallExpression{Callee: &ast.Identifier{Name: "input.timeframe"}},
			wantError: true,
		},
		{
			name:      "unsupported expression type",
			expr:      &ast.ConditionalExpression{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewSecurityArgumentExtractor(tt.generator)
			result, err := extractor.ExtractTimeframe(tt.expr)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got result: %+v", result)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", result.Code, tt.wantCode)
			}

			if result.IsRuntime != tt.wantRuntime {
				t.Errorf("IsRuntime = %v, want %v", result.IsRuntime, tt.wantRuntime)
			}
		})
	}
}

func TestTimeframeNormalization_AllFormats(t *testing.T) {
	extractor := NewSecurityArgumentExtractor(nil)

	tests := []struct {
		input    string
		expected string
	}{
		{"D", "1D"},
		{"W", "1W"},
		{"M", "1M"},
		{"1D", "1D"},
		{"1W", "1W"},
		{"1M", "1M"},
		{"5D", "5D"},
		{"2W", "2W"},
		{"3M", "3M"},
		{"1h", "1h"},
		{"4h", "4h"},
		{"1m", "1m"},
		{"5m", "5m"},
		{"15", "15"},
		{"60", "60"},
		{"", ""},
		{"d", "d"},
		{"w", "w"},
		{"m", "m"},
		{"custom", "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractor.normalizeTimeframe(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeTimeframe(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGeneratorContext_VariableResolutionBehavior(t *testing.T) {
	tests := []struct {
		name      string
		variables map[string]string
		constants map[string]interface{}
		testCases []struct {
			identName   string
			extractType string
			wantCode    string
			wantRuntime bool
			wantError   bool
		}
	}{
		{
			name:      "empty generator state",
			variables: map[string]string{},
			constants: map[string]interface{}{},
			testCases: []struct {
				identName   string
				extractType string
				wantCode    string
				wantRuntime bool
				wantError   bool
			}{
				{
					identName:   "unknown",
					extractType: "symbol",
					wantCode:    `"unknown"`,
					wantRuntime: false,
				},
				{
					identName:   "unknown",
					extractType: "timeframe",
					wantError:   true,
				},
			},
		},
		{
			name: "constants take precedence over variables",
			variables: map[string]string{
				"sym": "string",
				"tf":  "string",
			},
			constants: map[string]interface{}{
				"sym": "BTCUSDT",
				"tf":  "1H",
			},
			testCases: []struct {
				identName   string
				extractType string
				wantCode    string
				wantRuntime bool
				wantError   bool
			}{
				{
					identName:   "sym",
					extractType: "symbol",
					wantCode:    `"BTCUSDT"`,
					wantRuntime: false,
				},
				{
					identName:   "tf",
					extractType: "timeframe",
					wantCode:    `"1H"`,
					wantRuntime: false,
				},
			},
		},
		{
			name: "mixed variable types",
			variables: map[string]string{
				"stringSym": "string",
				"intVar":    "int",
				"boolVar":   "bool",
				"floatVar":  "float",
			},
			testCases: []struct {
				identName   string
				extractType string
				wantCode    string
				wantRuntime bool
				wantError   bool
			}{
				{
					identName:   "stringSym",
					extractType: "symbol",
					wantCode:    "stringSym",
					wantRuntime: true,
				},
				{
					identName:   "intVar",
					extractType: "symbol",
					wantCode:    `"intVar"`,
					wantRuntime: false,
				},
				{
					identName:   "intVar",
					extractType: "timeframe",
					wantError:   true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &generator{
				variables: tt.variables,
				constants: tt.constants,
			}
			extractor := NewSecurityArgumentExtractor(gen)

			for _, tc := range tt.testCases {
				t.Run(tc.identName+"_"+tc.extractType, func(t *testing.T) {
					ident := &ast.Identifier{Name: tc.identName}

					var result *ExtractionResult
					var err error

					if tc.extractType == "symbol" {
						result, err = extractor.ExtractSymbol(ident)
					} else {
						result, err = extractor.ExtractTimeframe(ident)
					}

					if tc.wantError {
						if err == nil {
							t.Errorf("expected error but got result: %+v", result)
						}
						return
					}

					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}

					if result.Code != tc.wantCode {
						t.Errorf("Code = %q, want %q", result.Code, tc.wantCode)
					}

					if result.IsRuntime != tc.wantRuntime {
						t.Errorf("IsRuntime = %v, want %v", result.IsRuntime, tc.wantRuntime)
					}
				})
			}
		})
	}
}

func TestErrorMessages_AllFailurePaths(t *testing.T) {
	tests := []struct {
		name             string
		generator        *generator
		expr             ast.Expression
		extractType      string
		expectedErrorMsg string
	}{
		{
			name:             "nil symbol expression",
			extractType:      "symbol",
			expr:             nil,
			expectedErrorMsg: "symbol expression is nil",
		},
		{
			name:             "nil timeframe expression",
			extractType:      "timeframe",
			expr:             nil,
			expectedErrorMsg: "timeframe expression is nil",
		},
		{
			name:             "unsupported symbol type",
			extractType:      "symbol",
			expr:             &ast.UnaryExpression{Operator: "!"},
			expectedErrorMsg: "unsupported symbol expression type",
		},
		{
			name:             "unsupported timeframe type",
			extractType:      "timeframe",
			expr:             &ast.ObjectExpression{},
			expectedErrorMsg: "unsupported timeframe expression type",
		},
		{
			name:             "call expression without generator context - symbol",
			extractType:      "symbol",
			expr:             &ast.CallExpression{Callee: &ast.Identifier{Name: "getSymbol"}},
			expectedErrorMsg: "requires generator context",
		},
		{
			name:             "call expression without generator context - timeframe",
			extractType:      "timeframe",
			expr:             &ast.CallExpression{Callee: &ast.Identifier{Name: "getTimeframe"}},
			expectedErrorMsg: "requires generator context",
		},
		{
			name:             "unknown timeframe identifier",
			extractType:      "timeframe",
			expr:             &ast.Identifier{Name: "unknownTf"},
			generator:        &generator{variables: map[string]string{}},
			expectedErrorMsg: "unsupported timeframe identifier",
		},
		{
			name:             "invalid symbol literal type",
			extractType:      "symbol",
			expr:             &ast.Literal{Value: true},
			expectedErrorMsg: "invalid symbol literal type",
		},
		{
			name:             "invalid timeframe literal type",
			extractType:      "timeframe",
			expr:             &ast.Literal{Value: 123.45},
			expectedErrorMsg: "invalid timeframe literal type",
		},
		{
			name:        "unsupported member expression - symbol",
			extractType: "symbol",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "wrong"},
				Property: &ast.Identifier{Name: "symbol"},
			},
			expectedErrorMsg: "unsupported member expression for symbol",
		},
		{
			name:        "unsupported member expression - timeframe",
			extractType: "timeframe",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "wrong"},
				Property: &ast.Identifier{Name: "timeframe"},
			},
			expectedErrorMsg: "unsupported member expression for timeframe",
		},
		{
			name:        "member expression nil object",
			extractType: "symbol",
			expr: &ast.MemberExpression{
				Object:   nil,
				Property: &ast.Identifier{Name: "tickerid"},
			},
			expectedErrorMsg: "nil object or property",
		},
		{
			name:        "member expression nil property",
			extractType: "symbol",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "syminfo"},
				Property: nil,
			},
			expectedErrorMsg: "nil object or property",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewSecurityArgumentExtractor(tt.generator)

			var err error
			if tt.extractType == "symbol" {
				_, err = extractor.ExtractSymbol(tt.expr)
			} else {
				_, err = extractor.ExtractTimeframe(tt.expr)
			}

			if err == nil {
				t.Errorf("expected error containing %q but got no error", tt.expectedErrorMsg)
				return
			}

			if !strings.Contains(err.Error(), tt.expectedErrorMsg) {
				t.Errorf("error = %q, want error containing %q", err.Error(), tt.expectedErrorMsg)
			}
		})
	}
}

func TestBoundaryConditions_ExtremeInputs(t *testing.T) {
	extractor := NewSecurityArgumentExtractor(nil)

	t.Run("very long symbol", func(t *testing.T) {
		longSymbol := strings.Repeat("A", 1000)
		result, err := extractor.ExtractSymbol(&ast.Literal{Value: longSymbol})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result.Code, longSymbol) {
			t.Error("long symbol not preserved")
		}
		if result.IsRuntime {
			t.Error("literal should not be runtime")
		}
	})

	t.Run("very long timeframe", func(t *testing.T) {
		longTf := strings.Repeat("1", 1000)
		result, err := extractor.ExtractTimeframe(&ast.Literal{Value: longTf})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result.Code, longTf) {
			t.Error("long timeframe not preserved")
		}
		if result.IsRuntime {
			t.Error("literal should not be runtime")
		}
	})

	t.Run("empty string symbol", func(t *testing.T) {
		result, err := extractor.ExtractSymbol(&ast.Literal{Value: ""})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Code != `""` {
			t.Errorf("empty symbol: got %q, want %q", result.Code, `""`)
		}
	})

	t.Run("empty string timeframe", func(t *testing.T) {
		result, err := extractor.ExtractTimeframe(&ast.Literal{Value: ""})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Code != `""` {
			t.Errorf("empty timeframe: got %q, want %q", result.Code, `""`)
		}
	})

	t.Run("unicode in symbol", func(t *testing.T) {
		unicode := "BTC日本語USDT"
		result, err := extractor.ExtractSymbol(&ast.Literal{Value: unicode})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result.Code, unicode) {
			t.Error("unicode symbol not preserved")
		}
	})

	t.Run("special characters in symbol", func(t *testing.T) {
		specialSymbols := []string{
			"BTC-USDT",
			"BINANCE:BTCUSDT",
			"BTC/USDT",
			"BTC.USDT",
			"BTC_USDT",
		}
		for _, sym := range specialSymbols {
			result, err := extractor.ExtractSymbol(&ast.Literal{Value: sym})
			if err != nil {
				t.Errorf("symbol %q: unexpected error: %v", sym, err)
				continue
			}
			if !strings.Contains(result.Code, sym) {
				t.Errorf("symbol %q not preserved in result %q", sym, result.Code)
			}
		}
	})
}
