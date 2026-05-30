package codegen

import (
	"strings"
	"testing"
)

// TestArrowTupleSecurityGenerator_ConcreteBarMapperAssertion verifies that the
// generated code obtains a *request.SecurityBarMapper via ConcreteBarMappers, not
// through SecurityBarMappers, which exposes only the BarIndexMapper interface and
// lacks the concrete methods needed for lookahead and parent linkage.
func TestArrowTupleSecurityGenerator_ConcreteBarMapperAssertion(t *testing.T) {
	const source = `//@version=5
strategy("t")
getRange(sym) =>
    [h, l] = request.security(sym, "D", [high, low])
    h - l
r = getRange(syminfo.tickerid)
plot(r)
`
	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilePineScript: %v", err)
	}

	// Must use ConcreteBarMappers for the type assertion.
	if !strings.Contains(code, `arrowCtx.ConcreteBarMappers[secKey].(*request.SecurityBarMapper)`) {
		t.Errorf("expected concrete bar mapper type assertion in:\n%s", code)
	}
	// Must NOT use SecurityBarMappers as the concrete access point.
	if strings.Contains(code, `arrowCtx.SecurityBarMappers[secKey].(*request`) {
		t.Errorf("must not type-assert from SecurityBarMappers (interface); must use ConcreteBarMappers:\n%s", code)
	}
	// Guard check uses SecurityBarMappers (interface lookup only) — that is correct.
	if !strings.Contains(code, `arrowCtx.SecurityBarMappers[secKey]`) {
		t.Errorf("guard check should still read from SecurityBarMappers:\n%s", code)
	}
}

// TestArrowTupleSecurityGenerator_GoLocalsAfterBlock verifies that after the
// security block closes, each tuple variable is extracted as a Go local via
//
//	name := nameSeries.GetCurrent()
//
// Downstream statements in the arrow body can then reference the tuple names
// directly without a redundant series read.
func TestArrowTupleSecurityGenerator_GoLocalsAfterBlock(t *testing.T) {
	tests := []struct {
		name   string
		source string
		locals []string // names that must appear as Go locals after the block
	}{
		{
			name: "two OHLCV fields",
			source: `//@version=5
strategy("t")
f(sym) =>
    [h, l] = request.security(sym, "D", [high, low])
    h - l
plot(f(syminfo.tickerid))
`,
			locals: []string{"h", "l"},
		},
		{
			name: "three OHLCV fields",
			source: `//@version=5
strategy("t")
f(sym) =>
    [o, h, l] = request.security(sym, "D", [open, high, low])
    (h + l) / 2 - o
plot(f(syminfo.tickerid))
`,
			locals: []string{"o", "h", "l"},
		},
		{
			name: "single variable",
			source: `//@version=5
strategy("t")
f(sym) =>
    [c] = request.security(sym, "D", [close])
    c
plot(f(syminfo.tickerid))
`,
			locals: []string{"c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.source)
			if err != nil {
				t.Fatalf("compilePineScript: %v", err)
			}

			for _, varName := range tt.locals {
				want := varName + " := " + varName + "Series.GetCurrent()"
				if !strings.Contains(code, want) {
					t.Errorf("expected Go local %q in:\n%s", want, code)
				}
			}
		})
	}
}

// TestArrowTupleSecurityGenerator_NaNFallback verifies that both guard paths
// (security context not found, bar mapper not found) emit NaN set calls for
// every variable in the tuple — none should be left undefined.
func TestArrowTupleSecurityGenerator_NaNFallback(t *testing.T) {
	const source = `//@version=5
strategy("t")
f(sym) =>
    [h, l] = request.security(sym, "D", [high, low])
    h - l
plot(f(syminfo.tickerid))
`
	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilePineScript: %v", err)
	}

	// Each fallback branch must set NaN for both h and l.
	// Count occurrences: context-not-found and mapper-not-found each add one set.
	hNaN := strings.Count(code, "hSeries.Set(math.NaN())")
	lNaN := strings.Count(code, "lSeries.Set(math.NaN())")

	if hNaN < 2 {
		t.Errorf("expected at least 2 NaN fallbacks for hSeries, got %d in:\n%s", hNaN, code)
	}
	if lNaN < 2 {
		t.Errorf("expected at least 2 NaN fallbacks for lSeries, got %d in:\n%s", lNaN, code)
	}
}

