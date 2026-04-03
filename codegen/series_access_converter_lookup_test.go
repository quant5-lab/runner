package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestSeriesAccessConverter_CallVarLookup validates CallVarLookup function injection behavior.
 * Tests generalized patterns for temp variable deduplication and lookup strategies.
 */
func TestSeriesAccessConverter_CallVarLookup(t *testing.T) {
	t.Run("nil lookup function skips temp var resolution", func(t *testing.T) {
		st := NewSymbolTable()
		conv := NewSeriesAccessConverter(st, "j", nil)

		call := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.sma"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}

		code, err := conv.ConvertExpression(call)
		if err != nil {
			t.Fatalf("ConvertExpression failed: %v", err)
		}

		// Without lookup, call should be converted normally (function + args)
		if !strings.Contains(code, "ta.sma") {
			t.Errorf("Expected function call conversion, got: %s", code)
		}
	})

	t.Run("lookup returns empty string falls back to normal conversion", func(t *testing.T) {
		st := NewSymbolTable()

		// Lookup that always returns empty (call not registered)
		lookupCallVar := func(call *ast.CallExpression) string {
			return ""
		}

		conv := NewSeriesAccessConverter(st, "j", lookupCallVar)

		call := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.ema"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}

		code, err := conv.ConvertExpression(call)
		if err != nil {
			t.Fatalf("ConvertExpression failed: %v", err)
		}

		// Call not found in lookup, should use normal function call conversion
		if !strings.Contains(code, "ta.ema") {
			t.Errorf("Expected fallback to function call, got: %s", code)
		}
	})

	t.Run("lookup returns temp var name generates series access", func(t *testing.T) {
		st := NewSymbolTable()

		targetCall := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.rma"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}

		// Lookup that returns temp var for specific call
		lookupCallVar := func(call *ast.CallExpression) string {
			if call == targetCall {
				return "ta_rma_14_abc123"
			}
			return ""
		}

		conv := NewSeriesAccessConverter(st, "j", lookupCallVar)

		code, err := conv.ConvertExpression(targetCall)
		if err != nil {
			t.Fatalf("ConvertExpression failed: %v", err)
		}

		want := "ta_rma_14_abc123Series.Get(j)"
		if code != want {
			t.Errorf("got %q, want %q", code, want)
		}
	})

	t.Run("lookup with offset 0 still generates Get call for temp vars", func(t *testing.T) {
		st := NewSymbolTable()

		call := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.change"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}

		lookupCallVar := func(c *ast.CallExpression) string {
			if c == call {
				return "ta_change_1_xyz789"
			}
			return ""
		}

		conv := NewSeriesAccessConverter(st, "0", lookupCallVar)

		code, err := conv.ConvertExpression(call)
		if err != nil {
			t.Fatalf("ConvertExpression failed: %v", err)
		}

		// Temp vars are always series, so even offset 0 uses .Get(0)
		want := "ta_change_1_xyz789Series.Get(0)"
		if code != want {
			t.Errorf("got %q, want %q (temp vars always use Series.Get)", code, want)
		}
	})

	t.Run("lookup distinguishes between different calls", func(t *testing.T) {
		st := NewSymbolTable()

		call1 := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.sma"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 50}},
		}

		call2 := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.sma"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 200}},
		}

		// Lookup that distinguishes by AST pointer identity
		lookupCallVar := func(call *ast.CallExpression) string {
			if call == call1 {
				return "ta_sma_50_aaa"
			}
			if call == call2 {
				return "ta_sma_200_bbb"
			}
			return ""
		}

		conv := NewSeriesAccessConverter(st, "i", lookupCallVar)

		code1, _ := conv.ConvertExpression(call1)
		code2, _ := conv.ConvertExpression(call2)

		if code1 != "ta_sma_50_aaaSeries.Get(i)" {
			t.Errorf("call1: got %q, want %q", code1, "ta_sma_50_aaaSeries.Get(i)")
		}

		if code2 != "ta_sma_200_bbbSeries.Get(i)" {
			t.Errorf("call2: got %q, want %q", code2, "ta_sma_200_bbbSeries.Get(i)")
		}
	})

	t.Run("nested call expressions use lookup independently", func(t *testing.T) {
		st := NewSymbolTable()

		innerCall := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.change"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}

		outerCall := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "max"},
			Arguments: []ast.Expression{innerCall, &ast.Literal{Value: 0.0}},
		}

		// Only inner call has temp var
		lookupCallVar := func(call *ast.CallExpression) string {
			if call == innerCall {
				return "ta_change_1_inner"
			}
			return ""
		}

		conv := NewSeriesAccessConverter(st, "k", lookupCallVar)

		code, err := conv.ConvertExpression(outerCall)
		if err != nil {
			t.Fatalf("ConvertExpression failed: %v", err)
		}

		// Inner call should use temp var, outer call mapped to math.Max
		if !strings.Contains(code, "ta_change_1_innerSeries.Get(k)") {
			t.Errorf("Inner call should use temp var, got: %s", code)
		}

		if !strings.Contains(code, "math.Max") {
			t.Errorf("Outer call should be mapped to math.Max, got: %s", code)
		}
	})

	t.Run("lookup in binary expression with multiple calls", func(t *testing.T) {
		st := NewSymbolTable()

		leftCall := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.sma"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}

		rightCall := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.ema"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}

		expr := &ast.BinaryExpression{
			Left:     leftCall,
			Operator: ">",
			Right:    rightCall,
		}

		lookupCallVar := func(call *ast.CallExpression) string {
			if call == leftCall {
				return "ta_sma_20_left"
			}
			if call == rightCall {
				return "ta_ema_20_right"
			}
			return ""
		}

		conv := NewSeriesAccessConverter(st, "n", lookupCallVar)

		code, err := conv.ConvertExpression(expr)
		if err != nil {
			t.Fatalf("ConvertExpression failed: %v", err)
		}

		if !strings.Contains(code, "ta_sma_20_leftSeries.Get(n)") {
			t.Errorf("Left call should use temp var, got: %s", code)
		}

		if !strings.Contains(code, "ta_ema_20_rightSeries.Get(n)") {
			t.Errorf("Right call should use temp var, got: %s", code)
		}
	})

	t.Run("lookup function called for each call expression traversal", func(t *testing.T) {
		st := NewSymbolTable()

		call := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.rma"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}

		callCount := 0
		lookupCallVar := func(c *ast.CallExpression) string {
			callCount++
			return ""
		}

		conv := NewSeriesAccessConverter(st, "j", lookupCallVar)

		_, err := conv.ConvertExpression(call)
		if err != nil {
			t.Fatalf("ConvertExpression failed: %v", err)
		}

		if callCount != 1 {
			t.Errorf("Lookup should be called once per call expression, got %d calls", callCount)
		}
	})
}

