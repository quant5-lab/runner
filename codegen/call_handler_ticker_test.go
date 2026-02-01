package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestTickerFunctionHandler_CanHandle validates ticker function recognition
 *
 * Tests that the TickerFunctionHandler correctly identifies all ticker modifier
 * functions (with/without namespace prefix) and defers non-ticker functions
 * to subsequent handlers in the call chain.
 */
func TestTickerFunctionHandler_CanHandle(t *testing.T) {
	handler := NewTickerFunctionHandler()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		// Chart type modifiers without prefix (global namespace)
		{"heikinashi", "heikinashi", true},
		{"renko", "renko", true},
		{"kagi", "kagi", true},
		{"linebreak", "linebreak", true},
		{"pointfigure", "pointfigure", true},

		// Chart type modifiers with ticker namespace
		{"ticker.heikinashi", "ticker.heikinashi", true},
		{"ticker.renko", "ticker.renko", true},
		{"ticker.kagi", "ticker.kagi", true},
		{"ticker.linebreak", "ticker.linebreak", true},
		{"ticker.pointfigure", "ticker.pointfigure", true},

		// Ticker ID manipulation functions
		{"ticker.new", "ticker.new", true},
		{"ticker.modify", "ticker.modify", true},
		{"ticker.standard", "ticker.standard", true},
		{"ticker.inherit", "ticker.inherit", true},

		// Non-ticker functions (should not handle)
		{"ta.sma", "ta.sma", false},
		{"strategy.entry", "strategy.entry", false},
		{"plot", "plot", false},
		{"security", "security", false},
		{"unknown_func", "unknown", false},
		{"empty", "", false},

		// Case sensitivity validation
		{"HEIKINASHI", "HEIKINASHI", false},
		{"Ticker.New", "Ticker.New", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.CanHandle(tt.funcName)
			if got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

/* TestTickerFunctionHandler_GenerateCode validates ticker function code generation
 *
 * Tests comprehensive symbol expression handling across all ticker modifier types,
 * validating proper translation to Go runtime ticker package calls with correct
 * argument extraction and type handling.
 */
func TestTickerFunctionHandler_GenerateCode(t *testing.T) {
	tests := []struct {
		name           string
		call           *ast.CallExpression
		setupVars      map[string]string
		expectError    bool
		validateOutput func(t *testing.T, code string)
	}{
		// Heikinashi: Simple ticker modifier with single symbol argument
		{
			name: "heikinashi with syminfo.tickerid",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "heikinashi"},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "syminfo"},
						Property: &ast.Identifier{Name: "tickerid"},
					},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if code != "ticker.Heikinashi(ctx.Symbol)" {
					t.Errorf("Expected ticker.Heikinashi(ctx.Symbol), got: %q", code)
				}
			},
		},
		{
			name: "heikinashi with string literal",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "heikinashi"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				expected := "ticker.Heikinashi(\"BTCUSDT\")"
				if code != expected {
					t.Errorf("Expected %q, got: %q", expected, code)
				}
			},
		},
		{
			name: "ticker.heikinashi with namespace prefix",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "heikinashi"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "ETHUSDT"},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "ticker.Heikinashi") {
					t.Error("Expected ticker.Heikinashi call")
				}
				if !strings.Contains(code, "ETHUSDT") {
					t.Error("Expected ETHUSDT symbol")
				}
			},
		},
		{
			name: "heikinashi with variable symbol",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "heikinashi"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "mySymbol"},
				},
			},
			setupVars: map[string]string{
				"mySymbol": "string",
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if code != "ticker.Heikinashi(mySymbol)" {
					t.Errorf("Expected ticker.Heikinashi(mySymbol), got: %q", code)
				}
			},
		},

		// Renko: Multi-parameter chart type with style and box size
		{
			name: "renko with all parameters",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "renko"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
					&ast.Literal{Value: "ATR"},
					&ast.Literal{Value: 14.0},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				expected := "ticker.Renko(\"BTCUSDT\", \"ATR\", float64(14))"
				if code != expected {
					t.Errorf("Expected %q, got: %q", expected, code)
				}
			},
		},
		{
			name: "renko with traditional style",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "renko"},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "syminfo"},
						Property: &ast.Identifier{Name: "tickerid"},
					},
					&ast.Literal{Value: "Traditional"},
					&ast.Literal{Value: 10.0},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "ticker.Renko") {
					t.Error("Expected ticker.Renko call")
				}
				if !strings.Contains(code, "ctx.Symbol") {
					t.Error("Expected ctx.Symbol for syminfo.tickerid")
				}
				if !strings.Contains(code, "Traditional") {
					t.Error("Expected Traditional style")
				}
			},
		},

		// Kagi: Reversal-based chart with reversal amount
		{
			name: "kagi with percentage reversal",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "kagi"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
					&ast.Literal{Value: 3.0},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				expected := "ticker.Kagi(\"BTCUSDT\", float64(3))"
				if code != expected {
					t.Errorf("Expected %q, got: %q", expected, code)
				}
			},
		},
		{
			name: "kagi with syminfo.ticker",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "kagi"},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "syminfo"},
						Property: &ast.Identifier{Name: "ticker"},
					},
					&ast.Literal{Value: 5.0},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "ctx.Symbol") {
					t.Error("Expected ctx.Symbol for syminfo.ticker")
				}
			},
		},

		// LineBreak: Three-line break chart
		{
			name: "linebreak with line count",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "linebreak"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
					&ast.Literal{Value: 3.0},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				expected := "ticker.LineBreak(\"BTCUSDT\", int(3))"
				if code != expected {
					t.Errorf("Expected %q, got: %q", expected, code)
				}
			},
		},

		// Ticker.new: Construct custom ticker ID
		{
			name: "ticker.new with exchange and symbol",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "new"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BINANCE"},
					&ast.Literal{Value: "BTCUSDT"},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				expected := "\"BINANCE\" + \":\" + \"BTCUSDT\""
				if code != expected {
					t.Errorf("Expected %q, got: %q", expected, code)
				}
			},
		},
		{
			name: "ticker.new with variables",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "new"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "exchange"},
					&ast.Identifier{Name: "symbol"},
				},
			},
			setupVars: map[string]string{
				"exchange": "string",
				"symbol":   "string",
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if code != "exchange + \":\" + symbol" {
					t.Errorf("Expected variable concatenation, got: %q", code)
				}
			},
		},

		// Ticker.standard: Extract base symbol from modified ticker
		{
			name: "ticker.standard with modified ticker ID",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "standard"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "HEIKINASHI:BTCUSDT"},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				expected := "ticker.NewModifierParser().ExtractBaseSymbol(\"HEIKINASHI:BTCUSDT\")"
				if code != expected {
					t.Errorf("Expected %q, got: %q", expected, code)
				}
			},
		},
		{
			name: "ticker.standard with no arguments",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "standard"},
				},
				Arguments: []ast.Expression{},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if code != "ctx.Symbol" {
					t.Errorf("Expected ctx.Symbol for no args, got: %q", code)
				}
			},
		},

		// Ticker.modify: Returns current symbol (no modification)
		{
			name: "ticker.modify returns current symbol",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "modify"},
				},
				Arguments: []ast.Expression{},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if code != "ctx.Symbol" {
					t.Errorf("Expected ctx.Symbol, got: %q", code)
				}
			},
		},

		// Ticker.inherit: Inherit context from another symbol
		{
			name: "ticker.inherit with modifier and target",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "inherit"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "HEIKINASHI"},
					&ast.Literal{Value: "ETHUSDT"},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "ctx.Symbol") {
					t.Error("Expected ctx.Symbol prefix")
				}
				if !strings.Contains(code, "ETHUSDT") {
					t.Error("Expected target symbol ETHUSDT")
				}
			},
		},

		// Error cases: Insufficient arguments
		{
			name: "heikinashi missing symbol argument",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "heikinashi"},
				Arguments: []ast.Expression{},
			},
			expectError: true,
		},
		{
			name: "renko missing parameters",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "renko"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
				},
			},
			expectError: true,
		},
		{
			name: "kagi missing reversal amount",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "kagi"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
				},
			},
			expectError: true,
		},
		{
			name: "linebreak missing line count",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "linebreak"},
				Arguments: []ast.Expression{},
			},
			expectError: true,
		},
		{
			name: "ticker.new missing symbol",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "new"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BINANCE"},
				},
			},
			expectError: true,
		},
		{
			name: "ticker.inherit missing target symbol",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "inherit"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "HEIKINASHI"},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()

			// Setup variable context if needed
			if tt.setupVars != nil {
				for name, typ := range tt.setupVars {
					g.variables[name] = typ
				}
			}

			handler := NewTickerFunctionHandler()
			code, err := handler.GenerateCode(g, tt.call)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if code == "" {
				t.Error("Expected generated code, got empty string")
			}

			if tt.validateOutput != nil {
				tt.validateOutput(t, code)
			}
		})
	}
}