// TestArrowTupleSecurityGenerator_OHLCVFieldRouting verifies that OHLCV
// identifiers in the tuple expression list resolve to direct secCtx.Data[idx]
// field access rather than the streaming evaluator path.
func TestArrowTupleSecurityGenerator_OHLCVFieldRouting(t *testing.T) {
	tests := []struct {
		field   string
		goField string
	}{
		{"high", "High"},
		{"low", "Low"},
		{"open", "Open"},
		{"close", "Close"},
		{"volume", "Volume"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			source := `//@version=5
strategy("t")
f(sym) =>
    [v] = request.security(sym, "D", [` + tt.field + `])
    v
plot(f(syminfo.tickerid))
`
			code, err := compilePineScript(source)
			if err != nil {
				t.Fatalf("compilePineScript: %v", err)
			}

			want := "secCtx.Data[secBarIdx]." + tt.goField
			if !strings.Contains(code, want) {
				t.Errorf("expected direct OHLCV field access %q for %q in:\n%s", want, tt.field, code)
			}
			// Should NOT go through the streaming evaluator for plain OHLCV.
			if strings.Contains(code, "GetOrCreateSecurityEvaluators") {
				t.Errorf("OHLCV field %q must not use streaming evaluator path:\n%s", tt.field, code)
			}
		})
	}
}

// TestArrowTupleSecurityGenerator_ArrowContextIsolation verifies that the generated
// code reads from arrowCtx rather than the top-level securityContexts/securityBarMappers
// maps, which only exist in the main bar loop scope and would be undefined inside
// an arrow function body.
func TestArrowTupleSecurityGenerator_ArrowContextIsolation(t *testing.T) {
	const source = `//@version=5
strategy("t")
f(sym) =>
    [h, l] = request.security(sym, "D", [high, low])
    h - l
plot(f(syminfo.tickerid))
`
	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilePineScript: %v", err)
	}

	// Arrow body must access security contexts through arrowCtx.
	if !strings.Contains(code, "arrowCtx.SecurityContexts[secKey]") {
		t.Errorf("expected arrowCtx.SecurityContexts lookup in arrow body:\n%s", code)
	}
	if !strings.Contains(code, "arrowCtx.SecurityBarMappers[secKey]") {
		t.Errorf("expected arrowCtx.SecurityBarMappers lookup in arrow body:\n%s", code)
	}
	if !strings.Contains(code, "arrowCtx.ConcreteBarMappers[secKey]") {
		t.Errorf("expected arrowCtx.ConcreteBarMappers lookup in arrow body:\n%s", code)
	}

	// Must NOT reference the top-level scope maps without the arrowCtx prefix.
	// Trim arrowCtx references so we can check for bare identifiers.
	stripped := strings.ReplaceAll(code, "arrowCtx.SecurityContexts", "__REPLACED__")
	stripped = strings.ReplaceAll(stripped, "arrowCtx.SecurityBarMappers", "__REPLACED__")
	stripped = strings.ReplaceAll(stripped, "arrowCtx.ConcreteBarMappers", "__REPLACED__")
	if strings.Contains(stripped, "securityContexts[") {
		t.Errorf("arrow body must not reference bare securityContexts[] (top-level scope):\n%s", code)
	}
	if strings.Contains(stripped, "securityBarMappers[") {
		t.Errorf("arrow body must not reference bare securityBarMappers[] (top-level scope):\n%s", code)
	}
}

// TestArrowTupleSecurityGenerator_LookaheadSameTimeframe verifies that the
// generated guard enables lookahead when the security timeframe equals the chart
// timeframe — this mirrors the Pine semantics for non-repainting security calls.
func TestArrowTupleSecurityGenerator_LookaheadSameTimeframe(t *testing.T) {
	const source = `//@version=5
strategy("t")
f(sym) =>
    [h, l] = request.security(sym, "D", [high, low])
    h - l
plot(f(syminfo.tickerid))
`
	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilePineScript: %v", err)
	}

	// The lookahead guard must compare the requested timeframe to ctx.Timeframe.
	if !strings.Contains(code, `"1D" == ctx.Timeframe`) && !strings.Contains(code, `ctx.Timeframe == "1D"`) {
		t.Errorf("expected lookahead timeframe comparison in:\n%s", code)
	}
	if !strings.Contains(code, "secLookahead = true") {
		t.Errorf("expected secLookahead = true branch in:\n%s", code)
	}
}

// TestArrowTupleSecurityGenerator_SeriesAllocatedForEachVar confirms that each
// tuple variable gets its own arrow-local series — required for per-bar
// ForwardSeriesBuffer semantics inside the arrow context.
func TestArrowTupleSecurityGenerator_SeriesAllocatedForEachVar(t *testing.T) {
	const source = `//@version=5
strategy("t")
f(sym) =>
    [a, b, c] = request.security(sym, "D", [open, high, low])
    a + b + c
plot(f(syminfo.tickerid))
`
	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilePineScript: %v", err)
	}

	for _, name := range []string{"a", "b", "c"} {
		want := name + "Series := arrowCtx.GetOrCreateSeries(" + `"` + name + `"`
		if !strings.Contains(code, want) {
			t.Errorf("expected series allocation %q in:\n%s", want, code)
		}
	}
}