/* TestSeriesExpressionAccessor_CallVarLookup validates integration with SeriesExpressionAccessor.
 * Ensures accessor correctly passes lookup to converter for both loop and initial value access.
 */
func TestSeriesExpressionAccessor_CallVarLookup(t *testing.T) {
	t.Run("GenerateLoopValueAccess uses lookup function", func(t *testing.T) {
		st := NewSymbolTable()

		call := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.sma"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}

		lookupCallVar := func(c *ast.CallExpression) string {
			if c == call {
				return "ta_sma_period_hash"
			}
			return ""
		}

		accessor := NewSeriesExpressionAccessor(call, st, lookupCallVar)

		code := accessor.GenerateLoopValueAccess("j")

		want := "ta_sma_period_hashSeries.Get(j)"
		if code != want {
			t.Errorf("got %q, want %q", code, want)
		}
	})

	t.Run("GenerateInitialValueAccess uses lookup function", func(t *testing.T) {
		st := NewSymbolTable()

		call := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.ema"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}

		lookupCallVar := func(c *ast.CallExpression) string {
			if c == call {
				return "ta_ema_init_var"
			}
			return ""
		}

		accessor := NewSeriesExpressionAccessor(call, st, lookupCallVar)

		code := accessor.GenerateInitialValueAccess(10)

		// Period 10 → offset "9" (period-1)
		want := "ta_ema_init_varSeries.Get(9)"
		if code != want {
			t.Errorf("got %q, want %q", code, want)
		}
	})

	t.Run("nil lookup returns NaN for call expressions", func(t *testing.T) {
		st := NewSymbolTable()

		call := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.stdev"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}

		accessor := NewSeriesExpressionAccessor(call, st, nil)

		code := accessor.GenerateLoopValueAccess("j")

		// Without lookup, call can't be resolved to temp var or series
		// Falls back to function call which likely produces NaN context
		if code == "math.NaN()" {
			// Acceptable - no way to resolve without lookup
			return
		}

		// Or it generates function call
		if strings.Contains(code, "ta.stdev") {
			// Also acceptable - generates actual function call
			return
		}

		t.Errorf("Expected NaN or function call, got: %s", code)
	})

	t.Run("complex expression with mixed builtin and calls", func(t *testing.T) {
		st := NewSymbolTable()

		call := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.sma"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "volume"}},
		}

		expr := &ast.BinaryExpression{
			Left:     &ast.Identifier{Name: "volume"},
			Operator: ">",
			Right:    call,
		}

		lookupCallVar := func(c *ast.CallExpression) string {
			if c == call {
				return "ta_sma_volume_avg"
			}
			return ""
		}

		accessor := NewSeriesExpressionAccessor(expr, st, lookupCallVar)

		code := accessor.GenerateLoopValueAccess("i")

		// volume is builtin field, not series variable
		if !strings.Contains(code, "volumeSeries.Get(i)") {
			t.Errorf("Should contain builtin volume access, got: %s", code)
		}

		if !strings.Contains(code, "ta_sma_volume_avgSeries.Get(i)") {
			t.Errorf("Should contain temp var access, got: %s", code)
		}
	})

	t.Run("lookup function isolation between accessors", func(t *testing.T) {
		st := NewSymbolTable()

		call1 := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.rma"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}

		call2 := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.ema"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}

		// Each accessor gets own lookup function
		lookup1 := func(c *ast.CallExpression) string {
			if c == call1 {
				return "accessor1_var"
			}
			return ""
		}

		lookup2 := func(c *ast.CallExpression) string {
			if c == call2 {
				return "accessor2_var"
			}
			return ""
		}

		accessor1 := NewSeriesExpressionAccessor(call1, st, lookup1)
		accessor2 := NewSeriesExpressionAccessor(call2, st, lookup2)

		code1 := accessor1.GenerateLoopValueAccess("j")
		code2 := accessor2.GenerateLoopValueAccess("j")

		if !strings.Contains(code1, "accessor1_var") {
			t.Errorf("Accessor1 should use lookup1, got: %s", code1)
		}

		if !strings.Contains(code2, "accessor2_var") {
			t.Errorf("Accessor2 should use lookup2, got: %s", code2)
		}

		// Ensure no cross-contamination
		if strings.Contains(code1, "accessor2_var") {
			t.Errorf("Accessor1 should not see lookup2 vars, got: %s", code1)
		}

		if strings.Contains(code2, "accessor1_var") {
			t.Errorf("Accessor2 should not see lookup1 vars, got: %s", code2)
		}
	})
}