/* TestTickerFunctionHandler_NilSafety validates graceful handling of malformed AST nodes
 *
 * Tests that the handler properly handles nil or invalid AST structures without
 * panicking, returning appropriate error codes or empty results.
 */
func TestTickerFunctionHandler_NilSafety(t *testing.T) {
	handler := NewTickerFunctionHandler()
	g := newTestGenerator()

	tests := []struct {
		name        string
		call        *ast.CallExpression
		expectEmpty bool
	}{
		{
			name: "nil callee",
			call: &ast.CallExpression{
				Callee:    nil,
				Arguments: []ast.Expression{},
			},
			expectEmpty: true,
		},
		{
			name: "valid ticker call",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "heikinashi"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
				},
			},
			expectEmpty: false,
		},
		{
			name: "non-ticker function should return empty",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.sma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := handler.GenerateCode(g, tt.call)

			if tt.expectEmpty {
				if code != "" && err == nil {
					t.Errorf("Expected empty code for %q, got: %q", tt.name, code)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if code == "" {
					t.Error("Expected non-empty code")
				}
			}
		})
	}
}

/* TestTickerFunctionHandler_SymbolExpressionTypes validates all supported symbol input types
 *
 * Tests that the handler correctly processes different types of symbol expressions:
 * - Literal strings (explicit symbols)
 * - MemberExpression (syminfo.tickerid, syminfo.ticker)
 * - Identifier (variable references)
 * - Constants (resolved at compile time)
 */
