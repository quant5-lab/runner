package codegen

import (
	"strings"
	"testing"
)

func TestSeriesInitCoercer_Coerce(t *testing.T) {
	coercer := NewSeriesInitCoercer()

	tests := []struct {
		name   string
		code   string
		goType GoValueType
		check  func(t *testing.T, result string)
	}{
		/* float64 pass-through — output equals input */
		{"float64 literal", "0.01", GoFloat64, expectExact("0.01")},
		{"float64 NaN", "math.NaN()", GoFloat64, expectExact("math.NaN()")},
		{"float64 function call", "float64(context.TimeframeMultiplier(ctx.Timeframe))", GoFloat64,
			expectExact("float64(context.TimeframeMultiplier(ctx.Timeframe))")},
		{"float64 series accessor", "closeSeries.GetCurrent()", GoFloat64, expectExact("closeSeries.GetCurrent()")},
		{"float64 empty string", "", GoFloat64, expectExact("")},

		/* bool → IIFE wrapping */
		{"bool true literal", "true", GoBool, expectBoolIIFE("true")},
		{"bool false literal", "false", GoBool, expectBoolIIFE("false")},
		{"bool comparison", "(ctx.BarIndex == 0)", GoBool, expectBoolIIFE("(ctx.BarIndex == 0)")},
		{"bool field access", "ctx.IsDaily", GoBool, expectBoolIIFE("ctx.IsDaily")},
		{"bool series equality", "session_isfirstbarSeries.GetCurrent() == 1.0", GoBool,
			expectBoolIIFE("session_isfirstbarSeries.GetCurrent() == 1.0")},
		{"bool empty string", "", GoBool, expectBoolIIFE("")},

		/* string → NaN fallback */
		{"string literal", `"stock"`, GoString, expectStringNaN(`"stock"`)},
		{"string variable", "syminfo_tickerid", GoString, expectStringNaN("syminfo_tickerid")},
		{"string field access", "ctx.Timezone", GoString, expectStringNaN("ctx.Timezone")},
		{"string color literal", `"#FFFFFF"`, GoString, expectStringNaN(`"#FFFFFF"`)},
		{"string empty string", "", GoString, expectStringNaN("")},

		/* unknown GoValueType falls through to default (pass-through) */
		{"unknown type 99", "someExpr", GoValueType(99), expectExact("someExpr")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := coercer.Coerce(tt.code, tt.goType)
			tt.check(t, result)
		})
	}
}

func expectExact(want string) func(*testing.T, string) {
	return func(t *testing.T, got string) {
		t.Helper()
		if got != want {
			t.Errorf("expected exact %q, got %q", want, got)
		}
	}
}

func expectBoolIIFE(expr string) func(*testing.T, string) {
	return func(t *testing.T, got string) {
		t.Helper()
		if !strings.HasPrefix(got, "func() float64 {") {
			t.Errorf("expected IIFE prefix, got: %s", got)
		}
		if !strings.Contains(got, "if "+expr+" { return 1.0 } else { return 0.0 }") {
			t.Errorf("expected bool branch for %q, got: %s", expr, got)
		}
		if !strings.HasSuffix(got, "}()") {
			t.Errorf("expected IIFE invocation suffix, got: %s", got)
		}
	}
}

func expectStringNaN(expr string) func(*testing.T, string) {
	return func(t *testing.T, got string) {
		t.Helper()
		if !strings.HasPrefix(got, "math.NaN()") {
			t.Errorf("expected math.NaN() prefix, got: %s", got)
		}
		if !strings.Contains(got, expr) {
			t.Errorf("expected reference to %q, got: %s", expr, got)
		}
	}
}

func TestGoValueType_String(t *testing.T) {
	tests := []struct {
		vt       GoValueType
		expected string
	}{
		{GoFloat64, "float64"},
		{GoBool, "bool"},
		{GoString, "string"},
		{GoValueType(99), "float64"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			if tc.vt.String() != tc.expected {
				t.Errorf("GoValueType(%d).String() = %q, want %q", tc.vt, tc.vt.String(), tc.expected)
			}
		})
	}
}
