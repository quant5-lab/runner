package codegen

import (
	"strings"
	"testing"
)

// TestNestedChildContextAllocator_FirstAllocation validates that the first allocation
// for any function name yields key "<funcName>_1".
func TestNestedChildContextAllocator_FirstAllocation(t *testing.T) {
	a := NewNestedChildContextAllocator()
	expr := a.AllocateChildContextExpr("dirmov")

	if !strings.Contains(expr, `"dirmov_1"`) {
		t.Errorf("First allocation must generate key \"dirmov_1\", got: %s", expr)
	}
	if !strings.Contains(expr, "GetOrCreateChildContext") {
		t.Errorf("Must emit GetOrCreateChildContext call, got: %s", expr)
	}
}

// TestNestedChildContextAllocator_CounterIncrements validates that repeated allocations
// for the same function name produce sequentially numbered keys.
func TestNestedChildContextAllocator_CounterIncrements(t *testing.T) {
	a := NewNestedChildContextAllocator()

	for i, wantKey := range []string{`"foo_1"`, `"foo_2"`, `"foo_3"`} {
		expr := a.AllocateChildContextExpr("foo")
		if !strings.Contains(expr, wantKey) {
			t.Errorf("Call %d: expected key %s in expression, got: %s", i+1, wantKey, expr)
		}
	}
}

// TestNestedChildContextAllocator_IndependentCounters validates that counters for
// different function names are independent — allocating for "a" does not affect "b".
func TestNestedChildContextAllocator_IndependentCounters(t *testing.T) {
	a := NewNestedChildContextAllocator()

	a.AllocateChildContextExpr("dirmov")
	a.AllocateChildContextExpr("dirmov")
	exprFoo := a.AllocateChildContextExpr("foo")

	if !strings.Contains(exprFoo, `"foo_1"`) {
		t.Errorf("Independent counter: foo must start at 1 regardless of dirmov calls, got: %s", exprFoo)
	}
}

// TestNestedChildContextAllocator_ResetClearsAllCounters validates that Reset() returns
// all counters to zero so the next allocation starts at _1 again.
func TestNestedChildContextAllocator_ResetClearsAllCounters(t *testing.T) {
	a := NewNestedChildContextAllocator()
	a.AllocateChildContextExpr("dirmov")
	a.AllocateChildContextExpr("dirmov")

	a.Reset()

	expr := a.AllocateChildContextExpr("dirmov")
	if !strings.Contains(expr, `"dirmov_1"`) {
		t.Errorf("After Reset, first allocation must restart at _1, got: %s", expr)
	}
}

// TestNestedChildContextAllocator_ResetIndependence validates that Reset() clears counters
// for all previously seen function names, not just the most recently used one.
func TestNestedChildContextAllocator_ResetIndependence(t *testing.T) {
	a := NewNestedChildContextAllocator()
	a.AllocateChildContextExpr("alpha")
	a.AllocateChildContextExpr("alpha")
	a.AllocateChildContextExpr("beta")

	a.Reset()

	for _, funcName := range []string{"alpha", "beta"} {
		expr := a.AllocateChildContextExpr(funcName)
		if !strings.Contains(expr, `"`+funcName+`_1"`) {
			t.Errorf("After Reset, %s must restart at _1, got: %s", funcName, expr)
		}
	}
}

// TestNestedChildContextAllocator_EmittedExpressionStructure validates the shape of the
// generated Go expression: it must be a complete GetOrCreateChildContext call with a
// quoted string key argument.
func TestNestedChildContextAllocator_EmittedExpressionStructure(t *testing.T) {
	tests := []struct {
		funcName   string
		callNumber int
		wantKey    string
	}{
		{"dirmov", 1, `"dirmov_1"`},
		{"supertrend", 1, `"supertrend_1"`},
		{"hull", 2, `"hull_2"`},
		{"ta_pivot_point_levels", 3, `"ta_pivot_point_levels_3"`},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			a := NewNestedChildContextAllocator()
			var expr string
			for i := 0; i < tt.callNumber; i++ {
				expr = a.AllocateChildContextExpr(tt.funcName)
			}
			if !strings.Contains(expr, tt.wantKey) {
				t.Errorf("Expected key %s in expression, got: %s", tt.wantKey, expr)
			}
			if !strings.HasPrefix(expr, "arrowCtx.GetOrCreateChildContext(") {
				t.Errorf("Expression must begin with arrowCtx.GetOrCreateChildContext(, got: %s", expr)
			}
		})
	}
}

// TestNestedChildContextAllocator_MultipleResets validates that the allocator can be
// Reset() multiple times and always restarts cleanly.
func TestNestedChildContextAllocator_MultipleResets(t *testing.T) {
	a := NewNestedChildContextAllocator()

	for round := 0; round < 3; round++ {
		a.Reset()
		expr := a.AllocateChildContextExpr("fn")
		if !strings.Contains(expr, `"fn_1"`) {
			t.Errorf("Round %d: after Reset, first allocation must be fn_1, got: %s", round, expr)
		}
		// Allocate more so the next Reset has something non-trivial to clear.
		a.AllocateChildContextExpr("fn")
		a.AllocateChildContextExpr("other")
	}
}

// TestNestedUDFInArrowBody_UsesChildContext validates that a tuple-returning UDF called
// inside another arrow function body emits GetOrCreateChildContext (persistent child
// context, reused across bars) rather than context.NewArrowContext (fresh context that
// destroys all internal series state on every bar).
//
// The GetOrCreateChildContext path is taken only for tuple-returning UDFs because those
// require a dedicated context variable at the call site.
func TestNestedUDFInArrowBody_UsesChildContext(t *testing.T) {
	src := `//@version=5
indicator("test")
inner() =>
    a = ta.ema(close, 5)
    b = ta.rma(close, 10)
    [a, b]
outer() =>
    [x, y] = inner()
    x + y
result = outer()
plot(result)
`
	code, err := compilePineScript(src)
	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	if !strings.Contains(code, "GetOrCreateChildContext") {
		t.Error("Tuple-returning UDF call inside arrow body must emit GetOrCreateChildContext for persistent child context")
	}
}

// TestTopLevelUDFCall_UsesFreshContext validates that a tuple-returning UDF called at
// the top level (outside any arrow body) emits context.NewArrowContext rather than
// GetOrCreateChildContext — each top-level call site is independent.
func TestTopLevelUDFCall_UsesFreshContext(t *testing.T) {
	src := `//@version=5
indicator("test")
inner() =>
    a = ta.ema(close, 5)
    b = ta.rma(close, 10)
    [a, b]
[x, y] = inner()
result = x + y
plot(result)
`
	code, err := compilePineScript(src)
	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	if !strings.Contains(code, "context.NewArrowContext") {
		t.Error("Top-level UDF call must emit context.NewArrowContext")
	}
}

// TestNestedUDFInArrowBody_ChildContextKeyStability validates that two calls to the same
// tuple-returning inner function within one arrow body get distinct, stable keys (_1 and _2).
func TestNestedUDFInArrowBody_ChildContextKeyStability(t *testing.T) {
	src := `//@version=5
indicator("test")
inner() =>
    a = ta.ema(close, 5)
    b = ta.rma(close, 10)
    [a, b]
outer() =>
    [a1, b1] = inner()
    [a2, b2] = inner()
    a1 + b1 + a2 + b2
result = outer()
plot(result)
`
	code, err := compilePineScript(src)
	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	count := strings.Count(code, "GetOrCreateChildContext")
	if count < 2 {
		t.Errorf("Two inner() tuple calls must emit two distinct GetOrCreateChildContext keys, found %d", count)
	}
}