func TestTickerFunctionHandler_SymbolExpressionTypes(t *testing.T) {
	tests := []struct {
		name           string
		symbolExpr     ast.Expression
		setupVars      map[string]string
		setupConsts    map[string]interface{}
		expectedSymbol string
	}{
		{
			name:           "literal string symbol",
			symbolExpr:     &ast.Literal{Value: "BTCUSDT"},
			expectedSymbol: "\"BTCUSDT\"",
		},
		{
			name: "syminfo.tickerid member expression",
			symbolExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "syminfo"},
				Property: &ast.Identifier{Name: "tickerid"},
			},
			expectedSymbol: "ctx.Symbol",
		},
		{
			name: "syminfo.ticker member expression",
			symbolExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "syminfo"},
				Property: &ast.Identifier{Name: "ticker"},
			},
			expectedSymbol: "ctx.Symbol",
		},
		{
			name:       "string variable identifier",
			symbolExpr: &ast.Identifier{Name: "mySymbol"},
			setupVars: map[string]string{
				"mySymbol": "string",
			},
			expectedSymbol: "mySymbol",
		},
		{
			name:       "constant string identifier",
			symbolExpr: &ast.Identifier{Name: "SYMBOL_CONST"},
			setupConsts: map[string]interface{}{
				"SYMBOL_CONST": "ETHUSDT",
			},
			expectedSymbol: "\"ETHUSDT\"",
		},
		{
			name:           "plain identifier (treated as string literal)",
			symbolExpr:     &ast.Identifier{Name: "AAPL"},
			expectedSymbol: "\"AAPL\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()

			// Setup test context
			if tt.setupVars != nil {
				for name, typ := range tt.setupVars {
					g.variables[name] = typ
				}
			}
			if tt.setupConsts != nil {
				for name, val := range tt.setupConsts {
					g.constants[name] = val
				}
			}

			handler := NewTickerFunctionHandler()
			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "heikinashi"},
				Arguments: []ast.Expression{tt.symbolExpr},
			}

			code, err := handler.GenerateCode(g, call)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			expectedCode := "ticker.Heikinashi(" + tt.expectedSymbol + ")"
			if code != expectedCode {
				t.Errorf("Expected %q, got: %q", expectedCode, code)
			}
		})
	}
}

