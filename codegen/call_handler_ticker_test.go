package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTickerFunctionHandler_CanHandle(t *testing.T) {
	handler := NewTickerFunctionHandler()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"heikinashi", "heikinashi", true},
		{"renko", "renko", true},
		{"kagi", "kagi", true},
		{"linebreak", "linebreak", true},
		{"pointfigure", "pointfigure", true},

		{"ticker.heikinashi", "ticker.heikinashi", true},
		{"ticker.renko", "ticker.renko", true},
		{"ticker.kagi", "ticker.kagi", true},
		{"ticker.linebreak", "ticker.linebreak", true},
		{"ticker.pointfigure", "ticker.pointfigure", true},
		{"range", "range", true},
		{"ticker.range", "ticker.range", true},

		{"ticker.new", "ticker.new", true},
		{"ticker.modify", "ticker.modify", true},
		{"ticker.standard", "ticker.standard", true},
		{"ticker.inherit", "ticker.inherit", true},

		{"ta.sma", "ta.sma", false},
		{"strategy.entry", "strategy.entry", false},
		{"plot", "plot", false},
		{"security", "security", false},
		{"unknown_func", "unknown", false},
		{"empty", "", false},

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

func TestTickerFunctionHandler_GenerateCode(t *testing.T) {
	tests := []struct {
		name           string
		call           *ast.CallExpression
		setupVars      map[string]string
		expectError    bool
		validateOutput func(t *testing.T, code string)
	}{
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
				expected := `ticker.New("BINANCE", "BTCUSDT", "", "", "", "")`
				if code != expected {
					t.Errorf("Expected %q, got: %q", expected, code)
				}
			},
		},
		{
			name: "ticker.new with session and adjustment",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "new"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BINANCE"},
					&ast.Literal{Value: "BTCUSDT"},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "session"},
						Property: &ast.Identifier{Name: "regular"},
					},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "adjustment"},
						Property: &ast.Identifier{Name: "splits"},
					},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				expected := `ticker.New("BINANCE", "BTCUSDT", ticker.SessionRegular, ticker.AdjustmentSplits, "", "")`
				if code != expected {
					t.Errorf("Expected %q, got: %q", expected, code)
				}
			},
		},
		{
			name: "ticker.new with all six arguments",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "new"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "NYSE"},
					&ast.Literal{Value: "AAPL"},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "session"},
						Property: &ast.Identifier{Name: "extended"},
					},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "adjustment"},
						Property: &ast.Identifier{Name: "dividends"},
					},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "backadjustment"},
						Property: &ast.Identifier{Name: "on"},
					},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "settlement_as_close"},
						Property: &ast.Identifier{Name: "off"},
					},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				expected := `ticker.New("NYSE", "AAPL", ticker.SessionExtended, ticker.AdjustmentDividends, ticker.BackAdjustmentOn, ticker.SettlementOff)`
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
				expected := `ticker.New(exchange, symbol, "", "", "", "")`
				if code != expected {
					t.Errorf("Expected %q, got: %q", expected, code)
				}
			},
		},
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
				expected := "ticker.Standard(\"HEIKINASHI:BTCUSDT\")"
				if code != expected {
					t.Errorf("Expected %q, got: %q", expected, code)
				}
			},
		},
		{
			name: "ticker.standard with syminfo.tickerid",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "standard"},
				},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "syminfo"},
						Property: &ast.Identifier{Name: "tickerid"},
					},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				expected := "ticker.Standard(ctx.Symbol)"
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
		{
			name: "ticker.modify with session",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "modify"},
				},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "syminfo"},
						Property: &ast.Identifier{Name: "tickerid"},
					},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "session"},
						Property: &ast.Identifier{Name: "extended"},
					},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				expected := `ticker.Modify(ctx.Symbol, ticker.SessionExtended, "", "", "")`
				if code != expected {
					t.Errorf("Expected %q, got: %q", expected, code)
				}
			},
		},
		{
			name: "ticker.modify with tickerid only",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "modify"},
				},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "syminfo"},
						Property: &ast.Identifier{Name: "tickerid"},
					},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				expected := `ticker.Modify(ctx.Symbol, "", "", "", "")`
				if code != expected {
					t.Errorf("Expected %q, got: %q", expected, code)
				}
			},
		},
		{
			name: "ticker.modify with all four modifiers",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "modify"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BINANCE:BTCUSDT"},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "session"},
						Property: &ast.Identifier{Name: "regular"},
					},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "adjustment"},
						Property: &ast.Identifier{Name: "none"},
					},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "backadjustment"},
						Property: &ast.Identifier{Name: "off"},
					},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "settlement_as_close"},
						Property: &ast.Identifier{Name: "on"},
					},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				expected := `ticker.Modify("BINANCE:BTCUSDT", ticker.SessionRegular, ticker.AdjustmentNone, ticker.BackAdjustmentOff, ticker.SettlementOn)`
				if code != expected {
					t.Errorf("Expected %q, got: %q", expected, code)
				}
			},
		},
		{
			name: "ticker.inherit with source and target",
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
				expected := `ticker.Inherit("HEIKINASHI", "ETHUSDT")`
				if code != expected {
					t.Errorf("Expected %q, got: %q", expected, code)
				}
			},
		},
		{
			name: "ticker.inherit with syminfo expressions",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "inherit"},
				},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "syminfo"},
						Property: &ast.Identifier{Name: "tickerid"},
					},
					&ast.Literal{Value: "ETHUSDT"},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				expected := `ticker.Inherit(ctx.Symbol, "ETHUSDT")`
				if code != expected {
					t.Errorf("Expected %q, got: %q", expected, code)
				}
			},
		},
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
		{
			name: "ticker.modify missing tickerid",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "modify"},
				},
				Arguments: []ast.Expression{},
			},
			expectError: true,
		},
		{
			name: "pointfigure missing required args",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "pointfigure"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
				},
			},
			expectError: true,
		},
		{
			name: "pointfigure with all parameters",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "pointfigure"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: "ATR"},
					&ast.Literal{Value: 14.0},
					&ast.Literal{Value: 3.0},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "ticker.PointFigure") {
					t.Error("Expected ticker.PointFigure call")
				}
				if !strings.Contains(code, "BTCUSDT") {
					t.Error("Expected symbol BTCUSDT")
				}
				if !strings.Contains(code, "ATR") {
					t.Error("Expected style ATR")
				}
			},
		},
		{
			name: "pointfigure with minimum args uses defaults",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "pointfigure"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "AAPL"},
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: "Traditional"},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "ticker.PointFigure") {
					t.Error("Expected ticker.PointFigure call")
				}
				if !strings.Contains(code, "float64(14)") {
					t.Error("Expected default param 14")
				}
				if !strings.Contains(code, "float64(3)") {
					t.Error("Expected default reversal 3")
				}
			},
		},
		{
			name: "pointfigure with syminfo.tickerid",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "pointfigure"},
				},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "syminfo"},
						Property: &ast.Identifier{Name: "tickerid"},
					},
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: "ATR"},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "ctx.Symbol") {
					t.Error("Expected ctx.Symbol for syminfo.tickerid")
				}
				if !strings.Contains(code, "ticker.PointFigure") {
					t.Error("Expected ticker.PointFigure call")
				}
			},
		},
		{
			name: "ticker.range with literal symbol",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "range"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				expected := `ticker.Range("BTCUSDT")`
				if code != expected {
					t.Errorf("Expected %q, got: %q", expected, code)
				}
			},
		},
		{
			name: "ticker.range with syminfo.tickerid",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "range"},
				},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "syminfo"},
						Property: &ast.Identifier{Name: "tickerid"},
					},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				expected := "ticker.Range(ctx.Symbol)"
				if code != expected {
					t.Errorf("Expected %q, got: %q", expected, code)
				}
			},
		},
		{
			name: "bare range with literal symbol",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "range"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "AAPL"},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				expected := `ticker.Range("AAPL")`
				if code != expected {
					t.Errorf("Expected %q, got: %q", expected, code)
				}
			},
		},
		{
			name: "ticker.range missing symbol argument",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "range"},
				},
				Arguments: []ast.Expression{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()

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
		{
			name: "ticker.range sets hasTickerCalls flag",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "range"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
				},
			},
			expectTickerImport: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			handler := NewTickerFunctionHandler()

			if g.hasTickerCalls {
				t.Error("Generator should not have hasTickerCalls set initially")
			}

			_, err := handler.GenerateCode(g, tt.call)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.expectTickerImport && !g.hasTickerCalls {
				t.Error("Expected hasTickerCalls flag to be set")
			}
		})
	}
}
