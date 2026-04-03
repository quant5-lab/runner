package codegen

import (
	"fmt"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestBuiltinTrueRangeAccessor_CodeGeneration covers all three generate methods:
// the IIFE structure, the three-component true-range formula, and the bar-0 NaN
// guard required by Pine's ta.tr default (handle_na=false, no previous close).
func TestBuiltinTrueRangeAccessor_CodeGeneration(t *testing.T) {
	accessor := NewBuiltinTrueRangeAccessor()

	t.Run("GenerateLoopValueAccess", func(t *testing.T) {
		for _, loopVar := range []string{"j", "i", "idx", "0"} {
			result := accessor.GenerateLoopValueAccess(loopVar)

			requiredParts := []string{
				"func() float64",
				"barIdx := ctx.BarIndex-" + loopVar,
				"if barIdx < 1 { return math.NaN() }",
				"prevClose := ctx.Data[barIdx-1].Close",
				"currentBar := ctx.Data[barIdx]",
				"math.Max(currentBar.High - currentBar.Low",
				"math.Abs(currentBar.High - prevClose)",
				"math.Abs(currentBar.Low - prevClose)",
			}
			for _, part := range requiredParts {
				if !contains(result, part) {
					t.Errorf("loopVar=%q missing %q\nGot: %s", loopVar, part, result)
				}
			}
		}
	})

	t.Run("GenerateInitialValueAccess embeds period-1 offset", func(t *testing.T) {
		for _, period := range []int{1, 5, 14, 20} {
			result := accessor.GenerateInitialValueAccess(period)
			offset := fmt.Sprintf("%d", period-1)

			if !contains(result, "ctx.BarIndex-"+offset) {
				t.Errorf("period=%d: expected offset ctx.BarIndex-%s\nGot: %s", period, offset, result)
			}
			if !contains(result, "math.Max(currentBar.High - currentBar.Low") {
				t.Errorf("period=%d: missing true-range formula\nGot: %s", period, result)
			}
			if !contains(result, "if barIdx < 1 { return math.NaN() }") {
				t.Errorf("period=%d: missing bar-0 NaN guard\nGot: %s", period, result)
			}
		}
	})

	t.Run("GenerateCurrentValueAccess", func(t *testing.T) {
		result := accessor.GenerateCurrentValueAccess()

		requiredParts := []string{
			"func() float64",
			"if ctx.BarIndex < 1 { return math.NaN() }",
			"prevClose := ctx.Data[ctx.BarIndex-1].Close",
			"currentBar := ctx.Data[ctx.BarIndex]",
			"math.Max(currentBar.High - currentBar.Low",
			"math.Abs(currentBar.High - prevClose)",
			"math.Abs(currentBar.Low - prevClose)",
		}
		for _, part := range requiredParts {
			if !contains(result, part) {
				t.Errorf("missing %q\nGot: %s", part, result)
			}
		}
	})
}

func TestBuiltinTrueRangeAccessor_InterfaceContracts(t *testing.T) {
	accessor := NewBuiltinTrueRangeAccessor()

	if got := accessor.GetBaseOffset(); got != 0 {
		t.Errorf("GetBaseOffset() = %d, want 0", got)
	}
	if got := accessor.GetPreamble(); got != "" {
		t.Errorf("GetPreamble() = %q, want empty string", got)
	}
}

// TestBuiltinTrueRange_NeverGeneratesSeriesAccess enforces that tr always resolves
// to direct OHLCV indexing — storing tr in a Series would break historical subscript
// semantics and conflict with the AccessGenerator contract.
func TestBuiltinTrueRange_NeverGeneratesSeriesAccess(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()
	accessor := NewBuiltinTrueRangeAccessor()

	cases := []struct {
		name   string
		result func() string
	}{
		{"current bar", func() string { return handler.GenerateCurrentBarAccess("tr") }},
		{"security context", func() string { return handler.GenerateSecurityContextAccess("tr") }},
		{"historical offset 1", func() string { return handler.GenerateHistoricalAccess("tr", 1) }},
		{"loop value access", func() string { return accessor.GenerateLoopValueAccess("j") }},
		{"initial value access", func() string { return accessor.GenerateInitialValueAccess(14) }},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := c.result()
			if contains(result, "trSeries.Get(") {
				t.Errorf("generated forbidden trSeries access\nGot: %s", result)
			}
			if !contains(result, "math.Max") {
				t.Errorf("missing inline calculation (math.Max)\nGot: %s", result)
			}
		})
	}
}

