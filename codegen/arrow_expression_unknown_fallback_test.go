package codegen

import (
	"strings"
	"testing"
)

// TestArrowExpressionPosition_UnknownCall_DegradesToNaN verifies that an unknown
// function call in arrow expression position produces math.NaN() — a valid Go
// expression — rather than a comment stub or an error.  The contract covers all
// syntactic positions in which the call can appear as an expression value.
func TestArrowExpressionPosition_UnknownCall_DegradesToNaN(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{
			name: "bare unknown as sole return expression",
			script: `//@version=5
indicator("Test")
f(src) => unknownFunc(src)
plot(f(close), "f")
`,
		},
		{
			name: "namespaced unknown as sole return expression",
			script: `//@version=5
indicator("Test")
f(src) => request.dividends(src)
plot(f(close), "f")
`,
		},
		{
			name: "unknown as left operand in additive expression",
			script: `//@version=5
indicator("Test")
f(src) => unknownFunc(src) + 1.0
plot(f(close), "f")
`,
		},
		{
			name: "unknown as right operand in multiplicative expression",
			script: `//@version=5
indicator("Test")
f(src) => src * unknownFunc(src)
plot(f(close), "f")
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.script)
			if err != nil {
				t.Fatalf("compilePineScript must not error for unknown call in arrow expression position: %v", err)
			}
			if strings.Contains(code, "// TODO") {
				t.Errorf("generated code must not embed TODO comment as expression value:\n%s", code)
			}
			if !strings.Contains(code, "math.NaN()") {
				t.Errorf("generated code must contain math.NaN() stub for unknown function:\n%s", code)
			}
		})
	}
}

// TestArrowExpressionPosition_KnownTACall_NotDegraded verifies the boundary between
// implemented and unimplemented functions: known TA calls must never produce a TODO stub.
// math.NaN() may appear in the output as a warmup-period sentinel; that is expected.
func TestArrowExpressionPosition_KnownTACall_NotDegraded(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{
			name: "ta.sma in arrow body",
			script: `//@version=5
indicator("Test")
f(src) => ta.sma(src, 3)
plot(f(close), "f")
`,
		},
		{
			name: "ta.ema in arrow body",
			script: `//@version=5
indicator("Test")
f(src) => ta.ema(src, 5)
plot(f(close), "f")
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.script)
			if err != nil {
				t.Fatalf("compilePineScript must not error for known TA call in arrow body: %v", err)
			}
			if strings.Contains(code, "// TODO") {
				t.Errorf("known TA function must not produce TODO stub:\n%s", code)
			}
		})
	}
}
