package security

import (
	"testing"
)

/* TestAnalyzeAST_ExpressionNesting verifies security() detection in all expression contexts */
func TestAnalyzeAST_ExpressionNesting(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		expectedCount int
	}{
		{
			name: "conditional_ternary",
			code: `
indicator("Test")
h = input(false)
src = h ? request.security(syminfo.tickerid, "1D", close) : close
`,
			expectedCount: 1,
		},
		{
			name: "conditional_both_branches",
			code: `
indicator("Test")
src = condition ? security("BTCUSDT", "1D", close) : security("ETHUSDT", "1h", high)
`,
			expectedCount: 2,
		},
		{
			name: "conditional_test_position",
			code: `
indicator("Test")
result = security("BTCUSDT", "1D", close) > 100 ? 1 : 0
`,
			expectedCount: 1,
		},
		{
			name: "binary_expression_left",
			code: `
indicator("Test")
combined = request.security("BTCUSDT", "1D", close) + close
`,
			expectedCount: 1,
		},
		{
			name: "binary_expression_right",
			code: `
indicator("Test")
combined = close + request.security("BTCUSDT", "1D", high)
`,
			expectedCount: 1,
		},
		{
			name: "binary_expression_both",
			code: `
indicator("Test")
combined = request.security("BTCUSDT", "1D", close) + security("ETHUSDT", "1h", high)
`,
			expectedCount: 2,
		},
		{
			name: "unary_expression_negation",
			code: `
indicator("Test")
negated = not security("BTCUSDT", "1D", close > open)
`,
			expectedCount: 1,
		},
		{
			name: "unary_expression_numeric",
			code: `
indicator("Test")
negative = -security("BTCUSDT", "1D", close)
`,
			expectedCount: 1,
		},
		{
			name: "function_argument",
			code: `
indicator("Test")
smoothed = ta.sma(request.security("BTCUSDT", "1D", close), 20)
`,
			expectedCount: 1,
		},
		{
			name: "nested_function_arguments",
			code: `
indicator("Test")
result = ta.sma(request.security("BTCUSDT", "1D", ta.ema(security("ETHUSDT", "1h", close), 10)), 20)
`,
			expectedCount: 2,
		},
		{
			name: "logical_and_expression",
			code: `
indicator("Test")
cond = close > 50 and security("BTCUSDT", "1D", high) > 60000
`,
			expectedCount: 1,
		},
		{
			name: "logical_or_expression",
			code: `
indicator("Test")
cond = security("BTCUSDT", "1D", close) < 30000 or security("ETHUSDT", "1h", high) < 2000
`,
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := parseCode(t, tt.code)
			calls := AnalyzeAST(program)

			if len(calls) != tt.expectedCount {
				t.Errorf("expected %d security calls, got %d", tt.expectedCount, len(calls))
			}

			for i, call := range calls {
				if call.Symbol == "" {
					t.Errorf("Call %d: symbol should not be empty", i)
				}
				if call.Timeframe == "" {
					t.Errorf("Call %d: timeframe should not be empty", i)
				}
				if call.Expression == nil {
					t.Errorf("Call %d: expression should not be nil", i)
				}
			}
		})
	}
}

/* TestAnalyzeAST_StatementTypes verifies security() detection across statement types */
func TestAnalyzeAST_StatementTypes(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		expectedCount int
	}{
		{
			name: "variable_declaration",
			code: `
indicator("Test")
dailyClose = request.security("BTCUSDT", "1D", close)
`,
			expectedCount: 1,
		},
		{
			name: "expression_statement",
			code: `
indicator("Test")
plot(request.security("BTCUSDT", "1D", close))
`,
			expectedCount: 1,
		},
		{
			name: "if_statement_test",
			code: `
indicator("Test")
if request.security("BTCUSDT", "1D", close) > 100
    x = 1
`,
			expectedCount: 1,
		},
		{
			name: "if_statement_body",
			code: `
indicator("Test")
if close > 100
    dailyClose = request.security("BTCUSDT", "1D", close)
`,
			expectedCount: 1,
		},
		{
			name: "if_else_both",
			code: `
indicator("Test")
if close > 100
    x = request.security("BTCUSDT", "1D", close)
else
    y = security("ETHUSDT", "1h", high)
`,
			expectedCount: 2,
		},
		{
			name: "for_loop_range",
			code: `
indicator("Test")
for i = 0 to request.security("BTCUSDT", "1D", close)
    x = i
`,
			expectedCount: 1,
		},
		{
			name: "for_loop_body",
			code: `
indicator("Test")
for i = 0 to 10
    x = request.security("BTCUSDT", "1D", close)
`,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := parseCode(t, tt.code)
			calls := AnalyzeAST(program)

			if len(calls) != tt.expectedCount {
				t.Errorf("expected %d security calls, got %d", tt.expectedCount, len(calls))
			}
		})
	}
}

