package codegen

import (
	"strings"
	"testing"
)

// TestArrowTupleDispatch covers tuple destructuring in all contexts:
// top-level, inside arrow body, and the full 3-way dispatch hierarchy
// (UDF → TA → security → generic fallback).
func TestArrowTupleDispatch(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		mustContain []string
		mustAvoid   []string
	}{
		{
			name: "UDF tuple at top level emits direct call not temp_ vars",
			script: `//@version=5
strategy("t")
pair(a, b) =>
    [a + b, a - b]
[s, d] = pair(close, open)
plot(s)
`,
			mustContain: []string{"s, d := pair("},
			mustAvoid:   []string{"temp_s", "temp_d"},
		},
		{
			name: "UDF tuple inside arrow body emits direct call",
			script: `//@version=5
strategy("t")
inner(a, b) =>
    [a * b, a / b]
outer(x) =>
    [p, q] = inner(x, 2.0)
    p + q
plot(outer(close))
`,
			mustContain: []string{"p, q := inner("},
			mustAvoid:   []string{"temp_p", "temp_q"},
		},
		{
			name: "TA tuple (ta.macd) dispatches through TA handler not generic",
			script: `//@version=5
strategy("t")
[macdLine, signalLine, histLine] = ta.macd(close, 12, 26, 9)
plot(macdLine)
`,
			mustContain: []string{"macdLine", "signalLine", "histLine"},
			mustAvoid:   []string{"temp_macdLine"},
		},
		{
			// Top-level unknown tuples emit a TODO stub with zeroed series assignments.
			// The temp_ dual-storage pattern applies only inside arrow bodies.
			name: "unknown-function tuple at top level emits TODO stub with zero series",
			script: `//@version=5
strategy("t")
[x, y] = unknownTupleFunc(close)
plot(x)
`,
			mustContain: []string{"unknownTupleFunc", "xSeries.Set(0.0)", "ySeries.Set(0.0)"},
			mustAvoid:   []string{"temp_x", "temp_y"},
		},
		{
			// Inside an arrow body, unknown tuple functions use the temp_ dual-storage
			// pattern: temp vars carry the raw value while named vars track series state.
			name: "unknown-function tuple inside arrow body uses temp_ dual-storage pattern",
			script: `//@version=5
strategy("t")
outer(x) =>
    [p, q] = unknownTupleInArrow(x)
    p + q
plot(outer(close))
`,
			mustContain: []string{"temp_p", "temp_q", "pSeries.Set(p)", "qSeries.Set(q)"},
		},
		{
			name: "multi-return UDF with outer-scope series captures passes captures at call site",
			script: `//@version=5
strategy("t")
indicator_pair(src, len) =>
    fast = ta.ema(src, len)
    slow = ta.ema(src, len * 2)
    [fast, slow]
[f, s] = indicator_pair(close, 14)
plot(f)
`,
			mustContain: []string{"f, s := indicator_pair("},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.script)
			if err != nil {
				t.Fatalf("compilePineScript: %v", err)
			}
			v := NewCodeVerifier(code, t)
			v.MustContain(tt.mustContain...)
			v.MustNotContain(tt.mustAvoid...)
		})
	}
}

// TestArrowTupleDispatch_SecurityTuple verifies that a security() call with an array
// literal as the third argument dispatches through the security tuple handler.
func TestArrowTupleDispatch_SecurityTuple(t *testing.T) {
	const script = `//@version=5
strategy("t")
[h, l] = security("AAPL", "D", [high, low])
plot(h)
`
	code, err := compilePineScript(script)
	if err != nil {
		t.Fatalf("compilePineScript: %v", err)
	}

	if !strings.Contains(code, "hSeries") || !strings.Contains(code, "lSeries") {
		t.Errorf("expected series declarations for h and l in:\n%s", code)
	}
	if !strings.Contains(code, "securityContexts") {
		t.Errorf("expected security context lookup in:\n%s", code)
	}
}