/* TestTickerFunctionHandler_IntegrationWithCallRouter validates handler chain behavior
 *
 * Tests that the ticker handler properly integrates with the call expression router,
 * including precedence ordering and fallback to unknown function handler.
 */
func TestTickerFunctionHandler_IntegrationWithCallRouter(t *testing.T) {
	router := NewCallExpressionRouter()
	g := newTestGenerator()

	tests := []struct {
		name           string
		call           *ast.CallExpression
		expectHandled  bool
		validateOutput func(t *testing.T, code string)
	}{
		{
			name: "ticker function handled by ticker handler",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "heikinashi"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
				},
			},
			expectHandled: true,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "ticker.Heikinashi") {
					t.Error("Expected ticker handler to generate code")
				}
			},
		},
		{
			name: "non-ticker function bypasses ticker handler",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "plotshape"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "condition"},
				},
			},
			expectHandled: true,
			validateOutput: func(t *testing.T, code string) {
				if strings.Contains(code, "ticker.") {
					t.Error("Ticker handler should not handle non-ticker functions")
				}
				// Should fall through to UnknownFunctionHandler
				if !strings.Contains(code, "//") {
					t.Error("Expected TODO comment from UnknownFunctionHandler")
				}
			},
		},
		{
			name: "math function bypasses ticker handler",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "abs"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "value"},
				},
			},
			expectHandled: true,
			validateOutput: func(t *testing.T, code string) {
				if strings.Contains(code, "ticker.") {
					t.Error("Ticker handler should not handle math functions")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := router.RouteCall(g, tt.call)

			if err != nil {
				t.Fatalf("Router error: %v", err)
			}

			if tt.expectHandled && code == "" {
				t.Error("Expected handler to generate code")
			}

			if tt.validateOutput != nil {
				tt.validateOutput(t, code)
			}
		})
	}
}

/* TestTickerFunctionHandler_GeneratorStateManagement validates hasTickerCalls flag
 *
 * Tests that the handler correctly sets the generator's hasTickerCalls flag,
 * which triggers import of the runtime ticker package in generated code.
 */
func TestTickerFunctionHandler_GeneratorStateManagement(t *testing.T) {
	tests := []struct {
		name               string
		call               *ast.CallExpression
		expectTickerImport bool
	}{
		{
			name: "ticker function sets hasTickerCalls flag",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "heikinashi"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
				},
			},
			expectTickerImport: true,
		},
		{
			name: "ticker.new sets hasTickerCalls flag",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "new"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BINANCE"},
					&ast.Literal{Value: "BTCUSDT"},
				},
			},
			expectTickerImport: true,
		},
		{
			name: "ticker.standard sets hasTickerCalls flag",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "standard"},
				},
				Arguments: []ast.Expression{},
			},
			expectTickerImport: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			handler := NewTickerFunctionHandler()

			// Verify initial state
			if g.hasTickerCalls {
				t.Error("Generator should not have hasTickerCalls set initially")
			}

			_, err := handler.GenerateCode(g, tt.call)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify flag was set
			if tt.expectTickerImport && !g.hasTickerCalls {
				t.Error("Expected hasTickerCalls flag to be set")
			}
		})
	}
}
