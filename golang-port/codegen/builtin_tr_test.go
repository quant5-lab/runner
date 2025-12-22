package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestBuiltinTrueRangeAccessor_GenerateLoopValueAccess(t *testing.T) {
	accessor := NewBuiltinTrueRangeAccessor()

	tests := []struct {
		name    string
		loopVar string
	}{
		{"loop with j", "j"},
		{"loop with i", "i"},
		{"loop with idx", "idx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := accessor.GenerateLoopValueAccess(tt.loopVar)

			/* Verify inline IIFE with loop offset */
			expectedComponents := []string{
				"func() float64",
				"barIdx := ctx.BarIndex-" + tt.loopVar,
				"math.Max",
				"High", "Low", "Close",
			}

			for _, component := range expectedComponents {
				if !contains(result, component) {
					t.Errorf("GenerateLoopValueAccess(%s) missing expected component: %s\nGot: %s", tt.loopVar, component, result)
				}
			}

			/* Verify first bar edge case handling */
			if !contains(result, "if barIdx < 1") {
				t.Errorf("GenerateLoopValueAccess(%s) missing first bar check\nGot: %s", tt.loopVar, result)
			}
		})
	}
}

func TestBuiltinTrueRangeAccessor_GenerateInitialValueAccess(t *testing.T) {
	accessor := NewBuiltinTrueRangeAccessor()

	tests := []struct {
		name   string
		period int
	}{
		{"period 14", 14},
		{"period 20", 20},
		{"period 1", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := accessor.GenerateInitialValueAccess(tt.period)

			/* Verify inline IIFE with period offset */
			expectedComponents := []string{
				"func() float64",
				"math.Max",
				"High", "Low", "Close",
			}

			for _, component := range expectedComponents {
				if !contains(result, component) {
					t.Errorf("GenerateInitialValueAccess(%d) missing expected component: %s\nGot: %s", tt.period, component, result)
				}
			}

			/* Verify first bar edge case handling */
			if !contains(result, "if barIdx < 1") {
				t.Errorf("GenerateInitialValueAccess(%d) missing first bar check\nGot: %s", tt.period, result)
			}
		})
	}
}

func TestBuiltinTrueRangeAccessor_EdgeCases(t *testing.T) {
	accessor := NewBuiltinTrueRangeAccessor()

	t.Run("Loop access with offset 0", func(t *testing.T) {
		result := accessor.GenerateLoopValueAccess("0")
		if !contains(result, "barIdx := ctx.BarIndex-0") {
			t.Errorf("GenerateLoopValueAccess(0) should handle zero offset correctly\nGot: %s", result)
		}
	})

	t.Run("Initial value with period 1", func(t *testing.T) {
		result := accessor.GenerateInitialValueAccess(1)
		/* Period 1 means offset 0, should still have tr calculation */
		if !contains(result, "math.Max") {
			t.Errorf("GenerateInitialValueAccess(1) should generate tr calculation\nGot: %s", result)
		}
	})
}

func TestBuiltinTrueRangeAccessor_FirstBarFormula(t *testing.T) {
	accessor := NewBuiltinTrueRangeAccessor()

	t.Run("First bar uses High-Low", func(t *testing.T) {
		result := accessor.GenerateLoopValueAccess("j")

		/* Verify first bar formula: High - Low */
		if !contains(result, "return ctx.Data[barIdx].High - ctx.Data[barIdx].Low") {
			t.Errorf("First bar case should use High - Low\nGot: %s", result)
		}
	})

	t.Run("Subsequent bars use max of three components", func(t *testing.T) {
		result := accessor.GenerateLoopValueAccess("j")

		/* Verify full true range formula */
		expectedFormulaParts := []string{
			"prevClose := ctx.Data[barIdx-1].Close",
			"math.Max(currentBar.High - currentBar.Low",
			"math.Abs(currentBar.High - prevClose)",
			"math.Abs(currentBar.Low - prevClose)",
		}

		for _, part := range expectedFormulaParts {
			if !contains(result, part) {
				t.Errorf("True range formula missing component: %s\nGot: %s", part, result)
			}
		}
	})
}