/* TestAnalyzeAST_ComplexNesting verifies deeply nested security() detection */
func TestAnalyzeAST_ComplexNesting(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		expectedCount int
		minDepth      int
	}{
		{
			name: "triple_nesting",
			code: `
indicator("Test")
result = ta.sma(close > 50 ? request.security("BTCUSDT", "1D", ta.ema(close, 10)) : close, 20)
`,
			expectedCount: 1,
			minDepth:      3,
		},
		{
			name: "multiple_at_different_depths",
			code: `
indicator("Test")
x = request.security("BTCUSDT", "1D", close)
y = ta.sma(security("ETHUSDT", "1h", close), 20)
z = close > 100 ? security("BNBUSDT", "1W", high) : 0
`,
			expectedCount: 3,
			minDepth:      1,
		},
		{
			name: "conditional_with_nested_calls",
			code: `
indicator("Test")
result = close > request.security("BTCUSDT", "1D", high) ? 
    security("ETHUSDT", "1h", close) : 
    security("BNBUSDT", "1W", low)
`,
			expectedCount: 3,
			minDepth:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := parseCode(t, tt.code)
			calls := AnalyzeAST(program)

			if len(calls) != tt.expectedCount {
				t.Errorf("Expected %d security calls, got %d", tt.expectedCount, len(calls))
			}

			for i, call := range calls {
				if call.Symbol == "" || call.Timeframe == "" || call.Expression == nil {
					t.Errorf("Call %d: incomplete security call data", i)
				}
			}
		})
	}
}

/* TestAnalyzeAST_SymbolAndTimeframeVariations verifies parameter extraction accuracy */
func TestAnalyzeAST_SymbolAndTimeframeVariations(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		expectedSymbol string
		expectedTF     string
	}{
		{
			name: "literal_symbol_literal_tf",
			code: `
indicator("Test")
x = request.security("BTCUSDT", "1D", close)
`,
			expectedSymbol: "BTCUSDT",
			expectedTF:     "1D",
		},
		{
			name: "runtime_symbol",
			code: `
indicator("Test")
x = request.security(syminfo.tickerid, "1h", close)
`,
			expectedSymbol: "syminfo.tickerid",
			expectedTF:     "1h",
		},
		{
			name: "runtime_timeframe",
			code: `
indicator("Test")
x = request.security("BTCUSDT", timeframe.period, close)
`,
			expectedSymbol: "BTCUSDT",
			expectedTF:     "timeframe.period",
		},
		{
			name: "both_runtime",
			code: `
indicator("Test")
x = request.security(syminfo.tickerid, timeframe.period, close)
`,
			expectedSymbol: "syminfo.tickerid",
			expectedTF:     "timeframe.period",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := parseCode(t, tt.code)
			calls := AnalyzeAST(program)

			if len(calls) != 1 {
				t.Fatalf("Expected 1 security call, got %d", len(calls))
			}

			call := calls[0]
			if call.Symbol != tt.expectedSymbol {
				t.Errorf("Expected symbol '%s', got '%s'", tt.expectedSymbol, call.Symbol)
			}
			if call.Timeframe != tt.expectedTF {
				t.Errorf("Expected timeframe '%s', got '%s'", tt.expectedTF, call.Timeframe)
			}
		})
	}
}