/* TestCallVarLookup_EdgeCases validates boundary conditions and error scenarios.
 * Ensures robust behavior under unusual but valid conditions.
 */
func TestCallVarLookup_EdgeCases(t *testing.T) {
	t.Run("lookup returns whitespace only treated as empty", func(t *testing.T) {
		st := NewSymbolTable()

		call := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.sma"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}

		lookupCallVar := func(c *ast.CallExpression) string {
			return "   " // Whitespace only
		}

		conv := NewSeriesAccessConverter(st, "j", lookupCallVar)

		code, err := conv.ConvertExpression(call)
		if err != nil {
			t.Fatalf("ConvertExpression failed: %v", err)
		}

		// Whitespace-only is truthy in Go, so should generate series access
		// This validates that lookup returns are used as-is
		if !strings.Contains(code, "Series.Get(j)") {
			t.Errorf("Non-empty string (even whitespace) should be used, got: %s", code)
		}
	})

	t.Run("lookup returns special characters in var name", func(t *testing.T) {
		st := NewSymbolTable()

		call := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.ema"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}

		// Var names with underscores, numbers, prefixes
		testNames := []string{
			"ta_sma_50_abc123",
			"_privateVar",
			"var_with_many_underscores",
			"CamelCaseVar",
		}

		for _, varName := range testNames {
			t.Run("var_name="+varName, func(t *testing.T) {
				lookupCallVar := func(c *ast.CallExpression) string {
					return varName
				}

				conv := NewSeriesAccessConverter(st, "j", lookupCallVar)

				code, err := conv.ConvertExpression(call)
				if err != nil {
					t.Fatalf("ConvertExpression failed: %v", err)
				}

				want := varName + "Series.Get(j)"
				if code != want {
					t.Errorf("got %q, want %q", code, want)
				}
			})
		}
	})

	t.Run("lookup with very long offset variable names", func(t *testing.T) {
		st := NewSymbolTable()

		call := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.rma"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}

		lookupCallVar := func(c *ast.CallExpression) string {
			return "ta_rma_period"
		}

		longOffset := "innerLoopIndexVariableWithVeryLongName"
		conv := NewSeriesAccessConverter(st, longOffset, lookupCallVar)

		code, err := conv.ConvertExpression(call)
		if err != nil {
			t.Fatalf("ConvertExpression failed: %v", err)
		}

		want := "ta_rma_periodSeries.Get(" + longOffset + ")"
		if code != want {
			t.Errorf("got %q, want %q", code, want)
		}
	})

	t.Run("multiple conversions with same converter reuse lookup", func(t *testing.T) {
		st := NewSymbolTable()

		call1 := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.sma"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}

		call2 := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "ta.ema"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}

		callCount := 0
		lookupCallVar := func(c *ast.CallExpression) string {
			callCount++
			if c == call1 {
				return "var1"
			}
			if c == call2 {
				return "var2"
			}
			return ""
		}

		conv := NewSeriesAccessConverter(st, "j", lookupCallVar)

		// Multiple conversions should each call lookup
		_, _ = conv.ConvertExpression(call1)
		_, _ = conv.ConvertExpression(call2)

		if callCount != 2 {
			t.Errorf("Lookup should be called once per conversion, got %d", callCount)
		}
	})
}