func TestArrowFunctionTACallGenerator_CreateAccessorForTr(t *testing.T) {
	gen := newTestGenerator()
	taGen := NewArrowFunctionTACallGenerator(gen)

	trIdentifier := &ast.Identifier{Name: "tr"}
	accessor, err := taGen.createAccessorFromExpression(trIdentifier)

	if err != nil {
		t.Errorf("createAccessorFromExpression(tr) returned error: %v", err)
	}

	if accessor == nil {
		t.Fatal("createAccessorFromExpression(tr) returned nil accessor")
	}

	/* Verify correct accessor type */
	if _, ok := accessor.(*BuiltinTrueRangeAccessor); !ok {
		t.Errorf("createAccessorFromExpression(tr) returned wrong type: %T, want *BuiltinTrueRangeAccessor", accessor)
	}
}

func TestArrowFunctionTACallGenerator_TrNotConfusedWithVariable(t *testing.T) {
	gen := newTestGenerator()
	gen.variables = map[string]string{"my_tr": "float"}
	taGen := NewArrowFunctionTACallGenerator(gen)

	t.Run("tr builtin returns BuiltinTrueRangeAccessor", func(t *testing.T) {
		trIdentifier := &ast.Identifier{Name: "tr"}
		accessor, err := taGen.createAccessorFromExpression(trIdentifier)

		if err != nil {
			t.Fatalf("createAccessorFromExpression(tr) error: %v", err)
		}

		if _, ok := accessor.(*BuiltinTrueRangeAccessor); !ok {
			t.Errorf("tr should return BuiltinTrueRangeAccessor, got %T", accessor)
		}
	})

	t.Run("my_tr parameter returns ArrowFunctionParameterAccessor", func(t *testing.T) {
		myTrIdentifier := &ast.Identifier{Name: "my_tr"}
		accessor, err := taGen.createAccessorFromExpression(myTrIdentifier)

		if err != nil {
			t.Fatalf("createAccessorFromExpression(my_tr) error: %v", err)
		}

		if _, ok := accessor.(*ArrowFunctionParameterAccessor); !ok {
			t.Errorf("my_tr parameter should return ArrowFunctionParameterAccessor, got %T", accessor)
		}
	})
}

func TestBuiltinTrueRange_IntegrationWithTAFunctions(t *testing.T) {
	/* Test that tr accessor is correctly used in TA function contexts */
	gen := newTestGenerator()
	taGen := NewArrowFunctionTACallGenerator(gen)

	trIdentifier := &ast.Identifier{Name: "tr"}
	accessor, err := taGen.createAccessorFromExpression(trIdentifier)

	if err != nil {
		t.Fatalf("createAccessorFromExpression(tr) error: %v", err)
	}

	t.Run("tr with RMA loop iteration", func(t *testing.T) {
		loopAccess := accessor.GenerateLoopValueAccess("j")

		/* Verify inline calculation in loop */
		expectedPatterns := []string{
			"func() float64",
			"barIdx := ctx.BarIndex-j",
			"math.Max",
			"prevClose",
		}

		for _, pattern := range expectedPatterns {
			if !contains(loopAccess, pattern) {
				t.Errorf("RMA loop access missing pattern: %s", pattern)
			}
		}

		/* Verify NO Series.Get() */
		if contains(loopAccess, "Series.Get(") || contains(loopAccess, "trSeries") {
			t.Errorf("RMA loop should not use Series.Get(), got: %s", loopAccess)
		}
	})

	t.Run("tr with SMA initial value", func(t *testing.T) {
		initialAccess := accessor.GenerateInitialValueAccess(20)

		/* Verify inline calculation for initial value */
		if !contains(initialAccess, "func() float64") {
			t.Errorf("SMA initial value should generate inline IIFE")
		}

		if !contains(initialAccess, "math.Max") {
			t.Errorf("SMA initial value should calculate true range")
		}
	})
}