func TestBuiltinTrueRange_ConsistencyAcrossContexts(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	cases := []struct {
		name            string
		result          func() string
		extraComponents []string
	}{
		{
			name:            "current bar",
			result:          func() string { return handler.GenerateCurrentBarAccess("tr") },
			extraComponents: []string{"bar.High", "bar.Low"},
		},
		{
			name:   "security context",
			result: func() string { return handler.GenerateSecurityContextAccess("tr") },
			extraComponents: []string{
				"highSeries.GetCurrent()",
				"lowSeries.GetCurrent()",
			},
		},
		{
			name:   "historical",
			result: func() string { return handler.GenerateHistoricalAccess("tr", 1) },
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := c.result()
			if !contains(result, "math.Max") {
				t.Errorf("%s: missing inline calculation (math.Max)\nGot: %s", c.name, result)
			}
			for _, comp := range c.extraComponents {
				if !contains(result, comp) {
					t.Errorf("%s: missing %q\nGot: %s", c.name, comp, result)
				}
			}
		})
	}
}

// TestArrowFunctionTACallGenerator_TrNotConfusedWithVariable guards against the
// accessor factory conflating the tr builtin with same-prefix user variables — the
// two resolve to different accessor types with incompatible code-generation paths.
func TestArrowFunctionTACallGenerator_TrNotConfusedWithVariable(t *testing.T) {
	gen := newTestGenerator()
	gen.variables = map[string]string{"my_tr": "float"}
	taGen := newTestArrowTAGenerator(gen)

	t.Run("bare tr returns BuiltinTrueRangeAccessor", func(t *testing.T) {
		accessor, err := taGen.accessorFactory.CreateAccessorForExpression(&ast.Identifier{Name: "tr"})
		if err != nil {
			t.Fatalf("CreateAccessorForExpression(tr) error: %v", err)
		}
		if _, ok := accessor.(*TrueRangeAccessGenerator); !ok {
			t.Errorf("expected TrueRangeAccessGenerator, got %T", accessor)
		}
	})

	t.Run("user variable my_tr returns ArrowFunctionParameterAccessor", func(t *testing.T) {
		accessor, err := taGen.accessorFactory.CreateAccessorForExpression(&ast.Identifier{Name: "my_tr"})
		if err != nil {
			t.Fatalf("CreateAccessorForExpression(my_tr) error: %v", err)
		}
		if _, ok := accessor.(*ArrowFunctionParameterAccessor); !ok {
			t.Errorf("expected ArrowFunctionParameterAccessor, got %T", accessor)
		}
	})
}

func TestBuiltinTrueRange_InTACallContexts(t *testing.T) {
	gen := newTestGenerator()
	taGen := newTestArrowTAGenerator(gen)

	accessor, err := taGen.accessorFactory.CreateAccessorForExpression(&ast.Identifier{Name: "tr"})
	if err != nil {
		t.Fatalf("CreateAccessorForExpression(tr) error: %v", err)
	}

	if _, ok := accessor.(*TrueRangeAccessGenerator); !ok {
		t.Fatalf("expected TrueRangeAccessGenerator, got %T", accessor)
	}

	t.Run("loop access is inline with no Series.Get", func(t *testing.T) {
		code := accessor.GenerateLoopValueAccess("j")
		for _, required := range []string{"func() float64", "idx := ctx.BarIndex - j", "math.Max", "pc"} {
			if !contains(code, required) {
				t.Errorf("missing %q in loop access\nGot: %s", required, code)
			}
		}
		if contains(code, "Series.Get(") || contains(code, "trSeries") {
			t.Errorf("loop access must not reference a stored series\nGot: %s", code)
		}
	})

	t.Run("initial value access is inline", func(t *testing.T) {
		code := accessor.GenerateInitialValueAccess(20)
		if !contains(code, "func() float64") {
			t.Errorf("initial value access missing IIFE wrapper\nGot: %s", code)
		}
		if !contains(code, "math.Max") {
			t.Errorf("initial value access missing true-range formula\nGot: %s", code)
		}
	})

	t.Run("current value access is inline", func(t *testing.T) {
		code := accessor.GenerateCurrentValueAccess()
		if !contains(code, "func() float64") {
			t.Errorf("current value access missing IIFE wrapper\nGot: %s", code)
		}
		if !contains(code, "math.Max") {
			t.Errorf("current value access missing true-range formula\nGot: %s", code)
		}
	})
}

