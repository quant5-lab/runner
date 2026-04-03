package security

import (
	"testing"
)

/* TestAnalyzeAST_ArrowFunctionSecurity verifies security() detection across arrow function body shapes */
func TestAnalyzeAST_ArrowFunctionSecurity(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		expectedCount int
	}{
		{
			name: "single_line_body",
			code: `
indicator("Test")
getHTFClose() => request.security(syminfo.tickerid, '1D', close)
`,
			expectedCount: 1,
		},
		{
			name: "multi_line_body",
			code: `
indicator("Test")
getHTF(tf) =>
    val = request.security(syminfo.tickerid, tf, close)
    val * 2
`,
			expectedCount: 1,
		},
		{
			name: "conditional_expression_body",
			code: `
indicator("Test")
getVal(useDaily) =>
    useDaily ? request.security(syminfo.tickerid, '1D', close) : close
`,
			expectedCount: 1,
		},
		{
			name: "legacy_security_call",
			code: `
indicator("Test")
getHTF() => security(syminfo.tickerid, '1D', close)
`,
			expectedCount: 1,
		},
		{
			name: "mixed_security_and_ta",
			code: `
indicator("Test")
getSmoothed() =>
    raw = request.security(syminfo.tickerid, '1D', close)
    ta.sma(raw, 10)
`,
			expectedCount: 1,
		},
		{
			name: "no_security_ta_only",
			code: `
indicator("Test")
calcSMA(src, len) => ta.sma(src, len)
`,
			expectedCount: 0,
		},
		{
			name: "no_security_arithmetic_only",
			code: `
indicator("Test")
double(x) => x * 2
`,
			expectedCount: 0,
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

/* TestAnalyzeAST_ArrowFunctionSecurityCoexistence verifies detection across mixed top-level and arrow scopes */
func TestAnalyzeAST_ArrowFunctionSecurityCoexistence(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		expectedCount int
	}{
		{
			name: "top_level_and_arrow",
			code: `
indicator("Test")
daily = request.security("BTCUSDT", "1D", close)
getHTFHigh() => request.security(syminfo.tickerid, '4h', high)
`,
			expectedCount: 2,
		},
		{
			name: "multiple_arrow_functions",
			code: `
indicator("Test")
getDaily() => request.security(syminfo.tickerid, '1D', close)
getWeekly() => request.security(syminfo.tickerid, '1W', close)
`,
			expectedCount: 2,
		},
		{
			name: "arrow_without_security_alongside_arrow_with",
			code: `
indicator("Test")
calcSMA(src) => ta.sma(src, 20)
getHTF() => request.security(syminfo.tickerid, '1D', close)
`,
			expectedCount: 1,
		},
		{
			name: "top_level_only_no_arrow",
			code: `
indicator("Test")
calcSMA(src) => ta.sma(src, 20)
daily = request.security("BTCUSDT", "1D", close)
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