func TestBuiltinTrueRange_InArrowFunctionContext(t *testing.T) {
	/* Test tr accessor in arrow function TA call generator */
	gen := newTestGenerator()
	taGen := NewArrowFunctionTACallGenerator(gen)

	trIdentifier := &ast.Identifier{Name: "tr"}
	accessor, err := taGen.createAccessorFromExpression(trIdentifier)

	if err != nil {
		t.Fatalf("Arrow function createAccessorFromExpression(tr) error: %v", err)
	}

	/* Verify accessor is BuiltinTrueRangeAccessor */
	if _, ok := accessor.(*BuiltinTrueRangeAccessor); !ok {
		t.Fatalf("Arrow function should return BuiltinTrueRangeAccessor for tr, got %T", accessor)
	}

	/* Verify inline tr in arrow context loop */
	loopCode := accessor.GenerateLoopValueAccess("j")

	expectedPatterns := []string{
		"func() float64",
		"barIdx := ctx.BarIndex-j",
		"math.Max",
		"prevClose",
	}

	for _, pattern := range expectedPatterns {
		if !contains(loopCode, pattern) {
			t.Errorf("Arrow function tr loop missing pattern: %s", pattern)
		}
	}

	/* Critical: Verify NO trSeries.Get() */
	if contains(loopCode, "trSeries.Get(") || contains(loopCode, "Series.Get(") {
		t.Errorf("Arrow function should not generate Series.Get() for tr, got: %s", loopCode)
	}
}

func TestBuiltinTrueRange_ConsistencyAcrossContexts(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	t.Run("All contexts generate tr calculation", func(t *testing.T) {
		contexts := []struct {
			name   string
			method func() string
		}{
			{"current bar", func() string { return handler.GenerateCurrentBarAccess("tr") }},
			{"security context", func() string { return handler.GenerateSecurityContextAccess("tr") }},
			{"historical", func() string { return handler.GenerateHistoricalAccess("tr", 1) }},
		}

		for _, ctx := range contexts {
			t.Run(ctx.name, func(t *testing.T) {
				result := ctx.method()

				/* All contexts should generate inline calculation */
				requiredComponents := []string{"math.Max", "High", "Low"}
				for _, comp := range requiredComponents {
					if !contains(result, comp) {
						t.Errorf("%s context missing component: %s\nGot: %s", ctx.name, comp, result)
					}
				}
			})
		}
	})
}

func TestBuiltinTrueRange_NeverGeneratesSeriesAccess(t *testing.T) {
	/* Regression test: tr should NEVER generate Series.Get() calls */
	handler := NewBuiltinIdentifierHandler()
	accessor := NewBuiltinTrueRangeAccessor()

	tests := []struct {
		name   string
		method func() string
	}{
		{
			"current bar",
			func() string { return handler.GenerateCurrentBarAccess("tr") },
		},
		{
			"security context",
			func() string { return handler.GenerateSecurityContextAccess("tr") },
		},
		{
			"historical offset 1",
			func() string { return handler.GenerateHistoricalAccess("tr", 1) },
		},
		{
			"loop value access",
			func() string { return accessor.GenerateLoopValueAccess("j") },
		},
		{
			"initial value access",
			func() string { return accessor.GenerateInitialValueAccess(14) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method()

			/* Verify NO Series.Get() pattern */
			forbiddenPatterns := []string{
				"trSeries.Get(",
				".Get(tr",
				"Series.Get(",
			}

			for _, pattern := range forbiddenPatterns {
				if contains(result, pattern) {
					t.Errorf("%s generated forbidden Series access pattern: %s\nGot: %s", tt.name, pattern, result)
				}
			}

			/* Verify inline calculation markers present */
			if !contains(result, "math.Max") {
				t.Errorf("%s should generate inline calculation with math.Max\nGot: %s", tt.name, result)
			}
		})
	}
}