// TestTrueRangeAccessGenerator_HandleNASemantics covers TrueRangeAccessGenerator,
// which mirrors Pine's ta.tr(true) (handle_na=true): bar 0 returns high−low because
// there is no previous close to exclude, so the range is unambiguous.
func TestTrueRangeAccessGenerator_HandleNASemantics(t *testing.T) {
	g := NewTrueRangeAccessGenerator()

	t.Run("bar 0 returns high minus low, not NaN", func(t *testing.T) {
		code := g.GenerateCurrentValueAccess()
		if contains(code, "return math.NaN()") {
			t.Errorf("TrueRangeAccessGenerator must not return NaN on bar 0\nGot: %s", code)
		}
		if !contains(code, "return h - l") {
			t.Errorf("TrueRangeAccessGenerator must return h-l on bar 0\nGot: %s", code)
		}
		if !contains(code, "if idx == 0") {
			t.Errorf("missing bar-0 branch 'if idx == 0'\nGot: %s", code)
		}
	})

	t.Run("normal bars apply full three-component formula", func(t *testing.T) {
		code := g.GenerateLoopValueAccess("j")
		for _, part := range []string{"math.Max(h-l", "math.Abs(h-pc)", "math.Abs(l-pc)"} {
			if !contains(code, part) {
				t.Errorf("missing formula component %q\nGot: %s", part, code)
			}
		}
	})

	t.Run("GetBaseOffset is 0", func(t *testing.T) {
		if got := g.GetBaseOffset(); got != 0 {
			t.Errorf("GetBaseOffset() = %d, want 0", got)
		}
	})
}

// TestTrueRangeAccessors_Bar0SemanticContrast enforces the semantic split between the
// two TR accessor types at bar 0, which must be preserved to match Pine's behavior:
//
//   - BuiltinTrueRangeAccessor (user-facing ta.tr, handle_na=false) → NaN on bar 0
//   - TrueRangeAccessGenerator (internal ta.atr seed, handle_na=true) → high−low on bar 0
func TestTrueRangeAccessors_Bar0SemanticContrast(t *testing.T) {
	handleNAFalse := NewBuiltinTrueRangeAccessor()
	handleNATrue := NewTrueRangeAccessGenerator()

	t.Run("handle_na=false bar 0 is NaN", func(t *testing.T) {
		code := handleNAFalse.GenerateCurrentValueAccess()
		if !contains(code, "return math.NaN()") {
			t.Errorf("BuiltinTrueRangeAccessor bar 0 must return math.NaN()\nGot: %s", code)
		}
	})

	t.Run("handle_na=true bar 0 is high minus low", func(t *testing.T) {
		code := handleNATrue.GenerateCurrentValueAccess()
		if contains(code, "return math.NaN()") {
			t.Errorf("TrueRangeAccessGenerator must not return NaN on bar 0\nGot: %s", code)
		}
		if !contains(code, "return h - l") {
			t.Errorf("TrueRangeAccessGenerator must return h-l on bar 0\nGot: %s", code)
		}
	})

	t.Run("normal bars share identical three-component formula structure", func(t *testing.T) {
		falseCode := handleNAFalse.GenerateLoopValueAccess("j")
		trueCode := handleNATrue.GenerateLoopValueAccess("j")

		for _, code := range []string{falseCode, trueCode} {
			if !contains(code, "math.Max") {
				t.Errorf("normal bar formula missing math.Max\nGot: %s", code)
			}
			if !contains(code, "math.Abs") {
				t.Errorf("normal bar formula missing math.Abs\nGot: %s", code)
			}
		}
	})
}
